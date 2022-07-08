package g45_nft

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

func CommandDeployNFT() *cli.Command {
	return &cli.Command{
		Name:    "deploy-nft",
		Aliases: []string{"d"},
		Usage:   "Deploy G45-NFT Smart Contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			assetType, err := app.PromptChoose("Asset token type", []string{"public", "private"}, "public")
			if app.HandlePromptErr(err) {
				return nil
			}

			code := utils.G45_NFT_PUBLIC
			if assetType == "private" {
				code = utils.G45_NFT_PRIVATE
			}

			txId, err := walletInstance.InstallSmartContract([]byte(code), true)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txId)
			return nil
		},
	}
}

func CommandInitNFT() *cli.Command {
	return &cli.Command{
		Name:    "init-nft",
		Aliases: []string{"in"},
		Usage:   "Init store NFT (one time thing)",
		Action: func(ctx *cli.Context) error {
			nftAssetToken := ctx.Args().First()
			var err error

			if nftAssetToken == "" {
				nftAssetToken, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			collectionSCID, err := app.Prompt("Enter collection scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			supply, err := app.PromptInt("Enter supply", 1)
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

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, nftAssetToken, "InitStore", []rpc.Argument{
				{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
				{Name: "supply", DataType: rpc.DataUint64, Value: supply},
				{Name: "metadata", DataType: rpc.DataString, Value: metadata},
				{Name: "frozenMetadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
				{Name: "frozenSupply", DataType: rpc.DataUint64, Value: uFreezeSupply},
			}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txId)
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
			nftAssetToken := ctx.Args().First()
			var err error

			if nftAssetToken == "" {
				nftAssetToken, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			supply, err := app.PromptInt("Enter supply", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, nftAssetToken, "AddSupply", []rpc.Argument{
				{Name: "supply", DataType: rpc.DataUint64, Value: supply},
			}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txId)

			return nil
		},
	}
}

func CommandFreezeSupply() *cli.Command {
	return &cli.Command{
		Name:    "freeze-supply",
		Aliases: []string{"fs"},
		Usage:   "Freeze supply of the NFT",
		Action: func(ctx *cli.Context) error {
			nftAssetToken := ctx.Args().First()
			var err error

			if nftAssetToken == "" {
				nftAssetToken, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, nftAssetToken, "FreezeSupply", []rpc.Argument{}, true)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txId)
			return nil
		},
	}
}

func CommandFreezeMetadata() *cli.Command {
	return &cli.Command{
		Name:    "freeze-metadata",
		Aliases: []string{"fm"},
		Usage:   "Freeze metadata of the NFT",
		Action: func(ctx *cli.Context) error {
			nftAssetToken := ctx.Args().First()
			var err error

			if nftAssetToken == "" {
				nftAssetToken, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, nftAssetToken, "FreezeMetadata", []rpc.Argument{}, true)
			if err != nil {
				fmt.Println(txId)
				return nil
			}

			fmt.Println(txId)
			return nil
		},
	}
}

func CommandSetMetadata() *cli.Command {
	return &cli.Command{
		Name:    "set-metadata",
		Aliases: []string{"fm"},
		Usage:   "Set/edit metadata of the NFT",
		Action: func(ctx *cli.Context) error {
			nftAssetToken := ctx.Args().First()
			var err error

			if nftAssetToken == "" {
				nftAssetToken, err = app.Prompt("Enter nft asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			metadata, err := app.Prompt("Set new metadata", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, nftAssetToken, "SetMetadata", []rpc.Argument{
				{Name: "metadata", DataType: rpc.DataString, Value: metadata},
			}, true)

			if err != nil {
				fmt.Println(txId)
				return nil
			}

			fmt.Println(txId)
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
				fmt.Println("Valid G45-NFT (PUBLIC) Smart Contract.")
			case utils.G45_NFT_PRIVATE:
				fmt.Println("Valid G45-NFT (PRIVATE) Smart Contract.")
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
		Name:    "deploy-collection",
		Aliases: []string{"dc"},
		Usage:   "Deploy G45-NFT-COLLECTION Smart Contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			walletInstance.InstallSmartContract([]byte(utils.G45_NFT_COLLECTION), true)
			return nil
		},
	}
}

func CommandCollectionAddNFT() *cli.Command {
	return &cli.Command{
		Name:    "collection-add-nft",
		Aliases: []string{"ca"},
		Usage:   "Add new NFT to G45-NFT-COLLECTION",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			nftAssetToken, err := app.Prompt("Enter nft asset token", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "Add", []rpc.Argument{
				{Name: "nft", DataType: rpc.DataString, Value: nftAssetToken},
			}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txId)
			return nil
		},
	}
}

func CommandSetupWizard() *cli.Command {
	return &cli.Command{
		Name:  "setup-wizard",
		Usage: "Script to deploy entire collection from metadata.json",
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

			var metadataCollection utils.NFTMetadataCollection
			err = json.Unmarshal(content, &metadataCollection)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			/*
				rarity_path, err := app.Prompt("Enter rarity csv file path", "")
				if app.HandlePromptErr(err) {
					return nil
				}

				rarity_file, err := os.Open(rarity_path)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				reader := csv.NewReader(rarity_file)
				rarity_records, _ := reader.ReadAll()
			*/

			assetType, err := app.PromptChoose("NFT type", []string{"public", "private"}, "public")
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
					log.Fatal(err)
				}
				fmt.Println(collectionSCID)
			} else {
				collectionSCID, err = app.Prompt("G45-NFT-COLLECTION asset token?", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			startIndex, err := app.PromptInt("NFT start index", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			nftCode := utils.G45_NFT_PUBLIC
			if assetType == "private" {
				nftCode = utils.G45_NFT_PRIVATE
			}

			for index, nft := range metadataCollection.Collection {
				if int64(index) >= startIndex {
					fmt.Println("Install NFT")
					nftSCID, err := walletInstance.InstallSmartContract([]byte(nftCode), false)
					if err != nil {
						log.Fatal(err)
					}

					time.Sleep(250 * time.Millisecond)

					fmt.Println("Add to collection: " + nftSCID)
					_, err = walletInstance.CallSmartContract(2, collectionSCID, "Add", []rpc.Argument{
						{Name: "nft", DataType: rpc.DataString, Value: nftSCID},
					}, false)

					if err != nil {
						log.Fatal(err)
					}

					time.Sleep(250 * time.Millisecond)
					sMetadata := ""
					for _, attr := range nft.Attributes {
						sMetadata = sMetadata + "&" + attr.TraitType + "=" + attr.Value
					}

					fmt.Println("InitStore: " + sMetadata)
					_, err = walletInstance.CallSmartContract(2, nftSCID, "InitStore", []rpc.Argument{
						{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
						{Name: "supply", DataType: rpc.DataUint64, Value: 1},
						{Name: "metadata", DataType: rpc.DataString, Value: sMetadata},
						{Name: "frozenMetadata", DataType: rpc.DataUint64, Value: 0},
						{Name: "frozenSupply", DataType: rpc.DataUint64, Value: 0},
					}, false)

					if err != nil {
						log.Fatal(err)
					}

					time.Sleep(250 * time.Millisecond)
				}
			}

			return nil
		},
	}
}

func App() *cli.App {
	return &cli.App{
		Name:        "g45-nft",
		Description: "Deploy & manage G45-NFT asset tokens",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandDeployNFT(),
			CommandDeployCollection(),
			CommandInitNFT(),
			CommandAddSupply(),
			CommandSetMetadata(),
			CommandFreezeMetadata(),
			CommandFreezeSupply(),
			CommandCheckValidNFT(),
			CommandCollectionAddNFT(),
			CommandSetupWizard(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
