package fetcher

import (
	"os"
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
func (f *mockFetcher) Fetch(repo string) error {
	dir := filepath.Join(append([]string{f.Dir}, strings.Split(repo, "/")...)...)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.Wrapf(err, "failed to create directories: %s", dir)
		}
	}
	return nil
}

// Cleanup cleans up all directories used by mockFetcher.
func (f *mockFetcher) Cleanup() error {
	err := os.RemoveAll(f.Dir)
	if err != nil {
		return errors.Wrapf(err, "failed to remove test vendor directory: %s", err)
	}
	return nil
}

func NewMock() *mockFetcher {
	return &mockFetcher{Dir: "vendor-test"}
}
