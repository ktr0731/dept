// Package filegen provides a Go file generator which
// contains required Go tools in import statement.
// Generated source is used for collecting dependencies of Go tools by 'go get'.
package filegen

import (
	"fmt"
	"go/format"
	"io"

	"github.com/ktr0731/dept/deptfile"
)

var tmpl = `package tools
import (
	%s
)
`

func Generate(w io.Writer, df *deptfile.File) {
	var s string
	for _, r := range df.Requirements {
		s += fmt.Sprintf("_ \"%s\"\n", r.Name)
	}
	b, err := format.Source([]byte(fmt.Sprintf(tmpl, s)))
	if err != nil {
		panic("invalid source passed")
	}
	w.Write(b)
}
