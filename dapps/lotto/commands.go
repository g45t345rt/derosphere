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

func Commands() []*cli.Command {
	return []*cli.Command{CommandDraws()}
}
