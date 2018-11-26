package cmd_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/gocmd"
)

func TestBuildRun(t *testing.T) {
	t.Run("Run returns 1 because gotool.mod is not found", func(t *testing.T) {
		mockUI := newMockUI()
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: func(f func(projectDir string, gomod *deptfile.GoMod) error) error {
				return deptfile.ErrNotFound
			},
		}

		cmd := cmd.NewBuild(mockUI, nil, mockWorkspace)

		code := cmd.Run(nil)

		if code != 1 {
			t.Errorf("Run must return code 1, but got %d", code)
		}
		if eout := mockUI.ErrorWriter().String(); !strings.Contains(eout, "dept init") {
			t.Errorf("Run must show 'dept init' related error message in case of gotool.mod is not found, but '%s'", eout)
		}
	})

	t.Run("Run returns code 0 normally", func(t *testing.T) {
		cases := map[string]struct {
			loadedTools []*deptfile.Require
		}{
			"get a new tool": {
				loadedTools: []*deptfile.Require{
					{Path: "github.com/ktr0731/evans"},
					{Path: "github.com/ktr0731/itunes-cli", CommandPath: []string{"/itunes"}},
				},
			},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				mockUI := newMockUI()
				mockGoCMD := &gocmd.CommandMock{
					BuildFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
				}
				mockWorkspace := &deptfile.WorkspacerMock{
					DoFunc: func(f func(projectDir string, df *deptfile.GoMod) error) error {
						df := &deptfile.GoMod{Require: c.loadedTools}
						return f("", df)
					},
				}
				cmd := cmd.NewBuild(mockUI, mockGoCMD, mockWorkspace)

				code := cmd.Run(nil)
				if code != 0 {
					t.Errorf("Run must return 0, but got %d (err = %s)", code, mockUI.ErrorWriter().String())
				}

				if n := len(mockGoCMD.BuildCalls()); n != len(c.loadedTools) {
					t.Errorf("Build must be called %d times, but actual %d", len(c.loadedTools), n)
				}
			})
		}
	})
}