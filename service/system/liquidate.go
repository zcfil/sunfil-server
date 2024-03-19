package system

import (
	"context"
	"errors"
	"fmt"
	c_abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-state-types/abi"
	filecoin_big "github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	miner2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/miner"
	cbor "github.com/ipfs/go-ipld-cbor"
	uuidGo "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
	"zcfil-server/build"
	"zcfil-server/contract"
	"zcfil-server/define"
	"zcfil-server/global"
	"zcfil-server/lotusrpc"
	"zcfil-server/model/system"
	"zcfil-server/model/system/request"
	"zcfil-server/model/system/response"
	"zcfil-server/utils"
)

var LiquidateServiceApi = new(LiquidateService)

type LiquidateService struct {
}

// GetContractWarnNode Obtain node alarm nodes
func (l *LiquidateService) GetContractWarnNode() error {

	// Query existing data
	dataMap := make(map[string]response.OldWarnNodeInfo)
	var cwn []response.OldWarnNodeInfo
	sql := ` SELECT contract_address,agree_count,reject_count,abstain_count,total_vote,solidify,modify_control,
                    c.warn_begin_time,c.warn_end_time,c.vote_begin_time,c.vote_end_time,c.node_status
               FROM sys_contract_warn_node c ORDER BY c.created_at DESC LIMIT 1000`
	err := global.ZC_DB.Raw(sql).Scan(&cwn).Error
	if err != nil {
		return err
	}

	addrStr := ""
	for _, v := range cwn {

		dataMap[v.ContractAddress] = v
		if v.NodeStatus != define.VoteEnd && v.NodeStatus != define.VoteClose {
			if addrStr == "" {
				addrStr = `'` + v.ContractAddress + `'`
			} else {
				addrStr += `,'` + v.ContractAddress + `'`
			}
		}
	}

	// Alarm node data
	type WarnNode struct {
		MinerId      uint64
		Addr         common.Address
		TotalBalance big.Int
		TotalDebt    big.Int
		WarnTime     big.Int
		AgreeCount   big.Int
		RejectCount  big.Int
		AbstainCount big.Int
		TotalVote    big.Int
	}

	var wn []WarnNode
	if err := contract.VoteContract.CallContract(contract.VoteCheckWarnNode.Keccak256(), &wn); err != nil {
		global.ZC_LOG.Error("Failed to obtain alarm node:", zap.Error(err))
		return err
	}

	wnCount := 0
	for _, v := range wn {
		if v.MinerId > 0 {
			wnCount++
		}
	}

	// Clear alarm data, delete voting proposals and ticket collectors
	if wnCount == 0 {
		if addrStr != "" {
			err = l.DelWarnVote(addrStr)
			if err != nil {
				return err
			}
		}

	} else {
		// Delete first, then modify
		err := global.ZC_DB.Model(&system.SysContractWarnNode{}).Where("node_status not in(3,4) and deleted_at is null").Update("deleted_at", time.Now()).Error
		if err != nil {
			return err
		}

		// Voting nodes
		var voteWn []WarnNode
		if err := contract.VoteContract.CallContract(contract.VoteGetWarnNode.Keccak256(), &voteWn); err != nil {
			global.ZC_LOG.Error("Failed to obtain voting proposal:", zap.Error(err))
			return err
		}

		voteWnMap := make(map[string]WarnNode)
		for _, v := range voteWn {
			voteWnMap[v.Addr.String()] = v
		}

		addrAr := ""
		warnNodeAr := make([]system.SysContractWarnNode, 0)
		for _, v := range wn {

			if v.MinerId == 0 {
				continue
			}

			// interest rate
			var debtRate, liquidateRate, warnPeriod, votePeriod big.Int
			param := contract.RateGetRateContractParam.Keccak256()
			target, b := new(big.Int).SetString(strconv.Itoa(int(v.MinerId)), 10)
			if !b {
				return errors.New("minerId conversion failed ")
			}
			targetData := common.LeftPadBytes(target.Bytes(), 32)
			param = append(param, targetData...)

			if err := contract.RateContract.CallContract(param, &debtRate, &liquidateRate, &warnPeriod, &votePeriod); err != nil {
				global.ZC_LOG.Error("Failed to obtain interest rate parameters:", zap.Error(err))
				return err
			}

			totalBalance := "0"
			maddr, err := address.NewFromString(fmt.Sprintf("f0%d", v.MinerId))
			if err != nil {
				global.ZC_LOG.Error("NewFromString:", zap.Error(err))
				return err
			}

			mact, err := lotusrpc.FullApi.StateGetActor(context.TODO(), maddr, types.EmptyTSK)
			if err != nil {
				global.ZC_LOG.Error("StateGetActor:", zap.Error(err))
				return err
			} else {
				totalBalance = mact.Balance.String()
			}

			var withdrawAmount big.Int
			if mact.Balance.GreaterThan(types.NewInt(0)) {

				param = contract.RateWithdrawalLimitView.Keccak256()
				target, _ = new(big.Int).SetString(strconv.Itoa(int(v.MinerId)), 10)
				balance, _ := new(big.Int).SetString(mact.Balance.String(), 10)
				targetData = common.LeftPadBytes(target.Bytes(), 32)
				availableBalanceData := common.LeftPadBytes(balance.Bytes(), 32)
				param = append(param, targetData...)
				param = append(param, availableBalanceData...)

				if err := contract.RateContract.CallContract(param, &withdrawAmount); err != nil {
					global.ZC_LOG.Error("Failed to obtain withdrawal limit:", zap.Error(err))
					return err
				}
			}

			withdraw := utils.StrToFloat64(withdrawAmount.String(), utils.FourBit)
			totalBalances := utils.StrToFloat64(totalBalance, utils.FourBit)
			totalDebts := utils.StrToFloat64(v.TotalDebt.String(), utils.FourBit)
			warnTimes, _ := strconv.Atoi(v.WarnTime.String())
			availableBalance := totalBalances - totalDebts
			var nodeDebtRate float64
			if totalBalances > 0 {
				nodeDebtRate = totalDebts / totalBalances
			}

			if availableBalance <= 0 {
				availableBalance = 0
			}

			warnBeginTime := utils.UnixToTime(warnTimes)
			warnEndTime := utils.UnixToTime(warnTimes).AddDate(0, 0, int(warnPeriod.Int64()))
			voteEndTime := warnEndTime.AddDate(0, 0, int(votePeriod.Int64()))

			scw := &system.SysContractWarnNode{
				ContractAddress: v.Addr.String(),
				MinerId:         fmt.Sprintf("f0%d", v.MinerId),
				NodeBalance:     totalBalances,
				NodeDebt:        totalDebts,
				NodeAvailable:   utils.FloatAccurateBit(availableBalance, utils.FourBit),
				DebtRate:        utils.StrToFloat64(debtRate.String(), utils.TwoBit) * 100,
				LiquidateRate:   utils.StrToFloat64(liquidateRate.String(), utils.TwoBit) * 100,
				NodeDebtRate:    utils.FloatAccurateBit(nodeDebtRate*100, utils.TwoBit),
				WarnBeginTime:   warnBeginTime,
				WarnEndTime:     warnEndTime,
				VoteBeginTime:   warnEndTime,
				VoteEndTime:     voteEndTime,
				VoteStatus:      define.VoteTNotInEffect,
				NodeStatus:      define.VoteWarn,
				Withdraw:        withdraw,
			}

			warnNodeAr = append(warnNodeAr, *scw)
			if addrAr == "" {
				addrAr = `'` + v.Addr.String() + `'`
			} else {
				addrAr += `,'` + v.Addr.String() + `'`
			}
		}

		if addrAr != "" {

			for _, v := range warnNodeAr {

				st, ok := dataMap[v.ContractAddress]
				if ok {

					if st.NodeStatus == define.VoteEnd {
						continue
					}

					// Obtain vote counting
					if st0, ok1 := voteWnMap[v.ContractAddress]; ok1 {
						v.AgreeCount = utils.StrToFloat64(st0.AgreeCount.String(), utils.FourBit)
						v.RejectCount = utils.StrToFloat64(st0.RejectCount.String(), utils.FourBit)
						v.AbstainCount = utils.StrToFloat64(st0.AbstainCount.String(), utils.FourBit)
						v.TotalVote = utils.StrToFloat64(st0.TotalVote.String(), utils.FourBit)
					}

					v.WarnBeginTime = st.WarnBeginTime
					v.WarnEndTime = st.WarnEndTime
					v.VoteBeginTime = st.VoteBeginTime
					v.VoteEndTime = st.VoteEndTime
					v.Solidify = st.Solidify
					v.NodeStatus = st.NodeStatus

					// Node status 1 in alarm, 2 in voting, 3 ended
					if time.Now().After(v.WarnEndTime) && v.NodeStatus == define.VoteWarn {
						// Fixed voters, proposals, and voting weights
						if v.Solidify == 0 {
							param := contract.VoteSolidifiedVote.Keccak256()
							addr := common.HexToAddress(v.ContractAddress)
							addrData := common.LeftPadBytes(addr.Bytes(), 32)
							param = append(param, addrData...)
							global.ZC_LOG.Info("GetContractWarnNode fixed voters")
							_, err := contract.VoteContract.PushContract(param)
							if err != nil {
								global.ZC_LOG.Error("Fixed voter, proposal, failed voting weight:", zap.Error(err))
								return err
							}

							v.Solidify = 1
						}
						v.NodeStatus = define.VoteVoting

					} else if time.Now().After(v.VoteEndTime) && v.NodeStatus == define.VoteVoting {

						// Proposal status 1 is effective, 2 is not effective
						if ((v.AgreeCount + v.RejectCount + v.AbstainCount) / v.TotalVote) > define.TakeEffectRate {
							v.VoteStatus = define.VoteTakeEffect
						} else {
							v.VoteStatus = define.VoteTNotInEffect
						}

						// Voting results
						resAr := make([]int, 0)
						resAr = append(resAr, int(v.AgreeCount), int(v.RejectCount), int(v.AbstainCount))
						sort.Ints(resAr)

						switch resAr[len(resAr)-1] {
						case 0:
							v.VoteResults = define.VoteExpired
						case int(v.AgreeCount):
							v.VoteResults = define.VoteAgree
							break
						case int(v.RejectCount):
							v.VoteResults = define.VoteReject
							break
						case int(v.AbstainCount):
							v.VoteResults = define.VoteAbstain
							break
						}

						v.NodeStatus = define.VoteEnd
						// Delete Contract Proposal Map
						param := contract.VoteDelwarnNodeMap.Keccak256()
						addr := common.HexToAddress(v.ContractAddress)
						addrData := common.LeftPadBytes(addr.Bytes(), 32)
						param = append(param, addrData...)
						global.ZC_LOG.Info("GetContractWarnNode delete Contract Proposal Map")
						_, err := contract.VoteContract.PushContract(param)
						if err != nil {
							global.ZC_LOG.Error("Failed to delete contract proposal Map:", zap.Error(err))
							return err
						}
					}

					sql := ` update sys_contract_warn_node set abstain_count=?,agree_count=?,debt_rate=?,deleted_at=?,liquidate_rate=?,node_available=?,node_balance=?,
															node_debt=?,node_debt_rate=?,node_status=?,reject_count=?,total_vote=?,vote_results=?,vote_status=?,updated_at=?,
															solidify = ?,withdraw = ?,process=0
													  where contract_address = ? `
					var paramReq []interface{}
					paramReq = append(paramReq, v.AbstainCount, v.AgreeCount, v.DebtRate, nil, v.LiquidateRate, v.NodeAvailable, v.NodeBalance)
					paramReq = append(paramReq, v.NodeDebt, v.NodeDebtRate, v.NodeStatus, v.RejectCount, v.TotalVote, v.VoteResults, v.VoteStatus, time.Now(), v.Solidify, v.Withdraw)
					paramReq = append(paramReq, v.ContractAddress)

					err = global.ZC_DB.Exec(sql, paramReq...).Error

				} else {

					err = global.ZC_DB.Model(&system.SysContractWarnNode{}).Create(&v).Error
				}

				if err != nil {
					return err
				}
			}
		}

		global.ZC_LOG.Info("Successfully obtained alarm node data:" + utils.GetNowStr())

	}

	return nil

}

// ReContractWarnNode Retrieve node alarm nodes again
func (l *LiquidateService) ReContractWarnNode() error {
	go func() {
		time.Sleep(time.Second * 30)
		err := l.GetContractWarnNode()
		if err != nil {
			global.ZC_LOG.Error("GetContractWarnNode:", zap.Error(err))
			return
		}
	}()

	return nil
}

// GetWarnNodeList Alarm node information
func (l *LiquidateService) GetWarnNodeList(info request.WarnNodeReq) (list []response.WarnNodeInfo, total int64, err error) {

	if info.NodeStatus > 5 {
		return nil, total, errors.New("NodeStatus data is error")
	}
	if info.PageSize == 0 || info.Page == 0 {
		info.Page = 1
		info.PageSize = 10
	}

	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)

	db := global.ZC_DB.Model(&system.SysContractWarnNode{})
	var contractNode []system.SysContractWarnNode
	sql := ""
	var param []interface{}
	if info.NodeStatus > 0 {

		switch info.NodeStatus {
		case 2:
			sql += " node_status in(2,3)"
			break
		case 1:
			sql += " node_status =1"
			break
		case 3:
			sql += " node_status =4"
			break
		}
	}
	err = db.Where(sql, param...).Count(&total).Error
	if err != nil {
		return nil, total, err
	}
	err = db.Where(sql, param...).Limit(limit).Offset(offset).Order("id desc").Find(&contractNode).Error
	if err != nil {
		return nil, total, err
	}

	var wf []response.WarnNodeInfo

	for _, v := range contractNode {

		riskDuration := time.Now().Sub(v.WarnBeginTime).String()
		riskDuration = strings.Split(riskDuration, "m")[0] + "m"
		voteExpired := v.VoteEndTime.Sub(time.Now()).String()
		voteExpired = strings.Split(voteExpired, "m")[0] + "m"

		wfParam := &response.WarnNodeInfo{
			Id:              v.ID,
			MinerId:         v.MinerId,
			ContractAddress: v.ContractAddress,
			NodeBalance:     fmt.Sprintf("%0.4f FIL", v.NodeBalance),
			NodeDebt:        fmt.Sprintf("%0.4f FIL", v.NodeDebt),
			NodeAvailable:   fmt.Sprintf("%0.4f FIL", v.NodeAvailable),
			DebtRate:        fmt.Sprintf("%0.2f", v.DebtRate) + "%",
			LiquidateRate:   fmt.Sprintf("%0.2f", v.LiquidateRate) + "%",
			NodeDebtRate:    fmt.Sprintf("%0.2f", v.NodeDebtRate) + "%",
			WarnBeginTime:   v.WarnBeginTime,
			VoteEndTime:     v.VoteEndTime,
			RiskDuration:    riskDuration,
			DiffRate:        fmt.Sprintf("%0.2f", utils.FloatAccurateBit(v.NodeDebtRate-v.LiquidateRate, utils.TwoBit)) + "%",
			VoteExpired:     voteExpired,
		}

		var voteAr []int
		voteAr = append(voteAr, int(v.AgreeCount), int(v.RejectCount), int(v.AbstainCount))
		sort.Ints(voteAr)

		switch voteAr[len(voteAr)-1] {
		case 0:
			wfParam.VoteResults = define.VoteExpired
			wfParam.ResultsRate = "0%"
			break
		case int(v.AgreeCount):
			wfParam.VoteResults = define.VoteAgree
			wfParam.ResultsRate = fmt.Sprintf("%0.2f", utils.FloatAccurateBit(v.AgreeCount/v.TotalVote*100, utils.TwoBit)) + "%"
			break
		case int(v.RejectCount):
			wfParam.VoteResults = define.VoteReject
			wfParam.ResultsRate = fmt.Sprintf("%0.2f", utils.FloatAccurateBit(v.RejectCount/v.TotalVote*100, utils.TwoBit)) + "%"
			break
		case int(v.AbstainCount):
			wfParam.VoteResults = define.VoteAbstain
			wfParam.ResultsRate = fmt.Sprintf("%0.2f", utils.FloatAccurateBit(v.AbstainCount/v.TotalVote*100, utils.TwoBit)) + "%"
			break
		}

		fmt.Println(fmt.Sprintf("data:%+v", wfParam))
		wf = append(wf, *wfParam)
	}

	return wf, total, err

}

// WarnNodeDetail Alarm node details
func (l *LiquidateService) WarnNodeDetail(info request.WarnNodeReq) (map[string]interface{}, error) {

	resMap := make(map[string]interface{})
	var contractNode system.SysContractWarnNode
	if info.Id <= 0 {
		return resMap, errors.New("id is null")
	}
	err := global.ZC_DB.Model(&system.SysContractWarnNode{}).Where("id", info.Id).Find(&contractNode).Error
	if err != nil {
		return resMap, err
	}

	if (contractNode == system.SysContractWarnNode{}) {
		return resMap, xerrors.New("contractNode data is null")
	}

	addr, err := address.NewFromString(contractNode.MinerId)
	if err != nil {
		return nil, xerrors.Errorf("NewFromString err:", err)
	}

	res, err := NodeQueryServiceApi.HandleMinerInfo(addr)
	if err != nil {
		return nil, xerrors.Errorf("HandleMinerInfo err:", err)
	}

	node := &response.NodeInfo{
		Available: fmt.Sprintf("%0.4f FIL", utils.StrToFloat64(res.Available.String(), utils.FourBit)),
		Vesting:   res.Vesting,
		Pledge:    res.Pledge,
		Active:    res.Active,
		Faulty:    res.Faulty,
		Live:      res.Live,
	}

	head, err := lotusrpc.FullApi.ChainHead(context.TODO())
	if err != nil {
		return nil, xerrors.Errorf("ChainHead err:", err)
	}

	// Normal sector
	activeSet, err := lotusrpc.FullApi.StateMinerActiveSectors(context.TODO(), addr, head.Key())
	if err != nil {
		return nil, xerrors.Errorf("StateMinerActiveSectors err:", err)
	}

	// Abnormal sector
	faultySectors, err := l.GetFaultySectors(contractNode.MinerId)
	if err != nil {
		return nil, xerrors.Errorf("StateMinerActiveSectors err:", err)
	}

	var sectors []string
	for _, v := range faultySectors {
		sectors = append(sectors, strconv.Itoa(int(v)))
	}

	for _, v := range activeSet {
		sectors = append(sectors, v.SectorNumber.String())
	}

	// Expected termination sector return fil
	var stopTotalFil float64
	if len(sectors) > 0 {
		stopTotalFil, err = l.GetStopSectorsValue(contractNode.MinerId, sectors)
		if err != nil {
			fmt.Println(fmt.Sprintf("GetStopSectorsValue err:%s", err))
		}
	}

	var wf response.WarnNodeInfo
	wf.MinerId = contractNode.MinerId
	wf.NodeBalance = fmt.Sprintf("%0.4f FIL", contractNode.NodeBalance)
	wf.NodeDebt = fmt.Sprintf("%0.4f FIL", contractNode.NodeDebt)
	wf.NodeAvailable = fmt.Sprintf("%0.4f FIL", contractNode.NodeAvailable)
	wf.DebtRate = fmt.Sprintf("%0.2f", contractNode.DebtRate)
	wf.LiquidateRate = fmt.Sprintf("%0.2f", contractNode.LiquidateRate) + "%"
	wf.NodeDebtRate = fmt.Sprintf("%0.2f", contractNode.NodeDebtRate) + "%"
	riskAr := strings.Split(time.Now().Sub(contractNode.WarnBeginTime).String(), "m")
	wf.RiskDuration = riskAr[0] + "m"
	wf.VoteBeginTime = contractNode.VoteBeginTime
	wf.VoteEndTime = contractNode.VoteEndTime
	wf.DiffRate = fmt.Sprintf("%0.2f", contractNode.NodeDebtRate-contractNode.LiquidateRate) + "%"
	wf.Withdraw = fmt.Sprintf("%0.4f FIL", contractNode.Withdraw)
	wf.AgreeCount = fmt.Sprintf("%0.4f sunFIL", contractNode.AgreeCount)
	if contractNode.AgreeCount == 0 || contractNode.TotalVote == 0 {
		wf.AgreeRate = "0%"
	} else {
		wf.AgreeRate = fmt.Sprintf("%0.2f", utils.FloatAccurateBit(contractNode.AgreeCount/contractNode.TotalVote, utils.TwoBit)*100) + "%"
	}
	wf.RejectCount = fmt.Sprintf("%0.4f sunFIL", contractNode.RejectCount)
	if contractNode.RejectCount == 0 || contractNode.TotalVote == 0 {
		wf.RejectRate = "0%"
	} else {
		wf.RejectRate = fmt.Sprintf("%0.2f", utils.FloatAccurateBit(contractNode.RejectCount/contractNode.TotalVote, utils.TwoBit)*100) + "%"
	}
	wf.AbstainCount = fmt.Sprintf("%0.4f sunFIL", contractNode.AbstainCount)
	if contractNode.AbstainCount == 0 || contractNode.TotalVote == 0 {
		wf.AbstainRate = "0%"
	} else {
		wf.AbstainRate = fmt.Sprintf("%0.2f", utils.FloatAccurateBit(contractNode.AbstainCount/contractNode.TotalVote, utils.TwoBit)*100) + "%"
	}
	if (contractNode.AgreeCount+contractNode.RejectCount+contractNode.AbstainCount) == 0 || contractNode.TotalVote == 0 {
		wf.VoteRate = "0%"
	} else {
		wf.VoteRate = fmt.Sprintf("%0.2f", utils.FloatAccurateBit((contractNode.AgreeCount+contractNode.RejectCount+contractNode.AbstainCount)/contractNode.TotalVote*100, utils.TwoBit)) + "%"
	}

	var voteAr []int
	voteAr = append(voteAr, int(contractNode.AgreeCount), int(contractNode.RejectCount), int(contractNode.AbstainCount))
	sort.Ints(voteAr)

	// 1 agree, 2 oppose, 3 abstain, 4 expire
	switch voteAr[len(voteAr)-1] {
	case 0:
		wf.VoteResults = define.VoteExpired
		wf.ResultsRate = "0%"
		break
	case int(contractNode.AgreeCount):
		wf.VoteResults = define.VoteAgree
		wf.ResultsRate = fmt.Sprintf("%0.2f", utils.FloatAccurateBit(contractNode.AgreeCount/contractNode.TotalVote, utils.TwoBit)) + "%"
		break
	case int(contractNode.RejectCount):
		wf.VoteResults = define.VoteReject
		wf.ResultsRate = fmt.Sprintf("%0.2f", utils.FloatAccurateBit(contractNode.RejectCount/contractNode.TotalVote, utils.TwoBit)) + "%"
		break
	case int(contractNode.AbstainCount):
		wf.VoteResults = define.VoteAbstain
		wf.ResultsRate = fmt.Sprintf("%0.2f", utils.FloatAccurateBit(contractNode.AbstainCount/contractNode.TotalVote, utils.TwoBit)) + "%"
		break
	}

	if stopTotalFil == 0 {
		node.TerminateFile = "0 FIL"
	} else {
		f1 := utils.StrToFloat64(fmt.Sprintf("%0.f", stopTotalFil), utils.FourBit)
		node.TerminateFile = fmt.Sprintf("%0.f File", f1)
	}

	resMap["contractNode"] = wf
	resMap["minerInfo"] = node

	return resMap, nil

}

// Liquidation Liquidation
func (l *LiquidateService) Liquidation() error {
	var swn []system.SysContractWarnNode
	err := global.ZC_DB.Model(&system.SysContractWarnNode{}).Where("process=? and node_status=?", define.VoteNotProcess, define.VoteEnd).Find(&swn).Error
	if err != nil {
		return err
	}

	for _, v := range swn {
		// Check if there is any outstanding debt
		debtStr, err := l.GetDebt(v.ContractAddress)
		if err != nil {
			return err
		}

		ds, _ := strconv.ParseFloat(debtStr, 64)
		debt := utils.FloatAccurateBit(ds, utils.FourBit)
		fmt.Println(fmt.Sprintf("Node:%s,arrears:%f", v.MinerId, debt))
		if debt > 0 {
			// Available balance
			availBalance, err := l.GetAvailBalance(v.MinerId)
			if err != nil {
				return xerrors.Errorf("Getting available balance: %w", err)
			}

			ab, _ := strconv.ParseFloat(availBalance, 64)
			nodeAvailBalance := utils.FloatAccurateBit(ab, utils.FourBit)

			fmt.Println(fmt.Sprintf("Node:%s,available balance:%f", v.MinerId, nodeAvailBalance))
			// The voting proposal takes effect and is passed
			if v.VoteResults == define.VoteAgree && v.VoteStatus == define.VoteTakeEffect {
				// Repayment first, if the balance is insufficient, terminate the equal value sector before repayment
				amountStr := ""
				fullRepayment := false
				if nodeAvailBalance >= debt {
					amountStr = debtStr
					fullRepayment = true
				} else {
					amountStr = availBalance
				}

				if ab > 100000000000000000 {
					for {
						err = l.Repayment(v.MinerId, amountStr, fullRepayment, define.DecisionLiquidation)
						if err != nil {
							time.Sleep(time.Second * 3)
							fmt.Println("Repayment err:", err)
							continue
						}

						fmt.Println(fmt.Sprintf("Node:%s,repayment successful,repayment amount:%s", v.MinerId, amountStr))
						break
					}
				}

				// Check again if there is still any outstanding debt
				debtStr, err = l.GetDebt(v.ContractAddress)
				if err != nil {
					return err
				}

				ds, _ = strconv.ParseFloat(debtStr, 64)
				fmt.Println(fmt.Sprintf("Node:%s,remaining debt:%f", v.MinerId, ds))

				// Clearing sector
				if ds > 0 {
					// Change control address
					if v.ModifyControl == define.UnChange {

						err = l.ModifyControlAddr(v.MinerId)
						if err != nil {
							return err
						}

						err = global.ZC_DB.Model(&system.SysContractWarnNode{}).Where("id", v.ID).Update("modify_control", define.Change).Error
						if err != nil {
							return err
						}
					}

					pledgeAmount, stopSector, err := l.GetStopSectorsTotal(v.MinerId, debtStr)
					if err != nil {
						return err
					}

					// Terminate sector
					if len(stopSector) > 0 {
						err = l.StopSectors(v.MinerId, stopSector)
						if err != nil {
							return err
						} else {
							fmt.Println(fmt.Sprintf("Node:%s,terminate sector completion", v.MinerId))
							// Sector termination log
							sectorsStr := ""
							for _, v := range stopSector {
								if sectorsStr == "" {
									sectorsStr = v
								} else {
									sectorsStr += "," + v
								}
							}

							paramData := &system.SysSectorRecords{
								MinerId:      v.MinerId,
								Sectors:      sectorsStr,
								SectorCount:  len(stopSector),
								PledgeAmount: fmt.Sprintf("%f", pledgeAmount),
							}
							err = global.ZC_DB.Create(paramData).Error
							if err != nil {
								return err
							}
						}

						// Available balance
						availBalance, err := l.GetAvailBalance(v.MinerId)
						if err != nil {
							return xerrors.Errorf("Getting available balance: %w", err)
						}

						ab, _ := strconv.ParseFloat(availBalance, 64)
						if ab > 0 {
							// Repayment
							for {
								err = l.Repayment(v.MinerId, availBalance, false, define.DecisionLiquidation)
								if err != nil {
									time.Sleep(time.Second * 3)
									fmt.Println("Repayment err:", err)
									continue
								}

								fmt.Println(fmt.Sprintf("Node:%s,successful second repayment,repayment amount:%s", v.MinerId, availBalance))
								break
							}
						}
					}
				}
				err = global.ZC_DB.Model(&system.SysContractWarnNode{}).Where("id = ?", v.ID).Update("Process", define.VoteProcess).Error
				if err != nil {
					return err
				}

				// Objection, Waiver, Expiration, Proposal Not Effective
			} else if v.VoteResults == define.VoteReject || v.VoteResults == define.VoteAbstain || v.VoteResults == define.VoteExpired || v.VoteStatus == define.VoteTNotInEffect {
				// Interest rate difference+1, making debt ratio<liquidation rate
				diffRate := v.NodeDebtRate - v.LiquidateRate
				diffRate += 1
				debtStr = fmt.Sprintf("%f", nodeAvailBalance*diffRate/100)
				debtStr = strings.Split(debtStr, ".")[0]

				// repayment
				if ab > 0 {
					for {
						err = l.Repayment(v.MinerId, debtStr, false, define.ProtectLiquidation)
						if err != nil {
							time.Sleep(time.Second * 3)
							fmt.Println("Repayment err:", err)
							continue
						}

						fmt.Println(fmt.Sprintf("Protection liquidation,node:%s,repayment successful,repayment amount:%s", v.MinerId, debtStr))
						break
					}
				}

				err := global.ZC_DB.Model(&system.SysContractWarnNode{}).Where("id = ?", v.ID).Update("node_status", define.VoteClose).Error
				if err != nil {
					return err
				}
			}
		} else {
			err := global.ZC_DB.Model(&system.SysContractWarnNode{}).Where("id = ?", v.ID).Update("node_status", define.VoteClose).Error
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// TimingRepayment Timed deduction
func (l *LiquidateService) TimingRepayment() error {
	var swn []system.SysContractWarnNode
	err := global.ZC_DB.Model(&system.SysContractWarnNode{}).Where("process=? and node_status=?", define.VoteProcess, define.VoteEnd).Find(&swn).Error
	if err != nil {
		return err
	}

	for _, v := range swn {
		// Check if there is any outstanding debt
		debtStr, err := l.GetDebt(v.ContractAddress)
		if err != nil {
			return err
		}

		ds, _ := strconv.ParseFloat(debtStr, 64)
		debt := utils.FloatAccurateBit(ds, utils.FourBit)

		availBalance, err := l.GetAvailBalance(v.MinerId)
		if err != nil {
			return xerrors.Errorf("Getting available balance: %w", err)
		}

		ab, _ := strconv.ParseFloat(availBalance, 64)
		nodeAvailBalance := utils.FloatAccurateBit(ab, utils.FourBit)

		if debt > 0 && nodeAvailBalance > 0 {
			// Less than 0.5FIL, not deducted
			if nodeAvailBalance < 500000000000000000 {
				continue
			}

			amountStr := ""
			fullRepayment := false
			if nodeAvailBalance >= debt {
				amountStr = debtStr
				fullRepayment = true
			} else {
				amountStr = availBalance
			}

			opType := 0
			if v.VoteResults == define.VoteAgree && v.VoteStatus == define.VoteTakeEffect {
				opType = define.DecisionLiquidation
			} else {
				opType = define.ProtectLiquidation
			}

			// repayment
			err = l.Repayment(v.MinerId, amountStr, fullRepayment, opType)
			if err != nil {
				return err
			}
		} else {
			err := global.ZC_DB.Model(&system.SysContractWarnNode{}).Where("id = ?", v.ID).Update("node_status", define.VoteClose).Error
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetAvailBalance Available balance
func (l *LiquidateService) GetAvailBalance(minerId string) (string, error) {
	ctx := context.TODO()
	var balance string
	addr, err := address.NewFromString(minerId)
	if err != nil {
		return balance, xerrors.Errorf("NewFromString err:", err)
	}

	mact, err := lotusrpc.FullApi.StateGetActor(ctx, addr, types.EmptyTSK)
	if err != nil {
		return balance, err
	}

	tbs := blockstore.NewTieredBstore(blockstore.NewAPIBlockstore(lotusrpc.FullApi), blockstore.NewMemory())
	mas, err := miner.Load(adt.WrapStore(ctx, cbor.NewCborStore(tbs)), mact)
	if err != nil {
		return balance, err
	}

	availBalance, err := mas.AvailableBalance(mact.Balance)
	if err != nil {
		return balance, xerrors.Errorf("Getting available balance: %w", err)
	}

	return availBalance.String(), nil

}

// GetDebt Debt
func (l *LiquidateService) GetDebt(contractAddress string) (string, error) {
	param := contract.DebtGetBalanceOf.Keccak256()
	addr := common.HexToAddress(contractAddress)
	paramData := common.LeftPadBytes(addr.Bytes(), 32)
	param = append(param, paramData...)
	var debt big.Int
	if err := contract.DebtContract.CallContract(param, &debt); err != nil {
		fmt.Println("Failed to obtain debt:", err)
		return "0", err
	}
	return debt.String(), nil
}

// Repayment Repayment
func (l *LiquidateService) Repayment(minerId, amountStr string, fullRepayment bool, opType int) error {

	if strings.Contains(minerId, "f") || strings.Contains(minerId, "t") {
		minerId = strings.Replace(minerId, "f0", "", -1)
		minerId = strings.Replace(minerId, "t0", "", -1)
	}
	param := contract.OpNodeMethodRepayment.Keccak256()
	target, _ := new(big.Int).SetString(minerId, 10)
	amount, _ := new(big.Int).SetString(amountStr, 10)
	targetData := common.LeftPadBytes(target.Bytes(), 32)
	amountData := common.LeftPadBytes(amount.Bytes(), 32)
	var fullData []byte
	if fullRepayment {
		isTotal, _ := new(big.Int).SetString("1", 10)
		fullData = common.LeftPadBytes(isTotal.Bytes(), 32)
	} else {
		isTotal, _ := new(big.Int).SetString("0", 10)
		fullData = common.LeftPadBytes(isTotal.Bytes(), 32)
	}

	param = append(param, targetData...)
	param = append(param, amountData...)
	param = append(param, fullData...)
	if _, err := contract.OpNodeContract.PushContract(param); err != nil {
		fmt.Println("Repayment contract failed:", err)
		return err
	} else {
		minerIds, _ := strconv.Atoi(minerId)
		toAddr := utils.GetAddr(contract.DefaultWallet)
		paramData := &system.SysNodeRecords{
			ToAddr:    toAddr.String(),
			Applied:   true,
			Cid:       uuidGo.NewV4().String(),
			ActorId:   uint64(minerIds),
			OpContent: amountStr,
			OpType:    opType,
		}
		err = global.ZC_DB.Create(paramData).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// GetStopSectorsTotal Obtain the termination sector return amount
func (l *LiquidateService) GetStopSectorsTotal(minerId string, debt string) (float64, []string, error) {
	ctx := context.TODO()
	// Get sectors
	var allSectors []uint64
	var backFil float64
	dt, _ := strconv.ParseFloat(debt, 64)
	stopSector := make([]string, 0)
	faultySectors, err := l.GetFaultySectors(minerId)
	if err != nil {
		return backFil, stopSector, err
	}

	allSectors = append(allSectors, faultySectors...)
	addr, err := address.NewFromString(minerId)
	if err != nil {
		return backFil, stopSector, xerrors.Errorf("NewFromString err:", err)
	}

	head, err := lotusrpc.FullApi.ChainHead(ctx)
	if err != nil {
		return backFil, stopSector, err
	}

	// Normal sector
	activeSet, err := lotusrpc.FullApi.StateMinerActiveSectors(ctx, addr, head.Key())
	if err != nil {
		return backFil, stopSector, err
	}

	for _, v := range activeSet {
		num, _ := strconv.Atoi(v.SectorNumber.String())
		allSectors = append(allSectors, uint64(num))
	}

	if len(allSectors) > 0 {
		for i, v := range allSectors {
			if backFil >= dt {
				break
			}
			sectorDeposit, err := l.GetStopSectorsValue(minerId, []string{strconv.Itoa(int(allSectors[i]))})
			if err != nil {
				return backFil, stopSector, err
			}

			backFil += sectorDeposit
			stopSector = append(stopSector, strconv.Itoa(int(v)))
		}
	}

	return backFil, stopSector, nil
}

// GetFaultySectors Get abnormal sectors
func (l *LiquidateService) GetFaultySectors(minerId string) ([]uint64, error) {
	stor := store.ActorStore(context.TODO(), blockstore.NewAPIBlockstore(lotusrpc.FullApi))

	maddr, err := address.NewFromString(minerId)
	if err != nil {
		return nil, err
	}

	mact, err := lotusrpc.FullApi.StateGetActor(context.TODO(), maddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	mas, err := miner.Load(stor, mact)
	if err != nil {
		return nil, err
	}

	var faultySectors []uint64
	err = mas.ForEachDeadline(func(dlIdx uint64, dl miner.Deadline) error {
		return dl.ForEachPartition(func(partIdx uint64, part miner.Partition) error {
			faults, err := part.FaultySectors()
			if err != nil {
				return err
			}
			return faults.ForEach(func(num uint64) error {
				faultySectors = append(faultySectors, num)
				return nil
			})
		})
	})
	if err != nil {
		return nil, err
	}

	return faultySectors, nil
}

// GetStopSectorsValue Obtain termination sector value
func (l *LiquidateService) GetStopSectorsValue(minerId string, sectors []string) (fil float64, err error) {
	ctx := context.TODO()
	var fileValue float64
	maddr, err := address.NewFromString(minerId)
	if err != nil {
		return fileValue, err
	}
	mi, err := lotusrpc.FullApi.StateMinerInfo(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return fileValue, err
	}

	terminationDeclarationParams := []miner2.TerminationDeclaration{}

	for _, sn := range sectors {
		sectorNum, err := strconv.ParseUint(sn, 10, 64)
		if err != nil {
			return fileValue, fmt.Errorf("could not parse sector number: %w", err)
		}

		sectorbit := bitfield.New()
		sectorbit.Set(sectorNum)

		loca, err := lotusrpc.FullApi.StateSectorPartition(ctx, maddr, abi.SectorNumber(sectorNum), types.EmptyTSK)
		if err != nil {
			return fileValue, fmt.Errorf("get state sector partition %s", err)
		}

		para := miner2.TerminationDeclaration{
			Deadline:  loca.Deadline,
			Partition: loca.Partition,
			Sectors:   sectorbit,
		}
		terminationDeclarationParams = append(terminationDeclarationParams, para)
	}

	terminateSectorParams := &miner2.TerminateSectorsParams{
		Terminations: terminationDeclarationParams,
	}

	sp, err := actors.SerializeParams(terminateSectorParams)
	if err != nil {
		return fileValue, xerrors.Errorf("serializing params: %w", err)
	}

	msg := &types.Message{
		From:   mi.Worker,
		To:     maddr,
		Method: builtin.MethodsMiner.TerminateSectors,
		Value:  filecoin_big.Zero(),
		Params: sp,
	}

	invocResult, err := lotusrpc.FullApi.StateCall(ctx, msg, types.EmptyTSK)
	if err != nil {
		return fileValue, xerrors.Errorf("fail to state call: %w", err)
	}

	checkValue := false
	for _, im := range invocResult.ExecutionTrace.Subcalls {
		fmt.Println(fmt.Sprintf(""))
		if strings.Contains("f099,t099", im.Msg.To.String()) {
			checkValue = true
			fileValue, _ = strconv.ParseFloat(im.Msg.Value.String(), 64)
			break
		}
	}

	if checkValue {
		ts, err := lotusrpc.FullApi.ChainHead(ctx)
		if err != nil {
			return fileValue, err
		}
		sectorId, _ := strconv.Atoi(sectors[0])
		si, err := lotusrpc.FullApi.StateSectorGetInfo(ctx, maddr, abi.SectorNumber(uint64(sectorId)), ts.Key())
		if err != nil {
			return fileValue, err
		}
		if si == nil {
			return fileValue, xerrors.Errorf("sector %d for miner %s not found", sectorId, maddr)
		}

		sectorPledge, _ := strconv.ParseFloat(si.InitialPledge.String(), 64)
		//退回质押=扇区质押-预估罚息
		fileValue = sectorPledge - fileValue

		return fileValue, nil
	}

	return fileValue, errors.New("fil is zero")
}

// StopSectors Terminate sector
func (l *LiquidateService) StopSectors(minerId string, sectors []string) error {
	ctx := context.TODO()
	maddr, err := address.NewFromString(minerId)
	if err != nil {
		return err
	}

	terminationDeclarationParams := []miner2.TerminationDeclaration{}

	for _, sn := range sectors {
		sectorNum, err := strconv.ParseUint(sn, 10, 64)
		if err != nil {
			return fmt.Errorf("could not parse sector number: %w", err)
		}

		sectorbit := bitfield.New()
		sectorbit.Set(sectorNum)

		loca, err := lotusrpc.FullApi.StateSectorPartition(context.TODO(), maddr, abi.SectorNumber(sectorNum), types.EmptyTSK)
		if err != nil {
			return fmt.Errorf("get state sector partition %s", err)
		}

		para := miner2.TerminationDeclaration{
			Deadline:  loca.Deadline,
			Partition: loca.Partition,
			Sectors:   sectorbit,
		}

		terminationDeclarationParams = append(terminationDeclarationParams, para)
	}

	terminateSectorParams := &miner2.TerminateSectorsParams{
		Terminations: terminationDeclarationParams,
	}

	sp, err := actors.SerializeParams(terminateSectorParams)
	if err != nil {
		return xerrors.Errorf("serializing params: %w", err)
	}

	faddr := utils.GetAddr(contract.DefaultControlWallet)
	idAddr, err := lotusrpc.FullApi.StateLookupID(context.TODO(), faddr, types.EmptyTSK)
	if err == nil {
		fmt.Println("ID address: ", idAddr)
	} else {
		return err
	}

	var fromAddr address.Address
	fromAddr = faddr

	smsg, err := lotusrpc.FullApi.MpoolPushMessage(ctx, &types.Message{
		From:   fromAddr,
		To:     maddr,
		Method: builtin.MethodsMiner.TerminateSectors,
		Value:  filecoin_big.Zero(),
		Params: sp,
	}, nil)
	if err != nil {
		return xerrors.Errorf("mpool push message: %w", err)
	}

	fmt.Println("sent termination message:", smsg.Cid())

	wait, err := lotusrpc.FullApi.StateWaitMsg(ctx, smsg.Cid(), uint64(0), api.LookbackNoLimit, true)
	if err != nil {
		return err
	}

	if wait.Receipt.ExitCode.IsError() {
		return fmt.Errorf("terminate sectors message returned exit %d", wait.Receipt.ExitCode)
	}

	return nil
}

// ModifyControlAddr Change control address
func (l *LiquidateService) ModifyControlAddr(minerId string) error {
	addr, err := address.NewFromString(minerId)
	if err != nil {
		return err
	}

	minerInfo, err := lotusrpc.FullApi.StateMinerInfo(context.TODO(), addr, types.EmptyTSK)
	if err != nil {
		return err
	}

	contractAbi, err := c_abi.JSON(strings.NewReader(build.MinerContracts))
	if err != nil {
		return err
	}
	workerParam := minerInfo.Worker.String()
	if strings.Contains(minerId, "f") || strings.Contains(minerId, "t") {
		minerId = strings.Replace(minerId, "f0", "", -1)
		minerId = strings.Replace(minerId, "t0", "", -1)
	}
	if strings.Contains(workerParam, "f") || strings.Contains(workerParam, "t") {
		workerParam = strings.Replace(workerParam, "f0", "", -1)
		workerParam = strings.Replace(workerParam, "t0", "", -1)
	}
	target, _ := new(big.Int).SetString(minerId, 10)     //10
	worker, _ := new(big.Int).SetString(workerParam, 10) //10

	faddr := utils.GetAddr(contract.DefaultControlWallet)
	idAddr, err := lotusrpc.FullApi.StateLookupID(context.TODO(), faddr, types.EmptyTSK)
	if err == nil {
		fmt.Println("ID address: ", idAddr)
	} else {
		return err
	}
	controlAddr := idAddr.String()
	if strings.Contains(controlAddr, "f") || strings.Contains(controlAddr, "t") {
		controlAddr = strings.Replace(controlAddr, "f0", "", -1)
		controlAddr = strings.Replace(controlAddr, "t0", "", -1)
	}

	fmt.Println(fmt.Sprintf("Change control address target:%s,worker:%s,controlAddr:%s", target.String(), worker.String(), controlAddr))
	var controlAr []*big.Int
	control, _ := new(big.Int).SetString(controlAddr, 10)
	controlAr = append(controlAr, control)

	bytes, err := contractAbi.Pack("changeWorkerAddress", target, worker, controlAr)
	if err != nil {
		fmt.Println("Pack err:", err)
		return err
	} else {
		bytes = bytes[4:]
	}

	param := contract.MinerChangeWorkerAddress.Keccak256()
	param = append(param, bytes...)
	if _, err := contract.MinerContract.PushContract(param); err != nil {
		global.ZC_LOG.Error("Change control address contract call failed:", zap.Error(err))
		return err
	}

	return nil
}

// DelWarnVote Proposal to delete alarm data
func (l *LiquidateService) DelWarnVote(addr string) error {
	if addr == "" {
		return errors.New("addr is null")
	}
	addrAr := strings.Split(addr, ",")
	addrParam := ""
	for _, v := range addrAr {
		if addrParam == "" {
			addrParam = v
		} else {
			addrParam += `,` + v
		}
	}

	if addrParam != "" {
		err := global.ZC_DB.Delete(&system.SysContractWarnNode{}, "contract_address in("+addrParam+")").Error
		if err != nil {
			return err
		}
	}

	// Delete voting proposal, voters
	global.ZC_LOG.Info("Delete voting proposal, voters")
	for _, v := range addrAr {
		param := contract.VoteDelwarnNodeMap.Keccak256()
		addr := common.HexToAddress(v)
		addrData := common.LeftPadBytes(addr.Bytes(), 32)
		param = append(param, addrData...)
		_, err := contract.VoteContract.PushContract(param)
		if err != nil {
			global.ZC_LOG.Error("Delete voting proposal, voter failed:", zap.Error(err))
			return err
		}
	}

	return nil
}

// GetWarnNodeCount Obtain the number of voting alarms
func (l *LiquidateService) GetWarnNodeCount() (*response.WarnCount, error) {

	wc := &response.WarnCount{}
	var total int64
	err := global.ZC_DB.Model(&system.SysContractWarnNode{}).Where("node_status in(1,2)").Count(&total).Error
	if err != nil {
		return wc, err
	}
	wc.Total = int(total)
	return wc, nil
}
