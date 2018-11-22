package deptfile_test

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/deptfile/internal/deptfileutil"
)

var verbose = flag.Bool("verbose", false, "verbose mode")
var l *log.Logger

func init() {
	flag.Parse()
	if *verbose {
		l = log.New(os.Stderr, "[debug] ", log.LstdFlags|log.Lshortfile)
	} else {
		l = log.New(ioutil.Discard, "[debug] ", log.Lshortfile)
	}
}

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
	l.Printf("setupEnv: created a temp dir %s\n", dir)

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %s", err)
	}

	err = deptfileutil.Copy(filepath.Join(dir, deptfile.DeptfileName), filepath.Join("testdata", deptfile.DeptfileName))
	if err != nil {
		t.Fatalf("failed to open and read testdata/gotool.mod: %s", err)
	}
	err = deptfileutil.Copy(filepath.Join(dir, deptfile.DeptfileSumName), filepath.Join("testdata", deptfile.DeptfileSumName))
	if err != nil {
		t.Fatalf("failed to open and read testdata/gotool.mod: %s", err)
	}

	os.Chdir(dir)
	return func() {
		os.Chdir(pwd)
		os.RemoveAll(dir)
	}
}

func TestCreate(t *testing.T) {
	t.Run("Create must return ErrAlreadyExist because deptfile exists", func(t *testing.T) {
		cleanup := setupEnv(t)
		defer cleanup()

		err := deptfile.Create(context.Background())
		if err == nil {
			t.Error("Create must return an error, but got nil")
		}
	})

	t.Run("Create must not return errors", func(t *testing.T) {
		cleanup := setupEnv(t)
		defer cleanup()

		if err := os.Remove(deptfile.DeptfileName); err != nil {
			t.Fatalf("failed to remove go.mod from the temp dir: %s", err)
		}
		if err := os.Remove(deptfile.DeptfileSumName); err != nil {
			t.Fatalf("failed to remove go.sum from the temp dir: %s", err)
		}

		err := deptfile.Create(context.Background())
		if err != nil {
			t.Fatalf("Create must not return an error, but got: %s", err)
		}

		if _, err := os.Stat(deptfile.DeptfileName); os.IsNotExist(err) {
			t.Error("after Create called, deptfile is in current dir, but missing")
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("Load must return ErrNotFound because go.mod missing", func(t *testing.T) {
		cleanup := setupEnv(t)
		defer cleanup()

		w := &deptfile.Workspace{SourcePath: "."}
		err := w.Do(func(string, *deptfile.GoMod) error {
			if err := os.Remove("go.mod"); err != nil {
				t.Fatalf("failed to remove go.mod from the temp dir: %s", err)
			}

			_, err := deptfile.Load(context.Background())
			return err
		})
		if err != deptfile.ErrNotFound {
			t.Fatalf("Load must return ErrNotFound, but got '%s'", err)
		}
	})

	t.Run("Load must return *File if deptfile found", func(t *testing.T) {
		cleanup := setupEnv(t)
		defer cleanup()

		w := &deptfile.Workspace{SourcePath: "."}
		err := w.Do(func(string, *deptfile.GoMod) error {
			m, err := deptfile.Load(context.Background())
			if err != nil {
				t.Fatalf("deptfile must be load by Load, but got an error: %s", err)
			}

			if n := len(m.Require); n != 6 {
				t.Errorf("number of required dependecies must be 6, but actual %d", n)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Do must not return errors, but got %s", err)
		}
	})
}
