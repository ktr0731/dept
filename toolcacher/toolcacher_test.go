package toolcacher_test

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	"os"
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
	}

	tc, err := toolcacher.New(gocmd)
	if err != nil {
		t.Fatalf("failed to instantiate a toolcacher: %s", err)
	}

	return tc, gocmd, func() {
		os.RemoveAll(dir)
	}
}

func TestCacher(t *testing.T) {
	t.Run("Find must return ErrCacheMiss because foo is not found", func(t *testing.T) {
		tc, gocmd, cleanup := setup(t)
		defer cleanup()

		_, err := tc.Find("foo", "v0.1.0")
		if err != toolcacher.ErrCacheMiss {
			t.Fatalf("Find must return ErrCacheMiss, but got %s", err)
		}

		if n := len(gocmd.EnvCalls()); n != 1 {
			t.Errorf("'go env' must be call once to get $GOPATH, but actual %d times called", n)
		}
	})

	t.Run("Find must not return any errors because tool was cached", func(t *testing.T) {
		tc, gocmd, cleanup := setup(t)
		defer cleanup()

		f, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatalf("failed to create a temp file")
		}
		defer f.Close()

		pkgName := "github.com/hoge/fuga/foo"
		version := "v0.1.0"
		cachePath, err := tc.Put(f.Name(), pkgName, version)
		if err != nil {
			t.Fatalf("Put must not return any errors, but got %s", err)
		}

		cachePath2, err := tc.Find(pkgName, version)
		if err != nil {
			t.Fatalf("Find must not return any errors, but got %s", err)
		}

		if cachePath != cachePath2 {
			t.Errorf("the path which Put returned and the path which Get returned must be equal, but actual '%s' and '%s'", cachePath, cachePath2)
		}

		if n := len(gocmd.EnvCalls()); n != 1 {
			t.Errorf("'go env' must be call once to get $GOPATH, but actual %d times called", n)
		}
	})

	t.Run("Get builds a new tool for caching", func(t *testing.T) {
		tc, gocmd, cleanup := setup(t)
		defer cleanup()

		pkgName := "github.com/hoge/fuga/foo"
		version := "v0.1.0"
		cachePath, err := tc.Get(context.Background(), pkgName, version)
		if err != nil {
			t.Fatalf("Get must not return any errors, but got %s", err)
		}

		// Create a file which is used as a pseudo binary file.
		// Next Get call will checks it to determine build a new binary or not.
		if n := len(gocmd.BuildCalls()); n != 1 {
			t.Errorf("'go build' must be call once to build the passed tool, but actual %d times called", n)
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

		if cachePath != cachePath2 {
			t.Errorf("Two returned paths must be equal, but actual '%s' and '%s'", cachePath, cachePath2)
		}
	})
}
