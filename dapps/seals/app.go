package seals

import (
	"fmt"
	"log"
	"net/url"

	"github.com/deroproject/derohe/rpc"
	"github.com/fatih/color"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/rpc_client"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/pkg/browser"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

var DAPP_NAME = "seals"

var COLLECTION_SC_ID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "",
	"simulator": "d8cc9975a9ba8db60623515d11547a28cf2974babe4b779d55153f2d5805d781",
}

func getCollectionSCID() string {
	return COLLECTION_SC_ID[app.Context.Config.Env]
}

type SealNFT struct {
	Id               string
	FrozenMetadata   bool
	FrozenSupply     bool
	Supply           uint64
	FileNumber       int
	Rarity           float64
	TraitBackground  string
	TraitBase        string
	TraitEyes        string
	TraitHairAndHats string
	TraitShirts      string
	TraitTattoo      string
}

func (sn *SealNFT) Traits() string {
	return sn.TraitBackground + ", " + sn.TraitBase + ", " + sn.TraitEyes + ", " + sn.TraitHairAndHats + ", " + sn.TraitShirts + ", " + sn.TraitTattoo
}

func initData() {
	sqlQuery := `
		create table if not exists seals_collection (
			nft varchar primary key,
			frozen_metadata boolean,
			frozen_supply boolean,
			supply integer,
			file_number integer,
			rarity real,
			trait_background varchar,
			trait_base varchar,
			trait_eyes varchar,
			trait_hairAndHats varchar,
			trait_shirts varchar,
			trait_tattoo varchar
		);
	`

	db := app.Context.DB

	_, err := db.Exec(sqlQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func sync() {
	daemon := app.Context.WalletInstance.Daemon
	scid := getCollectionSCID()
	itemCount := daemon.GetSCItemCount(scid, "nft_count")
	//counts := utils.GetCounts()
	itemAt := uint64(0) //counts[DAPP_NAME]
	chunk := uint64(1000)
	db := app.Context.DB

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	var i uint64
	for i = itemAt; i < itemCount; i += chunk {
		var items []rpc_client.Item
		end := i + chunk
		if end > itemCount {
			itemAt = itemCount
			items = daemon.GetSCItems(scid, "nft_", i, itemCount)
		} else {
			itemAt = end
			items = daemon.GetSCItems(scid, "nft_", i, itemAt)
		}

		for _, item := range items {
			nftId := item.Value
			result, err := daemon.GetSC(&rpc.GetSC_Params{
				SCID:      nftId,
				Code:      false,
				Variables: true,
			})

			if err != nil {
				log.Fatal(err)
			}

			nft := utils.ParseG45NFT(nftId, result)

			values, err := url.ParseQuery(nft.Metadata)
			if err != nil {
				log.Fatal(err)
			}

			query := `
				insert into seals_collection (nft, frozen_metadata,	frozen_supply, supply, file_number,
					rarity, trait_background, trait_base, trait_eyes, trait_hairAndHats, trait_shirts, trait_tattoo)
				values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				on conflict(nft) do update 
				set nft = ?
			`

			_, err = tx.Exec(query, nft.Id, nft.FrozenMetadata, nft.FrozenSupply, nft.Supply,
				values.Get("fileNumber"), values.Get("rarity"), values.Get("trait_background"),
				values.Get("trait_base"), values.Get("trait_eyes"), values.Get("trait_hairAndHats"),
				values.Get("trait_shirts"), values.Get("trait_tattoo"), nft.Id,
			)

			if err != nil {
				log.Fatal(err)
			}
		}

		// we if ticket price null means a lotto has been cancelled and deleted from sc
		//_, err := tx.Exec("delete from lotto where ticket_price is null")
		//	if err != nil {
		//	log.Fatal(err)
		//}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

		utils.SetCount(DAPP_NAME, itemAt)
	}
}

func displayNFTTable(nfts []SealNFT) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("", "NFT", "File Number", "Rarity", "Traits")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for index, nft := range nfts {
		tbl.AddRow(index, nft.Id, nft.FileNumber, nft.Rarity, nft.Traits())
	}

	tbl.Print()
	if len(nfts) == 0 {
		fmt.Println("No NFTs yet.")
	}
}

func CommandList() *cli.Command {
	return &cli.Command{
		Name:    "list-collection",
		Aliases: []string{"l"},
		Usage:   "List NFT collection",
		Action: func(c *cli.Context) error {
			sync()

			db := app.Context.DB
			query := `select * from seals_collection`

			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
			}

			var nfts []SealNFT
			for rows.Next() {
				var nft SealNFT
				err = rows.Scan(&nft.Id, &nft.FrozenMetadata, &nft.FrozenSupply, &nft.Supply,
					&nft.FileNumber, &nft.Rarity, &nft.TraitBackground, &nft.TraitBase, &nft.TraitEyes,
					&nft.TraitHairAndHats, &nft.TraitShirts, &nft.TraitTattoo,
				)

				if err != nil {
					log.Fatal(err)
				}

				nfts = append(nfts, nft)
			}

			displayNFTTable(nfts)
			return nil
		},
	}
}

func CommandViewNFT() *cli.Command {
	return &cli.Command{
		Name:    "view-nft",
		Aliases: []string{"vn"},
		Usage:   "Open NFT image with asset token",
		Action: func(ctx *cli.Context) error {
			sync()

			nft := ctx.Args().First()
			var err error

			if nft == "" {
				nft, err = app.Prompt("Enter nft", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			db := app.Context.DB
			query := `select file_number from seals_collection where nft = ?`

			row := db.QueryRow(query, nft)
			err = row.Err()
			if err != nil {
				fmt.Println(err)
				return nil
			}

			var fileNumber string
			row.Scan(&fileNumber)

			browser.OpenURL("https://imagedelivery.net/zAjZFa6f2RjCu5A0cXIeHA/dero-seals-" + fileNumber + "/default")
			return nil
		},
	}
}

func CommandViewImage() *cli.Command {
	return &cli.Command{
		Name:    "view-image",
		Aliases: []string{"vi"},
		Usage:   "Open NFT image with file number",
		Action: func(ctx *cli.Context) error {
			fileNumber := ctx.Args().First()
			var err error

			if fileNumber == "" {
				fileNumber, err = app.Prompt("Enter file number", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			browser.OpenURL("https://imagedelivery.net/zAjZFa6f2RjCu5A0cXIeHA/dero-seals-" + fileNumber + "/default")
			return nil
		},
	}
}

func App() *cli.App {
	initData()

	return &cli.App{
		Name:        "seals",
		Description: "Dero Seals NFT project.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandList(),
			CommandViewNFT(),
			CommandViewImage(),
		},
		Authors: []*cli.Author{
			{Name: "billoetree"},
			{Name: "MERU"},
			{Name: "g45t345rt"},
		},
	}
}
