package t345_nft

import (
	"github.com/urfave/cli/v2"
)

func CommandSetRoyaltyFees() *cli.Command {
	return &cli.Command{
		Name:    "set-royaltyfees",
		Aliases: []string{"sr"},
		Usage:   "Set royalty fees when transfering NFTs",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandSetBurn() *cli.Command {
	return &cli.Command{
		Name:    "set-burn",
		Aliases: []string{"sb"},
		Usage:   "Set if you can burn the NFTs",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandBurn() *cli.Command {
	return &cli.Command{
		Name:    "burn",
		Aliases: []string{"b"},
		Usage:   "Burn an NFT",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandDeploy() *cli.Command {
	return &cli.Command{
		Name:    "deploy",
		Aliases: []string{"m"},
		Usage:   "Deploy NFT collection",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandMint() *cli.Command {
	return &cli.Command{
		Name:    "mint",
		Aliases: []string{"m"},
		Usage:   "Mint an NFT",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandSet() *cli.Command {
	return &cli.Command{
		Name:    "set-nft",
		Aliases: []string{"sn"},
		Usage:   "Edit/update an NFT",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandInitStore() *cli.Command {
	return &cli.Command{
		Name:    "init-store",
		Aliases: []string{"is"},
		Usage:   "Initialize NFT collection",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandTransfer() *cli.Command {
	return &cli.Command{
		Name:    "transfer",
		Aliases: []string{"t"},
		Usage:   "Transfer an NFT",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandClaimTransfer() *cli.Command {
	return &cli.Command{
		Name:    "claim-transfer",
		Aliases: []string{"ct"},
		Usage:   "Claim the NFT",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandCancelTransfer() *cli.Command {
	return &cli.Command{
		Name:    "cancel-transfer",
		Aliases: []string{"cc"},
		Usage:   "Cancel live transfer",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func App() *cli.App {
	return &cli.App{
		Name:        "t345-nft",
		Description: "Deploy & manage T345-NFT Smart Contract.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandDeploy(),
			CommandInitStore(),
			CommandMint(),
			CommandSet(),
			CommandBurn(),
			CommandSetBurn(),
			CommandSetRoyaltyFees(),
			CommandTransfer(),
			CommandClaimTransfer(),
			CommandCancelTransfer(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
