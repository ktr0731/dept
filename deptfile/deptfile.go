package deptfile

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ktr0731/dept/deptfile/internal/deptfileutil"
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

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(dir)
	defer os.Chdir(pwd)

	// TODO: module name
	if err := exec.CommandContext(ctx, "go", "mod", "init", "tools").Run(); err != nil {
		return err
	}

	deptfileutil.Copy(filepath.Join(pwd, DeptfileName), filepath.Join(dir, "go.mod"))
	deptfileutil.Copy(filepath.Join(pwd, DeptfileSumName), filepath.Join(dir, "go.sum"))

	f, err := os.Create(DeptfileName)
	if err != nil {
		return errors.Wrap(err, "failed to create deptfile")
	}
	defer f.Close()

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

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a temp dir")
	}
	defer os.RemoveAll(dir)

	if err := deptfileutil.Copy(filepath.Join(dir, "go.mod"), DeptfileName); err != nil {
		return nil, err
	}
	if err := deptfileutil.Copy(filepath.Join(dir, "go.sum"), DeptfileSumName); err != nil {
		return nil, err
	}

	os.Chdir(dir)

	var out, eout bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "mod", "edit", "-json")
	cmd.Stdout = &out
	cmd.Stderr = &eout
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "failed to execute 'go mod edit -json': %s", eout.String())
	}

	var m GoMod
	err = json.NewDecoder(&out).Decode(&m)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open or decode %s", DeptfileName)
	}

	return &m, nil
}
