package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/blang/semver/v4"
	deroConfig "github.com/deroproject/derohe/config"
	"github.com/deroproject/derohe/cryptography/crypto"
	deroWallet "github.com/deroproject/derohe/walletapi"
	"github.com/fatih/color"
	"github.com/g45t345rt/derosphere/config"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

func displayWalletsTable() {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("", "Name", "Daemon", "Wallet")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for index, w := range Context.walletInstances {
		tbl.AddRow(index, w.Name, w.DaemonAddress, w.GetConnectionAddress())
	}

	tbl.Print()
	if len(Context.walletInstances) == 0 {
		fmt.Println("No wallets")
	}
}

func editWalletInstanceDaemon(walletInstance *WalletInstance) error {
setDaemon:
	nodeType, err := PromptChoose("Set node from", []string{"local", "trustednode", "rpc"}, "local")
	if err != nil {
		return err
	}

	switch nodeType {
	case "local":
		walletInstance.DaemonAddress = fmt.Sprintf("http://localhost:%d", deroConfig.Mainnet.RPC_Default_Port)
	case "trustednode":
		remoteNodeEnv, err := PromptChoose("Remote node environment?", []string{"mainnet", "testnet"}, "mainnet")
		if err != nil {
			return err
		}

		switch remoteNodeEnv {
		case "mainnet":
			walletInstance.DaemonAddress = fmt.Sprintf("http://%s", deroConfig.Mainnet_seed_nodes[0])
		case "testnet":
			walletInstance.DaemonAddress = fmt.Sprintf("http://%s", deroConfig.Testnet_seed_nodes[0])
		}
	case "rpc":
		address, err := Prompt("Enter node rpc address", fmt.Sprintf("http://localhost:%d", deroConfig.Mainnet.RPC_Default_Port))
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

	fmt.Println("Daemon rpc connection was successful.")

	return nil
}

func editWalletInstanceWallet(walletInstance *WalletInstance) error {
	walletType, err := PromptChoose("Set wallet connection from", []string{"rpc", "file"}, "rpc")
	if err != nil {
		return err
	}

	switch walletType {
	case "rpc":
		address, err := Prompt("Enter wallet rpc address", "")
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
		walletFilePath, err := Prompt("Enter wallet file location", "")
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
				name, err = Prompt("Enter wallet name", "")
				if HandlePromptErr(err) {
					return nil
				}
			}

			_, walletInstance := Context.GetWalletInstance(name)
			if walletInstance != nil {
				fmt.Println("A wallet with this name is already attached.")
				name = ""
				goto name
			}

			walletInstance = new(WalletInstance)

			if name == "" {
				fmt.Println("Name cannot be empty.")
				goto name
			}

			walletInstance.Name = name

			err = editWalletInstanceDaemon(walletInstance)
			if HandlePromptErr(err) {
				return nil
			}

			err = editWalletInstanceWallet(walletInstance)
			if HandlePromptErr(err) {
				return nil
			}

			walletInstance.Save()
			Context.AddWalletInstance(walletInstance)

			fmt.Println("New wallet attached and saved.")
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
			walletIndex, walletInstance := Context.GetWalletInstance(name)
			if walletInstance == nil {
				fmt.Println("This wallet does not exists.")
				return nil
			}

			yes, err := PromptYesNo("Are you sure?", true)
			if HandlePromptErr(err) {
				return nil
			}

			if !yes {
				return nil
			}

			walletInstance.Del()
			Context.RemoveWalletInstance(walletIndex)

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
			_, walletInstance := Context.GetWalletInstance(name)
			if walletInstance == nil {
				fmt.Println("This wallet does not exists.")
				return nil
			}

			editType, err := PromptChoose("What do you want to change?", []string{"daemon", "wallet"}, "")
			if HandlePromptErr(err) {
				return nil
			}

			switch editType {
			case "daemon":
				err = editWalletInstanceDaemon(walletInstance)
				if HandlePromptErr(err) {
					return nil
				}

				walletInstance.Save()
			case "wallet":
				err = editWalletInstanceWallet(walletInstance)
				if HandlePromptErr(err) {
					return nil
				}

				walletInstance.Save()
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
			displayWalletsTable()
			return nil
		},
	}
}

func OpenWalletAction(ctx *cli.Context) error {
	walletName := ctx.Args().First()
	var err error = nil

setWalletName:
	if walletName == "" {
		walletName, err = Prompt("Enter wallet name", "")
		if HandlePromptErr(err) {
			return nil
		}
	}

	_, walletInstance := Context.GetWalletInstance(walletName)
	if walletInstance == nil {
		fmt.Println("Wallet does not exists.")
		walletName = ""
		goto setWalletName
	}

	err = walletInstance.Open()
	if HandlePromptErr(err) {
		return nil
	}

	Context.SetCurrentWalletInstance(walletInstance)
	fmt.Println("Wallet connection successful.")
	Context.rootApp = WalletApp()

	return nil
}

func CommandOpenWallet() *cli.Command {
	return &cli.Command{
		Name:    "open",
		Aliases: []string{"o"},
		Usage:   "Use a specific wallet for interacting with apps",
		Action:  OpenWalletAction,
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
				walletFileName, err = Prompt("Enter new wallet filename", "")
				if HandlePromptErr(err) {
					return nil
				}
			}

			if walletFileName == "" {
				fmt.Println("Wallet filename can't be empty.")
				goto setWalletName
			}

			folder := fmt.Sprintf("%s/%s", config.WALLET_FOLDER_PATH, Context.config.Env)
			utils.CreateFoldersIfNotExists(folder)
			filePath := fmt.Sprintf("%s/%s.wallet", folder, walletFileName)

			_, err = os.Stat(filePath)
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Printf("A wallet already exists at this location: %s\n", filePath)
				walletFileName = ""
				goto setWalletName
			}

			createType, err := PromptChoose("Create disk wallet from", []string{"seed-random", "seed-words", "seed-hex"}, "seed-random")
			if HandlePromptErr(err) {
				return nil
			}

			switch createType {
			case "seed-random":
				password, err := PromptPassword("Enter new wallet password")
				if HandlePromptErr(err) {
					return nil
				}

				wallet, err := deroWallet.Create_Encrypted_Wallet_Random(filePath, password)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				fmt.Println("SEED")

				fmt.Println("#########")
				fmt.Println(wallet.GetSeed())
				fmt.Println("#########")

				err = wallet.Save_Wallet()
				if err != nil {
					fmt.Println(err)
					return nil
				}
			case "seed-words":
				seed, err := Prompt("Enter seed (25 words)", "")
				if HandlePromptErr(err) {
					return nil
				}

				password, err := PromptPassword("Enter new wallet password")
				if HandlePromptErr(err) {
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
				seed, err := Prompt("Enter seed (64 chars)", "")
				if HandlePromptErr(err) {
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

				password, err := PromptPassword("Enter new wallet password")
				if HandlePromptErr(err) {
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

func CommandVersion(version semver.Version) *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Display current version",
		Action: func(ctx *cli.Context) error {
			fmt.Println(config.Version)
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
				env, err = PromptChoose("Enter environment", []string{"mainnet", "testnet", "simulator"}, "")
				if HandlePromptErr(err) {
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

			Context.SetEnv(env)
			return nil
		},
	}
}

func Commands() []*cli.Command {
	return []*cli.Command{
		WalletCommands(),
		CommandSetEnv(),
		CommandVersion(config.Version),
		CommandExit(),
	}
}
