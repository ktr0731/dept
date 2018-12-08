//go:generate moq -out mock_gen.go . Cacher

package toolcacher

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ktr0731/dept/gocmd"
	"github.com/ktr0731/dept/logger"
	"github.com/pkg/errors"
)

var (
	errCacheMiss = errors.New("specified tool is not found")
)

type Cacher interface {
	// Get finds a cached tool path which satisfies the passed pkgName and version.
	// If it is not cached, Get builds a new one.
	Get(ctx context.Context, pkgName, version string) (path string, err error)
	// Clear removes all cached tools.
	Clear(ctx context.Context) error
}

type cacher struct {
	gocmd        gocmd.Command
	rootPath     string
	downloadOnce sync.Once
}

func New(gocmd gocmd.Command) (Cacher, error) {
	r, err := gocmd.Env(context.Background(), "GOPATH")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get $GOPATH")
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read $GOPATH value")
	}
	gopath := strings.TrimSpace(string(b))
	rootPath := filepath.Join(gopath, "pkg", "dept")
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		if err := os.MkdirAll(rootPath, 0755); err != nil {
			return nil, errors.Wrap(err, "failed to create a cache dir")
		}
	}
	return &cacher{
		gocmd:    gocmd,
		rootPath: rootPath,
	}, nil
}

func (c *cacher) find(key string) (string, error) {
	_, err := os.Stat(key)
	if os.IsNotExist(err) {
		return "", errCacheMiss
	}
	if err != nil {
		return "", errors.Wrap(err, "failed to get tool binary info")
	}
	return key, nil
}

func (c *cacher) Get(ctx context.Context, pkgName, version string) (string, error) {
	outPath := c.cachePath(pkgName, version)
	cachePath, err := c.find(outPath)
	if err == nil {
		logger.Printf("tool cache found: %s", cachePath)
		return cachePath, nil
	}

	if err == errCacheMiss {
		logger.Printf("cache passed tool: %s %s", outPath, pkgName)

		var err error
		c.downloadOnce.Do(func() {
			logger.Println("downloading modules")
			if err = c.gocmd.ModDownload(ctx); err != nil {
				err = errors.Wrap(err, "failed to download module dependencies")
			}
		})
		if err != nil {
			return "", err
		}

		if err := c.gocmd.Build(ctx, "-o", outPath, pkgName); err != nil {
			return "", errors.Wrapf(err, "failed to cache tool %s to %s", pkgName, outPath)
		}
		return outPath, nil
	}
	return "", errors.Wrap(err, "failed to find the passed tool")
}

func (c *cacher) Clear(ctx context.Context) error {
	logger.Printf("remove %s", c.rootPath)
	err := os.RemoveAll(c.rootPath)
	if err != nil {
		return errors.Wrap(err, "failed to remove all cached tools")
	}
	return nil
}

func (c *cacher) cachePath(pkgName, version string) string {
	if pkgName == "" || version == "" {
		panic("pkgName and version must not be nil")
	}
	return filepath.Join(
		c.rootPath,
		fmt.Sprintf("%s-%s", strings.Replace(pkgName, "/", "-", -1), version),
	)
}
