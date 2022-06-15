package username

import (
	"fmt"

	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/urfave/cli/v2"
)

var SCID_MAINNET = ""
var SCID_TESTNET = ""
var SCID_SIMULATOR = ""

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

			scid := ""
			//signer := walletInstance.GetAddress()
			ringsize := uint64(2)

			arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Register"}
			arg2 := rpc.Argument{Name: "name", DataType: "S", Value: username}

			txid, err := walletInstance.EstimateFeesAndTransfer(scid, ringsize, rpc.Arguments{
				arg1,
				arg2,
			})

			/*
				arg_sc := rpc.Argument{Name: "SC_ID", DataType: "H", Value: scid}
				arg_sc_action := rpc.Argument{Name: "SC_ACTION", DataType: "U", Value: 0}

				arg1 := rpc.Argument{Name: "entrypoint", DataType: "S", Value: "Register"}
				arg2 := rpc.Argument{Name: "name", DataType: "S", Value: username}

				estimate, err := walletInstance.Daemon.GetGasEstimate(&rpc.GasEstimate_Params{
					Ringsize: ringsize,
					Signer:   signer,
					SC_RPC: rpc.Arguments{
						arg_sc,
						arg_sc_action,
						arg1,
						arg2,
					},
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
					SC_ID:    scid,
					Ringsize: ringsize,
					Fees:     estimate.GasStorage,
					SC_RPC: rpc.Arguments{
						arg1,
						arg2,
					},
				})
			*/

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
