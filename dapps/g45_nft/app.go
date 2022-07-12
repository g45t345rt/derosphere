package g45_nft

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/rpc_client"
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
			assetType, err := app.PromptChoose("Asset token type", []string{"public", "private"}, "private")
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

func CommandInitStoreNFT() *cli.Command {
	return &cli.Command{
		Name:    "init-store",
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
				{Name: "freezeMetadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
				{Name: "freezeSupply", DataType: rpc.DataUint64, Value: uFreezeSupply},
			}, []rpc.Transfer{}, true)

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
			}, []rpc.Transfer{}, true)

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
			txId, err := walletInstance.CallSmartContract(2, nftAssetToken, "FreezeSupply", []rpc.Argument{}, []rpc.Transfer{}, true)
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
			txId, err := walletInstance.CallSmartContract(2, nftAssetToken, "FreezeMetadata", []rpc.Argument{}, []rpc.Transfer{}, true)
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
		Aliases: []string{"sm"},
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
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
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
			txId, err := walletInstance.InstallSmartContract([]byte(utils.G45_NFT_COLLECTION), true)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txId)
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

			nftAssetToken, err := app.Prompt("Enter nft asset token", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			index, err := app.PromptInt("Enter index", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "Set", []rpc.Argument{
				{Name: "nft", DataType: rpc.DataString, Value: nftAssetToken},
				{Name: "index", DataType: rpc.DataUint64, Value: index},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txId)
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

			nftAssetToken, err := app.Prompt("Enter nft asset token", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, collectionSCID, "Del", []rpc.Argument{
				{Name: "nft", DataType: rpc.DataString, Value: nftAssetToken},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txId)
			return nil
		},
	}
}

func CommandCollectionLock() *cli.Command {
	return &cli.Command{
		Name:    "collection-lock",
		Aliases: []string{"cl"},
		Usage:   "Lock G45-NFT-COLLECTION - can't add or delete nfts",
		Action: func(ctx *cli.Context) error {
			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			txId, err := walletInstance.CallSmartContract(2, scid, "Lock", []rpc.Argument{}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txId)
			return nil
		},
	}
}

func CommandDeployEntireCollection() *cli.Command {
	return &cli.Command{
		Name:  "deploy-entire-collection",
		Usage: "Script to deploy entire NFT collection",
		Action: func(ctx *cli.Context) error {
			app.Context.StopInactivityTimer = true
			app.Context.StopPromptRefresh = true

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

			endIndex, err := app.PromptInt("NFT end index", 99)
			if app.HandlePromptErr(err) {
				return nil
			}

			nftCode := utils.G45_NFT_PUBLIC
			if assetType == "private" {
				nftCode = utils.G45_NFT_PRIVATE
			}

			sleep := 1000 * time.Millisecond

			for i := startIndex; i <= endIndex; i++ {
				fmt.Printf("Install NFT - %d\n", i)
				nftSCID, err := walletInstance.InstallSmartContract([]byte(nftCode), false)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(nftSCID)
				time.Sleep(sleep)

				fmt.Println("Set to collection: " + nftSCID)
				setTxId, err := walletInstance.CallSmartContract(2, collectionSCID, "Set", []rpc.Argument{
					{Name: "nft", DataType: rpc.DataString, Value: nftSCID},
					{Name: "index", DataType: rpc.DataUint64, Value: i},
				}, []rpc.Transfer{}, false)

				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(setTxId)
				time.Sleep(sleep)
			}

			app.Context.StopInactivityTimer = false
			app.Context.StopPromptRefresh = false
			return nil
		},
	}
}

func CommandInitStoreCollectionNFTs() *cli.Command {
	return &cli.Command{
		Name:  "init-collection-nfts",
		Usage: "Script to deploy collection NFTs from metadata.json",
		Action: func(ctx *cli.Context) error {
			app.Context.StopInactivityTimer = true
			app.Context.StopPromptRefresh = true

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

			collectionSCID, err := app.Prompt("G45-NFT-COLLECTION asset token?", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			uFreezeMetadata := 0
			freezeMetadata, err := app.PromptYesNo("Freeze metadata", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeMetadata {
				uFreezeMetadata = 1
			}

			uFreezeSupply := 0
			freezeSupply, err := app.PromptYesNo("Freeze supply", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if freezeSupply {
				uFreezeSupply = 1
			}

			sleep := 1000 * time.Millisecond

			result, err := walletInstance.Daemon.GetSC(&rpc.GetSC_Params{
				Code:      true,
				Variables: false,
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

			daemon := walletInstance.Daemon
			commitCount := daemon.GetSCCommitCount(collectionSCID)
			commitAt := uint64(0)
			chunk := uint64(1000)
			nftKey, _ := regexp.Compile(`state_nft_(.+)`)

			for i := commitAt; i < commitCount; i += chunk {
				var commits []rpc_client.Commit
				end := i + chunk
				if end > commitCount {
					commitAt = commitCount
					commits = daemon.GetSCCommits(collectionSCID, i, commitCount)
				} else {
					commitAt = end
					commits = daemon.GetSCCommits(collectionSCID, i, commitAt)
				}

				for _, commit := range commits {
					key := commit.Key

					if strings.HasPrefix(key, "state_nft_") {
						nftAssetToken := nftKey.ReplaceAllString(key, "$1")
						index, err := strconv.ParseUint(commit.Value, 10, 64)
						if err != nil {
							log.Fatal(err)
						}

						nft := metadataCollection.Collection[index]

						sMetadata := ""
						for _, attr := range nft.Attributes {
							sMetadata = sMetadata + "&" + attr.TraitType + "=" + attr.Value
						}

						fmt.Println("InitStore: " + sMetadata)
						storeTxId, err := walletInstance.CallSmartContract(2, nftAssetToken, "InitStore", []rpc.Argument{
							{Name: "collection", DataType: rpc.DataString, Value: collectionSCID},
							{Name: "supply", DataType: rpc.DataUint64, Value: 1},
							{Name: "metadata", DataType: rpc.DataString, Value: sMetadata},
							{Name: "freezeMetadata", DataType: rpc.DataUint64, Value: uFreezeMetadata},
							{Name: "freezeSupply", DataType: rpc.DataUint64, Value: uFreezeSupply},
						}, []rpc.Transfer{}, false)

						if err != nil {
							fmt.Println(err)
							continue
						}

						fmt.Println(storeTxId)
						time.Sleep(sleep)
					}
				}
			}

			app.Context.StopInactivityTimer = false
			app.Context.StopPromptRefresh = false
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
			CommandDeployNFT(),
			CommandDeployCollection(),
			CommandInitStoreNFT(),
			CommandAddSupply(),
			CommandSetMetadata(),
			CommandFreezeMetadata(),
			CommandFreezeSupply(),
			CommandCheckValidNFT(),
			CommandCollectionSetNFT(),
			CommandCollectionDelNFT(),
			CommandCollectionLock(),
			CommandDeployEntireCollection(),
			CommandInitStoreCollectionNFTs(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
