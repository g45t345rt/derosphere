package nameservice

import (
	"encoding/hex"
	"fmt"

	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/urfave/cli/v2"
)

var SC_ID = "0000000000000000000000000000000000000000000000000000000000000001"

type Name struct {
	Name    string
	Address string
}

func CommandRegister() *cli.Command {
	return &cli.Command{
		Name:    "register",
		Aliases: []string{"r"},
		Usage:   "Register a name to your address",
		Action: func(c *cli.Context) error {
			username := c.Args().First()
			var err error
			if username == "" {
				username, err = app.Prompt("Enter name", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			walletInstance := app.Context.WalletInstance
			txid, err := walletInstance.CallSmartContract(2, SC_ID, "Register", []rpc.Argument{
				{Name: "name", DataType: rpc.DataString, Value: username},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txid)
			return nil
		},
	}
}

func CommandNames() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "List names registered to your address",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			result, err := walletInstance.Daemon.GetSC(&rpc.GetSC_Params{
				SCID:      SC_ID,
				Variables: true,
				Code:      false,
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			var names []Name
			for key, value := range result.VariableStringKeys {
				if key != "C" {
					a, _ := hex.DecodeString(fmt.Sprint(value))
					addr, _ := rpc.NewAddressFromCompressedKeys(a)
					names = append(names, Name{
						Name:    key,
						Address: addr.String(),
					})
				}
			}

			app.Context.DisplayTable(len(names), func(i int) []interface{} {
				n := names[i]
				return []interface{}{
					i, n.Name, n.Address,
				}
			}, []interface{}{"", "Name", "Address"}, 25)
			return nil
		},
	}
}

func App() *cli.App {
	return &cli.App{
		Name:        "nameservice",
		Description: "Register multiple names to receive DERO from others.",
		Version:     "0.0.1",
		Commands: []*cli.Command{
			CommandRegister(),
			CommandNames(),
		},
		Authors: []*cli.Author{
			{Name: "Captain"},
			{Name: "g45t345rt"},
		},
	}
}
