package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/gocmd"
	"github.com/ktr0731/dept/logger"
	"github.com/mitchellh/cli"
)

const (
	appName    = "dept"
	appVersion = "0.1.0"
)

var minVersion = version.Must(version.NewVersion("1.11.0"))

var (
	stdin  io.Reader = os.Stdin
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

// Run receives args, then run dept.
func Run(args []string) (int, error) {
	ver := version.Must(version.NewVersion(strings.TrimPrefix(runtime.Version(), "go")))
	if !isCompatibleGoVersion(ver) {
		return 1, errors.New("dept needs Go 1.11 or later")
	}

	if isLimitedGoModSupport(ver) {
		const key = "GO111MODULE"
		oldVal := os.Getenv(key)
		os.Setenv(key, "on")
		defer func() {
			os.Setenv(key, oldVal)
		}()
	}

	app := cli.NewCLI(appName, appVersion)

	app.Commands = map[string]cli.CommandFactory{
		"init": func() (cli.Command, error) {
			return cmd.NewInit(newUI()), nil
		},
		"get": func() (cli.Command, error) {
			return cmd.NewGet(
				newUI(),
				gocmd.New(),
				&deptfile.Workspace{},
			), nil
		},
		"remove": func() (cli.Command, error) {
			return cmd.NewRemove(
				newUI(),
				gocmd.New(),
				&deptfile.Workspace{},
			), nil
		},
		"build": func() (cli.Command, error) {
			return cmd.NewBuild(
				newUI(),
				gocmd.New(),
				&deptfile.Workspace{},
			), nil
		},
		"list": func() (cli.Command, error) {
			return cmd.NewList(
				newUI(),
				&deptfile.Workspace{},
			), nil
		},
	}

	f := flag.NewFlagSet("main", flag.ExitOnError)
	verbose := f.Bool("v", false, "verbose output")

	app.HelpWriter = os.Stdout
	app.HelpFunc = func(c map[string]cli.CommandFactory) string {
		// Replace basic help header by new one
		// because it doesn't show optional flags.
		header := fmt.Sprintf(
			"Usage: %s [-v] [--version] [--help] <command> [<args>]",
			app.Name)
		s := cli.BasicHelpFunc(app.Name)(c)
		i := strings.Index(s, "\n")
		s = strings.Replace(s, s[:i], header, 1)
		return s
	}

	f.Usage = func() { fmt.Println(app.HelpFunc(app.Commands)) }
	f.Parse(args)
	if *verbose {
		logger.SetOutput(os.Stderr)
	}

	app.Args = f.Args()
	return app.Run()
}

func newUI() cli.Ui {
	return &cli.BasicUi{
		Reader:      stdin,
		Writer:      stdout,
		ErrorWriter: stderr,
	}
}

// isCompatibleGoVersion checks whether the Go version has Go modules support.
// version must be formatted as semver form.
func isCompatibleGoVersion(v *version.Version) bool {
	return v.Compare(minVersion) >= 0
}

// isLimitedGoModSupport checks whether the Go version is limited support for Go modules.
// If Go modules incompatible version passed, isLimitedGoModSupport will panic.
func isLimitedGoModSupport(v *version.Version) bool {
	if !isCompatibleGoVersion(v) {
		panic("Go modules incompatible version passed")
	}
	minor := v.Segments()[1]
	return minor == 11
}
