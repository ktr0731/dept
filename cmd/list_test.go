package cmd_test

import (
	"fmt"
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
						{Path: "github.com/ktr0731/itunes-cli/itunes"},
					},
				})
			},
		}
		cmd := cmd.NewList(mockUI, mockWorkspace)

		code := cmd.Run(nil)
		if code != 0 {
			t.Fatalf("Run must return 0, but got %d", code)
		}
		fmt.Println(mockUI.Writer().String())

		sp := strings.Split(strings.TrimSpace(mockUI.Writer().String()), "\n")
		if len(sp) != 2 {
			t.Errorf("Run must show 2 tools, but got %d", len(sp))
		}
	})
}
