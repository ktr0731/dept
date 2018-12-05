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
	"github.com/ktr0731/dept/logger"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type buildFlagSet struct {
	*flag.FlagSet

	outputDir string
}

func newBuildFlagSet() *buildFlagSet {
	bf := &buildFlagSet{FlagSet: flag.NewFlagSet("build", flag.ExitOnError)}
	bf.StringVar(&bf.outputDir, "d", "", "Output dir to store built Go tools")
	return bf
}

// buildCommand builds Go tools based on gotool.mod.
type buildCommand struct {
	f         *buildFlagSet
	ui        cli.Ui
	gocmd     gocmd.Command
	workspace deptfile.Workspacer
}

func (c *buildCommand) UI() cli.Ui {
	return c.ui
}

func (c *buildCommand) Help() string {
	return fmt.Sprintf("Usage: dept build\n\n%s", FlagUsage(c.f.FlagSet, false))
}

func (c *buildCommand) Synopsis() string {
	return fmt.Sprintf("Build all tools based on %s", deptfile.FileName)
}

func (c *buildCommand) Run(args []string) int {
	if err := c.f.Parse(args); err != nil {
		c.UI().Error(err.Error())
		return 1
	}

	outputDir := c.f.outputDir
	if outputDir != "" {
		outputDir, _ = filepath.Abs(outputDir)
	}

	return run(c, func(ctx context.Context) error {
		err := c.workspace.Do(func(projRoot string, df *deptfile.File) error {
			outputDir = resolveOutputDir(projRoot, outputDir)

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
					binPath := filepath.Join(outputDir, outputName)
					logger.Printf("building %s to %s", t.path, binPath)
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
	return &buildCommand{
		f:         newBuildFlagSet(),
		ui:        ui,
		gocmd:     gocmd,
		workspace: workspace,
	}
}
