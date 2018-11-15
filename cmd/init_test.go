package cmd_test

import (
	"os"
	"testing"

	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/deptfile"
)

func changeDeptfileName(name string) func() {
	old := deptfile.DeptfileName
	deptfile.DeptfileName = name
	return func() {
		deptfile.DeptfileName = old
	}
}

func TestInitRun(t *testing.T) {
	const deptfileName = "test.toml"
	cleanup := changeDeptfileName(deptfileName)
	defer func() {
		os.Remove(deptfileName)
		cleanup()
	}()

	mockUI := newMockUI()
	cmd := cmd.NewInit(mockUI)

	t.Run("Run returns 0 normally", func(t *testing.T) {
		code := cmd.Run(nil)
		if code != 0 {
			t.Errorf("Run must finish normally, but got exit code %d", code)
		}

		if _, err := os.Stat(deptfileName); os.IsNotExist(err) {
			t.Error("deptfile must be created, but missing")
		}
	})

	t.Run("Run returns 1 with ErrAlreadyExist", func(t *testing.T) {
		code := cmd.Run(nil)
		if code != 1 {
			t.Error("Run must be failed, but got exit normally")
		}

		if len(mockUI.ErrorWriter().String()) == 0 {
			t.Error("Run must write happened errors to ErrorWriter, but missing")
		}

		if _, err := os.Stat(deptfileName); os.IsExist(err) {
			t.Error("deptfile must not be created, but found")
		}
	})
}
