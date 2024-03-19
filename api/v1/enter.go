package v1

import (
	"zcfil-server/api/v1/rpc"
	"zcfil-server/api/v1/system"
)

type ApiGroup struct {
	SystemApiGroup system.ApiGroup
	LotusApiGroup  rpc.ApiGroup
}

var ApiGroupApp = new(ApiGroup)
