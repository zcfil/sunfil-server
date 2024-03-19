package system

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"zcfil-server/model/common/response"
	"zcfil-server/model/system/request"
)

type LiquidateApi struct {
}

// WarnNodeList Alarm node
func (l *LiquidateApi) WarnNodeList(c *gin.Context) {
	var warnNodeReq request.WarnNodeReq
	err := c.ShouldBindQuery(&warnNodeReq)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := liquidateService.GetWarnNodeList(warnNodeReq)
	if err != nil {
		response.FailWithMessage(fmt.Sprintf("Getting information failure:%+v", err), c)
		return
	}

	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     warnNodeReq.Page,
		PageSize: warnNodeReq.PageSize,
	}, "Success", c)
}

// WarnNodeDetail Alarm node details
func (l *LiquidateApi) WarnNodeDetail(c *gin.Context) {

	var warnNodeReq request.WarnNodeReq
	err := c.ShouldBindQuery(&warnNodeReq)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	res, err := liquidateService.WarnNodeDetail(warnNodeReq)
	if err != nil {
		response.FailWithMessage(fmt.Sprintf("Getting information failure:%+v", err), c)
		return
	}

	response.OkWithDetailed(res, "Success", c)
}

// Liquidation Liquidation
func (l *LiquidateApi) Liquidation(c *gin.Context) {

	err := liquidateService.Liquidation()
	if err != nil {
		response.FailWithMessage(fmt.Sprintf("Getting information failure:%+v", err), c)
		return
	}

	response.OkWithDetailed(nil, "Success", c)
}

// WarnNodeCount Obtain the number of voting alarms
func (l *LiquidateApi) WarnNodeCount(c *gin.Context) {

	res, err := liquidateService.GetWarnNodeCount()
	if err != nil {
		response.FailWithMessage(fmt.Sprintf("Getting information failure:%+v", err), c)
		return
	}

	response.OkWithDetailed(res, "Success", c)
}

// ContractWarnNode Obtain node alarm data
func (l *LiquidateApi) ContractWarnNode(c *gin.Context) {

	err := liquidateService.ReContractWarnNode()
	if err != nil {
		response.FailWithMessage(fmt.Sprintf("Getting information failure:%+v", err), c)
		return
	}

	response.OkWithDetailed(nil, "Success", c)
}
