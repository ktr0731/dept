package cmd_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
)

func TestListRun(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working dir: %s", err)
	}
	setup := func() func() {
		return func() {
			os.Chdir(cwd)
		}
	}

	t.Run("Run shows direction packages with code 0 normally", func(t *testing.T) {
		cases := map[string]struct {
			args     []string
			expected int
		}{
			"shows direction packages only": {expected: 3},
			"shows all packages":            {args: []string{"-i"}, expected: 6},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				cleanup := setup()
				defer cleanup()

				mockUI := newMockUI()
				workspace := &deptfile.Workspace{SourcePath: "testdata"}
				cmd := cmd.NewList(mockUI, workspace)

				code := cmd.Run(c.args)
				if code != 0 {
					fmt.Println(mockUI.Writer().String())
					t.Fatalf("Run must return 0, but got %d", code)
				}

				sp := strings.Split(strings.TrimSpace(mockUI.Writer().String()), "\n")
				if len(sp) != c.expected {
					t.Errorf("Run must show %d tools, but got %d", c.expected, len(sp))
				}
			})
		}
	})
}
