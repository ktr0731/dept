package filegen_test

import (
	"os"
	"testing"

	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/filegen"
)

func TestGenerator(t *testing.T) {
	df := &deptfile.File{
		Requirements: []*deptfile.Requirement{
			{Name: "github.com/foo/bar"},
			{Name: "github.com/ktr0731/evans"},
		},
	}
	filegen.Generate(os.Stdout, df)
}
