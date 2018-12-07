package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/toolcacher"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

var (
	syscallExec = syscall.Exec
)

type execCommand struct {
	args       []string
	ui         cli.Ui
	workspace  deptfile.Workspacer
	toolcacher toolcacher.Cacher
}

func (c *execCommand) UI() cli.Ui {
	return c.ui
}

func (c *execCommand) Help() string {
	return fmt.Sprintf("Usage: dept exec <tool name>")
}

func (c *execCommand) Synopsis() string {
	return fmt.Sprintf("Execute passed tool")
}

func (c *execCommand) Run([]string) int {
	// In exec command, we don't use Run's args, use c.args instead.
	// The reason is described at app.go.
	args := c.args
	return run(c, func(ctx context.Context) error {
		if len(args) == 0 {
			return errShowHelp
		}

		toolName := args[0]

		var toolPkgName, toolVersion string
		err := c.workspace.Do(func(projRoot string, df *deptfile.File) error {
			for _, r := range df.Require {
				forToolsWithOutputName(r, func(path, outputName string) bool {
					if outputName == toolName || (outputName == "" && filepath.Base(path) == toolName) {
						toolPkgName = path
						toolVersion = r.Version
						return false
					}
					return true
				})
			}
			if toolPkgName == "" || toolVersion == "" {
				return errors.Errorf(`command '%s' is not in %s (available tools can be see 'dept list -f "{{ .Name }}"')`, toolName, deptfile.FileName)
			}

			cachePath, err := c.toolcacher.Get(ctx, toolPkgName, toolVersion)
			if err != nil {
				return errors.Wrap(err, "failed to get a cached tool path")
			}

			err = syscallExec(cachePath, append([]string{toolName}, args[1:]...), os.Environ())
			if err != nil {
				return errors.Wrapf(err, "failed to execute the specified tool: %s", strings.Join(append([]string{toolName}, args[1:]...), " "))
			}
			return nil
		})
		return err
	})
	return 0
}

// NewExec returns an initialized execCommand instance.
func NewExec(
	args []string,
	ui cli.Ui,
	workspace deptfile.Workspacer,
	toolcacher toolcacher.Cacher,
) cli.Command {
	return &execCommand{
		args:       args,
		ui:         ui,
		workspace:  workspace,
		toolcacher: toolcacher,
	}
}
