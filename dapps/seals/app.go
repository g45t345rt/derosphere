package seals

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/pkg/browser"
	"github.com/urfave/cli/v2"
)

var DAPP_NAME = "seals"

var COLLECTION_SC_ID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "e1f37876324f8692dd126131931236c90d65417b9d61665e4ea8f849931468f4",
	"simulator": "bcda6eff38e9d8c91ab441459c47bcf20196b7431760ab871f53fd27175f059c",
}

func getCollectionSCID() string {
	return COLLECTION_SC_ID[app.Context.Config.Env]
}

type SealNFT struct {
	Token            sql.NullString
	FrozenMetadata   sql.NullBool
	FrozenSupply     sql.NullBool
	Supply           sql.NullInt64
	Metadata         sql.NullString
	Id               sql.NullInt64
	Rarity           sql.NullFloat64
	TraitBackground  sql.NullString
	TraitBase        sql.NullString
	TraitEyes        sql.NullString
	TraitHairAndHats sql.NullString
	TraitShirts      sql.NullString
	TraitTattoo      sql.NullString
	TraitFacialHair  sql.NullString
}

type NFTMetadata struct {
	Id         uint64            `json:"id"`
	Rarity     float64           `json:"rarity"`
	Attributes map[string]string `json:"attributes"`
}

func emptyStringToUnderscore(value string) string {
	if value == "" {
		return "_"
	}

	return value
}

func (sn *SealNFT) Traits() string {
	var traits []string
	traits = append(traits,
		emptyStringToUnderscore(sn.TraitBackground.String),
		emptyStringToUnderscore(sn.TraitBase.String),
		emptyStringToUnderscore(sn.TraitEyes.String),
		emptyStringToUnderscore(sn.TraitHairAndHats.String),
		emptyStringToUnderscore(sn.TraitShirts.String),
		emptyStringToUnderscore(sn.TraitTattoo.String),
		emptyStringToUnderscore(sn.TraitFacialHair.String),
	)

	return strings.Join(traits, ", ")
}

func initData() {
	query := `
		create table if not exists dapps_seals_collection (
			scid varchar primary key,
			frozen_metadata boolean,
			frozen_supply boolean,
			supply bigint,
			metadata string,
			id integer,
			rarity real,
			trait_background varchar,
			trait_base varchar,
			trait_eyes varchar,
			trait_hairAndHats varchar,
			trait_shirts varchar,
			trait_tattoo varchar,
			trait_facialHair varchar
		);
	`

	db := app.Context.DB

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func update() error {
	daemon := app.Context.WalletInstance.Daemon
	collectionSCID := getCollectionSCID()

	db := app.Context.DB
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`delete from dapps_seals_collection;`)
	if err != nil {
		return err
	}

	collection := utils.G45_C{}
	result, err := daemon.GetSC(&rpc.GetSC_Params{
		SCID:      collectionSCID,
		Code:      true,
		Variables: true,
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	err = collection.Parse(collectionSCID, result)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for assetSCID := range collection.Assets {
		nft := utils.G45_NFT{}

		result, err := daemon.GetSC(&rpc.GetSC_Params{
			SCID:      assetSCID,
			Code:      true,
			Variables: true,
		})
		if err != nil {
			fmt.Printf("%s %s\n", assetSCID, err.Error())
			continue
		}

		err = nft.Parse(assetSCID, result)
		if err != nil {
			fmt.Printf("%s %s\n", assetSCID, err.Error())
			continue
		}

		var metadata NFTMetadata
		err = json.Unmarshal([]byte(nft.Metadata), &metadata)
		if err != nil {
			fmt.Println(err)
			continue
		}

		query := `
					insert into dapps_seals_collection (scid, metadata, id,
						rarity, trait_background, trait_base, trait_eyes, trait_hairAndHats, trait_shirts, trait_tattoo, trait_facialHair)
					values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				`

		_, err = tx.Exec(query, nft.SCID, nft.Metadata,
			metadata.Id, metadata.Rarity, metadata.Attributes["background"],
			metadata.Attributes["base"], metadata.Attributes["eyes"], metadata.Attributes["hair_and_hats"],
			metadata.Attributes["shirts"], metadata.Attributes["tattoo"], metadata.Attributes["facial_hair"],
		)

		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func CommandUpdate() *cli.Command {
	return &cli.Command{
		Name:    "update",
		Aliases: []string{"u"},
		Usage:   "Update collection and all nfts",
		Action: func(ctx *cli.Context) error {
			err := update()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("List updated.")
			}
			return nil
		},
	}
}

func CommandList() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "List NFT collection",
		Action: func(c *cli.Context) error {
			db := app.Context.DB
			query := `select scid, frozen_metadata, frozen_supply, supply, metadata, id, rarity,
			trait_background, trait_base, trait_eyes, trait_hairAndHats, trait_shirts, trait_tattoo
			from dapps_seals_collection order by rarity desc`

			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
			}

			var nfts []SealNFT
			for rows.Next() {
				var nft SealNFT
				err = rows.Scan(&nft.Token, &nft.FrozenMetadata, &nft.FrozenSupply, &nft.Supply, &nft.Metadata,
					&nft.Id, &nft.Rarity, &nft.TraitBackground, &nft.TraitBase, &nft.TraitEyes,
					&nft.TraitHairAndHats, &nft.TraitShirts, &nft.TraitTattoo,
				)

				if err != nil {
					log.Fatal(err)
				}

				nfts = append(nfts, nft)
			}

			app.Context.DisplayTable(len(nfts), func(i int) []interface{} {
				nft := nfts[i]
				return []interface{}{
					nft.Token.String, nft.Supply.Int64, nft.FrozenMetadata.Bool, nft.FrozenSupply.Bool, nft.Id.Int64, nft.Rarity.Float64, nft.Traits(),
				}
			}, []interface{}{"NFT", "Supply", "Frozen Metadata", "Frozen Supply", "Id", "Rarity", "Traits"}, 25)
			return nil
		},
	}
}

func CommandCount() *cli.Command {
	return &cli.Command{
		Name:    "count",
		Aliases: []string{"c"},
		Usage:   "Number of NFTs in the collection",
		Action: func(c *cli.Context) error {
			db := app.Context.DB
			query := `select count(*) from dapps_seals_collection`

			row := db.QueryRow(query)
			var count int
			err := row.Scan(&count)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(count)
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
			nft := ctx.Args().First()
			var err error

			if nft == "" {
				nft, err = app.Prompt("Enter nft", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			db := app.Context.DB
			query := `select id from dapps_seals_collection where token = ?`

			row := db.QueryRow(query, nft)
			err = row.Err()
			if err != nil {
				fmt.Println(err)
				return nil
			}

			var id string
			row.Scan(&id)

			browser.OpenURL("https://imagedelivery.net/zAjZFa6f2RjCu5A0cXIeHA/dero-seals-" + id + "/default")
			fmt.Println("NFT image opened in the browser.")
			return nil
		},
	}
}

func CommandViewImage() *cli.Command {
	return &cli.Command{
		Name:    "view-image",
		Aliases: []string{"vi"},
		Usage:   "Open NFT image with ID",
		Action: func(ctx *cli.Context) error {
			id := ctx.Args().First()
			var err error

			if id == "" {
				id, err = app.Prompt("Enter ID", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			browser.OpenURL("https://imagedelivery.net/zAjZFa6f2RjCu5A0cXIeHA/dero-seals-" + id + "/default")
			fmt.Println("NFT image opened in the browser.")
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
			CommandUpdate(),
			CommandList(),
			CommandViewNFT(),
			CommandViewImage(),
			CommandCount(),
		},
		Authors: []*cli.Author{
			{Name: "billoetree"},
			{Name: "MERU"},
			{Name: "g45t345rt"},
		},
	}
}
