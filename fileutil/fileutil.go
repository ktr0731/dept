package fileutil

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Copy copies 'from' to 'to'.
// If directries are not exist, Copy creates these.
func Copy(to, from string) error {
	toDir := filepath.Dir(to)
	if _, err := os.Stat(toDir); os.IsNotExist(err) {
		if err := os.MkdirAll(toDir, 0755); err != nil {
			return errors.Wrap(err, "failed to create destination directories")
		}
	}

	ff, err := os.Open(from)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", from)
	}
	defer ff.Close()
	fi, err := ff.Stat()
	if err != nil {
		return errors.Wrapf(err, "failed to get file info from %s", from)
	}

	tf, err := os.Create(to)
	if err != nil {
		return errors.Wrapf(err, "failed to create %s", to)
	}
	defer tf.Close()
	if err := tf.Chmod(fi.Mode()); err != nil {
		return errors.Wrap(err, "failed to chmod cached binary")
	}

	_, err = io.Copy(tf, ff)
	if err != nil {
		return errors.Wrapf(err, "failed to copy from %s to %s", from, to)
	}

	return nil
}
