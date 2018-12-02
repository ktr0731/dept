package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/filegen"
	"github.com/ktr0731/dept/gocmd"
	"github.com/ktr0731/dept/logger"
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
	return fmt.Sprintf("Remove a tool from %s", deptfile.FileName)
}

func (c *removeCommand) Run(args []string) int {
	return run(c, func() error {
		if len(args) != 1 {
			return errShowHelp
		}

		ctx := context.Background()

		err := c.workspace.Do(func(projRoot string, df *deptfile.File) error {
			path := args[0]
			repo, _, err := normalizePath(path)
			if err != nil {
				return err
			}

			var all, cnt int
			requires := make([]string, 0, len(df.Require))
			for _, r := range df.Require {
				forTools(r, func(path string) bool {
					all++
					if repo != path {
						requires = append(requires, path)
						cnt++
					}
					return true
				})
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

			logger.Println("removing unnecessary tools and indirection dependencies")
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
