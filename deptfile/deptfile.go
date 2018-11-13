package deptfile

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

var deptfileName = "dept.toml"

var (
	ErrNotFound = errors.Errorf("%s not found", deptfileName)
)

type File struct {
	Requirements []*Requirement `toml:"requirements"`
}

type Requirement struct {
	Name string `toml:"name"`
}

func (f *File) Encode(w io.Writer) error {
	if err := toml.NewEncoder(w).Encode(f); err != nil {
		return errors.Wrap(err, "failed to encode deptfile")
	}
	return nil
}

// Load loads deptfile from current directory.
// If deptfile not found, Load returns ErrNotFound.
func Load() (*File, error) {
	if _, err := os.Stat(deptfileName); os.IsNotExist(err) {
		return nil, ErrNotFound
	}

	var f File
	_, err := toml.DecodeFile(deptfileName, &f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open or decode %s", deptfileName)
	}

	return &f, nil
}
