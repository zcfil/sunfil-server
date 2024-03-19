package system

import (
	"github.com/gin-gonic/gin"
	v1 "zcfil-server/api/v1"
)

type LiquidateRouter struct{}

func (l *LiquidateRouter) InitLiquidateRouter(Router *gin.RouterGroup) {
	hostGroupRouter := Router.Group("liquidate")
	hostGroupApi := v1.ApiGroupApp.SystemApiGroup.LiquidateApi
	{
		hostGroupRouter.GET("warnNodeList", hostGroupApi.WarnNodeList)
		hostGroupRouter.GET("warnNodeDetail", hostGroupApi.WarnNodeDetail)
		hostGroupRouter.GET("liquidation", hostGroupApi.Liquidation)
		hostGroupRouter.GET("warnNodeCount", hostGroupApi.WarnNodeCount)
		hostGroupRouter.GET("contractWarnNode", hostGroupApi.ContractWarnNode)

	}
}
