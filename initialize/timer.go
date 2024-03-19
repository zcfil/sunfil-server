package initialize

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-state-types/abi"
	builtintypes "github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/robfig/cron"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"
	apiSystem "zcfil-server/api/v1/system"
	"zcfil-server/contract"
	"zcfil-server/define"
	"zcfil-server/global"
	"zcfil-server/initialize/internal"
	"zcfil-server/lotusrpc"
	modelSystem "zcfil-server/model/system"
	"zcfil-server/service"
	"zcfil-server/service/system"
	"zcfil-server/utils"
	"zcfil-server/utils/redisClient"
)

// RecordStackPoolBalance Regularly record pledge pool information to the database
func RecordStackPoolBalance() {
	log.Println("Start recording the total amount information of the pledge pool")
	c := cron.New()
	// At each full 4-hour time point, execute corresponding tasks on a scheduled basis
	c.AddFunc("0 0 0,4,8,12,16,20 * * ?", func() {
		// Obtain total amount
		var totalBalance big.Int
		if err := contract.DebtContract.CallContract(contract.TotalFilBalance.Keccak256(), &totalBalance); err != nil {
			fmt.Println("Failed to obtain the total amount information of the pledge pool from the smart contract!", err)
		}

		// Obtain the pledged amount
		var stakeBalance big.Int
		if err := contract.StakeContract.CallContract(contract.StakeTotalSupply.Keccak256(), &stakeBalance); err != nil {
			fmt.Println("Failed to obtain record pledge pool amount from smart contract!", err)
		}

		// Obtain the interest rate of pledge
		var stakeRate big.Int
		if err := contract.RateContract.CallContract(contract.DepositRate.Keccak256(), &stakeRate); err != nil {
			fmt.Println("Failed to obtain record pledge interest rate from smart contract!", err)
		}

		// Obtain funding utilization rate
		var financeUseRate big.Int
		if err := contract.RateContract.CallContract(contract.PoolUseRate.Keccak256(), &financeUseRate); err != nil {
			fmt.Println("Failed to obtain fund utilization rate from smart contract!", err)
		}

		// Obtain the pond amount
		var lastAvailableBalance big.Int
		if err := contract.StakeContract.CallContract(contract.GetContractAmount.Keccak256(), &lastAvailableBalance); err != nil {
			fmt.Println("Failed to obtain remaining available current asset information from smart contract!", err)
		}

		baseRate := math.Pow(10, 18)

		_totalBalance, _ := strconv.ParseFloat(totalBalance.String(), 64)
		_stakeBalance, _ := strconv.ParseFloat(stakeBalance.String(), 64)
		_stakeRate, _ := strconv.ParseFloat(stakeRate.String(), 64)
		_financeUseRate, _ := strconv.ParseFloat(financeUseRate.String(), 64)
		_lastAvailableBalance, _ := strconv.ParseFloat(lastAvailableBalance.String(), 64)

		_totalBalance2 := utils.FloatAccurateBit(_totalBalance/baseRate, utils.TwoBit)
		_stakeBalance2 := utils.FloatAccurateBit(_stakeBalance/baseRate, utils.TwoBit)
		_lastAvailableBalance2 := utils.FloatAccurateBit(_lastAvailableBalance/baseRate, utils.TwoBit)

		_stakeRate4 := utils.FloatAccurateBit(_stakeRate/baseRate, utils.FourBit)
		_financeUseRate4 := utils.FloatAccurateBit(_financeUseRate/baseRate, utils.FourBit)

		stakeInfo := modelSystem.SysStakeInfo{
			ZC_MODEL:        global.ZC_MODEL{CreatedAt: time.Now(), UpdatedAt: time.Now()},
			TotalBalance:    _totalBalance2,
			StakeBalance:    _stakeBalance2,
			StakeRate:       _stakeRate4,
			FinanceUseRate:  _financeUseRate4,
			ContractBalance: _lastAvailableBalance2,
		}
		stakeInfoService := system.StakeInfoService{}
		err := stakeInfoService.CreateSysStakeInfo(stakeInfo)
		if err != nil {
			log.Println("Failed to record the total amount information of the pledge pool!", err.Error())
			return
		}
		log.Println("The total amount data of the pledge pool has been successfully recorded!")
	})
	go c.Start()
	defer c.Stop()
	select {}
}

// RecordDebtInfo Regularly record loan pool information to the database
func RecordDebtInfo() {
	log.Println("Start recording the total amount information of the pledge pool")
	c := cron.New()
	// At each full 4-hour time point, execute corresponding tasks on a scheduled basis
	c.AddFunc("0 0 0,4,8,12,16,20 * * ?", func() {
		// Obtain the amount of borrowing and lending
		var debtBalance big.Int
		if err := contract.DebtContract.CallContract(contract.GetTotalSupply.Keccak256(), &debtBalance); err != nil {
			fmt.Println("Failed to retrieve recorded loan pool amount from smart contract!", err)
		}

		// Obtain interest rates for borrowing and lending
		var debtRate big.Int
		if err := contract.RateContract.CallContract(contract.LoanRate.Keccak256(), &debtRate); err != nil {
			fmt.Println("Failed to obtain recorded loan interest rate from smart contract!", err)
		}

		baseRate := math.Pow(10, 18)

		_debtBalance, _ := strconv.ParseFloat(debtBalance.String(), 64)
		_debtRate, _ := strconv.ParseFloat(debtRate.String(), 64)

		_debtBalance2 := utils.FloatAccurateBit(_debtBalance/baseRate, utils.TwoBit)

		_debtRate4 := utils.FloatAccurateBit(_debtRate/baseRate, utils.FourBit)

		debtInfo := modelSystem.SysDebtInfo{
			ZC_MODEL:    global.ZC_MODEL{CreatedAt: time.Now(), UpdatedAt: time.Now()},
			DebtBalance: _debtBalance2,
			DebtRate:    _debtRate4,
		}
		debtInfoService := system.DebtInfoService{}
		err := debtInfoService.CreateSysDebtInfo(debtInfo)
		if err != nil {
			log.Println("Failed to record the total amount information of the pledge pool!", err.Error())
			return
		}
		log.Println("The total amount data of the pledge pool has been successfully recorded!")
	})
	go c.Start()
	defer c.Stop()
	select {}
}

func RecordNodeDebtNum() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	// Execute every 30 seconds
	for range ticker.C {
		list, err := apiSystem.NodeInfoService.GetNodeList()
		if err != nil {
			global.ZC_LOG.Error("Failed to obtain node list!", zap.Error(err))
			return
		}
		if len(list) == 0 {
			global.ZC_LOG.Error("No node data, skip to the next cycle", zap.Error(err))
			return
		}

		var sys sync.WaitGroup
		sys.Add(len(list))
		// Simultaneously obtaining the loan situation of the corresponding node
		for _, val := range list {
			go func(nodeInfo modelSystem.SysNodeInfo) {
				defer sys.Done()
				internal.UpdateNodeDebt(nodeInfo)
			}(val)
		}

		sys.Wait()
	}
}

func ChainTicker() {
	ticker := time.NewTicker(time.Second * 1)
	height, err := service.ServiceGroupApp.SystemServiceGroup.GetHeight()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("---ChainTicker Begin---")
	for {
		select {
		case <-ticker.C:
			if err = scanChain(height); err != nil {
				log.Println(err)
				go func(i int64) {
					for {
						if err := scanChain(i); err != nil {
							time.Sleep(time.Second)
							continue
						}
						break
					}
				}(height)
			}
		case msg := <-global.PushNoticeChan:
			recordsMap, err := service.ServiceGroupApp.SystemServiceGroup.GetMultisigRecords()
			if err != nil {
				log.Println(err)
				continue
			}
			if err = internal.MsgOperatorInfo(context.Background(), msg.Cid(), msg, recordsMap); err != nil {
				log.Println(err)
				continue
			}
		}
		if height < utils.BlockHeight() {
			// Abnormal synchronization, accelerating speed
			height++
			service.ServiceGroupApp.SystemServiceGroup.UpdateHeight(height)
			ticker = time.NewTicker(time.Second * 1)
		} else {
			ticker = time.NewTicker(time.Second * 15)
		}
	}
}

// Chain scanning
func scanChain(height int64) error {
	fmt.Println("---scanChain Begin---")
	var ctx = context.Background()
	tip, err := lotusrpc.FullApi.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(height), types.EmptyTSK)
	if err != nil {
		log.Println("ChainGetTipSetByHeight:", height, err)
		return nil
	}

	var msgs = make(map[cid.Cid]*types.Message)
	for _, cd := range tip.Cids() {
		blk, err := lotusrpc.FullApi.ChainGetBlockMessages(ctx, cd)
		if err != nil {
			return err
		}
		for _, bls := range blk.BlsMessages {
			msgs[bls.Cid()] = bls.VMMessage()
		}
		for _, secpk := range blk.SecpkMessages {
			msgs[secpk.Cid()] = secpk.VMMessage()
		}
	}

	recordsMap, err := service.ServiceGroupApp.SystemServiceGroup.GetMultisigRecords()
	if err != nil {
		return err
	}
	log.Println("Starting height：", height, ",number of chain messages：", len(msgs), ",number of multiple signed wallets in the database：", len(recordsMap))
	defer func() {
		if err != nil {
			log.Println("Abnormal skip height：", height, err)
		}
	}()
	//Originally, message CID() has a replacement situation
	for chainCid, msg := range msgs {
		if msg.Method == builtintypes.MethodsMultisig.Approve {
			log.Println(recordsMap, msg.To.String(), recordsMap[msg.To.String()])
		}
		err = internal.MsgOperatorInfo(ctx, chainCid, msg, recordsMap)
		return err
	}

	return nil
}

// SynPledgeInfo Synchronize pledge information
func SynPledgeInfo() {
	if err := service.ServiceGroupApp.SystemServiceGroup.PledgeInfoService.SynPledgeToContract(); err != nil {
		log.Println("SynPledgeToContract err:", err)
	}
}

// ContractWarnNode Obtain node alarm information
func ContractWarnNode() {
	ticker := time.NewTicker(time.Hour)
	for {
		select {
		case <-ticker.C:
			if err := service.ServiceGroupApp.SystemServiceGroup.LiquidateService.GetContractWarnNode(); err != nil {
				log.Println("GetContractWarnNode err:", err)
			}
		}
	}
}

// TimingRepayment Timed deduction
func TimingRepayment() {
	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ticker.C:
			if err := service.ServiceGroupApp.SystemServiceGroup.LiquidateService.TimingRepayment(); err != nil {
				log.Println("GetContractWarnNode err:", err)
			}
		}
	}
}

// SaveBlockRewards Save block rewards
func SaveBlockRewards() {
	service.ServiceGroupApp.SystemServiceGroup.PledgeInfoService.BlockRewardsTimer()
}

// UpdateNodeOperation Regularly update database node operator information
func UpdateNodeOperation() {
	log.Println("Starting to update database node operator information")
	ticker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()

	// Timed execution
	for range ticker.C {
		list, err := apiSystem.NodeInfoService.GetNodeList()
		if err != nil {
			global.ZC_LOG.Error("Failed to obtain node list!", zap.Error(err))
			return
		}
		if len(list) == 0 {
			global.ZC_LOG.Error("No node data, skip to the next cycle", zap.Error(err))
			return
		}

		// Simultaneously obtaining the loan situation of the corresponding node
		for _, val := range list {
			go func(nodeInfo modelSystem.SysNodeInfo) {
				pond, err := contract.MinerContract.MinerMethodGetMinerByActorId(strconv.FormatInt(int64(nodeInfo.NodeName), 10))
				if err != nil {
					log.Println("Failed to obtain node "+strconv.FormatInt(int64(nodeInfo.NodeName), 10)+" information", err)
					return
				} else {
					nodeInfo.Operator = pond.Operator
				}
				if nodeInfo.Operator != contract.EmptyAddress {
					nodeInfo.Status = contract.NodeStatusJoining
				}

				baseRate := math.Pow(10, 18)

				var debtRate, liquidateRate, warnPeriod, votePeriod big.Int
				debtRateParam := contract.RateGetRateContractParam.Keccak256()
				target, _ := new(big.Int).SetString(strconv.FormatInt(int64(nodeInfo.NodeName), 10), 10)
				targetData := common.LeftPadBytes(target.Bytes(), 32)
				debtRateParam = append(debtRateParam, targetData...)
				if err := contract.RateContract.CallContract(debtRateParam, &debtRate, &liquidateRate, &warnPeriod, &votePeriod); err != nil {
					log.Println("Abnormal acquisition of interest rate parameters callContract err:", err)
				} else {
					maxDebtRatio, _ := strconv.ParseFloat(debtRate.String(), 64)
					nodeInfo.MaxDebtRate = utils.FloatAccurateBit(maxDebtRatio/baseRate, utils.FourBit)

					threshold, _ := strconv.ParseFloat(liquidateRate.String(), 64)
					nodeInfo.LiquidateRate = utils.FloatAccurateBit(threshold/baseRate, utils.FourBit)
				}

				if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysNodeOperatorAndStatus(nodeInfo.NodeName, nodeInfo); err != nil {
					log.Println("Failed to modify node "+strconv.FormatInt(int64(nodeInfo.NodeName), 10)+" information", err)
					return
				}
			}(val)

		}

	}
}

// UpdateSolidityData Regularly obtain contract information and write it to the cache
func UpdateSolidityData() {
	log.Println("Starting to periodically retrieve contract information and write it to the cache")
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	redisInfo := redisClient.NewDefaultRedisStore()

	// Execute every 30 seconds
	for range ticker.C {
		/*Sidebar related*/
		// Pull the interest rate of pledge
		go func() {
			var stakeAPY big.Int
			if err := contract.RateContract.CallContract(contract.DepositRate.Keccak256(), &stakeAPY); err != nil {
				fmt.Println("Failed to obtain pledge interest rate information from smart contract!", err)
				return
			} else {
				err = redisInfo.Set(define.StakeAPYKey, stakeAPY.String())
				if err != nil {
					fmt.Println("Failed to write pledge interest rate to cache!", err)
				}
			}
		}()

		// Obtain interest rates for borrowing and lending
		go func() {
			var loanAPY big.Int
			if err := contract.RateContract.CallContract(contract.LoanRate.Keccak256(), &loanAPY); err != nil {
				fmt.Println("Failed to obtain loan interest rate information from smart contract!", err)
				return
			} else {
				err = redisInfo.Set(define.LoanAPYKey, loanAPY.String())
				if err != nil {
					fmt.Println("Failed to write loan interest rate to cache!", err)
				}
			}
		}()

		// Currency pool utilization rate
		go func() {
			var financeUseRate big.Int
			if err := contract.RateContract.CallContract(contract.PoolUseRate.Keccak256(), &financeUseRate); err != nil {
				fmt.Println("Failed to obtain coin pool utilization information from smart contract!", err)
				return
			} else {
				err = redisInfo.Set(define.FinanceUseRateKey, financeUseRate.String())
				if err != nil {
					fmt.Println("Currency pool utilization write to cache failed!", err)
				}
			}
		}()

		// Total pool volume
		go func() {
			var totalBalance big.Int
			if err := contract.DebtContract.CallContract(contract.TotalFilBalance.Keccak256(), &totalBalance); err != nil {
				fmt.Println("Failed to obtain the total amount information of the pledge pool from the smart contract!", err)
				return
			} else {
				err = redisInfo.Set(define.TotalBalanceKey, totalBalance.String())
				if err != nil {
					fmt.Println("Failed to write the total amount of pledge pool to the cache!", err)
				}
			}
		}()

		// Remaining available current assets
		go func() {
			var lastAvailableBalance big.Int
			if err := contract.StakeContract.CallContract(contract.GetContractAmount.Keccak256(), &lastAvailableBalance); err != nil {
				fmt.Println("Failed to obtain remaining available current asset information from smart contract!", err)
				return
			} else {
				err = redisInfo.Set(define.LastAvailableBalanceKey, lastAvailableBalance.String())
				if err != nil {
					fmt.Println("Failed to write available liquid assets to cache!", err)
				}
			}
		}()

		// Number of pledged individuals obtained
		go func() {
			var stakeNum big.Int
			if err := contract.StakeContract.CallContract(contract.StakeAddressNum.Keccak256(), &stakeNum); err != nil {
				fmt.Println("Failed to obtain information on the number of pledgers from the smart contract!", err)
				return
			} else {
				err = redisInfo.Set(define.StakeNumKey, stakeNum.String())
				if err != nil {
					fmt.Println("Failed to write the number of pledged individuals to the cache!", err)
				}
			}
		}()

		// Obtain risk reserve coefficient
		go func() {
			var riskCoefficient big.Int
			if err := contract.RateContract.CallContract(contract.RiskCoefficient.Keccak256(), &riskCoefficient); err != nil {
				fmt.Println("Failed to obtain risk reserve coefficient information from smart contract!", err)
				return
			} else {
				err = redisInfo.Set(define.RiskCoefficientKey, riskCoefficient.String())
				if err != nil {
					fmt.Println("Failed to write risk reserve coefficient to cache!", err)
				}
			}
		}()

		// Obtain operational reserve coefficient
		go func() {
			var businessCoefficient big.Int
			if err := contract.RateContract.CallContract(contract.OmCoefficient.Keccak256(), &businessCoefficient); err != nil {
				fmt.Println("Failed to obtain risk reserve coefficient information from smart contract!", err)
				return
			} else {
				err = redisInfo.Set(define.BusinessCoefficientKey, businessCoefficient.String())
				if err != nil {
					fmt.Println("Operation reserve coefficient write to cache failed!", err)
				}
			}
		}()

		// Obtain FIL price
		go func() {
			filPrice, err := GetFilCoinPrice()
			if err != nil {
				fmt.Println("Failed to obtain fil price", err)
				return
			} else {
				err = redisInfo.Set(define.FilCoinPriceKey, utils.Strval(filPrice))
				if err != nil {
					fmt.Println("Fil price write cache failed!", err)
				}
			}
		}()

		// Obtain total loan amount
		go func() {
			var totalLoan big.Int
			if err := contract.DebtContract.CallContract(contract.GetTotalSupply.Keccak256(), &totalLoan); err != nil {
				fmt.Println("Failed to obtain total loan amount information from smart contract!", err)
				return
			} else {
				err = redisInfo.Set(define.TotalLoanKey, totalLoan.String())
				if err != nil {
					fmt.Println("Failed to write the total loan amount to the cache!", err)
				}
			}
		}()

		// Obtain the total pledged amount
		go func() {
			var totalStake big.Int
			if err := contract.StakeContract.CallContract(contract.GetTotalSupply.Keccak256(), &totalStake); err != nil {
				fmt.Println("Failed to obtain pledge interest rate information from smart contract!", err)
				return
			} else {
				err = redisInfo.Set(define.TotalStakeKey, totalStake.String())
				if err != nil {
					fmt.Println("Failed to write the total pledged amount to the cache!", err)
				}
			}
		}()
	}
}

// GetFilCoinPrice Obtain fil value
func GetFilCoinPrice() (float64, error) {
	url := "https://api.filutils.com/api/v2/network/filprice"

	data, err := RequestDo(url, "", nil, time.Second*15, http.MethodPost)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}

	type DataInfo struct {
		NewlyPrice float64 `json:"newlyPrice"`
	}

	type ResponseInfo struct {
		Code int      `json:"code"`
		Data DataInfo `json:"data"`
		Msg  string   `json:"msg"`
	}

	var dataRes ResponseInfo
	if err = json.Unmarshal(data, &dataRes); err != nil {
		log.Println(err.Error())
		return 0, err
	}

	return dataRes.Data.NewlyPrice, nil
}

// RequestDo Encapsulate HTTP requests methodType POST,GET
func RequestDo(url, token string, request interface{}, timeout time.Duration, methodType string) ([]byte, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	var buf []byte
	if request != nil {
		var err error
		buf, err = json.Marshal(request)
		if err != nil {
			log.Println("Serialization failed：", err.Error())
			return nil, err
		}
	}

	path := url

	req, err := http.NewRequest(methodType, path, bytes.NewReader(buf))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, err
	}
	// Set header
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.FormatInt(req.ContentLength, 10))
	if token != "" {
		req.Header.Set("x-token", token)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request error：", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Request error：", *resp)
		return nil, fmt.Errorf("Request error code：%d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}
