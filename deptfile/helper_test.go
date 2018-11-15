package deptfile_test

import "github.com/ktr0731/dept/deptfile"

func changeDeptfileName(name string) func() {
	old := deptfile.DeptfileName
	deptfile.DeptfileName = name
	return func() {
		deptfile.DeptfileName = old
	}
}
