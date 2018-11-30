package deptfile

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ktr0731/modfile"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
)

var (
	DeptfileName    = "gotool.mod"
	DeptfileSumName = "gotool.sum"
)

var (
	// ErrNotFound represents deptfile not found.
	ErrNotFound = errors.Errorf("%s not found", DeptfileName)
	// ErrAlreadyExist represents deptfile alredy exist.
	ErrAlreadyExist = errors.New("already exist")
)

// TODO: rename to Deptfile
type GoMod struct {
	Require []*Require
	f       *modfile.File
}

// Require represents a parsed direct requirement.
// A Require has least one Tool.
type Require struct {
	Path    string
	Version string
	// TODO: root package representation
	ToolPaths []*Tool
}

func (r *Require) format() string {
	s := r.Path
	if len(r.ToolPaths) == 0 {
		return s
	}
	toolPaths := make([]string, 0, len(r.ToolPaths))
	for _, t := range r.ToolPaths {
		toolPaths = append(toolPaths, t.format())
	}
	s += ":" + strings.Join(toolPaths, ",")
	return s
}

// Tool represents a tool that is belongs to a module.
// In deptfile representation, a module is represents as a Require.
// Path is the absolute tool path from the module root.
// If Path is empty, it means the package of the tool is in the module root.
// Name is the tool name.
type Tool struct {
	Path string
	Name string
}

func (t *Tool) format() string {
	s := t.Path
	if n := t.Name; n != "" {
		s += "@" + n
	}
	return s
}

// parseDeptfile parses a file which named fname as a deptfile.
// The differences between deptfile and go.mod is just one point,
// deptfile's each path has also command paths.
//
// For example:
//   "github.com/ktr0731/evans": module is github.com/ktr0731/evans, the command path is the module root.
//   "github.com/ktr0731/itunes-cli:/itunes": module is github.com/ktr0731/itunes-cli, the command path is /itunes.
//   "honnef.co/go/tools:/cmd/staticcheck,/cmd/unused": module is honnef.co/go/tools, command paths are /cmd/staticcheck and /cmd/unused.
//
// Also parseDeptfile returns the canonical modfile. It has been removed command paths.
// So, it is go.mod compatible.
//
// parseDeptfile returns ErrNotFound if fname is not found.
// TODO: rename syntax
func parseDeptfile(fname string) (*GoMod, *modfile.File, error) {
	data, err := ioutil.ReadFile(fname)
	if os.IsNotExist(err) {
		return nil, nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to open %s", fname)
	}
	f, err := modfile.Parse(filepath.Base(fname), data, nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse %s", fname)
	}

	tmp, err := copystructure.Copy(f)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to deep copy modfile.File")
	}
	canonical := tmp.(*modfile.File)

	// Convert from modfile.File.Require to deptfile.Require.
	requires := make([]*Require, 0, len(f.Require))
	for i, r := range f.Require {
		// Skip indirect requirements because deptfile focuses on direct requirements (= managed tools) only.
		if r.Indirect {
			continue
		}

		var toolPaths []*Tool
		path := r.Mod.Path

		// main package is not in the module root
		if i := strings.LastIndex(r.Mod.Path, ":"); i != -1 {
			path = r.Mod.Path[:i]
			for _, toolPath := range strings.Split(r.Mod.Path[i+1:], ",") {
				toolPaths = append(toolPaths, &Tool{Path: toolPath})
			}
		} else {
			// TODO
		}

		requires = append(requires, &Require{
			Path:      path,
			Version:   r.Mod.Version,
			ToolPaths: toolPaths,
		})
		canonical.Require[i].Mod.Path = path
		canonical.Require[i].Syntax.Token[0] = path
	}
	canonical.SetRequire(canonical.Require)
	return &GoMod{Require: requires, f: f}, canonical, nil
}

func convertGoModToDeptfile(fname string, gomod *GoMod) (*modfile.File, error) {
	data, err := ioutil.ReadFile(fname)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", fname)
	}
	f, err := modfile.Parse(filepath.Base(fname), data, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", fname)
	}

	// no any additional information
	if gomod == nil {
		return f, nil
	}

	path2req := map[string]*Require{}
	for _, r := range gomod.Require {
		path2req[r.Path] = r
	}

	for i := range f.Require {
		if f.Require[i].Indirect {
			continue
		}
		req, ok := path2req[f.Require[i].Mod.Path]
		// new tool
		if !ok {
			continue
		}
		p := req.format()
		f.Require[i].Mod.Path = p

		// require statement is oneline.
		if f.Require[i].Syntax.Token[0] == "require" {
			f.Require[i].Syntax.Token[1] = p
		} else {
			f.Require[i].Syntax.Token[0] = p
		}
	}

	f.SetRequire(f.Require)

	return f, nil
}

// Create creates a new deptfile.
// If already created, Create returns ErrAlreadyExist.
func Create(ctx context.Context) error {
	if _, err := os.Stat(DeptfileName); err == nil {
		return ErrAlreadyExist
	}

	var err error
	w := &Workspace{
		SourcePath: ".",
		DoNotCopy:  true,
	}
	err = w.Do(func(string, *GoMod) error {
		// TODO: module name
		err = exec.CommandContext(ctx, "go", "mod", "init", "tools").Run()
		if err != nil {
			return errors.Wrap(err, "failed to init Go modules")
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
