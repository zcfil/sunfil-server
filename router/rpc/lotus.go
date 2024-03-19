package rpc

import (
	"github.com/gin-gonic/gin"
	v1 "zcfil-server/api/v1"
)

type LotusGroupRouter struct{}

func (s *LotusGroupRouter) InitLotusGroupRouter(Router *gin.RouterGroup) {
	lotusGroupRouter := Router.Group("rpc")
	LotusApi := v1.ApiGroupApp.LotusApiGroup
	{
		lotusGroupRouter.POST("v0", LotusApi.RpcV1)
		lotusGroupRouter.POST("v1", LotusApi.RpcV1)
	}
}
