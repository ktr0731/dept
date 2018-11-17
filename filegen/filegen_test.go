package filegen_test

import (
	"os"
	"testing"

	"github.com/ktr0731/dept/filegen"
)

func TestGenerator(t *testing.T) {
	df := []string{
		"github.com/foo/bar",
		"github.com/ktr0731/evans",
	}
	filegen.Generate(os.Stdout, df)
}
