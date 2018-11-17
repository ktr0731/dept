package cmd

import (
	"context"

	"github.com/ktr0731/dept/deptfile"
)

func ChangeDeptfileLoad(f func(context.Context) (*deptfile.GoMod, error)) func() {
	deptfileLoad = f
	return func() {
		deptfileLoad = deptfile.Load
	}
}
