package rpc

import (
	"fmt"

	"github.com/ybbus/jsonrpc/v2"
)

type Daemon struct {
	Address  string
	Endpoint string
	client   jsonrpc.RPCClient
}

func (d *Daemon) SetClient(address string) {
	d.Address = address
	d.Endpoint = fmt.Sprintf("%s/json_rpc", d.Address)
	d.client = jsonrpc.NewClient(d.Endpoint)
}

func (d *Daemon) Ping() (string, error) {
	var result string
	err := d.client.CallFor(&result, "DERO.Ping")
	return result, err
}

func (d *Daemon) GetInfo() (*RPCGetInfoResult, error) {
	var result *RPCGetInfoResult
	err := d.client.CallFor(&result, "DERO.GetInfo")
	return result, err
}

func (d *Daemon) GetSC(params interface{}) (*RPCGetSCResult, error) {
	var result *RPCGetSCResult
	err := d.client.CallFor(&result, "DERO.GetSC", params)

	if err != nil {
		return nil, err
	}

	return result, nil
}
