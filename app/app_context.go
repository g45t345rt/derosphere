package app

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"

	deroConfig "github.com/deroproject/derohe/config"
	"github.com/deroproject/derohe/globals"
	"github.com/g45t345rt/derosphere/config"
	"github.com/g45t345rt/derosphere/utils"
)

type Config struct {
	Env              string
	CloseWalletAfter uint64
}

type AppContext struct {
	Config            Config
	UseApp            string
	rootApp           *cli.App
	walletApp         *cli.App
	DAppApp           *cli.App
	WalletInstance    *WalletInstance
	walletInstances   []*WalletInstance
	readlineInstance  *readline.Instance
	DB                *sql.DB
	StopPromptRefresh bool // prompt auto refresh every second to display block height - use this arg to disable and show other prompt
}

var Context *AppContext

func InitAppContext(rootApp *cli.App, walletApp *cli.App) {
	app := new(AppContext)
	app.rootApp = rootApp
	app.walletApp = walletApp
	app.UseApp = "rootApp"

	instance, err := readline.New("")
	if err != nil {
		log.Fatal(err)
	}

	app.readlineInstance = instance
	app.LoadConfig()
	app.LoadDB()
	app.LoadWalletInstances()

	Context = app
}

func (app *AppContext) Run() {
	var m sync.Mutex
	// auto refresh block height
	go func() {
		for {
			m.Lock()
			if !app.StopPromptRefresh {
				app.RefreshPrompt()
			}
			m.Unlock()

			time.Sleep(1 * time.Second)
		}
	}()

	// listen to keypress
	lastActivity := time.Now()
	app.readlineInstance.Config.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		lastActivity = time.Now()
		return nil, 0, false
	})

	// close wallet after certain amount of time
	go func() {
		for {
			if !app.StopPromptRefresh {
				closeWalletAfter := time.Duration(app.Config.CloseWalletAfter) * time.Second
				if closeWalletAfter > 0 && app.WalletInstance != nil && time.Now().After(lastActivity.Add(closeWalletAfter)) {
					m.Lock()
					app.WalletInstance.Close()
					app.WalletInstance = nil
					app.UseApp = "rootApp"
					app.DAppApp = nil
					fmt.Printf("\nWallet close after %ds of inactivity.\n", app.Config.CloseWalletAfter)
					m.Unlock()
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()

out:
	for {
		app.RefreshPrompt()
		line, err := app.readlineInstance.Readline()

		switch err {
		case io.EOF:
			break out
		case readline.ErrInterrupt:
			app.StopPromptRefresh = true
			yes, _ := PromptYesNo("Are you sure you want to quit?", false)
			app.StopPromptRefresh = false
			if !yes {
				continue
			}

			break out
		case nil:
		default:
			log.Fatal(err)
			break out
		}

		switch app.UseApp {
		case "rootApp":
			app.rootApp.Run(strings.Fields("cmd " + line))
		case "walletApp":
			app.walletApp.Run(strings.Fields("cmd " + line))
		case "dappApp":
			app.DAppApp.Run(strings.Fields("cmd " + line))
		}
	}
}

func (app *AppContext) LoadDB() {
	dataFolder := config.DATA_FOLDER
	utils.CreateFoldersIfNotExists(dataFolder)
	db, err := sql.Open("sqlite3", dataFolder+"/"+app.Config.Env+".db")
	if err != nil {
		log.Fatal(err)
	}

	sql := `
		create table if not exists app_wallets (
			id integer primary key,
			name varchar unique,
			daemon_rpc varchar,
			wallet_rpc varchar,
			wallet_path varchar
		);
	`

	_, err = db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}

	app.DB = db
}

func (app *AppContext) setEnvGlobals() {
	// we need this if want to use wallet SetOnlineMode() and sync wallet with daemon
	switch app.Config.Env {
	case "mainnet":
		globals.Config = deroConfig.Mainnet
	case "testnet":
		globals.Config = deroConfig.Testnet
	case "simulator":
		globals.Config = deroConfig.Testnet
		globals.Arguments["--simulator"] = true
	}
}

func (app *AppContext) SetEnv(env string) {
	app.Config.Env = env

	app.setEnvGlobals()
	app.SaveConfig()
	app.DB.Close()
	app.LoadDB()
	app.LoadWalletInstances()
}

func (app *AppContext) SetWalletInactivity(timeout uint64) {
	app.Config.CloseWalletAfter = timeout
	app.SaveConfig()
}

func (app *AppContext) GetWalletInstance(name string) (index int, wallet *WalletInstance) {
	for i, w := range app.walletInstances {
		if w.Name == name {
			return i, w
		}
	}

	return -1, nil
}

func (app *AppContext) GetWalletInstances() []*WalletInstance {
	return app.walletInstances
}

func (app *AppContext) RefreshPrompt() {
	prompt := fmt.Sprintf("[%s] > ", app.Config.Env)

	if app.WalletInstance != nil {
		daemonHeight, _ := app.WalletInstance.Daemon.GetHeight()
		walletHeight := app.WalletInstance.GetHeight()
		prompt = fmt.Sprintf("[%s] (%d/%d) > %s > ", app.Config.Env, walletHeight, daemonHeight.Height, app.WalletInstance.Name)

		if app.DAppApp != nil {
			prompt = fmt.Sprintf("%s%s > ", prompt, app.DAppApp.Name)
		}
	}

	app.readlineInstance.SetPrompt(prompt)
	app.readlineInstance.Refresh()
}

func (app *AppContext) LoadConfig() {
	content, err := ioutil.ReadFile("./data/config.json")
	if err != nil {
		app.Config.Env = "mainnet"
		app.Config.CloseWalletAfter = 180 // default 180s (3min)
		return
	}

	err = json.Unmarshal(content, &app.Config)
	if err != nil {
		log.Fatal(err)
	}

	app.setEnvGlobals()
}

func (app *AppContext) SaveConfig() {
	configString, err := json.Marshal(app.Config)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./data/config.json", configString, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func (app *AppContext) LoadWalletInstances() {
	app.walletInstances = []*WalletInstance{}

	query := `
		select * from app_wallets
	`

	rows, err := app.DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		walletInstance := new(WalletInstance)
		err = rows.Scan(&walletInstance.Id,
			&walletInstance.Name,
			&walletInstance.DaemonAddress,
			&walletInstance.WalletAddress,
			&walletInstance.WalletPath,
		)

		if err != nil {
			log.Fatal(err)
		}

		app.walletInstances = append(app.walletInstances, walletInstance)
	}
}

type TableRowDef func(i int) []interface{}

func (app *AppContext) DisplayTable(count int, rowFunc TableRowDef, headers []interface{}, pageSize int) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	cursor := 0

	tbl := table.New(headers...)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

printTable:
	if count == 0 {
		fmt.Println("No data available")
		return
	}

	end := cursor + pageSize
	for cursor < end {
		if cursor >= count {
			break
		}

		row := rowFunc(cursor)
		tbl.AddRow(row...)
		cursor++
	}

	tbl.Print()

	if cursor < count {
		fmt.Println("Press any key to load more or q to exit...")
		app.readlineInstance.Terminal.EnterRawMode()
		defer app.readlineInstance.Terminal.ExitRawMode()

		buf := bufio.NewReader(os.Stdin)
		ru, _, err := buf.ReadRune()
		if ru != 113 && err == nil {
			//cursor += pageSize
			goto printTable
		}
	}
}
