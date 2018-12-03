package cmd_test

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/ktr0731/dept/cmd"
	"github.com/mitchellh/cli"
)

type mockUI struct {
	*cli.BasicUi
}

func (u *mockUI) Writer() *bytes.Buffer {
	return u.BasicUi.Writer.(*bytes.Buffer)
}

func (u *mockUI) ErrorWriter() *bytes.Buffer {
	return u.BasicUi.ErrorWriter.(*bytes.Buffer)
}

func newMockUI() *mockUI {
	return &mockUI{
		&cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      new(bytes.Buffer),
			ErrorWriter: new(bytes.Buffer),
		},
	}
}

func TestExcludeFlagUsage(t *testing.T) {
	f := flag.NewFlagSet("test", flag.ExitOnError)
	f.Bool("foo", false, "")
	f.Bool("bar", false, "")
	f.Bool("baz", false, "")
	excludes := []string{"bar"}
	actual := cmd.ExcludeFlagUsage(f, false, excludes)
	if strings.Contains(actual, "bar") {
		t.Errorf("ExcludeFlagUsage must hide flag 'bar', but found:\n%s", actual)
	}
}
