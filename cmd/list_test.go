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
			DoFunc: func(f func(projectDir string, gomod *deptfile.GoMod) error) error {
				return f("", &deptfile.GoMod{
					Require: []*deptfile.Require{
						{Path: "github.com/ktr0731/evans"},
						{Path: "github.com/ktr0731/itunes-cli", CommandPath: []string{"/itunes"}},
						{Path: "honnef.co/go/tools", CommandPath: []string{"/cmd/unused", "/cmd/staticcheck"}},
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
