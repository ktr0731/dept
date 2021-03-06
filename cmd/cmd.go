package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/logger"
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

func run(c command, f func(ctx context.Context) error) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		logger.Println("interrupted")
		cancel()
		logger.Println("context canceled")
	}()

	err := f(ctx)
	if err == nil {
		return 0
	}

	switch errors.Cause(err) {
	case errShowHelp:
		c.UI().Output(c.Help())
	case context.Canceled:
		logger.Println("command successfully finished")
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
		fmt.Fprintf(&b, "\t%s", f.Usage)
		if f.DefValue != "" {
			fmt.Fprintf(&b, ` (default "%s")`, f.DefValue)
		}
		fmt.Fprint(&b, "\n")
	})
	return b.String()
}

// ExcludeFlagUsage is the same as FlagUsage, but excludes a slice passed as exclude.
// Note that ExcludeFlagUsage must be called before Parse.
func ExcludeFlagUsage(f *flag.FlagSet, repeatable bool, exclude []string) string {
	var excludes = make(map[string]interface{}, len(exclude))
	for _, f := range exclude {
		excludes[f] = nil
	}
	newOne := flag.NewFlagSet(f.Name(), f.ErrorHandling())
	f.VisitAll(func(f *flag.Flag) {
		if _, found := excludes[f.Name]; !found {
			newOne.Var(f.Value, f.Name, f.Usage)
		}
	})
	return FlagUsage(newOne, repeatable)
}

// normalizePath normalizes a passed path.
// It trims any schemes like 'https://'.
// Also, it parse the module version from path.
// For example,
//   'https://github.com/ktr0731/itunes-cli/itunes@latest'
//     Repo = 'github.com/ktr0731/itunes-cli/itunes'
//     Ver  = '@latest'
func normalizePath(path string) (repo, ver string, err error) {
	var u *url.URL
	u, err = url.Parse(path)
	if err != nil {
		return "", "", errors.Wrap(err, "invalid Repo passed")
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
// Note that, version are ignored.
func forTools(r *deptfile.Require, f func(path string) bool) {
	forToolsWithOutputName(r, func(path, _ string) bool {
		return f(path)
	})
}

// forToolsWithOutputName is like forTools, but also pass outputName of each tool.
// If out is empty, it means out is the same as filepath.Base(path).
func forToolsWithOutputName(r *deptfile.Require, f func(path, outputName string) bool) {
	if r == nil {
		return
	}
	for _, t := range r.ToolPaths {
		p := r.Path
		if t.Path != "/" {
			p += t.Path
		}
		if ok := f(p, t.Name); !ok {
			return
		}
	}
}

func resolveOutputDir(projRoot, flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if b := os.Getenv("GOBIN"); b != "" {
		logger.Printf("output dir = $GOBIN (%s)", b)
		return b
	} else {
		p := filepath.Join(projRoot, "_tools")
		logger.Printf("output dir = %s", p)
		return p
	}
}
