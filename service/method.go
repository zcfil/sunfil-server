package service

import (
	"zcfil-server/model/rpc/request"
	"zcfil-server/model/rpc/response"
)

var FilecoinMethod = map[string]func(rpc *request.JsonRpc) *response.JsonRpcResult{
	MpoolPush: ServiceGroupApp.LotusServiceGroup.MpoolPush,
}

const (
	SunPondMinerList = "Filecoin.SunPondMinerList"
	MpoolPush        = "Filecoin.MpoolPush"
)
