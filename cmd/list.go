package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/ktr0731/dept/deptfile"
	"github.com/mitchellh/cli"
)

// listCommand list ups managed dependencies.
type listCommand struct {
	f         *flag.FlagSet
	ui        cli.Ui
	workspace *deptfile.Workspace
}

func (c *listCommand) UI() cli.Ui {
	return c.ui
}

func (c *listCommand) Help() string {
	return fmt.Sprintf("Usage: dept list\n\n%s", FlagUsage(c.f))
}

func (c *listCommand) Synopsis() string {
	return fmt.Sprintf("List ups all tools based on %s", deptfile.DeptfileName)
}

func (c *listCommand) Run(args []string) int {
	c.f.Parse(args)

	indirect := c.f.Lookup("i").Value.String() == "true"

	args = c.f.Args()

	return run(c, func() error {
		ctx := context.Background()

		err := c.workspace.Do(func(projRoot string) error {
			df, err := deptfileLoad(ctx)
			if err != nil {
				return err
			}

			requires := make([]string, 0, len(df.Require))
			for _, r := range df.Require {
				if !indirect && r.Indirect {
					continue
				}
				requires = append(requires, fmt.Sprintf("%s %s", r.Path, r.Version))
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
	workspace *deptfile.Workspace,
) cli.Command {
	f := flag.NewFlagSet("list", flag.ExitOnError)
	f.Bool("i", false, "Show all packages which including indirection packages")
	return &listCommand{
		f:         f,
		ui:        ui,
		workspace: workspace,
	}
}
