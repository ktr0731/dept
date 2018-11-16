package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/ktr0731/dept/app"
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

	code, err := app.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run dept: %s\n", err)
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
