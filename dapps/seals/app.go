package seals

import (
	"github.com/urfave/cli/v2"
)

func App() *cli.App {
	return &cli.App{
		Name:        "seals",
		Description: "Dero Seals NFT project.",
		Version:     "0.0.1",
		Commands:    []*cli.Command{},
		Authors: []*cli.Author{
			{Name: "billoetree"},
			{Name: "M-M"},
			{Name: "g45t345rt"},
		},
	}
}
