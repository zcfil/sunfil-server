package rpc

import (
	"zcfil-server/service"
)

type ApiGroup struct {
	LotusV1Api
}

var (
	lotusService = service.ServiceGroupApp.LotusServiceGroup.LotusV1Service
)
