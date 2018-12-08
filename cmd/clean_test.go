package cmd_test

import (
	"context"
	"testing"

	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/toolcacher"
)

func TestCleanRun(t *testing.T) {
	mockUI := newMockUI()
	mockToolcacher := &toolcacher.CacherMock{
		ClearFunc: func(ctx context.Context) error {
			return nil
		},
	}
	cmd := cmd.NewClean(mockUI, mockToolcacher)
	code := cmd.Run(nil)
	if code != 0 {
		t.Errorf("code must be 0, but got %d", code)
	}

	if n := len(mockToolcacher.ClearCalls()); n != 1 {
		t.Errorf("Clear must be called once, but called %d times", n)
	}
}
