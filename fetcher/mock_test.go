package fetcher_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ktr0731/dept/fetcher"
)

func TestNewMock(t *testing.T) {
	f := fetcher.NewMock()
	defer f.Cleanup()

	err := f.Fetch(context.Background(), "foo/bar/baz")
	if err != nil {
		t.Fatalf("Fetch must not return some errors, but got an error: %s", err)
	}

	dir := filepath.Join(f.Dir, "foo/bar/baz")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Fetch must create directories '%s', but missing", dir)
	}

	os.Chdir(dir)

	b, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		t.Errorf("created repository must be a Git repository, but Git command failed: %s", err)
	}

	if out := string(b); !strings.HasPrefix(out, "git@github.com") {
		t.Errorf("the result of 'git remote get-url origin' must be contain 'git@github.com', but actual %s", out)
	}
}
