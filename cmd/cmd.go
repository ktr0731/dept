package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ktr0731/dept/deptfile"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
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
	case context.Canceled:
		c.UI().Error("command canceled")
	case deptfile.ErrNotFound:
		c.UI().Error("deptfile missing. please do 'dept init'")
	default:
		c.UI().Error(err.Error())
	}
	return 1
}

// FlagUsage gets available flags and usage from f.
func FlagUsage(f *flag.FlagSet, repeatable bool) string {
	var b strings.Builder
	if repeatable {
		b.WriteString("Available repeatable flags are:\n")
	} else {
		b.WriteString("Available flags are:\n")
	}
	f.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(&b, "    -%s", f.Name)
		name, _ := flag.UnquoteUsage(f)
		if name != "" {
			fmt.Fprintf(&b, " %s", name)
		} else {
			b.WriteString("\t")
		}
		fmt.Fprintf(&b, "\t%s\n", f.Usage)
	})
	if repeatable {
		b.WriteString("\nNote that all repeatable flags must be put after normal flags.\n")
	}
	return b.String()
}

// normalizePath normalizes a passed path.
// It trims any schemes like 'https://'.
// Also, it parse the module version from path.
// For example,
//   'https://github.com/ktr0731/itunes-cli/itunes@latest'
//     repo = 'github.com/ktr0731/itunes-cli/itunes'
//     ver  = '@latest'
func normalizePath(path string) (repo, ver string, err error) {
	var u *url.URL
	u, err = url.Parse(path)
	if err != nil {
		return "", "", errors.Wrap(err, "invalid repo passed")
	}

	path = filepath.Clean(u.Host + u.Path)

	if i := strings.Index(path, "@"); i != -1 {
		repo = path[:i]
		ver = path[i+1:]
	} else {
		repo = path
	}
	return
}

// forTools iterates r, then pass each tool path to f.
func forTools(r *deptfile.Require, f func(path string) bool) {
	if len(r.CommandPath) == 0 {
		f(r.Path)
		return
	}
	for _, t := range r.CommandPath {
		if ok := f(r.Path + t); !ok {
			return
		}
	}
}
