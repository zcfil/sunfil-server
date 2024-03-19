package response

import (
	"reflect"
)

type JsonRpcResult struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Id      int         `json:"id"`
	Error   interface{} `json:"error,omitempty"`
}

func JsonRpcResultError(err error, id int) *JsonRpcResult {
	if err != nil {
		return &JsonRpcResult{Jsonrpc: "2.0", Result: nil, Id: id, Error: err.Error()}
	}
	return &JsonRpcResult{Jsonrpc: "2.0", Result: nil, Id: id}
}

func JsonRpcResultOk(result interface{}, id int) *JsonRpcResult {
	return &JsonRpcResult{Jsonrpc: "2.0", Result: result, Id: id}
}

func IsEmptyParams(param interface{}) bool {
	if param == nil {
		return true
	}
	if reflect.ValueOf(param).Type().String() != "[]interface {}" {
		return true
	}
	if len(param.([]interface{})) == 0 {
		return true
	}
	return false
}
