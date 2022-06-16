package dnd

import (
	"github.com/urfave/cli/v2"
)

func App() *cli.App {
	return &cli.App{
		Name:        "dnd",
		Description: "Trading card fantasy world. NFTs by JoyRaptor!",
		Version:     "0.0.1",
		Commands:    []*cli.Command{},
		Authors: []*cli.Author{
			{Name: "JoyRaptor"},
			{Name: "g45t345rt"},
		},
	}
}
