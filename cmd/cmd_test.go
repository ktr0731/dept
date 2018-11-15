package cmd_test

import (
	"bytes"
	"os"

	"github.com/mitchellh/cli"
)

type testUI struct {
	*cli.BasicUi
}

func (u *testUI) Writer() *bytes.Buffer {
	return u.BasicUi.Writer.(*bytes.Buffer)
}

func (u *testUI) ErrorWriter() *bytes.Buffer {
	return u.BasicUi.ErrorWriter.(*bytes.Buffer)
}

func newTestUI() *testUI {
	return &testUI{
		&cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      new(bytes.Buffer),
			ErrorWriter: new(bytes.Buffer),
		},
	}
}
