package cli

import (
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/deroproject/derohe/globals"
	"github.com/fatih/color"
	"github.com/g45t345rt/derosphere/dapps"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

func displayAppsTable() {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("", "Name", "Description", "Version")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for index, dapp := range dapps.GetDApps() {
		tbl.AddRow(index, dapp.Name, dapp.Description, dapp.Version)
	}

	tbl.Print()
}

func CommandWalletInfo() *cli.Command {
	return &cli.Command{
		Name:    "info",
		Aliases: []string{"i"},
		Usage:   "Wallet generic information",
		Action: func(ctx *cli.Context) error {
			w := Context.currentWalletInstance
			fmt.Println("Name: ", w.Name)
			fmt.Println("Daemon: ", w.DaemonAddress)
			fmt.Println("Wallet: ", w.GetConnectionAddress())
			return nil
		},
	}
}

func CommandSwitchWallet() *cli.Command {
	return &cli.Command{
		Name:   "switch",
		Usage:  "Quicky change to another wallet",
		Action: OpenWalletAction,
	}
}

func CommandListDApps() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "Show a list of available apps",
		Action: func(ctx *cli.Context) error {
			displayAppsTable()
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
			dapp := dapps.FindDApp(dappName)
			if dapp == nil {
				fmt.Println("DApp does not exists.")
				return nil
			}

			Context.SetCurrentDApp(dapp)

			return nil
		},
	}
}

func CommandDAppBack() *cli.Command {
	return &cli.Command{
		Name:  "back",
		Usage: "Back to wallet",
		Action: func(ctx *cli.Context) error {
			Context.SetCurrentDApp(nil)
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
			Context.SetCurrentWalletInstance(nil)
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
			fmt.Println(Context.currentWalletInstance.GetAddress())
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
			balance := Context.currentWalletInstance.GetBalance()
			fmt.Printf("%s\n", globals.FormatMoney(balance))
			return nil
		},
	}
}

func DAppApp(dapp *dapps.DApp) *cli.App {
	return &cli.App{
		Name:                  "",
		CustomAppHelpTemplate: AppTemplate,
		Commands: append(dapp.Commands,
			CommandDAppInfo(),
			CommandDAppBack(),
			CommandSwitchWallet(),
			CommandVersion(semver.MustParse(dapp.Version)),
			CommandExit(),
		),
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
