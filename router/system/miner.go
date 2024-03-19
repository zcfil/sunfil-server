package system

import (
	"github.com/gin-gonic/gin"
	v1 "zcfil-server/api/v1"
)

type MinerRouter struct{}

func (m *MinerRouter) InitMinerRouter(Router *gin.RouterGroup) {
	minerGroupRouter := Router.Group("miner")
	minerApi := v1.ApiGroupApp.SystemApiGroup.MinerApi
	{
		minerGroupRouter.GET("pondMinerList", minerApi.PondMinerList)
		minerGroupRouter.GET("pondOperator", minerApi.PondOperator)
		minerGroupRouter.GET("minerInfo", minerApi.MinerInfo)
		minerGroupRouter.GET("workerInfo", minerApi.WorkerInfo)
		minerGroupRouter.GET("departInfo", minerApi.DepartInfo)
		minerGroupRouter.GET("msigInspect", minerApi.MsigInspect)
	}
}
