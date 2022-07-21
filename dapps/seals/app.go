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
	"github.com/g45t345rt/derosphere/rpc_client"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/pkg/browser"
	"github.com/urfave/cli/v2"
)

var DAPP_NAME = "seals"

var COLLECTION_SC_ID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "9fdf71cc0a563cc616e849b902f0503f5253a430c975bc4e06ee91c2814d3a8c",
	"simulator": "e43a6e0ad77917fd66ff00b685aeb6e95af7437b5f09b68d5c556e2fb54be0b7",
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

func clearData() {
	query := `
		delete from dapps_seals_collection
	`

	db := app.Context.DB

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func sync() {
	daemon := app.Context.WalletInstance.Daemon
	scid := getCollectionSCID()
	commitCount := daemon.GetSCCommitCount(scid)
	counts := utils.GetCounts()
	commitAt := counts[DAPP_NAME]

	if commitAt == 0 {
		clearData()
	}

	chunk := uint64(1000)
	db := app.Context.DB
	nftKey, _ := regexp.Compile(`state_nft_(.+)`)

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	for i := commitAt; i < commitCount; i += chunk {
		var commits []rpc_client.Commit
		end := i + chunk
		if end > commitCount {
			commitAt = commitCount
			commits = daemon.GetSCCommits(scid, i, commitCount)
		} else {
			commitAt = end
			commits = daemon.GetSCCommits(scid, i, commitAt)
		}

		for _, commit := range commits {
			key := commit.Key

			if strings.HasPrefix(key, "state_nft_") {
				assetTokenSCID := nftKey.ReplaceAllString(key, "$1")

				if commit.Action == "S" {
					result, err := daemon.GetSC(&rpc.GetSC_Params{
						SCID:      assetTokenSCID,
						Code:      true,
						Variables: true,
					})

					if err != nil {
						log.Fatal(err)
					}

					nft, err := utils.ParseG45NFT(assetTokenSCID, result)
					if err != nil {
						fmt.Println(err)
						continue
					}

					values, err := url.ParseQuery(nft.Metadata)
					if err != nil {
						log.Fatal(err)
					}

					query := `
						insert into dapps_seals_collection (token, frozen_metadata,	frozen_supply, supply, metadata, file_number,
							rarity, trait_background, trait_base, trait_eyes, trait_hairAndHats, trait_shirts, trait_tattoo)
						values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
						on conflict(token) do update 
						set token = ?
					`

					_, err = tx.Exec(query, nft.Token, nft.FrozenMetadata, nft.FrozenSupply, nft.Supply, nft.Metadata,
						utils.NewNullString(values.Get("id")), utils.NewNullString(values.Get("rarity")), values.Get("trait_background"),
						values.Get("trait_base"), values.Get("trait_eyes"), values.Get("trait_hairAndHats"),
						values.Get("trait_shirts"), values.Get("trait_tattoo"), nft.Token,
					)

					if err != nil {
						log.Fatal(err)
					}
				} else if commit.Action == "D" {
					query := `delete from dapps_seals_collection where token = ?`

					_, err = tx.Exec(query, assetTokenSCID)

					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

		utils.SetCount(DAPP_NAME, commitAt)
	}
}

func CommandList() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "List NFT collection",
		Action: func(c *cli.Context) error {
			sync()

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
			sync()

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
