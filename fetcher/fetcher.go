package fetcher

import "context"

type Fetcher interface {
	Fetch(ctx context.Context, repo string) error
}

type gitFetcher struct{}

func (f *gitFetcher) Fetch(ctx context.Context, repo string) error {
	return nil
}

// New returns an instance of the default fetcher implementation.
func New() Fetcher {
	return &gitFetcher{}
}
