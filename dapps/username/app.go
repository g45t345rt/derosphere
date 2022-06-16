package username

import (
	"fmt"

	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/urfave/cli/v2"
)

var SC_ID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "",
	"simulator": "900f10626046c2160bbaa9bdaee9bf025ff8596d10d5da8af0c6638ba50277f9",
}

func CommandRegister() *cli.Command {
	return &cli.Command{
		Name:    "register",
		Aliases: []string{"r"},
		Usage:   "",
		Action: func(c *cli.Context) error {
			username, err := app.Prompt("Enter username", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance

			scid := SC_ID[app.Context.Config.Env]
			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Register"}
			arg2 := rpc.Argument{Name: "name", DataType: rpc.DataString, Value: username}

			txid, err := walletInstance.EstimateFeesAndTransfer(scid, uint64(2), rpc.Arguments{
				arg1,
				arg2,
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

func CommandUnRegister() *cli.Command {
	return &cli.Command{
		Name:    "unregister",
		Aliases: []string{"u"},
		Usage:   "",
		Action: func(c *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			yes, err := app.PromptYesNo("Are you sure?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if !yes {
				return nil
			}

			scid := SC_ID[app.Context.Config.Env]
			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Unregister"}

			txid, err := walletInstance.EstimateFeesAndTransfer(scid, uint64(2), rpc.Arguments{
				arg1,
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

func App() *cli.App {
	return &cli.App{
		Name:        "username",
		Description: "Register a single username used by other dApps.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandRegister(),
			CommandUnRegister(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
