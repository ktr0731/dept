package cmd

import (
	"flag"
	"syscall"
)

func NewOutputFlagValue(f *flag.FlagSet) *outputFlagValue {
	return &outputFlagValue{Values: []struct{ Out, Path string }{}, f: f}
}

func NormalizePath(path string) (repo, ver string, err error) {
	return normalizePath(path)
}

func ChnageSyscallExec(f func(argv0 string, argv []string, envv []string) (err error)) func() {
	syscallExec = f
	return func() {
		syscallExec = syscall.Exec
	}
}
