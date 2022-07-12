package nft_trade

import (
	"fmt"
	"log"

	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

var DAPP_NAME = "nft-trade"

var EXCHANGE_SC_ID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "",
	"simulator": "e5d487596821e5540f607d1b16fcbe8fa53a7eafe737e3c7913ea28cf652dda4",
}

func getExchangeSCID() string {
	return EXCHANGE_SC_ID[app.Context.Config.Env]
}

type Exchange struct {
	Id         uint64
	Amount     uint64
	AssetId    string
	ForAssetId string
	ForAmount  uint64
	Seller     string
}

func initData() {
	sqlQuery := `
		create table if not exists dapps_nft_trade_exchanges (
			id integer primary key,
			amount integer,
			asset_id varchar,
			for_asset_id varchar,
			for_amount varchar,
			seller varchar
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
	scid := getExchangeSCID()
	itemCount := daemon.GetSCItemCount(scid, "item_count")
	itemAt := uint64(0)
	chunk := uint64(1000)
	db := app.Context.DB

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	columns := []string{"_amount", "_assetId", "_forAssetId", "_forAmount", "_seller"}
	for i := itemAt; i < itemCount; i += chunk {
		var keyValues map[string]string
		end := i + chunk
		if end > itemCount {
			itemAt = itemCount
			keyValues = daemon.GetSCKeyValues(scid, "item_", i, itemCount, columns)
		} else {
			itemAt = end
			keyValues = daemon.GetSCKeyValues(scid, "item_", i, itemAt, columns)
		}

		if err != nil {
			log.Fatal(err)
		}

		for a := i; i < itemCount; a++ {
			amount := keyValues[fmt.Sprintf("item_%d_amount", i)]
			assetId := keyValues[fmt.Sprintf("item_%d_assetId", i)]
			forAmount := keyValues[fmt.Sprintf("item_%d_forAmount", i)]
			forAssetId := keyValues[fmt.Sprintf("item_%d_forAssetId", i)]
			seller := keyValues[fmt.Sprintf("item_%d_seller", i)]

			query := `
				insert into dapps_nft_trade_exchanges (id, amount, asset_id, for_asset_id, for_amount, seller)
				values (?, ?, ?, ?, ?, ?)
				on conflict(id) do update 
				set id = ?
			`

			_, err = tx.Exec(query, i, amount, assetId, forAssetId, forAmount, seller)

			if err != nil {
				log.Fatal(err)
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

		utils.SetCount(DAPP_NAME, itemAt)
	}
}

func CommandListAuction() *cli.Command {
	return &cli.Command{
		Name:    "auction",
		Aliases: []string{"au"},
		Usage:   "NFT in auction",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}

func CommandListExchange() *cli.Command {
	return &cli.Command{
		Name:    "exchange",
		Aliases: []string{"ex"},
		Usage:   "List NFTs you can buy",
		Action: func(c *cli.Context) error {
			sync()

			db := app.Context.DB

			query := `
				select asset_id, amount, for_asset_id, for_amount, seller
				from dapps_nft_trade_exchanges
				left join dapps_seals_collection as c1 on c1.nft = asset_id
				left join dapps_username as u on u.wallet_address = seller
			`

			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
			}

			var exchanges []Exchange
			for rows.Next() {
				var exchange Exchange
				err = rows.Scan(&exchange.AssetId, &exchange.Amount, &exchange.ForAssetId, &exchange.ForAmount, &exchange.Seller)
				if err != nil {
					log.Fatal(err)
				}

				exchanges = append(exchanges, exchange)
			}

			app.Context.DisplayTable(len(exchanges), func(i int) []interface{} {
				e := exchanges[i]

				nft := fmt.Sprintf("%d %s", e.Amount, e.AssetId)

				price := ""
				if e.ForAssetId == "" {
					price = globals.FormatMoney(e.ForAmount) + " DERO"
				} else {
					price = fmt.Sprintf("%d %s", e.ForAmount, e.ForAssetId)
				}

				return []interface{}{
					i, nft, price, e.Seller,
				}
			}, []interface{}{"", "NFT", "Price", "Seller"}, 25)

			return nil
		},
	}
}

func CommandSellExchange() *cli.Command {
	return &cli.Command{
		Name:    "sell",
		Aliases: []string{"s"},
		Usage:   "Sell NFT on exchange",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandCancelSellExchange() *cli.Command {
	return &cli.Command{
		Name:    "cancel-sell",
		Aliases: []string{"cs"},
		Usage:   "Cancel sell exchange",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandBuyExchange() *cli.Command {
	return &cli.Command{
		Name:    "buy",
		Aliases: []string{"b"},
		Usage:   "Buy exchange",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func CommandViewNFT() *cli.Command {
	return &cli.Command{
		Name:    "view",
		Aliases: []string{"v"},
		Usage:   "View specific NFT",
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
				Variables: true,
				SCID:      scid,
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			nft, err := utils.ParseG45NFT(scid, result)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			nft.Print()

			return nil
		},
	}
}

func App() *cli.App {
	initData()

	return &cli.App{
		Name:        "nft-trade",
		Description: "Browse, buy, sell and auction NFTs.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandListExchange(),
			CommandSellExchange(),
			CommandCancelSellExchange(),
			CommandBuyExchange(),
			CommandListAuction(),
			CommandViewNFT(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
