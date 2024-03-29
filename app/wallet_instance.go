package app

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/transaction"
	"github.com/deroproject/derohe/walletapi"
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
	WalletDisk    *walletapi.Wallet_Disk
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

		wallet, err := walletapi.Open_Encrypted_Wallet(w.WalletPath, password)
		if err != nil {
			if err.Error() == "Invalid Password" {
				fmt.Println("Invalid password")
				goto retryPass
			}

			return err
		}

		w.WalletDisk = wallet

		httpKey, err := regexp.Compile("https?://")
		if err != nil {
			return err
		}

		globals.Arguments["--daemon-address"] = httpKey.ReplaceAllString(w.DaemonAddress, "")
		w.WalletDisk.SetNetwork(globals.IsMainnet())
		w.WalletDisk.SetOnlineMode()
		go walletapi.Keep_Connectivity()
	}

	return nil
}

func (w *WalletInstance) Close() {
	w.Daemon = nil
	w.WalletRPC = nil

	if w.WalletDisk != nil {
		w.WalletDisk.Close_Encrypted_Wallet()
		w.WalletDisk = nil
	}
}

func (w *WalletInstance) Save() error {
	sql := `
		update app_wallets set name = ?, daemon_rpc = ?, wallet_rpc = ?, wallet_path = ? where id == ?
	`

	_, err := Context.DB.Exec(sql, w.Name, w.DaemonAddress, w.WalletAddress, w.WalletPath, w.Id)
	return err
}

func (w *WalletInstance) Add() error {
	sql := `
		insert into app_wallets(name, daemon_rpc, wallet_rpc, wallet_path)
		values (?,?,?,?)
	`

	res, err := Context.DB.Exec(sql, w.Name, w.DaemonAddress, w.WalletAddress, w.WalletPath)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	w.Id = id
	Context.walletInstances = append(Context.walletInstances, w)
	return nil
}

func (w *WalletInstance) Del(listIndex int) error {
	sql := `
		delete from app_wallets where id == ?
	`

	_, err := Context.DB.Exec(sql, w.Id)
	if err != nil {
		return err
	}

	Context.walletInstances = append(Context.walletInstances[:listIndex], Context.walletInstances[listIndex+1:]...)
	return nil
}

func (w *WalletInstance) GetConnectionAddress() string {
	if w.WalletAddress != "" {
		return fmt.Sprintf("[rpc]%s", w.WalletAddress)
	} else if w.WalletPath != "" {
		return fmt.Sprintf("[file]%s", w.WalletPath)
	}
	return ""
}

func (w *WalletInstance) IsRegistered() bool {
	registered := false
	if w.WalletRPC != nil {
		registered, _ = w.WalletRPC.GetRegistered()
	} else if w.WalletDisk != nil {
		registered = w.WalletDisk.IsRegistered()
	}

	return registered
}

func (w *WalletInstance) GetAddress() (string, error) {
	if w.WalletRPC != nil {
		result, err := w.WalletRPC.GetAddress()
		if err != nil {
			return "", err
		}
		return result, nil
	} else if w.WalletDisk != nil {
		return w.WalletDisk.GetAddress().String(), nil
	}

	return "", nil
}

func (w *WalletInstance) GetSeed() (string, error) {
	if w.WalletRPC != nil {
		seed, err := w.WalletRPC.GetSeed()
		if err != nil {
			return "", err
		}
		return seed, nil
	} else if w.WalletDisk != nil {
		return w.WalletDisk.GetSeed(), nil
	}

	return "", nil
}

func (w *WalletInstance) GetBalance(scid crypto.Hash) (uint64, error) {
	if w.WalletRPC != nil {
		result, err := w.WalletRPC.GetBalance(&rpc.GetBalance_Params{
			SCID: scid,
		})
		if err != nil {
			return 0, err
		}

		return result.Balance, nil
	} else if w.WalletDisk != nil {
		err := w.WalletDisk.Sync_Wallet_Memory_With_Daemon_internal(scid)
		if err != nil {
			return 0, err
		}

		m_balance, _ := w.WalletDisk.Get_Balance_scid(scid)
		return m_balance, nil
	}

	return 0, nil
}

func (w *WalletInstance) GetHeight() (uint64, error) {
	if w.WalletRPC != nil {
		result, err := w.WalletRPC.GetHeight()
		if err != nil {
			return 0, err
		}

		return result.Height, nil
	} else if w.WalletDisk != nil {
		return w.WalletDisk.Get_Height(), nil
	}

	return 0, nil
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

func (w *WalletInstance) Transfer(p *rpc.Transfer_Params) (string, error) {
	if w.WalletRPC != nil {
		result, err := w.WalletRPC.Transfer(p)
		if err != nil {
			return "", err
		}

		return result.TXID, nil
	} else if w.WalletDisk != nil {
		for _, t := range p.Transfers {
			_, err := t.Payload_RPC.CheckPack(transaction.PAYLOAD0_LIMIT)
			if err != nil {
				return "", err
			}
		}

		if len(p.SC_Code) >= 1 {
			if sc, err := base64.StdEncoding.DecodeString(p.SC_Code); err == nil {
				p.SC_Code = string(sc)
			}
		}

		if p.SC_Code != "" && p.SC_ID == "" {
			p.SC_RPC = append(p.SC_RPC, rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_INSTALL)})
			p.SC_RPC = append(p.SC_RPC, rpc.Argument{Name: rpc.SCCODE, DataType: rpc.DataString, Value: p.SC_Code})
		}

		if p.SC_ID != "" {
			p.SC_RPC = append(p.SC_RPC, rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)})
			p.SC_RPC = append(p.SC_RPC, rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(p.SC_ID)})
			if p.SC_Code != "" {
				p.SC_RPC = append(p.SC_RPC, rpc.Argument{Name: rpc.SCCODE, DataType: rpc.DataString, Value: p.SC_Code})
			}
		}

		tx, err := w.WalletDisk.TransferPayload0(p.Transfers, p.Ringsize, false, p.SC_RPC, p.Fees, false)
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

	params := rpc.GasEstimate_Params{
		Ringsize:  transfer.Ringsize,
		Transfers: transfer.Transfers,
		SC_RPC:    transfer.SC_RPC,
	}

	if params.Ringsize == 2 {
		signer, err := w.GetAddress()
		if err != nil {
			return "", err
		}
		params.Signer = signer
	}

	estimate, err := w.Daemon.GetGasEstimate(&params)

	if err != nil {
		return "", err
	}

	transfer.Fees = estimate.GasStorage
	yes, err := PromptYesNo(fmt.Sprintf("TX fees are %s. Do you want to send the transaction?", rpc.FormatMoney(transfer.Fees)), false)
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

func (walletInstance *WalletInstance) InstallSmartContract(code []byte, ringsize uint64, args []rpc.Argument, promptFees bool) (string, error) {
	codeBase64 := base64.StdEncoding.EncodeToString(code)
	signer, err := walletInstance.GetAddress()
	if err != nil {
		return "", err
	}

	sc_rpc := rpc.Arguments{
		{Name: "entrypoint", DataType: rpc.DataString, Value: codeBase64},
	}

	sc_rpc = append(sc_rpc, args[:]...)

	estimate, err := walletInstance.Daemon.GetGasEstimate(&rpc.GasEstimate_Params{
		SC_Code: codeBase64,
		SC_RPC:  sc_rpc,
		Signer:  signer,
	})

	if err != nil {
		return "", err
	}

	fees := estimate.GasStorage

	if promptFees {
		yes, err := PromptYesNo(fmt.Sprintf("TX fees are %s. Do you want to send the transaction?", rpc.FormatMoney(fees)), false)
		if err != nil {
			return "", err
		}

		if !yes {
			return "", fmt.Errorf("cancelled")
		}
	}

	txid, err := walletInstance.Transfer(&rpc.Transfer_Params{
		SC_Code:  codeBase64,
		Ringsize: ringsize,
		Fees:     fees,
		SC_RPC:   args,
	})

	if err != nil {
		return "", err
	}

	return txid, nil
}

func (walletInstance *WalletInstance) CallSmartContract(ringsize uint64, scid string, entrypoint string, args []rpc.Argument, transfers []rpc.Transfer, promptFees bool) (string, error) {
	sc_rpc := rpc.Arguments{
		{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},  // 'SC_ACTION' value should be of type uint64
		{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)}, // 'SC_ID' value should be of type Hash
		{Name: "entrypoint", DataType: rpc.DataString, Value: entrypoint},
	}

	sc_rpc = append(sc_rpc, args[:]...)

	signer := ""
	var err error
	if ringsize == 2 {
		signer, err = walletInstance.GetAddress()
		if err != nil {
			return "", err
		}
	}

	estimate, err := walletInstance.Daemon.GetGasEstimate(&rpc.GasEstimate_Params{
		Ringsize:  ringsize,
		SC_RPC:    sc_rpc,
		Transfers: transfers,
		Signer:    signer,
	})

	if err != nil {
		return "", err
	}

	fees := estimate.GasStorage

	if promptFees {
		yes, err := PromptYesNo(fmt.Sprintf("TX fees are %s. Do you want to send the transaction?", rpc.FormatMoney(fees)), false)
		if err != nil {
			return "", err
		}

		if !yes {
			return "", fmt.Errorf("cancelled")
		}
	}

	txid, err := walletInstance.Transfer(&rpc.Transfer_Params{
		SC_RPC:    sc_rpc,
		Transfers: transfers,
		Ringsize:  ringsize,
		Fees:      fees,
	})

	if err != nil {
		return "", err
	}

	return txid, nil
}

func (walletInstance *WalletInstance) RunTxChecker(txid string) {
	tries := 25
	waitInterval := 2 * time.Second
	var i int

	fmt.Printf("Checking transaction... TXID: %s\n", txid)
	// TODO fmt.Println("Type anything to skip")
	for i = 0; i < tries; i++ {
		result, err := walletInstance.Daemon.GetTransaction(&rpc.GetTransaction_Params{
			Tx_Hashes: []string{txid},
		})

		if err != nil {
			fmt.Println(err)
			break
		}

		txInfo := result.Txs[0]
		if !txInfo.In_pool && txInfo.ValidBlock == "" {
			fmt.Println("Invalid transaction")
			break
		}

		txBlockHeight := txInfo.Block_Height

		if txBlockHeight != -1 {
			fmt.Printf("Successful transaction at block %d\n", txBlockHeight)
			break
		}

		time.Sleep(waitInterval)
	}

	if i == tries {
		fmt.Println("Can't confirm transaction. Number of tries exceeded.")
	}
}

func (walletInstance *WalletInstance) WaitTransaction(txid string) error {
	fmt.Printf("Waiting for transaction... %s\n", txid)

	startHeight := uint64(0)
	tries := 25
	waitInterval := 2 * time.Second
	var i int
	var err error

	for i = 0; i < tries; i++ {
		time.Sleep(waitInterval)

		var result *rpc.GetTransaction_Result
		result, err = walletInstance.Daemon.GetTransaction(&rpc.GetTransaction_Params{
			Tx_Hashes: []string{txid},
		})

		if err != nil {
			continue
		}

		txInfo := result.Txs[0]
		if !txInfo.In_pool && txInfo.ValidBlock == "" {
			err = errors.New("invalid transaction")
			break
		}

		txBlockHeight := txInfo.Block_Height

		var currentHeight uint64
		currentHeight, err = walletInstance.GetHeight()
		if err != nil {
			continue
		}

		if startHeight == 0 {
			startHeight = currentHeight
		}

		//fmt.Printf("%d %d %d\n", currentHeight, txBlockHeight, startHeight)

		if txBlockHeight != -1 {
			err = nil
			break
		}

		if currentHeight >= startHeight+2 {
			err = errors.New("stuck transaction")
			break
		}
	}

	if i == tries {
		err = errors.New(`maximum tries exceeded`)
	}

	return err
}
