package cmd

import (
	"os"

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
		if _, err := os.Stat(deptfile.DeptfileName); err == nil {
			return deptfile.ErrAlreadyExist
		}

		f, err := os.Create(deptfile.DeptfileName)
		if err != nil {
			return errors.Wrap(err, "failed to create deptfile")
		}
		defer f.Close()

		df := &deptfile.File{Requirements: []*deptfile.Requirement{}}
		df.Writer = f
		if err := df.Encode(); err != nil {
			return errors.Wrap(err, "failed to write the initial content to deptfile")
		}
		return nil
	})
}

// NewInit returns an initialized init command instance.
func NewInit(ui cli.Ui) cli.Command {
	return &initCommand{ui: ui}
}
