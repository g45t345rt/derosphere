package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/transaction"
	"github.com/g45t345rt/derosphere/app"
	"github.com/g45t345rt/derosphere/config"
	"github.com/g45t345rt/derosphere/dapps"
	"github.com/g45t345rt/derosphere/utils"
	"github.com/urfave/cli/v2"
)

func CommandWalletInfo() *cli.Command {
	return &cli.Command{
		Name:    "info",
		Aliases: []string{"i"},
		Usage:   "Wallet generic information",
		Action: func(ctx *cli.Context) error {
			w := app.Context.WalletInstance
			fmt.Println("Name: ", w.Name)
			fmt.Println("Daemon: ", w.DaemonAddress)
			fmt.Println("Wallet: ", w.GetConnectionAddress())
			fmt.Println("Registered: ", w.IsRegistered())
			return nil
		},
	}
}

func CommandDaemonInfo() *cli.Command {
	return &cli.Command{
		Name:    "daemon-info",
		Aliases: []string{"di"},
		Usage:   "Connected daemon information",
		Action: func(ctx *cli.Context) error {
			w := app.Context.WalletInstance

			result, err := w.Daemon.GetInfo()
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("Height: ", result.Height)
			fmt.Println("Testnet: ", result.Testnet)
			fmt.Println("Network: ", result.Network)
			fmt.Println("Version: ", result.Version)
			return nil
		},
	}
}

func CommandRegisterWallet() *cli.Command {
	return &cli.Command{
		Name:    "register",
		Aliases: []string{"r"},
		Usage:   "Register wallet with blockchain (can take up to 2 hours - POW anti-spam)",
		Action: func(ctx *cli.Context) error {
			w := app.Context.WalletInstance

			if w.IsRegistered() {
				fmt.Println("Wallet already registered.")
				return nil
			}

			if w.WalletDisk == nil {
				fmt.Println("Can't register from rpc wallet yet. TODO")
				return nil
			}

			wallet := w.WalletDisk

			fmt.Println("Please wait while the app solves the POW to register the new wallet...")
			fmt.Println("Can take a few hours!")

			var regTx *transaction.Transaction
			chanTx := make(chan *transaction.Transaction)

			counter := 0
			found := false // need this to cancel other parallel loop
			maxThreads := runtime.GOMAXPROCS(0)
			fmt.Printf("Using %d threads\n", maxThreads)

			for i := 0; i < maxThreads; i++ {
				go func() {
					for !found {
						tempTx := wallet.GetRegistrationTX()
						hash := tempTx.GetHash()

						if hash[0] == 0 && hash[1] == 0 && hash[2] == 0 {
							chanTx <- tempTx
							found = true
							break
						}

						counter++
						fmt.Printf("%d tries\r", counter)
					}
				}()
			}

			regTx = <-chanTx
			fmt.Println("Valid registration tx found!")
			fmt.Println("Sending transaction to blockchain...")
			err := wallet.SendTransaction(regTx)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("Your wallet was succesfully registered.")
			return nil
		},
	}
}

func CommandWalletSeed() *cli.Command {
	return &cli.Command{
		Name:  "seed",
		Usage: "Display wallet seed",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			fmt.Println(walletInstance.GetSeed())
			return nil
		},
	}
}

func CommandSwitchWallet() *cli.Command {
	return &cli.Command{
		Name:    "switch",
		Usage:   "Quickly change to another wallet",
		Aliases: []string{"s"},
		Action: func(ctx *cli.Context) error {
			return OpenWalletAction(ctx, "")
		},
	}
}

func CommandListDApps() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "Show a list of available apps",
		Action: func(ctx *cli.Context) error {
			dapps := dapps.List()
			app.Context.DisplayTable(len(dapps), func(i int) []interface{} {
				dapp := dapps[i]
				return []interface{}{
					i, dapp.Name, dapp.Description, dapp.Version, utils.AppAuthors(dapp),
				}
			}, []interface{}{"", "Name", "Description", "Version", "Authors"}, 25)
			return nil
		},
	}
}

func CommandDAppInfo() *cli.Command {
	return &cli.Command{
		Name:    "info",
		Aliases: []string{"i"},
		Usage:   "A basic description of the application",
		Action: func(ctx *cli.Context) error {
			app := ctx.App
			fmt.Printf("Name: %s\n", app.Name)
			fmt.Printf("Description: %s\n", app.Description)
			fmt.Printf("Authors: %s\n", utils.AppAuthors(app))
			return nil
		},
	}
}

func CommandOpenDApp() *cli.Command {
	return &cli.Command{
		Name:    "open",
		Aliases: []string{"o"},
		Usage:   "Open speficic app",
		Action: func(ctx *cli.Context) error {
			dappName := ctx.Args().First()
			var err error

		setAppName:
			if dappName == "" {
				dappName, err = app.Prompt("Enter app name/index", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			appIndex, err := strconv.ParseInt(dappName, 10, 64)
			var dapp *cli.App
			if err == nil {
				list := dapps.List()
				if int(appIndex) < len(list) {
					dapp = dapps.List()[appIndex]
				}
			} else {
				dapp = dapps.Find(dappName)
			}

			if dapp == nil {
				fmt.Println("App does not exists.")
				dappName = ""
				goto setAppName
			}

			app.Context.DAppApp = DAppApp(dapp)
			app.Context.UseApp = "dappApp"

			return nil
		},
	}
}

func CommandDAppBack() *cli.Command {
	return &cli.Command{
		Name:    "back",
		Usage:   "Back to wallet",
		Aliases: []string{"b"},
		Action: func(ctx *cli.Context) error {
			app.Context.DAppApp = nil
			app.Context.UseApp = "walletApp"
			return nil
		},
	}
}

func CommandDApp() *cli.Command {
	return &cli.Command{
		Name:               "app",
		Usage:              "App commands",
		CustomHelpTemplate: utils.AppTemplate,
		Subcommands: []*cli.Command{
			CommandListDApps(),
			CommandOpenDApp(),
		},
		Action: func(ctx *cli.Context) error {
			ctx.App.Run([]string{"cmd", "help"})
			return nil
		},
	}
}

func CommandCloseWallet() *cli.Command {
	return &cli.Command{
		Name:    "close",
		Aliases: []string{"c"},
		Usage:   "Close wallet",
		Action: func(ctx *cli.Context) error {
			app.Context.WalletInstance.Close()
			app.Context.WalletInstance = nil
			app.Context.UseApp = "rootApp"
			return nil
		},
	}
}

func CommandWalletAddress() *cli.Command {
	return &cli.Command{
		Name:    "address",
		Aliases: []string{"a"},
		Usage:   "Wallet address",
		Action: func(ctx *cli.Context) error {
			addr, err := app.Context.WalletInstance.GetAddress()
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(addr)
			return nil
		},
	}
}

func CommandWalletBalance() *cli.Command {
	return &cli.Command{
		Name:    "balance",
		Aliases: []string{"b"},
		Usage:   "Wallet balance",
		Action: func(ctx *cli.Context) error {
			var scid crypto.Hash // default DERO scid
			balance, err := app.Context.WalletInstance.GetBalance(scid)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Printf("%s\n", globals.FormatMoney(balance))
			return nil
		},
	}
}

func CommandDisplayTransaction() *cli.Command {
	return &cli.Command{
		Name:    "view-transaction",
		Aliases: []string{"vtx"},
		Usage:   "Display transaction information",
		Action: func(ctx *cli.Context) error {
			txId := ctx.Args().First()
			var err error
			walletInstance := app.Context.WalletInstance

			if txId == "" {
				txId, err = app.Prompt("Enter txid", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			result, err := walletInstance.Daemon.GetTransaction(&rpc.GetTransaction_Params{
				Tx_Hashes: []string{txId},
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			var tx transaction.Transaction
			tx_bin, _ := hex.DecodeString(result.Txs_as_hex[0])
			tx.Deserialize(tx_bin)

			fmt.Println(tx.Height)
			fmt.Println(tx)
			return nil
		},
	}
}

func CommandWalletBurn() *cli.Command {
	return &cli.Command{
		Name:    "burn",
		Aliases: []string{"bu"},
		Usage:   "Burn DERO/ASSET_TOKEN",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			assetToken, err := app.Prompt("Enter asset token (empty for burning DERO)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			burnAmount := uint64(0)
			if assetToken == "" {
				burnAmount, err = app.PromptDero("Enter burn amount (in Dero)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			} else {
				burnAmount, err = app.PromptUInt("Enter burn amount (atomic value)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			transfer := rpc.Transfer{
				SCID: crypto.HashHexToHash(assetToken),
				Burn: burnAmount,
			}

			ringsize, err := app.PromptUInt("Set ringsize", 2)
			if app.HandlePromptErr(err) {
				return nil
			}

			prompt := ""
			if assetToken == "" {
				prompt = fmt.Sprintf("Are you sure you want to burn %s DERO", globals.FormatMoney(transfer.Burn))
			} else {
				prompt = fmt.Sprintf("Are you sure you want to burn %d TOKEN", transfer.Burn)
			}

			yes, err := app.PromptYesNo(prompt, false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if !yes {
				return nil
			}

			yes, err = app.PromptYesNo("The funds will literally burn like it never existed! Are your really sure?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if !yes {
				return nil
			}

			txid, err := walletInstance.Transfer(&rpc.Transfer_Params{
				Ringsize: ringsize,
				Transfers: []rpc.Transfer{
					transfer,
				},
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txid)
			return nil
		},
	}
}

func CommandWalletTransfer() *cli.Command {
	return &cli.Command{
		Name:    "transfer",
		Aliases: []string{"t"},
		Usage:   "Transfer DERO/ASSET_TOKEN to another address",
		Action: func(ctx *cli.Context) error {

			walletInstance := app.Context.WalletInstance

			assetToken, err := app.Prompt("Enter asset token (empty for sending DERO)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			addressOrName, err := app.Prompt("Enter address/name", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			_, err = globals.ParseValidateAddress(addressOrName)
			if err != nil {
				result, err := walletInstance.Daemon.NameToAddress(&rpc.NameToAddress_Params{
					Name: addressOrName,
				})

				if err != nil {
					fmt.Println(err)
					return nil
				}

				addressOrName = result.Address
				fmt.Printf("Address found: %s\n", addressOrName)
			}

			amount := uint64(0)
			if assetToken == "" {
				amount, err = app.PromptDero("Enter amount (in Dero)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			} else {
				amount, err = app.PromptUInt("Enter amount (atomic value)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			ringsize, err := app.PromptUInt("Set ringsize", 2)
			if app.HandlePromptErr(err) {
				return nil
			}

			transfer := rpc.Transfer{
				SCID:        crypto.HashHexToHash(assetToken),
				Destination: addressOrName,
				Amount:      amount,
			}

			arguments := rpc.Arguments{}

			comment, err := app.Prompt("Comment", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			if comment != "" {
				arguments = append(arguments, rpc.Argument{
					Name:     rpc.RPC_COMMENT,
					DataType: rpc.DataString,
					Value:    comment,
				})
			}

			sPortNumber, err := app.Prompt("Destination port number", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			if sPortNumber != "" {
				portNumber, err := strconv.ParseUint(sPortNumber, 10, 64)
				if err != nil {
					fmt.Println(err)
				}

				arguments = append(arguments, rpc.Argument{
					Name:     rpc.RPC_DESTINATION_PORT,
					DataType: rpc.DataUint64,
					Value:    portNumber,
				})
			}

			transfer.Payload_RPC = arguments

			prompt := ""
			if assetToken == "" {
				prompt = fmt.Sprintf("Are you sure you want to send %s DERO to %s", globals.FormatMoney(transfer.Amount), addressOrName)
			} else {
				prompt = fmt.Sprintf("Are you sure you want to send %d TOKEN to %s", transfer.Amount, addressOrName)
			}

			yes, err := app.PromptYesNo(prompt, false)
			if app.HandlePromptErr(err) {
				return nil
			}

			if !yes {
				return nil
			}

			txid, err := walletInstance.Transfer(&rpc.Transfer_Params{
				Ringsize: ringsize,
				Transfers: []rpc.Transfer{
					transfer,
				},
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txid)
			return nil
		},
	}
}

func CommandWalletTransferFromFile() *cli.Command {
	return &cli.Command{
		Name:    "transfer-from-file",
		Aliases: []string{"tff"},
		Usage:   "Transfer from json file",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			filePath, err := app.Prompt("Filepath", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			var transfer *rpc.Transfer_Params
			err = json.Unmarshal(content, &transfer)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			for _, t := range transfer.Transfers {
				addr := t.Destination

				if !strings.HasPrefix(addr, "dero") {
					result, err := walletInstance.Daemon.NameToAddress(&rpc.NameToAddress_Params{
						Name: addr,
					})
					if err != nil {
						fmt.Println(addr, err)
						return nil
					}
					fmt.Println(addr, "->", result.Address)
				} else {
					_, err = walletInstance.Daemon.GetEncrypedBalance(&rpc.GetEncryptedBalance_Params{
						Address:    t.Destination,
						TopoHeight: -1,
					})

					if err != nil {
						fmt.Println(addr, err)
						return nil
					}
					fmt.Println(addr)
				}
			}

			txid, err := walletInstance.Transfer(transfer)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txid)
			return nil
		},
	}
}

func CommandWalletTransactions() *cli.Command {
	return &cli.Command{
		Name:    "transactions",
		Aliases: []string{"txs"},
		Usage:   "Show transaction history",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			filterType := ctx.Args().First()

			in := filterType == "incoming" || filterType == ""
			out := filterType == "outgoing" || filterType == ""
			coinbase := filterType == "coinbase" || filterType == ""

			entries, err := walletInstance.GetTransfers(&rpc.Get_Transfers_Params{
				In:       in,
				Out:      out,
				Coinbase: coinbase,
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			app.Context.DisplayTable(len(entries), func(i int) []interface{} {
				tx := entries[i]
				return []interface{}{
					i, globals.FormatMoney(tx.Amount), globals.FormatMoney(tx.Burn), globals.FormatMoney(tx.Fees), tx.Time,
					tx.Height, tx.Destination, tx.Coinbase, tx.TXID,
				}
			}, []interface{}{"", "Amount", "Burn", "Fees", "Time", "Height", "Destination", "Coinbase", "TXID"}, 25)
			return nil
		},
	}
}

func CommandInstallSC() *cli.Command {
	return &cli.Command{
		Name:    "install",
		Aliases: []string{"i"},
		Usage:   "Install smart contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance
			codeFilePath, err := app.Prompt("Enter code filepath", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			code, err := ioutil.ReadFile(codeFilePath)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			txId, err := walletInstance.InstallSmartContract(code, 2, []rpc.Argument{}, true)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func CommandUpdateSC() *cli.Command {
	return &cli.Command{
		Name:    "update",
		Aliases: []string{"u"},
		Usage:   "Update smart contract",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			codeFilePath, err := app.Prompt("Enter new code filepath", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			code, err := ioutil.ReadFile(codeFilePath)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			codeString := string(code)
			txId, err := walletInstance.CallSmartContract(2, scid, "UpdateCode", []rpc.Argument{
				{Name: "code", DataType: rpc.DataString, Value: codeString},
			}, []rpc.Transfer{}, true)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

type SCFunc struct {
	Name string
	Args []SCFuncArg
}

type SCFuncArg struct {
	Name string
	Type string
}

func CommandCallSC() *cli.Command {
	return &cli.Command{
		Name:    "call",
		Aliases: []string{"c"},
		Usage:   "Call smart contract function",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			scid, err := app.Prompt("Enter scid", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			result, err := walletInstance.Daemon.GetSC(&rpc.GetSC_Params{
				SCID:      scid,
				Code:      true,
				Variables: false,
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			matchFunctions, err := regexp.Compile(`Function ([A-Z]\w+)\(?(.+)\)`)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			funcs := make(map[string]SCFunc)
			var funcNames []string

			values := matchFunctions.FindAllStringSubmatch(result.Code, -1)
			for _, value := range values {
				funcName := value[1]
				/*if funcName == "Initialize" || funcName == "PrivateInitialize" || funcName == "UpdateCode" {
					continue
				}*/

				scFunc := SCFunc{
					Name: funcName,
				}

				sArgs := value[2]
				if sArgs != "(" {
					args := strings.Split(sArgs, ",")

					for _, arg := range args {
						def := strings.Split(strings.Trim(arg, " "), " ")
						scFunc.Args = append(scFunc.Args, SCFuncArg{
							Name: def[0],
							Type: def[1],
						})
					}
				}

				funcs[funcName] = scFunc
				funcNames = append(funcNames, funcName)
			}

			funcName, err := app.PromptChoose("Function to excute", funcNames, "")
			//funcName, err := app.Prompt("Function to excute", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			function := funcs[funcName]

			args := []rpc.Argument{}
			for _, arg := range function.Args {
				switch arg.Type {
				case "Uint64":
					valueUInt, err := app.PromptUInt(arg.Name, 0)
					if app.HandlePromptErr(err) {
						return nil
					}

					args = append(args, rpc.Argument{
						Name:     arg.Name,
						DataType: rpc.DataUint64,
						Value:    valueUInt,
					})
				case "String":
					sArg := rpc.Argument{}

					valueString, err := app.Prompt(arg.Name, "")
					if app.HandlePromptErr(err) {
						return nil
					}

					isHashString, err := app.PromptYesNo(fmt.Sprintf("Is arg %s hash string?", arg.Name), false)
					if app.HandlePromptErr(err) {
						return nil
					}

					if isHashString {
						sArg = rpc.Argument{
							Name:     arg.Name,
							DataType: rpc.DataHash,
							Value:    crypto.HashHexToHash(valueString),
						}
					} else {
						sArg = rpc.Argument{
							Name:     arg.Name,
							DataType: rpc.DataString,
							Value:    valueString,
						}
					}

					args = append(args, sArg)
				default:
					fmt.Println("Unknown arg type")
					return nil
				}
			}

			ringSize, err := app.PromptUInt("Ringsize", 2)
			if app.HandlePromptErr(err) {
				return nil
			}

			burnAssetOrDero, err := app.PromptYesNo("Burn asset or dero?", false)
			if app.HandlePromptErr(err) {
				return nil
			}

			var transfer []rpc.Transfer

			if burnAssetOrDero {
				assetToken, err := app.Prompt("Enter asset token (empty for sending DERO)", "")
				if app.HandlePromptErr(err) {
					return nil
				}

				amount, err := app.PromptUInt("Enter amount (atomic value)", 0)
				if app.HandlePromptErr(err) {
					return nil
				}

				randomAddresses, err := walletInstance.Daemon.GetRandomAddresses(nil)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				transfer = append(transfer, rpc.Transfer{
					SCID:        crypto.HashHexToHash(assetToken),
					Burn:        amount,
					Destination: randomAddresses.Address[0],
				})
			}

			txId, err := walletInstance.CallSmartContract(ringSize, scid, function.Name, args, transfer, true)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			walletInstance.RunTxChecker(txId)
			return nil
		},
	}
}

func SCCommands() *cli.Command {
	return &cli.Command{
		Name:               "sc",
		Usage:              "Smart contract commands",
		CustomHelpTemplate: utils.AppTemplate,
		Subcommands: []*cli.Command{
			CommandInstallSC(),
			CommandUpdateSC(),
			CommandCallSC(),
		},
		Action: func(ctx *cli.Context) error {
			ctx.App.Run([]string{"cmd", "help"})
			return nil
		},
	}
}

func CommandAssetBalance() *cli.Command {
	return &cli.Command{
		Name:    "asset-balance",
		Aliases: []string{"ab"},
		Usage:   "Display specific asset balance",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			scid := ctx.Args().First()
			var err error

			if scid == "" {
				scid, err = app.Prompt("Enter asset token", "")
				if app.HandlePromptErr(err) {
					return nil
				}
			}

			hash := crypto.HashHexToHash(scid)
			balance, err := walletInstance.GetBalance(hash)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(balance)
			return nil
		},
	}
}

func CommandClearCommitCount() *cli.Command {
	return &cli.Command{
		Name:    "clear-commit",
		Aliases: []string{"cc"},
		Usage:   "Clear commit count to resync values",
		Action: func(ctx *cli.Context) error {
			name := app.Context.DAppApp.Name
			count := utils.Count{Filename: config.GetCountFilename(app.Context.Config.Env)}
			err := count.Load()
			if err != nil {
				return err
			}

			count.Set(name, 0)
			err = count.Save()
			if err != nil {
				return err
			}

			fmt.Printf("dapps [%s] count reset\n", name)
			return nil
		},
	}
}

func CommandAccountExists() *cli.Command {
	return &cli.Command{
		Name:    "wallet-exists",
		Usage:   "Check if wallet address as balance",
		Aliases: []string{"we"},
		Action: func(ctx *cli.Context) error {

			addr, err := app.Prompt("Enter wallet address", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			result, err := app.Context.WalletInstance.Daemon.GetEncrypedBalance(&rpc.GetEncryptedBalance_Params{
				Address:    addr,
				TopoHeight: -1,
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(result)
			return nil
		},
	}
}

func DAppApp(dapp *cli.App) *cli.App {
	return &cli.App{
		Name:                  dapp.Name,
		Description:           dapp.Description,
		CustomAppHelpTemplate: utils.AppTemplate,
		Authors:               dapp.Authors,
		Before: func(c *cli.Context) error {
			app.Context.StopPromptRefresh = true
			return nil
		},
		After: func(c *cli.Context) error {
			app.Context.StopPromptRefresh = false
			return nil
		},
		Commands: append(dapp.Commands,
			CommandDAppInfo(),
			CommandDAppBack(),
			DAppWalletCommands(),
			CommandClearCommitCount(),
			//CommandSwitchWallet(),
			CommandVersion(dapp.Name, semver.MustParse(dapp.Version)),
			CommandExit(),
		),
		Action: func(ctx *cli.Context) error {
			fmt.Println("Command not found. Type 'help' for a list of commands.")
			return nil
		},
	}
}

func DAppWalletCommands() *cli.Command {
	return &cli.Command{
		Name:               "wallet",
		Aliases:            []string{"w"},
		Usage:              "Wallet commands",
		CustomHelpTemplate: utils.AppTemplate,
		Subcommands: []*cli.Command{
			CommandWalletInfo(),
			CommandWalletTransfer(),
			CommandWalletBalance(),
			CommandAssetBalance(),
			CommandWalletAddress(),
			CommandWalletTransactions(),
			CommandSwitchWallet(),
		},
		Action: func(ctx *cli.Context) error {
			ctx.App.Run([]string{"cmd", "help"})
			return nil
		},
	}
}

func CommandGetEncrypedBalance() *cli.Command {
	return &cli.Command{
		Name:    "get-encrypted-balance",
		Aliases: []string{"geb"},
		Usage:   "Get encrupted balance for wallet token",
		Action: func(ctx *cli.Context) error {
			walletInstance := app.Context.WalletInstance

			walletAddress, err := app.Prompt("Wallet Address", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			scid, err := app.Prompt("Token ID (SCID)", "")
			if app.HandlePromptErr(err) {
				return nil
			}

			result, err := walletInstance.Daemon.GetEncrypedBalance(&rpc.GetEncryptedBalance_Params{
				Address:    walletAddress,
				SCID:       crypto.HashHexToHash(scid),
				TopoHeight: -1,
			})

			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(result)
			return nil
		},
	}
}

func WalletApp() *cli.App {
	return &cli.App{
		Name:                  "",
		CustomAppHelpTemplate: utils.AppTemplate,
		Before: func(c *cli.Context) error {
			app.Context.StopPromptRefresh = true
			return nil
		},
		After: func(c *cli.Context) error {
			app.Context.StopPromptRefresh = false
			return nil
		},
		Commands: []*cli.Command{
			CommandWalletInfo(),
			CommandDaemonInfo(),
			CommandDApp(),
			CommandWalletTransfer(),
			CommandWalletTransferFromFile(),
			CommandWalletBurn(),
			CommandWalletBalance(),
			CommandAssetBalance(),
			CommandWalletAddress(),
			CommandDisplayTransaction(),
			CommandWalletTransactions(),
			CommandWalletSeed(),
			CommandRegisterWallet(),
			CommandSwitchWallet(),
			CommandAccountExists(),
			CommandGetEncrypedBalance(),
			SCCommands(),
			CommandCloseWallet(),
			CommandExit(),
		},
		Action: func(ctx *cli.Context) error {
			fmt.Println("Command not found. Type 'help' for a list of commands.")
			return nil
		},
	}
}

func WalletCommands() *cli.Command {
	return &cli.Command{
		Name:               "wallet",
		Aliases:            []string{"w"},
		Usage:              "Wallet commands",
		CustomHelpTemplate: utils.AppTemplate,
		Subcommands: []*cli.Command{
			CommandOpenWallet(),
			CommandAttachWallet(),
			CommandEditWallet(),
			CommandDetachWallet(),
			CommandCreateWallet(),
			CommandListWallets(),
		},
		Action: func(ctx *cli.Context) error {
			ctx.App.Run([]string{"cmd", "help"})
			return nil
		},
	}
}
