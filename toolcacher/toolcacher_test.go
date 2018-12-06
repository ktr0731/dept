package toolcacher_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ktr0731/dept/toolcacher"
)

func TestCacher(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", dir)
	defer os.Setenv("GOPATH", gopath)

	tc, err := toolcacher.New()
	if err != nil {
		t.Errorf("failed to instantiate a toolcacher: %s", err)
	}

	cachePath, err := tc.Find("foo", "v0.1.0")
	if err != toolcacher.ErrCacheMiss {
		t.Fatalf("Find must return ErrCacheMiss, but got %s", err)
	}

	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("failed to create a temp file")
	}
	defer f.Close()

	cachePath, err = tc.Put(f.Name(), filepath.Base(f.Name()), "v0.1.0")
	if err != nil {
		t.Fatalf("Put must not return any errors, but got %s", err)
	}

	cachePath2, err := tc.Find(filepath.Base(f.Name()), "v0.1.0")
	if err != nil {
		t.Fatalf("Find must not return any errors, but got %s", err)
	}

	if cachePath != cachePath2 {
		t.Errorf("the path which Put returned and the path which Get returned must be equal, but actual '%s' and '%s'", cachePath, cachePath2)
	}
}
