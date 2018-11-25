package cmd

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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
	workspace deptfile.Workspacer
}

func (c *getCommand) UI() cli.Ui {
	return c.ui
}

// Help shows the help message.
// Before call Help, getCommand.f must be initialized.
func (c *getCommand) Help() string {
	return fmt.Sprintf("Usage: dept get <package>\n\n%s", FlagUsage(c.f))
}

func (c *getCommand) Synopsis() string {
	return "Get a new tool as a dependency"
}

func (c *getCommand) Run(args []string) int {
	c.f.Parse(args)

	output := c.f.Lookup("o").Value.String()
	outputDir := c.f.Lookup("d").Value.String()
	update := c.f.Lookup("u").Value.String() == "true"

	args = c.f.Args()

	return run(c, func() error {
		if len(args) != 1 {
			return errShowHelp
		}

		ctx := context.Background()

		err := c.workspace.Do(func(projRoot string, df *deptfile.GoMod) error {
			path := args[0]
			repo, ver, err := normalizeRepo(path)
			if err != nil {
				return err
			}

			if output == "" {
				output = filepath.Base(repo)
			}

			// key: tool name, val: full path for the tool.
			foundTool := map[string]string{}
			requireMap := map[string]*deptfile.Require{}
			requires := make([]string, 0, len(df.Require))
			for _, r := range df.Require {
				requireMap[r.Path] = r
				if r.CommandPath != nil {
					for _, cmdPath := range r.CommandPath {
						requires = append(requires, r.Path+cmdPath)
						foundTool[filepath.Base(cmdPath)] = r.Path + cmdPath
					}
				} else {
					requires = append(requires, r.Path)
					foundTool[filepath.Base(r.Path)] = r.Path
				}
			}

			if path, ok := foundTool[output]; ok && repo != path {
				return errors.Errorf("tool names conflicted: %s and %s. please rename tool name by -o option.", repo, path)
			}
			requires = append(requires, repo)
			mroot, err := getModuleRoot(ctx, c.gocmd, repo)
			if err != nil {
				return err
			}

			// TODO: multi packages support
			var r *deptfile.Require
			cmdPath := strings.TrimPrefix(repo, mroot)
			if req, ok := requireMap[mroot]; ok {
				if req.CommandPath == nil {
					req.CommandPath = []string{cmdPath}
				} else {
					req.CommandPath = append(req.CommandPath, cmdPath)
				}
				r = req
			} else {
				r = &deptfile.Require{
					Path:        mroot,
					CommandPath: []string{cmdPath},
				}
			}
			df.Require = append(df.Require, r)

			f, err := os.Create("tools.go")
			if err != nil {
				return errors.Wrap(err, "failed to create a temp file which contains required Go tools in the import statement")
			}
			defer os.Remove("tools.go")
			defer f.Close()
			filegen.Generate(f, requires)

			// Always getCommand runs Get.
			// If an unmanaged tool is passed with -u option, '// indirect' is marked
			// because it is not included in gotool.mod.
			if err := c.gocmd.Get(ctx, path, "./..."); err != nil {
				return errors.Wrap(err, "failed to get Go tools dependencies")
			}

			// If also -u is passed, update repo to the latest.
			if update && ver == "" {
				if err := c.gocmd.Get(ctx, "-u", repo); err != nil {
					return errors.Wrap(err, "failed to get Go tools dependencies")
				}
			}

			binPath := filepath.Join(projRoot, outputDir, output)
			if err := c.gocmd.Build(ctx, "-o", binPath, repo); err != nil {
				return errors.Wrapf(err, "failed to buld %s (bin path = %s)", repo, binPath)
			}

			return nil
		})
		return err
	})
}

func getModuleRoot(ctx context.Context, gocmd gocmd.Command, path string) (string, error) {
	res, err := gocmd.List(ctx, "-f", `{{ .Module.Path }}`, path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get the module root of %s", path)
	}
	b, err := ioutil.ReadAll(res)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert io.Reader to string")
	}
	return strings.TrimSpace(string(b)), nil
}

// NewGet returns an initialized get command instance.
func NewGet(
	ui cli.Ui,
	gocmd gocmd.Command,
	workspace deptfile.Workspacer,
) cli.Command {
	f := flag.NewFlagSet("get", flag.ExitOnError)
	f.String("o", "", "Output name")
	f.String("d", "_tools", "Output dir to store built Go tools")
	f.Bool("u", false, "Update the specified tool to the latest version")
	return &getCommand{
		f:         f,
		ui:        ui,
		gocmd:     gocmd,
		workspace: workspace,
	}
}
