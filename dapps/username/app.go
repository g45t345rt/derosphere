package username

import (
	"fmt"
	"log"
	"strings"

	"github.com/deroproject/derohe/rpc"
	"github.com/fatih/color"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/config"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/rodaine/table"
	"github.com/tidwall/buntdb"
	"github.com/urfave/cli/v2"
)

var SC_ID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "",
	"simulator": "900f10626046c2160bbaa9bdaee9bf025ff8596d10d5da8af0c6638ba50277f9",
}

type Name struct {
	Name    string
	Address string
}

func dbFolderPath() string {
	return config.DAPPS_FOLDER + "/" + app.Context.Config.Env + "/username"
}

func dbFilePath(scid string) string {
	return dbFolderPath() + "/" + scid + ".db"
}

func getSCID() string {
	return SC_ID[app.Context.Config.Env]
}

func openDBAndSync(scid string) *buntdb.DB {
	utils.CreateFoldersIfNotExists(dbFolderPath())
	db, err := buntdb.Open(dbFilePath(scid))
	if err != nil {
		log.Fatal(err)
	}

	err = db.CreateIndex("names", "state_name_*", buntdb.IndexString)
	if err != nil {
		log.Fatal(err)
	}

	daemon := app.Context.WalletInstance.Daemon
	err = utils.SyncCommits(db, daemon, scid)
	if err != nil {
		log.Fatal(err)
	}

	return db
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
		Usage:   "Register yourself a nice username",
		Action: func(c *cli.Context) error {
			username, err := app.Prompt("Enter username", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			walletInstance := app.Context.WalletInstance

			scid := getSCID()
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
		Usage:   "Unregister your current username",
		Action: func(c *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			yes, err := app.PromptYesNo("Are you sure?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if !yes {
				return nil
			}

			scid := getSCID()
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

func CommandListNames() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "List of registered names",
		Action: func(c *cli.Context) error {
			scid := getSCID()
			db := openDBAndSync(scid)
			defer db.Close()

			var names []Name
			err := db.View(func(tx *buntdb.Tx) error {
				err := tx.Ascend("names", func(key, value string) bool {
					address := strings.Replace(key, "state_name_", "", -1)
					names = append(names, Name{
						Name:    value,
						Address: address,
					})
					return true
				})

				return err
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			displayNamesTable(names)
			return nil
		},
	}
}

func CommandName() *cli.Command {
	return &cli.Command{
		Name:    "name",
		Usage:   "What is my name?",
		Aliases: []string{"n"},
		Action: func(c *cli.Context) error {
			scid := getSCID()
			db := openDBAndSync(scid)
			defer db.Close()

			walletInstance := app.Context.WalletInstance
			address := walletInstance.GetAddress()
			err := db.View(func(tx *buntdb.Tx) error {
				val, err := tx.Get("state_name_" + address)
				if err != nil {
					return err
				}

				fmt.Println(val)
				return nil
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

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
			CommandListNames(),
			CommandName(),
		},
		Authors: []*cli.Author{
			{Name: "g45t345rt"},
		},
	}
}
