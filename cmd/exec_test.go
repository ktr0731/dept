package cmd_test

import (
	"context"
	"testing"

	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/toolcacher"
)

func TestExecRun(t *testing.T) {
	t.Run("Run returns code 1 because no arguments passed", func(t *testing.T) {
		mockUI := newMockUI()
		cmd := cmd.NewExec(mockUI, nil, nil)

		code := cmd.Run(nil)
		if code != 1 {
			t.Errorf("Run must return 1, but got %d (err = %s)", code, mockUI.ErrorWriter().String())
		}
	})

	t.Run("Run returns code 1 because the specified tool is not found", func(t *testing.T) {
		cases := map[string]struct {
			loadedTool *deptfile.Require
		}{
			"no tools": {
				loadedTool: nil,
			},
			"not found": {
				loadedTool: &deptfile.Require{
					Path: "github.com/ktr0731/itunes-cli",
					ToolPaths: []*deptfile.Tool{
						{Path: "/itunes"},
					},
				},
			},
			"same tool name, but different out name": {
				loadedTool: &deptfile.Require{
					Path: "github.com/ktr0731/salias",
					ToolPaths: []*deptfile.Tool{
						{Path: "/", Name: "sa"},
					},
				},
			},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				mockUI := newMockUI()
				mockWorkspace := &deptfile.WorkspacerMock{
					DoFunc: func(f func(projectDir string, df *deptfile.File) error) error {
						df := &deptfile.File{Require: []*deptfile.Require{c.loadedTool}}
						return f("", df)
					},
				}
				cmd := cmd.NewExec(mockUI, mockWorkspace, nil)

				code := cmd.Run([]string{"salias"})
				if code != 1 {
					t.Errorf("Run must return 1, but got %d", code)
				}
			})
		}
	})

	t.Run("Run returns code 0", func(t *testing.T) {
		cases := map[string]struct {
			loadedTool *deptfile.Require
		}{
			"top-level": {
				loadedTool: &deptfile.Require{
					Path:    "github.com/ktr0731/salias",
					Version: "v0.1.0",
					ToolPaths: []*deptfile.Tool{
						{Path: "/"},
					},
				},
			},
			"nested": {
				loadedTool: &deptfile.Require{
					Path:    "github.com/ktr0731/salias",
					Version: "v0.1.0",
					ToolPaths: []*deptfile.Tool{
						{Path: "/cmd/salias"},
					},
				},
			},
			"top-level with renaming": {
				loadedTool: &deptfile.Require{
					Path:    "github.com/ktr0731/saliasv2",
					Version: "v0.1.0",
					ToolPaths: []*deptfile.Tool{
						{Path: "/", Name: "salias"},
					},
				},
			},
			"nested with renaming": {
				loadedTool: &deptfile.Require{
					Path:    "github.com/ktr0731/salias",
					Version: "v0.1.0",
					ToolPaths: []*deptfile.Tool{
						{Path: "/cmd/saliasv2", Name: "salias"},
					},
				},
			},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				mockUI := newMockUI()
				mockWorkspace := &deptfile.WorkspacerMock{
					DoFunc: func(f func(projectDir string, df *deptfile.File) error) error {
						df := &deptfile.File{Require: []*deptfile.Require{c.loadedTool}}
						return f("", df)
					},
				}
				mockToolcacher := &toolcacher.CacherMock{
					GetFunc: func(ctx context.Context, pkgName string, version string) (string, error) {
						return "", nil
					},
				}
				cleanup := cmd.ChnageSyscallExec(func(argv0 string, argv []string, envv []string) (err error) {
					return nil
				})
				defer cleanup()

				cmd := cmd.NewExec(mockUI, mockWorkspace, mockToolcacher)

				code := cmd.Run([]string{"salias"})
				if code != 0 {
					t.Errorf("Run must return 0, but got %d (err = %s)", code, mockUI.ErrorWriter().String())
				}
			})
		}
	})
}
