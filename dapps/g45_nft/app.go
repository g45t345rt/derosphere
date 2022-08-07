package g45_nft

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

func CommandDeployNFT() *cli.Command {
	return &cli.Command{
		Name:    "deploy",
		Aliases: []string{"d"},
		Usage:   "Deploy G45-NFT Smart Contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			assetType, err := app.PromptChoose("Asset token type", []string{"public", "private"}, "private")
			if app.HandlePromptErr(err) {
				return nil
			}

			code := utils.G45_NFT_PRIVATE
			if assetType == "public" {
				code = utils.G45_NFT_PUBLIC
			}

			txId, err := walletInstance.InstallSmartContract([]byte(code), true)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func CommandInitStoreNFT() *cli.Command {
	return &cli.Command{
		Name:    "init-store",
		Aliases: []string{"in"},
		Usage:   "Init store NFT (one time thing)",
		Action: func(ctx *cli.Context) error {
			nftSCID := ctx.Args().First()
			var err error

			if nftSCID == "" {
				nftSCID, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			collectionSCID, err := app.Prompt("Enter collection scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			supply, err := app.PromptUInt("Enter supply", 1)
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

			freezeSupply, err := app.PromptYesNo("Freeze supply?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeSupply := 0
			if freezeSupply {
				uFreezeSupply = 1
			}

			freezeCollection, err := app.PromptYesNo("Freeze collection?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeCollection := 0
			if freezeCollection {
				uFreezeCollection = 1
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, nftSCID, "InitStore", []rpc.Argument{
				{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
				{Name: "supply", DataType: rpc.DataUint64, Value: supply},
				{Name: "metadata", DataType: rpc.DataString, Value: metadata},
				{Name: "freezeCollection", DataType: rpc.DataUint64, Value: uFreezeCollection},
				{Name: "freezeSupply", DataType: rpc.DataUint64, Value: uFreezeSupply},
				{Name: "freezeMetadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
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

func CommandAddSupply() *cli.Command {
	return &cli.Command{
		Name:    "add-supply",
		Aliases: []string{"as"},
		Usage:   "Add supply to NFT",
		Action: func(ctx *cli.Context) error {
			nftSCID := ctx.Args().First()
			var err error

			if nftSCID == "" {
				nftSCID, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			supply, err := app.PromptUInt("Enter supply", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, nftSCID, "AddSupply", []rpc.Argument{
				{Name: "supply", DataType: rpc.DataUint64, Value: supply},
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
		Usage:   "Set NFT collection SCID supply to NFT",
		Action: func(ctx *cli.Context) error {
			nftSCID := ctx.Args().First()
			var err error

			if nftSCID == "" {
				nftSCID, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			collectionSCID, err := app.Prompt("Enter collection SCID", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, nftSCID, "SetCollection", []rpc.Argument{
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
		Usage:   "Freeze G45-NFT (supply, metadata or collection)",
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
		Usage:   "Set/edit metadata of the NFT",
		Action: func(ctx *cli.Context) error {
			nftSCID := ctx.Args().First()
			var err error

			if nftSCID == "" {
				nftSCID, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			metadata, err := app.Prompt("Set new metadata", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, nftSCID, "SetMetadata", []rpc.Argument{
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
		Usage:   "Publicly burn asset and edit supply",
		Action: func(ctx *cli.Context) error {
			nftSCID := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if nftSCID == "" {
				nftSCID, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			burnAmount, err := app.PromptUInt("Burn", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			transfer := rpc.Transfer{
				SCID: crypto.HashHexToHash(nftSCID),
				//Destination: randomAddresses.Address[0],
				Burn: burnAmount,
			}

			txId, err := walletInstance.CallSmartContract(2, nftSCID, "Burn", []rpc.Argument{}, []rpc.Transfer{
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
			nftSCID := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if nftSCID == "" {
				nftSCID, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			amount, err := app.PromptUInt("Amount", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			/*randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}*/

			transfer := rpc.Transfer{
				SCID: crypto.HashHexToHash(nftSCID),
				//Destination: randomAddresses.Address[0],
				Burn: amount,
			}

			txId, err := walletInstance.CallSmartContract(2, nftSCID, "DisplayToken", []rpc.Argument{}, []rpc.Transfer{
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
			nftSCID := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			var err error

			if nftSCID == "" {
				nftSCID, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			amount, err := app.PromptUInt("Amount", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			txId, err := walletInstance.CallSmartContract(2, nftSCID, "RetrieveToken", []rpc.Argument{
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

func CommandCheckValidNFT() *cli.Command {
	return &cli.Command{
		Name:    "check-valid",
		Aliases: []string{"cv"},
		Usage:   "Check if smart contract is a valid G45-NFT/G45-NFT-COLLECTION",
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
			case utils.G45_NFT_PUBLIC:
				fmt.Println("Valid G45-NFT-PUBLIC Smart Contract.")
			case utils.G45_NFT_PRIVATE:
				fmt.Println("Valid G45-NFT-PRIVATE Smart Contract.")
			case utils.G45_NFT_COLLECTION:
				fmt.Println("Valid G45-NFT-COLLECTION Smart Contract.")
			default:
				fmt.Println("Not a valid G45-NFT.")
			}

			return nil
		},
	}
}

func CommandDeployCollection() *cli.Command {
	return &cli.Command{
		Name:    "collection-deploy",
		Aliases: []string{"cd"},
		Usage:   "Deploy G45-NFT-COLLECTION Smart Contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.InstallSmartContract([]byte(utils.G45_NFT_COLLECTION), true)
			if err != nil {
				fmt.Println(err)
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

func CommandCollectionSetNFT() *cli.Command {
	return &cli.Command{
		Name:    "collection-set-nft",
		Aliases: []string{"cs"},
		Usage:   "Set new NFT to G45-NFT-COLLECTION",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			nftSCID, err := app.Prompt("Enter nft asset token", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			index, err := app.PromptUInt("Enter index", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "SetNft", []rpc.Argument{
				{Name: "nft", DataType: rpc.DataString, Value: nftSCID},
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

func CommandCollectionDelNFT() *cli.Command {
	return &cli.Command{
		Name:    "collection-del-nft",
		Aliases: []string{"cd"},
		Usage:   "Del NFT from G45-NFT-COLLECTION",
		Action: func(ctx *cli.Context) error {
			collectionSCID, err := app.Prompt("Enter collection scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			nftSCID, err := app.Prompt("Enter nft asset token", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, collectionSCID, "DelNft", []rpc.Argument{
				{Name: "nft", DataType: rpc.DataString, Value: nftSCID},
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
		Usage:   "Freeze G45-NFT-COLLECTION (collection/nfts or metadata)",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeCollection := 0
			freezeCollection, err := app.PromptYesNo("Freeze collection/nfts?", false)
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
		Usage:   "Set/edit metadata of the NFT Collection",
		Action: func(ctx *cli.Context) error {
			nftSCID := ctx.Args().First()
			var err error

			if nftSCID == "" {
				nftSCID, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			metadata, err := app.Prompt("Set new metadata", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, nftSCID, "SetMetadata", []rpc.Argument{
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

func CommandDeployEntireCollection() *cli.Command {
	return &cli.Command{
		Name:  "deploy-entire-collection",
		Usage: "Script to deploy entire NFT collection",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			assetType, err := app.PromptChoose("NFT type", []string{"public", "private"}, "private")
			if app.HandlePromptErr(err) {
				return nil
			}

			installCollection, err := app.PromptYesNo("Install G45-NFT-COLLECTION?", true)
			if app.HandlePromptErr(err) {
				return nil
			}

			var collectionSCID string
			if installCollection {
				fmt.Println("Install NFT Collection")
				collectionSCID, err = walletInstance.InstallSmartContract([]byte(utils.G45_NFT_COLLECTION), false)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				walletInstance.RunTxChecker(collectionSCID)
			} else {
				collectionSCID, err = app.Prompt("G45-NFT-COLLECTION asset token?", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			startIndex, err := app.PromptUInt("NFT start index", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			endIndex, err := app.PromptUInt("NFT end index", 99)
			if app.HandlePromptErr(err) {
				return nil
			}

			nftCode := utils.G45_NFT_PUBLIC
			if assetType == "private" {
				nftCode = utils.G45_NFT_PRIVATE
			}

			//sleep := 1000 * time.Millisecond
			for i := startIndex; i <= endIndex; i++ {
			install_nft:
				fmt.Printf("Install NFT - %d\n", i)
				nftSCID, err := walletInstance.InstallSmartContract([]byte(nftCode), false)
				if err != nil {
					fmt.Println(err)
					goto install_nft
				}

				err = walletInstance.WaitTransaction(nftSCID)
				if err != nil {
					fmt.Println(err)
					goto install_nft
				}

			set_collection:
				fmt.Println("Set to collection: " + nftSCID)
				setTxId, err := walletInstance.CallSmartContract(2, collectionSCID, "SetNft", []rpc.Argument{
					{Name: "nft", DataType: rpc.DataString, Value: nftSCID},
					{Name: "index", DataType: rpc.DataUint64, Value: i},
				}, []rpc.Transfer{}, false)

				if err != nil {
					fmt.Println(err)
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

func CommandInitStoreCollectionNFTs() *cli.Command {
	return &cli.Command{
		Name:  "init-store-nfts",
		Usage: "Script to initialize store of NFTs from metadata.json",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			metadata_path, err := app.Prompt("Enter metadata file path", "")
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

			collectionSCID, err := app.Prompt("G45-NFT-COLLECTION asset token?", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			supply, err := app.PromptUInt("Supply", 1)
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

			uFreezeSupply := 0
			freezeSupply, err := app.PromptYesNo("Freeze supply?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeSupply {
				uFreezeSupply = 1
			}

			uFreezeCollection := 0
			freezeCollection, err := app.PromptYesNo("Freeze collection?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeCollection {
				uFreezeCollection = 1
			}

			result, err := walletInstance.Daemon.GetSC(&rpc.GetSC_Params{
				Code:      true,
				Variables: true,
				SCID:      collectionSCID,
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			if result.Code != utils.G45_NFT_COLLECTION {
				fmt.Println("Not a valid G45-NFT-COLLECTION Smart Contract.")
				return nil
			}

			nftKey, _ := regexp.Compile(`nft_(.+)`)

			for key, value := range result.VariableStringKeys {
				if nftKey.Match([]byte(key)) {
					nftAssetToken := nftKey.ReplaceAllString(key, "$1")
					index := uint64(value.(float64))
					nft := metadataList[index]

					bMetadata, err := json.Marshal(nft)
					if err != nil {
						fmt.Println(err)
						continue
					}

					sMetadata := string(bMetadata)
					fmt.Println("InitStore: " + sMetadata)
					storeTxId, err := walletInstance.CallSmartContract(2, nftAssetToken, "InitStore", []rpc.Argument{
						{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
						{Name: "supply", DataType: rpc.DataUint64, Value: supply},
						{Name: "metadata", DataType: rpc.DataString, Value: sMetadata},
						{Name: "freezeCollection", DataType: rpc.DataUint64, Value: uFreezeCollection},
						{Name: "freezeSupply", DataType: rpc.DataUint64, Value: uFreezeSupply},
						{Name: "freezeMetadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
					}, []rpc.Transfer{}, false)

					if err != nil {
						fmt.Println(err)
						continue
					}

					err = walletInstance.WaitTransaction(storeTxId)
					if err != nil {
						fmt.Println(err)
						break
					}

					fmt.Println(storeTxId)
				}
			}

			return nil
		},
	}
}

func CommandViewNFT() *cli.Command {
	return &cli.Command{
		Name:    "view-nft",
		Aliases: []string{"vn"},
		Usage:   "Display nft metadata and more",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance

			nft, err := utils.GetG45NFT(scid, walletInstance.Daemon)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			nft.Print()
			return nil
		},
	}
}

func CommandViewNFTCollection() *cli.Command {
	return &cli.Command{
		Name:    "view-nft-collection",
		Aliases: []string{"vnc"},
		Usage:   "Display nft collection metadata and more",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			nftCollection, err := utils.GetG45NftCollection(scid, walletInstance.Daemon)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			nftCollection.Print()
			return nil
		},
	}
}

func App() *cli.App {
	return &cli.App{
		Name:        "g45-nft",
		Description: "Deploy & manage G45-NFT asset tokens.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandViewNFT(),
			CommandDeployNFT(),
			CommandDeployCollection(),
			CommandInitStoreNFT(),
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
			CommandCheckValidNFT(),
			CommandViewNFTCollection(),
			CommandCollectionTransferOwnership(),
			CommandCollectionCancelTransferOwnership(),
			CommandCollectionClaimOwnership(),
			CommandSetCollectionMetadata(),
			CommandCollectionSetNFT(),
			CommandCollectionDelNFT(),
			CommandCollectionFreeze(),
			CommandDeployEntireCollection(),
			CommandInitStoreCollectionNFTs(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
