package cli

import (
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/deroproject/derohe/globals"
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

func CommandSwitchWallet() *cli.Command {
	return &cli.Command{
		Name:    "switch",
		Usage:   "Quickly change to another wallet",
		Aliases: []string{"s"},
		Action:  OpenWalletAction,
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
			dapp := dapps.Find(dappName)
			if dapp == nil {
				fmt.Println("App does not exists.")
				return nil
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
		CustomHelpTemplate: AppTemplate,
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

func DAppApp(app *cli.App) *cli.App {
	return &cli.App{
		Name:                  app.Name,
		Description:           app.Description,
		CustomAppHelpTemplate: AppTemplate,
		Authors:               app.Authors,
		Commands: append(app.Commands,
			CommandDAppInfo(),
			CommandDAppBack(),
			CommandSwitchWallet(),
			CommandVersion(semver.MustParse(app.Version)),
			CommandExit(),
		),
		Action: func(ctx *cli.Context) error {
			fmt.Println("Command not found. Type 'help' for a list of commands.")
			return nil
		},
	}
}

func WalletApp() *cli.App {
	return &cli.App{
		Name:                  "",
		CustomAppHelpTemplate: AppTemplate,
		Commands: []*cli.Command{
			CommandWalletInfo(),
			CommandDApp(),
			CommandWalletBalance(),
			CommandWalletAddress(),
			CommandSwitchWallet(),
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
		CustomHelpTemplate: AppTemplate,
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
