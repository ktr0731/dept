package cmd

import (
	"errors"
	"flag"
	"fmt"
	"strings"

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

// FlagUsage gets available flags and usage from f.
func FlagUsage(f *flag.FlagSet) string {
	var b strings.Builder
	b.WriteString("Available flags are:\n")
	f.VisitAll(func(f *flag.Flag) {
		name, _ := flag.UnquoteUsage(f)
		fmt.Fprintf(&b, "    -%s %s   %s\n", f.Name, name, f.Usage)
	})
	return b.String()
}
