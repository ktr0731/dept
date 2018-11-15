package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/ktr0731/dept/builder"
	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/fetcher"
	"github.com/mitchellh/cli"
)

func main() {
	ver := version.Must(version.NewVersion(strings.TrimPrefix(runtime.Version(), "go")))
	if !isCompatibleGoVersion(ver) {
		fmt.Fprintf(os.Stderr, "dept needs Go 1.11 or later")
		os.Exit(1)
	}

	if isLimitedGoModSupport(ver) {
		const key = "GO111MODULE"
		oldVal := os.Getenv(key)
		os.Setenv(key, "on")
		defer func() {
			os.Setenv(key, oldVal)
		}()
	}

	app := cli.NewCLI("dept", "0.1.0")
	app.Args = os.Args[1:]

	// TODO: move
	df, err := deptfile.Load()
	if app.Subcommand() != "init" && err == deptfile.ErrNotFound {
		fmt.Fprintf(os.Stderr, "deptfile missing. please do 'dept init'\n")
		os.Exit(1)
	} else if app.Subcommand() != "init" && err != nil {
		log.Fatalln(err)
	}

	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	fetcher := fetcher.New()
	builder := builder.New()

	app.Commands = map[string]cli.CommandFactory{
		"init": cmd.Init(ui),
		"get":  cmd.Get(ui, fetcher, builder, df),
	}
	code, err := app.Run()
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(code)
}

var minVersion = version.Must(version.NewVersion("1.11.0"))

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
