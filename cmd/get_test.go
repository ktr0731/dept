package cmd_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/gocmd"
	"github.com/pkg/errors"
)

func TestGetRun(t *testing.T) {
	doNothing := func(f func(projectDir string, gomod *deptfile.GoMod) error) error { return f("", nil) }

	t.Run("Run returns code 1 because no arguments passed", func(t *testing.T) {
		mockUI := newMockUI()
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: doNothing,
		}
		cmd := cmd.NewGet(mockUI, nil, mockWorkspace)

		code := cmd.Run(nil)
		if code != 1 {
			t.Errorf("Run must be 1 because command need to show help message")
		}
		if out := mockUI.Writer().String(); !strings.HasPrefix(out, "Usage: ") {
			t.Errorf("Run must write help message to Writer, but actual '%s'", out)
		}
	})

	t.Run("Run returns 1 because gotool.mod is not found", func(t *testing.T) {
		mockUI := newMockUI()
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: func(f func(projectDir string, gomod *deptfile.GoMod) error) error {
				return deptfile.ErrNotFound
			},
		}

		cmd := cmd.NewGet(mockUI, nil, mockWorkspace)

		repo := "github.com/ktr0731/go-modules-test"
		code := cmd.Run([]string{repo})

		if code != 1 {
			t.Errorf("Run must return code 1, but got %d", code)
		}
		if eout := mockUI.ErrorWriter().String(); !strings.Contains(eout, "dept init") {
			t.Errorf("Run must show 'dept init' related error message in case of gotool.mod is not found, but '%s'", eout)
		}
	})

	t.Run("Run returns code 0 normally", func(t *testing.T) {
		cases := map[string]struct {
			loadedTools     []*deptfile.Require
			root            string
			args            []string
			expectedRequire *deptfile.Require
			update          bool
		}{
			"get a new tool": {
				args:            []string{"github.com/ktr0731/evans"},
				root:            "github.com/ktr0731/evans",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/evans"},
			},
			"get a new tool with HTTP scheme": {
				args:            []string{"https://github.com/ktr0731/evans"},
				root:            "github.com/ktr0731/evans",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/evans"},
			},
			"get a new tool that is not in the module root": {
				args:            []string{"github.com/ktr0731/itunes-cli/itunes"},
				root:            "github.com/ktr0731/itunes-cli",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/itunes-cli", CommandPath: []string{"/itunes"}},
			},
			"add a new tool that is not in the module root": {
				loadedTools: []*deptfile.Require{{Path: "honnef.co/go/tools", CommandPath: []string{"/cmd/unused"}}},
				args:        []string{"honnef.co/go/tools/cmd/staticcheck"},
				root:        "honnef.co/go/tools",
				expectedRequire: &deptfile.Require{
					Path:        "honnef.co/go/tools",
					CommandPath: []string{"/cmd/unused", "/cmd/staticcheck"},
				},
			},
			"add a new tool that is in the module root, but has command path": {
				loadedTools: []*deptfile.Require{{Path: "github.com/foo/bar", CommandPath: []string{"/cmd/baz"}}},
				args:        []string{"github.com/foo/bar"},
				root:        "github.com/foo/bar",
				expectedRequire: &deptfile.Require{
					Path:        "github.com/foo/bar",
					CommandPath: []string{"/", "/cmd/baz"},
				},
			},
			"add a new tool that is not in the module root, but has top-level command": {
				loadedTools: []*deptfile.Require{{Path: "github.com/foo/bar"}},
				args:        []string{"github.com/foo/bar/cmd/baz"},
				root:        "github.com/foo/bar",
				expectedRequire: &deptfile.Require{
					Path:        "github.com/foo/bar",
					CommandPath: []string{"/", "/cmd/baz"},
				},
			},
			"add a duplicated tool": {
				loadedTools: []*deptfile.Require{{Path: "github.com/foo/bar", CommandPath: []string{"/cmd/baz"}}},
				args:        []string{"github.com/foo/bar/cmd/baz"},
				root:        "github.com/foo/bar",
				expectedRequire: &deptfile.Require{
					Path:        "github.com/foo/bar",
					CommandPath: []string{"/cmd/baz"},
				},
			},
			"update a tool": {
				loadedTools:     []*deptfile.Require{{Path: "github.com/ktr0731/evans"}},
				args:            []string{"github.com/ktr0731/evans"},
				root:            "github.com/ktr0731/evans",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/evans"},
				update:          true,
			},
			"-u also works if the specified tool is not found": {
				args:            []string{"github.com/ktr0731/evans"},
				root:            "github.com/ktr0731/evans",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/evans"},
				update:          true,
			},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				mockUI := newMockUI()
				mockGoCMD := &gocmd.CommandMock{
					GetFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
					BuildFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
					ListFunc: func(ctx context.Context, args ...string) (io.Reader, error) {
						return strings.NewReader(c.root), nil
					},
				}
				mockWorkspace := &deptfile.WorkspacerMock{
					DoFunc: func(f func(projectDir string, df *deptfile.GoMod) error) error {
						df := &deptfile.GoMod{Require: c.loadedTools}
						if err := f("", df); err != nil {
							return err
						}

						if n := len(df.Require); n != 1 {
							t.Fatalf("expected only 1 tool, but %d", n)
						}
						if diff := cmp.Diff(df.Require[0], c.expectedRequire); diff != "" {
							t.Fatalf("modified deptfile.GoMod is not equal to the expected one:\n%s", diff)
						}
						return nil
					},
				}
				cmd := cmd.NewGet(mockUI, mockGoCMD, mockWorkspace)

				var args []string
				if c.update {
					args = append([]string{"-u"}, c.args...)
				} else {
					args = c.args
				}
				code := cmd.Run(args)
				if code != 0 {
					t.Errorf("Run must return 0, but got %d (err = %s)", code, mockUI.ErrorWriter().String())
				}

				if c.update {
					if n := len(mockGoCMD.GetCalls()); n != 2 {
						t.Errorf("Get must be called twice, but actual %d", n)
					}
				} else {
					if n := len(mockGoCMD.GetCalls()); n != 1 {
						t.Errorf("Get must be called once, but actual %d", n)
					}
				}

				if n := len(mockGoCMD.BuildCalls()); n != 1 {
					t.Errorf("Build must be called once, but actual %d", n)
				}
			})
		}
	})

	t.Run("deptfile is not modified when command failed", func(t *testing.T) {
		mockUI := newMockUI()
		mockGoCMD := &gocmd.CommandMock{
			ListFunc: func(ctx context.Context, args ...string) (io.Reader, error) {
				return nil, errors.New("an error")
			},
		}
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: func(f func(projectDir string, gomod *deptfile.GoMod) error) error {
				return f("", &deptfile.GoMod{
					Require: []*deptfile.Require{},
				})
			},
		}

		cmd := cmd.NewGet(mockUI, mockGoCMD, mockWorkspace)

		repo := "github.com/ktr0731/go-modules-test"
		code := cmd.Run([]string{repo})
		if code != 1 {
			t.Errorf("Run must be 1, but got %d", code)
		}

		if n := len(mockGoCMD.ListCalls()); n != 1 {
			t.Errorf("List must be called once, but actual %d (err = %s)", n, mockUI.ErrorWriter().String())
		}
	})

	t.Run("Run returns an error when tool names conflicted", func(t *testing.T) {
		cases := map[string]struct {
			hasErr  bool
			require string
		}{
			"same tools":                                {hasErr: false, require: "github.com/ktr0731/evans"},
			"same tools with versions":                  {hasErr: false, require: "github.com/ktr0731/evans@v0.2.0"},
			"different tools, different names":          {hasErr: false, require: "github.com/foo/bar"},
			"different tools, same names":               {hasErr: true, require: "github.com/foo/evans"},
			"different tools, same names with versions": {hasErr: true, require: "github.com/foo/evans@v0.2.0"},
			"invalid module version syntax":             {hasErr: true, require: "github.com/foo/evans@"},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				repo, _, err := cmd.NormalizePath(c.require)
				if err != nil {
					t.Fatalf("failed to normalize c.require (%s): %s", c.require, err)
				}

				mockUI := newMockUI()
				mockGoCMD := &gocmd.CommandMock{
					GetFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
					BuildFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
					ListFunc: func(ctx context.Context, args ...string) (io.Reader, error) {
						return strings.NewReader(repo), nil
					},
				}
				mockWorkspace := &deptfile.WorkspacerMock{
					DoFunc: func(f func(projectDir string, gomod *deptfile.GoMod) error) error {
						return f("", &deptfile.GoMod{
							Require: []*deptfile.Require{
								{Path: "github.com/ktr0731/evans"},
							},
						})
					},
				}

				cmd := cmd.NewGet(mockUI, mockGoCMD, mockWorkspace)

				code := cmd.Run([]string{c.require})
				if c.hasErr {
					if code != 1 {
						t.Errorf("Run must return 1, but got %d", code)
					}
					if mockUI.ErrorWriter().String() == "" {
						t.Error("Run must output an error message, but empty")
					}
				} else {
					if code != 0 {
						t.Errorf("Run must return 0, but got %d", code)
					}
					if eout := mockUI.ErrorWriter().String(); eout != "" {
						t.Errorf("Run must not output any error messages, but got '%s'", eout)
					}
				}
			})
		}
	})
}
