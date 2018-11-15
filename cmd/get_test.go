package cmd_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ktr0731/dept/builder"
	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/fetcher"
	"github.com/pkg/errors"
)

func TestGetRun(t *testing.T) {
	t.Run("Run returns code 1 because no arguments passed", func(t *testing.T) {
		mockUI := newMockUI()
		cmd := cmd.NewGet(mockUI, nil, nil)

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
			BuildFunc: func(ctx context.Context, dir string) error { return nil },
		}

		var out bytes.Buffer
		df := &deptfile.File{
			Requirements: []*deptfile.Requirement{},
			Writer:       &out,
		}
		cleanup := cmd.ChangeDeptfileLoad(func() (*deptfile.File, error) {
			return df, nil
		})
		defer cleanup()

		cmd := cmd.NewGet(mockUI, mockFetcher, mockBuilder)

		repo := "github.com/ktr0731/go-modules-test"
		code := cmd.Run([]string{repo})
		if code != 0 {
			t.Errorf("Run must be 0, but got %d", code)
		}

		if n := len(mockFetcher.FetchCalls()); n != 1 {
			t.Errorf("Fetch must be called once, but actual %d", n)
		}

		if n := len(mockBuilder.BuildCalls()); n != 1 {
			t.Errorf("Build must be called once, but actual %d", n)
		}

		if n := len(df.Requirements); n != 1 {
			t.Errorf("Get must add passed dependency to requirements, but %d dependency found", n)
		}

		if out.String() == "" {
			t.Error("Encode must be called")
		}
	})

	t.Run("deptfile is not modified when command failed", func(t *testing.T) {
		mockUI := newMockUI()
		mockFetcher := &fetcher.FetcherMock{
			FetchFunc: func(ctx context.Context, repo string) error { return errors.New("an error") },
		}

		var out bytes.Buffer
		df := &deptfile.File{
			Requirements: []*deptfile.Requirement{},
			Writer:       &out,
		}
		cleanup := cmd.ChangeDeptfileLoad(func() (*deptfile.File, error) {
			return df, nil
		})
		defer cleanup()

		cmd := cmd.NewGet(mockUI, mockFetcher, nil)

		repo := "github.com/ktr0731/go-modules-test"
		code := cmd.Run([]string{repo})
		if code != 1 {
			t.Errorf("Run must be 1, but got %d", code)
		}

		if n := len(mockFetcher.FetchCalls()); n != 1 {
			t.Errorf("Fetch must be called once, but actual %d", n)
		}

		if n := len(df.Requirements); n != 0 {
			t.Errorf("Get must not add passed dependency to requirements when command failed, but %d dependency found", n)
		}

		if out.String() != "" {
			t.Error("Encode must not be called")
		}
	})
}
