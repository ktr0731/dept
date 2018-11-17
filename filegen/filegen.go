// Package filegen provides a Go file generator which
// contains required Go tools in import statement.
// Generated source is used for collecting dependencies of Go tools by 'go get'.
package filegen

import (
	"fmt"
	"go/format"
	"io"
)

var tmpl = `package tools
import (
	%s
)
`

func Generate(w io.Writer, paths []string) {
	var s string
	for _, p := range paths {
		s += fmt.Sprintf("_ \"%s\"\n", p)
	}
	b, err := format.Source([]byte(fmt.Sprintf(tmpl, s)))
	if err != nil {
		panic("invalid source passed")
	}
	w.Write(b)
}
