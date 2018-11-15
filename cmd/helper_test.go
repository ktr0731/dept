package cmd

import "github.com/ktr0731/dept/deptfile"

func ChangeDeptfileLoad(f func() (*deptfile.File, error)) func() {
	deptfileLoad = f
	return func() {
		deptfileLoad = deptfile.Load
	}
}
