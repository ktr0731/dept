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
	"github.com/ktr0731/dept/fileutil"
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
// Also setupEnv copies go.mod and go.sum from cwd to the temp dir.
// Finally, it changes current working directory to the temp dir.
// Callers must call returned function (cleanup function) at end of each test.
func setupEnv(t *testing.T, cwd string) func() {
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

	err = fileutil.Copy(filepath.Join(dir, deptfile.FileName), filepath.Join(cwd, deptfile.FileName))
	if err != nil {
		t.Fatalf("failed to open and read testdata/gotool.mod: %s", err)
	}
	err = fileutil.Copy(filepath.Join(dir, deptfile.FileSumName), filepath.Join(cwd, deptfile.FileSumName))
	if err != nil {
		t.Fatalf("failed to open and read testdata/gotool.mod: %s", err)
	}

	_ = os.Chdir(dir)
	return func() {
		_ = os.Chdir(pwd)
		os.RemoveAll(dir)
	}
}

func TestCreate(t *testing.T) {
	t.Run("Create must return ErrAlreadyExist because deptfile exists", func(t *testing.T) {
		cleanup := setupEnv(t, filepath.Join("testdata", "normal"))
		defer cleanup()

		err := deptfile.Create(context.Background())
		if err == nil {
			t.Error("Create must return an error, but got nil")
		}
	})

	t.Run("Create must not return errors", func(t *testing.T) {
		cleanup := setupEnv(t, filepath.Join("testdata", "normal"))
		defer cleanup()

		if err := os.Remove(deptfile.FileName); err != nil {
			t.Fatalf("failed to remove go.mod from the temp dir: %s", err)
		}
		if err := os.Remove(deptfile.FileSumName); err != nil {
			t.Fatalf("failed to remove go.sum from the temp dir: %s", err)
		}

		err := deptfile.Create(context.Background())
		if err != nil {
			t.Fatalf("Create must not return an error, but got: %s", err)
		}

		if _, err := os.Stat(deptfile.FileName); os.IsNotExist(err) {
			t.Error("after Create called, deptfile is in current dir, but missing")
		}
	})
}
