package fetcher

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// mockFetcher is a mock of getFetcher.
// It works without network communication.
type mockFetcher struct {
	Dir string
}

// Fetch is a mock of gitFetcher.Fetch.
// It creates directory specified by repo.
// If repo is separated by '/', it will be interpret as each directory.
// For example, repo is 'foo/bar/baz', created directories are:
//
//   vendor-test
//     - foo
//       - bar
//         - baz
//
// vendor-test is used to the destination of mockFetcher. It is specified by mockFetcher.Dir.
// If vendor-test is not exist, Fetch creates it.
func (f *mockFetcher) Fetch(ctx context.Context, repo string) error {
	dir := filepath.Join(append([]string{f.Dir}, strings.Split(repo, "/")...)...)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.Wrapf(err, "failed to create directories: %s", dir)
		}
	}

	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get current directory")
	}
	defer os.Chdir(pwd)
	os.Chdir(dir)

	if err := exec.CommandContext(ctx, "git", "init").Run(); err != nil {
		return errors.Wrap(err, "failed to init passed repo as a Git repository")
	}

	// To run `go build` as Go modules aware mode, we need to set origin as a GitHub repository.
	// GitHub repository URL has no meaning.
	if err := exec.CommandContext(ctx, "git", "remote", "add", "origin", "https://github.com/ktr0731/evans").Run(); err != nil {
		return errors.Wrap(err, "failed to set temporary origin")
	}

	return nil
}

// Cleanup cleans up all directories used by mockFetcher.
func (f *mockFetcher) Cleanup() error {
	err := os.RemoveAll(f.Dir)
	if err != nil {
		return errors.Wrap(err, "failed to remove test vendor directory")
	}
	return nil
}

func NewMock() *mockFetcher {
	return &mockFetcher{Dir: "vendor-test"}
}
