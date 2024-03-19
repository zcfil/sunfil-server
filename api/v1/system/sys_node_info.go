package system

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/v0api"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	"github.com/gin-gonic/gin"
	cbor "github.com/ipfs/go-ipld-cbor"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
	"zcfil-server/contract"
	"zcfil-server/global"
	"zcfil-server/lotusrpc"
	"zcfil-server/model/common/response"
	modelSystem "zcfil-server/model/system"
	sysRequest "zcfil-server/model/system/request"
	sysResponse "zcfil-server/model/system/response"
	"zcfil-server/service/system"
	"zcfil-server/utils"
)

type NodeInfoApi struct{}

// GetNodeAddress Get node address
func (s *NodeInfoApi) GetNodeAddress(c *gin.Context) {
	var req sysRequest.NodeAddressReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	addr, err := address.NewFromString(req.NodeId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ethAddr, _, err := EthAddrFromFilecoinAddress(context.Background(), addr, lotusrpc.FullApi0)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(map[string]interface{}{"ethAddr": ethAddr.String()}, c)
}

// GetSysNodeList Paging to obtain node list
func (s *NodeInfoApi) GetSysNodeList(c *gin.Context) {
	var nodeReq sysRequest.NodeListReq
	err := c.ShouldBindQuery(&nodeReq)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if nodeReq.PageSize == 0 {
		nodeReq.PageSize = 10
	}
	if nodeReq.Page == 0 {
		nodeReq.Page = 1
	}

	list, total, err := NodeInfoService.GetSysNodeInfoList(nodeReq)
	if err != nil {
		global.ZC_LOG.Error("Getting information failure!", zap.Error(err))
		response.FailWithMessage("Getting information failure", c)
		return
	}

	baseRate := math.Pow(10, 18)
	for i := 0; i < len(list); i++ {
		list[i].NodeId = "f0" + strconv.FormatInt(list[i].NodeName, 10)
		list[i].NodeTotalBalance = utils.FloatAccurateBit(list[i].Balance/baseRate, utils.TwoBit)
		list[i].DebtValue = list[i].DebtBalance
		if list[i].NodeTotalBalance == 0 {
			list[i].DebtRatio = 0
		} else {
			list[i].DebtRatio = utils.FloatAccurateBit(list[i].DebtValue/list[i].NodeTotalBalance, utils.TwoBit)
		}
		list[i].NodeOwnershipBalance = utils.FloatAccurateBit(list[i].NodeTotalBalance-list[i].DebtValue, utils.TwoBit)
		if list[i].Operator == nodeReq.Operator {
			list[i].IsOperator = true
		}
	}

	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     nodeReq.Page,
		PageSize: nodeReq.PageSize,
	}, "Success", c)
}

// GetNodeRepayInfo Obtain node repayment information
func (s *NodeInfoApi) GetNodeRepayInfo(c *gin.Context) {
	var req sysRequest.NodeInfoReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	addr, err := address.NewFromString(req.NodeId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	eaddr, _, err := EthAddrFromFilecoinAddress(context.Background(), addr, lotusrpc.FullApi0)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	resp := sysResponse.NodeRepayInfo{}

	var outstandingDebt big.Int
	param := contract.DebtGetBalanceOf.Keccak256()
	debtAddr := common.HexToAddress(eaddr.String())
	paramData := common.LeftPadBytes(debtAddr.Bytes(), 32)
	param = append(param, paramData...)
	if err = contract.DebtContract.CallContract(param, &outstandingDebt); err != nil {
		fmt.Println("Failed to obtain outstanding debt from smart contract!", err)
		resp.OutstandingDebt = "0"
	} else {
		resp.OutstandingDebt = outstandingDebt.String()
	}

	resp.MaxLoanLimit = contract.RateContract.GetMaxBorrowableAmount(req.NodeId[2:]).String()

	response.OkWithData(resp, c)
}

// GetNodeWithdrawInfo Obtain node withdrawal information
func (s *NodeInfoApi) GetNodeWithdrawInfo(c *gin.Context) {
	var req sysRequest.NodeInfoReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	resp := sysResponse.NodeWithdrawalInfo{}

	var maxWithdrawLimit big.Int
	withdrawParam := contract.MaxWithdrawalAmount.Keccak256()
	nodeId, _ := new(big.Int).SetString(req.NodeId[2:], 10) //10
	paddedAmount := common.LeftPadBytes(nodeId.Bytes(), 32) //获取数量的编码
	param := append(withdrawParam, paddedAmount...)
	if err = contract.RateContract.CallContract(param, &maxWithdrawLimit); err != nil {
		fmt.Println("Failed to obtain the maximum withdrawal limit from the smart contract!", err)
		resp.MaxWithdrawBalance = "0"
	} else {
		resp.MaxWithdrawBalance = maxWithdrawLimit.String()
	}

	response.OkWithData(resp, c)
}

// GetNodeDataInfo Obtain node data information
func (s *NodeInfoApi) GetNodeDataInfo(c *gin.Context) {
	var req sysRequest.NodeInfoReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	resp, err := dealNodeInfo(c, req)
	if err != nil {
		return
	}

	response.OkWithData(resp, c)
}

// Processing node information
func dealNodeInfo(c *gin.Context, req sysRequest.NodeInfoReq) (sysResponse.NodeDetailInfo, error) {
	baseRate := math.Pow(10, 18)
	resp := sysResponse.NodeDetailInfo{}

	var nodeInfo modelSystem.SysNodeInfo
	var nodeName int64
	nodeInfoService := system.NodeInfoService{}
	if strings.Contains(req.NodeId, "f") || strings.Contains(req.NodeId, "t") {
		nodeName, _ = strconv.ParseInt(req.NodeId[2:], 10, 64)
	}
	nodeInfo, err := nodeInfoService.GetSysNodeInfo(uint64(nodeName))
	if err != nil {
		log.Println("error：GetSysNodeInfo：", err.Error())
	} else {
		if nodeInfo.Operator == req.Operator {
			resp.IsOperator = true
		}
	}

	addr, err := address.NewFromString(req.NodeId)
	if err != nil {
		log.Println(err.Error())
		response.FailWithMessage("Unknown node number："+req.NodeId, c)
		return sysResponse.NodeDetailInfo{}, err
	}

	mact, err := lotusrpc.FullApi.StateGetActor(context.TODO(), addr, types.EmptyTSK)
	if err != nil {
		fmt.Println("Failed to obtain the total balance of the node!", err)
	} else {
		_nodeTotalBalance, _ := strconv.ParseFloat(mact.Balance.String(), 64)
		resp.NodeTotalBalance = utils.FloatAccurateBit(_nodeTotalBalance/baseRate, utils.TwoBit)
	}

	if nodeAvailableBalance, err := liquidateService.GetAvailBalance(req.NodeId); err != nil {
		fmt.Println("Failed to obtain available balance of node!", err)
	} else {
		_nodeAvailableBalance, _ := strconv.ParseFloat(nodeAvailableBalance, 64)
		resp.NodeAvailableBalance = utils.FloatAccurateBit(_nodeAvailableBalance/baseRate, utils.TwoBit)
	}

	tbs := blockstore.NewTieredBstore(blockstore.NewAPIBlockstore(lotusrpc.FullApi), blockstore.NewMemory())
	mas, err := miner.Load(adt.WrapStore(context.TODO(), cbor.NewCborStore(tbs)), mact)
	if err != nil {
		fmt.Println("Deal blockstore tbs failed!", err)
	}
	lockedFunds, err := mas.LockedFunds()
	if err != nil {
		fmt.Println("Getting locked funds: %w", err)
	} else {
		_sectorPledge, _ := strconv.ParseFloat(lockedFunds.InitialPledgeRequirement.String(), 64)
		_preCommitDeposits, _ := strconv.ParseFloat(lockedFunds.PreCommitDeposits.String(), 64)
		resp.SectorPledge = utils.FloatAccurateBit((_sectorPledge+_preCommitDeposits)/baseRate, utils.TwoBit)
	}

	mb, err := lotusrpc.FullApi.StateMarketBalance(context.TODO(), addr, types.EmptyTSK)
	if err != nil {
		fmt.Println("Getting market balance: %w", err)
	} else {
		_locked, _ := strconv.ParseFloat(mb.Locked.String(), 64)
		resp.LockInRewards = utils.FloatAccurateBit(_locked/baseRate, utils.TwoBit)
	}

	if len(req.NodeDebt) != 0 {
		_debtNum, _ := strconv.ParseFloat(req.NodeDebt, 64)
		resp.DebtValue = utils.FloatAccurateBit(_debtNum/baseRate, utils.TwoBit)
	}

	if resp.NodeTotalBalance == 0 {
		resp.DebtRatio = 0
	} else {
		resp.DebtRatio = utils.FloatAccurateBit(resp.DebtValue/resp.NodeTotalBalance, utils.FourBit)
	}

	if len(req.MaxWithdrawLimit) != 0 {
		_maxWithdrawLimit, _ := strconv.ParseFloat(req.MaxWithdrawLimit, 64)
		resp.WithdrawalThreshold = utils.FloatAccurateBit(_maxWithdrawLimit/baseRate, utils.TwoBit)
	}

	resp.NodeOwnershipBalance = utils.FloatAccurateBit(resp.NodeTotalBalance-resp.DebtValue, utils.TwoBit)
	resp.MaxDebtRatio = nodeInfo.MaxDebtRate
	resp.SettlementThreshold = nodeInfo.LiquidateRate

	return resp, nil
}

func EthAddrFromFilecoinAddress(ctx context.Context, addr address.Address, fnapi v0api.FullNode) (ethtypes.EthAddress, address.Address, error) {
	var faddr address.Address
	var err error

	switch addr.Protocol() {
	case address.BLS, address.SECP256K1:
		faddr, err = fnapi.StateLookupID(ctx, addr, types.EmptyTSK)
		if err != nil {
			return ethtypes.EthAddress{}, addr, err
		}
	case address.Actor, address.ID:
		faddr, err = fnapi.StateLookupID(ctx, addr, types.EmptyTSK)
		if err != nil {
			return ethtypes.EthAddress{}, addr, err
		}
		fAct, err := fnapi.StateGetActor(ctx, faddr, types.EmptyTSK)
		if err != nil {
			return ethtypes.EthAddress{}, addr, err
		}
		if fAct.Address != nil && (*fAct.Address).Protocol() == address.Delegated {
			faddr = *fAct.Address
		}
	case address.Delegated:
		faddr = addr
	default:
		return ethtypes.EthAddress{}, addr, xerrors.Errorf("Filecoin address doesn't match known protocols")
	}

	ethAddr, err := ethtypes.EthAddressFromFilecoinAddress(faddr)
	if err != nil {
		return ethtypes.EthAddress{}, addr, err
	}

	return ethAddr, faddr, nil
}

// GetAvailableAssets Obtain available current assets
func (s *NodeInfoApi) GetAvailableAssets(c *gin.Context) {
	var availableAssets float64
	var lastAvailableBalance big.Int
	if err := contract.StakeContract.CallContract(contract.GetContractAmount.Keccak256(), &lastAvailableBalance); err != nil {
		fmt.Println("Failed to obtain remaining available current asset information from smart contract!", err)
		response.FailWithMessage("Get available assets failed", c)
		return
	}

	_lastAvailableBalance, _ := strconv.ParseFloat(lastAvailableBalance.String(), 64)
	availableAssets = utils.FloatAccurateBit(_lastAvailableBalance/math.Pow(10, 18), utils.TwoBit)
	response.OkWithData(map[string]interface{}{"availableAssets": availableAssets}, c)
}

// Processing node list information
func dealNodeListInfo(c *gin.Context, req sysRequest.NodeInfoReq) (sysResponse.NodeDetailInfo, error) {
	baseRate := math.Pow(10, 18)
	resp := sysResponse.NodeDetailInfo{}

	addr, err := address.NewFromString(req.NodeId)
	if err != nil {
		log.Println(err.Error())
		response.FailWithMessage("Unknown node number："+req.NodeId, c)
		return sysResponse.NodeDetailInfo{}, err
	}

	mact, err := lotusrpc.FullApi.StateGetActor(context.TODO(), addr, types.EmptyTSK)
	if err != nil {
		fmt.Println("Failed to obtain the total balance of the node!", err)
	} else {
		_nodeTotalBalance, _ := strconv.ParseFloat(mact.Balance.String(), 64)
		resp.NodeTotalBalance = utils.FloatAccurateBit(_nodeTotalBalance/baseRate, utils.TwoBit)
	}

	eaddr, _, err := EthAddrFromFilecoinAddress(context.Background(), addr, lotusrpc.FullApi0)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return resp, err
	}
	if debtNum, err := liquidateService.GetDebt(eaddr.String()); err != nil {
		fmt.Println("Failed to obtain available balance of node!", err)
	} else {
		_debtNum, _ := strconv.ParseFloat(debtNum, 64)
		resp.DebtValue = utils.FloatAccurateBit(_debtNum/baseRate, utils.TwoBit)
	}

	if resp.NodeTotalBalance == 0 {
		resp.DebtRatio = 0
	} else {
		resp.DebtRatio = utils.FloatAccurateBit(resp.DebtValue/resp.NodeTotalBalance, utils.TwoBit)
	}

	resp.NodeOwnershipBalance = utils.FloatAccurateBit(resp.NodeTotalBalance-resp.DebtValue, utils.TwoBit)

	var debtRate, liquidateRate, warnPeriod, votePeriod big.Int
	debtRateParam := contract.RateGetRateContractParam.Keccak256()
	target, _ := new(big.Int).SetString(req.NodeId[2:], 10)
	targetData := common.LeftPadBytes(target.Bytes(), 32)
	debtRateParam = append(debtRateParam, targetData...)
	if err := contract.RateContract.CallContract(debtRateParam, &debtRate, &liquidateRate, &warnPeriod, &votePeriod); err != nil {
		fmt.Println("Abnormal acquisition of interest rate parameters. CallContract err:", err)
	}

	maxDebtRatio, _ := strconv.ParseFloat(debtRate.String(), 64)
	resp.MaxDebtRatio = utils.FloatAccurateBit(maxDebtRatio/baseRate, utils.FourBit)

	threshold, _ := strconv.ParseFloat(liquidateRate.String(), 64)
	resp.SettlementThreshold = utils.FloatAccurateBit(threshold/baseRate, utils.FourBit)
	return resp, nil
}
