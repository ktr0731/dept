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
	"github.com/ktr0731/dept/logger"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type outputFlagValue struct {
	Values []struct{ Out, Path string }
	f      *flag.FlagSet
}

func (v *outputFlagValue) Set(out string) error {
	path := v.f.Arg(0)
	if v.f.NArg() < 1 || (len(out) > 0 && out[0] == '-') || (len(path) > 0 && path[0] == '-') {
		return errShowHelp
	}
	v.Values = append(v.Values, struct{ Out, Path string }{out, path})
	return v.f.Parse(v.f.Args()[1:])
}

func (v *outputFlagValue) String() string {
	s := make([]string, 0, len(v.Values))
	for _, v := range v.Values {
		s = append(s, fmt.Sprintf("%s=%s", v.Path, v.Out))
	}
	return strings.Join(s, ", ")
}

type getFlagSet struct {
	*flag.FlagSet

	outputDir   string
	update      bool
	outputNames *outputFlagValue
}

func newGetFlagSet() *getFlagSet {
	gf := &getFlagSet{FlagSet: flag.NewFlagSet("get", flag.ContinueOnError)}

	// Suppress outputting by flag, delegate to cli.Command instead.
	gf.SetOutput(ioutil.Discard)
	gf.StringVar(&gf.outputDir, "d", "", "Output dir to store built Go tools")
	gf.BoolVar(&gf.update, "u", false, "Update the specified tool to the latest version")

	gf.outputNames = &outputFlagValue{Values: []struct{ Out, Path string }{}, f: gf.FlagSet}
	gf.Var(gf.outputNames, "o", "Output name (first arg is output name, second arg is path)")

	return gf
}

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
	ui        cli.Ui
	gocmd     gocmd.Command
	workspace deptfile.Workspacer

	f *getFlagSet
}

func (c *getCommand) UI() cli.Ui {
	return c.ui
}

var getHelpTmpl = `Usage: dept get <path [path ...]>

get installs the passed Go tools to the specified directory.
If $GOBIN enabled, it will be used preferentially.
-u flag updates the passed Go tools. If there are no args,
updates all Go tools which is already installed.

%s
%s
Examples:

    $ dept get github.com/mitchellh/gox
    $ dept get github.com/mitchellh/gox@v0.3.0
    $ dept get github.com/mitchellh/gox@v0.1.0

    $ dept get -o it github.com/ktr0731/itunes-cli/itunes

    $ dept get -d bin github.com/mitchellh/gox
    $ GOBIN=$PWD/bin dept get github.com/mitchellh/gox
`

// Help shows the help message.
// Before call Help, getCommand.f must be initialized.
func (c *getCommand) Help() string {
	return fmt.Sprintf(
		getHelpTmpl,
		ExcludeFlagUsage(c.f.FlagSet, false, []string{"o"}),
		ExcludeFlagUsage(c.f.FlagSet, true, []string{"d", "u"}))
}

func (c *getCommand) Synopsis() string {
	return "Get a new tool as a dependency"
}

func (c *getCommand) Run(args []string) int {
	if err := c.f.Parse(args); err != nil {
		c.UI().Error(err.Error())
		return 1
	}
	args = c.f.Args()

	outputDir := c.f.outputDir
	if outputDir != "" {
		outputDir, _ = filepath.Abs(outputDir)
	}
	update := c.f.update

	return run(c, func(ctx context.Context) error {
		err := c.workspace.Do(func(projRoot string, df *deptfile.File) error {
			paths, err := c.initModPaths(ctx, c.f.outputNames.Values, args)
			if err != nil {
				return err
			}

			if len(paths) == 0 && !update {
				return errShowHelp
			}

			cleanup, err := generateGoFile(df, paths)
			if err != nil {
				return err
			}
			defer cleanup()

			if len(paths) == 0 && update {
				logger.Println("updating all tools to the latest version")
				if err := c.gocmd.Get(ctx, "-u", "-d"); err != nil {
					return errors.Wrap(err, "failed to update Go tools")
				}
				return nil
			}

			// Always getCommand runs Get.
			// If an unmanaged tool is passed with -u option, '// indirect' will be marked
			// because it is not included in gotool.mod.
			getArgs := make([]string, 0, 1+len(paths))
			getArgs = append(getArgs, "-d")
			for _, p := range paths {
				getArgs = append(getArgs, p.modPath())
			}
			logger.Println("getting all dependencies")
			if err := c.gocmd.Get(ctx, append(getArgs, ".")...); err != nil {
				return errors.Wrap(err, "failed to get Go tools dependencies")
			}

			eg, ctx := errgroup.WithContext(ctx)
			outputDir = resolveOutputDir(projRoot, outputDir)
			for _, path := range paths {
				path := path
				eg.Go(func() error {
					// If also -u is passed, update Repo to the latest.
					if update && path.Ver == "" {
						logger.Printf("updating %s to the latest version", path.Repo)
						if err := c.gocmd.Get(ctx, "-u", "-d", path.Repo); err != nil {
							return errors.Wrap(err, "failed to get Go tools dependencies")
						}
					}

					var binPath string
					if path.Out != "" {
						binPath = filepath.Join(outputDir, path.Out)
					} else {
						binPath = filepath.Join(outputDir, filepath.Base(path.Repo))
					}
					logger.Printf("building %s to %s", path.Repo, binPath)
					if err := c.gocmd.Build(ctx, "-o", binPath, path.Repo); err != nil {
						return errors.Wrapf(err, "failed to buld %s (bin path = %s)", path.Repo, binPath)
					}

					return nil
				})
			}
			if err := eg.Wait(); errors.Cause(err) == context.Canceled {
				return context.Canceled
			} else if err != nil {
				return errors.Wrap(err, "failed to build tools")
			}

			return nil
		})
		return err
	})
}

// initModPaths parses passed paths and collect its module roots.
// initModPaths must be call inside of a workspace.
func (c *getCommand) initModPaths(ctx context.Context, argsWithFlag []struct{ Out, Path string }, args []string) ([]*path, error) {
	found := map[string]interface{}{}

	paths := make([]*path, 0, len(argsWithFlag)+len(args))
	if len(argsWithFlag) == 0 && len(args) == 0 {
		return nil, nil
	}

	eg, ctx := errgroup.WithContext(ctx)

	appendPath := func(p, out string) error {
		if len(p) > 0 && p[0] == '-' {
			return errors.Errorf("found '%s' after args. all flags must be put before args", p)
		}
		repo, ver, err := normalizePath(p)
		if err != nil {
			return err
		}
		if _, ok := found[repo]; ok {
			return nil
		}
		path := &path{Val: p, Repo: repo, Ver: ver, Out: out}
		paths = append(paths, path)
		eg.Go(func() (err error) {
			path.ModRoot, err = getModuleRoot(ctx, c.gocmd, path.Repo)
			return
		})
		return nil
	}

	for _, a := range argsWithFlag {
		if err := appendPath(a.Path, a.Out); err != nil {
			return nil, err
		}
	}
	for _, path := range args {
		if err := appendPath(path, ""); err != nil {
			return nil, err
		}
	}
	return paths, eg.Wait()
}

// generateGoFile generate a Go file which imports df.Require and paths.
// File name is always "tools.go", also package name is "tools".
// Returned func is a cleanup function.
func generateGoFile(df *deptfile.File, paths []*path) (func(), error) {
	importPaths := make([]string, 0, len(df.Require))
	for _, path := range paths {
		var targetReq *deptfile.Require
		for _, r := range df.Require {
			if r.Path == path.ModRoot {
				tmp, err := copystructure.Copy(r)
				if err != nil {
					return nil, errors.Wrap(err, "failed to deepcopy a Require")
				}
				targetReq = tmp.(*deptfile.Require)
			}
			var err error
			var i int
			forTools(r, func(importPath string) bool {
				importPaths = append(importPaths, importPath)
				if toolNameConflicted(importPath, path.Repo) {
					err = errors.Errorf("tool names conflicted: %s and %s. please rename tool name by -o option.", path.Repo, importPath)
					return false
				}
				// If -o passed with updating, rename tool to it.
				if importPath == path.Repo && path.Out != "" {
					r.ToolPaths[i].Name = path.Out
				}
				i++
				return true
			})
			if err != nil {
				return nil, err
			}
		}

		importPaths = append(importPaths, path.Repo)

		df.Require = appendRequire(df.Require, targetReq, path)
	}

	f, err := os.Create("tools.go")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a temp file which contains required Go tools in the import statement")
	}
	defer f.Close()
	filegen.Generate(f, importPaths)

	return func() {
		if err := recover(); err != nil {
			os.Remove("tools.go")
			panic(err)
		}
		os.Remove("tools.go")
	}, nil
}

func getModuleRoot(ctx context.Context, gocmd gocmd.Command, path string) (string, error) {
	logger.Printf("get the module root of %s", path)
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
// If r is nil, appendRequire assigns a new one with Path=ModRoot.
// If r is nil, it means the module of r is not managed yet.
//
// uri is the full path for tool r.
// ModRoot is the module root of r.
func appendRequire(reqs []*deptfile.Require, r *deptfile.Require, path *path) []*deptfile.Require {
	toolPath := strings.TrimPrefix(path.Repo, path.ModRoot)
	// a new module
	if r == nil {
		r = &deptfile.Require{Path: path.ModRoot}
		var t *deptfile.Tool
		// tool is not in the module root.
		if toolPath != "" {
			t = &deptfile.Tool{Path: toolPath, Name: path.Out}
		} else {
			t = &deptfile.Tool{Path: "/", Name: path.Out}
		}
		r.ToolPaths = append(r.ToolPaths, t)
		return append(reqs, r)
	}

	var t *deptfile.Tool
	if toolPath != "" {
		t = &deptfile.Tool{Path: toolPath, Name: path.Out}
	} else {
		t = &deptfile.Tool{Path: "/", Name: path.Out}
		toolPath = "/"
	}
	var duplicated bool
	for _, another := range r.ToolPaths {
		if another.Path == toolPath {
			duplicated = true
			break
		}
	}
	if duplicated {
		return reqs
	}
	r.ToolPaths = append(r.ToolPaths, t)

	sort.Slice(r.ToolPaths, func(i, j int) bool {
		return len(r.ToolPaths[i].Path) < len(r.ToolPaths[j].Path)
	})
	for i := range reqs {
		if reqs[i].Path == path.ModRoot {
			reqs[i] = r
			return reqs
		}
	}
	panic("must not reach to here")
}

type path struct {
	// Val is the original value of path.
	// For example, 'github.com/ktr0731/salias@v0.1.0'
	Val string
	// ModRoot is the module root of path without the version.
	// For example, 'github.com/ktr0731/salias'
	ModRoot string
	// Repo is the repository name of Val.
	// For example, 'github.com/ktr0731/salias'
	Repo string
	// Ver is the version of Val.
	// For example, '@v0.1.0'
	Ver string
	// Out is the output name of the tool specified by path.
	// For example, 'salias'
	// If Out is empty, it means Out is same as filepath.Base(Repo).
	Out string
}

// modPath returns the completely module path which includes module's version.
func (p *path) modPath() string {
	if p.Ver != "" {
		return p.ModRoot + "@" + p.Ver
	}
	return p.ModRoot
}

// NewGet returns an initialized get command instance.
func NewGet(
	ui cli.Ui,
	gocmd gocmd.Command,
	workspace deptfile.Workspacer,
) cli.Command {
	return &getCommand{
		f:         newGetFlagSet(),
		ui:        ui,
		gocmd:     gocmd,
		workspace: workspace,
	}
}
