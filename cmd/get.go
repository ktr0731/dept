package cmd

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/filegen"
	"github.com/ktr0731/dept/gocmd"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/copystructure"
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

	outputName := c.f.Lookup("o").Value.String()
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
			repo, ver, err := normalizePath(path)
			if err != nil {
				return err
			}

			if outputName == "" {
				outputName = filepath.Base(repo)
			}

			modRoot, err := getModuleRoot(ctx, c.gocmd, repo)
			if err != nil {
				return err
			}

			var targetReq *deptfile.Require
			importPaths := make([]string, 0, len(df.Require))
			for _, r := range df.Require {
				if r.Path == modRoot {
					tmp, err := copystructure.Copy(r)
					if err != nil {
						return errors.Wrap(err, "failed to deepcopy a Require")
					}
					targetReq = tmp.(*deptfile.Require)
				}
				var err error
				forTools(r, func(path string) bool {
					importPaths = append(importPaths, path)
					if toolNameConflicted(r.Path, repo) {
						err = errors.Errorf("tool names conflicted: %s and %s. please rename tool name by -o option.", repo, path)
						return false
					}
					return true
				})
				if err != nil {
					return err
				}
			}

			importPaths = append(importPaths, repo)

			// TODO: multi packages support
			df.Require = appendRequire(df.Require, targetReq, modRoot, repo)

			f, err := os.Create("tools.go")
			if err != nil {
				return errors.Wrap(err, "failed to create a temp file which contains required Go tools in the import statement")
			}
			defer os.Remove("tools.go")
			defer f.Close()
			filegen.Generate(f, importPaths)

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

			binPath := filepath.Join(projRoot, outputDir, outputName)
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

// toolNameConflicted returns whether each tool in p1 and p2 conflicted.
// Note that filepath.Base(p) is the tool name.
//
// If p1 and p2 are the same value, it will be regarded as not conflicted.
func toolNameConflicted(p1, p2 string) bool {
	if p1 == p2 {
		return false
	}
	c1, c2 := filepath.Base(p1), filepath.Base(p2)
	return c1 == c2
}

// appendRequire appends a tool requirment that named r to reqs.
// If r is nil, appendRequire assigns a new one with Path=modRoot.
// If r is nil, it means the module of r is not managed yet.
//
// uri is the full path for tool r.
// modRoot is the module root of r.
func appendRequire(reqs []*deptfile.Require, r *deptfile.Require, modRoot, uri string) []*deptfile.Require {
	path := strings.TrimPrefix(uri, modRoot)
	if r == nil {
		r = &deptfile.Require{Path: modRoot}
		if path != "" {
			r.CommandPath = []string{path}
		}
		return append(reqs, r)
	}
	if path != "" {
		if r.CommandPath == nil {
			// A module already has a tool in the module root package.
			r.CommandPath = []string{"/", path}
		} else {
			if len(r.CommandPath) == 0 {
				r.CommandPath = append(r.CommandPath, path)
			} else {
				var duplicated bool
				forTools(r, func(another string) bool {
					if another == uri {
						duplicated = true
						return false
					}
					return true
				})
				if duplicated {
					return reqs
				}
				r.CommandPath = append(r.CommandPath, path)
			}
		}
	} else {
		if r.CommandPath != nil {
			r.CommandPath = append(r.CommandPath, "/")
		}
		// If r.CommandPath is nil,
		// it means that getCommand launched with '-u' option.
	}
	sort.Slice(r.CommandPath, func(i, j int) bool {
		return len(r.CommandPath[i]) < len(r.CommandPath[j])
	})
	for i := range reqs {
		if reqs[i].Path == modRoot {
			reqs[i] = r
			return reqs
		}
	}
	panic("must not reach to here")
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
