package cmd_test

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/gocmd"
)

func TestRemoveRun(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working dir: %s", err)
	}
	setup := func(m *deptfile.GoMod) func() {
		if m == nil {
			m = &deptfile.GoMod{
				Require: []deptfile.Require{},
			}
		}
		cleanup := cmd.ChangeDeptfileLoad(func(context.Context) (*deptfile.GoMod, error) {
			return m, nil
		})
		return func() {
			cleanup()
			// workspace.Do also change back to current working dir.
			// However, we will specify SourcePath to "testdata", so changed dir will be "testdata", not package dir.
			// We change directory to package dir manually.
			os.Chdir(cwd)
		}
	}

	t.Run("Run returns code 1 because no arguments passed", func(t *testing.T) {
		mockUI := newMockUI()
		workspace := &deptfile.Workspace{SourcePath: "testdata"}
		cmd := cmd.NewRemove(mockUI, nil, workspace)

		code := cmd.Run(nil)
		if code != 1 {
			t.Errorf("Run must be 1 because command need to show help message")
		}
		if out := mockUI.Writer().String(); !strings.HasPrefix(out, "Usage: ") {
			t.Errorf("Run must write help message to Writer, but actual '%s'", out)
		}
	})

	t.Run("Run returns 1 because gotool.mod is not found", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("failed to create a temp dir: %s", err)
		}

		mockUI := newMockUI()
		workspace := &deptfile.Workspace{SourcePath: dir}

		cleanup := setup(nil)
		defer cleanup()

		cmd := cmd.NewRemove(mockUI, nil, workspace)

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
		m := &deptfile.GoMod{
			Require: []deptfile.Require{
				{Path: "github.com/wa2/kazusa"},
				{Path: "github.com/wa2/setsuna"},
				{Path: "github.com/wa2/haruki", Indirect: true},
			},
		}

		cases := map[string]struct {
			repo   string
			hasErr bool
		}{
			"normal":        {repo: "github.com/wa2/kazusa"},
			"normal with /": {repo: "github.com/wa2/kazusa/"},
			"indirection package is not able to remove": {repo: "github.com/wa2/haruki"},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				cleanup := setup(m)
				defer cleanup()

				mockUI := newMockUI()
				mockGoCMD := &gocmd.CommandMock{
					ModTidyFunc: func(ctx context.Context) error {
						return nil
					},
				}
				workspace := &deptfile.Workspace{SourcePath: "testdata"}
				cmd := cmd.NewRemove(mockUI, mockGoCMD, workspace)

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
