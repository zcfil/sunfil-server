package service

import (
	"zcfil-server/service/rpc"
	"zcfil-server/service/system"
)

type ServiceGroup struct {
	SystemServiceGroup system.ServiceGroup
	LotusServiceGroup  rpc.ServiceGroup
}

var ServiceGroupApp = new(ServiceGroup)
