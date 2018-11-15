package cmd

import (
	"os"

	"github.com/ktr0731/dept/deptfile"
	"github.com/mitchellh/cli"
)

// getCommand gets a passed Go tool from the remote repository.
// get generate the artifact as follows.
//
//   1. clone the remote repository.
//   2. find go.mod file from the cloned repository.
//     a. if go.mod missing, generate go.mod.
//   3. run 'go build' with Go modules aware mode.
//   4. move the artifact to user's repository.
//   5. update own config file.
//
type getCommand struct {
	ui cli.Ui
	df *deptfile.File
}

func (c *getCommand) UI() cli.Ui {
	return c.ui
}

func (c *getCommand) Help() string {
	return "Usage: dept get <url>"
}

func (c *getCommand) Synopsis() string {
	return "Get new CLI tool as a dependency"
}

func (c *getCommand) Run(args []string) int {
	return run(c, func() error {
		if len(args) != 1 {
			return errShowHelp
		}

		// ctx := context.Background()

		path := args[0]

		c.df.Requirements = append(c.df.Requirements, &deptfile.Requirement{path})
		c.df.Encode(os.Stdout)

		// log.Println("create temp dir")
		// name, err := ioutil.TempDir("", "dept")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// // defer os.RemoveAll(name)
		//
		// log.Println("created: ", name)
		//
		// err = exec.CommandContext(ctx, "git", "clone", fmt.Sprintf("https://%s", path), name).Run()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		//
		// log.Println("cloned: ", path)
		//
		// modPath := filepath.Join(name, path, "go.mod")
		// if _, err := os.Stat(modPath); os.IsNotExist(err) {
		// 	log.Println("go.mod not found")
		// }

		return nil
	})
}

// Get returns an initialized get command instance.
func Get(ui cli.Ui, df *deptfile.File) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &getCommand{
			ui: ui,
			df: df,
		}, nil
	}
}

type dependency struct {
	Revision string `json:"revision"`
	Digest   string `json:"digest"`
}
