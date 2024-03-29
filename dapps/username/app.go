package username

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"database/sql"

	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/config"
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
		create table if not exists dapps_username (
			wallet_address varchar primary key,
			name varchar
		);
	`

	db := app.Context.DB
	_, err := db.Exec(sqlQuery)
	if err != nil {
		log.Fatal(err)
	}

	count := utils.Count{Filename: config.GetCountFilename(app.Context.Config.Env)}
	err = count.Load()
	if err != nil {
		log.Fatal(err)
	}

	commitAt := count.Get(DAPP_NAME)
	if commitAt == 0 {
		sqlQuery = `
			delete from dapps_username
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
	count := utils.Count{Filename: config.GetCountFilename(app.Context.Config.Env)}
	err := count.Load()
	if err != nil {
		log.Fatal(err)
	}

	commitAt := count.Get(DAPP_NAME)
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
		insert into dapps_username (wallet_address, name)
		values (?,?)
		on conflict(wallet_address) do update set name = ?
	`

	setTx, err := tx.Prepare(sqlQuery)
	if err != nil {
		log.Fatal(err)
	}

	defer setTx.Close()

	sqlQuery = `
		delete from dapps_username where wallet_address == ?
	`

	delTx, err := tx.Prepare(sqlQuery)
	if err != nil {
		log.Fatal(err)
	}

	defer setTx.Close()

	for i := commitAt; i < commitCount; i += chunk {
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

		count.Set(DAPP_NAME, commitAt)
		err = count.Save()
		if err != nil {
			log.Fatal(err)
		}
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
			txid, err := walletInstance.CallSmartContract(2, scid, "Register", []rpc.Argument{
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
			txid, err := walletInstance.CallSmartContract(2, scid, "Unregister", []rpc.Argument{}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txid)
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
				select wallet_address, name from dapps_username
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
			walletAddress, err := app.Context.WalletInstance.GetAddress()
			if err != nil {
				fmt.Println(err)
				return nil
			}

			sqlQuery := `select name from dapps_username where wallet_address == ?`

			row := db.QueryRow(sqlQuery, walletAddress)
			var name string
			err = row.Scan(&name)
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
