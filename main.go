package main

import (
	"fmt"
	"os"

	"github.com/ktr0731/dept/app"
)

func main() {
	code, err := app.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run dept: %s\n", err)
	}

	os.Exit(code)
}
