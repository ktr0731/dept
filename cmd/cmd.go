package cmd

import (
	"bytes"
	"errors"
	"flag"

	"github.com/ktr0731/dept/deptfile"
	"github.com/mitchellh/cli"
)

var (
	errShowHelp = errors.New("show help")
)

type command interface {
	cli.Command
	UI() cli.Ui
}

func run(c command, f func() error) int {
	err := f()
	if err == nil {
		return 0
	}

	switch err {
	case errShowHelp:
		c.UI().Output(c.Help())
	case deptfile.ErrNotFound:
		c.UI().Error("deptfile missing. please do 'dept init'")
	default:
		c.UI().Error(err.Error())
	}
	return 1
}

// flagUsage gets available flags and usage from f.
func flagUsage(f *flag.FlagSet) string {
	var buf bytes.Buffer
	old := f.Output()
	f.SetOutput(&buf)
	f.PrintDefaults()
	f.SetOutput(old)
	return buf.String()
}
