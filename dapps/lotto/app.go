package lotto

import (
	"fmt"
	"log"

	"github.com/deroproject/derohe/rpc"
	"github.com/fatih/color"
	"github.com/g45t345rt/derosphere/app"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"

	_ "github.com/mattn/go-sqlite3"
)

var SC_ID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "",
	"simulator": "95b938ea2aa43a9ddc7a30db3dbcaa311f3bb99320d4b8bde62f42739d35d0b8",
}

func getSCID() string {
	return SC_ID[app.Context.Config.Env]
}

type Lotto struct {
	TxId           string
	MaxTickets     uint64
	TicketPrice    uint64
	TicketCount    uint64
	BaseReward     uint64
	Duration       uint64
	UniqueWallet   bool
	PasswordHash   string
	DrawTimestamp  uint64
	ClaimTxId      string
	ClaimTimestamp uint64
	StartTimestamp uint64
	Winner         string
	WinningTicket  uint64
	WinnerComment  string
	Owner          string
}

type LottoTicket struct {
	LottoTxId    string
	TicketNumber uint64
	Owner        string
	Timestamp    uint64
	PlayTxId     string
}

func initData() {
	sqlQuery := `
		create table if not exists lotto (
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
			owner varchar
		);

		create table if not exists lotto_tickets (
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

func displayLiveLottoTable(lottories []Lotto) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("", "Owner", "")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for index, l := range lottories {
		tbl.AddRow(index, l.Owner)
	}

	tbl.Print()
	if len(lottories) == 0 {
		fmt.Println("No lottery available")
	}
}

func CommandViewResult() *cli.Command {
	return &cli.Command{
		Name:    "result",
		Aliases: []string{"r"},
		Usage:   "View lottery draws / result",
		Action: func(c *cli.Context) error {
			fmt.Println("draw result")
			return nil
		},
	}
}

func CommandBuyTicket() *cli.Command {
	return &cli.Command{
		Name:    "buy",
		Aliases: []string{"b"},
		Usage:   "Buy a ticket",
		Action: func(c *cli.Context) error {
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

			duration, err := app.PromptInt("Duration", 0)
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

			fmt.Println(password)

			startTimestamp, err := app.PromptInt("Start timestamp (unix)", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			baseReward, err := app.PromptInt("Base reward", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			antiSpamFee := int64(100000)
			arg_sc := rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: scid}
			arg_sc_action := rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: rpc.SC_CALL}
			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Create"}
			arg2 := rpc.Argument{Name: "maxTickets", DataType: rpc.DataUint64, Value: maxTickets}
			arg3 := rpc.Argument{Name: "ticketPrice", DataType: rpc.DataUint64, Value: ticketPrice}
			arg4 := rpc.Argument{Name: "duration", DataType: rpc.DataUint64, Value: duration}
			arg5 := rpc.Argument{Name: "uniqueWallet", DataType: rpc.DataUint64, Value: uniqueWallet}
			arg6 := rpc.Argument{Name: "passwordHash", DataType: rpc.DataString, Value: ""}
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
