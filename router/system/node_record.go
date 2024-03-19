package system

import (
	"github.com/gin-gonic/gin"
	v1 "zcfil-server/api/v1"
)

type NodeRecordRouter struct{}

func (n *NodeRecordRouter) InitNodeRecordRouter(Router *gin.RouterGroup) {
	nodeRecordRouter := Router.Group("nodeRecord")
	nodeRecordApi := v1.ApiGroupApp.SystemApiGroup.NodeRecordApi
	{
		nodeRecordRouter.GET("nodeRecordList", nodeRecordApi.GetNodeRecordList)
		nodeRecordRouter.GET("testWithdraw", nodeRecordApi.GetTestWithdraw)
	}
}
