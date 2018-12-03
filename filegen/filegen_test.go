package filegen_test

import (
	"bytes"
	"testing"

	"github.com/ktr0731/dept/filegen"
)

func TestGenerator(t *testing.T) {
	expected := `package tools

import (
	_ "github.com/foo/bar"
	_ "github.com/ktr0731/evans"
)
`
	df := []string{
		"github.com/foo/bar",
		"github.com/ktr0731/evans",
	}
	var buf bytes.Buffer
	filegen.Generate(&buf, df)
	if actual := buf.String(); expected != actual {
		t.Errorf("expected:\n%s\n\nactual:\n%s", expected, actual)
	}
}

func TestGeneratorInvalidSource(t *testing.T) {
	df := []string{
		`"github.com/foo/bar`,
	}
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("Generate must panic")
		}
	}()
	var buf bytes.Buffer
	filegen.Generate(&buf, df)
}
