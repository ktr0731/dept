package fetcher

type Fetcher interface {
	Fetch(repo string) error
}

type gitFetcher struct{}

func (f *gitFetcher) Fetch(repo string) error {
	return nil
}

// New returns an instance of the default fetcher implementation.
func New() Fetcher {
	return &gitFetcher{}
}
