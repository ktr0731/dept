package deptfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ktr0731/dept/deptfile"
)

func TestDo(t *testing.T) {
	cwd := "testdata"
	absCWD, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatalf("failed to get abs path of %s", cwd)
	}

	w := &deptfile.Workspace{
		SourcePath: cwd,
	}
	err = w.Do(func(proj string) {
		newcwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current working dir: %s", err)
		}
		if cwd == newcwd {
			t.Errorf("current dir in Do must not be equal to dir outside of Do")
		}
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
}
