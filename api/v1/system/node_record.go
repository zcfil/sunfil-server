package system

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
	"zcfil-server/config"
	"zcfil-server/contract"
	"zcfil-server/define"
	"zcfil-server/global"
	"zcfil-server/lotusrpc"
	"zcfil-server/model/common/response"
	modelSystem "zcfil-server/model/system"
	sysReq "zcfil-server/model/system/request"
	"zcfil-server/service/system"
	"zcfil-server/utils"
)

type NodeRecordApi struct {
}

type Res1 struct {
	ActorId  big.Int `json:"actorId"`
	Operator string  `json:"operator"`
	OwnerId  big.Int `json:"ownerId"`
}

// GetNodeRecordList List of node operation information obtained by classification
func (m *NodeRecordApi) GetNodeRecordList(c *gin.Context) {
	var req sysReq.GetNodeRecordReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if strings.Contains(req.NodeId, "t0") || strings.Contains(req.NodeId, "f0") {
		req.NodeId = req.NodeId[2:]
	}

	list, total, err := NodeRecordsService.GetNodeRecordList(req)
	if err != nil {
		global.ZC_LOG.Error("Getting information failure!", zap.Error(err))
		response.FailWithMessage("Getting information failure", c)
		return
	}

	location, _ := time.LoadLocation("Asia/Shanghai")
	timeNow, _ := time.ParseInLocation(config.TimeFormat, time.Now().Format(config.TimeFormat), location)

	for i := 0; i < len(list); i++ {
		updatedAt, _ := time.ParseInLocation(config.TimeFormat, list[i].UpdatedAt.Format(config.TimeFormat), location)
		list[i].TimeDuration = timeNow.Sub(updatedAt).String()

		list[i].ActorId = "f0" + utils.Strval(list[i].NodeId)

		if req.OpType != define.NodeRecordChangeStr {
			if req.OpType != define.NodeRecordWithdrawStr {
				list[i].Amount = list[i].OpContent
			} else {
				if strings.Contains(list[i].OpContent, ",") {
					contentArr := strings.Split(list[i].OpContent, ",")
					list[i].ToAddr = contentArr[0]
					list[i].Amount = contentArr[1]
				} else {
					list[i].Amount = list[i].OpContent
				}
			}
		}
	}

	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "Success", c)
}

// GetTestWithdraw List of node operation information obtained by classification
func (m *NodeRecordApi) GetTestWithdraw(c *gin.Context) {

	// 获取节点列表
	list, err := NodeInfoService.GetNodeList()
	if err != nil {
		global.ZC_LOG.Error("failed to obtain node list!", zap.Error(err))
		return
	}
	if len(list) == 0 {
		global.ZC_LOG.Error("no node data, skip to the next cycle", zap.Error(err))
		return
	}

	var sys sync.WaitGroup
	sys.Add(len(list))

	for _, val := range list {
		go func(nodeInfo modelSystem.SysNodeInfo) {
			defer sys.Done()
			UpdateNodeDebt(nodeInfo)
		}(val)
	}
}

// UpdateNodeDebt Obtain node debt information
func UpdateNodeDebt(nodeInfo modelSystem.SysNodeInfo) {
	var debtBalance big.Int
	param := contract.DebtGetBalanceOf.Keccak256()
	addr := common.HexToAddress(nodeInfo.NodeAddress)
	paramData := common.LeftPadBytes(addr.Bytes(), 32)
	param = append(param, paramData...)
	if err := contract.DebtContract.CallContract(param, &debtBalance); err != nil {
		fmt.Println("Failed to retrieve recorded loan pool amount from smart contract!", err)
		return
	}

	_debtBalance, _ := strconv.ParseFloat(debtBalance.String(), 64)
	_debtBalance = _debtBalance / math.Pow(10, 18)

	nodeInfo.DebtBalance = fmt.Sprintf("%.2f", _debtBalance)

	nodeAddr, err := address.NewFromString("f0" + strconv.FormatInt(int64(nodeInfo.NodeName), 10))
	if err != nil {
		fmt.Println("Node number processing failed!", err)
	} else {
		nodeInfo.Balance, err = GetMinerBalance(nodeAddr)
		if err != nil {
			fmt.Println("Failed to obtain node total amount!", err)
		}
	}

	nodeInfoService := system.NodeInfoService{}
	err = nodeInfoService.UpdateNodeInfo(&nodeInfo)
	if err != nil {
		log.Println("Node "+strconv.FormatInt(int64(nodeInfo.NodeName), 10)+", failed to record debt information and node total data!", err.Error())
		return
	}
}

// GetMinerBalance Obtain node quota
func GetMinerBalance(maddr address.Address) (float64, error) {
	var ctx = context.Background()
	mact, err := lotusrpc.FullApi.StateGetActor(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return 0, err
	}
	balance := mact.Balance.String()
	nodeBalance, _ := strconv.ParseFloat(balance, 64)
	return nodeBalance, nil
}
