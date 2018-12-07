// +build e2e

package cmd

import (
	"bytes"
	"os/exec"
	"syscall"
)

func ChangeSyscallExec(stdout, stderr *bytes.Buffer) func() {
	syscallExec = func(argv0 string, argv []string, envv []string) (err error) {
		cmd := exec.Command(argv0, argv[1:]...)
		cmd.Env = envv
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	}
	return func() {
		syscallExec = syscall.Exec
	}
}
