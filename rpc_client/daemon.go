package rpc_client

import (
	"fmt"

	"github.com/deroproject/derohe/rpc"
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

func (d *Daemon) GetInfo() (*rpc.GetInfo_Result, error) {
	var result *rpc.GetInfo_Result
	err := d.client.CallFor(&result, "DERO.GetInfo")
	return result, err
}

func (d *Daemon) GetSC(params *rpc.GetSC_Params) (*rpc.GetSC_Result, error) {
	var result *rpc.GetSC_Result
	err := d.client.CallFor(&result, "DERO.GetSC", params)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Daemon) GetGasEstimate(params *rpc.GasEstimate_Params) (*rpc.GasEstimate_Result, error) {
	var result *rpc.GasEstimate_Result
	err := d.client.CallFor(&result, "DERO.GetGasEstimate", params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Daemon) GetRandomAddresses(params *rpc.GetRandomAddress_Params) (*rpc.GetRandomAddress_Result, error) {
	var result *rpc.GetRandomAddress_Result
	err := d.client.CallFor(&result, "DERO.GetRandomAddress", params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Daemon) GetHeight() (*rpc.Daemon_GetHeight_Result, error) {
	var result *rpc.Daemon_GetHeight_Result
	err := d.client.CallFor(&result, "DERO.GetHeight")
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Daemon) GetBlock(params *rpc.GetBlock_Params) (*rpc.GetBlock_Result, error) {
	var result *rpc.GetBlock_Result
	err := d.client.CallFor(&result, "DERO.GetBlock")
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Daemon) GetTransaction(params *rpc.GetTransaction_Params) (*rpc.GetTransaction_Result, error) {
	var result *rpc.GetTransaction_Result
	err := d.client.CallFor(&result, "DERO.GetTransaction", params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Daemon) NameToAddress(params *rpc.NameToAddress_Params) (*rpc.NameToAddress_Result, error) {
	var result *rpc.NameToAddress_Result
	err := d.client.CallFor(&result, "DERO.NameToAddress", params)
	if err != nil {
		return nil, err
	}

	return result, nil
}
