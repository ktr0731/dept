package deptfile

import (
	"os"
	"testing"
)

func TestRunInWorkspace(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working dir: %s", err)
	}
	err = RunInWorkspace(func() {
		newPWD, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current working dir: %s", err)
		}
		if pwd == newPWD {
			t.Errorf("current dir in RunInWorkspace must not be equal to dir outside of RunInWorkspace")
		}
	})
	if err != nil {
		t.Errorf("RunInWorkspace must not return errors, but got an error: %s", err)
	}
	pwd2, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current working dir: %s", err)
	}
	if pwd != pwd2 {
		t.Errorf("current working dir which called before RunInWorkspace and after one must be equal")
	}
}
