package cmd

import (
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
		fmt.Fprintf(&b, "    -%s", f.Name)
		name, _ := flag.UnquoteUsage(f)
		if name != "" {
			fmt.Fprintf(&b, " %s", name)
		} else {
			b.WriteString("\t")
		}
		fmt.Fprintf(&b, "\t%s\n", f.Usage)
	})
	return b.String()
}

func normalizeRepo(path string) (repo, ver string, err error) {
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
