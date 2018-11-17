package deptfile

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// RunInWorkspace provides a temporary workspace to edit gotool.mod.
// RunInWorkspace copies from the project gotool.mod to a temporary workspace
// as a go.mod.
// Then, RunInWorkspace change the current dir to the workspace.
// After that, RunInWorkspace changes back the current dir and remove the created workspace.
func RunInWorkspace(f func()) error {
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get current dir")
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return errors.Wrap(err, "failed to create a temp dir")
	}
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	defer os.Chdir(pwd)

	f()

	return nil
}
