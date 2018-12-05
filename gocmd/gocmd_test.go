package gocmd_test

import (
	"context"
	"testing"
	"time"

	"github.com/ktr0731/dept/gocmd"
)

func TestTimeoutErr(t *testing.T) {
	err := &gocmd.TimeoutErr{Command: "foo bar"}
	expected := "command 'foo bar' timed out"
	actual := err.Error()
	if expected != actual {
		t.Errorf("expected %s, but got %s", expected, actual)
	}
}

func TestCommand(t *testing.T) {
	cases := map[string]func(context.Context, gocmd.Command) error{
		"Get":         func(ctx context.Context, cmd gocmd.Command) error { return cmd.Get(ctx, "github.com/ktr0731/dept") },
		"Build":       func(ctx context.Context, cmd gocmd.Command) error { return cmd.Build(ctx) },
		"ModTidy":     func(ctx context.Context, cmd gocmd.Command) error { return cmd.ModTidy(ctx) },
		"ModDownload": func(ctx context.Context, cmd gocmd.Command) error { return cmd.ModDownload(ctx) },
		"List": func(ctx context.Context, cmd gocmd.Command) error {
			_, err := cmd.List(ctx, "github.com/ktr0731/dept")
			return err
		},
	}

	runNormalTest := func(t *testing.T, cmd gocmd.Command, c func(ctx context.Context, cmd gocmd.Command) error) {
		t.Run("normal", func(t *testing.T) {
			ctx := context.Background()
			err := c(ctx, cmd)
			if err != nil {
				t.Fatalf("must not return errors, but got %s", err)
			}
		})
	}

	runCancelTest := func(t *testing.T, cmd gocmd.Command, c func(ctx context.Context, cmd gocmd.Command) error) {
		t.Run("cancel", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			// Do cancel immediately.
			cancel()
			err := c(ctx, cmd)
			if err == nil {
				t.Fatal("must return errors, but nil")
			}
			if err != context.Canceled {
				t.Errorf("must return context.Canceled, but actual %s", err)
			}
		})
	}

	runTimeoutTest := func(t *testing.T, cmd gocmd.Command, c func(ctx context.Context, cmd gocmd.Command) error) {
		t.Run("timeout", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 0*time.Nanosecond)
			// Do cancel immediately.
			cancel()
			err := c(ctx, cmd)
			if err == nil {
				t.Fatal("must return errors, but nil")
			}
			if _, ok := err.(*gocmd.TimeoutErr); !ok {
				t.Errorf("must return *TimeoutErr, but actual %T", err)
			}
		})
	}

	for name, f := range cases {
		t.Run(name, func(t *testing.T) {
			runNormalTest(t, gocmd.New(), f)
			runCancelTest(t, gocmd.New(), f)
			runTimeoutTest(t, gocmd.New(), f)
		})
	}
}
