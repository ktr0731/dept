package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/filegen"
	"github.com/ktr0731/dept/gocmd"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

// getCommand gets a passed Go tool from the remote repository.
// get generate the artifact as follows.
//
//   1. load deptfile.
//   2. generate Go code which imports all required tools.
//   3. run 'go get' with Go modules aware mode to collect dependencies of 2.
//   4. build binaries.
//   5. TODO: Gopkg.toml
//
type getCommand struct {
	f         *flag.FlagSet
	ui        cli.Ui
	gocmd     gocmd.Command
	workspace *deptfile.Workspace
}

func (c *getCommand) UI() cli.Ui {
	return c.ui
}

// Help shows the help message.
// Before call Help, getCommand.f must be initialized.
func (c *getCommand) Help() string {
	return fmt.Sprintf("Usage: dept get <url>\n%s", flagUsage(c.f))
}

func (c *getCommand) Synopsis() string {
	return "Get new CLI tool as a dependency"
}

// Used only mocking
var (
	deptfileLoad = deptfile.Load
)

func (c *getCommand) Run(args []string) int {
	output := c.f.String("o", "", "output name")
	c.f.Parse(args)

	args = c.f.Args()

	return run(c, func() error {
		if len(args) != 1 {
			return errShowHelp
		}

		ctx := context.Background()

		err := c.workspace.Do(func(projRoot string) error {
			df, err := deptfileLoad(ctx)
			if err != nil {
				return err
			}

			repo := args[0]

			if *output == "" {
				*output = filepath.Base(repo)
			}

			// TODO: cleanup imports
			requires := make([]string, 0, len(df.Require))
			for _, r := range df.Require {
				if p := filepath.Base(r.Path); *output == p {
					return errors.Errorf("tool names conflicted: %s and %s. please rename tool name by -o option.", repo, r.Path)
				}
				requires = append(requires, r.Path)
			}
			requires = append(requires, repo)

			f, err := os.Create("tools.go")
			if err != nil {
				return errors.Wrap(err, "failed to create a temp file which contains required Go tools in the import statement")
			}
			defer f.Close()
			filegen.Generate(f, requires)

			if err := c.gocmd.Get(ctx); err != nil {
				return errors.Wrap(err, "failed to get Go tools dependencies")
			}

			binPath := filepath.Join(projRoot, "_tools", *output)
			if err := c.gocmd.Build(ctx, "-o", binPath, repo); err != nil {
				return errors.Wrapf(err, "failed to buld %s (bin path = %s)", repo, binPath)
			}

			return nil
		})
		return err
	})
}

// NewGet returns an initialized get command instance.
func NewGet(
	ui cli.Ui,
	gocmd gocmd.Command,
	workspace *deptfile.Workspace,
) cli.Command {
	return &getCommand{
		f:         flag.NewFlagSet("get", flag.ExitOnError),
		ui:        ui,
		gocmd:     gocmd,
		workspace: workspace,
	}
}
