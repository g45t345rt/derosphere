package app

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	deroWallet "github.com/deroproject/derohe/walletapi"
	"github.com/g45t345rt/derosphere/rpc_client"
)

type WalletInstance struct {
	Id            int64
	Name          string
	DaemonAddress string
	WalletAddress string
	WalletPath    string
	Daemon        *rpc_client.Daemon
	WalletRPC     *rpc_client.Wallet
	WalletDisk    *deroWallet.Wallet_Disk
}

func (w *WalletInstance) SetupDaemon() error {
	w.Daemon = new(rpc_client.Daemon)
	w.Daemon.SetClient(w.DaemonAddress)

	_, err := w.Daemon.GetInfo()
	if err != nil {
		return err
	}

	return nil
}

func (w *WalletInstance) Open() error {
	fmt.Println("Connecting to daemon rpc...")
	err := w.SetupDaemon()
	if err != nil {
		return err
	}

	fmt.Println("Daemon rpc connection was successful.")

	if w.WalletAddress != "" {
		walletRPC := new(rpc_client.Wallet)
		walletRPC.SetClient(w.WalletAddress)

		count := 0
	checkAuth:
		fmt.Println("Connecting to wallet rpc...")
		needAuth, err := walletRPC.NeedAuth()
		if err != nil {
			return err
		}

		if needAuth {
			if count == 0 {
				fmt.Println("Wallet rpc requires authentication...")
			} else {
				fmt.Println("Invalid username or password. Retry...")
			}

			username, err := Prompt("Enter username", "")
			if err != nil {
				return err
			}

			password, err := PromptPassword("Enter password")
			if err != nil {
				return err
			}

			walletRPC.SetClientWithAuth(w.WalletAddress, username, password)
			count++
			goto checkAuth
		}

		w.WalletRPC = walletRPC
	} else if w.WalletPath != "" {
		/*wd, err := os.Getwd()
		if err != nil {
			return err
		}*/

		//path := filepath.ToSlash(fmt.Sprintf("%s/%s", wd, w.WalletPath))
		path := filepath.ToSlash(w.WalletPath)
		fmt.Println(path)
		_, err = os.Stat(path)

		if err != nil {
			return err
		}

	retryPass:
		password, err := PromptPassword("Enter wallet password")
		if err != nil {
			return err
		}

		wallet, err := deroWallet.Open_Encrypted_Wallet(w.WalletPath, password)
		if err != nil {
			if err.Error() == "Invalid Password" {
				fmt.Println("Invalid password")
				goto retryPass
			}

			return err
		}

		w.WalletDisk = wallet
		globals.Arguments["--daemon-address"] = strings.Replace(w.DaemonAddress, "http://", "", -1)
		w.WalletDisk.SetNetwork(globals.IsMainnet())
		w.WalletDisk.SetOnlineMode()
		go deroWallet.Keep_Connectivity()
	}

	return nil
}

func (w *WalletInstance) Close() {
	w.Daemon = nil
	w.WalletRPC = nil

	if w.WalletDisk != nil {
		Context.StopPromptRefresh = true
		w.WalletDisk.Close_Encrypted_Wallet()
		w.WalletDisk = nil
		Context.StopPromptRefresh = false
	}
}

func (w *WalletInstance) Save() {
	sql := `
		update app_wallets set name = ?, daemon_rpc = ?, wallet_rpc = ?, wallet_path = ? where id == ?
	`

	_, err := Context.DB.Exec(sql, w.Name, w.DaemonAddress, w.WalletAddress, w.WalletPath, w.Id)
	if err != nil {
		log.Fatal(err)
	}
}

func (w *WalletInstance) Add() {
	sql := `
		insert into app_wallets(name, daemon_rpc, wallet_rpc, wallet_path)
		values (?,?,?,?)
	`

	res, err := Context.DB.Exec(sql, w.Name, w.DaemonAddress, w.WalletAddress, w.WalletPath)
	if err != nil {
		log.Fatal(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	w.Id = id
	Context.walletInstances = append(Context.walletInstances, w)
}

func (w *WalletInstance) Del(listIndex int) {
	sql := `
		delete from app_wallets where id == ?
	`

	_, err := Context.DB.Exec(sql, w.Id)
	if err != nil {
		log.Fatal(err)
	}

	Context.walletInstances = append(Context.walletInstances[:listIndex], Context.walletInstances[listIndex+1:]...)
}

func (w *WalletInstance) GetConnectionAddress() string {
	if w.WalletAddress != "" {
		return fmt.Sprintf("[rpc]%s", w.WalletAddress)
	} else if w.WalletPath != "" {
		return fmt.Sprintf("[file]%s", w.WalletPath)
	}
	return ""
}

func (w *WalletInstance) GetAddress() string {
	if w.WalletRPC != nil {
		result, err := w.WalletRPC.GetAddress()
		if err != nil {
			log.Fatal(err)
		}
		return result
	} else if w.WalletDisk != nil {
		return w.WalletDisk.GetAddress().String()
	}

	return ""
}

func (w *WalletInstance) GetSeed() string {
	if w.WalletRPC != nil {
		seed, err := w.WalletRPC.GetSeed()
		if err != nil {
			log.Fatal(err)
		}
		return seed
	} else if w.WalletDisk != nil {
		return w.WalletDisk.GetSeed()
	}

	return ""
}

func (w *WalletInstance) GetBalance() uint64 {
	if w.WalletRPC != nil {
		result, err := w.WalletRPC.GetBalance()
		if err != nil {
			log.Fatal(err)
		}
		return result.Balance
	} else if w.WalletDisk != nil {
		m_balance, _ := w.WalletDisk.Get_Balance()
		return m_balance
	}

	return 0
}

func (w *WalletInstance) GetHeight() uint64 {
	if w.WalletRPC != nil {
		result, err := w.WalletRPC.GetHeight()
		if err != nil {
			log.Fatal(err)
		}

		return result.Height
	} else if w.WalletDisk != nil {
		return w.WalletDisk.Get_Height()
	}

	return 0
}

func (w *WalletInstance) GetTransfers(params *rpc.Get_Transfers_Params) ([]rpc.Entry, error) {
	if w.WalletRPC != nil {
		result, err := w.WalletRPC.GetTransfers(params)
		if err != nil {
			return nil, err
		}

		return result.Entries, nil
	} else if w.WalletDisk != nil {
		entries := w.WalletDisk.Show_Transfers(
			params.SCID,
			params.Coinbase,
			params.In,
			params.Out,
			params.Min_Height,
			params.Max_Height,
			params.Sender,
			params.Receiver,
			params.DestinationPort,
			0,
		)

		return entries, nil
	}

	return nil, nil
}

func (w *WalletInstance) Transfer(params *rpc.Transfer_Params) (string, error) {
	if w.WalletRPC != nil {
		result, err := w.WalletRPC.Transfer(params)
		if err != nil {
			return "", err
		}

		return result.TXID, nil
	} else if w.WalletDisk != nil {
		tx, err := w.WalletDisk.TransferPayload0(params.Transfers, params.Ringsize, false, params.SC_RPC, params.Fees, false)
		if err != nil {
			return "", err
		}

		err = w.WalletDisk.SendTransaction(tx)
		if err != nil {
			return "", err
		}

		return tx.GetHash().String(), nil
	}

	return "", nil
}

func (w *WalletInstance) EstimateFeesAndTransfer(transfer *rpc.Transfer_Params) (string, error) {
	signer := w.GetAddress()

	estimate, err := w.Daemon.GetGasEstimate(&rpc.GasEstimate_Params{
		Ringsize:  transfer.Ringsize,
		Signer:    signer,
		Transfers: transfer.Transfers,
		SC_RPC:    transfer.SC_RPC,
	})

	if err != nil {
		return "", err
	}

	transfer.Fees = estimate.GasStorage
	yes, err := PromptYesNo(fmt.Sprintf("Fees are %s", rpc.FormatMoney(transfer.Fees)), false)
	if HandlePromptErr(err) {
		return "", err
	}

	if !yes {
		return "", errors.New("transaction cancelled")
	}

	txid, err := w.Transfer(transfer)

	if err != nil {
		return "", err
	}

	return txid, nil
}
