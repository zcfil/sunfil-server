package utils

var (
	JsonRpc = Rules{"Jsonrpc": {NotEmpty()}, "Method": {NotEmpty()}}
)
