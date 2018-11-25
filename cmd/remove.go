package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/filegen"
	"github.com/ktr0731/dept/gocmd"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

// removeCommand removes a passed Go tool from gotool.mod as follows.
//
//   1. load deptfile.
//   2. check whether the passed tool is vendored, or not.
//   3. generate Go code from updated deptfile.
//   4. run 'go mod tidy' to remove unnecessary dependencies.
//
type removeCommand struct {
	ui        cli.Ui
	gocmd     gocmd.Command
	workspace deptfile.Workspacer
}

func (c *removeCommand) UI() cli.Ui {
	return c.ui
}

func (c *removeCommand) Help() string {
	return "Usage: dept remove <package>"
}

func (c *removeCommand) Synopsis() string {
	return fmt.Sprintf("Remove a tool from %s", deptfile.DeptfileName)
}

func (c *removeCommand) Run(args []string) int {
	return run(c, func() error {
		if len(args) != 1 {
			return errShowHelp
		}

		ctx := context.Background()

		err := c.workspace.Do(func(projRoot string, df *deptfile.GoMod) error {
			path := args[0]
			repo, _, err := normalizeRepo(path)
			if err != nil {
				return err
			}

			var all, cnt int
			requires := make([]string, 0, len(df.Require))
			for _, r := range df.Require {
				if len(r.CommandPath) == 0 {
					all++
					if repo != r.Path {
						requires = append(requires, r.Path)
						cnt++
					}
				} else {
					for _, cmdPath := range r.CommandPath {
						all++
						p := r.Path + cmdPath
						if repo != p {
							requires = append(requires, r.Path+cmdPath)
							cnt++
						}
					}
				}
			}
			if all == cnt {
				return errors.Errorf("%s not found in gotool.mod", repo)
			}

			f, err := os.Create("tools.go")
			if err != nil {
				return errors.Wrap(err, "failed to create a temp file which contains required Go tools in the import statement")
			}
			defer os.Remove("tools.go")
			defer f.Close()
			filegen.Generate(f, requires)

			if err := c.gocmd.ModTidy(ctx); err != nil {
				return errors.Wrap(err, "failed to remove the tool from gotool.mod")
			}

			return nil
		})
		return err
	})
}

// NewRemove returns an initialized removeCommand instance.
func NewRemove(
	ui cli.Ui,
	gocmd gocmd.Command,
	workspace deptfile.Workspacer,
) cli.Command {
	return &removeCommand{
		ui:        ui,
		gocmd:     gocmd,
		workspace: workspace,
	}
}
