package request

import "github.com/filecoin-project/lotus/chain/types"

type JsonRpc struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int         `json:"id"`
}

type JsonParam struct {
	Message types.Message
}
