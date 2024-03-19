package rpc

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"zcfil-server/model/common/response"
	"zcfil-server/model/rpc/request"
	"zcfil-server/service"
	"zcfil-server/utils"
)

type LotusV1Api struct{}

func (l *LotusV1Api) RpcV1(c *gin.Context) {
	var req request.JsonRpc
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err = utils.Verify(req, utils.JsonRpc); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	fun := service.FilecoinMethod[req.Method]
	if fun == nil {
		fun = lotusService.RequestLotusFunc
	}

	c.JSON(http.StatusOK, fun(&req))
}
