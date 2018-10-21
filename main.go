package main

import (
	"log"
	"os"

	"github.com/ktr0731/dept/cmd"
	"github.com/mitchellh/cli"
)

func main() {
	app := cli.NewCLI("dept", "0.1.0")
	app.Commands = map[string]cli.CommandFactory{
		"add": cmd.Add,
	}
	app.Args = os.Args[1:]
	code, err := app.Run()
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(code)
}
