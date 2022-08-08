package asset_trade

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
	"github.com/g45t345rt/derosphere/config"
	"github.com/g45t345rt/derosphere/rpc_client"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

var DAPP_NAME = "asset-trade"

var EXCHANGE_SCID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "d8301b171c1554c15a553f26dfd8754963c9201781fd46eb4c547685029afeb8",
	"simulator": "caabd45f02847409f585ada62ee5ad5906f0a3b8f18108be794e9479fc7620f9",
}

var AUCTION_SCID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "a8b7153181a9da75eed78bc523d9496025768569e3ecb8dff60966e7934bcbb1",
	"simulator": "9185fc87f5b48e2a1f26c597a73018cfb405da5ce4a5a32d805249746c762898",
}

func getExchangeSCID() string {
	return EXCHANGE_SCID[app.Context.Config.Env]
}

func getAuctionSCID() string {
	return AUCTION_SCID[app.Context.Config.Env]
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
	ExpireTimestamp   sql.NullInt64
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

func (e *Exchange) DisplayExpireTimestamp() string {
	if e.ExpireTimestamp.Valid {
		return time.Unix(e.ExpireTimestamp.Int64, 0).Local().String()
	}

	return ""
}

type Auction struct {
	Id             sql.NullInt64
	StartAmount    sql.NullInt64
	SellAssetId    sql.NullString
	SellAmount     sql.NullInt64
	StartTimestamp sql.NullInt64
	Duration       sql.NullInt64
	Seller         sql.NullString
	BidAssetId     sql.NullString
	MinBidAmount   sql.NullInt64
	BidSum         sql.NullInt64
	BidCount       sql.NullInt64
	Timestamp      sql.NullInt64
	Complete       sql.NullBool
	LastBidder     sql.NullString
}

func (a *Auction) DisplayStartTimestamp() string {
	if a.StartTimestamp.Valid {
		return time.Unix(a.StartTimestamp.Int64, 0).Local().String()
	}

	return ""
}

func (a *Auction) DisplayTimestamp() string {
	if a.Timestamp.Valid {
		return time.Unix(a.Timestamp.Int64, 0).Local().String()
	}

	return ""
}

type Bid struct {
	AuId         sql.NullInt64
	Bidder       sql.NullString
	LockedAmount sql.NullInt64
	Timestamp    sql.NullString
}

func initData() {
	sqlQuery := `
		create table if not exists dapps_asset_trade_exchanges (
			id bigint primary key,
			sellAmount bigint,
			sellAssetId varchar,
			buyAssetId varchar,
			buyAmount varchar,
			seller varchar,
			timestamp bigint,
			complete boolean,
			completeTimestamp bigint,
			buyer varchar,
			expireTimestamp bigint
		);

		create table if not exists dapps_asset_trade_auctions (
			id bigint primary key,
			sellAssetId varchar,
			sellAmount bigint,
			startAmount bigint,
			startTimestamp bigint,
			duration bigint,
			seller varchar,
			bidAssetId varchar,
			minBidAmount bigint,
			bidSum bigint,
			bidCount bigint,
			timestamp bigint,
			complete boolean,
			lastBidder varchar
		);

		create table if not exists dapps_asset_trade_auctions_bids (
			auId bigint,
			bidder varchar,
			lockedAmount bigint,
			timestamp bigint,
			primary key (bidder, auId)
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
		delete from dapps_asset_trade_exchanges;
		delete from dapps_asset_trade_auctions;
		delete from dapps_asset_trade_auctions_bids;
	`

	db := app.Context.DB

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func syncAuction() {
	daemon := app.Context.WalletInstance.Daemon
	scid := getAuctionSCID()
	commitCount := daemon.GetSCCommitCount(scid)
	count := utils.Count{Filename: config.GetCountFilename(app.Context.Config.Env)}
	err := count.Load()
	if err != nil {
		log.Fatal(err)
	}

	name := DAPP_NAME + "-auction"
	commitAt := count.Get(name)

	if commitAt == 0 {
		clearData()
	}

	chunk := uint64(1000)
	db := app.Context.DB
	auKey, _ := regexp.Compile(`state_au_(\d+)_(.+)`)
	bidKey, _ := regexp.Compile(`state_au_(\d+)_bid_(.+)_(.+)`)

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

			if bidKey.Match([]byte(key)) {
				auId := bidKey.ReplaceAllString(key, "$1")
				signer := bidKey.ReplaceAllString(key, "$2")
				columnName := bidKey.ReplaceAllString(key, "$3")

				switch commit.Action {
				case "S":
					query := fmt.Sprintf(`insert into dapps_asset_trade_auctions_bids (auId, bidder, %s)
					values (?, ?, ?)
					on conflict(auId, bidder) do update
					set %s = ?`, columnName, columnName)

					_, err = tx.Exec(query, auId, signer, commit.Value, commit.Value)

					if err != nil {
						log.Fatal(err)
					}
				}
			} else if auKey.Match([]byte(key)) {
				auId := auKey.ReplaceAllString(key, "$1")
				columnName := auKey.ReplaceAllString(key, "$2")

				switch commit.Action {
				case "S":
					query := fmt.Sprintf(`insert into dapps_asset_trade_auctions (id, %s)
						values (?, ?)
						on conflict(id) do update
						set %s = ?`, columnName, columnName)

					_, err = tx.Exec(query, auId, commit.Value, commit.Value)

					if err != nil {
						log.Fatal(err)
					}
				case "D":
					query := fmt.Sprintf(`
					update dapps_asset_trade_auctions
					set %s = null
					where id = ?
				`, columnName)

					_, err := tx.Exec(query, auId)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		query := `delete from dapps_asset_trade_auctions
		where sellAssetId is null`

		_, err := tx.Exec(query)
		if err != nil {
			log.Fatal(err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

		count.Set(name, commitAt)
		err = count.Save()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func syncExchange() {
	daemon := app.Context.WalletInstance.Daemon
	scid := getExchangeSCID()
	commitCount := daemon.GetSCCommitCount(scid)
	count := utils.Count{Filename: config.GetCountFilename(app.Context.Config.Env)}
	err := count.Load()
	if err != nil {
		log.Fatal(err)
	}

	name := DAPP_NAME + "-exchange"
	commitAt := count.Get(name)

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

				switch commit.Action {
				case "S":
					query := fmt.Sprintf(`
					insert into dapps_asset_trade_exchanges (id, %s)
					values (?, ?)
					on conflict(id) do update 
					set %s = ?
				`, columnName, columnName)

					_, err = tx.Exec(query, exId, commit.Value, commit.Value)

					if err != nil {
						log.Fatal(err)
					}
				case "D":
					query := fmt.Sprintf(`
					update dapps_asset_trade_exchanges
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

		query := `delete from dapps_asset_trade_exchanges
		where sellAssetId is null`

		_, err := tx.Exec(query)
		if err != nil {
			log.Fatal(err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

		count.Set(name, commitAt)
		err = count.Save()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func CommandListAuction() *cli.Command {
	return &cli.Command{
		Name:    "list-auction",
		Aliases: []string{"la"},
		Usage:   "Assets in auction",
		Action: func(c *cli.Context) error {
			syncAuction()

			db := app.Context.DB

			query := `
				select id, startAmount, sellAssetId, sellAmount, startTimestamp, duration, seller, bidAssetId, minBidAmount, bidSum, bidCount, timestamp
				from dapps_asset_trade_auctions
			`

			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
			}

			var auctions []Auction
			for rows.Next() {
				var auction Auction
				err = rows.Scan(&auction.Id, &auction.StartAmount, &auction.SellAssetId, &auction.SellAmount, &auction.StartTimestamp,
					&auction.Duration, &auction.Seller, &auction.BidAssetId, &auction.MinBidAmount,
					&auction.BidSum, &auction.BidCount, &auction.Timestamp)

				if err != nil {
					log.Fatal(err)
				}

				auctions = append(auctions, auction)
			}

			app.Context.DisplayTable(len(auctions), func(i int) []interface{} {
				a := auctions[i]
				return []interface{}{
					a.Id.Int64, a.StartAmount.Int64, a.SellAssetId.String, a.DisplayStartTimestamp(), a.Duration.Int64,
					a.Seller.String, a.BidAssetId.String, a.MinBidAmount.Int64, a.BidSum.Int64, a.BidCount.Int64, a.DisplayTimestamp(),
				}
			}, []interface{}{"Id", "Start Amount", "Sell Asset ID", "Start Timestamp", "Duration",
				"Seller", "Bid Asset ID", "Min Bid Amount", "Bid Sum", "Bid count", "Timestamp"}, 5)
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

			bidAssetId, err := app.Prompt("Enter asset id you want to auction for (empty for DERO)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			startAmount := uint64(0)
			minBidAmount := uint64(0)
			if bidAssetId == "" {
				bidAssetId = crypto.ZEROHASH.String() //"0000000000000000000000000000000000000000000000000000000000000000"
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
			scid := getAuctionSCID()

			/*randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}*/

			transfer := rpc.Transfer{
				SCID: crypto.HashHexToHash(sellAssetId),
				Burn: amount,
				//Destination: randomAddresses.Address[0],
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "CreateAuction", []rpc.Argument{
				{Name: "sellAssetId", DataType: rpc.DataString, Value: sellAssetId},
				{Name: "bidAssetId", DataType: rpc.DataString, Value: bidAssetId},
				{Name: "startAmount", DataType: rpc.DataUint64, Value: startAmount},
				{Name: "minBidAmount", DataType: rpc.DataUint64, Value: minBidAmount},
				{Name: "startTimestamp", DataType: rpc.DataUint64, Value: startTimestamp},
				{Name: "duration", DataType: rpc.DataUint64, Value: duration},
			}, []rpc.Transfer{
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

func CommandCancelAuction() *cli.Command {
	return &cli.Command{
		Name:    "cancel-auction",
		Aliases: []string{"ce"},
		Usage:   "Cancel auction",
		Action: func(ctx *cli.Context) error {
			sAuId := ctx.Args().First()
			var err error

			if sAuId == "" {
				sAuId, err = app.Prompt("Enter auction id", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			walletInstance := app.Context.WalletInstance
			scid := getAuctionSCID()

			auId, err := strconv.ParseUint(sAuId, 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "CancelAuction", []rpc.Argument{
				{Name: "auId", DataType: rpc.DataUint64, Value: auId},
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

func CommandBidAuction() *cli.Command {
	return &cli.Command{
		Name:    "bid",
		Aliases: []string{"bi"},
		Usage:   "Bid on the auction",
		Action: func(ctx *cli.Context) error {
			syncAuction()
			sAuId := ctx.Args().First()
			walletInstance := app.Context.WalletInstance
			db := app.Context.DB

			var err error

			if sAuId == "" {
				sAuId, err = app.Prompt("Enter auction id", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			auId, err := strconv.ParseUint(sAuId, 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			scid := getAuctionSCID()

			/*randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}*/

			query := `
				select id, startAmount, sellAssetId, sellAmount, startTimestamp, duration, seller, bidAssetId, minBidAmount, bidSum, bidCount, timestamp
				from dapps_asset_trade_auctions
				where id = ?
			`

			row := db.QueryRow(query, auId)

			var auction Auction
			err = row.Scan(&auction.Id, &auction.StartAmount, &auction.SellAssetId, &auction.SellAmount,
				&auction.StartTimestamp, &auction.Duration, &auction.Seller, &auction.BidAssetId, &auction.MinBidAmount,
				&auction.BidSum, &auction.BidCount, &auction.Timestamp,
			)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			bidAmount := uint64(0)
			if auction.BidAssetId.String == crypto.ZEROHASH.String() {
				bidAmount, err = app.PromptDero("Bid amount (in Dero)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			} else {
				bidAmount, err = app.PromptUInt("Bid amount", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			transfer := rpc.Transfer{
				SCID: crypto.HashHexToHash(auction.BidAssetId.String),
				Burn: bidAmount,
				//Destination: randomAddresses.Address[0],
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "Bid", []rpc.Argument{
				{Name: "auId", DataType: rpc.DataUint64, Value: auId},
			}, []rpc.Transfer{
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

func CommandListAuctionBids() *cli.Command {
	return &cli.Command{
		Name:    "list-auction-bids",
		Aliases: []string{"lab"},
		Usage:   "List auction bids",
		Action: func(ctx *cli.Context) error {
			syncAuction()
			sAuId := ctx.Args().First()
			db := app.Context.DB

			var err error

			if sAuId == "" {
				sAuId, err = app.Prompt("Enter auction id", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			auId, err := strconv.ParseUint(sAuId, 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			query := `
				select auId, bidder, lockedAmount, timestamp
				from dapps_asset_trade_auctions_bids
				where auId = ?
			`

			rows, err := db.Query(query, auId)
			if err != nil {
				log.Fatal(err)
			}

			var bids []Bid
			for rows.Next() {
				var bid Bid
				err = rows.Scan(&bid.AuId, &bid.Bidder, &bid.LockedAmount, &bid.Timestamp)

				if err != nil {
					log.Fatal(err)
				}

				bids = append(bids, bid)
			}

			app.Context.DisplayTable(len(bids), func(i int) []interface{} {
				b := bids[i]
				return []interface{}{
					b.Bidder.String, b.LockedAmount.Int64, b.Timestamp.String,
				}
			}, []interface{}{"Bidder", "Locked Ammount", "Timestamp"}, 5)
			return nil
		},
	}
}

func CommandSetMinBidAuction() *cli.Command {
	return &cli.Command{
		Name:    "set-minbid",
		Aliases: []string{"sm"},
		Usage:   "Set auction minimum bid",
		Action: func(ctx *cli.Context) error {
			sAuId := ctx.Args().First()
			walletInstance := app.Context.WalletInstance

			var err error

			if sAuId == "" {
				sAuId, err = app.Prompt("Enter auction id", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			auId, err := strconv.ParseUint(sAuId, 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			amount, err := app.PromptUInt("Enter minimum bid amount", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			scid := getAuctionSCID()
			txId, err := walletInstance.CallSmartContract(2, scid, "SetAuctionMinBid", []rpc.Argument{
				{Name: "auId", DataType: rpc.DataUint64, Value: auId},
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

func CommandCheckoutAuction() *cli.Command {
	return &cli.Command{
		Name:    "checkout-auction",
		Aliases: []string{"ca"},
		Usage:   "Checkout finished auction",
		Action: func(ctx *cli.Context) error {
			sAuId := ctx.Args().First()
			walletInstance := app.Context.WalletInstance

			var err error

			if sAuId == "" {
				sAuId, err = app.Prompt("Enter auction id", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			auId, err := strconv.ParseUint(sAuId, 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			scid := getAuctionSCID()
			txId, err := walletInstance.CallSmartContract(2, scid, "CheckoutAuction", []rpc.Argument{
				{Name: "auId", DataType: rpc.DataUint64, Value: auId},
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

func CommandRetrieveLockedFundsAuction() *cli.Command {
	return &cli.Command{
		Name:    "get-lockedfunds",
		Aliases: []string{"gl"},
		Usage:   "Checkout finished auction",
		Action: func(ctx *cli.Context) error {
			sAuId := ctx.Args().First()
			walletInstance := app.Context.WalletInstance

			var err error

			if sAuId == "" {
				sAuId, err = app.Prompt("Enter auction id", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			auId, err := strconv.ParseUint(sAuId, 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			scid := getAuctionSCID()
			txId, err := walletInstance.CallSmartContract(2, scid, "RetrieveLockedFunds", []rpc.Argument{
				{Name: "auId", DataType: rpc.DataUint64, Value: auId},
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

func CommandListExchange() *cli.Command {
	return &cli.Command{
		Name:    "list-exchange",
		Aliases: []string{"le"},
		Usage:   "Assets/NFTs you can buy",
		Action: func(c *cli.Context) error {
			syncExchange()

			db := app.Context.DB

			query := `
				select id, sellAmount, sellAssetId, buyAssetId, buyAmount, seller, timestamp, complete, completeTimestamp, buyer
				from dapps_asset_trade_exchanges
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
			sellAssetId, err := app.Prompt("Enter asset id to sell (empty for DERO)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			sellAmount := uint64(0)
			if sellAssetId == "" {
				sellAssetId = crypto.ZEROHASH.String() //"0000000000000000000000000000000000000000000000000000000000000000"
				sellAmount, err = app.PromptDero("Enter how much you want (in Dero)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			} else {
				iAmount, err := app.PromptUInt("Enter amount (atomic value)", 1)
				if app.HandlePromptErr(err) {
					return nil
				}

				sellAmount = uint64(iAmount)
			}

			buyAssetId, err := app.Prompt("Enter asset id you want in exchange (empty for DERO)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			buyAmount := uint64(0)
			if buyAssetId == "" {
				buyAssetId = crypto.ZEROHASH.String() //"0000000000000000000000000000000000000000000000000000000000000000"
				buyAmount, err = app.PromptDero("Enter how much you want (in Dero)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			} else {
				iAmount, err := app.PromptUInt("Enter amount (atomic value)", 1)
				if app.HandlePromptErr(err) {
					return nil
				}

				buyAmount = uint64(iAmount)
			}

			expireTimestamp, err := app.PromptUInt("Expire timestamp (unix)", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance
			scid := getExchangeSCID()

			/*randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}*/

			transfer := rpc.Transfer{
				SCID: crypto.HashHexToHash(sellAssetId),
				Burn: sellAmount,
				//Destination: randomAddresses.Address[0],
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "CreateExchange", []rpc.Argument{
				{Name: "sellAssetId", DataType: rpc.DataString, Value: sellAssetId},
				{Name: "buyAssetId", DataType: rpc.DataString, Value: buyAssetId},
				{Name: "buyAmount", DataType: rpc.DataUint64, Value: buyAmount},
				{Name: "expireTimestamp", DataType: rpc.DataUint64, Value: expireTimestamp},
			}, []rpc.Transfer{
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

			walletInstance.RunTxChecker(txId)
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
			syncExchange()

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
				from dapps_asset_trade_exchanges
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

			/*randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}*/

			transfer = rpc.Transfer{
				Burn: uint64(exchange.BuyAmount.Int64),
				//Destination: randomAddresses.Address[0],
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

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func CommandViewAsset() *cli.Command {
	return &cli.Command{
		Name:    "view",
		Aliases: []string{"v"},
		Usage:   "View asset",
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

			asset, err := utils.GetG45_AT(scid, walletInstance.Daemon)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			asset.Print()
			return nil
		},
	}
}

func App() *cli.App {
	initData()

	return &cli.App{
		Name:        "asset-trade",
		Description: "Browse, buy, sell and auction assets.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandListExchange(),
			CommandCreateExchange(),
			CommandCancelExchange(),
			CommandBuyExchange(),
			CommandListAuction(),
			CommandCreateAuction(),
			CommandCancelAuction(),
			CommandSetMinBidAuction(),
			CommandBidAuction(),
			CommandListAuctionBids(),
			CommandCheckoutAuction(),
			CommandRetrieveLockedFundsAuction(),
			CommandViewAsset(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
