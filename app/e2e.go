// +build e2e

package app

import "io"

func SetStdout(out io.Writer) {
	stdout = out
}

func SetStderr(eout io.Writer) {
	stderr = eout
}
