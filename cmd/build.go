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
	"golang.org/x/sync/errgroup"
)

// buildCommand builds Go tools based on gotool.mod.
type buildCommand struct {
	f         *flag.FlagSet
	ui        cli.Ui
	gocmd     gocmd.Command
	workspace *deptfile.Workspace
}

func (c *buildCommand) UI() cli.Ui {
	return c.ui
}

func (c *buildCommand) Help() string {
	return fmt.Sprintf("Usage: dept build\n\n%s", FlagUsage(c.f))
}

func (c *buildCommand) Synopsis() string {
	return fmt.Sprintf("Build all tools based on %s", deptfile.DeptfileName)
}

func (c *buildCommand) Run(args []string) int {
	c.f.Parse(args)

	outputDir := c.f.Lookup("d").Value.String()

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
				if r.Indirect {
					continue
				}
				requires = append(requires, r.Path)
			}

			f, err := os.Create("tools.go")
			if err != nil {
				return errors.Wrap(err, "failed to create a temp file which contains required Go tools in the import statement")
			}
			defer f.Close()
			filegen.Generate(f, requires)

			eg, ctx := errgroup.WithContext(ctx)
			for _, r := range requires {
				r := r
				eg.Go(func() error {
					binPath := filepath.Join(projRoot, outputDir, filepath.Base(r))
					if err := c.gocmd.Build(ctx, "-o", binPath, r); err != nil {
						return errors.Wrapf(err, "failed to buld %s (bin path = %s)", r, binPath)
					}
					return nil
				})
			}
			return eg.Wait()
		})
		return err
	})
}

// NewBuild returns an initialized buildCommand instance.
func NewBuild(
	ui cli.Ui,
	gocmd gocmd.Command,
	workspace *deptfile.Workspace,
) cli.Command {
	f := flag.NewFlagSet("build", flag.ExitOnError)
	f.String("d", "_tools", "Output dir to store built Go tools")
	return &buildCommand{
		f:         f,
		ui:        ui,
		gocmd:     gocmd,
		workspace: workspace,
	}
}
