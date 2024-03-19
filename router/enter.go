package router

import (
	"zcfil-server/router/rpc"
	"zcfil-server/router/system"
)

type RouterGroup struct {
	System system.RouterGroup
	Rpc    rpc.LotusGroupRouter
}

var RouterGroupApp = new(RouterGroup)
