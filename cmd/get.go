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
	"golang.org/x/sync/errgroup"
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
	rf        repeatableFlagSet
	ui        cli.Ui
	gocmd     gocmd.Command
	workspace deptfile.Workspacer
}

func (c *getCommand) UI() cli.Ui {
	return c.ui
}

var getHelpTmpl = `Usage: dept get <package>

%s
%s`

// Help shows the help message.
// Before call Help, getCommand.f must be initialized.
func (c *getCommand) Help() string {
	return fmt.Sprintf(getHelpTmpl, FlagUsage(c.f, false), FlagUsage(c.rf(), true))
}

func (c *getCommand) Synopsis() string {
	return "Get a new tool as a dependency"
}

func (c *getCommand) Run(args []string) int {
	// First, parse normal flags only.
	// If repeatable flags or args found, break the loop.
	var normalFlags []string
	for i := 0; i < len(args); i++ {
		if args[i] == "-o" || !strings.HasPrefix(args[i], "-") {
			normalFlags = args[:i]
			break
		}
	}
	if err := c.f.Parse(normalFlags); err != nil {
		c.UI().Error(err.Error())
		return 1
	}

	var outputDir string
	var update bool
	outputDir = c.f.Lookup("d").Value.String()
	update = c.f.Lookup("u").Value.String() == "true"

	// Repeatable flags + all args.
	args = args[len(normalFlags):]

	return run(c, func() error {
		if len(args) == 0 {
			return errShowHelp
		}

		ctx := context.Background()

		paths, err := c.parseArgs(ctx, args)
		if err != nil {
			return err
		}

		err = c.workspace.Do(func(projRoot string, df *deptfile.File) error {
			if outputDir == "" {
				if b := os.Getenv("GOBIN"); b != "" {
					outputDir = b
				} else {
					outputDir = filepath.Join(projRoot, "_tools")
				}
			}

			importPaths := make([]string, 0, len(df.Require))
			for _, path := range paths {
				var targetReq *deptfile.Require
				for _, r := range df.Require {
					if r.Path == path.modRoot {
						tmp, err := copystructure.Copy(r)
						if err != nil {
							return errors.Wrap(err, "failed to deepcopy a Require")
						}
						targetReq = tmp.(*deptfile.Require)
					}
					var err error
					var i int
					forTools(r, func(importPath string) bool {
						importPaths = append(importPaths, importPath)
						if toolNameConflicted(importPath, path.repo) {
							err = errors.Errorf("tool names conflicted: %s and %s. please rename tool name by -o option.", path.repo, importPath)
							return false
						}
						// If -o passed with updating, rename tool to it.
						if importPath == path.repo && path.out != "" {
							r.ToolPaths[i].Name = path.out
						}
						i++
						return true
					})
					if err != nil {
						return err
					}
				}

				importPaths = append(importPaths, path.repo)

				df.Require = appendRequire(df.Require, targetReq, path)
			}

			f, err := os.Create("tools.go")
			if err != nil {
				return errors.Wrap(err, "failed to create a temp file which contains required Go tools in the import statement")
			}
			defer os.Remove("tools.go")
			defer f.Close()
			filegen.Generate(f, importPaths)

			// Always getCommand runs Get.
			// If an unmanaged tool is passed with -u option, '// indirect' will be marked
			// because it is not included in gotool.mod.
			getArgs := make([]string, 0, 1+len(paths))
			getArgs = append(getArgs, "-d")
			for _, p := range paths {
				getArgs = append(getArgs, p.modPath())
			}
			if err := c.gocmd.Get(ctx, append(getArgs, "./...")...); err != nil {
				return errors.Wrap(err, "failed to get Go tools dependencies")
			}

			handlePanic := func() {
				if err := recover(); err != nil {
					f.Close()
					os.Remove("tools.go")
					panic(err)
				}
			}
			eg, ctx := errgroup.WithContext(ctx)
			for _, path := range paths {
				path := path
				eg.Go(func() error {
					defer handlePanic()

					// If also -u is passed, update repo to the latest.
					if update && path.ver == "" {
						if err := c.gocmd.Get(ctx, "-u", "-d", path.repo); err != nil {
							return errors.Wrap(err, "failed to get Go tools dependencies")
						}
					}

					var binPath string
					if path.out != "" {
						binPath = filepath.Join(outputDir, path.out)
					} else {
						binPath = filepath.Join(outputDir, filepath.Base(path.repo))
					}
					if err := c.gocmd.Build(ctx, "-o", binPath, path.repo); err != nil {
						return errors.Wrapf(err, "failed to buld %s (bin path = %s)", path.repo, binPath)
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

func (c *getCommand) parseArgs(ctx context.Context, args []string) ([]*path, error) {
	paths, err := c.rf.Parse(args)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, errShowHelp
	}
	eg, ctx := errgroup.WithContext(ctx)
	for _, p := range paths {
		p := p
		eg.Go(func() (err error) {
			p.modRoot, err = getModuleRoot(ctx, c.gocmd, p.repo)
			return
		})
	}
	return paths, eg.Wait()
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
func appendRequire(reqs []*deptfile.Require, r *deptfile.Require, path *path) []*deptfile.Require {
	// defer pp.Println(reqs)
	// TODO: version, rename?

	toolPath := strings.TrimPrefix(path.repo, path.modRoot)
	// a new module
	if r == nil {
		r = &deptfile.Require{Path: path.modRoot}
		var t *deptfile.Tool
		// tool is not in the module root.
		if toolPath != "" {
			t = &deptfile.Tool{Path: toolPath, Name: path.out}
		} else {
			t = &deptfile.Tool{Path: "/", Name: path.out}
		}
		r.ToolPaths = append(r.ToolPaths, t)
		return append(reqs, r)
	}

	var t *deptfile.Tool
	if toolPath != "" {
		t = &deptfile.Tool{Path: toolPath, Name: path.out}
	} else {
		t = &deptfile.Tool{Path: "/", Name: path.out}
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
		if reqs[i].Path == path.modRoot {
			reqs[i] = r
			return reqs
		}
	}
	panic("must not reach to here")
}

type path struct {
	// val is the original value of path.
	// For example, 'github.com/ktr0731/salias@v0.1.0'
	val string
	// modRoot is the module root of path without the version.
	// For example, 'github.com/ktr0731/salias'
	modRoot string
	// repo is the repository name of val.
	// For example, 'github.com/ktr0731/salias'
	repo string
	// ver is the version of val.
	// For example, '@v0.1.0'
	ver string
	// out is the output name of the tool specified by path.
	// For example, 'salias'
	// If out is empty, it means out is same as filepath.Base(repo).
	out string
}

// modPath returns the completely module path which includes module's version.
func (p *path) modPath() string {
	if p.ver != "" {
		return p.modRoot + "@" + p.ver
	}
	return p.modRoot
}

type repeatableFlagSet func() *flag.FlagSet

var defaultRepeatableFlagSet repeatableFlagSet = func() *flag.FlagSet {
	f := flag.NewFlagSet("subget", flag.ExitOnError)
	f.String("o", "", "Output name")
	return f
}

// TODO: use flags
func (f repeatableFlagSet) Parse(args []string) ([]*path, error) {
	pargs := make([]*path, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch len(args[i:]) {
		case 0:
			return pargs, nil
		case 1, 2:
			// -o requires two arguments.
			if args[i] == "-o" {
				return nil, errShowHelp
			}
			p := args[i]
			repo, ver, err := normalizePath(p)
			if err != nil {
				return nil, err
			}
			pargs = append(pargs, &path{val: p, repo: repo, ver: ver})
		default:
			if args[i] == "-o" {
				p, out := args[i+2], args[i+1]
				repo, ver, err := normalizePath(p)
				if err != nil {
					return nil, err
				}
				pargs = append(pargs, &path{val: p, repo: repo, ver: ver, out: out})
				i += 2
			} else {
				p := args[i]
				repo, ver, err := normalizePath(p)
				if err != nil {
					return nil, err
				}
				pargs = append(pargs, &path{val: p, repo: repo, ver: ver})
			}
		}
	}
	return pargs, nil
}

// NewGet returns an initialized get command instance.
func NewGet(
	ui cli.Ui,
	gocmd gocmd.Command,
	workspace deptfile.Workspacer,
) cli.Command {
	f := flag.NewFlagSet("get", flag.ContinueOnError)
	// Suppress outputting by flag, delegate to cli.Command instead.
	f.SetOutput(ioutil.Discard)
	f.String("d", "", "Output dir to store built Go tools")
	f.Bool("u", false, "Update the specified tool to the latest version")
	rf := defaultRepeatableFlagSet
	return &getCommand{
		f:         f,
		rf:        rf,
		ui:        ui,
		gocmd:     gocmd,
		workspace: workspace,
	}
}
