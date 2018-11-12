package deptfile_test

import (
	"bytes"
	"testing"

	"github.com/ktr0731/dept/deptfile"
)

func TestFile(t *testing.T) {
	expected := `[[requirements]]
  name = "github.com/ktr0731/evans"
`
	f := &deptfile.File{}
	f.Requirements = append(f.Requirements, &deptfile.Requirement{Name: "github.com/ktr0731/evans"})

	var buf bytes.Buffer
	err := f.Encode(&buf)
	if err != nil {
		t.Fatalf("expected Encode has no errors, but got an error: %s", err)
	}
	if actual := buf.String(); expected != actual {
		t.Errorf("expected %s, but got %s", expected, actual)
	}
}
