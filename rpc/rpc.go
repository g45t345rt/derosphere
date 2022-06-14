package rpc

type RPCGetSCParams struct {
	SCID       string   `json:"scid"`
	Code       bool     `json:"code"`
	Variables  bool     `json:"variables"`
	KeysString []string `json:"keysstring"`
}

type RPCGetSCResult struct {
	Balance      int                    `json:"balance"`
	Code         string                 `json:"code"`
	Status       string                 `json:"status"`
	Balances     map[string]int         `json:"balances"`
	StringKeys   map[string]interface{} `json:"stringkeys"`
	ValuesString []string               `json:"valuesstring"`
}

type RPCGetRandomAddressResult struct {
	Status  string   `json:"status"`
	Address []string `json:"address"`
}

type RPCGetAddressResult struct {
	Address string `json:"address"`
}

type RPCGetBalanceResult struct {
	Balance         uint64 `json:"balance"`
	UnlockedBalance uint64 `json:"unlocked_balance"`
}

type RPCGasEstimateResult struct {
	GasCompute uint64 `json:"gascompute"`
	GasStorage uint64 `json:"gasstorage"`
	Status     string `json:"status"`
}

type RPCTransferResult struct {
	TxID string `json:"txid"`
}

type RPCGetInfoResult struct {
	Difficulty   uint64 `json:"difficulty"`
	Height       uint64 `json:"height"`
	StableHeight uint64 `json:"stableheight"`
	TopoHeight   uint64 `json:"topoheight"`
	Testnet      bool   `json:"testnet"`
	Version      string `json:"version"`
	Status       string `json:"status"`
}
