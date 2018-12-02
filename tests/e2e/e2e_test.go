package e2e

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/ktr0731/dept/app"
)

func setupOutput() (*bytes.Buffer, *bytes.Buffer, func()) {
	var out, eout bytes.Buffer
	app.Stdout = &out
	app.Stderr = &eout
	return &out, &eout, func() {
		app.Stdout = os.Stdout
		app.Stderr = os.Stderr
	}
}

func TestGet(t *testing.T) {
	cases := map[string]struct {
		args []string
	}{
		"get a new tool": {args: []string{"-v", "get", "github.com/ktr0731/salias"}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			// out, eout, cleanup := setupOutput()
			// defer cleanup()

			code, err := app.Run(c.args)
			if err != nil {
				// pp.Println(out.String())
				t.Fatal(err)
			}
			pp.Println(code)
			// pp.Println(out.String())
			// pp.Println(eout.String())
		})
	}
}

func BenchmarkGet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code, err := app.Run([]string{"-v", "get", "github.com/ktr0731/salias"})
		fmt.Println(code)
		if err != nil {
			b.Fatal(err)
		}
	}
}
