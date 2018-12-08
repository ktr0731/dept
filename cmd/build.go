package cmd

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/fileutil"
	"github.com/ktr0731/dept/gocmd"
	"github.com/ktr0731/dept/logger"
	"github.com/ktr0731/dept/toolcacher"
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
	f          *buildFlagSet
	ui         cli.Ui
	gocmd      gocmd.Command
	workspace  deptfile.Workspacer
	toolcacher toolcacher.Cacher
}

func (c *buildCommand) UI() cli.Ui {
	return c.ui
}

func (c *buildCommand) Help() string {
	return fmt.Sprintf("Usage: dept build\n\n%s", FlagUsage(c.f.FlagSet, false))
}

func (c *buildCommand) Synopsis() string {
	return "Build all tools and copy these to the specified dir"
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
			tools := []*tool{}
			for _, r := range df.Require {
				forToolsWithOutputName(r, func(path, outputName string) bool {
					requires = append(requires, path)
					tools = append(tools, &tool{Path: path, Name: outputName, Version: r.Version})
					return true
				})
			}

			eg, ctx := errgroup.WithContext(ctx)
			for _, t := range tools {
				t := t
				eg.Go(func() error {
					cachePath, err := c.toolcacher.Get(ctx, t.Path, t.Version)
					if err != nil {
						return errors.Wrapf(err, "failed to get cache of %s", t.Path)
					}
					var outputName string
					if t.Name != "" {
						outputName = t.Name
					} else {
						outputName = filepath.Base(t.Path)
					}
					binPath := filepath.Join(outputDir, outputName)
					logger.Printf("copy %s from %s to %s", t.Path, cachePath, binPath)
					if err := fileutil.Copy(binPath, cachePath); err != nil {
						return errors.Wrapf(err, "failed to copy %s from %s to %s", t.Path, cachePath, binPath)
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
	toolcacher toolcacher.Cacher,
) cli.Command {
	return &buildCommand{
		f:          newBuildFlagSet(),
		ui:         ui,
		gocmd:      gocmd,
		workspace:  workspace,
		toolcacher: toolcacher,
	}
}
