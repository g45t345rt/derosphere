package lotto

import (
	"crypto"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/rpc_client"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"

	_ "github.com/mattn/go-sqlite3"
)

var DAPP_NAME = "lotto"

var SC_ID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "",
	"simulator": "95b938ea2aa43a9ddc7a30db3dbcaa311f3bb99320d4b8bde62f42739d35d0b8",
}

func getSCID() string {
	return SC_ID[app.Context.Config.Env]
}

type Lotto struct {
	TxId           sql.NullString `db:"tx_id"`
	MaxTickets     sql.NullInt64  `db:"max_tickets"`
	TicketPrice    sql.NullInt64  `db:"ticket_price"`
	TicketCount    sql.NullInt64  `db:"ticket_count"`
	BaseReward     sql.NullInt64  `db:"base_reward"`
	Duration       sql.NullInt64  `db:"duration"`
	UniqueWallet   sql.NullBool   `db:"unique_wallet"`
	PasswordHash   sql.NullString `db:"password_hash"`
	DrawTimestamp  sql.NullInt64  `db:"draw_timestamp"`
	ClaimTxId      sql.NullString `db:"claim_tx_id"`
	ClaimTimestamp sql.NullInt64  `db:"claim_timestamp"`
	StartTimestamp sql.NullInt64  `db:"start_timestamp"`
	Winner         sql.NullString `db:"winner"`
	WinningTicket  sql.NullInt64  `db:"winning_ticket"`
	WinnerComment  sql.NullString `db:"winner_comment"`
	Owner          sql.NullString `db:"owner"`
	OwnerName      sql.NullString `db:"owner_name"`
	WinnerName     sql.NullString `db:"winner_name"`
}

func (l *Lotto) DisplayDrawTimestamp() string {
	if l.DrawTimestamp.Valid {
		return time.Unix(l.DrawTimestamp.Int64, 0).Local().String()
	}

	return ""
}

func (l *Lotto) DisplayStartTimestamp() string {
	return time.Unix(l.StartTimestamp.Int64, 0).Local().String()
}

func (l *Lotto) DisplayCreator() string {
	creator := l.Owner.String
	if l.OwnerName.Valid {
		creator = l.OwnerName.String
	}

	return creator
}

func (l *Lotto) DisplayTickets() string {
	tickets := fmt.Sprintf("%d / ∞", l.TicketCount.Int64)
	if l.MaxTickets.Int64 > 0 {
		tickets = fmt.Sprintf("%d / %d", l.TicketCount.Int64, l.MaxTickets.Int64)
	}

	return tickets
}

func (l *Lotto) DisplayWinnerReward() string {
	winnerReward := globals.FormatMoney(0)

	winnerRewardValue := l.BaseReward.Int64 + l.TicketPrice.Int64*l.MaxTickets.Int64
	if l.BaseReward.Int64 > 0 {
		winnerReward = fmt.Sprintf(">= %s", globals.FormatMoney(uint64(l.BaseReward.Int64)))
	}

	if winnerRewardValue > l.BaseReward.Int64 {
		winnerReward = globals.FormatMoney(uint64(winnerRewardValue))
	}

	return winnerReward
}

func (l *Lotto) DisplayWinner() string {
	if l.WinnerName.Valid {
		return l.WinnerName.String
	}

	return l.Winner.String
}

func (l *Lotto) Print() {
	fmt.Println("Creator:", l.DisplayCreator())
	fmt.Println("Tickets:", l.DisplayTickets())
	fmt.Println("Ticket price:", globals.FormatMoney(uint64(l.TicketPrice.Int64)))
	fmt.Println("Winner reward:", l.DisplayWinnerReward())
	fmt.Println("Base reward:", globals.FormatMoney(uint64(l.BaseReward.Int64)))
	fmt.Println("Start timestamp:", l.DisplayStartTimestamp())
	fmt.Println("Draw timestamp:", l.DisplayDrawTimestamp())
	fmt.Println("Unique wallet:", l.UniqueWallet.Bool)
	fmt.Println("Password lock:", l.PasswordHash.Valid)
}

type LottoTicket struct {
	LottoTxId    sql.NullString `db:"lotto_tx_id"`
	TicketNumber sql.NullInt64  `db:"ticker_number"`
	Owner        sql.NullString `db:"owner"`
	Timestamp    sql.NullInt64  `db:"timestamp"`
	PlayTxId     sql.NullString `db:"play_tx_id"`
	OwnerName    sql.NullString `db:"owner_name"`
}

func (lt *LottoTicket) DisplayOwner() string {
	if lt.OwnerName.Valid {
		return lt.OwnerName.String
	}

	return lt.Owner.String
}

func initData() {
	sqlQuery := `
		create table if not exists dapps_lotto (
			tx_id varchar primary key,
			max_tickets bigint,
			ticket_price bigint,
			ticket_count bigint,
			base_reward bigint,
			duration bigint,
			unique_wallet boolean,
			password_hash varchar,
			draw_timestamp bigint,
			claim_tx_id varchar,
			claim_timestamp bigint,
			start_timestamp bigint,
			winner varchar,
			winning_ticket bigint,
			winner_comment varchar,
			owner varchar,
			anti_spam_fee bigint
		);

		create table if not exists dapps_lotto_tickets (
			lotto_tx_id varchar,
			ticket_number bigint,
			owner varchar,
			timestamp bigint,
			play_tx_id varchar,
			primary key(lotto_tx_id, ticket_number)
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
	scid := getSCID()
	commitCount := daemon.GetSCCommitCount(scid)
	counts := utils.GetCounts()
	commitAt := counts[DAPP_NAME]
	chunk := uint64(1000)
	db := app.Context.DB

	isLotto, _ := regexp.Compile(`state_lotto_`)
	lottoKey, _ := regexp.Compile(`state_lotto_([a-zA-Z0-9-]+)_(.+)`)

	isTicket, _ := regexp.Compile(`state_lotto_[a-zA-Z0-9-]+_ticket_\d+`)
	ticketKey, _ := regexp.Compile(`state_lotto_([a-zA-Z0-9-]+)_ticket_(\d+)_(.+)`)

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

			if isTicket.Match([]byte(key)) {
				txId := ticketKey.ReplaceAllString(key, "$1")
				ticketNumber := ticketKey.ReplaceAllString(key, "$2")
				columnName := ticketKey.ReplaceAllString(key, "$3")

				if commit.Action == "S" {
					query := fmt.Sprintf(`
						insert into lotto_tickets (lotto_tx_id, ticket_number, %s)
						values (?, ?, ?)
						on conflict(lotto_tx_id, ticket_number) do update 
						set %s = ?
					`, columnName, columnName)

					_, err := tx.Exec(query, txId, ticketNumber, commit.Value, commit.Value)
					if err != nil {
						log.Fatal(err)
					}
				}
			} else if isLotto.Match([]byte(key)) {
				txId := lottoKey.ReplaceAllString(key, "$1")
				columnName := lottoKey.ReplaceAllString(key, "$2")

				if strings.HasPrefix(columnName, "unique_ticket_") {
					continue
				}

				if commit.Action == "S" {
					query := fmt.Sprintf(`
						insert into lotto (tx_id, %s)
						values (?, ?)
						on conflict(tx_id) do update 
						set %s = ?
					`, columnName, columnName)

					_, err := tx.Exec(query, txId, commit.Value, commit.Value)
					if err != nil {
						log.Fatal(err)
					}
				} else if commit.Action == "D" {
					query := fmt.Sprintf(`
					  update lotto
						set %s = null
						where tx_id = ?
					`, columnName)

					_, err := tx.Exec(query, txId)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		// we if ticket price null means a lotto has been cancelled and deleted from sc
		_, err := tx.Exec("delete from dapps_lotto where ticket_price is null")
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

func promptTxId(c *cli.Context) (string, error) {
	txId := c.Args().First()
	var err error

	if txId == "" {
		txId, err = app.Prompt("Enter txid", "")
		if err != nil {
			return "", err
		}
	}

	return txId, nil
}

func getLotto(db *sql.DB, txId string) (*Lotto, error) {
	query := `
		select tx_id, ticket_price, max_tickets, base_reward, ticket_count,
			unique_wallet, password_hash, draw_timestamp, claim_tx_id, claim_timestamp,
			start_timestamp, winner, winning_ticket, winner_comment, owner, u1.name as owner_name, u2.name as winner_name
		from dapps_lotto
		left join dapps_username as u1 on u1.wallet_address = owner
		left join dapps_username as u2 on u2.wallet_address = winner
		where tx_id = ?
	`

	row := db.QueryRow(query, txId)
	var lotto Lotto
	err := row.Scan(&lotto.TxId, &lotto.TicketPrice, &lotto.MaxTickets, &lotto.BaseReward,
		&lotto.TicketCount, &lotto.UniqueWallet, &lotto.PasswordHash, &lotto.DrawTimestamp, &lotto.ClaimTxId,
		&lotto.ClaimTimestamp, &lotto.StartTimestamp, &lotto.Winner, &lotto.WinningTicket, &lotto.WinnerComment,
		&lotto.Owner, &lotto.OwnerName, &lotto.WinnerName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Lotto does not exists")
		} else {
			return nil, err
		}
	}

	return &lotto, nil
}

func CommandViewResult() *cli.Command {
	return &cli.Command{
		Name:    "result",
		Aliases: []string{"r"},
		Usage:   "View lottery draws / result",
		Action: func(c *cli.Context) error {
			sync()

			db := app.Context.DB

			query := `
				select tx_id, ticket_price, max_tickets, base_reward, ticket_count,
					unique_wallet, password_hash, draw_timestamp, claim_tx_id, claim_timestamp,
					start_timestamp, winner, winning_ticket, winner_comment, owner, u1.name as owner_name, u2.name as winner_name
				from dapps_lotto
				left join dapps_username as u1 on u1.wallet_address = owner
				left join dapps_username as u2 on u2.wallet_address = winner
				where draw_timestamp is not null
			`

			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
			}

			var lottos []Lotto
			for rows.Next() {
				var lotto Lotto
				err = rows.Scan(&lotto.TxId, &lotto.TicketPrice, &lotto.MaxTickets, &lotto.BaseReward,
					&lotto.TicketCount, &lotto.UniqueWallet, &lotto.PasswordHash, &lotto.DrawTimestamp, &lotto.ClaimTxId,
					&lotto.ClaimTimestamp, &lotto.StartTimestamp, &lotto.Winner, &lotto.WinningTicket, &lotto.WinnerComment,
					&lotto.Owner, &lotto.OwnerName, &lotto.WinnerName,
				)

				if err != nil {
					log.Fatal(err)
				}

				lottos = append(lottos, lotto)
			}

			app.Context.DisplayTable(len(lottos), func(i int) []interface{} {
				l := lottos[i]
				return []interface{}{
					i, l.DisplayTickets(), l.WinningTicket.Int64, l.DisplayWinner(), l.DisplayWinnerReward(), l.ClaimTimestamp.Valid, l.DisplayDrawTimestamp(), l.TxId.String,
				}
			}, []interface{}{"", "Tickets", "Winning Ticket", "Winner", "Winner Reward", "Claimed", "Draw timestamp", "TxId"}, 5)
			return nil
		},
	}
}

func CommandBuyTicket() *cli.Command {
	return &cli.Command{
		Name:  "buy",
		Usage: "Buy ticket",
		Action: func(c *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			scid := getSCID()
			db := app.Context.DB

			txId, err := promptTxId(c)
			if app.HandlePromptErr(err) {
				return nil
			}

			lotto, err := getLotto(db, txId)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			ticketPrice := lotto.TicketPrice.Int64

			userPasswordHash := ""

			if lotto.PasswordHash.Valid && lotto.PasswordHash.String != "" {
				password, err := app.PromptPassword("Password")
				if app.HandlePromptErr(err) {
					return nil
				}

				if password != "" {
					walletAddress := walletInstance.GetAddress()
					owner := "" // TODO - get owner lotto address
					hasher := crypto.SHA3_256.New()

					// first hash
					hasher.Write([]byte(strings.Join([]string{owner, fmt.Sprintf("%d", ticketPrice), password}, ".")))
					userPasswordHash = hex.EncodeToString(hasher.Sum(nil))
					hasher.Reset()

					// second hash
					hasher.Write([]byte(strings.Join([]string{txId, userPasswordHash}, ".")))
					userPasswordHash = hex.EncodeToString(hasher.Sum(nil))
					hasher.Reset()

					// third hash
					hasher.Write([]byte(strings.Join([]string{walletAddress, userPasswordHash}, ".")))
					userPasswordHash = hex.EncodeToString(hasher.Sum(nil))
				}
			}

			arg_sc := rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: scid}
			arg_sc_action := rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: rpc.SC_CALL}
			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Play"}
			arg2 := rpc.Argument{Name: "txId", DataType: rpc.DataString, Value: txId}
			arg3 := rpc.Argument{Name: "userPasswordHash", DataType: rpc.DataString, Value: userPasswordHash}

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			txid, err := walletInstance.EstimateFeesAndTransfer(&rpc.Transfer_Params{
				Ringsize: 2,
				Transfers: []rpc.Transfer{
					{
						Burn:        uint64(ticketPrice),
						Destination: randomAddresses.Address[0],
					},
				},
				SC_RPC: rpc.Arguments{
					arg_sc, arg_sc_action, arg1, arg2, arg3,
				},
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txid)
			return nil
		},
	}
}

func CommandCreateLotto() *cli.Command {
	return &cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
		Usage:   "Create custom lottery",
		Action: func(c *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			scid := getSCID()

			maxTickets, err := app.PromptInt("Max tickets", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			ticketPrice, err := app.PromptInt("Ticket price", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			duration, err := app.PromptInt("Duration (in seconds)", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			uniqueWalletBool, err := app.PromptYesNo("Unique wallet ?", true)
			if app.HandlePromptErr(err) {
				return nil
			}

			uniqueWallet := 0
			if uniqueWalletBool {
				uniqueWallet = 1
			}

			password, err := app.PromptPassword("Password")
			if app.HandlePromptErr(err) {
				return nil
			}

			startTimestamp, err := app.PromptUInt("Start timestamp (unix)", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			baseReward, err := app.PromptInt("Base reward", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			passwordHash := ""
			if password != "" {
				walletAddress := walletInstance.GetAddress()
				hasher := crypto.SHA3_256.New()
				hasher.Write([]byte(strings.Join([]string{walletAddress, fmt.Sprintf("%d", ticketPrice), password}, ".")))
				passwordHash = hex.EncodeToString(hasher.Sum(nil))
			}

			antiSpamFee := int64(100000)
			arg_sc := rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: scid}
			arg_sc_action := rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: rpc.SC_CALL}
			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Create"}
			arg2 := rpc.Argument{Name: "maxTickets", DataType: rpc.DataUint64, Value: maxTickets}
			arg3 := rpc.Argument{Name: "ticketPrice", DataType: rpc.DataUint64, Value: ticketPrice}
			arg4 := rpc.Argument{Name: "duration", DataType: rpc.DataUint64, Value: duration}
			arg5 := rpc.Argument{Name: "uniqueWallet", DataType: rpc.DataUint64, Value: uniqueWallet}
			arg6 := rpc.Argument{Name: "passwordHash", DataType: rpc.DataString, Value: passwordHash}
			arg7 := rpc.Argument{Name: "startTimestamp", DataType: rpc.DataUint64, Value: startTimestamp}

			randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			txid, err := walletInstance.EstimateFeesAndTransfer(&rpc.Transfer_Params{
				Ringsize: 2,
				Transfers: []rpc.Transfer{
					{
						Burn:        uint64(antiSpamFee + baseReward),
						Destination: randomAddresses.Address[0],
					},
				},
				SC_RPC: rpc.Arguments{
					arg_sc, arg_sc_action, arg1, arg2, arg3, arg4, arg5, arg6, arg7,
				},
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txid)

			return nil
		},
	}
}

func CommandCancelLotto() *cli.Command {
	return &cli.Command{
		Name:    "cancel",
		Aliases: []string{"ca"},
		Usage:   "Cancel your lottery",
		Action: func(c *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			scid := getSCID()
			db := app.Context.DB

			txId, err := promptTxId(c)
			if app.HandlePromptErr(err) {
				return nil
			}

			_, err = getLotto(db, txId)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			arg_sc := rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: scid}
			arg_sc_action := rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: rpc.SC_CALL}
			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Cancel"}
			arg2 := rpc.Argument{Name: "txId", DataType: rpc.DataString, Value: txId}

			txid, err := walletInstance.EstimateFeesAndTransfer(&rpc.Transfer_Params{
				Ringsize: 2,
				SC_RPC: rpc.Arguments{
					arg_sc, arg_sc_action, arg1, arg2,
				},
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txid)
			return nil
		},
	}
}

func CommandDrawLotto() *cli.Command {
	return &cli.Command{
		Name:    "draw",
		Aliases: []string{"d"},
		Usage:   "Draw the lottery",
		Action: func(c *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			scid := getSCID()
			db := app.Context.DB

			txId, err := promptTxId(c)
			if app.HandlePromptErr(err) {
				return nil
			}

			lotto, err := getLotto(db, txId)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			if time.Now().Unix() < lotto.DrawTimestamp.Int64 {
				fmt.Println("Can't draw yet.")
				return nil
			}

			arg_sc := rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: scid}
			arg_sc_action := rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: rpc.SC_CALL}
			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Draw"}
			arg2 := rpc.Argument{Name: "txId", DataType: rpc.DataString, Value: txId}

			txid, err := walletInstance.EstimateFeesAndTransfer(&rpc.Transfer_Params{
				Ringsize: 2,
				SC_RPC: rpc.Arguments{
					arg_sc, arg_sc_action, arg1, arg2,
				},
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txid)
			return nil
		},
	}
}

func CommandClaimReward() *cli.Command {
	return &cli.Command{
		Name:    "claim",
		Aliases: []string{"cl"},
		Usage:   "Claim lotto reward",
		Action: func(c *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			scid := getSCID()
			db := app.Context.DB

			txId, err := promptTxId(c)
			if app.HandlePromptErr(err) {
				return nil
			}

			lotto, err := getLotto(db, txId)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			password := ""

			if lotto.PasswordHash.Valid && lotto.PasswordHash.String != "" {
				password, err = app.PromptPassword("Enter password")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			comment, err := app.Prompt("Enter comment (optional)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			arg_sc := rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: scid}
			arg_sc_action := rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: rpc.SC_CALL}
			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "ClaimReward"}
			arg2 := rpc.Argument{Name: "txId", DataType: rpc.DataString, Value: txId}
			arg3 := rpc.Argument{Name: "comment", DataType: rpc.DataString, Value: comment}
			arg4 := rpc.Argument{Name: "password", DataType: rpc.DataString, Value: password}

			txid, err := walletInstance.EstimateFeesAndTransfer(&rpc.Transfer_Params{
				Ringsize: 2,
				SC_RPC: rpc.Arguments{
					arg_sc, arg_sc_action, arg1, arg2, arg3, arg4,
				},
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(txid)
			return nil
		},
	}
}

func CommandViewLotto() *cli.Command {
	return &cli.Command{
		Name:    "view",
		Aliases: []string{"v"},
		Usage:   "View lottery specifications",
		Action: func(c *cli.Context) error {
			sync()

			db := app.Context.DB

			txId, err := promptTxId(c)
			if app.HandlePromptErr(err) {
				return nil
			}

			lotto, err := getLotto(db, txId)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			lotto.Print()
			return nil
		},
	}
}

func CommandLiveLotto() *cli.Command {
	return &cli.Command{
		Name:    "live",
		Aliases: []string{"l"},
		Usage:   "Available lottery that you can participate",
		Action: func(c *cli.Context) error {
			sync()

			db := app.Context.DB

			query := `
				select tx_id, ticket_price, max_tickets, base_reward, ticket_count,
					unique_wallet, password_hash, draw_timestamp, claim_tx_id, claim_timestamp,
					start_timestamp, winner, winning_ticket, winner_comment, owner, u1.name as owner_name, u2.name as winner_name
				from dapps_lotto
				left join dapps_username as u1 on u1.wallet_address = owner
				left join dapps_username as u2 on u2.wallet_address = winner
				where draw_timestamp is null
			`

			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
			}

			var lottos []Lotto
			for rows.Next() {
				var lotto Lotto
				err = rows.Scan(&lotto.TxId, &lotto.TicketPrice, &lotto.MaxTickets, &lotto.BaseReward,
					&lotto.TicketCount, &lotto.UniqueWallet, &lotto.PasswordHash, &lotto.DrawTimestamp, &lotto.ClaimTxId,
					&lotto.ClaimTimestamp, &lotto.StartTimestamp, &lotto.Winner, &lotto.WinningTicket, &lotto.WinnerComment,
					&lotto.Owner, &lotto.OwnerName, &lotto.WinnerName,
				)

				if err != nil {
					log.Fatal(err)
				}

				lottos = append(lottos, lotto)
			}

			app.Context.DisplayTable(len(lottos), func(i int) []interface{} {
				l := lottos[i]
				return []interface{}{
					i, l.DisplayCreator(), l.DisplayTickets(), globals.FormatMoney(uint64(l.TicketPrice.Int64)), l.DisplayWinnerReward(), l.TxId.String,
				}
			}, []interface{}{"", "Creator", "Tickets", "Ticket Price", "Winner Reward", "TxId"}, 10)
			return nil
		},
	}
}

func CommandLottoTickets() *cli.Command {
	return &cli.Command{
		Name:    "tickets",
		Aliases: []string{"t"},
		Usage:   "View lotto tickets",
		Action: func(c *cli.Context) error {
			sync()

			db := app.Context.DB

			txId, err := promptTxId(c)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			query := `
				select lotto_tx_id, ticket_number, owner, timestamp, play_tx_id,
				u1.name as owner_name
				from dapps_lotto_tickets
				left join dapps_username as u1 on u1.wallet_address = owner
				where lotto_tx_id = ?
			`

			rows, err := db.Query(query, txId)
			if err != nil {
				log.Fatal(err)
			}

			var tickets []LottoTicket
			for rows.Next() {
				var ticket LottoTicket
				err = rows.Scan(&ticket.LottoTxId, &ticket.TicketNumber, &ticket.Owner,
					&ticket.Timestamp, &ticket.PlayTxId, &ticket.OwnerName,
				)

				if err != nil {
					log.Fatal(err)
				}

				tickets = append(tickets, ticket)
			}

			app.Context.DisplayTable(len(tickets), func(i int) []interface{} {
				t := tickets[i]
				return []interface{}{
					i, t.DisplayOwner(), t.TicketNumber.Int64, t.Timestamp.Int64, t.PlayTxId.String,
				}
			}, []interface{}{"", "Owner", "Ticket Number", "Timestamp", "TxId"}, 5)
			return nil
		},
	}
}

func App() *cli.App {
	initData()

	return &cli.App{
		Name:        "lotto",
		Description: "Official custom lottery pool. Create your own type of lottery.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandLiveLotto(),
			CommandViewLotto(),
			CommandLottoTickets(),
			CommandBuyTicket(),
			CommandCreateLotto(),
			CommandCancelLotto(),
			CommandDrawLotto(),
			CommandClaimReward(),
			CommandViewResult(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
