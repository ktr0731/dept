package gomodutil

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

// Copy copies 'from' to 'to'.
func Copy(to, from string) error {
	ff, err := os.Open(from)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", from)
	}
	defer ff.Close()

	tf, err := os.Create(to)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", to)
	}
	defer tf.Close()

	_, err = io.Copy(tf, ff)
	if err != nil {
		return errors.Wrapf(err, "failed to copy from %s to %s", from, to)
	}

	return nil
}
