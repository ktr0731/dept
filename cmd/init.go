package cmd

import (
	"context"

	"github.com/ktr0731/dept/deptfile"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

// initCommand create a new deptfile.
type initCommand struct {
	ui cli.Ui
}

func (c *initCommand) UI() cli.Ui {
	return c.ui
}

func (c *initCommand) Help() string {
	return "Usage: dept init"
}

func (c *initCommand) Synopsis() string {
	return "Create new deptfile"
}

func (c *initCommand) Run(args []string) int {
	return run(c, func() error {
		if err := deptfile.Create(context.Background()); err != nil {
			return errors.Wrap(err, "failed to create a new deptfile")
		}
		return nil
	})
}

// NewInit returns an initialized init command instance.
func NewInit(ui cli.Ui) cli.Command {
	return &initCommand{ui: ui}
}
