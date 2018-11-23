package cmd_test

import (
	"context"
	"strings"
	"testing"

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
			paths  []string
			args   []string
			update bool
		}{
			"get a new tool":                  {args: []string{"github.com/ktr0731/evans"}},
			"get a new tool with HTTP scheme": {args: []string{"https://github.com/ktr0731/evans"}},
			"update a tool": {
				paths:  []string{"github.com/ktr0731/evans"},
				args:   []string{"github.com/ktr0731/evans"},
				update: true,
			},
			"-u also works if the specified tool is not found": {args: []string{"github.com/ktr0731/evans"}, update: true},
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
				}
				mockWorkspace := &deptfile.WorkspacerMock{
					DoFunc: func(f func(projectDir string, gomod *deptfile.GoMod) error) error {
						requires := make([]*deptfile.Require, 0, len(c.paths))
						for _, p := range c.paths {
							requires = append(requires, &deptfile.Require{Path: p})
						}
						return f("", &deptfile.GoMod{Require: requires})
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
			GetFunc: func(ctx context.Context, pkgs ...string) error {
				return errors.New("an error")
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

		if n := len(mockGoCMD.GetCalls()); n != 1 {
			t.Errorf("Get must be called once, but actual %d (err = %s)", n, mockUI.ErrorWriter().String())
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
				mockUI := newMockUI()
				mockGoCMD := &gocmd.CommandMock{
					GetFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
					BuildFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
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
