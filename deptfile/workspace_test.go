package deptfile_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/modfile"
	"github.com/pkg/errors"
)

var _ deptfile.Workspacer = (*deptfile.Workspace)(nil)

// TestDo runs tests against to testdata/gotool.mod.
func TestDo(t *testing.T) {
	testRequires := func(t *testing.T, r []*deptfile.Require, cond map[string]func(r *deptfile.Require) error) {
		t.Helper()
		passed := map[string]bool{}
		for _, r := range r {
			for modName, cond := range cond {
				if _, ok := passed[modName]; !ok {
					passed[modName] = false
				}
				if r.Path != modName {
					continue
				}
				if err := cond(r); err != nil {
					t.Errorf("%s: %s", r.Path, err)
				} else {
					passed[modName] = true
				}
			}
		}
		for modName, passed := range passed {
			if !passed {
				t.Errorf("tests for %s are not passed. invalid condition?", modName)
			}
		}
	}

	t.Run("workspace changes directory to a temp dir and copies gotool.mod to there", func(t *testing.T) {
		cases := map[string]struct {
			dir        string
			numRequire int
			testcases  map[string]func(r *deptfile.Require) error
		}{
			"some required tools": {
				dir:        "normal",
				numRequire: 4,
				testcases: map[string]func(r *deptfile.Require) error{
					"github.com/ktr0731/evans": func(r *deptfile.Require) error {
						if r.CommandPath != nil {
							return errors.New("CommandPath must be nil if the module has main package in the module root")
						}
						return nil
					},
					"honnef.co/go/tools": func(r *deptfile.Require) error {
						if r.CommandPath == nil {
							return errors.New("CommandPath must not be nil if the module has main package other than the module root")
						}
						if n := len(r.CommandPath); n != 2 {
							return errors.Errorf("expected 2 command in this module, but got %d", n)
						}
						if r.CommandPath[0] != "/cmd/staticcheck" {
							return errors.Errorf("expected r.CommandPath[0] is '/cmd/staticcheck', but %s", r.CommandPath[0])
						}
						if r.CommandPath[1] != "/cmd/unused" {
							return errors.Errorf("expected r.CommandPath[0] is '/cmd/unused', but %s", r.CommandPath[0])
						}
						return nil
					},
				},
			},
			"only one require tool": {
				dir:        "oneline",
				numRequire: 1,
				testcases: map[string]func(r *deptfile.Require) error{
					"github.com/ktr0731/evans": func(r *deptfile.Require) error {
						if r.CommandPath != nil {
							return errors.New("CommandPath must be nil if the module has main package in the module root")
						}
						return nil
					},
				},
			},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				cleanup := setupEnv(t, filepath.Join("testdata", c.dir))
				defer cleanup()

				cwd, err := os.Getwd()
				if err != nil {
					t.Fatalf("failed to get cwd: %s", err)
				}

				w := &deptfile.Workspace{
					SourcePath: cwd,
				}
				err = w.Do(func(proj string, gomod *deptfile.GoMod) error {
					if gomod == nil {
						t.Fatalf("deptfile must not be nil, but nil")
					}

					if n := len(gomod.Require); n != c.numRequire {
						t.Errorf("deptfile must have %d required modules (managed tools), but %d", c.numRequire, n)
					}

					testRequires(t, gomod.Require, c.testcases)

					newcwd, err := os.Getwd()
					if err != nil {
						t.Fatalf("failed to get current working dir: %s", err)
					}
					if cwd == newcwd {
						t.Errorf("current dir in Do must not be equal to dir outside of Do")
					}
					return nil
				})
				if err != nil {
					t.Errorf("Do must not return errors, but got an error: %s", err)
				}

				cwd2, err := os.Getwd()
				if err != nil {
					t.Errorf("failed to get current working dir: %s", err)
				}
				if cwd != cwd2 {
					t.Errorf("current working dir which called before Do and after one must be equal, but %s and %s", cwd, cwd2)
				}

				if _, err := os.Stat(deptfile.DeptfileName); os.IsNotExist(err) {
					t.Errorf("gotool.mod must be in the current dir, but not found")
				}

				checkGoModSyntax(t)
			})
		}
	})

	t.Run("workspace returns ErrNotFound", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("failed to create a temp dir: %s", err)
		}
		w := &deptfile.Workspace{SourcePath: dir}
		err = w.Do(func(proj string, gomod *deptfile.GoMod) error {
			return nil
		})
		if err != deptfile.ErrNotFound {
			t.Errorf("workspace must return ErrNotFound because gotool.mod is not found in current working dir, but '%s'", err)
		}
	})
}

func checkGoModSyntax(t *testing.T) {
	b, err := ioutil.ReadFile(deptfile.DeptfileName)
	if err != nil {
		t.Fatalf("failed to read %s", deptfile.DeptfileName)
	}
	_, err = modfile.Parse(deptfile.DeptfileName, b, nil)
	if err != nil {
		fmt.Println(string(b))
		t.Fatalf("failed to parse %s: %s", deptfile.DeptfileName, err)
	}
}
