package fetcher_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ktr0731/dept/fetcher"
)

func TestNewMock(t *testing.T) {
	f := fetcher.NewMock()
	defer f.Cleanup()

	err := f.Fetch("foo/bar/baz")
	if err != nil {
		t.Fatalf("Fetch must not return some errors, but got an error: %s", err)
	}

	dir := filepath.Join(f.Dir, "foo/bar/baz")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Fetch must create directories '%s', but missing", dir)
	}
}
