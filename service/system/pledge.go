package system

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/chain/actors/builtin/reward"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
	cbor "github.com/ipfs/go-ipld-cbor"
	"log"
	"math/big"
	"strconv"
	"time"
	"zcfil-server/contract"
	"zcfil-server/define"
	"zcfil-server/global"
	"zcfil-server/lotusrpc"
	"zcfil-server/model/system"
	"zcfil-server/utils"
)

type PledgeInfoService struct {
}

// SaveBlockRewards Save block rewards
func (p *PledgeInfoService) SaveBlockRewards(tipset *types.TipSet, height abi.ChainEpoch) error {
	log.Println("SaveBlockRewards synchronous height:", height.String())
	ctx := context.TODO()
	s1, _ := lotusrpc.FullApi.ChainGetTipSetByHeight(ctx, height, types.TipSetKey{})
	// Total reward for height
	heightReward := types.NewInt(0)
	heightBlock := types.NewInt(0)
	blockRewards := types.NewInt(0)
	if s1 != nil && len(s1.Blocks()) > 0 {
		for _, block := range s1.Blocks() {
			tbs := blockstore.NewTieredBstore(blockstore.NewAPIBlockstore(lotusrpc.FullApi), blockstore.NewMemory())
			rewardActor, err1 := lotusrpc.FullApi.StateGetActor(ctx, reward.Address, tipset.Key())
			if err1 != nil {
				log.Println("StateGetActor err:", err1.Error())
				return err1
			}
			rewardActorState, err := reward.Load(adt.WrapStore(ctx, cbor.NewCborStore(tbs)), rewardActor)
			if err != nil {
				log.Println("Load err:", err.Error())
				return err
			}
			blockReward, err := rewardActorState.ThisEpochReward()
			if err != nil {
				log.Println("rewardActorState err:", err.Error())
				return err
			}

			// Reward per block
			minerReward := types.BigDiv(types.BigMul(types.NewInt(uint64(block.ElectionProof.WinCount)), blockReward), types.NewInt(uint64(builtin.ExpectedLeadersPerEpoch)))

			if blockRewards.Equals(types.NewInt(0)) {
				blockRewards = minerReward
			}
			// Rewards for each height block
			heightReward = types.BigAdd(heightReward, types.BigMul(types.NewInt(uint64(block.ElectionProof.WinCount)), minerReward))
			// Count the number of blocks
			heightBlock = types.BigAdd(heightBlock, types.NewInt(uint64(block.ElectionProof.WinCount)))
		}
	}

	hs, _ := strconv.Atoi(height.String())

	var count int64
	err := global.ZC_DB.Model(&system.SysRewardsInfo{}).Where("block_height", hs).Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {

		param := &system.SysRewardsInfo{
			HeightRewards: heightReward.String(),
			BlockRewards:  blockRewards.String(),
			BlockCount:    int(heightBlock.Int64()),
			BlockHeight:   hs,
			BlockTime:     utils.BlockHeightToStr(int64(hs)),
		}

		err = global.ZC_DB.Model(&system.SysRewardsInfo{}).Create(param).Error
		if err != nil {
			return err
		}
	}

	// o calculations at 12:00 am
	if utils.BlockHeightToStr(int64(hs)) == utils.TimeToStr(time.Now(), utils.YYMMDD)+" 00:00:00" {

		// Total computing power
		pow, err := lotusrpc.FullApi.StateMinerPower(ctx, address.Address{}, types.EmptyTSK)
		if err != nil {
			return err
		}
		log.Println("pow:", pow.TotalPower.QualityAdjPower.String())

		type Rewards struct {
			AllRewards string `json:"all_rewards"`
		}

		var rd Rewards
		sql := ` select sum(b.height_rewards/1000000000000000000) AS all_rewards from sys_rewards_info b where b.block_time >=(now() - interval 24 hour) `
		err = global.ZC_DB.Raw(sql).Scan(&rd).Error
		if err != nil {
			return err
		}

		totalRewards, _ := strconv.ParseFloat(rd.AllRewards, 64)
		totalPow, _ := strconv.ParseFloat(pow.TotalPower.QualityAdjPower.String(), 64)
		totalPow = totalPow / define.TibCardinality
		profit := utils.FloatAccurateBit(totalRewards/totalPow, utils.FourBit)

		log.Println(fmt.Sprintf("totalRewards:%v,totalPow:%v,profit:%v", totalRewards, totalPow, profit))

		// Pledge
		circ, err := lotusrpc.FullApi.StateVMCirculatingSupplyInternal(ctx, tipset.Key())
		if err != nil {
			return err
		}

		// Pledge=storage pledge+consensus pledge
		// Storage pledge (expected revenue per unit of computing power in the next 20 days) * average daily production per T
		// Consensus pledge=0.3 (base) * circulation * (ratio of computing power size to the larger value of baseline or full network computing power)
		filCirculating := utils.StrToFloat64(circ.FilCirculating.String(), utils.FourBit)
		storagePledge := define.NexT20Days * profit
		consensusPledge := define.Cardinality * filCirculating * (define.Tib / totalPow)
		pledge := utils.FloatAccurateBit(storagePledge+consensusPledge, utils.FourBit)
		log.Println(fmt.Sprintf("filCirculating:%v,storagePledge:%v,consensusPledge:%v,pledge:%v", filCirculating, storagePledge, consensusPledge, pledge))

		profit = profit * utils.PointLength
		pledge = pledge * utils.PointLength

		param := &system.SysPledgeInfo{
			Profit:       fmt.Sprintf("%0.f", profit),
			Pledge:       fmt.Sprintf("%0.f", pledge),
			TotalRevenue: rd.AllRewards,
			TotalPower:   pow.TotalPower.QualityAdjPower.String(),
			PledgeTime:   utils.TimeToStr(time.Now(), utils.YYMMDD),
		}

		var total int64
		err = global.ZC_DB.Model(&system.SysPledgeInfo{}).Where("pledge_time=?", param.PledgeTime).Count(&total).Error
		if err != nil {
			return err
		}

		if total == 0 {

			err = global.ZC_DB.Model(&system.SysPledgeInfo{}).Create(&param).Error
			if err != nil {
				return err
			}
		}

		// Earnings, staking only retains data for one week
		delTime := utils.TimeToStr(time.Now().AddDate(0, 0, -7), utils.YYMMDD)
		sql = "delete from sys_rewards_info where block_time < ?"
		err = global.ZC_DB.Exec(sql, delTime).Error
		if err != nil {
			return err
		}

		sql = "delete from sys_pledge_info WHERE pledge_time < ?"
		err = global.ZC_DB.Exec(sql, delTime).Error
		if err != nil {
			return err
		}

		// Synchronize pledge information
		p.SynPledgeToContract()
	}

	return nil
}

func (p *PledgeInfoService) BlockRewardsTimer() error {
	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ticker.C:
			ctx := context.TODO()
			tipset, err := lotusrpc.FullApi.ChainHead(ctx)
			if err != nil {
				return err
			}

			height := tipset.Height() - abi.ChainEpoch(1)
			var srf system.SysRewardsInfo
			err = global.ZC_DB.Model(&system.SysRewardsInfo{}).Order("block_height desc").Limit(1).Find(&srf).Error
			if err != nil {
				log.Println("SysRewardsInfo query err:", err)
			}

			h, _ := strconv.Atoi(height.String())
			if srf.BlockHeight == h {
				continue
			}
			if (srf == system.SysRewardsInfo{}) {
				// Accelerate progress
				ticker = time.NewTicker(time.Second)
			} else if srf.BlockHeight < h {
				// Accelerate progress
				ticker = time.NewTicker(time.Second)
				srf.BlockHeight++
				height = abi.ChainEpoch(srf.BlockHeight)
			} else {
				// Restore for 30 seconds
				ticker = time.NewTicker(time.Second * 30)
				srf.BlockHeight++
				height = abi.ChainEpoch(srf.BlockHeight)
			}

			p.SaveBlockRewards(tipset, height)
		}
	}

}

// SynPledgeToContract Synchronize pledge information to the contract
func (p *PledgeInfoService) SynPledgeToContract() error {
	pledgeDay := utils.TimeToStr(time.Now().AddDate(0, 0, -7), utils.YYMMDD)
	var spf []system.SysPledgeInfo
	err := global.ZC_DB.Model(&system.SysPledgeInfo{}).Where("pledge_time > ?", pledgeDay).Find(&spf).Error
	if err != nil {
		return err
	}

	if len(spf) == 0 {
		return errors.New("pledge is nil")
	}

	var avgProfit, avgPledge float64
	for _, v := range spf {

		profit, _ := strconv.ParseFloat(v.Profit, 64)
		avgProfit += profit
		pledge, _ := strconv.ParseFloat(v.Pledge, 64)
		avgPledge += pledge
	}

	avgProfit = avgProfit / float64(len(spf))
	avgPledge = avgPledge / float64(len(spf))

	param := contract.RateSetPledge.Keccak256()
	profit, _ := new(big.Int).SetString(fmt.Sprintf("%0.f", avgProfit), 10)
	pledge, _ := new(big.Int).SetString(fmt.Sprintf("%0.f", avgPledge), 10)
	profitData := common.LeftPadBytes(profit.Bytes(), 32)
	pledgeData := common.LeftPadBytes(pledge.Bytes(), 32)
	param = append(param, profitData...)
	param = append(param, pledgeData...)

	_, err = contract.RateContract.PushContract(param)
	if err != nil {
		fmt.Println("RateSetPledge PushContract err:", err)
		return err
	} else {
		fmt.Println(fmt.Sprintf("Successfully synchronized pledge information, avgProfit:%f,avgPledge:%f", avgProfit, avgPledge))
	}

	return nil

}
