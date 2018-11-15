package deptfile

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

var DeptfileName = "dept.toml"

var (
	// ErrNotFound represents deptfile not found.
	ErrNotFound = errors.Errorf("%s not found", DeptfileName)
	// ErrAlreadyExist represents deptfile alredy exist.
	ErrAlreadyExist = errors.New("already exist")
)

type File struct {
	Requirements []*Requirement `toml:"requirements"`

	Writer io.Writer `toml:"-"`
}

type Requirement struct {
	Name string `toml:"name"`
}

// Encode encodes itself to f.Writer. If f.Writer is nil, os.Stdout will be used instead.
// Encoding format is TOML.
func (f *File) Encode() error {
	w := f.Writer
	if w == nil {
		w = os.Stdout
	}
	if err := toml.NewEncoder(w).Encode(f); err != nil {
		return errors.Wrap(err, "failed to encode deptfile")
	}
	return nil
}

// Load loads deptfile from current directory.
// If deptfile not found, Load returns ErrNotFound.
func Load() (*File, error) {
	if _, err := os.Stat(DeptfileName); os.IsNotExist(err) {
		return nil, ErrNotFound
	}

	var f File
	_, err := toml.DecodeFile(DeptfileName, &f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open or decode %s", DeptfileName)
	}

	return &f, nil
}
