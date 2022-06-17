package app

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

	"github.com/g45t345rt/derosphere/utils"
)

type Config struct {
	Env string
}

type AppContext struct {
	Config           Config
	UseApp           string
	rootApp          *cli.App
	walletApp        *cli.App
	DAppApp          *cli.App
	WalletInstance   *WalletInstance
	walletInstances  []*WalletInstance
	readlineInstance *readline.Instance
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

	//defer instance.Close()

	app.readlineInstance = instance
	app.LoadConfig()
	app.LoadWalletInstances()

	Context = app
}

func (app *AppContext) Run() {
out:
	for {
		app.RefreshPrompt()
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

func (app *AppContext) SetEnv(env string) {
	app.Config.Env = env
	app.SaveConfig()
	app.LoadWalletInstances()
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

func (app *AppContext) AddWalletInstance(wallet *WalletInstance) {
	app.walletInstances = append(app.walletInstances, wallet)
}

func (app *AppContext) RemoveWalletInstance(index int) {
	app.walletInstances = append(app.walletInstances[:index], app.walletInstances[index+1:]...)
}

func (app *AppContext) RefreshPrompt() {
	prompt := fmt.Sprintf("[%s] > ", app.Config.Env)

	if app.WalletInstance != nil {
		prompt = fmt.Sprintf("[%s] > %s > ", app.Config.Env, app.WalletInstance.Name)
	}

	if app.DAppApp != nil {
		prompt = fmt.Sprintf("[%s] > %s > %s > ", app.Config.Env, app.WalletInstance.Name, app.DAppApp.Name)
	}

	app.readlineInstance.SetPrompt(prompt)
}

func (app *AppContext) LoadConfig() {
	content, err := ioutil.ReadFile("./data/config.json")
	if err != nil {
		app.Config.Env = "mainnet"
		return
	}

	err = json.Unmarshal(content, &app.Config)
	if err != nil {
		log.Fatal(err)
	}
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

	folder := "data"
	utils.CreateFoldersIfNotExists(folder)

	db, err := buntdb.Open(fmt.Sprintf("./%s/wallets.db", folder))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(tx *buntdb.Tx) error {
		tx.Ascend("", func(key, value string) bool {
			walletInstance := new(WalletInstance)
			walletInstance.Unmarshal(value)
			if walletInstance.Env == app.Config.Env {
				app.walletInstances = append(app.walletInstances, walletInstance)
			}

			return true
		})

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
