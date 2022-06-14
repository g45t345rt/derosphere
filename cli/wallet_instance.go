package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	deroWallet "github.com/deroproject/derohe/walletapi"
	"github.com/g45t345rt/derosphere/config"
	"github.com/g45t345rt/derosphere/rpc"
	"github.com/tidwall/buntdb"
)

type WalletInstance struct {
	Name          string
	DaemonAddress string
	WalletAddress string
	WalletPath    string
	Daemon        *rpc.Daemon
	WalletRPC     *rpc.Wallet
	WalletDisk    *deroWallet.Wallet_Disk
}

func (w *WalletInstance) SetupDaemon() error {
	w.Daemon = new(rpc.Daemon)
	w.Daemon.SetClient(w.DaemonAddress)

	_, err := w.Daemon.GetInfo()
	if err != nil {
		return err
	}

	return nil
}

func (w *WalletInstance) Open() error {
	err := w.SetupDaemon()
	if err != nil {
		return err
	}

	fmt.Println("Daemon rpc connection was successful.")

	if w.WalletAddress != "" {
		walletRPC := new(rpc.Wallet)
		walletRPC.SetClient(w.WalletAddress)

		count := 0
	checkAuth:
		needAuth := walletRPC.NeedAuth()

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
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		path := filepath.ToSlash(fmt.Sprintf("%s/%s", wd, w.WalletPath))
		fmt.Println(path)
		_, err = os.Stat(path)

		if err != nil {
			return err
		}

		password, err := PromptPassword("Enter wallet password")
		if err != nil {
			return err
		}

		wallet, err := deroWallet.Open_Encrypted_Wallet(w.WalletPath, password)
		if err != nil {
			return err
		}

		w.WalletDisk = wallet
		w.WalletDisk.SetDaemonAddress(w.DaemonAddress)
	}

	return nil
}

func (w *WalletInstance) Close() {
	w.Daemon = nil
	w.WalletRPC = nil

	if w.WalletDisk != nil {
		w.WalletDisk.Close_Encrypted_Wallet()
	}
}

func (w *WalletInstance) Save() {
	db, err := buntdb.Open(config.DB_WALLETS_FILEPATH)

	if err != nil {
		log.Fatal(err)
	}

	data, err := w.Marshal()
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(w.Name, string(data), nil)
		return err
	})
}

func (w *WalletInstance) Del() {
	db, err := buntdb.Open(config.DB_WALLETS_FILEPATH)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(w.Name)
		return err
	})
}

func (w *WalletInstance) Marshal() ([]byte, error) {
	instance := map[string]interface{}{
		"name":   w.Name,
		"daemon": w.DaemonAddress,
	}

	if w.WalletAddress != "" {
		instance["wallet_rpc"] = w.WalletAddress
	} else if w.WalletPath != "" {
		instance["wallet_path"] = w.WalletPath
	}

	return json.Marshal(instance)
}

func (w *WalletInstance) Unmarshal(data string) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		log.Fatal(err)
	}

	w.Name = fmt.Sprint(result["name"])
	w.DaemonAddress = fmt.Sprint(result["daemon"])

	if result["wallet_rpc"] != nil {
		w.WalletAddress = fmt.Sprint(result["wallet_rpc"])
	} else if result["wallet_path"] != nil {
		w.WalletPath = fmt.Sprint(result["wallet_path"])
	}
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

func (w *WalletInstance) GetBalance() uint64 {
	if w.WalletRPC != nil {
		result, err := w.WalletRPC.GetBalance()
		if err != nil {
			log.Fatal(err)
		}
		return result.Balance
	}

	return 0
}

func LoadWalletInstances() {
	Context.walletInstances = []*WalletInstance{}

	folder := config.DATA_FOLDER
	_, err := os.Stat(folder)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(folder, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}

	db, err := buntdb.Open(fmt.Sprintf("%s/wallets.db", folder))
	if err != nil {
		log.Fatal(err)
	}

	err = db.View(func(tx *buntdb.Tx) error {
		tx.Ascend("", func(key, value string) bool {
			walletInstance := new(WalletInstance)
			walletInstance.Unmarshal(value)
			Context.walletInstances = append(Context.walletInstances, walletInstance)
			return true
		})

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
