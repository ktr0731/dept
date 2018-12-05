package cmd

import (
	"flag"
)

func NewOutputFlagValue(f *flag.FlagSet) *outputFlagValue {
	return &outputFlagValue{Values: []struct{ Out, Path string }{}, f: f}
}

func NormalizePath(path string) (repo, ver string, err error) {
	return normalizePath(path)
}
