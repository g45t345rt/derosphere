package asset_trade

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/config"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

var DAPP_NAME = "asset-trade"

var EXCHANGE_SCID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "d8301b171c1554c15a553f26dfd8754963c9201781fd46eb4c547685029afeb8",
	"simulator": "ad0be83e6443c9002eb063e9443230e250f2a2007bacb6a141e72e8eea3a1bf1",
}

var AUCTION_SCID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "a8b7153181a9da75eed78bc523d9496025768569e3ecb8dff60966e7934bcbb1",
	"simulator": "c82f7fcf1ce54c54a80e92535356814827ea3af071da39778dea64c16931d5d8",
}

func getExchangeSCID() string {
	return EXCHANGE_SCID[app.Context.Config.Env]
}

func getAuctionSCID() string {
	return AUCTION_SCID[app.Context.Config.Env]
}

type Order struct {
	Id              sql.NullInt64
	Type            sql.NullString
	AssetAmount     sql.NullInt64
	AssetBalance    sql.NullInt64
	AssetId         sql.NullString
	PriceAssetId    sql.NullString
	UnitPrice       sql.NullInt64
	Creator         sql.NullString
	Timestamp       sql.NullInt64
	Close           sql.NullBool
	PriceAmount     sql.NullInt64
	PriceBalance    sql.NullInt64
	OneTxOnly       sql.NullBool
	ExpireTimestamp sql.NullInt64
}

func (e *Order) DisplayTimestamp() string {
	return time.Unix(e.Timestamp.Int64, 0).Local().String()
}

func (e *Order) DisplayExpireTimestamp() string {
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
		create table if not exists dapps_asset_trade_orders (
			id bigint primary key,
			type varchar,
			assetAmount bigint,
			assetBalance bigint,
			assetId varchar,
			priceAssetId varchar,
			unitPrice bigint,
			creator varchar,
			timestamp bigint,
			close boolean,
			oneTxOnly boolean,
			priceAmount bigint,
			priceBalance bigint,
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
			close boolean,
			lastBidder varchar
		);

		create table if not exists dapps_asset_trade_auctions_bids (
			auId bigint,
			bidder varchar,
			lockedAmount bigint,
			timestamp bigint,
			primary key (bidder, auId)
		);

		create table if not exists dapps_asset_trade_orders_txs (
			odId bigint,
			id bigint,
			sender varchar,
			assetSent bigint,
			assetReceived bigint,
			amountSent bigint,
			amountReceived bigint,
			timestamp bigint,
			txId varchar,
			fee bigint,
			primary key (odId, id)
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
		delete from dapps_asset_trade_orders;
		delete from dapps_asset_trade_orders_txs;
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
	commitCount, err := daemon.GetSCCommitCountV2(scid)
	if err != nil {
		log.Fatal(err)
	}

	count := utils.Count{Filename: config.GetCountFilename(app.Context.Config.Env)}
	err = count.Load()
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
	auKey, _ := regexp.Compile(`au_(\d+)_(.+)`)
	bidKey, _ := regexp.Compile(`au_(\d+)_bid_(.+)_(.+)`)

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		} else {
			tx.Commit()

			count.Set(name, commitAt)
			err = count.Save()
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	for i := commitAt; i < commitCount; i += chunk {
		var commits []map[string]interface{}
		end := i + chunk
		if end > commitCount {
			commitAt = commitCount
			commits, err = daemon.GetSCCommitsV2(scid, i, commitCount)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			commitAt = end
			commits, err = daemon.GetSCCommitsV2(scid, i, commitAt)
			if err != nil {
				log.Fatal(err)
			}
		}

		for _, changes := range commits {
			for key, value := range changes {
				if bidKey.Match([]byte(key)) {
					auId := bidKey.ReplaceAllString(key, "$1")
					signer := bidKey.ReplaceAllString(key, "$2")
					columnName := bidKey.ReplaceAllString(key, "$3")

					deleted := value == -1
					if !deleted {
						query := fmt.Sprintf(`insert into dapps_asset_trade_auctions_bids (auId, bidder, %s)
						values (?, ?, ?)
						on conflict(auId, bidder) do update
						set %s = ?`, columnName, columnName)

						_, err = tx.Exec(query, auId, signer, value, value)
					}
				} else if auKey.Match([]byte(key)) {
					auId := auKey.ReplaceAllString(key, "$1")
					columnName := auKey.ReplaceAllString(key, "$2")

					deleted := value == -1
					if !deleted {
						query := fmt.Sprintf(`insert into dapps_asset_trade_auctions (id, %s)
							values (?, ?)
							on conflict(id) do update
							set %s = ?`, columnName, columnName)

						_, err = tx.Exec(query, auId, value, value)
					}
				}
			}
		}
	}
}

func syncExchange() {
	daemon := app.Context.WalletInstance.Daemon
	scid := getExchangeSCID()
	commitCount, err := daemon.GetSCCommitCountV2(scid)
	if err != nil {
		log.Fatal(err)
	}

	count := utils.Count{Filename: config.GetCountFilename(app.Context.Config.Env)}
	err = count.Load()
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
	odKey, _ := regexp.Compile(`od_(\d+)_(.+)`)
	txKey, _ := regexp.Compile(`od_(\d+)_tx_(.+)_(.+)`)

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		} else {
			tx.Commit()

			count.Set(name, commitAt)
			err = count.Save()
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	for i := commitAt; i < commitCount; i += chunk {
		var commits []map[string]interface{}
		end := i + chunk
		if end > commitCount {
			commitAt = commitCount
			commits, err = daemon.GetSCCommitsV2(scid, i, commitCount)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			commitAt = end
			commits, err = daemon.GetSCCommitsV2(scid, i, commitAt)
			if err != nil {
				log.Fatal(err)
			}
		}

		for _, changes := range commits {
			for key, value := range changes {
				if txKey.Match([]byte(key)) {
					odId := txKey.ReplaceAllString(key, "$1")
					id := txKey.ReplaceAllString(key, "$2")
					columnName := txKey.ReplaceAllString(key, "$3")

					deleted := value == -1
					if !deleted {
						_, err = tx.Exec(fmt.Sprintf(`insert into dapps_asset_trade_orders_txs (odId,id,%s) values (?,?,?) on conflict(odId,id) do update set %s = ?;`, columnName, columnName),
							odId, id, value, value)
					}
				} else if odKey.Match([]byte(key)) {
					odId := odKey.ReplaceAllString(key, "$1")
					columnName := odKey.ReplaceAllString(key, "$2")
					if columnName == "txCtr" {
						continue
					}

					deleted := value == -1
					if !deleted {
						query := fmt.Sprintf(`
							insert into dapps_asset_trade_orders (id, %s)
							values (?, ?)
							on conflict(id) do update 
							set %s = ?
						`, columnName, columnName)

						_, err = tx.Exec(query, odId, value, value)
					}
				}
			}
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

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			transfer := rpc.Transfer{
				SCID:        crypto.HashHexToHash(sellAssetId),
				Burn:        amount,
				Destination: randomAddresses.Address[0],
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

func CommandCloseAuction() *cli.Command {
	return &cli.Command{
		Name:    "close-auction",
		Aliases: []string{"ce"},
		Usage:   "close auction",
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

			txId, err := walletInstance.CallSmartContract(2, scid, "CloseAuction", []rpc.Argument{
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

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			transfer := rpc.Transfer{
				SCID:        crypto.HashHexToHash(auction.BidAssetId.String),
				Burn:        bidAmount,
				Destination: randomAddresses.Address[0],
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

func CommandListOrders() *cli.Command {
	return &cli.Command{
		Name:    "list-orders",
		Aliases: []string{"lo"},
		Usage:   "Assets/NFTs you can buy or sell",
		Action: func(c *cli.Context) error {
			syncExchange()

			db := app.Context.DB

			query := `
				select id, assetAmount, assetId, priceAssetId, unitPrice, creator, timestamp, close
				from dapps_asset_trade_orders
				where close = false
			`

			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
			}

			var orders []Order
			for rows.Next() {
				var order Order
				err = rows.Scan(&order.Id, &order.AssetAmount, &order.AssetId, &order.PriceAssetId,
					&order.UnitPrice, &order.Creator, &order.Timestamp, &order.Close)

				if err != nil {
					log.Fatal(err)
				}

				orders = append(orders, order)
			}

			app.Context.DisplayTable(len(orders), func(i int) []interface{} {
				e := orders[i]
				return []interface{}{
					e.Id.Int64, e.AssetAmount.Int64, e.AssetId.String, e.PriceAssetId.String, e.UnitPrice.Int64,
					e.Creator.String, e.DisplayTimestamp(), e.Close.Bool,
				}
			}, []interface{}{"Id", "Sell Amount", "Sell Asset ID", "Buy Asset ID", "Buy Amount",
				"Creator", "Timestamp", "Close"}, 5)
			return nil
		},
	}
}

func CommandCreateOrder() *cli.Command {
	return &cli.Command{
		Name:    "create-order",
		Aliases: []string{"co"},
		Usage:   "Create order",
		Action: func(ctx *cli.Context) error {
			odType, err := app.PromptChoose("Sell or buy?", []string{"sell", "buy"}, "sell")
			if app.HandlePromptErr(err) {
				return nil
			}

			assetId, err := app.Prompt("Enter asset id", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			assetAmount, err := app.PromptUInt("Enter amount (atomic value)", 1)
			if app.HandlePromptErr(err) {
				return nil
			}

			priceAssetId, err := app.Prompt("Enter price asset id (empty for DERO)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			unitPrice := uint64(0)
			if priceAssetId == "" {
				priceAssetId = crypto.ZEROHASH.String() //"0000000000000000000000000000000000000000000000000000000000000000"
				unitPrice, err = app.PromptDero("Enter unit price (in Dero)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			} else {
				unitPrice, err = app.PromptUInt("Enter unit price (atomic value)", 1)
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			expireTimestamp, err := app.PromptUInt("Expire timestamp (unix)", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			uOneTx := uint64(0)
			oneTx, err := app.PromptYesNo("One transaction only?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if oneTx {
				uOneTx = 1
			}

			walletInstance := app.Context.WalletInstance
			scid := getExchangeSCID()

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			transfer := rpc.Transfer{}

			switch odType {
			case "sell":
				transfer = rpc.Transfer{
					SCID:        crypto.HashHexToHash(assetId),
					Burn:        assetAmount,
					Destination: randomAddresses.Address[0],
				}
			case "buy":
				transfer = rpc.Transfer{
					SCID:        crypto.HashHexToHash(priceAssetId),
					Burn:        assetAmount * unitPrice,
					Destination: randomAddresses.Address[0],
				}
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "CreateOrder", []rpc.Argument{
				{Name: "odType", DataType: rpc.DataString, Value: odType},
				{Name: "assetId", DataType: rpc.DataString, Value: assetId},
				{Name: "priceAssetId", DataType: rpc.DataString, Value: priceAssetId},
				{Name: "unitPrice", DataType: rpc.DataUint64, Value: unitPrice},
				{Name: "expireTimestamp", DataType: rpc.DataUint64, Value: expireTimestamp},
				{Name: "oneTxOnly", DataType: rpc.DataUint64, Value: uOneTx},
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

func CommandCloseOrder() *cli.Command {
	return &cli.Command{
		Name:    "close-order",
		Aliases: []string{"ce"},
		Usage:   "Close order",
		Action: func(ctx *cli.Context) error {
			sOdId := ctx.Args().First()
			var err error

			if sOdId == "" {
				sOdId, err = app.Prompt("Enter order id", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			walletInstance := app.Context.WalletInstance
			scid := getExchangeSCID()

			odId, err := strconv.ParseUint(sOdId, 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "CloseOrder", []rpc.Argument{
				{Name: "odId", DataType: rpc.DataUint64, Value: odId},
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

func CommandBuyOrSell() *cli.Command {
	return &cli.Command{
		Name:    "buysell-order",
		Aliases: []string{"bso"},
		Usage:   "Buy or sell asset order",
		Action: func(ctx *cli.Context) error {
			syncExchange()

			sOdId := ctx.Args().First()
			var err error

			if sOdId == "" {
				sOdId, err = app.Prompt("Enter order id", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			odId, err := strconv.ParseUint(sOdId, 10, 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance := app.Context.WalletInstance
			scid := getExchangeSCID()

			query := `
				select id, type, assetAmount, assetId, priceAssetId, creator, timestamp, close, unitPrice
				from dapps_asset_trade_orders
				where id = ?
			`

			db := app.Context.DB
			row := db.QueryRow(query, odId)

			var transfer rpc.Transfer
			var order Order
			err = row.Scan(&order.Id, &order.Type, &order.AssetAmount, &order.AssetId, &order.PriceAssetId,
				&order.Creator, &order.Timestamp, &order.Close, &order.UnitPrice)

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
				Destination: randomAddresses.Address[0],
			}

			switch order.Type.String {
			case "sell":
				qty, err := app.PromptUInt("Enter quantity", 0)
				if app.HandlePromptErr(err) {
					return nil
				}

				amount := qty * uint64(order.UnitPrice.Int64)
				if order.PriceAssetId.String == crypto.ZEROHASH.String() {
					fmt.Printf("You will send %s for %d asset [%s]\n", globals.FormatMoney(amount)+" DERO", qty, order.AssetId.String)
				} else {
					fmt.Printf("You will send %d for %d asset [%s]\n", amount, qty, order.AssetId.String)
				}

				transfer.SCID = crypto.HashHexToHash(order.PriceAssetId.String)
				transfer.Burn = amount
			case "buy":
				qty, err := app.PromptUInt("Enter quantity", 0)
				if app.HandlePromptErr(err) {
					return nil
				}

				totalAmount := qty * uint64(order.UnitPrice.Int64)
				if order.PriceAssetId.String == crypto.ZEROHASH.String() {
					fmt.Printf("You will send %d asset [%s] for %s\n", qty, order.AssetId.String, globals.FormatMoney(totalAmount)+" DERO")
				} else {
					fmt.Printf("You will send %d asset [%s] for %d\n", qty, order.AssetId.String, totalAmount)
				}

				transfer.SCID = crypto.HashHexToHash(order.AssetId.String)
				transfer.Burn = qty
			}

			txId, err := walletInstance.CallSmartContract(2, scid, "BuyOrSell", []rpc.Argument{
				{Name: "odId", DataType: rpc.DataUint64, Value: odId},
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
			CommandListOrders(),
			CommandCreateOrder(),
			CommandCloseOrder(),
			CommandBuyOrSell(),
			CommandListAuction(),
			CommandCreateAuction(),
			CommandCloseAuction(),
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
