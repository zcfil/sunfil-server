package system

import (
	"github.com/gin-gonic/gin"
	v1 "zcfil-server/api/v1"
)

type HostGroupRouter struct{}

func (s *HostGroupRouter) InitSysHostGroupRouter(Router *gin.RouterGroup) {
	hostGroupRouter := Router.Group("sysNodeInfo")
	hostGroupApi := v1.ApiGroupApp.SystemApiGroup.NodeInfoApi
	{
		hostGroupRouter.GET("getNodeAddress", hostGroupApi.GetNodeAddress)
		hostGroupRouter.GET("getNodeInfoList", hostGroupApi.GetSysNodeList)
		hostGroupRouter.GET("getNodeRepayInfo", hostGroupApi.GetNodeRepayInfo)
		hostGroupRouter.GET("getNodeWithdrawInfo", hostGroupApi.GetNodeWithdrawInfo)
		hostGroupRouter.GET("getNodeDataInfo", hostGroupApi.GetNodeDataInfo)
		hostGroupRouter.GET("getAvailableAssets", hostGroupApi.GetAvailableAssets)
	}
}
