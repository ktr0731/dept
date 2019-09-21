//go:generate moq -out mock_gen.go . Workspacer

package deptfile

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ktr0731/dept/fileutil"
	"github.com/ktr0731/modfile"
	"github.com/pkg/errors"
)

// Workspacer provides an environment to edit go.mod and go.sum.
type Workspacer interface {
	// Do copies gotool.mod and gotool.sum to the workspace.
	// If gotool.mod is not found, Do returns ErrNotFound.
	Do(f func(projectDir string, gomod *File) error) error
}

// Workspace is an implementation for Workspacer.
// The environment is created in a temp dir.
type Workspace struct {
	// SourcePath is the root path for finding go.mod and go.sum.
	// If SourcePath is empty, SourcePath is the Git project root.
	SourcePath string
	// DoNotCopy doesn't copy gotool.mod and gotool.sum to the workspace.
	DoNotCopy bool
	// DoNotUpdate doesn't update gotool.mod and gotool.sum.
	// It is used for commands which doesn't need to update gotool.mod such like 'dept build'.
	DoNotUpdate bool
}

// Do copies from the project gotool.mod to a temporary workspace
// as a go.mod.
// Then, Do change the current dir to the workspace.
// After that, Do changes back the current dir and remove the created workspace.
//
// f receives projectDir which is the project root dir.
func (w *Workspace) Do(f func(projectDir string, gomod *File) error) error {
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
			cwd, err = os.Getwd()
			if err != nil {
				return errors.Wrap(err, "failed to get the current working dir")
			}
		}
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return errors.Wrap(err, "failed to create a temp dir")
	}
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	defer os.Chdir(cwd)

	var gomod *File
	var canonicalModFile *modfile.File
	// Parse deptfile and write out canonical formed modfile to go.mod.
	// After that, f treats this go.mod.
	if !w.DoNotCopy {
		gomod, canonicalModFile, err = parseDeptfile(filepath.Join(cwd, FileName))
		if err == ErrNotFound {
			return ErrNotFound
		}
		if err != nil {
			return errors.Wrap(err, "failed to initialize *File")
		}
		b, err := canonicalModFile.Format()
		if err != nil {
			return errors.Wrap(err, "failed to format canonicalized modfile")
		}
		if err := ioutil.WriteFile("go.mod", b, 0644); err != nil {
			return errors.Wrap(err, "failed to write out go.mod")
		}

		// ignore errors because it is auto-generated file.
		fileutil.Copy("go.sum", filepath.Join(cwd, FileSumName))
	}

	if err := f(cwd, gomod); err != nil {
		return err
	}

	if w.DoNotUpdate {
		return nil
	}

	df, err := convertGoModToDeptfile("go.mod", gomod)
	if err != nil {
		return errors.Wrap(err, "failed to convert from go.mod to deptfile")
	}

	b, err := df.Format()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(cwd, FileName), b, 0644); err != nil {
		return errors.Wrap(err, "failed to write gotool.mod")
	}

	fileutil.Copy(filepath.Join(cwd, FileSumName), "go.sum")

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
	return strings.TrimSpace(p), nil
}
