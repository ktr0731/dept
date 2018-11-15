//go:generate moq -out mock.go . Fetcher

package fetcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const vendorDir = "vendor"

type Fetcher interface {
	Fetch(ctx context.Context, repo string) error
}

type gitFetcher struct{}

func (f *gitFetcher) Fetch(ctx context.Context, repo string) error {
	dir := VendorDir(repo)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "failed to create vendor directories")
	}
	if !strings.HasPrefix("http", repo) {
		repo = fmt.Sprintf("https://%s", repo)
	}
	if err := exec.CommandContext(ctx, "git", "clone", repo, dir).Run(); err != nil {
		return errors.Wrapf(err, "failed to clone the remote repository (url = %s, dest = %s)", repo, dir)
	}
	return nil
}

// New returns an instance of the default fetcher implementation.
func New() Fetcher {
	return &gitFetcher{}
}

// VendorDir receives an argument repo and
// returns the directory path which stored the downloaded remote repository.
func VendorDir(repo string) string {
	return filepath.Join(append([]string{vendorDir}, strings.Split(repo, "/")...)...)
}
