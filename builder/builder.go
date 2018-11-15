//go:generate moq -out mock.go . Builder

package builder

import (
	"context"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// Builder builds the artifact from a passed directory
// using 'go build' with Go modules aware mode.
// dir must be exist and main package.
type Builder interface {
	Build(ctx context.Context, dir string) error
}

type builder struct{}

func (b *builder) Build(ctx context.Context, dir string) error {
	// TODO: use the project root
	os.Chdir(dir)
	if err := exec.CommandContext(ctx, "go", "build").Run(); err != nil {
		return errors.Wrapf(err, "failed to build %s repository", dir)
	}
	return nil
}

func New() Builder {
	return &builder{}
}
