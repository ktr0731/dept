package cmd_test

import (
	"bytes"
	"os"

	"github.com/mitchellh/cli"
)

type mockUI struct {
	*cli.BasicUi
}

func (u *mockUI) Writer() *bytes.Buffer {
	return u.BasicUi.Writer.(*bytes.Buffer)
}

func (u *mockUI) ErrorWriter() *bytes.Buffer {
	return u.BasicUi.ErrorWriter.(*bytes.Buffer)
}

func newMockUI() *mockUI {
	return &mockUI{
		&cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      new(bytes.Buffer),
			ErrorWriter: new(bytes.Buffer),
		},
	}
}
