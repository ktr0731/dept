package cmd_test

import (
	"strings"
	"testing"

	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
)

func TestListRun(t *testing.T) {
	t.Run("Run shows direction packages with code 0 normally", func(t *testing.T) {
		mockUI := newMockUI()
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
				return f("", &deptfile.File{
					Require: []*deptfile.Require{
						{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
						{Path: "github.com/ktr0731/itunes-cli", ToolPaths: []*deptfile.Tool{{Path: "/itunes"}}},
						{Path: "honnef.co/go/tools", ToolPaths: []*deptfile.Tool{{Path: "/cmd/unused"}, {Path: "/cmd/staticcheck"}}},
					},
				})
			},
		}
		cmd := cmd.NewList(mockUI, mockWorkspace)

		code := cmd.Run(nil)
		if code != 0 {
			t.Fatalf("Run must return 0, but got %d", code)
		}

		sp := strings.Split(strings.TrimSpace(mockUI.Writer().String()), "\n")
		if len(sp) != 4 {
			t.Errorf("Run must show 4 tools, but got %d", len(sp))
		}
	})

	t.Run("Run shows only specified tools", func(t *testing.T) {
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
				return f("", &deptfile.File{
					Require: []*deptfile.Require{
						{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
						{Path: "github.com/ktr0731/itunes-cli", ToolPaths: []*deptfile.Tool{{Path: "/itunes"}}},
						{Path: "honnef.co/go/tools", ToolPaths: []*deptfile.Tool{{Path: "/cmd/unused"}, {Path: "/cmd/staticcheck"}}},
					},
				})
			},
		}

		cases := map[string]struct {
			args     []string
			expected []string
		}{
			"only one tool": {
				args:     []string{"github.com/ktr0731/evans"},
				expected: []string{"github.com/ktr0731/evans"},
			},
			"only one tool which is in a sub package": {
				args:     []string{"honnef.co/go/tools/cmd/unused"},
				expected: []string{"honnef.co/go/tools/cmd/unused"},
			},
			"only one module": {
				args: []string{"honnef.co/go/tools"},
				expected: []string{
					"honnef.co/go/tools/cmd/unused",
					"honnef.co/go/tools/cmd/staticcheck",
				},
			},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				mockUI := newMockUI()
				cmd := cmd.NewList(mockUI, mockWorkspace)

				code := cmd.Run(c.args)
				if code != 0 {
					t.Fatalf("Run must return 0, but got %d", code)
				}

				actual := strings.Split(strings.TrimSpace(mockUI.Writer().String()), "\n")
				if len(c.expected) != len(actual) {
					t.Errorf("expected: %s, but got %s", c.expected, actual)
				}
			})
		}
	})

	t.Run("Run shows tools with -f based format", func(t *testing.T) {
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
				return f("", &deptfile.File{
					Require: []*deptfile.Require{
						{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
						{Path: "github.com/ktr0731/itunes-cli", ToolPaths: []*deptfile.Tool{{Path: "/itunes", Name: "it"}}},
						{Path: "honnef.co/go/tools", ToolPaths: []*deptfile.Tool{{Path: "/cmd/unused"}, {Path: "/cmd/staticcheck"}}},
					},
				})
			},
		}

		cases := map[string]struct {
			args     []string
			expected []string
			hasErr   bool
		}{
			"{{.Name}}": {
				args:     []string{"-f", "{{.Name}}"},
				expected: []string{"evans", "it", "unused", "staticcheck"},
			},
			"invalid format": {
				args:   []string{"-f", "{{"},
				hasErr: true,
			},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				mockUI := newMockUI()
				cmd := cmd.NewList(mockUI, mockWorkspace)

				code := cmd.Run(c.args)
				if c.hasErr {
					if code == 0 {
						t.Error("Run must return 1, but got 0")
					}
					return
				}
				if code != 0 {
					t.Fatalf("Run must return 0, but got %d", code)
				}

				actual := mockUI.Writer().String()
				expected := strings.Join(c.expected, "\n") + "\n"
				if expected != actual {
					t.Errorf("expected: %s, but got %s", expected, actual)
				}
			})
		}
	})
}
