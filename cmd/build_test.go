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
	"github.com/ktr0731/dept/toolcacher"
)

func TestBuildRun(t *testing.T) {
	t.Run("Run returns 1 because gotool.mod is not found", func(t *testing.T) {
		mockUI := newMockUI()
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
				return deptfile.ErrNotFound
			},
		}

		cmd := cmd.NewBuild(mockUI, nil, mockWorkspace, nil)

		code := cmd.Run(nil)

		if code != 1 {
			t.Errorf("Run must return code 1, but got %d", code)
		}
		if n := len(mockWorkspace.DoCalls()); n != 1 {
			t.Errorf("Do must be called only once, but %d", n)
		}
		if eout := mockUI.ErrorWriter().String(); !strings.Contains(eout, "dept init") {
			t.Errorf("Run must show 'dept init' related error message in case of gotool.mod is not found, but '%s'", eout)
		}
	})

	t.Run("Run returns code 0 normally", func(t *testing.T) {
		cases := map[string]struct {
			loadedTools []*deptfile.Require
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
			},
		}

		setup := func(t *testing.T) func() {
			dir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("failed to create a temp dir: %s", err)
			}
			old := os.Getenv("GOBIN")
			os.Setenv("GOBIN", dir)
			return func() {
				os.Setenv("GOBIN", old)
				os.RemoveAll(dir)
			}
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				cleanup := setup(t)
				defer cleanup()

				mockUI := newMockUI()
				mockGoCMD := &gocmd.CommandMock{
					ModDownloadFunc: func(ctx context.Context) error { return nil },
				}
				mockWorkspace := &deptfile.WorkspacerMock{
					DoFunc: func(f func(projectDir string, df *deptfile.File) error) error {
						df := &deptfile.File{Require: c.loadedTools}
						return f("", df)
					},
				}

				f, err := ioutil.TempFile("", "")
				if err != nil {
					t.Fatal(err, "failed to create a temp file")
				}
				defer f.Close()
				mockToolCacher := &toolcacher.CacherMock{
					GetFunc: func(ctx context.Context, pkgName string, version string) (string, error) {
						return f.Name(), nil
					},
				}
				cmd := cmd.NewBuild(mockUI, mockGoCMD, mockWorkspace, mockToolCacher)

				code := cmd.Run(nil)
				if code != 0 {
					t.Errorf("Run must return 0, but got %d (err = %s)", code, mockUI.ErrorWriter().String())
				}

				if n := len(mockToolCacher.GetCalls()); n != len(c.loadedTools) {
					t.Errorf("Get must be called %d times, but actual %d", len(c.loadedTools), n)
				}
			})
		}
	})
}
