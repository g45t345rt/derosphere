package g45_at

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

func CommandDeploy() *cli.Command {
	return &cli.Command{
		Name:    "deploy",
		Aliases: []string{"d"},
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

func CommandDeployNFT() *cli.Command {
	return &cli.Command{
		Name:    "deploy-nft",
		Aliases: []string{"dn"},
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

func CommandAddSupply() *cli.Command {
	return &cli.Command{
		Name:    "mint",
		Aliases: []string{"mt"},
		Usage:   "Increase supply",
		Action: func(ctx *cli.Context) error {
			scid := ctx.Args().First()
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter scid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			supply, err := app.PromptUInt("Enter quantity", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "Mint", []rpc.Argument{
				{Name: "qty", DataType: rpc.DataUint64, Value: supply},
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

func CommandSetCollection() *cli.Command {
	return &cli.Command{
		Name:    "set-collection",
		Aliases: []string{"sc"},
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

func CommandFreeze() *cli.Command {
	return &cli.Command{
		Name:    "freeze",
		Aliases: []string{"f"},
		Usage:   "Freeze G45-AT (supply, metadata or collection)",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeSupply := 0
			freezeSupply, err := app.PromptYesNo("Freeze supply?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeSupply {
				uFreezeSupply = 1
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
				{Name: "supply", DataType: rpc.DataUint64, Value: uFreezeSupply},
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

func CommandSetMetadata() *cli.Command {
	return &cli.Command{
		Name:    "set-metadata",
		Aliases: []string{"sm"},
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

func CommandBurn() *cli.Command {
	return &cli.Command{
		Name:    "burn",
		Aliases: []string{"dt"},
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

func CommandDisplayToken() *cli.Command {
	return &cli.Command{
		Name:    "display-token",
		Aliases: []string{"dt"},
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

func CommandRetrieveToken() *cli.Command {
	return &cli.Command{
		Name:    "retrieve-token",
		Aliases: []string{"rt"},
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

func CommandTransferMinter() *cli.Command {
	return &cli.Command{
		Name:    "transfer-minter",
		Aliases: []string{"tm"},
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

func CommanCancelTransferMinter() *cli.Command {
	return &cli.Command{
		Name:    "cancel-transfer-minter",
		Aliases: []string{"ctm"},
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

func CommandClaimMinter() *cli.Command {
	return &cli.Command{
		Name:    "claim-minter",
		Aliases: []string{"cm"},
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

func CommandCheckValid() *cli.Command {
	return &cli.Command{
		Name:    "check-valid",
		Aliases: []string{"cv"},
		Usage:   "Check if smart contract is a valid G45-AT/G45-ATC",
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
				fmt.Println("Valid G45-ATC Smart Contract.")
			case utils.G45_NFT_PRIVATE_CODE:
				fmt.Println("Valid G45-NFT (Private) Smart Contract.")
			case utils.G45_NFT_PUBLIC_CODE:
				fmt.Println("Valid G45-NFT (Public) Smart Contract.")
			default:
				fmt.Println("Not a valid G45 Smart Contract.")
			}

			return nil
		},
	}
}

func DoCollectionDeploy() (string, error) {
	walletInstance := app.Context.WalletInstance

	metadataFormat, err := app.Prompt("Metadata format?", "json")
	if err != nil {
		return "", err
	}

	metadata, err := app.Prompt("Set metadata", "")
	if err != nil {
		return "", err
	}

	uFreezeCollection := 0
	freezeCollection, err := app.PromptYesNo("Freeze collection?", false)
	if err != nil {
		return "", err
	}

	if freezeCollection {
		uFreezeCollection = 1
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
		{Name: "freezeCollection", DataType: rpc.DataUint64, Value: uFreezeCollection},
		{Name: "freezeMetadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
	}, true)

	return txId, err
}

func CommandDeployCollection() *cli.Command {
	return &cli.Command{
		Name:    "collection-deploy",
		Aliases: []string{"cd"},
		Usage:   "Deploy G45-ATC Smart Contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			txId, err := DoCollectionDeploy()
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func CommandCollectionTransferOwnership() *cli.Command {
	return &cli.Command{
		Name:    "collection-transfer-ownership",
		Aliases: []string{"cto"},
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

func CommandCollectionClaimOwnership() *cli.Command {
	return &cli.Command{
		Name:    "collection-claim-ownership",
		Aliases: []string{"cco"},
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

func CommandCollectionCancelTransferOwnership() *cli.Command {
	return &cli.Command{
		Name:    "collection-cancel-transfer-ownership",
		Aliases: []string{"ccto"},
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

func CommandCollectionSetAsset() *cli.Command {
	return &cli.Command{
		Name:    "collection-set-asset",
		Aliases: []string{"cs"},
		Usage:   "Set new asset to G45-ATC",
		Action: func(ctx *cli.Context) error {
			collectionSCID, err := app.Prompt("Enter collection scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			assetSCID, err := app.Prompt("Enter asset token", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			index, err := app.PromptUInt("Enter index", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, collectionSCID, "SetAsset", []rpc.Argument{
				{Name: "asset", DataType: rpc.DataString, Value: assetSCID},
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

func CommandCollectionDelAsset() *cli.Command {
	return &cli.Command{
		Name:    "collection-del-asset",
		Aliases: []string{"cd"},
		Usage:   "Del asset from G45-ATC",
		Action: func(ctx *cli.Context) error {
			collectionSCID, err := app.Prompt("Enter collection scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			assetSCID, err := app.Prompt("Enter asset token", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, collectionSCID, "DelAsset", []rpc.Argument{
				{Name: "asset", DataType: rpc.DataString, Value: assetSCID},
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

func CommandCollectionFreeze() *cli.Command {
	return &cli.Command{
		Name:    "collection-freeze",
		Aliases: []string{"cf"},
		Usage:   "Freeze G45-ATC (collection/nfts or metadata)",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeCollection := 0
			freezeCollection, err := app.PromptYesNo("Freeze collection/assets/nfts?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeCollection {
				uFreezeCollection = 1
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
				{Name: "collection", DataType: rpc.DataUint64, Value: uFreezeCollection},
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

func CommandSetCollectionMetadata() *cli.Command {
	return &cli.Command{
		Name:    "collection-set-metadata",
		Aliases: []string{"csm"},
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

func CommandDeployNFTCollection() *cli.Command {
	return &cli.Command{
		Name:  "deploy-nft-collection",
		Usage: "Script to deploy nft collection",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			installCollection, err := app.PromptYesNo("Install G45-C?", true)
			if app.HandlePromptErr(err) {
				return nil
			}

			var collectionSCID string
			if installCollection {
				collectionSCID, err = DoCollectionDeploy()
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

			metadata_path, err := app.Prompt("Enter nfts metadata file path", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			content, err := ioutil.ReadFile(metadata_path)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			var metadataList []interface{}
			err = json.Unmarshal(content, &metadataList)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			assetType, err := app.PromptChoose("G45-NFT type", []string{"public", "private"}, "private")
			if app.HandlePromptErr(err) {
				return nil
			}

			scCode := utils.G45_NFT_PUBLIC_CODE
			if assetType == "private" {
				scCode = utils.G45_NFT_PRIVATE_CODE
			}

			startIndex, err := app.PromptUInt("Asset start index", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			endIndex, err := app.PromptUInt("Asset end index", 99)
			if app.HandlePromptErr(err) {
				return nil
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
				fmt.Printf("Install Asset - %d\n", index)
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
					goto install_asset
				}

			set_collection:
				fmt.Println("Set to collection: " + assetSCID)
				setTxId, err := walletInstance.CallSmartContract(2, collectionSCID, "SetAsset", []rpc.Argument{
					{Name: "asset", DataType: rpc.DataString, Value: assetSCID},
					{Name: "index", DataType: rpc.DataUint64, Value: index},
				}, []rpc.Transfer{}, false)

				if err != nil {
					fmt.Println(err)
					time.Sleep(2 * time.Second)
					goto set_collection
				}

				err = walletInstance.WaitTransaction(setTxId)
				if err != nil {
					fmt.Println(err)
					goto set_collection
				}
			}

			return nil
		},
	}
}

func CommandView() *cli.Command {
	return &cli.Command{
		Name:    "view",
		Aliases: []string{"v"},
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

func CommandViewCollection() *cli.Command {
	return &cli.Command{
		Name:    "view-collection",
		Aliases: []string{"vc"},
		Usage:   "Display G45-ATC metadata and more",
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

func App() *cli.App {
	return &cli.App{
		Name:        "g45-at",
		Description: "Deploy & manage G45-AT Smart Contract.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandView(),
			CommandDeploy(),
			CommandDeployNFT(),
			CommandDeployCollection(),
			CommandAddSupply(),
			CommandBurn(),
			CommandSetMetadata(),
			CommandSetCollection(),
			CommandDisplayToken(),
			CommandRetrieveToken(),
			CommandFreeze(),
			CommandTransferMinter(),
			CommanCancelTransferMinter(),
			CommandClaimMinter(),
			CommandCheckValid(),
			CommandViewCollection(),
			CommandCollectionTransferOwnership(),
			CommandCollectionCancelTransferOwnership(),
			CommandCollectionClaimOwnership(),
			CommandSetCollectionMetadata(),
			CommandCollectionSetAsset(),
			CommandCollectionDelAsset(),
			CommandCollectionFreeze(),
			CommandDeployNFTCollection(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
