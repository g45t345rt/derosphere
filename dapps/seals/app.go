package seals

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"regexp"
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
	"testnet":   "84f3153f4cb0b56ee8560904a83f2859ec92c5c08aa3b6d2c3bf9cd962703fda",
	"simulator": "e9e9f5f22a0798774fa070ec2f1ba8d5b8df93282d928f6042e787aa6b4ddde4",
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
	)

	return strings.Join(traits, ", ")
}

func initData() {
	query := `
		create table if not exists dapps_seals_collection (
			token varchar primary key,
			frozen_metadata boolean,
			frozen_supply boolean,
			supply bigint,
			metadata string,
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

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func update() error {
	daemon := app.Context.WalletInstance.Daemon
	scid := getCollectionSCID()

	db := app.Context.DB
	nftKey, _ := regexp.Compile(`nft_(.+)`)

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	result, err := daemon.GetSC(&rpc.GetSC_Params{
		SCID:      scid,
		Code:      true,
		Variables: true,
	})

	if err != nil {
		return err
	}

	_, err = tx.Exec(`delete from dapps_seals_collection;`)
	if err != nil {
		return err
	}

	for key := range result.VariableStringKeys {
		if strings.HasPrefix(key, "nft_") {
			assetTokenSCID := nftKey.ReplaceAllString(key, "$1")

			nft, err := utils.GetG45NFT(assetTokenSCID, daemon)
			if err != nil {
				fmt.Printf("%s %s\n", assetTokenSCID, err.Error())
				continue
			}

			values, err := url.ParseQuery(nft.Metadata)
			if err != nil {
				fmt.Println(err)
				continue
			}

			query := `
				insert into dapps_seals_collection (token, frozen_metadata,	frozen_supply, supply, metadata, file_number,
					rarity, trait_background, trait_base, trait_eyes, trait_hairAndHats, trait_shirts, trait_tattoo)
				values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`

			_, err = tx.Exec(query, nft.Token, nft.FrozenMetadata, nft.FrozenSupply, nft.Supply, nft.Metadata,
				utils.NewNullString(values.Get("id")), utils.NewNullString(values.Get("rarity")), values.Get("trait_background"),
				values.Get("trait_base"), values.Get("trait_eyes"), values.Get("trait_hairAndHats"),
				values.Get("trait_shirts"), values.Get("trait_tattoo"), nft.Token,
			)

			if err != nil {
				return err
			}
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
			query := `select token, frozen_metadata, frozen_supply, supply, metadata, file_number, rarity,
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
			query := `select file_number from dapps_seals_collection where token = ?`

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
		Usage:   "Open NFT image with file number",
		Action: func(ctx *cli.Context) error {
			id := ctx.Args().First()
			var err error

			if id == "" {
				id, err = app.Prompt("Enter file number", "")
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
