package deptfile

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

const (
	DeptfileName    = "gotool.mod"
	DeptfileSumName = "gotool.sum"
)

var (
	// ErrNotFound represents deptfile not found.
	ErrNotFound = errors.Errorf("%s not found", DeptfileName)
	// ErrAlreadyExist represents deptfile alredy exist.
	ErrAlreadyExist = errors.New("already exist")
)

// Create creates a new deptfile.
// If already created, Create returns ErrAlreadyExist.
func Create(ctx context.Context) error {
	if _, err := os.Stat(DeptfileName); err == nil {
		return ErrAlreadyExist
	}

	var err error
	w := &Workspace{
		SourcePath: ".",
		DoNotCopy:  true,
	}
	werr := w.Do(func(string) {
		// TODO: module name
		err = exec.CommandContext(ctx, "go", "mod", "init", "tools").Run()
		if err != nil {
			err = errors.Wrap(err, "failed to init Go modules")
			return
		}
	})
	if werr != nil {
		return errors.Wrap(werr, "failed to setup a workspace")
	}
	if err != nil {
		return err
	}

	return nil
}

// Load loads deptfile from current directory.
// If deptfile not found, Load returns ErrNotFound.
//
// 'go mod' command reads go.mod as a Go modules file so that
// we creates a temp dir and copies gotool.mod and gotool.sum to there.
// Then execute 'go mod' command.
func Load(ctx context.Context) (*GoMod, error) {
	if _, err := os.Stat(DeptfileName); os.IsNotExist(err) {
		return nil, ErrNotFound
	}

	var m GoMod
	var err error
	w := &Workspace{
		SourcePath: ".",
	}
	werr := w.Do(func(string) {
		var out, eout bytes.Buffer
		cmd := exec.CommandContext(ctx, "go", "mod", "edit", "-json")
		cmd.Stdout = &out
		cmd.Stderr = &eout
		if err := cmd.Run(); err != nil {
			err = errors.Wrapf(err, "failed to execute 'go mod edit -json': %s", eout.String())
			return
		}

		if e := eout.String(); e != "" {
			err = errors.New(e)
			return
		}

		err = json.NewDecoder(&out).Decode(&m)
		if err != nil {
			err = errors.Wrapf(err, "failed to open or decode %s", DeptfileName)
			return
		}
	})
	if werr != nil {
		return nil, errors.Wrap(werr, "failed to setup a workspace")
	}
	if err != nil {
		return nil, errors.Wrap(err, "an error occurred in the workspace")
	}
	return &m, nil
}
