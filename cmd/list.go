package cmd

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/ktr0731/dept/deptfile"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

type listFlagSet struct {
	*flag.FlagSet

	format string
}

func newListFlagSet() *listFlagSet {
	lf := &listFlagSet{FlagSet: flag.NewFlagSet("list", flag.ExitOnError)}
	lf.StringVar(&lf.format, "f", "{{.Path}} {{.Name}} {{.Version}}", "output format")
	return lf
}

// listCommand lists up managed dependencies.
type listCommand struct {
	f         *listFlagSet
	ui        cli.Ui
	workspace deptfile.Workspacer
}

func (c *listCommand) UI() cli.Ui {
	return c.ui
}

var listHelpTmpl = `Usage: dept list <path [path ...]>

list lists up tool information with some attributes.
-f formats output based on the passed format string.
Each item is represents as the following structure.

type tool struct {
	Path, Name, Version string
}

%s`

func (c *listCommand) Help() string {
	return fmt.Sprintf(listHelpTmpl, FlagUsage(c.f.FlagSet, false))
}

func (c *listCommand) Synopsis() string {
	return fmt.Sprintf("Lists up all tools based on %s", deptfile.FileName)
}

func (c *listCommand) Run(args []string) int {
	if err := c.f.Parse(args); err != nil {
		c.UI().Error(err.Error())
		return 1
	}

	args = c.f.Args()

	passed := map[string]interface{}{}
	for _, arg := range args {
		passed[arg] = nil
	}
	listAll := len(passed) == 0
	tmpl := `{{range .}}` + c.f.format + `{{"\n"}}{{end}}`
	return run(c, func(context.Context) error {
		t, err := template.New("list").Parse(tmpl)
		if err != nil {
			return errors.Wrapf(err, "failed to parse -f value '%s'", c.f.format)
		}
		err = c.workspace.Do(func(projRoot string, df *deptfile.File) error {
			requires := make([]*tool, 0, len(df.Require))
			for _, r := range df.Require {
				if !listAll {
					// If module roots passed, filter by that modules.
					if _, found := passed[r.Path]; found {
						forToolsWithOutputName(r, func(path, out string) bool {
							requires = appendListItem(requires, path, out, r.Version)
							return true
						})
						continue
					}

					// If module roots not found, step into each tool.
				}

				forToolsWithOutputName(r, func(path, out string) bool {
					if listAll {
						requires = appendListItem(requires, path, out, r.Version)
					} else if _, found := passed[path]; found {
						requires = appendListItem(requires, path, out, r.Version)
					}
					return true
				})
			}

			var buf bytes.Buffer
			if err := t.Execute(&buf, requires); err != nil {
				return err
			}
			if buf.Len() > 0 {
				// Trim last '\n'
				c.ui.Output(buf.String()[:buf.Len()-1])
			}
			return nil
		})
		return err
	})
}

func appendListItem(requires []*tool, path, out, version string) []*tool {
	t := &tool{Path: path, Name: out, Version: version}
	if out == "" {
		t.Name = filepath.Base(path)
	}
	return append(requires, t)
}

// NewList returns an initialized listCommand instance.
func NewList(
	ui cli.Ui,
	workspace deptfile.Workspacer,
) cli.Command {
	return &listCommand{
		f:         newListFlagSet(),
		ui:        ui,
		workspace: workspace,
	}
}
