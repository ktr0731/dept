package toolcacher

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ktr0731/dept/gocmd"
	"github.com/pkg/errors"
)

var (
	ErrCacheMiss = errors.New("specified tool is not found")
	ErrNotFound  = errors.New("original tool path is not found")
)

type Cacher interface {
	// Find finds a cached tool path which satisfies the passed toolName and version.
	// If it is not cached, FindTool returns ErrCacheMiss.
	Find(name, version string) (path string, err error)
	// Put puts a built binary to a cache storage.
	// The cached binary can be extract by Find(name, version) or the returned path.
	Put(originalPath, name, version string) (path string, err error)
}

type cacher struct {
	rootPath string
}

func New() (Cacher, error) {
	gocmd := gocmd.New()
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
		rootPath: rootPath,
	}, nil
}

func (c *cacher) Find(name, version string) (string, error) {
	toolName := cacheName(name, version)
	cachePath := filepath.Join(c.rootPath, toolName)
	_, err := os.Stat(cachePath)
	if os.IsNotExist(err) {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", errors.Wrap(err, "failed to get tool binary info")
	}
	return cachePath, nil
}

func (c *cacher) Put(path, name, version string) (string, error) {
	toolName := cacheName(name, version)
	cachePath := filepath.Join(c.rootPath, toolName)
	from, err := os.Open(path)
	if os.IsNotExist(err) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", errors.Wrap(err, "failed to open the target binary file")
	}
	defer from.Close()
	to, err := os.Create(cachePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to create a cache file")
	}
	defer to.Close()
	io.Copy(to, from)
	return cachePath, nil
}

func cacheName(name, version string) string {
	if name == "" || version == "" {
		panic("name and version must not be nil")
	}
	return fmt.Sprintf("%s-%s", name, version)
}
