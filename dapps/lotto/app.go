package lotto

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func CommandViewResult() *cli.Command {
	return &cli.Command{
		Name:    "result",
		Aliases: []string{"r"},
		Usage:   "View lottery draws / result",
		Action: func(c *cli.Context) error {
			fmt.Println("draw result")
			return nil
		},
	}
}

func CommandBuyTicket() *cli.Command {
	return &cli.Command{
		Name:    "buy",
		Aliases: []string{"b"},
		Usage:   "Buy a ticket",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}

func CommandCreateLotto() *cli.Command {
	return &cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
		Usage:   "Create custom lottery",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}

func CommandCancelLotto() *cli.Command {
	return &cli.Command{
		Name:    "cancel",
		Aliases: []string{"ca"},
		Usage:   "Cancel your lottery",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}

func CommandDrawLotto() *cli.Command {
	return &cli.Command{
		Name:    "draw",
		Aliases: []string{"d"},
		Usage:   "Draw the lottery",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}

func CommandClaimReward() *cli.Command {
	return &cli.Command{
		Name:    "claim",
		Aliases: []string{"cl"},
		Usage:   "Claim lotto reward",
		Action: func(c *cli.Context) error {
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
			CommandBuyTicket(),
			CommandCreateLotto(),
			CommandCancelLotto(),
			CommandDrawLotto(),
			CommandClaimReward(),
			CommandViewResult(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
