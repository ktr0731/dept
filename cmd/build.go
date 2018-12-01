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
	workspace deptfile.Workspacer
}

func (c *buildCommand) UI() cli.Ui {
	return c.ui
}

func (c *buildCommand) Help() string {
	return fmt.Sprintf("Usage: dept build\n\n%s", FlagUsage(c.f, false))
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

		err := c.workspace.Do(func(projRoot string, df *deptfile.GoMod) error {
			requires := make([]string, 0, len(df.Require))
			tools := []struct {
				path, outputName string
			}{}
			for _, r := range df.Require {
				forToolsWithOutputName(r, func(path, outputName string) bool {
					requires = append(requires, path)
					tools = append(tools, struct{ path, outputName string }{path, outputName})
					return true
				})
			}

			f, err := os.Create("tools.go")
			if err != nil {
				return errors.Wrap(err, "failed to create a temp file which contains required Go tools in the import statement")
			}
			defer os.Remove("tools.go")
			defer f.Close()
			filegen.Generate(f, requires)

			eg, ctx := errgroup.WithContext(ctx)
			for _, t := range tools {
				t := t
				eg.Go(func() error {
					var outputName string
					if t.outputName != "" {
						outputName = t.outputName
					} else {
						outputName = filepath.Base(t.path)
					}
					binPath := filepath.Join(projRoot, outputDir, outputName)
					if err := c.gocmd.Build(ctx, "-o", binPath, t.path); err != nil {
						return errors.Wrapf(err, "failed to buld %s (bin path = %s)", t.path, binPath)
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
	workspace deptfile.Workspacer,
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
