package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/blang/semver/v4"
	deroConfig "github.com/deroproject/derohe/config"
	"github.com/deroproject/derohe/cryptography/crypto"
	deroWallet "github.com/deroproject/derohe/walletapi"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/config"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

func getDaemonPort() int {
	daemonPort := 0
	switch app.Context.Config.Env {
	case "mainnet":
		daemonPort = deroConfig.Mainnet.RPC_Default_Port
	case "testnet":
		daemonPort = deroConfig.Testnet.RPC_Default_Port
	case "simulator":
		daemonPort = 20000
	}
	return daemonPort
}

func getWalletPort() int {
	walletPort := 0
	switch app.Context.Config.Env {
	case "mainnet":
		walletPort = deroConfig.Mainnet.Wallet_RPC_Default_Port
	case "testnet":
		walletPort = deroConfig.Testnet.Wallet_RPC_Default_Port
	case "simulator":
		walletPort = 30000
	}
	return walletPort
}

func editWalletInstanceDaemon(walletInstance *app.WalletInstance) error {
setDaemon:
	nodeType, err := app.PromptChoose("Set node from", []string{"local", "trustednode", "rpc"}, "local")
	if err != nil {
		return err
	}

	daemonPort := getDaemonPort()
	switch nodeType {
	case "local":
		walletInstance.DaemonAddress = fmt.Sprintf("http://localhost:%d", daemonPort)
	case "trustednode":
		switch app.Context.Config.Env {
		case "mainnet":
			walletInstance.DaemonAddress = fmt.Sprintf("http://%s", deroConfig.Mainnet_seed_nodes[0])
		case "testnet":
			walletInstance.DaemonAddress = fmt.Sprintf("http://%s", deroConfig.Testnet_seed_nodes[0])
		case "simulator":
			walletInstance.DaemonAddress = "http://localhost:20000"
		}
	case "rpc":
		address, err := app.Prompt("Enter node rpc address", fmt.Sprintf("http://localhost:%d", daemonPort))
		if err != nil {
			return err
		}

		walletInstance.DaemonAddress = address
	}

	err = walletInstance.SetupDaemon()
	if err != nil {
		fmt.Println(err)
		goto setDaemon
	}

	info, err := walletInstance.Daemon.GetInfo()
	if err != nil {
		return err
	}

	env := app.Context.Config.Env
	if env == "mainnet" && info.Testnet {
		return fmt.Errorf("Can't attach testnet/simulator daemon to mainnet environment.")
	}

	if (env == "testnet" || env == "simulator") && !info.Testnet {
		return fmt.Errorf("Can't attach mainnet daemon to testnet/simulator environment.")
	}

	fmt.Println("Daemon rpc connection was successful.")

	return nil
}

func editWalletInstanceWallet(walletInstance *app.WalletInstance) error {
	walletType, err := app.PromptChoose("Set wallet connection from", []string{"rpc", "file"}, "rpc")
	if err != nil {
		return err
	}

	walletPort := getWalletPort()
	switch walletType {
	case "rpc":
		address, err := app.Prompt("Enter wallet rpc address", fmt.Sprintf("http://localhost:%d", walletPort))
		if err != nil {
			return err
		}

		walletInstance.WalletAddress = address
		err = walletInstance.Open()
		if err != nil {
			fmt.Println(err)
			return err
		}

	case "file":
		walletFilePath, err := app.Prompt("Enter wallet file location", "")
		if err != nil {
			return err
		}

		walletInstance.WalletPath = walletFilePath
		err = walletInstance.Open()

		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}

func CommandAttachWallet() *cli.Command {
	return &cli.Command{
		Name:    "attach",
		Usage:   "Attach / add existing wallet",
		Aliases: []string{"a"},
		Action: func(ctx *cli.Context) error {
			name := ctx.Args().First()
			var err error = nil

		name:
			if name == "" {
				name, err = app.Prompt("Enter wallet name", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			_, walletInstance := app.Context.GetWalletInstance(name)
			if walletInstance != nil {
				fmt.Println("A wallet with this name is already attached.")
				name = ""
				goto name
			}

			walletInstance = new(app.WalletInstance)

			if name == "" {
				fmt.Println("Name cannot be empty.")
				goto name
			}

			walletInstance.Name = name

			err = editWalletInstanceDaemon(walletInstance)
			if app.HandlePromptErr(err) {
				return nil
			}

			err = editWalletInstanceWallet(walletInstance)
			if app.HandlePromptErr(err) {
				return nil
			}

			err = walletInstance.Add()
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Printf("New wallet %s attached and saved.", name)
			return nil
		},
	}
}

func CommandDetachWallet() *cli.Command {
	return &cli.Command{
		Name:    "detach",
		Usage:   "Detach / remove wallet",
		Aliases: []string{"d"},
		Action: func(ctx *cli.Context) error {
			name := ctx.Args().First()
			listIndex, walletInstance := app.Context.GetWalletInstance(name)
			if walletInstance == nil {
				fmt.Println("This wallet does not exists.")
				return nil
			}

			yes, err := app.PromptYesNo("Are you sure?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if !yes {
				return nil
			}

			err = walletInstance.Del(listIndex)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Printf("Wallet %s detached.\n", name)
			return nil
		},
	}
}

func CommandEditWallet() *cli.Command {
	return &cli.Command{
		Name:    "edit",
		Usage:   "Edit wallet",
		Aliases: []string{"e"},
		Action: func(ctx *cli.Context) error {
			name := ctx.Args().First()
			_, walletInstance := app.Context.GetWalletInstance(name)
			if walletInstance == nil {
				fmt.Println("This wallet does not exists.")
				return nil
			}

			editType, err := app.PromptChoose("What do you want to change?", []string{"daemon", "wallet"}, "")
			if app.HandlePromptErr(err) {
				return nil
			}

			switch editType {
			case "daemon":
				err = editWalletInstanceDaemon(walletInstance)
				if app.HandlePromptErr(err) {
					return nil
				}

				err := walletInstance.Save()
				if err != nil {
					fmt.Println(err)
					return nil
				}
			case "wallet":
				err = editWalletInstanceWallet(walletInstance)
				if app.HandlePromptErr(err) {
					return nil
				}

				err := walletInstance.Save()
				if err != nil {
					fmt.Println(err)
					return nil
				}
			}

			return nil
		},
	}
}

func CommandListWallets() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "List of your wallets",
		Action: func(ctx *cli.Context) error {
			walletInstances := app.Context.GetWalletInstances()
			app.Context.DisplayTable(len(walletInstances), func(i int) []interface{} {
				w := walletInstances[i]
				return []interface{}{
					i, w.Name, w.DaemonAddress, w.GetConnectionAddress(),
				}
			}, []interface{}{"", "Name", "Daemon", "Wallet"}, 25)
			return nil
		},
	}
}

func OpenWalletAction(ctx *cli.Context, useApp string) error {
	walletName := ctx.Args().First()
	var err error = nil

setWalletName:
	if walletName == "" {
		walletName, err = app.Prompt("Enter wallet name", "")
		if app.HandlePromptErr(err) {
			return nil
		}
	}

	_, walletInstance := app.Context.GetWalletInstance(walletName)
	if walletInstance == nil {
		fmt.Println("Wallet does not exists.")
		walletName = ""
		goto setWalletName
	}

	if app.Context.WalletInstance != nil {
		if app.Context.WalletInstance.Name == walletName {
			fmt.Println("Already connected to this wallet.")
			return nil
		}
	}

	err = walletInstance.Open()
	if app.HandlePromptErr(err) {
		return nil
	}

	if app.Context.WalletInstance != nil {
		app.Context.WalletInstance.Close()
	}

	app.Context.WalletInstance = walletInstance
	if useApp != "" {
		app.Context.UseApp = useApp // "walletApp"
	}

	fmt.Println("Wallet connection successful.")
	return nil
}

func CommandOpenWallet() *cli.Command {
	return &cli.Command{
		Name:    "open",
		Aliases: []string{"o"},
		Usage:   "Use a specific wallet for interacting with apps",
		Action: func(ctx *cli.Context) error {
			return OpenWalletAction(ctx, "walletApp")
		},
	}
}

func CommandCreateWallet() *cli.Command {
	return &cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
		Usage:   "Use a specific wallet for interacting with apps",
		Action: func(ctx *cli.Context) error {
			walletFileName := ctx.Args().First()
			var err error = nil

		setWalletName:
			if walletFileName == "" {
				walletFileName, err = app.Prompt("Enter new wallet filename", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			if walletFileName == "" {
				fmt.Println("Wallet filename can't be empty.")
				goto setWalletName
			}

			folder := fmt.Sprintf("%s/%s", config.WALLET_FOLDER_PATH, app.Context.Config.Env)
			utils.CreateFoldersIfNotExists(folder)
			filePath := fmt.Sprintf("%s/%s.wallet", folder, walletFileName)

			_, err = os.Stat(filePath)
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Printf("A wallet already exists at this location: %s\n", filePath)
				walletFileName = ""
				goto setWalletName
			}

			createType, err := app.PromptChoose("Create disk wallet from", []string{"seed-random", "seed-words", "seed-hex"}, "seed-random")
			if app.HandlePromptErr(err) {
				return nil
			}

			switch createType {
			case "seed-random":
				password, err := app.PromptPassword("Enter new wallet password")
				if app.HandlePromptErr(err) {
					return nil
				}

				wallet, err := deroWallet.Create_Encrypted_Wallet_Random(filePath, password)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				fmt.Println("####SEED####")
				fmt.Println(wallet.GetSeed())
				fmt.Println("####SEED####")

				err = wallet.Save_Wallet()
				if err != nil {
					fmt.Println(err)
					return nil
				}
			case "seed-words":
				seed, err := app.Prompt("Enter seed (25 words)", "")
				if app.HandlePromptErr(err) {
					return nil
				}

				password, err := app.PromptPassword("Enter new wallet password")
				if app.HandlePromptErr(err) {
					return nil
				}

				wallet, err := deroWallet.Create_Encrypted_Wallet_From_Recovery_Words(filePath, password, seed)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				err = wallet.Save_Wallet()
				if err != nil {
					fmt.Println(err)
					return nil
				}
			case "seed-hex":
				seed, err := app.Prompt("Enter seed (64 chars)", "")
				if app.HandlePromptErr(err) {
					return nil
				}

				if len(seed) >= 65 {
					fmt.Println("Hex seed is more than 65 chars")
					return nil
				}

				seedRaw, err := hex.DecodeString(seed)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				password, err := app.PromptPassword("Enter new wallet password")
				if app.HandlePromptErr(err) {
					return nil
				}

				wallet, err := deroWallet.Create_Encrypted_Wallet(filePath, password, new(crypto.BNRed).SetBytes(seedRaw))
				if err != nil {
					fmt.Println(err)
					return nil
				}

				err = wallet.Save_Wallet()
				if err != nil {
					fmt.Println(err)
					return nil
				}
			}

			fmt.Printf("Wallet successfully create at %s\n", filePath)

			return nil
		},
	}
}

func CommandVersion(name string, version semver.Version) *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Display current version",
		Action: func(ctx *cli.Context) error {
			fmt.Printf("%s v%s\n", name, version)
			return nil
		},
	}
}

func CommandExit() *cli.Command {
	return &cli.Command{
		Name:    "exit",
		Aliases: []string{"quit", "q"},
		Usage:   "Quit CLI application",
		Action: func(ctx *cli.Context) error {
			os.Exit(1)
			return nil
		},
	}
}

func CommandSetEnv() *cli.Command {
	return &cli.Command{
		Name:  "set-env",
		Usage: "Change environment",
		Action: func(ctx *cli.Context) error {
			env := ctx.Args().First()
			var err error = nil

			if env == "" {
				env, err = app.PromptChoose("Enter environment", []string{"mainnet", "testnet", "simulator"}, "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			switch env {
			case "mainnet":
			case "testnet":
			case "simulator":
			default:
				fmt.Println("Can't set. Invalid environment. Valid env are `mainnet`, `testnet` or `simulator`")
				return nil
			}

			app.Context.SetEnv(env)
			return nil
		},
	}
}

func CommandSetWalletInactivity() *cli.Command {
	return &cli.Command{
		Name:  "set-wallet-inactivity",
		Usage: "Close wallet if inactive after a certain amount of time. Default to 300s and 0 = always opened.",
		Action: func(ctx *cli.Context) error {
			timeoutString := ctx.Args().First()
			var err error = nil
			var timeout uint64

			if timeoutString != "" {
				timeout, err = strconv.ParseUint(timeoutString, 10, 64)
				if err != nil {
					fmt.Println(err)
					return nil
				}
			} else {
				timeout, err = app.PromptUInt("Enter timeout in second", 300)
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			app.Context.SetWalletInactivity(timeout)
			fmt.Printf("Wallet inactivity set to %ds\n", timeout)
			return nil
		},
	}
}

func Commands() []*cli.Command {
	return []*cli.Command{
		WalletCommands(),
		CommandSetEnv(),
		CommandSetWalletInactivity(),
		CommandVersion("derosphere", config.Version),
		CommandExit(),
	}
}

func RootApp() *cli.App {
	return &cli.App{
		Name:     "DeroSphere",
		Commands: Commands(),
		Before: func(c *cli.Context) error {
			app.Context.StopPromptRefresh = true
			return nil
		},
		After: func(c *cli.Context) error {
			app.Context.StopPromptRefresh = false
			return nil
		},
		CustomAppHelpTemplate: utils.AppTemplate,
		Action: func(ctx *cli.Context) error {
			fmt.Println("Command not found. Type 'help' for a list of commands.")
			return nil
		},
	}
}
