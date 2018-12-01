package cmd_test

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/gocmd"
	"github.com/pkg/errors"
)

func TestGetRun(t *testing.T) {
	doNothing := func(f func(projectDir string, gomod *deptfile.GoMod) error) error { return f("", nil) }
	emptyReader := strings.NewReader("")

	assertBuild := func(t *testing.T, expected *deptfile.Require, cmd *gocmd.CommandMock) {
		if n := len(cmd.BuildCalls()); n != 1 {
			t.Errorf("Build must be called once, but actual %d", n)
		}

		buildArgs := cmd.BuildCalls()[0].Args
		var actualName string
		if buildArgs[0] == "-o" {
			actualName = filepath.Base(buildArgs[1])
		} else {
			actualName = filepath.Base(buildArgs[0])
		}
		if actualName == "_tools" {
			t.Errorf("invalid output name: %s", actualName)
		}
	}

	t.Run("Run returns code 1 because no arguments passed", func(t *testing.T) {
		mockUI := newMockUI()
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: doNothing,
		}
		cmd := cmd.NewGet(mockUI, nil, mockWorkspace)

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
		mockGoCMD := &gocmd.CommandMock{
			ListFunc: func(ctx context.Context, args ...string) (io.Reader, error) {
				return emptyReader, nil
			},
		}
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: func(f func(projectDir string, gomod *deptfile.GoMod) error) error {
				return deptfile.ErrNotFound
			},
		}

		cmd := cmd.NewGet(mockUI, mockGoCMD, mockWorkspace)

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
			loadedTools     []*deptfile.Require
			root            string
			args            []string
			expectedRequire *deptfile.Require
			update          bool
		}{
			"get a new tool": {
				args:            []string{"github.com/ktr0731/evans"},
				root:            "github.com/ktr0731/evans",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
			},
			"get a new tool with HTTP scheme": {
				args:            []string{"https://github.com/ktr0731/evans"},
				root:            "github.com/ktr0731/evans",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
			},
			"get a new tool that is not in the module root": {
				args:            []string{"github.com/ktr0731/itunes-cli/itunes"},
				root:            "github.com/ktr0731/itunes-cli",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/itunes-cli", ToolPaths: []*deptfile.Tool{{Path: "/itunes"}}},
			},
			"add a new tool that is not in the module root": {
				loadedTools: []*deptfile.Require{{Path: "honnef.co/go/tools", ToolPaths: []*deptfile.Tool{{Path: "/cmd/unused"}}}},
				args:        []string{"honnef.co/go/tools/cmd/staticcheck"},
				root:        "honnef.co/go/tools",
				expectedRequire: &deptfile.Require{
					Path: "honnef.co/go/tools",
					ToolPaths: []*deptfile.Tool{
						{Path: "/cmd/unused"},
						{Path: "/cmd/staticcheck"},
					},
				},
			},
			"add a new tool that is in the module root, but has command path": {
				loadedTools: []*deptfile.Require{{Path: "github.com/foo/bar", ToolPaths: []*deptfile.Tool{{Path: "/cmd/baz"}}}},
				args:        []string{"github.com/foo/bar"},
				root:        "github.com/foo/bar",
				expectedRequire: &deptfile.Require{
					Path: "github.com/foo/bar",
					ToolPaths: []*deptfile.Tool{
						{Path: "/"},
						{Path: "/cmd/baz"},
					},
				},
			},
			"add a new tool that is not in the module root, but has top-level command": {
				loadedTools: []*deptfile.Require{
					{Path: "github.com/foo/bar", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
				},
				args: []string{"github.com/foo/bar/cmd/baz"},
				root: "github.com/foo/bar",
				expectedRequire: &deptfile.Require{
					Path: "github.com/foo/bar",
					ToolPaths: []*deptfile.Tool{
						{Path: "/"},
						{Path: "/cmd/baz"},
					},
				},
			},
			"add a duplicated tool": {
				loadedTools: []*deptfile.Require{
					{Path: "github.com/foo/bar", ToolPaths: []*deptfile.Tool{{Path: "/cmd/baz"}}},
				},
				args: []string{"github.com/foo/bar/cmd/baz"},
				root: "github.com/foo/bar",
				expectedRequire: &deptfile.Require{
					Path:      "github.com/foo/bar",
					ToolPaths: []*deptfile.Tool{{Path: "/cmd/baz"}},
				},
			},
			"update a tool": {
				loadedTools: []*deptfile.Require{
					{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
				},
				args:            []string{"github.com/ktr0731/evans"},
				root:            "github.com/ktr0731/evans",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
				update:          true,
			},
			"-u also works if the specified tool is not found": {
				args:            []string{"github.com/ktr0731/evans"},
				root:            "github.com/ktr0731/evans",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
				update:          true,
			},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				mockUI := newMockUI()
				mockGoCMD := &gocmd.CommandMock{
					GetFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
					BuildFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
					ListFunc: func(ctx context.Context, args ...string) (io.Reader, error) {
						return strings.NewReader(c.root), nil
					},
				}
				mockWorkspace := &deptfile.WorkspacerMock{
					DoFunc: func(f func(projectDir string, df *deptfile.GoMod) error) error {
						df := &deptfile.GoMod{Require: c.loadedTools}
						if err := f("", df); err != nil {
							return err
						}

						if n := len(df.Require); n != 1 {
							t.Fatalf("expected only 1 tool, but %d", n)
						}
						if diff := cmp.Diff(c.expectedRequire, df.Require[0]); diff != "" {
							t.Fatalf("modified deptfile.Require is not equal to the expected one:\n%s", diff)
						}
						return nil
					},
				}
				cmd := cmd.NewGet(mockUI, mockGoCMD, mockWorkspace)

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

				assertBuild(t, c.expectedRequire, mockGoCMD)
			})
		}
	})

	t.Run("deptfile is not modified when command failed", func(t *testing.T) {
		mockUI := newMockUI()
		mockGoCMD := &gocmd.CommandMock{
			ListFunc: func(ctx context.Context, args ...string) (io.Reader, error) {
				return nil, errors.New("an error")
			},
		}
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: func(f func(projectDir string, gomod *deptfile.GoMod) error) error {
				return f("", &deptfile.GoMod{
					Require: []*deptfile.Require{},
				})
			},
		}

		cmd := cmd.NewGet(mockUI, mockGoCMD, mockWorkspace)

		repo := "github.com/ktr0731/go-modules-test"
		code := cmd.Run([]string{repo})
		if code != 1 {
			t.Errorf("Run must be 1, but got %d", code)
		}

		if n := len(mockGoCMD.ListCalls()); n != 1 {
			t.Errorf("List must be called once, but actual %d (err = %s)", n, mockUI.ErrorWriter().String())
		}
	})

	t.Run("Run returns an error when tool names conflicted", func(t *testing.T) {
		cases := map[string]struct {
			hasErr  bool
			require string
		}{
			"same tools":                                {hasErr: false, require: "github.com/ktr0731/evans"},
			"same tools with versions":                  {hasErr: false, require: "github.com/ktr0731/evans@v0.2.0"},
			"different tools, different names":          {hasErr: false, require: "github.com/foo/bar"},
			"different tools, same names":               {hasErr: true, require: "github.com/foo/evans"},
			"different tools, same names with versions": {hasErr: true, require: "github.com/foo/evans@v0.2.0"},
			"invalid module version syntax":             {hasErr: true, require: "github.com/foo/evans@"},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				repo, _, err := cmd.NormalizePath(c.require)
				if err != nil {
					t.Fatalf("failed to normalize c.require (%s): %s", c.require, err)
				}

				mockUI := newMockUI()
				mockGoCMD := &gocmd.CommandMock{
					GetFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
					BuildFunc: func(ctx context.Context, pkgs ...string) error {
						return nil
					},
					ListFunc: func(ctx context.Context, args ...string) (io.Reader, error) {
						return strings.NewReader(repo), nil
					},
				}
				mockWorkspace := &deptfile.WorkspacerMock{
					DoFunc: func(f func(projectDir string, gomod *deptfile.GoMod) error) error {
						return f("", &deptfile.GoMod{
							Require: []*deptfile.Require{
								{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
							},
						})
					},
				}

				cmd := cmd.NewGet(mockUI, mockGoCMD, mockWorkspace)

				code := cmd.Run([]string{c.require})
				if c.hasErr {
					if code != 1 {
						t.Errorf("Run must return 1, but got %d", code)
					}
					if mockUI.ErrorWriter().String() == "" {
						t.Error("Run must output an error message, but empty")
					}
				} else {
					if code != 0 {
						t.Errorf("Run must return 0, but got %d", code)
					}
					if eout := mockUI.ErrorWriter().String(); eout != "" {
						t.Errorf("Run must not output any error messages, but got '%s'", eout)
					}
				}
			})
		}
	})
}

// Same as cmd.path.
type path struct {
	val  string
	repo string
	ver  string
	out  string
}

func TestRepeatableFlagSet(t *testing.T) {
	cases := []struct {
		args               []string
		expectedParsedArgs []*path
		hasErr             bool
	}{
		{
			args: []string{"-o", "wa2", "github.com/leaf/whitealbum2"},
			expectedParsedArgs: []*path{
				{
					val:  "github.com/leaf/whitealbum2",
					repo: "github.com/leaf/whitealbum2",
					out:  "wa2",
				},
			},
		},
		{
			args: []string{"-o", "wa2", "github.com/leaf/whitealbum2@v1.1.0"},
			expectedParsedArgs: []*path{
				{
					val:  "github.com/leaf/whitealbum2@v1.1.0",
					ver:  "v1.1.0",
					repo: "github.com/leaf/whitealbum2",
					out:  "wa2",
				},
			},
		},
		{
			args: []string{"github.com/leaf/whitealbum2@v1.1.0"},
			expectedParsedArgs: []*path{
				{
					val:  "github.com/leaf/whitealbum2@v1.1.0",
					ver:  "v1.1.0",
					repo: "github.com/leaf/whitealbum2",
				},
			},
		},
		{
			args: []string{"-o", "wa2", "github.com/leaf/whitealbum2", "-o", "ic", "github.com/leaf/whitealbum2/introductory-chapter"},
			expectedParsedArgs: []*path{
				{
					val:  "github.com/leaf/whitealbum2",
					repo: "github.com/leaf/whitealbum2",
					out:  "wa2",
				},
				{
					val:  "github.com/leaf/whitealbum2/introductory-chapter",
					repo: "github.com/leaf/whitealbum2/introductory-chapter",
					out:  "ic",
				},
			},
		},
		{
			args: []string{"-o", "wa2", "github.com/leaf/whitealbum2", "github.com/leaf/whitealbum2/introductory-chapter"},
			expectedParsedArgs: []*path{
				{
					val:  "github.com/leaf/whitealbum2",
					repo: "github.com/leaf/whitealbum2",
					out:  "wa2",
				},
				{
					val:  "github.com/leaf/whitealbum2/introductory-chapter",
					repo: "github.com/leaf/whitealbum2/introductory-chapter",
				},
			},
		},
		{
			args: []string{"github.com/leaf/whitealbum2", "-o", "ic", "github.com/leaf/whitealbum2/introductory-chapter"},
			expectedParsedArgs: []*path{
				{
					val:  "github.com/leaf/whitealbum2",
					repo: "github.com/leaf/whitealbum2",
				},
				{
					val:  "github.com/leaf/whitealbum2/introductory-chapter",
					repo: "github.com/leaf/whitealbum2/introductory-chapter",
					out:  "ic",
				},
			},
		},
		{
			args: []string{"github.com/leaf/whitealbum2@v1.1.0", "-o", "ic", "github.com/leaf/whitealbum2/introductory-chapter@v1.0.0", "-o", "cc", "github.com/leaf/whitealbum2/closing-chapter@v1.0.1"},
			expectedParsedArgs: []*path{
				{
					val:  "github.com/leaf/whitealbum2@v1.1.0",
					repo: "github.com/leaf/whitealbum2",
					ver:  "v1.1.0",
				},
				{
					val:  "github.com/leaf/whitealbum2/introductory-chapter@v1.0.0",
					repo: "github.com/leaf/whitealbum2/introductory-chapter",
					ver:  "v1.0.0",
					out:  "ic",
				},
				{
					val:  "github.com/leaf/whitealbum2/closing-chapter@v1.0.1",
					repo: "github.com/leaf/whitealbum2/closing-chapter",
					ver:  "v1.0.1",
					out:  "cc",
				},
			},
		},
		{args: []string{"-o"}, hasErr: true},
		{args: []string{"-o", "github.com/leaf/whitealbum2"}, hasErr: true},
		{args: []string{"github.com/leaf/whitealbum2", "-o"}, hasErr: true},
		{args: []string{"github.com/leaf/whitealbum2", "-o", "ic"}, hasErr: true},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.args, " "), func(t *testing.T) {
			f := cmd.RepeatableFlagSet
			parsedArgs, err := f.Parse(c.args)
			if c.hasErr {
				if err == nil {
					t.Error("RepeatableFlagSet must return an error, but got nil")
				}
				return
			} else {
				if err != nil {
					t.Fatalf("RepeatableFlagSet must not return any errors, but got %s", err)
				}
			}
			for i, expected := range c.expectedParsedArgs {
				actual := parsedArgs[i]
				cmd.AssertPath(t, expected.val, expected.repo, expected.ver, expected.out, actual)
			}
		})
	}
}
