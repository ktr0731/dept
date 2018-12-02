package main

import (
	"fmt"
	"os"

	"github.com/ktr0731/dept/app"
	"github.com/pkg/profile"
)

func main() {
	defer profile.Start().Stop()

	code, err := app.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run dept: %s\n", err)
	}

	os.Exit(code)
}
