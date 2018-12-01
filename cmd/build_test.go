package cmd_test

import (
	"context"
	"path/filepath"
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
			DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
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
			assert      func(t *testing.T, args []string)
		}{
			"build two tools": {
				loadedTools: []*deptfile.Require{
					{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
					{Path: "github.com/ktr0731/itunes-cli", ToolPaths: []*deptfile.Tool{{Path: "/itunes"}}},
				},
			},
			"build renamed tools": {
				loadedTools: []*deptfile.Require{
					{Path: "github.com/ktr0731/itunes-cli", ToolPaths: []*deptfile.Tool{{Path: "/itunes", Name: "it"}}},
				},
				assert: func(t *testing.T, args []string) {
					if n := filepath.Base(args[1]); n != "it" {
						t.Errorf("output name must be 'it', but actual '%s'", n)
					}
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
					DoFunc: func(f func(projectDir string, df *deptfile.File) error) error {
						df := &deptfile.File{Require: c.loadedTools}
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

				if c.assert != nil {
					buildArgs := mockGoCMD.BuildCalls()[0].Args
					c.assert(t, buildArgs)
				}
			})
		}
	})
}
