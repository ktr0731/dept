package cmd_test

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
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
	doNothing := func(f func(projectDir string, gomod *deptfile.File) error) error { return f("", nil) }
	emptyReader := strings.NewReader("")

	assertBuild := func(t *testing.T, expected *deptfile.Require, cmd *gocmd.CommandMock) {
		if n := len(cmd.BuildCalls()); n != 1 {
			t.Fatalf("Build must be called once, but actual %d", n)
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
			DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
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

	t.Run("Run returns 1 because some flags put after args", func(t *testing.T) {
		mockUI := newMockUI()
		mockGoCMD := &gocmd.CommandMock{
			ListFunc: func(ctx context.Context, args ...string) (io.Reader, error) {
				return emptyReader, nil
			},
		}
		mockWorkspace := &deptfile.WorkspacerMock{
			DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
				return deptfile.ErrNotFound
			},
		}

		cmd := cmd.NewGet(mockUI, mockGoCMD, mockWorkspace)

		code := cmd.Run([]string{"github.com/ktr0731/go-modules-test", "-o", "foo", "bar"})

		if code != 1 {
			t.Errorf("Run must return code 1, but got %d", code)
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
			"update a tool with renaming": {
				loadedTools: []*deptfile.Require{
					{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/"}}},
				},
				args:            []string{"-o", "ev", "github.com/ktr0731/evans"},
				root:            "github.com/ktr0731/evans",
				expectedRequire: &deptfile.Require{Path: "github.com/ktr0731/evans", ToolPaths: []*deptfile.Tool{{Path: "/", Name: "ev"}}},
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
					DoFunc: func(f func(projectDir string, df *deptfile.File) error) error {
						df := &deptfile.File{Require: c.loadedTools}
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
					if n := len(mockGoCMD.GetCalls()); n < 2 {
						t.Errorf("Get must be called twice or more, but actual %d", n)
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
			DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
				return f("", &deptfile.File{
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
					DoFunc: func(f func(projectDir string, gomod *deptfile.File) error) error {
						return f("", &deptfile.File{
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
	path string
	out  string
}

func TestOutputFlagValue(t *testing.T) {
	cases := []struct {
		args               []string
		expectedParsedArgs []*path
		expectedArgs       []string
		hasErr             bool
	}{
		{
			args: []string{"-o", "wa2", "github.com/leaf/whitealbum2"},
			expectedParsedArgs: []*path{
				{
					path: "github.com/leaf/whitealbum2",
					out:  "wa2",
				},
			},
		},
		{
			args: []string{"-o", "wa2", "github.com/leaf/whitealbum2@v1.1.0"},
			expectedParsedArgs: []*path{
				{
					path: "github.com/leaf/whitealbum2@v1.1.0",
					out:  "wa2",
				},
			},
		},
		{
			args:         []string{"github.com/leaf/whitealbum2@v1.1.0"},
			expectedArgs: []string{"github.com/leaf/whitealbum2@v1.1.0"},
		},
		{
			args: []string{"-o", "wa2", "github.com/leaf/whitealbum2", "-o", "ic", "github.com/leaf/whitealbum2/introductory-chapter"},
			expectedParsedArgs: []*path{
				{
					path: "github.com/leaf/whitealbum2",
					out:  "wa2",
				},
				{
					path: "github.com/leaf/whitealbum2/introductory-chapter",
					out:  "ic",
				},
			},
		},
		{
			args: []string{"-o", "wa2", "github.com/leaf/whitealbum2", "github.com/leaf/whitealbum2/introductory-chapter"},
			expectedParsedArgs: []*path{
				{
					path: "github.com/leaf/whitealbum2",
					out:  "wa2",
				},
			},
			expectedArgs: []string{"github.com/leaf/whitealbum2/introductory-chapter"},
		},
		{
			args:         []string{"github.com/leaf/whitealbum2"},
			expectedArgs: []string{"github.com/leaf/whitealbum2"},
		},
		{args: []string{"-o"}, hasErr: true},
		{args: []string{"-o", "github.com/leaf/whitealbum2"}, hasErr: true},
		{args: []string{"-o", "wa2", "github.com/leaf/whitealbum2", "-o"}, hasErr: true},
		{args: []string{"-o", "wa2", "github.com/leaf/whitealbum2", "-o", "foo"}, hasErr: true},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.args, " "), func(t *testing.T) {
			f := flag.NewFlagSet("test", flag.ContinueOnError)
			f.SetOutput(ioutil.Discard)
			v := cmd.NewOutputFlagValue(f)
			f.Var(v, "o", "")
			err := f.Parse(c.args)

			if c.hasErr {
				if err == nil {
					t.Error("outputFlagValue must return an error, but got nil")
				}
				return
			} else {
				if err != nil {
					t.Fatalf("outputFlagValue must not return any errors, but got %s", err)
				}
			}
			pargs := v.Values
			if len(c.expectedParsedArgs) != len(pargs) {
				t.Fatalf("number of expected args and vals must be equal, but %d and %d", len(c.expectedParsedArgs), len(pargs))
			}
			args := f.Args()
			if len(c.expectedArgs) != len(args) {
				t.Fatalf("number of expected args and vals must be equal, but %d and %d", len(c.expectedParsedArgs), len(args))
			}
			for i, arg := range c.expectedParsedArgs {
				if arg.path != pargs[i].Path {
					t.Errorf("path is wrong: expected = %s, actual = %s", arg.path, pargs[i].Path)
				}
				if arg.out != pargs[i].Out {
					t.Errorf("out is wrong: expected = %s, actual = %s", arg.out, pargs[i].Out)
				}
			}
			for i, path := range c.expectedArgs {
				if path != args[i] {
					t.Errorf("path is wrong: expected = %s, actual = %s", path, args[i])
				}
			}
		})
	}
}
