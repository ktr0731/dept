package deptfile

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ktr0731/dept/deptfile/internal/deptfileutil"
	"github.com/pkg/errors"
)

// Workspace provides an environment to edit go.mod and go.sum.
// The environment is created in a temp dir.
type Workspace struct {
	// SourcePath is the root path for finding go.mod and go.sum.
	// If SourcePath is empty, SourcePath is the Git project root.
	SourcePath string
}

// Do copies from the project gotool.mod to a temporary workspace
// as a go.mod.
// Then, Do change the current dir to the workspace.
// After that, Do changes back the current dir and remove the created workspace.
//
// f receives projectDir which is the project root dir.
func (w *Workspace) Do(f func(projectDir string)) error {
	var err error
	cwd := w.SourcePath
	if cwd != "" {
		cwd, err = filepath.Abs(cwd)
		if err != nil {
			return errors.Wrapf(err, "failed to get abs path from %s", cwd)
		}
	} else {
		cwd, err = projectRoot()
		if err != nil {
			return err
		}
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return errors.Wrap(err, "failed to create a temp dir")
	}
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	defer os.Chdir(cwd)

	if err := deptfileutil.Copy("go.mod", filepath.Join(cwd, DeptfileName)); err != nil {
		return errors.Wrap(err, "failed to copy gotool.mod to the workspace")
	}
	if err := deptfileutil.Copy("go.sum", filepath.Join(cwd, DeptfileSumName)); err != nil {
		return errors.Wrap(err, "failed to copy gotool.sum to the workspace")
	}

	f(cwd)

	return nil
}

func projectRoot() (string, error) {
	var out, eout bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "--show-cdup")
	cmd.Stdout = &out
	cmd.Stderr = &eout
	if err := cmd.Run(); err != nil {
		return "", errors.Wrapf(err, "failed to get the project root: %s", eout.String())
	}
	p, err := filepath.Abs(out.String())
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse '%s' as abs path", out.String())
	}
	return p, nil
}
