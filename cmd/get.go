package cmd

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/ktr0731/dept/builder"
	"github.com/ktr0731/dept/deptfile"
	"github.com/ktr0731/dept/fetcher"
	"github.com/ktr0731/dept/filegen"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

// getCommand gets a passed Go tool from the remote repository.
// get generate the artifact as follows.
//
//   1. load deptfile.
//   2. generate Go code which imports all required tools.
//   3. run 'go get' with Go modules aware mode to collect dependencies of 2.
//   4. build binaries.
//   5. TODO: Gopkg.toml
//
type getCommand struct {
	ui      cli.Ui
	fetcher fetcher.Fetcher
	builder builder.Builder
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

// Used only mocking
var deptfileLoad func(context.Context) (*deptfile.GoMod, error) = deptfile.Load

func (c *getCommand) Run(args []string) int {
	return run(c, func() error {
		if len(args) != 1 {
			return errShowHelp
		}

		ctx := context.Background()

		df, err := deptfileLoad(ctx)
		if err != nil {
			return err
		}

		repo := args[0]

		requires := make([]string, 0, len(df.Require))
		for _, r := range df.Require {
			requires = append(requires, r.Path)
		}

		var out bytes.Buffer
		filegen.Generate(&out, requires)

		io.Copy(os.Stdout, &out)

		err = c.fetcher.Fetch(ctx, repo)
		if err != nil {
			return errors.Wrap(err, "failed to fetch passed repository")
		}

		err = c.builder.Build(ctx, fetcher.VendorDir(repo))
		if err != nil {
			return errors.Wrap(err, "failed to build fetched repository")
		}

		return nil
	})
}

// NewGet returns an initialized get command instance.
func NewGet(
	ui cli.Ui,
	fetcher fetcher.Fetcher,
	builder builder.Builder,
) cli.Command {
	return &getCommand{
		ui:      ui,
		fetcher: fetcher,
		builder: builder,
	}
}
