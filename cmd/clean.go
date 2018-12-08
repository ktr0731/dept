package cmd

import (
	"fmt"

	"github.com/ktr0731/dept/toolcacher"
	"github.com/mitchellh/cli"
)

// cleanCommand cleans up all cached binaries.
type cleanCommand struct {
	ui         cli.Ui
	toolcacher toolcacher.Cacher
}

func (c *cleanCommand) UI() cli.Ui {
	return c.ui
}

func (c *cleanCommand) Help() string {
	return fmt.Sprint("Usage: dept clean")
}

func (c *cleanCommand) Synopsis() string {
	return fmt.Sprint("Cleans up all cached tools")
}

func (c *cleanCommand) Run(args []string) int {
	return run(c, c.toolcacher.Clear)
}

// NewClean returns an initialized cleanCommand instance.
func NewClean(
	ui cli.Ui,
	toolcacher toolcacher.Cacher,
) cli.Command {
	return &cleanCommand{
		ui:         ui,
		toolcacher: toolcacher,
	}
}
