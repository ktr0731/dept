package gomod_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ktr0731/dept/gomod"
	"github.com/ktr0731/dept/gomod/internal/gomodutil"
)

// setupEnv creates a new temp dir for testing.
// Also setupEnv copies go.mod and go.sum from testdata to the temp dir.
// Finally, it changes current working directory to the temp dir.
// Callers must call returned function (cleanup function) at end of each test.
func setupEnv(t *testing.T) func() {
	t.Helper()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %s", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %s", err)
	}

	err = gomodutil.Copy(filepath.Join(dir, gomod.DeptfileName), filepath.Join("testdata", gomod.DeptfileName))
	if err != nil {
		t.Fatalf("failed to open and read testdata/gotool.mod: %s", err)
	}
	err = gomodutil.Copy(filepath.Join(dir, gomod.DeptfileSumName), filepath.Join("testdata", gomod.DeptfileSumName))
	if err != nil {
		t.Fatalf("failed to open and read testdata/gotool.mod: %s", err)
	}

	os.Chdir(dir)
	return func() {
		os.Chdir(pwd)
		os.RemoveAll(dir)
	}
}

func TestLoad(t *testing.T) {
	t.Run("Load must return ErrNotFound because deptfile missing", func(t *testing.T) {
		cleanup := setupEnv(t)
		defer cleanup()

		if err := os.Remove(gomod.DeptfileName); err != nil {
			t.Fatalf("failed to remove go.mod from the temp dir: %s", err)
		}

		_, err := gomod.Load(context.Background())
		if err != gomod.ErrNotFound {
			t.Fatalf("Load must return ErrNotFound, but got %s", err)
		}
	})

	t.Run("Load must return *File if deptfile found", func(t *testing.T) {
		cleanup := setupEnv(t)
		defer cleanup()

		m, err := gomod.Load(context.Background())
		if err != nil {
			t.Errorf("deptfile must be load by Load, but got an error: %s", err)
		}

		if n := len(m.Require); n != 6 {
			t.Errorf("number of required dependecies must be 3, but actual %d", n)
		}
	})
}
