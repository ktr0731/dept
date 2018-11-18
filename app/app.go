package app

import (
	"os"

	"github.com/ktr0731/dept/cmd"
	"github.com/ktr0731/dept/gocmd"
	"github.com/mitchellh/cli"
)

const (
	name    = "dept"
	version = "0.1.0"
)

// Run receives args, then run dept.
func Run(args []string) (int, error) {
	app := cli.NewCLI(name, version)
	app.Args = args

	app.Commands = map[string]cli.CommandFactory{
		"init": func() (cli.Command, error) {
			return cmd.NewInit(newUI()), nil
		},
		"get": func() (cli.Command, error) {
			return cmd.NewGet(
				newUI(),
				gocmd.New(),
			), nil
		},
	}

	return app.Run()
}

func newUI() cli.Ui {
	return &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}
}
