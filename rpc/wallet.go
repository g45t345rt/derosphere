package rpc

import (
	"encoding/base64"
	"fmt"

	"github.com/ybbus/jsonrpc/v2"
)

type Wallet struct {
	Address  string
	Endpoint string
	Username string
	client   jsonrpc.RPCClient
}

func (c *Wallet) SetClient(address string) {
	c.Address = address
	c.Endpoint = fmt.Sprintf("%s/json_rpc", c.Address)
	c.client = jsonrpc.NewClient(c.Endpoint)
}

func (c *Wallet) SetClientWithAuth(address, username string, password string) {
	c.Address = address
	c.Endpoint = fmt.Sprintf("%s/json_rpc", c.Address)
	auth := username + ":" + password
	c.Username = username
	c.client = jsonrpc.NewClientWithOpts(c.Endpoint, &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(auth)),
		},
	})
}

func (c *Wallet) NeedAuth() bool {
	_, err := c.Echo()

	switch err.(type) {
	case nil:
		return false
	case *jsonrpc.HTTPError:
		return true
	default:
		return false
	}
}

func (c *Wallet) Echo() (string, error) {
	var result string
	err := c.client.CallFor(&result, "Echo")
	return result, err
}

func (c *Wallet) GetAddress() (string, error) {
	var result *RPCGetAddressResult
	err := c.client.CallFor(&result, "GetAddress")
	return result.Address, err
}

func (c *Wallet) GetBalance() (*RPCGetBalanceResult, error) {
	var result *RPCGetBalanceResult
	err := c.client.CallFor(&result, "GetBalance")
	return result, err
}
