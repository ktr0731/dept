package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/filegen"
	"github.com/ktr0731/dept/gocmd"
	"github.com/ktr0731/dept/logger"
	multierror "github.com/ktr0731/go-multierror"
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
	return "Usage: dept remove <path [path ...]>"
}

func (c *removeCommand) Synopsis() string {
	return fmt.Sprintf("Remove a tool from %s", deptfile.FileName)
}

func (c *removeCommand) Run(args []string) int {
	return run(c, func(ctx context.Context) error {
		if len(args) < 1 {
			return errShowHelp
		}

		err := c.workspace.Do(func(projRoot string, df *deptfile.File) error {
			repoMap := map[string]bool{}
			for _, path := range args {
				repo, _, err := normalizePath(path)
				if err != nil {
					return err
				}
				if _, found := repoMap[repo]; found {
					continue
				}
				repoMap[repo] = false
			}

			requires := make([]string, 0, len(df.Require))
			for _, r := range df.Require {
				forTools(r, func(path string) bool {
					if _, found := repoMap[path]; found {
						repoMap[path] = true
					} else {
						requires = append(requires, path)
					}
					return true
				})
			}
			var err error
			for repo, found := range repoMap {
				if !found {
					err = multierror.Append(err, errors.Errorf("%s not found in gotool.mod", repo))
				}
			}
			if err != nil {
				return err
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
