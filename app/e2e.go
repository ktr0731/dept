// +build e2e

package app

import "io"

func SetStdin(in io.Reader) {
	stdin = in
}

func SetStdout(out io.Writer) {
	stdout = out
}

func SetStderr(eout io.Writer) {
	stderr = eout
}
