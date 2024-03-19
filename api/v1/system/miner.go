package system

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/gin-gonic/gin"
	"log"
	"math/big"
	"zcfil-server/contract"
	"zcfil-server/model/common/response"
	"zcfil-server/model/system/request"
)

type MinerApi struct {
}

type Res struct {
	ActorId  big.Int `json:"actorId"`
	Operator string  `json:"operator"`
	OwnerId  big.Int `json:"ownerId"`
}

// PondMinerList Get the miner list
func (m *MinerApi) PondMinerList(c *gin.Context) {
	var res []Res
	if err := contract.MinerContract.CallContract(contract.MinerMethodGetMiners.Keccak256(), &res); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithDetailed(response.PageResult{
		List:     res,
		Total:    0,
		Page:     0,
		PageSize: 0,
	}, "Success", c)

}

// MinerInfo Obtain miner information
func (m *MinerApi) MinerInfo(c *gin.Context) {
	var actor request.ActorInfo
	if err := c.ShouldBindQuery(&actor); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	addr, err := address.NewFromString(actor.Actor)
	if err != nil {
		log.Println(err.Error())
		response.FailWithMessage("Unknown node number："+actor.Actor, c)
		return
	}
	res, err := nodeQueryService.HandleMinerInfo(addr)
	if err != nil {
		log.Println(err.Error(), addr.String())
		response.FailWithMessage("Failed to obtain node information："+addr.String(), c)
		return
	}
	response.OkWithData(res, c)
}

// WorkerInfo Get Worker Wallet Information
func (m *MinerApi) WorkerInfo(c *gin.Context) {
	var actor request.ActorInfo
	if err := c.ShouldBindQuery(&actor); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	addr, err := address.NewFromString(actor.Actor)
	if err != nil {
		log.Println(err.Error())
		response.FailWithMessage("Unknown node number："+actor.Actor, c)
		return
	}
	res, err := nodeQueryService.HandleWorkerInfo(addr)
	if err != nil {
		log.Println(err.Error(), addr.String())
		response.FailWithMessage("Failed to obtain node information："+addr.String(), c)
		return
	}
	response.OkWithData(res, c)
}

// PondOperator Get the current operator of the node
func (m *MinerApi) PondOperator(c *gin.Context) {
	var actor request.ActorInfo
	if err := c.ShouldBindQuery(&actor); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	addr, err := address.NewFromString(actor.Actor)
	if err != nil {
		log.Println(err.Error())
		response.FailWithMessage("Unknown node number："+actor.Actor, c)
		return
	}
	info, err := nodeQueryService.GetOperator(context.Background(), addr)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(info, c)

}

// DepartInfo Obtain resignation information
func (m *MinerApi) DepartInfo(c *gin.Context) {
	var actor request.ActorInfo
	if err := c.ShouldBindQuery(&actor); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	addr, err := address.NewFromString(actor.Actor)
	if err != nil {
		log.Println(err.Error())
		response.FailWithMessage("Unknown node number："+actor.Actor, c)
		return
	}
	res, err := nodeQueryService.DepartInfo(context.Background(), addr)
	if err != nil {
		log.Println(err.Error(), addr.String())
		response.FailWithMessage("Failed to obtain node information："+addr.String(), c)
		return
	}
	response.OkWithData(res, c)
}

// MsigInspect Obtain resignation information
func (m *MinerApi) MsigInspect(c *gin.Context) {
	var actor request.ActorInfo
	if err := c.ShouldBindQuery(&actor); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	addr, err := address.NewFromString(actor.Actor)
	if err != nil {
		log.Println(err.Error())
		response.FailWithMessage("Unknown node number："+actor.Actor, c)
		return
	}
	res, err := nodeQueryService.MsigInspect(context.Background(), addr)
	if err != nil {
		log.Println(err.Error(), addr.String())
		response.FailWithMessage("Failed to obtain node information："+addr.String(), c)
		return
	}
	response.OkWithData(res, c)
}
