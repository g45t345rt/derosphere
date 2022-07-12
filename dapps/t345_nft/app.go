package t345_nft

import (
	"github.com/urfave/cli/v2"
)

func App() *cli.App {
	return &cli.App{
		Name:        "t345-nft",
		Description: "Deploy & manage T345-NFT.",
		Version:     "0.0.1",
		Commands:    []*cli.Command{},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
