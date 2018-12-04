//go:generate moq -out mock_gen.go . Command

package gocmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type TimeOutErr struct {
	Command string
}

func (e *TimeOutErr) Error() string {
	return fmt.Sprintf("command '%s' timed out", e.Command)
}

// Command provides available Go commands.
// Each comamnd may return these errors:
//
//   - *TimeOutErr: comamnd timed out
//   - context.Canceled: context is canceled
//   - others: command execution error
//
type Command interface {
	// Get executes 'go get' with args.
	Get(ctx context.Context, args ...string) error
	// Build executes 'go build' with args.
	Build(ctx context.Context, args ...string) error
	// ModTidy executes 'go mod tidy'.
	ModTidy(ctx context.Context) error
	// List executes 'go list' with args.
	// The result is represents as an io.Reader.
	List(ctx context.Context, args ...string) (io.Reader, error)
}

// New returns a new instance of Command.
func New() Command {
	return &command{}
}

type command struct{}

func (c *command) Get(ctx context.Context, args ...string) error {
	return run(ctx, 15*time.Minute, "get", args)
}

func (c *command) Build(ctx context.Context, args ...string) error {
	return run(ctx, 15*time.Minute, "build", args)
}

func (c *command) ModTidy(ctx context.Context) error {
	return run(ctx, 3*time.Minute, "mod", []string{"tidy"})
}

func (c *command) List(ctx context.Context, args ...string) (io.Reader, error) {
	return runWithOutput(ctx, 10*time.Minute, "list", args)
}

func runWithOutput(ctx context.Context, timeout time.Duration, command string, args []string) (io.Reader, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var out, eout bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", append([]string{command}, args...)...)
	cmd.Stdout = &out
	cmd.Stderr = &eout

	return &out, runCommand(ctx, cmd)
}

func run(ctx context.Context, timeout time.Duration, command string, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var eout bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", append([]string{command}, args...)...)
	cmd.Stderr = &eout

	return runCommand(ctx, cmd)
}

func runCommand(ctx context.Context, cmd *exec.Cmd) error {
	err := cmd.Run()
	switch ctx.Err() {
	case context.Canceled:
		return context.Canceled
	case context.DeadlineExceeded:
		return &TimeOutErr{Command: cmd.Path + " " + strings.Join(cmd.Args, " ")}
	default:
	}
	if err != nil {
		return errors.Wrapf(err, "failed to execute '%s %s': '%s'", cmd.Path, strings.Join(cmd.Args, " "), cmd.Stderr.(*bytes.Buffer).String())
	}
	return nil
}
