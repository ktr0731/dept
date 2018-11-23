package cmd

import (
	"fmt"
	"strings"

	"github.com/ktr0731/dept/deptfile"
	"github.com/mitchellh/cli"
)

// listCommand list ups managed dependencies.
type listCommand struct {
	ui        cli.Ui
	workspace deptfile.Workspacer
}

func (c *listCommand) UI() cli.Ui {
	return c.ui
}

func (c *listCommand) Help() string {
	return fmt.Sprintf("Usage: dept list")
}

func (c *listCommand) Synopsis() string {
	return fmt.Sprintf("List ups all tools based on %s", deptfile.DeptfileName)
}

func (c *listCommand) Run(args []string) int {
	return run(c, func() error {
		err := c.workspace.Do(func(projRoot string, df *deptfile.GoMod) error {
			requires := make([]string, 0, len(df.Require))
			for _, r := range df.Require {
				if r.CommandPath != nil {
					for _, cmdPath := range r.CommandPath {
						requires = append(requires, fmt.Sprintf("%s %s", r.Path+cmdPath, r.Version))
					}
				} else {
					requires = append(requires, r.Path)
				}
			}

			c.ui.Output(strings.Join(requires, "\n"))
			return nil
		})
		return err
	})
}

// NewList returns an initialized listCommand instance.
func NewList(
	ui cli.Ui,
	workspace deptfile.Workspacer,
) cli.Command {
	return &listCommand{
		ui:        ui,
		workspace: workspace,
	}
}
