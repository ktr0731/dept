package cmd

import (
	"errors"

	"github.com/mitchellh/cli"
)

var (
	errShowHelp = errors.New("show help")
)

type command interface {
	cli.Command
	UI() cli.Ui
}

func run(c command, f func() error) int {
	err := f()
	if err == nil {
		return 0
	}

	switch err {
	case errShowHelp:
		c.UI().Output(c.Help())
	default:
		c.UI().Error(err.Error())
	}
	return 1
}
