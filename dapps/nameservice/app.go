package nameservice

import (
	"github.com/urfave/cli/v2"
)

func CommandRegister() *cli.Command {
	return &cli.Command{
		Name:    "register",
		Aliases: []string{"r"},
		Usage:   "Register a name to your address",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}

func App() *cli.App {
	return &cli.App{
		Name:        "nameservice",
		Description: "Register multiple names to receive DERO from others.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandRegister(),
		},
		Authors: []*cli.Author{
			{Name: "Captain"},
			{Name: "g45t345rt"},
		},
	}
}
