//go:generate moq -out mock_gen.go . Cacher

package toolcacher

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ktr0731/dept/fileutil"
	"github.com/ktr0731/dept/gocmd"
	"github.com/ktr0731/dept/logger"
	"github.com/pkg/errors"
)

var (
	ErrCacheMiss = errors.New("specified tool is not found")
	ErrNotFound  = errors.New("original tool path is not found")
)

type Cacher interface {
	// Find finds a cached tool path which satisfies the passed pkgName and version.
	// If it is not cached, FindTool returns ErrCacheMiss.
	Find(pkgName, version string) (path string, err error)
	// Get finds a cached tool path which satisfies the passed pkgName and version.
	// If it is not cached, Get builds a new one.
	Get(ctx context.Context, pkgName, version string) (path string, err error)
	// Put puts a built binary to a cache storage.
	// The cached binary can be extract by Find(pkgName, version) or the returned path.
	Put(originalPath, pkgName, version string) (path string, err error)
}

type cacher struct {
	gocmd    gocmd.Command
	rootPath string
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
		os.MkdirAll(rootPath, 0755)
	}
	return &cacher{
		gocmd:    gocmd,
		rootPath: rootPath,
	}, nil
}

func (c *cacher) Find(pkgName, version string) (string, error) {
	return c.find(c.cachePath(pkgName, version))
}

func (c *cacher) find(key string) (string, error) {
	_, err := os.Stat(key)
	if os.IsNotExist(err) {
		return "", ErrCacheMiss
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

	if err == ErrCacheMiss {
		logger.Printf("cache passed tool: %s %s", outPath, pkgName)
		if err := c.gocmd.Build(ctx, "-o", outPath, pkgName); err != nil {
			return "", errors.Wrapf(err, "failed to cache tool %s to %s", pkgName, outPath)
		}
		return outPath, nil
	}
	return "", errors.Wrap(err, "failed to find the passed tool")
}

func (c *cacher) Put(path, pkgName, version string) (string, error) {
	cachePath := c.cachePath(pkgName, version)
	if err := fileutil.Copy(cachePath, path); err != nil {
		return "", errors.Wrapf(err, "failed to copy %s to %s", path, cachePath)
	}
	return cachePath, nil
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
