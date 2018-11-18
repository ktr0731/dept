//go:generate moq -out mock_gen.go . Command

package gocmd

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type Command interface {
	Get(ctx context.Context, args ...string) error
	Build(ctx context.Context, args ...string) error
	ModTidy(ctx context.Context) error
}

func New() Command {
	return &command{}
}

type command struct{}

func (c *command) Get(ctx context.Context, args ...string) error {
	var eout bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", append([]string{"get"}, args...)...)
	cmd.Stderr = &eout
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to execute 'go get %s': '%s'", strings.Join(args, " "), eout.String())
	}
	return nil
}

func (c *command) Build(ctx context.Context, args ...string) error {
	var eout bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", append([]string{"build"}, args...)...)
	cmd.Stderr = &eout
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to execute 'go build %s': '%s'", strings.Join(args, " "), eout.String())
	}
	return nil
}

func (c *command) ModTidy(ctx context.Context) error {
	var eout bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Stderr = &eout
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to execute 'go mod tidy': '%s'", eout.String())
	}
	return nil
}
