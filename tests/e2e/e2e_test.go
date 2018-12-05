// +build e2e

package e2e

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ktr0731/dept/app"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type testcase struct {
	name   string
	args   []string
	code   int
	assert func(t *testing.T, out, eout *bytes.Buffer)
}

func setupOutput() (*bytes.Buffer, *bytes.Buffer, func()) {
	var out, eout bytes.Buffer
	app.SetStdout(&out)
	app.SetStderr(&eout)
	return &out, &eout, func() {
		app.SetStdout(os.Stdout)
		app.SetStderr(os.Stderr)
	}
}

func setupEnv(t *testing.T) func() {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %s", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get the current working dir: %s", err)
	}
	os.Chdir(dir)
	return func() {
		os.Chdir(cwd)
		os.RemoveAll(dir)
	}
}

func TestGet(t *testing.T) {
	cleanup := setupEnv(t)
	defer cleanup()

	cases := []testcase{
		{
			name: "fail to get because gotool.mod missing",
			args: []string{"get", "github.com/ktr0731/salias"},
			code: 1,
		},
		{
			name: "create a new gotool.mod",
			args: []string{"init"},
		},
		{
			name: "get a new tool",
			args: []string{"get", "github.com/ktr0731/salias"},
		},
		{
			name: "list tools",
			args: []string{"list"},
			assert: func(t *testing.T, out, eout *bytes.Buffer) {
				if !strings.Contains(out.String(), "github.com/ktr0731/salias") {
					t.Errorf("list must be list up 'github.com/ktr0731/salias', but missing:\n%s", out.String())
				}
			},
		},
		{
			name: "add a new tool again but it missing",
			args: []string{"get", "github.com/hashicorp/gox"},
			code: 1,
		},
		{
			name: "add a new tool again with specified version",
			args: []string{"get", "github.com/mitchellh/gox@v0.2.0"},
		},
		{
			name: "list tools again",
			args: []string{"list"},
			assert: func(t *testing.T, out, eout *bytes.Buffer) {
				if !strings.Contains(out.String(), "github.com/ktr0731/salias") {
					t.Errorf("list must be list up 'github.com/ktr0731/salias', but missing:\n%s", out.String())
				}
				if !strings.Contains(out.String(), "github.com/mitchellh/gox") {
					t.Errorf("list must be list up 'github.com/mitchellh/gox', but missing:\n%s", out.String())
				}
			},
		},
		{
			name: "upgrade gox to v0.3.0",
			args: []string{"get", "github.com/mitchellh/gox@v0.3.0"},
		},
		{
			name: "gox is upgraded to v0.3.0",
			args: []string{"list", "github.com/mitchellh/gox"},
			assert: func(t *testing.T, out, eout *bytes.Buffer) {
				s := strings.TrimSpace(out.String())
				sp := strings.Split(s, " ")
				if sp[len(sp)-1] != "v0.3.0" {
					t.Errorf("gox must be updated to v0.3.0, but %s", sp[len(sp)-1])
				}
			},
		},
		{
			name: "upgrade gox to the latest version",
			args: []string{"get", "-u", "github.com/mitchellh/gox"},
		},
		{
			name: "gox is updated to the latest version",
			args: []string{"list", "github.com/mitchellh/gox"},
			assert: func(t *testing.T, out, eout *bytes.Buffer) {
				s := strings.TrimSpace(out.String())
				sp := strings.Split(s, " ")
				if sp[len(sp)-1] == "v0.3.0" {
					t.Error("gox must be updated to the latest version, but v0.3.0")
				}
			},
		},
		{
			name: "downgrade gox to v0.2.0",
			args: []string{"get", "github.com/mitchellh/gox@v0.2.0"},
		},
		{
			name: "gox is downgraded to v0.2.0",
			args: []string{"list", "github.com/mitchellh/gox"},
			assert: func(t *testing.T, out, eout *bytes.Buffer) {
				s := strings.TrimSpace(out.String())
				sp := strings.Split(s, " ")
				if sp[len(sp)-1] != "v0.2.0" {
					t.Errorf("gox must be downgraded to v0.2.0, but %s", sp[len(sp)-1])
				}
			},
		},
		{
			name: "remove uninstalled tool",
			args: []string{"remove", "github.com/wa2/kazusa"},
			code: 1,
		},
		{
			name: "upgrade all tools",
			args: []string{"get", "-u"},
		},
		{
			name: "remove gox",
			args: []string{"remove", "github.com/mitchellh/gox"},
		},
		{
			name: "gox is uninstalled",
			args: []string{"list", "github.com/mitchellh/gox"},
			assert: func(t *testing.T, out, eout *bytes.Buffer) {
				if strings.Contains(out.String(), "github.com/mitchellh/gox") {
					t.Errorf("list must not be list up 'github.com/mitchellh/gox':\n%s", out.String())
				}
			},
		},
		{
			name: "add two new tools with renaming",
			args: []string{
				"get",
				"-d", "_tools",
				"-o", "it", "github.com/ktr0731/itunes-cli/itunes",
				"-o", "gc", "github.com/fullstorydev/grpcurl/cmd/grpcurl",
			},
			assert: func(t *testing.T, out, eout *bytes.Buffer) {
				if strings.Contains(out.String(), "github.com/ktr0731/itunes-cli/itunes") {
					t.Errorf("list must not be list up 'github.com/ktr0731/itunes-cli/itunes':\n%s", out.String())
				}
				if strings.Contains(out.String(), "github.com/fullstorydev/grpcurl/cmd/grpcurl") {
					t.Errorf("list must not be list up 'github.com/fullstorydev/grpcurl/cmd/grpcurl':\n%s", out.String())
				}
				_, err := os.Stat(filepath.Join("_tools", "it"))
				if os.IsNotExist(err) {
					t.Error("itunes must be installed as 'it', but not found")
				}
				_, err = os.Stat(filepath.Join("_tools", "gc"))
				if os.IsNotExist(err) {
					t.Error("grpcurl must be installed as 'gc', but not found")
				}
			},
		},
		{
			name: "build all tools",
			args: []string{"build", "-d", "bin"},
			assert: func(t *testing.T, out, eout *bytes.Buffer) {
				if _, err := os.Stat("bin"); os.IsNotExist(err) {
					t.Error("build must write out binaries to bin, but dir not found")
				}
				_, err := os.Stat(filepath.Join("_tools", "it"))
				if os.IsNotExist(err) {
					t.Error("itunes must be installed as 'it', but not found")
				}
				_, err = os.Stat(filepath.Join("_tools", "gc"))
				if os.IsNotExist(err) {
					t.Error("grpcurl must be installed as 'gc', but not found")
				}
				_, err = os.Stat(filepath.Join("bin", "salias"))
				if os.IsNotExist(err) {
					t.Error("salias must be installed, but not found")
				}
			},
		},
	}

	for _, c := range cases {
		do(t, c)
		if t.Failed() {
			return
		}
	}
}

func TestGetExamples(t *testing.T) {
	cleanup := setupEnv(t)
	defer cleanup()

	out, _, cleanupOutput := setupOutput()
	defer cleanupOutput()

	code, err := app.Run([]string{"get", "-h"})
	if code != 0 {
		t.Fatalf("expected code = 0, but got %d", code)
	}
	if err != nil {
		t.Fatalf("failed to get help message: %s", err)
	}
	code, err = app.Run([]string{"init"})
	if code != 0 {
		t.Fatalf("expected code = 0, but got %d", code)
	}
	if err != nil {
		t.Fatalf("failed to init deptfile: %s", err)
	}
	re := regexp.MustCompile(`.*\$.*dept\s(.*)$`)
	s := bufio.NewScanner(out)
	for s.Scan() {
		str := strings.TrimSpace(s.Text())
		sub := re.FindStringSubmatch(str)
		if sub == nil {
			continue
		}
		cmd := strings.Split(sub[1], " ")
		do(t, testcase{
			name: str,
			args: cmd,
		})
	}
}

func do(t *testing.T, c testcase) {
	var code int
	var err error
	out, eout, cleanup := setupOutput()
	defer cleanup()
	defer func() func() {
		fmt.Printf("   --- RUN : %s", c.name)
		return func() {
			if code == c.code {
				fmt.Printf("\r   --- PASS: %s\n", c.name)
			} else {
				fmt.Printf("\r   --- FAIL: %s\n", c.name)
				fmt.Printf("       expected code = %d, but got %d\n", c.code, code)
				if err != nil {
					fmt.Println(err.Error())
				}
				if eout.String() != "" {
					fmt.Println(eout.String())
				}
				t.Fail()
			}
		}
	}()()

	code, err = app.Run(c.args)
	if err != nil {
		return
	}

	if c.assert != nil {
		c.assert(t, out, eout)
		if t.Failed() {
			if err != nil {
				fmt.Println(err.Error())
			}
			if eout.String() != "" {
				fmt.Println(eout.String())
			}
			t.Fail()
		}
	}
}
