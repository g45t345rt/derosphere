package nameservice

import (
	"encoding/hex"
	"fmt"

	"github.com/deroproject/derohe/rpc"
	"github.com/fatih/color"
	"github.com/g45t345rt/derosphere/app"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

var SC_ID = "0000000000000000000000000000000000000000000000000000000000000001"

type Name struct {
	Name    string
	Address string
}

func displayNamesTable(names []Name) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("", "Name", "Address")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for index, n := range names {
		tbl.AddRow(index, n.Name, n.Address)
	}

	tbl.Print()
	if len(names) == 0 {
		fmt.Println("No names")
	}
}

func CommandRegister() *cli.Command {
	return &cli.Command{
		Name:    "register",
		Aliases: []string{"r"},
		Usage:   "Register a name to your address",
		Action: func(c *cli.Context) error {
			username, err := app.Prompt("Enter name", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance

			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Register"}
			arg2 := rpc.Argument{Name: "name", DataType: rpc.DataString, Value: username}

			txid, err := walletInstance.EstimateFeesAndTransfer(SC_ID, 2, nil, rpc.Arguments{
				arg1, arg2,
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

			displayNamesTable(names)
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
