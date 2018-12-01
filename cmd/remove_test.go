package cmd_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/gocmd"
)

func TestRemoveRun(t *testing.T) {
	doNothing := func(f func(projectDir string, gomod *deptfile.File) error) error { return f("", nil) }

	t.Run("Run returns code 1 because no arguments passed", func(t *testing.T) {
		mockUI := newMockUI()
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: doNothing,
		}
		cmd := cmd.NewRemove(mockUI, nil, mockWorkspace)

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
			DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
				return deptfile.ErrNotFound
			},
		}

		cmd := cmd.NewRemove(mockUI, nil, mockWorkspace)

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
			repo     string
			requires []*deptfile.Require
			hasErr   bool
		}{
			"tool not found": {
				repo:     "github.com/wa2/haruki",
				requires: []*deptfile.Require{{Path: "github.com/wa2/kazusa", ToolPaths: []*deptfile.Tool{{Path: "/"}}}},
				hasErr:   true,
			},
			"main package is in the module root": {
				repo:     "github.com/wa2/kazusa",
				requires: []*deptfile.Require{{Path: "github.com/wa2/kazusa", ToolPaths: []*deptfile.Tool{{Path: "/"}}}},
			},
			"main package is in the module root with /": {
				repo:     "github.com/wa2/kazusa/",
				requires: []*deptfile.Require{{Path: "github.com/wa2/kazusa", ToolPaths: []*deptfile.Tool{{Path: "/"}}}},
			},
			"main package is in the module root  with HTTP scheme": {
				repo:     "https://github.com/wa2/kazusa",
				requires: []*deptfile.Require{{Path: "github.com/wa2/kazusa", ToolPaths: []*deptfile.Tool{{Path: "/"}}}},
			},
			"main package is not in the module root": {
				repo: "github.com/leaf/wa2/cmd/closing",
				requires: []*deptfile.Require{
					{Path: "github.com/leaf/wa2", ToolPaths: []*deptfile.Tool{{Path: "/cmd/introductory"}, {Path: "/cmd/closing"}}},
				},
			},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				mockUI := newMockUI()
				mockGoCMD := &gocmd.CommandMock{
					ModTidyFunc: func(ctx context.Context) error {
						return nil
					},
				}
				mockWorkspace := &deptfile.WorkspacerMock{
					DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
						return f("", &deptfile.File{
							Require: c.requires,
						})
					},
				}
				cmd := cmd.NewRemove(mockUI, mockGoCMD, mockWorkspace)

				code := cmd.Run([]string{c.repo})
				if c.hasErr {
					if code == 0 {
						t.Errorf("Run must return 1, but got %d (err = %s)", code, mockUI.ErrorWriter().String())
					}
				} else {
					if code != 0 {
						t.Errorf("Run must return 0, but got %d (err = %s)", code, mockUI.ErrorWriter().String())
					}

					if n := len(mockGoCMD.ModTidyCalls()); n != 1 {
						t.Errorf("ModTidy must be called once, but actual %d", n)
					}
				}
			})
		}
	})

}
