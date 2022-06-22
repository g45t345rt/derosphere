package cli

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/blang/semver/v4"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/fatih/color"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/dapps"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

func displayDAppsTable() {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("", "Name", "Description", "Version", "Authors")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for index, dapp := range dapps.List() {
		tbl.AddRow(index, dapp.Name, dapp.Description, dapp.Version, utils.AppAuthors(dapp))
	}

	tbl.Print()
}

func displayWalletTransactions(entries []rpc.Entry) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("", "Amount", "Burn", "Fees", "Time", "Height", "Destination", "Coinbase", "TXID")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for index, tx := range entries {
		tbl.AddRow(
			index,
			globals.FormatMoney(tx.Amount),
			globals.FormatMoney(tx.Burn),
			globals.FormatMoney(tx.Fees),
			tx.Time,
			tx.Height,
			tx.Destination,
			tx.Coinbase,
			tx.TXID)
	}

	tbl.Print()
}

func CommandWalletInfo() *cli.Command {
	return &cli.Command{
		Name:    "info",
		Aliases: []string{"i"},
		Usage:   "Wallet generic information",
		Action: func(ctx *cli.Context) error {
			w := app.Context.WalletInstance
			fmt.Println("Name: ", w.Name)
			fmt.Println("Daemon: ", w.DaemonAddress)
			fmt.Println("Wallet: ", w.GetConnectionAddress())
			return nil
		},
	}
}

func CommandWalletSeed() *cli.Command {
	return &cli.Command{
		Name:  "seed",
		Usage: "Display wallet seed",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			fmt.Println(walletInstance.GetSeed())
			return nil
		},
	}
}

func CommandSwitchWallet() *cli.Command {
	return &cli.Command{
		Name:    "switch",
		Usage:   "Quickly change to another wallet",
		Aliases: []string{"s"},
		Action: func(ctx *cli.Context) error {
			return OpenWalletAction(ctx, "")
		},
	}
}

func CommandListDApps() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "Show a list of available apps",
		Action: func(ctx *cli.Context) error {
			displayDAppsTable()
			return nil
		},
	}
}

func CommandDAppInfo() *cli.Command {
	return &cli.Command{
		Name:    "info",
		Aliases: []string{"i"},
		Usage:   "A basic description of the application",
		Action: func(ctx *cli.Context) error {
			app := ctx.App
			fmt.Printf("Name: %s\n", app.Name)
			fmt.Printf("Description: %s\n", app.Description)
			fmt.Printf("Authors: %s\n", utils.AppAuthors(app))
			return nil
		},
	}
}

func CommandOpenDApp() *cli.Command {
	return &cli.Command{
		Name:    "open",
		Aliases: []string{"o"},
		Usage:   "Open speficic app",
		Action: func(ctx *cli.Context) error {
			dappName := ctx.Args().First()
			var err error

		setAppName:
			if dappName == "" {
				dappName, err = app.Prompt("Enter app name/index", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			appIndex, err := strconv.ParseInt(dappName, 10, 64)
			var dapp *cli.App
			if err == nil {
				list := dapps.List()
				if int(appIndex) < len(list) {
					dapp = dapps.List()[appIndex]
				}
			} else {
				dapp = dapps.Find(dappName)
			}

			if dapp == nil {
				fmt.Println("App does not exists.")
				dappName = ""
				goto setAppName
			}

			app.Context.DAppApp = DAppApp(dapp)
			app.Context.UseApp = "dappApp"

			return nil
		},
	}
}

func CommandDAppBack() *cli.Command {
	return &cli.Command{
		Name:    "back",
		Usage:   "Back to wallet",
		Aliases: []string{"b"},
		Action: func(ctx *cli.Context) error {
			app.Context.DAppApp = nil
			app.Context.UseApp = "walletApp"
			return nil
		},
	}
}

func CommandDApp() *cli.Command {
	return &cli.Command{
		Name:               "app",
		Usage:              "App commands",
		CustomHelpTemplate: utils.AppTemplate,
		Subcommands: []*cli.Command{
			CommandListDApps(),
			CommandOpenDApp(),
		},
		Action: func(ctx *cli.Context) error {
			ctx.App.Run([]string{"cmd", "help"})
			return nil
		},
	}
}

func CommandCloseWallet() *cli.Command {
	return &cli.Command{
		Name:    "close",
		Aliases: []string{"c"},
		Usage:   "Close wallet",
		Action: func(ctx *cli.Context) error {
			app.Context.WalletInstance.Close()
			app.Context.WalletInstance = nil
			app.Context.UseApp = "rootApp"
			return nil
		},
	}
}

func CommandWalletAddress() *cli.Command {
	return &cli.Command{
		Name:    "address",
		Aliases: []string{"a"},
		Usage:   "Wallet address",
		Action: func(ctx *cli.Context) error {
			fmt.Println(app.Context.WalletInstance.GetAddress())
			return nil
		},
	}
}

func CommandWalletBalance() *cli.Command {
	return &cli.Command{
		Name:    "balance",
		Aliases: []string{"b"},
		Usage:   "Wallet balance",
		Action: func(ctx *cli.Context) error {
			balance := app.Context.WalletInstance.GetBalance()
			fmt.Printf("%s\n", globals.FormatMoney(balance))
			return nil
		},
	}
}

func CommandWalletTransferDero() *cli.Command {
	return &cli.Command{
		Name:    "transfer",
		Aliases: []string{"t"},
		Usage:   "Transfer DERO to another address",
		Action: func(ctx *cli.Context) error {

			walletInstance := app.Context.WalletInstance

			assetToken, err := app.Prompt("Enter asset token (empty for sending DERO)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			addressOrName, err := app.Prompt("Enter address/name", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			amount, err := app.PromptInt("Enter amount", 0)
			if app.HandlePromptErr(err) {
				return nil
			}

			ringsize, err := app.PromptInt("Set ringsize", 2)
			if app.HandlePromptErr(err) {
				return nil
			}

			transfer := rpc.Transfer{SCID: crypto.HashHexToHash(assetToken), Destination: addressOrName, Amount: uint64(amount)}

			yes, err := app.PromptYesNo(fmt.Sprintf("Are you sure you want to send %s to %s", globals.FormatMoney(transfer.Amount), addressOrName), false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if !yes {
				return nil
			}

			txid, err := walletInstance.Transfer(&rpc.Transfer_Params{
				Ringsize: uint64(ringsize),
				Transfers: []rpc.Transfer{
					transfer,
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

func CommandUpdateSC() *cli.Command {
	return &cli.Command{
		Name:    "update",
		Aliases: []string{"u"},
		Usage:   "Update smart contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			codeFilePath, err := app.Prompt("Enter new code filepath", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			code, err := ioutil.ReadFile(codeFilePath)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			codeString := string(code)
			ringsize := uint64(2)
			sc_rpc := rpc.Arguments{
				{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: rpc.SC_CALL},
				{Name: rpc.SCID, DataType: rpc.DataHash, Value: scid},
				{Name: "entrypoint", DataType: rpc.DataString, Value: "UpdateCode"},
				{Name: "code", DataType: rpc.DataString, Value: codeString},
			}

			estimate, err := walletInstance.Daemon.GetGasEstimate(&rpc.GasEstimate_Params{
				Ringsize: ringsize,
				SC_RPC:   sc_rpc,
				Signer:   walletInstance.GetAddress(),
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fees := estimate.GasStorage
			yes, err := app.PromptYesNo(fmt.Sprintf("Fees are %s", rpc.FormatMoney(fees)), false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if !yes {
				return nil
			}

			txid, err := walletInstance.Transfer(&rpc.Transfer_Params{
				SC_RPC:   sc_rpc,
				Ringsize: ringsize,
				Fees:     fees,
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

func CommandWalletTransactions() *cli.Command {
	return &cli.Command{
		Name:    "transactions",
		Aliases: []string{"tx"},
		Usage:   "Show transaction history",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			filterType := ctx.Args().First()

			in := filterType == "incoming" || filterType == ""
			out := filterType == "outgoing" || filterType == ""
			coinbase := filterType == "coinbase" || filterType == ""

			entries, err := walletInstance.GetTransfers(&rpc.Get_Transfers_Params{
				In:       in,
				Out:      out,
				Coinbase: coinbase,
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			displayWalletTransactions(entries)
			return nil
		},
	}
}

func CommandInstallSC() *cli.Command {
	return &cli.Command{
		Name:    "install",
		Aliases: []string{"i"},
		Usage:   "Install smart contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			codeFilePath, err := app.Prompt("Enter code filepath", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			code, err := ioutil.ReadFile(codeFilePath)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			//codeString := string(code)
			codeBase64 := base64.StdEncoding.EncodeToString(code)
			ringsize := uint64(2)

			estimate, err := walletInstance.Daemon.GetGasEstimate(&rpc.GasEstimate_Params{
				SC_Code: codeBase64,
				SC_RPC: rpc.Arguments{
					{Name: "entrypoint", DataType: rpc.DataString, Value: codeBase64},
				},
				Signer: walletInstance.GetAddress(),
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fees := estimate.GasStorage
			yes, err := app.PromptYesNo(fmt.Sprintf("Fees are %s", rpc.FormatMoney(fees)), false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if !yes {
				return nil
			}

			txid, err := walletInstance.Transfer(&rpc.Transfer_Params{
				SC_Code: codeBase64,
				/*SC_RPC: rpc.Arguments{
					{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: rpc.SC_INSTALL},
					{Name: rpc.SCCODE, DataType: rpc.DataString, Value: codeString},
				},*/
				Ringsize: ringsize,
				Fees:     fees,
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

func SCCommands() *cli.Command {
	return &cli.Command{
		Name:               "sc",
		Usage:              "Smart contract commands",
		CustomHelpTemplate: utils.AppTemplate,
		Subcommands: []*cli.Command{
			CommandInstallSC(),
			CommandUpdateSC(),
		},
		Action: func(ctx *cli.Context) error {
			ctx.App.Run([]string{"cmd", "help"})
			return nil
		},
	}
}

func DAppApp(app *cli.App) *cli.App {
	return &cli.App{
		Name:                  app.Name,
		Description:           app.Description,
		CustomAppHelpTemplate: utils.AppTemplate,
		Authors:               app.Authors,
		Commands: append(app.Commands,
			CommandDAppInfo(),
			CommandDAppBack(),
			DAppWalletCommands(),
			//CommandSwitchWallet(),
			CommandVersion(app.Name, semver.MustParse(app.Version)),
			CommandExit(),
		),
		Action: func(ctx *cli.Context) error {
			fmt.Println("Command not found. Type 'help' for a list of commands.")
			return nil
		},
	}
}

func DAppWalletCommands() *cli.Command {
	return &cli.Command{
		Name:               "wallet",
		Aliases:            []string{"w"},
		Usage:              "Wallet commands",
		CustomHelpTemplate: utils.AppTemplate,
		Subcommands: []*cli.Command{
			CommandWalletInfo(),
			CommandWalletTransferDero(),
			CommandWalletBalance(),
			CommandWalletAddress(),
			CommandWalletTransactions(),
			CommandSwitchWallet(),
		},
		Action: func(ctx *cli.Context) error {
			ctx.App.Run([]string{"cmd", "help"})
			return nil
		},
	}
}

func WalletApp() *cli.App {
	return &cli.App{
		Name:                  "",
		CustomAppHelpTemplate: utils.AppTemplate,
		Commands: []*cli.Command{
			CommandWalletInfo(),
			CommandDApp(),
			CommandWalletTransferDero(),
			CommandWalletBalance(),
			CommandWalletAddress(),
			CommandWalletTransactions(),
			CommandWalletSeed(),
			CommandSwitchWallet(),
			SCCommands(),
			CommandCloseWallet(),
			CommandExit(),
		},
		Action: func(ctx *cli.Context) error {
			fmt.Println("Command not found. Type 'help' for a list of commands.")
			return nil
		},
	}
}

func WalletCommands() *cli.Command {
	return &cli.Command{
		Name:               "wallet",
		Aliases:            []string{"w"},
		Usage:              "Wallet commands",
		CustomHelpTemplate: utils.AppTemplate,
		Subcommands: []*cli.Command{
			CommandOpenWallet(),
			CommandAttachWallet(),
			CommandEditWallet(),
			CommandDetachWallet(),
			CommandCreateWallet(),
			CommandListWallets(),
		},
		Action: func(ctx *cli.Context) error {
			ctx.App.Run([]string{"cmd", "help"})
			return nil
		},
	}
}
