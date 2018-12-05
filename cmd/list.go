package cmd

import (
	"context"
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
	return fmt.Sprintf("Usage: dept list <path [path ...]>")
}

func (c *listCommand) Synopsis() string {
	return fmt.Sprintf("List ups all tools based on %s", deptfile.FileName)
}

func (c *listCommand) Run(args []string) int {
	passed := map[string]interface{}{}
	for _, arg := range args {
		passed[arg] = nil
	}
	listAll := len(passed) == 0
	return run(c, func(context.Context) error {
		err := c.workspace.Do(func(projRoot string, df *deptfile.File) error {
			requires := make([]string, 0, len(df.Require))
			for _, r := range df.Require {
				if !listAll {
					// If module roots passed, filter by that modules.
					if _, found := passed[r.Path]; found {
						forTools(r, func(path string) bool {
							requires = append(requires, fmt.Sprintf("%s %s", path, r.Version))
							return true
						})
						continue
					}

					// If module roots not found, step into each tool.
				}

				forTools(r, func(path string) bool {
					if listAll {
						requires = append(requires, fmt.Sprintf("%s %s", path, r.Version))
					} else if _, found := passed[path]; found {
						requires = append(requires, fmt.Sprintf("%s %s", path, r.Version))
					}
					return true
				})
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
