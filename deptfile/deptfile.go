package deptfile

import (
	"io"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
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
