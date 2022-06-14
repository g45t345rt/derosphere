package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/tidwall/buntdb"
	"github.com/urfave/cli/v2"

	"github.com/g45t345rt/derosphere/dapps"
	"github.com/g45t345rt/derosphere/utils"
)

type Config struct {
	Env string
}

type AppContext struct {
	config                Config
	rootApp               *cli.App
	currentDApp           *dapps.DApp
	currentWalletInstance *WalletInstance
	walletInstances       []*WalletInstance
	readlineInstance      *readline.Instance
}

var Context *AppContext

func InitAppContext() {
	app := new(AppContext)
	app.SetRootApp()

	instance, err := readline.New("")
	if err != nil {
		log.Fatal(err)
	}

	//defer instance.Close()

	app.readlineInstance = instance
	app.LoadConfig()
	app.LoadWalletInstances()
	//app.RefreshPrompt()

	Context = app
}

func (app *AppContext) SetRootApp() {
	app.rootApp = &cli.App{
		Name:                  "DeroSphere",
		Commands:              Commands(),
		CustomAppHelpTemplate: AppTemplate,
		Action: func(ctx *cli.Context) error {
			fmt.Println("Command not found. Type 'help' for a list of commands.")
			return nil
		},
	}
}

func (app *AppContext) Run() {
out:
	for {
		Context.RefreshPrompt()
		line, err := app.readlineInstance.Readline()

		switch err {
		case io.EOF:
			break out
		case readline.ErrInterrupt:
			yes, _ := PromptYesNo("Are you sure you want to quit?", false)
			if !yes {
				continue
			}

			break out
		case nil:
		default:
			log.Fatal(err)
			break out
		}

		app.rootApp.Run(strings.Fields("cmd " + line))
	}
}

func (app *AppContext) SetEnv(env string) {
	app.config.Env = env
	//app.RefreshPrompt()
	app.SaveConfig()
}

func (app *AppContext) SetCurrentWalletInstance(wallet *WalletInstance) {
	if app.currentWalletInstance != nil {
		app.currentWalletInstance.Close()
	}

	if wallet == nil {
		app.SetRootApp()
	}

	app.currentWalletInstance = wallet
	//app.RefreshPrompt()
}

func (app *AppContext) SetCurrentDApp(dapp *dapps.DApp) {
	if dapp != nil {
		app.rootApp = DAppApp(dapp)
	} else {
		app.rootApp = WalletApp()
	}

	app.currentDApp = dapp
	//app.RefreshPrompt()
}

func (app *AppContext) GetWalletInstance(name string) (index int, wallet *WalletInstance) {
	for i, w := range Context.walletInstances {
		if w.Name == name {
			return i, w
		}
	}

	return -1, nil
}

func (app *AppContext) AddWalletInstance(wallet *WalletInstance) {
	Context.walletInstances = append(Context.walletInstances, wallet)
}

func (app *AppContext) RemoveWalletInstance(index int) {
	Context.walletInstances = append(Context.walletInstances[:index], Context.walletInstances[index+1:]...)
}

func (app *AppContext) RefreshPrompt() {
	prompt := fmt.Sprintf("[%s] > ", app.config.Env)

	if app.currentWalletInstance != nil {
		prompt = fmt.Sprintf("[%s] > %s > ", app.config.Env, app.currentWalletInstance.Name)
	}

	if app.currentDApp != nil {
		prompt = fmt.Sprintf("[%s] > %s > %s > ", app.config.Env, app.currentWalletInstance.Name, app.currentDApp.Name)
	}

	app.readlineInstance.SetPrompt(prompt)
}

func (app *AppContext) LoadConfig() {
	content, err := ioutil.ReadFile("./data/config.json")
	if err != nil {
		app.config.Env = "mainnet"
		return
	}

	err = json.Unmarshal(content, &app.config)
	if err != nil {
		log.Fatal(err)
	}
}

func (app *AppContext) SaveConfig() {
	configString, err := json.Marshal(app.config)
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

	folder := "data"
	utils.CreateFoldersIfNotExists(folder)

	db, err := buntdb.Open(fmt.Sprintf("./%s/wallets.db", folder))
	if err != nil {
		log.Fatal(err)
	}

	err = db.View(func(tx *buntdb.Tx) error {
		tx.Ascend("", func(key, value string) bool {
			walletInstance := new(WalletInstance)
			walletInstance.Unmarshal(value)
			app.walletInstances = append(app.walletInstances, walletInstance)
			return true
		})

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
