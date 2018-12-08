package toolcacher_test

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ktr0731/dept/gocmd"
	"github.com/ktr0731/dept/toolcacher"
)

func setup(t *testing.T) (toolcacher.Cacher, *gocmd.CommandMock, func()) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %s", err)
	}

	gocmd := &gocmd.CommandMock{
		EnvFunc: func(ctx context.Context, args ...string) (io.Reader, error) {
			return strings.NewReader(dir), nil
		},
		BuildFunc: func(ctx context.Context, args ...string) error {
			return nil
		},
		ModDownloadFunc: func(ctx context.Context) error {
			return nil
		},
	}

	tc, err := toolcacher.New(gocmd)
	if err != nil {
		t.Fatalf("failed to instantiate a toolcacher: %s", err)
	}

	if n := len(gocmd.EnvCalls()); n != 1 {
		t.Fatalf("Env must be called once, but actual %d times called", n)
	}

	return tc, gocmd, func() {
		os.RemoveAll(dir)
	}
}

func TestCacher(t *testing.T) {
	t.Run("Get builds a new tool for caching", func(t *testing.T) {
		tc, gocmd, cleanup := setup(t)
		defer cleanup()

		pkgName := "github.com/hoge/fuga/foo"
		version := "v0.1.0"
		cachePath, err := tc.Get(context.Background(), pkgName, version)
		if err != nil {
			t.Fatalf("Get must not return any errors, but got '%s'", err)
		}

		// Create a file which is used as a pseudo binary file.
		// Next Get call will checks it to determine build a new binary or not.
		if n := len(gocmd.BuildCalls()); n != 1 {
			t.Errorf("'go build' must be call once to build the passed tool, but actual %d times called", n)
		}
		if n := len(gocmd.ModDownloadCalls()); n != 1 {
			t.Errorf("'go mod download' must be called once in a single execution, but actual %d times called", n)
		}
		fs := flag.NewFlagSet("test", flag.ExitOnError)
		outputPath := fs.String("o", "", "")
		fs.Parse(gocmd.BuildCalls()[0].Args)
		f, err := os.Create(*outputPath)
		if err != nil {
			t.Fatalf("failed to create a pseudo binary file: %s", err)
		}
		defer f.Close()

		cachePath2, err := tc.Get(context.Background(), pkgName, version)
		if err != nil {
			t.Fatalf("Get must not return any errors, but got %s", err)
		}

		if n := len(gocmd.BuildCalls()); n != 1 {
			t.Errorf("'go build' must not be call again, but %d times called", n)
		}
		if n := len(gocmd.ModDownloadCalls()); n != 1 {
			t.Errorf("'go mod download' must be called once in a single execution, but actual %d times called", n)
		}

		if cachePath != cachePath2 {
			t.Errorf("Two returned paths must be equal, but actual '%s' and '%s'", cachePath, cachePath2)
		}
	})

	t.Run("Get will panic", func(t *testing.T) {
		tc, _, cleanup := setup(t)
		defer cleanup()

		cases := []struct {
			pkgName, version string
		}{
			{"foo", ""},
			{"", "v0.1.0"},
			{"", ""},
		}
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s@%s", c.pkgName, c.version), func(t *testing.T) {
				defer func() {
					if err := recover(); err == nil {
						t.Error("Get must panic if pkgName or version is empty, but got nil")
					}
				}()

				_, _ = tc.Get(context.Background(), "", "")
			})
		}
	})

	t.Run("Clear removes the cache dir", func(t *testing.T) {
		tc, gocmd, cleanup := setup(t)
		defer cleanup()

		e, _ := gocmd.Env(context.Background())
		b, _ := ioutil.ReadAll(e)
		gopath := string(b)

		err := tc.Clear(context.TODO())
		if err != nil {
			t.Fatalf("Clear must not return any errors, but got %s", err)
		}

		dir := filepath.Join(gopath, "pkg", "dept")
		if _, err := os.Stat(dir); err == nil {
			t.Errorf("Clear must remove %s, but didn't", dir)
		}
	})
}
