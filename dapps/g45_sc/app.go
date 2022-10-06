package g45_sc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

/** G45-AT **/

func Command_G45_AT_View() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-view",
		Aliases: []string{"g45-at-v"},
		Usage:   "Display G45-AT metadata and more",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance

			asset := utils.G45_AT{}
			result, err := walletInstance.Daemon.GetSC(&rpc.GetSC_Params{
				SCID:      scid,
				Code:      true,
				Variables: true,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			err = asset.Parse(scid, result)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			asset.Print()
			return nil
		},
	}
}

func Command_G45_AT_Deploy() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-deploy",
		Aliases: []string{"g45-at-d"},
		Usage:   "Deploy G45-AT Smart Contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			assetType, err := app.PromptChoose("Asset token type", []string{"public", "private"}, "private")
			if app.HandlePromptErr(err) {
				return nil
			}

			code := utils.G45_AT_PRIVATE_CODE
			if assetType == "public" {
				code = utils.G45_AT_PUBLIC_CODE
			}

			collectionSCID, err := app.Prompt("Enter collection scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			decimals, err := app.PromptUInt("Enter token decimals", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			startSupply, err := app.PromptUInt("Enter amount to mint", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			metadataFormat, err := app.Prompt("Enter metadata format", "json")
			if app.HandlePromptErr(err) {
				return nil
			}

			metadata, err := app.Prompt("Enter metadata", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			freezeMetadata, err := app.PromptYesNo("Freeze metadata?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeMetadata := 0
			if freezeMetadata {
				uFreezeMetadata = 1
			}

			freezeMint, err := app.PromptYesNo("Freeze minting?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeMint := 0
			if freezeMint {
				uFreezeMint = 1
			}

			freezeCollection, err := app.PromptYesNo("Freeze collection?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeCollection := 0
			if freezeCollection {
				uFreezeCollection = 1
			}

			txId, err := walletInstance.InstallSmartContract([]byte(code), 2, []rpc.Argument{
				{Name: "startSupply", DataType: rpc.DataUint64, Value: startSupply},
				{Name: "decimals", DataType: rpc.DataUint64, Value: decimals},
				{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
				{Name: "metadataFormat", DataType: rpc.DataString, Value: metadataFormat},
				{Name: "metadata", DataType: rpc.DataString, Value: metadata},
				{Name: "freezeCollection", DataType: rpc.DataUint64, Value: uFreezeCollection},
				{Name: "freezeMint", DataType: rpc.DataUint64, Value: uFreezeMint},
				{Name: "freezeMetadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
			}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_AT_Mint() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-mint",
		Aliases: []string{"g45-at-m"},
		Usage:   "Increase supply (mint tokens)",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			supply, err := app.PromptUInt("Enter amount", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "Mint", []rpc.Argument{
				{Name: "amount", DataType: rpc.DataUint64, Value: supply},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_AT_SetCollection() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-setcollection",
		Aliases: []string{"g45-at-sc"},
		Usage:   "Set collection SCID to G45-AT",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			collectionSCID, err := app.Prompt("Enter collection SCID", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "SetCollection", []rpc.Argument{
				{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_AT_Freeze() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-freeze",
		Aliases: []string{"g45-at-f"},
		Usage:   "Freeze G45-AT (mint, metadata or collection)",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeMint := 0
			freezeMint, err := app.PromptYesNo("Freeze minting?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeMint {
				uFreezeMint = 1
			}

			uFreezeMetadata := 0
			freezeMetadata, err := app.PromptYesNo("Freeze metadata?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeMetadata {
				uFreezeMetadata = 1
			}

			uFreezeCollection := 0
			freezeCollection, err := app.PromptYesNo("Freeze collection?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeCollection {
				uFreezeCollection = 1
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "Freeze", []rpc.Argument{
				{Name: "mint", DataType: rpc.DataUint64, Value: uFreezeMint},
				{Name: "metadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
				{Name: "collection", DataType: rpc.DataUint64, Value: uFreezeCollection},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_AT_SetMetadata() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-setmetadata",
		Aliases: []string{"g45-at-sm"},
		Usage:   "Set/edit metadata of the asset",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			format, err := app.Prompt("Metadata format?", "json")
			if app.HandlePromptErr(err) {
				return nil
			}

			metadata, err := app.Prompt("Set new metadata", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "SetMetadata", []rpc.Argument{
				{Name: "format", DataType: rpc.DataString, Value: format},
				{Name: "metadata", DataType: rpc.DataString, Value: metadata},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_AT_Burn() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-burn",
		Aliases: []string{"g45-at-b"},
		Usage:   "Publicly burn tokens and edit supply",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			burnAmount, err := app.PromptUInt("Burn", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			transfer := rpc.Transfer{
				SCID:        crypto.HashHexToHash(scid),
				Destination: randomAddresses.Address[0],
				Burn:        burnAmount,
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "Burn", []rpc.Argument{}, []rpc.Transfer{
				transfer,
			}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_AT_DisplayToken() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-displaytoken",
		Aliases: []string{"g45-at-dt"},
		Usage:   "Display token in SC",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			amount, err := app.PromptUInt("Amount", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			transfer := rpc.Transfer{
				SCID:        crypto.HashHexToHash(scid),
				Destination: randomAddresses.Address[0],
				Burn:        amount,
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "DisplayToken", []rpc.Argument{}, []rpc.Transfer{
				transfer,
			}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_AT_RetrieveToken() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-retrievetoken",
		Aliases: []string{"g45-at-rt"},
		Usage:   "Retrieve token from SC",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			amount, err := app.PromptUInt("Amount", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "RetrieveToken", []rpc.Argument{
				{Name: "amount", DataType: rpc.DataUint64, Value: amount},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_AT_TransferMinter() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-transferminter",
		Aliases: []string{"g45-at-tm"},
		Usage:   "Initiate transfer minter",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			addr, err := app.Prompt("New owner address", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "TransferMinter", []rpc.Argument{
				{Name: "newMinter", DataType: rpc.DataString, Value: addr},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_AT_CancelTransferMinter() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-canceltransferminter",
		Aliases: []string{"g45-at-ctm"},
		Usage:   "Cancel ongoing transfer minter",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "CancelTransferMinter", []rpc.Argument{}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_AT_ClaimMinter() *cli.Command {
	return &cli.Command{
		Name:    "g45-at-claimminter",
		Aliases: []string{"g45-at-cm"},
		Usage:   "Claim minter pending transfer",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "ClaimMinter", []rpc.Argument{}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

/** G45-FAT **/

func Command_G45_FAT_View() *cli.Command {
	return &cli.Command{
		Name:    "g45-fat-view",
		Aliases: []string{"g45-fat-v"},
		Usage:   "Display G45-FAT metadata and more",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance

			asset := utils.G45_FAT{}
			result, err := walletInstance.Daemon.GetSC(&rpc.GetSC_Params{
				SCID:      scid,
				Code:      true,
				Variables: true,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			err = asset.Parse(scid, result)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			asset.Print()
			return nil
		},
	}
}

func Command_G45_FAT_Deploy() *cli.Command {
	return &cli.Command{
		Name:    "g45-fat-deploy",
		Aliases: []string{"g45-fat-d"},
		Usage:   "Deploy G45-FAT Smart Contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			assetType, err := app.PromptChoose("Asset token type", []string{"public", "private"}, "private")
			if app.HandlePromptErr(err) {
				return nil
			}

			code := utils.G45_FAT_PRIVATE_CODE
			if assetType == "public" {
				code = utils.G45_FAT_PUBLIC_CODE
			}

			collectionSCID, err := app.Prompt("Enter collection scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			decimals, err := app.PromptUInt("Enter token decimals", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			startSupply, err := app.PromptUInt("Enter max supply", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			metadataFormat, err := app.Prompt("Enter metadata format", "json")
			if app.HandlePromptErr(err) {
				return nil
			}

			metadata, err := app.Prompt("Enter metadata", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			freezeMetadata, err := app.PromptYesNo("Freeze metadata?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeMetadata := 0
			if freezeMetadata {
				uFreezeMetadata = 1
			}

			freezeCollection, err := app.PromptYesNo("Freeze collection?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeCollection := 0
			if freezeCollection {
				uFreezeCollection = 1
			}

			txId, err := walletInstance.InstallSmartContract([]byte(code), 2, []rpc.Argument{
				{Name: "maxSupply", DataType: rpc.DataUint64, Value: startSupply},
				{Name: "decimals", DataType: rpc.DataUint64, Value: decimals},
				{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
				{Name: "metadataFormat", DataType: rpc.DataString, Value: metadataFormat},
				{Name: "metadata", DataType: rpc.DataString, Value: metadata},
				{Name: "freezeCollection", DataType: rpc.DataUint64, Value: uFreezeCollection},
				{Name: "freezeMetadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
			}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_FAT_SetMetadata() *cli.Command {
	return &cli.Command{
		Name:    "g45-fat-setmetadata",
		Aliases: []string{"g45-fat-sm"},
		Usage:   "Set/edit metadata of the asset",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			format, err := app.Prompt("Metadata format?", "json")
			if app.HandlePromptErr(err) {
				return nil
			}

			metadata, err := app.Prompt("Set new metadata", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "SetMetadata", []rpc.Argument{
				{Name: "format", DataType: rpc.DataString, Value: format},
				{Name: "metadata", DataType: rpc.DataString, Value: metadata},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_FAT_SetCollection() *cli.Command {
	return &cli.Command{
		Name:    "g45-fat-setcollection",
		Aliases: []string{"g45-fat-sc"},
		Usage:   "Set collection SCID to G45-FAT",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			collectionSCID, err := app.Prompt("Enter collection SCID", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "SetCollection", []rpc.Argument{
				{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_FAT_Burn() *cli.Command {
	return &cli.Command{
		Name:    "g45-fat-burn",
		Aliases: []string{"g45-fat-b"},
		Usage:   "Publicly burn tokens and edit supply",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			burnAmount, err := app.PromptUInt("Burn", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			transfer := rpc.Transfer{
				SCID:        crypto.HashHexToHash(scid),
				Destination: randomAddresses.Address[0],
				Burn:        burnAmount,
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "Burn", []rpc.Argument{}, []rpc.Transfer{
				transfer,
			}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_FAT_Freeze() *cli.Command {
	return &cli.Command{
		Name:    "g45-fat-freeze",
		Aliases: []string{"g45-fat-f"},
		Usage:   "Freeze G45-FAT (metadata or collection)",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeMetadata := 0
			freezeMetadata, err := app.PromptYesNo("Freeze metadata?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeMetadata {
				uFreezeMetadata = 1
			}

			uFreezeCollection := 0
			freezeCollection, err := app.PromptYesNo("Freeze collection?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeCollection {
				uFreezeCollection = 1
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "Freeze", []rpc.Argument{
				{Name: "metadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
				{Name: "collection", DataType: rpc.DataUint64, Value: uFreezeCollection},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_FAT_DisplayToken() *cli.Command {
	return &cli.Command{
		Name:    "g45-fat-displaytoken",
		Aliases: []string{"g45-fat-dt"},
		Usage:   "Display token in SC",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			amount, err := app.PromptUInt("Amount", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			transfer := rpc.Transfer{
				SCID:        crypto.HashHexToHash(scid),
				Destination: randomAddresses.Address[0],
				Burn:        amount,
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "DisplayToken", []rpc.Argument{}, []rpc.Transfer{
				transfer,
			}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_FAT_RetrieveToken() *cli.Command {
	return &cli.Command{
		Name:    "g45-fat-retrievetoken",
		Aliases: []string{"g45-fat-rt"},
		Usage:   "Retrieve token from SC",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			amount, err := app.PromptUInt("Amount", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "RetrieveToken", []rpc.Argument{
				{Name: "amount", DataType: rpc.DataUint64, Value: amount},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

/** G45-NFT **/

func Command_G45_NFT_View() *cli.Command {
	return &cli.Command{
		Name:    "g45-nft-view",
		Aliases: []string{"g45-nft-v"},
		Usage:   "Display G45-NFT metadata and more",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance

			asset := utils.G45_NFT{}
			result, err := walletInstance.Daemon.GetSC(&rpc.GetSC_Params{
				SCID:      scid,
				Code:      true,
				Variables: true,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			err = asset.Parse(scid, result)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			asset.Print()
			return nil
		},
	}
}

func Command_G45_NFT_Deploy() *cli.Command {
	return &cli.Command{
		Name:    "g45-nft-deploy",
		Aliases: []string{"g45-nft-d"},
		Usage:   "Deploy G45-NFT Smart Contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			assetType, err := app.PromptChoose("Privacy type", []string{"public", "private"}, "private")
			if app.HandlePromptErr(err) {
				return nil
			}

			code := utils.G45_NFT_PRIVATE_CODE
			if assetType == "public" {
				code = utils.G45_NFT_PUBLIC_CODE
			}

			collectionSCID, err := app.Prompt("Enter collection scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			metadataFormat, err := app.Prompt("Enter metadata format", "json")
			if app.HandlePromptErr(err) {
				return nil
			}

			metadata, err := app.Prompt("Enter metadata", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			txId, err := walletInstance.InstallSmartContract([]byte(code), 2, []rpc.Argument{
				{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
				{Name: "metadataFormat", DataType: rpc.DataString, Value: metadataFormat},
				{Name: "metadata", DataType: rpc.DataString, Value: metadata},
			}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_NFT_DisplayNFT() *cli.Command {
	return &cli.Command{
		Name:    "g45-nft-displaynft",
		Aliases: []string{"g45-nft-dn"},
		Usage:   "Display token in SC",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			transfer := rpc.Transfer{
				SCID:        crypto.HashHexToHash(scid),
				Destination: randomAddresses.Address[0],
				Burn:        1,
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "DisplayNFT", []rpc.Argument{}, []rpc.Transfer{
				transfer,
			}, true)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_NFT_RetrieveNFT() *cli.Command {
	return &cli.Command{
		Name:    "g45-nft-retrievenft",
		Aliases: []string{"g45-nft-rn"},
		Usage:   "Retrieve token from SC",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "RetrieveNFT", []rpc.Argument{}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

/** G45-C **/

func Command_G45_C_View() *cli.Command {
	return &cli.Command{
		Name:    "g45-c-view",
		Aliases: []string{"g45-c-v"},
		Usage:   "Display G45-C metadata and more",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			collection := utils.G45_C{}
			result, err := walletInstance.Daemon.GetSC(&rpc.GetSC_Params{
				SCID:      scid,
				Code:      true,
				Variables: true,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			err = collection.Parse(scid, result)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			collection.Print()
			return nil
		},
	}
}

func G45_C_Deploy() (string, error) {
	walletInstance := app.Context.WalletInstance

	metadataFormat, err := app.Prompt("Metadata format?", "json")
	if err != nil {
		return "", err
	}

	metadata, err := app.Prompt("Set metadata", "")
	if err != nil {
		return "", err
	}

	uFreezeMetadata := 0
	freezeMetadata, err := app.PromptYesNo("Freeze metadata?", false)
	if err != nil {
		return "", err
	}

	if freezeMetadata {
		uFreezeMetadata = 1
	}

	txId, err := walletInstance.InstallSmartContract([]byte(utils.G45_C_CODE), 2, []rpc.Argument{
		{Name: "metadataFormat", DataType: rpc.DataString, Value: metadataFormat},
		{Name: "metadata", DataType: rpc.DataString, Value: metadata},
		{Name: "freezeMetadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
	}, true)

	return txId, err
}

func Command_G45_C_Deploy() *cli.Command {
	return &cli.Command{
		Name:    "g45-c-deploy",
		Aliases: []string{"g45-c-d"},
		Usage:   "Deploy G45-C Smart Contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			txId, err := G45_C_Deploy()
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_C_Freeze() *cli.Command {
	return &cli.Command{
		Name:    "g45-c-freeze",
		Aliases: []string{"g45-c-freeze"},
		Usage:   "Freeze G45-C (assets/metadata)",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeAssets := 0
			freezeAssets, err := app.PromptYesNo("Freeze assets/nfts?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeAssets {
				uFreezeAssets = 1
			}

			uFreezeMetadata := 0
			freezeMetadata, err := app.PromptYesNo("Freeze metadata?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeMetadata {
				uFreezeMetadata = 1
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "Freeze", []rpc.Argument{
				{Name: "assets", DataType: rpc.DataUint64, Value: uFreezeAssets},
				{Name: "metadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func G45_C_SetAssets(scId string, assets map[string]uint64, promptFees bool) error {
	startAt, err := app.PromptUInt("Start at index", 0)
	if app.HandlePromptErr(err) {
		return nil
	}

	maxAssetsPerEntry, err := app.PromptUInt("Assets per TX", 100)
	if app.HandlePromptErr(err) {
		return nil
	}

	var entries []map[string]uint64
	index := 0
	assetsEntry := make(map[string]uint64)
	for key, v := range assets {
		assetsEntry[key] = v
		index++

		if index >= int(maxAssetsPerEntry) {
			entries = append(entries, assetsEntry)
			assetsEntry = make(map[string]uint64)
			index = 0
		}
	}

	if len(assetsEntry) > 0 {
		entries = append(entries, assetsEntry)
	}

	fmt.Printf("%d assets with %d entries\n", len(assets), len(entries))
	walletInstance := app.Context.WalletInstance
	for i, entry := range entries {
		if i >= int(startAt) {
			data, err := json.Marshal(entry)
			if err != nil {
				fmt.Println(err)
				return nil
			}

		set_assets:
			fmt.Printf("Set Assets %d - %d assets\n", i, len(entry))
			txId, err := walletInstance.CallSmartContract(2, scId, "SetAssets", []rpc.Argument{
				{Name: "index", DataType: rpc.DataUint64, Value: uint64(i)},
				{Name: "assets", DataType: rpc.DataString, Value: string(data)},
			}, []rpc.Transfer{}, promptFees)
			if err != nil {
				fmt.Println(err)
				time.Sleep(2 * time.Second)
				goto set_assets
			}

			err = walletInstance.WaitTransaction(txId)
			if err != nil {
				fmt.Println(err)
				time.Sleep(2 * time.Second)
				goto set_assets
			}
		}
	}

	return nil
}

func Command_G45_C_SetAssets() *cli.Command {
	return &cli.Command{
		Name:    "g45-c-setassets",
		Aliases: []string{"g45-c-sa"},
		Usage:   "Set assets to G45-C",
		Action: func(ctx *cli.Context) error {
			collectionSCID, err := app.Prompt("Enter collection scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			metadataPath, err := app.Prompt("Enter assets metadata file path", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			content, err := ioutil.ReadFile(metadataPath)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			var assets map[string]uint64
			err = json.Unmarshal(content, &assets)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			err = G45_C_SetAssets(collectionSCID, assets, false)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			return nil
		},
	}
}

func Command_G45_C_DelAssets() *cli.Command {
	return &cli.Command{
		Name:    "g45-c-delassets",
		Aliases: []string{"g45-c-da"},
		Usage:   "Del assets from G45-C",
		Action: func(ctx *cli.Context) error {
			collectionSCID, err := app.Prompt("Enter collection scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			index, err := app.PromptUInt("Enter index to delete?", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, collectionSCID, "DelAssets", []rpc.Argument{
				{Name: "index", DataType: rpc.DataUint64, Value: index},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_C_SetMetadata() *cli.Command {
	return &cli.Command{
		Name:    "g45-c-setmetadata",
		Aliases: []string{"g45-c-sm"},
		Usage:   "Set/edit metadata of the collection",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			format, err := app.Prompt("Metadata format?", "json")
			if app.HandlePromptErr(err) {
				return nil
			}

			metadata, err := app.Prompt("Set new metadata", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "SetMetadata", []rpc.Argument{
				{Name: "format", DataType: rpc.DataString, Value: format},
				{Name: "metadata", DataType: rpc.DataString, Value: metadata},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_C_TransferOwnership() *cli.Command {
	return &cli.Command{
		Name:    "g45-c-transferownership",
		Aliases: []string{"g45-c-to"},
		Usage:   "Initiate collection transfer ownership",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			addr, err := app.Prompt("New owner address", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "TransferOwnership", []rpc.Argument{
				{Name: "newOwner", DataType: rpc.DataString, Value: addr},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_C_ClaimOwnership() *cli.Command {
	return &cli.Command{
		Name:    "g45-c-claimownership",
		Aliases: []string{"g45-c-co"},
		Usage:   "Claim collection ownership",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "ClaimOwnership", []rpc.Argument{}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_C_CancelTransferOwnership() *cli.Command {
	return &cli.Command{
		Name:    "g45-c-canceltransferownership",
		Aliases: []string{"g45-c-cto"},
		Usage:   "Cancel transfer collection ownership",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "CancelTransferOwnership", []rpc.Argument{}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func Command_G45_C_DeployNFTs() *cli.Command {
	return &cli.Command{
		Name:  "g45-c-deploynfts",
		Usage: "Script to deploy nfts with collection",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			installCollection, err := app.PromptYesNo("Install G45-C?", true)
			if app.HandlePromptErr(err) {
				return nil
			}

			var collectionSCID string
			if installCollection {
				collectionSCID, err = G45_C_Deploy()
				if app.HandlePromptErr(err) {
					return nil
				}

				walletInstance.RunTxChecker(collectionSCID)
			} else {
				collectionSCID, err = app.Prompt("G45-C Smart Contract?", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			metadataPath, err := app.Prompt("Enter nfts metadata file path", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			metadataContent, err := ioutil.ReadFile(metadataPath)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			var metadataList []interface{}
			err = json.Unmarshal(metadataContent, &metadataList)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			scType, err := app.PromptChoose("G45-NFT type", []string{"public", "private"}, "private")
			if app.HandlePromptErr(err) {
				return nil
			}

			scCode := utils.G45_NFT_PUBLIC_CODE
			if scType == "private" {
				scCode = utils.G45_NFT_PRIVATE_CODE
			}

			startIndex, err := app.PromptUInt("NFT start index", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			endIndex, err := app.PromptUInt("NFT end index", 99)
			if app.HandlePromptErr(err) {
				return nil
			}

			nfts := make(map[string]uint64)

			nftsOutputPath, err := app.Prompt("Enter nft list output file path", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			loadNFTs, err := app.PromptYesNo("Load nft list from output file?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if loadNFTs {
				nftsContent, err := ioutil.ReadFile(nftsOutputPath)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				err = json.Unmarshal(nftsContent, &nfts)
				if err != nil {
					fmt.Println(err)
					return nil
				}
			}

			//sleep := 1000 * time.Millisecond
			for index := startIndex; index <= endIndex; index++ {
				assetMetadata := metadataList[index]

				bMetadata, err := json.Marshal(assetMetadata)
				if err != nil {
					fmt.Println(err)
					continue
				}

				sMetadata := string(bMetadata)

			install_asset:
				fmt.Printf("Install NFT - %d\n", index)
				assetSCID, err := walletInstance.InstallSmartContract([]byte(scCode), 2, []rpc.Argument{
					{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
					{Name: "metadataFormat", DataType: rpc.DataString, Value: "json"},
					{Name: "metadata", DataType: rpc.DataString, Value: sMetadata},
				}, false)

				if err != nil {
					fmt.Println(err)
					time.Sleep(2 * time.Second)
					goto install_asset
				}

				err = walletInstance.WaitTransaction(assetSCID)
				if err != nil {
					fmt.Println(err)
					time.Sleep(2 * time.Second)
					goto install_asset
				}

				nfts[assetSCID] = index

				jsonNFTs, err := json.Marshal(nfts)
				if err != nil {
					return err
				}

				err = ioutil.WriteFile(nftsOutputPath, jsonNFTs, os.ModePerm)
				if err != nil {
					return err
				}
			}

			setCollectionAssets, err := app.PromptYesNo("Set all nfts in collection?", true)
			if app.HandlePromptErr(err) {
				return nil
			}

			if setCollectionAssets {
				err = G45_C_SetAssets(collectionSCID, nfts, false)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

/** Others **/

func CommandValidSC() *cli.Command {
	return &cli.Command{
		Name:    "valid-sc",
		Aliases: []string{"v-sc"},
		Usage:   "Check if smart contract is a valid G45-AT/G45-FAT/G45-NFT/G45-C",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			scid := ctx.Args().First()
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			result, err := walletInstance.Daemon.GetSC(&rpc.GetSC_Params{
				Code:      true,
				Variables: false,
				SCID:      scid,
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			//checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(result.Code)))
			switch result.Code {
			case utils.G45_AT_PUBLIC_CODE:
				fmt.Println("Valid G45-AT (Public) Smart Contract.")
			case utils.G45_AT_PRIVATE_CODE:
				fmt.Println("Valid G45-AT (Private) Smart Contract.")
			case utils.G45_C_CODE:
				fmt.Println("Valid G45-C Smart Contract.")
			case utils.G45_NFT_PRIVATE_CODE:
				fmt.Println("Valid G45-NFT (Private) Smart Contract.")
			case utils.G45_NFT_PUBLIC_CODE:
				fmt.Println("Valid G45-NFT (Public) Smart Contract.")
			case utils.G45_FAT_PRIVATE_CODE:
				fmt.Println("Valid G45-FAT (Private) Smart Contract.")
			case utils.G45_FAT_PUBLIC_CODE:
				fmt.Println("Valid G45-FAT (Public) Smart Contract.")
			default:
				fmt.Println("Not a valid G45 Smart Contract.")
			}

			return nil
		},
	}
}

func App() *cli.App {
	return &cli.App{
		Name:        "g45-sc",
		Description: "Deploy & manage G45 Smart Contract.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			// G45-AT
			Command_G45_AT_View(),
			Command_G45_AT_Deploy(),
			Command_G45_AT_SetMetadata(),
			Command_G45_AT_SetCollection(),
			Command_G45_AT_Mint(),
			Command_G45_AT_Burn(),
			Command_G45_AT_Freeze(),
			Command_G45_AT_DisplayToken(),
			Command_G45_AT_RetrieveToken(),
			Command_G45_AT_TransferMinter(),
			Command_G45_AT_TransferMinter(),
			Command_G45_AT_CancelTransferMinter(),
			Command_G45_AT_ClaimMinter(),
			// G45-FAT
			Command_G45_FAT_View(),
			Command_G45_FAT_Deploy(),
			Command_G45_FAT_SetMetadata(),
			Command_G45_FAT_SetCollection(),
			Command_G45_FAT_Burn(),
			Command_G45_FAT_Freeze(),
			Command_G45_FAT_DisplayToken(),
			Command_G45_FAT_RetrieveToken(),
			// G45-NFT
			Command_G45_NFT_View(),
			Command_G45_NFT_Deploy(),
			Command_G45_NFT_DisplayNFT(),
			Command_G45_NFT_RetrieveNFT(),
			// G45-C
			Command_G45_C_View(),
			Command_G45_C_Deploy(),
			Command_G45_C_Freeze(),
			Command_G45_C_SetAssets(),
			Command_G45_C_DelAssets(),
			Command_G45_C_SetMetadata(),
			Command_G45_C_TransferOwnership(),
			Command_G45_C_CancelTransferOwnership(),
			Command_G45_C_ClaimOwnership(),
			Command_G45_C_DeployNFTs(),
			// Others
			CommandValidSC(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
