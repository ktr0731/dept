package cmd_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ktr0731/dept/builder"
	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/fetcher"
)

func TestGetRun(t *testing.T) {
	t.Run("Run returns code 1 because no arguments passed", func(t *testing.T) {
		mockUI := newMockUI()
		cmd, err := cmd.Get(mockUI, nil, nil, nil)()
		if err != nil {
			t.Errorf("Get must not return some errors, but got an error: %s", err)
		}

		code := cmd.Run(nil)
		if code != 1 {
			t.Errorf("Run must be 1 because command need to show help message")
		}
		if out := mockUI.Writer().String(); !strings.HasPrefix(out, "Usage: ") {
			t.Errorf("Run must write help message to Writer, but actual '%s'", out)
		}
	})

	t.Run("Run returns code 0 normally", func(t *testing.T) {
		mockUI := newMockUI()
		mockFetcher := &fetcher.FetcherMock{
			FetchFunc: func(ctx context.Context, repo string) error { return nil },
		}
		mockBuilder := &builder.BuilderMock{
			BuildFunc: func() error { return nil },
		}

		df := &deptfile.File{Requirements: []*deptfile.Requirement{}}
		cmd, err := cmd.Get(mockUI, mockFetcher, mockBuilder, df)()
		if err != nil {
			t.Errorf("Get must not return some errors, but got an error: %s", err)
		}

		repo := "github.com/ktr0731/go-modules-test"
		code := cmd.Run([]string{repo})
		if code != 0 {
			t.Errorf("Run must be 0, but got %d", code)
		}

		if nr := len(df.Requirements); nr != 1 {
			t.Errorf("Get must add passed dependency to requirements, but %d dependency found", nr)
		}

		if n := len(mockFetcher.FetchCalls()); n != 1 {
			t.Errorf("Fetch must be called once, but actual %d", n)
		}

		if n := len(mockBuilder.BuildCalls()); n != 1 {
			t.Errorf("Build must be called once, but actual %d", n)
		}
	})
}
