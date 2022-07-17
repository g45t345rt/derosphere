package nft_trade

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/rpc_client"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

var DAPP_NAME = "nft-trade"

var EXCHANGE_SC_ID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "",
	"simulator": "a38eb330a23cba1ad7ebcb03cb1673a5ae4b7cce4549339cd6abf7548380ebe7",
}

func getExchangeSCID() string {
	return EXCHANGE_SC_ID[app.Context.Config.Env]
}

type Exchange struct {
	Id                sql.NullInt64
	SellAmount        sql.NullInt64
	SellAssetId       sql.NullString
	BuyAssetId        sql.NullString
	BuyAmount         sql.NullInt64
	Seller            sql.NullString
	Timestamp         sql.NullInt64
	Complete          sql.NullBool
	CompleteTimestamp sql.NullInt64
	Buyer             sql.NullString
}

func (e *Exchange) DisplayTimestamp() string {
	return time.Unix(e.Timestamp.Int64, 0).Local().String()
}

func (e *Exchange) DisplayCompleteTimestamp() string {
	if e.CompleteTimestamp.Valid {
		return time.Unix(e.CompleteTimestamp.Int64, 0).Local().String()
	}

	return ""
}

func initData() {
	sqlQuery := `
		create table if not exists dapps_nft_trade_exchanges (
			id integer primary key,
			sellAmount integer,
			sellAssetId varchar,
			buyAssetId varchar,
			buyAmount varchar,
			seller varchar,
			timestamp integer,
			complete boolean,
			completeTimestamp integer,
			buyer varchar
		);

		create table if not exists dapps_nft_trade_auctions (
			id integer primary key,
			sellAssetId varchar,
			startAmount integer,
			startTimestamp integer,
			duration integer,
			seller varchar,
			bidAssetId varchar,
			minBidAmount integer,
			bidSum integer,
			bidCount integer,
			timestamp integer
		);

		create table if not exists dapps_nft_trade_auctions_bids (
			bidder varchar,
			bidId integer,
			lockedAmount integer,
			timestamp integer,
			primary key (bidder, bidId)
		);
	`

	db := app.Context.DB

	_, err := db.Exec(sqlQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func clearData() {
	query := `
		delete from dapps_nft_trade_exchanges;
		delete from dapps_nft_trade_auctions;
		delete from dapps_nft_trade_auctions_bids;
	`

	db := app.Context.DB

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func syncAuction() {
	// TODO
}

func syncExchange() {
	daemon := app.Context.WalletInstance.Daemon
	scid := getExchangeSCID()
	commitCount := daemon.GetSCCommitCount(scid)
	counts := utils.GetCounts()
	commitAt := counts[DAPP_NAME]

	if commitAt == 0 {
		clearData()
	}

	chunk := uint64(1000)
	db := app.Context.DB
	exKey, _ := regexp.Compile(`state_ex_(\d+)_(.+)`)

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

			if strings.HasPrefix(key, "state_ex_") {
				exId := exKey.ReplaceAllString(key, "$1")
				columnName := exKey.ReplaceAllString(key, "$2")

				if commit.Action == "S" {
					query := fmt.Sprintf(`
						insert into dapps_nft_trade_exchanges (id, %s)
						values (?, ?)
						on conflict(id) do update 
						set %s = ?
					`, columnName, columnName)

					_, err = tx.Exec(query, exId, commit.Value, commit.Value)

					if err != nil {
						log.Fatal(err)
					}
				} else if commit.Action == "D" {
					query := fmt.Sprintf(`
					  update dapps_nft_trade_exchanges
						set %s = null
						where id = ?
					`, columnName)

					_, err := tx.Exec(query, exId)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		query := `delete from dapps_nft_trade_exchanges
		where sellAssetId is null`

		_, err := tx.Exec(query)
		if err != nil {
			log.Fatal(err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

		utils.SetCount(DAPP_NAME, commitAt)
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

func CommandCreateAuction() *cli.Command {
	return &cli.Command{
		Name:    "create-auction",
		Aliases: []string{"ca"},
		Usage:   "Create auction",
		Action: func(ctx *cli.Context) error {
			sellAssetId := ctx.Args().First()
			var err error

			if sellAssetId == "" {
				sellAssetId, err = app.Prompt("Enter asset id to sell", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			amount, err := app.PromptUInt("Enter asset amount", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			buyAssetId, err := app.Prompt("Enter asset id you want to auction for (empty for Dero)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			startAmount := uint64(0)
			minBidAmount := uint64(0)
			if buyAssetId == "" {
				buyAssetId = "0000000000000000000000000000000000000000000000000000000000000000"
				startAmount, err = app.PromptDero("Enter start amount (in Dero)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}

				minBidAmount, err = app.PromptDero("Enter min bid amount (in Dero)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			} else {
				startAmount, err = app.PromptUInt("Enter start amount of the asset", 1)
				if app.HandlePromptErr(err) {
					return nil
				}

				minBidAmount, err = app.PromptUInt("Enter min bid amount of the asset", 1)
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			startTimestamp, err := app.PromptUInt("Start timestamp (unix)", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			duration, err := app.PromptUInt("Duration (in seconds)", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			scid := getExchangeSCID()

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "CreateExchange", []rpc.Argument{
				{Name: "sellAssetId", DataType: rpc.DataString, Value: sellAssetId},
				{Name: "buyAssetId", DataType: rpc.DataString, Value: buyAssetId},
				{Name: "startAmount", DataType: rpc.DataUint64, Value: startAmount},
				{Name: "minBidAmount", DataType: rpc.DataUint64, Value: minBidAmount},
				{Name: "startTimestamp", DataType: rpc.DataUint64, Value: startTimestamp},
				{Name: "duration", DataType: rpc.DataUint64, Value: duration},
			}, []rpc.Transfer{
				{SCID: crypto.HashHexToHash(sellAssetId), Burn: uint64(amount), Destination: randomAddresses.Address[0]},
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

func CommandListExchange() *cli.Command {
	return &cli.Command{
		Name:    "list-exchange",
		Aliases: []string{"le"},
		Usage:   "List assets you can buy",
		Action: func(c *cli.Context) error {
			syncExchange()

			db := app.Context.DB

			query := `
				select id, sellAmount, sellAssetId, buyAssetId, buyAmount, seller, timestamp, complete, completeTimestamp, buyer
				from dapps_nft_trade_exchanges
				where complete = false
			`

			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
			}

			var exchanges []Exchange
			for rows.Next() {
				var exchange Exchange
				err = rows.Scan(&exchange.Id, &exchange.SellAmount, &exchange.SellAssetId, &exchange.BuyAssetId,
					&exchange.BuyAmount, &exchange.Seller, &exchange.Timestamp,
					&exchange.Complete, &exchange.CompleteTimestamp, &exchange.Buyer)

				if err != nil {
					log.Fatal(err)
				}

				exchanges = append(exchanges, exchange)
			}

			app.Context.DisplayTable(len(exchanges), func(i int) []interface{} {
				e := exchanges[i]
				return []interface{}{
					e.Id.Int64, e.SellAmount.Int64, e.SellAssetId.String, e.BuyAssetId.String, e.BuyAmount.Int64,
					e.Seller.String, e.DisplayTimestamp(), e.Complete.Bool, e.DisplayCompleteTimestamp(), e.Buyer.String,
				}
			}, []interface{}{"Id", "Sell Amount", "Sell Asset ID", "Buy Asset ID", "Buy Amount",
				"Seller", "Timestamp", "Complete", "Complete Timestamp", "Buyer"}, 5)
			return nil
		},
	}
}

func CommandCreateExchange() *cli.Command {
	return &cli.Command{
		Name:    "create-exchange",
		Aliases: []string{"ce"},
		Usage:   "Create exchange",
		Action: func(ctx *cli.Context) error {
			sellAssetId := ctx.Args().First()
			var err error

			if sellAssetId == "" {
				sellAssetId, err = app.Prompt("Enter asset id to sell", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			amount, err := app.PromptUInt("Enter asset amount", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			buyAssetId, err := app.Prompt("Enter asset id you want in exchange (empty for Dero)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			buyAmount := uint64(0)
			if buyAssetId == "" {
				buyAssetId = "0000000000000000000000000000000000000000000000000000000000000000"
				buyAmount, err = app.PromptDero("Enter how much you want (in Dero)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			} else {
				iAmount, err := app.PromptUInt("Enter amount of the asset", 1)
				if app.HandlePromptErr(err) {
					return nil
				}

				buyAmount = uint64(iAmount)
			}

			walletInstance := app.Context.WalletInstance
			scid := getExchangeSCID()

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "CreateExchange", []rpc.Argument{
				{Name: "sellAssetId", DataType: rpc.DataString, Value: sellAssetId},
				{Name: "buyAssetId", DataType: rpc.DataString, Value: buyAssetId},
				{Name: "buyAmount", DataType: rpc.DataUint64, Value: buyAmount},
			}, []rpc.Transfer{
				{SCID: crypto.HashHexToHash(sellAssetId), Burn: uint64(amount), Destination: randomAddresses.Address[0]},
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

func CommandCancelExchange() *cli.Command {
	return &cli.Command{
		Name:    "cancel-exchange",
		Aliases: []string{"ce"},
		Usage:   "Cancel exchange",
		Action: func(ctx *cli.Context) error {
			sExId := ctx.Args().First()
			var err error

			if sExId == "" {
				sExId, err = app.Prompt("Enter exchange id", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			walletInstance := app.Context.WalletInstance
			scid := getExchangeSCID()

			exId, err := strconv.ParseUint(sExId, 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "CancelExchange", []rpc.Argument{
				{Name: "exId", DataType: rpc.DataUint64, Value: exId},
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

func CommandBuyExchange() *cli.Command {
	return &cli.Command{
		Name:    "buy-asset",
		Aliases: []string{"ba"},
		Usage:   "Buy asset",
		Action: func(ctx *cli.Context) error {
			sExId := ctx.Args().First()
			var err error

			if sExId == "" {
				sExId, err = app.Prompt("Enter exchange id", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			exId, err := strconv.ParseUint(sExId, 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance := app.Context.WalletInstance
			scid := getExchangeSCID()

			query := `
				select id, sellAmount, sellAssetId, buyAssetId, buyAmount, seller, timestamp, complete, completeTimestamp
				from dapps_nft_trade_exchanges
				where id = ?
			`

			db := app.Context.DB
			row := db.QueryRow(query, exId)

			var transfer rpc.Transfer
			var exchange Exchange
			err = row.Scan(&exchange.Id, &exchange.SellAmount, &exchange.SellAssetId, &exchange.BuyAssetId,
				&exchange.BuyAmount, &exchange.Seller, &exchange.Timestamp, &exchange.Complete, &exchange.CompleteTimestamp)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			transfer = rpc.Transfer{
				Burn:        uint64(exchange.BuyAmount.Int64),
				Destination: randomAddresses.Address[0],
			}

			if exchange.BuyAssetId.Valid && exchange.BuyAssetId.String != "" {
				transfer.SCID = crypto.HashHexToHash(exchange.BuyAssetId.String)
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "Buy", []rpc.Argument{
				{Name: "exId", DataType: rpc.DataUint64, Value: exId},
			}, []rpc.Transfer{transfer}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txId)
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
		Description: "Browse, buy, sell and auction assets.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandListExchange(),
			CommandCreateExchange(),
			CommandCancelExchange(),
			CommandBuyExchange(),
			CommandListAuction(),
			CommandViewNFT(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
