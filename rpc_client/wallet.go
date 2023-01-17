package rpc_client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/deroproject/derohe/rpc"
	"github.com/ybbus/jsonrpc/v2"
)

type Wallet struct {
	Address  string
	Endpoint string
	Auth     string
	Client   jsonrpc.RPCClient
}

func (c *Wallet) SetClient(address string) {
	c.Address = address
	c.Endpoint = fmt.Sprintf("%s/json_rpc", c.Address)
	c.Client = jsonrpc.NewClient(c.Endpoint)
}

func (c *Wallet) SetClientWithAuth(address, username string, password string) {
	c.Address = address
	c.Endpoint = fmt.Sprintf("%s/json_rpc", c.Address)
	auth := username + ":" + password
	c.Auth = "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	c.Client = jsonrpc.NewClientWithOpts(c.Endpoint, &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": c.Auth,
		},
	})
}

func (c *Wallet) NeedAuth() (bool, error) {
	_, err := c.Echo()

	switch err.(type) {
	case nil:
		return false, nil
	case *jsonrpc.HTTPError:
		return true, nil
	default:
		return false, err
	}
}

func (c *Wallet) Echo() (string, error) {
	var result string
	err := c.Client.CallFor(&result, "Echo")
	return result, err
}

func (c *Wallet) GetAddress() (string, error) {
	var result *rpc.GetAddress_Result
	err := c.Client.CallFor(&result, "GetAddress")
	return result.Address, err
}

func (c *Wallet) GetBalance(params *rpc.GetBalance_Params) (*rpc.GetBalance_Result, error) {
	var result *rpc.GetBalance_Result
	err := c.Client.CallFor(&result, "GetBalance", params)
	return result, err
}

func (c *Wallet) GetRegistered() (bool, error) {
	res, err := c.Client.Call("GetBalance")
	if err != nil {
		return false, err
	}

	// if this address is not registered on the blockchain
	// the error code will be -32098 and message Account Unregistered
	if res.Error != nil {
		return false, nil
	}

	return true, nil
}

func (c *Wallet) Transfer(params *rpc.Transfer_Params) (*rpc.Transfer_Result, error) {
	var result *rpc.Transfer_Result
	err := c.Client.CallFor(&result, "Transfer", params)
	return result, err
}

func (c *Wallet) GetTransfers(params *rpc.Get_Transfers_Params) (*rpc.Get_Transfers_Result, error) {
	var result *rpc.Get_Transfers_Result
	err := c.Client.CallFor(&result, "GetTransfers", params)
	return result, err
}

func (c *Wallet) GetHeight() (*rpc.GetHeight_Result, error) {
	var result *rpc.GetHeight_Result
	err := c.Client.CallFor(&result, "GetHeight")
	return result, err
}

func (c *Wallet) GetSeed() (string, error) {
	var result *rpc.Query_Key_Result
	err := c.Client.CallFor(&result, "QueryKey", &rpc.Query_Key_Params{
		Key_type: "mnemonic",
	})

	if err != nil {
		return "", err
	}

	return result.Key, nil
}

// Useless func since I use transfer - keep it for archive
func (c *Wallet) InstallSC(code string) (string, error) {
	client := &http.Client{
		Timeout: 5000,
	}

	bcode := bytes.NewBufferString(code)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/install_sc", c.Address), bcode)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Auth != "" {
		req.Header.Set("Authorization", c.Auth)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	var data map[string]string
	err = json.NewDecoder(res.Body).Decode(&data)

	if err != nil {
		return "", err
	}

	return data["txid"], nil
}
