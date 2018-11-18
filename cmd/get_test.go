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
	"github.com/pkg/errors"
)

func TestGetRun(t *testing.T) {
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
		cmd := cmd.NewGet(mockUI, nil, workspace)

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

		cmd := cmd.NewGet(mockUI, nil, workspace)

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
			m      *deptfile.GoMod
			args   []string
			update bool
		}{
			"get a new tool": {args: []string{"github.com/ktr0731/evans"}},
			"update a tool": {
				m: &deptfile.GoMod{
					Require: []deptfile.Require{
						{Path: "github.com/ktr0731/evans"},
					},
				},
				args:   []string{"github.com/ktr0731/evans"},
				update: true,
			},
			"-u also works if the specified tool is not found": {args: []string{"github.com/ktr0731/evans"}, update: true},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				cleanup := setup(c.m)
				defer cleanup()

				mockUI := newMockUI()
				mockGoCMD := &gocmd.CommandMock{
					GetFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
					BuildFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
				}
				workspace := &deptfile.Workspace{SourcePath: "testdata"}
				cmd := cmd.NewGet(mockUI, mockGoCMD, workspace)

				var args []string
				if c.update {
					args = append([]string{"-u"}, c.args...)
				} else {
					args = c.args
				}
				code := cmd.Run(args)
				if code != 0 {
					t.Errorf("Run must return 0, but got %d (err = %s)", code, mockUI.ErrorWriter().String())
				}

				if c.update {
					if n := len(mockGoCMD.GetCalls()); n != 2 {
						t.Errorf("Get must be called twice, but actual %d", n)
					}
				} else {
					if n := len(mockGoCMD.GetCalls()); n != 1 {
						t.Errorf("Get must be called once, but actual %d", n)
					}
				}

				if n := len(mockGoCMD.BuildCalls()); n != 1 {
					t.Errorf("Build must be called once, but actual %d", n)
				}
			})
		}
	})

	t.Run("deptfile is not modified when command failed", func(t *testing.T) {
		mockUI := newMockUI()
		mockGoCMD := &gocmd.CommandMock{
			GetFunc: func(ctx context.Context, pkgs ...string) error {
				return errors.New("an error")
			},
		}
		workspace := &deptfile.Workspace{SourcePath: "testdata"}

		cleanup := setup(nil)
		defer cleanup()

		cmd := cmd.NewGet(mockUI, mockGoCMD, workspace)

		repo := "github.com/ktr0731/go-modules-test"
		code := cmd.Run([]string{repo})
		if code != 1 {
			t.Errorf("Run must be 1, but got %d", code)
		}

		if n := len(mockGoCMD.GetCalls()); n != 1 {
			t.Errorf("Get must be called once, but actual %d (err = %s)", n, mockUI.ErrorWriter().String())
		}
	})

	t.Run("Run returns an error when tool names conflicted", func(t *testing.T) {
		mockUI := newMockUI()
		workspace := &deptfile.Workspace{SourcePath: "testdata"}

		m := &deptfile.GoMod{
			Require: []deptfile.Require{
				{Path: "github.com/ktr0731/evans"},
			},
		}
		cleanup := setup(m)
		defer cleanup()

		cmd := cmd.NewGet(mockUI, nil, workspace)

		repo := "github.com/ktr0732/evans"
		code := cmd.Run([]string{repo})
		if code != 1 {
			t.Errorf("Run must be 1, but got %d", code)
		}

		if mockUI.ErrorWriter().String() == "" {
			t.Error("Run must output an error message, but empty")
		}
	})
}
