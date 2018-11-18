package deptfile_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ktr0731/dept/deptfile"
)

func TestDo(t *testing.T) {
	t.Run("workspace changes directory to a temp dir and copies gotool.mod to there", func(t *testing.T) {
		cwd := "testdata"
		absCWD, err := filepath.Abs(cwd)
		if err != nil {
			t.Fatalf("failed to get abs path of %s", cwd)
		}

		w := &deptfile.Workspace{
			SourcePath: cwd,
		}
		err = w.Do(func(proj string) error {
			newcwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get current working dir: %s", err)
			}
			if cwd == newcwd {
				t.Errorf("current dir in Do must not be equal to dir outside of Do")
			}
			return nil
		})
		if err != nil {
			t.Errorf("Do must not return errors, but got an error: %s", err)
		}
		cwd2, err := os.Getwd()
		if err != nil {
			t.Errorf("failed to get current working dir: %s", err)
		}

		if absCWD != cwd2 {
			t.Errorf("current working dir which called before Do and after one must be equal, but %s and %s", absCWD, cwd2)
		}
	})

	t.Run("workspace returns ErrNotFound", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("failed to create a temp dir: %s", err)
		}
		w := &deptfile.Workspace{SourcePath: dir}
		err = w.Do(func(proj string) error {
			return nil
		})
		if err != deptfile.ErrNotFound {
			t.Errorf("workspace must return ErrNotFound because gotool.mod is not found in current working dir")
		}
	})
}
