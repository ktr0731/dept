package deptfile

func ChangeDeptfileName(name string) func() {
	old := deptfileName
	deptfileName = name
	return func() {
		deptfileName = old
	}
}
