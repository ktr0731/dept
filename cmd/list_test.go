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
						{Path: "honnef.co/go/tools", ToolPaths: []*deptfile.Tool{{Path: "/cmd/unused"}, {Path: "/cmd/statickcheck"}}},
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
}
