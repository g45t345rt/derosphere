package lotto

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func CommandDraws() *cli.Command {
	return &cli.Command{
		Name:    "draws",
		Aliases: []string{"d"},
		Usage:   "View lottery draws / result",
		Action: func(c *cli.Context) error {
			fmt.Println("draw result")
			return nil
		},
	}
}

func App() *cli.App {
	return &cli.App{
		Name:        "lotto",
		Description: "Official custom lottery pool. Create your own type of lottery.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandDraws(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
