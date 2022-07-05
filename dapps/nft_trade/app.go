package nft_trade

import (
	"github.com/urfave/cli/v2"
)

func CommandAuction() *cli.Command {
	return &cli.Command{
		Name:    "auction",
		Aliases: []string{"au"},
		Usage:   "NFT in auction",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}

func CommandExchange() *cli.Command {
	return &cli.Command{
		Name:    "exchange",
		Aliases: []string{"ex"},
		Usage:   "List NFTs you can buy",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}

func CommandListCollection() *cli.Command {
	return &cli.Command{
		Name:    "list-collection",
		Aliases: []string{"lc"},
		Usage:   "List available NFT collection",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}

func App() *cli.App {
	return &cli.App{
		Name:        "nft-trade",
		Description: "Browse, buy, sell and auction NFTs.",
		Version:     "0.0.1",
		Commands:    []*cli.Command{},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
