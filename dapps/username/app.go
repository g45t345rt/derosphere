package username

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"database/sql"

	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/rpc_client"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

var DAPP_NAME = "username"

var SC_ID map[string]string = map[string]string{
	"mainnet":   "",
	"testnet":   "",
	"simulator": "900f10626046c2160bbaa9bdaee9bf025ff8596d10d5da8af0c6638ba50277f9",
}

type Name struct {
	Name    string
	Address string
}

func getSCID() string {
	return SC_ID[app.Context.Config.Env]
}

func initData() {
	sqlQuery := `
		create table if not exists username (
			wallet_address varchar primary key,
			name varchar
		);
	`

	db := app.Context.DB
	_, err := db.Exec(sqlQuery)
	if err != nil {
		log.Fatal(err)
	}

	// reset table
	counts := utils.GetCounts()
	commitAt := counts[DAPP_NAME]
	if commitAt == 0 {
		sqlQuery = `
			delete from username
		`

		_, err = db.Exec(sqlQuery)
		if err != nil {
			log.Fatal(err)
		}
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
	nameKey, err := regexp.Compile(`state_name_(.+)`)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	sqlQuery := `
		insert into username (wallet_address, name)
		values (?,?)
		on conflict(wallet_address) do update set name = ?
	`

	setTx, err := tx.Prepare(sqlQuery)
	if err != nil {
		log.Fatal(err)
	}

	defer setTx.Close()

	sqlQuery = `
		delete from username where wallet_address == ?
	`

	delTx, err := tx.Prepare(sqlQuery)
	if err != nil {
		log.Fatal(err)
	}

	defer setTx.Close()

	var i uint64
	for i = commitAt; i < commitCount; i += chunk {
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

			if strings.HasPrefix(commit.Key, "state_name_") {
				walletAddress := nameKey.ReplaceAllString(key, "$1")
				if commit.Action == "S" {
					_, err := setTx.Exec(walletAddress, commit.Value, commit.Value)
					if err != nil {
						log.Fatal(err)
					}

					continue
				}

				if commit.Action == "D" {
					_, err := delTx.Exec(walletAddress)
					if err != nil {
						log.Fatal(err)
					}

					continue
				}
			}
		}

		err := tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

		utils.SetCount(DAPP_NAME, commitAt)
	}
}

func CommandRegister() *cli.Command {
	return &cli.Command{
		Name:    "register",
		Aliases: []string{"r"},
		Usage:   "Register yourself a nice username",
		Action: func(c *cli.Context) error {
			username := c.Args().First()
			var err error

			if username == "" {
				username, err = app.Prompt("Enter username", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			walletInstance := app.Context.WalletInstance

			scid := getSCID()
			arg_sc := rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: scid}
			arg_sc_action := rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: rpc.SC_CALL}
			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Register"}
			arg2 := rpc.Argument{Name: "name", DataType: rpc.DataString, Value: username}

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
			arg_sc := rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: scid}
			arg_sc_action := rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: rpc.SC_CALL}
			arg1 := rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Unregister"}

			txid, err := walletInstance.EstimateFeesAndTransfer(&rpc.Transfer_Params{
				Ringsize: 2,
				SC_RPC: rpc.Arguments{
					arg_sc, arg_sc_action, arg1,
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

func CommandListNames() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "List of registered names",
		Action: func(c *cli.Context) error {
			sync()

			db := app.Context.DB

			query := `
				select wallet_address, name from username
			`

			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
			}

			var names []Name
			for rows.Next() {
				var name Name
				err = rows.Scan(&name.Address, &name.Name)
				if err != nil {
					log.Fatal(err)
				}

				names = append(names, name)
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

func CommandName() *cli.Command {
	return &cli.Command{
		Name:    "name",
		Usage:   "What is my username?",
		Aliases: []string{"n"},
		Action: func(c *cli.Context) error {
			sync()

			db := app.Context.DB
			walletAddress := app.Context.WalletInstance.GetAddress()
			sqlQuery := `select name from username where wallet_address == ?`

			row := db.QueryRow(sqlQuery, walletAddress)
			var name string
			err := row.Scan(&name)
			if err != nil {
				if err == sql.ErrNoRows {
					fmt.Println("You don't have a registered username")
				} else {
					log.Fatal(err)
				}
			} else {
				fmt.Println(name)
			}

			return nil
		},
	}
}

func App() *cli.App {
	initData()
	return &cli.App{
		Name:        DAPP_NAME,
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
