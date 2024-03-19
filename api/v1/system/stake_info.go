package system

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"math"
	"strconv"
	"time"
	"zcfil-server/config"
	"zcfil-server/define"
	"zcfil-server/global"
	"zcfil-server/model/common/request"
	"zcfil-server/model/common/response"
	systemResponse "zcfil-server/model/system/response"
	"zcfil-server/utils"
	"zcfil-server/utils/redisClient"
)

type StakeInfoApi struct{}

// GetTotalFilList Obtain the total pool fil line list
func (s *StakeInfoApi) GetTotalFilList(c *gin.Context) {
	var req request.GetFoldLineListReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, err := stakeInfoService.GetSysStakeInfoList(req)
	if err != nil {
		global.ZC_LOG.Error("Getting information failure!", zap.Error(err))
		response.FailWithMessage("Getting information failure", c)
		return
	}

	totalFilResp := make([]systemResponse.TotalFilResp, len(list))
	for key, val := range list {
		_createdAt := val.CreatedAt.Format("2006-01-02 15:00:00")
		dealTime, _ := time.ParseInLocation("2006-01-02 15:04:05", _createdAt, time.Local)
		totalFilResp[key].RecordTime = dealTime
		totalFilResp[key].TotalBalance = val.TotalBalance
	}

	sideResp := &systemResponse.TotalFilResp{RecordTime: time.Now()}

	redisInfo := redisClient.NewDefaultRedisStore()
	baseRate := math.Pow(10, 18)
	totalBalance := redisInfo.Get(define.TotalBalanceKey)
	if len(totalBalance) != 0 {
		_totalBalance, _ := strconv.ParseFloat(totalBalance, 64)
		sideResp.TotalBalance = utils.FloatAccurateBit(_totalBalance/baseRate, utils.TwoBit)
	}
	resp := make([]systemResponse.TotalFilResp, 0)
	resp = append(resp, *sideResp)
	resp = append(resp, totalFilResp...)

	response.OkWithDetailed(response.PageResult{
		List: resp,
	}, "Success", c)
}

// GetTotalFilSide Obtain the sidebar data of the total pool fil line list
func (s *StakeInfoApi) GetTotalFilSide(c *gin.Context) {
	sideResp := systemResponse.GetTotalFilSideResp{}
	baseRate := math.Pow(10, 18)

	redisInfo := redisClient.NewDefaultRedisStore()

	stakeAPY := redisInfo.Get(define.StakeAPYKey)
	if stakeAPY != "" {
		_stakeAPY, _ := strconv.ParseFloat(stakeAPY, 64)
		sideResp.StakeAPY = utils.FloatAccurateBit(_stakeAPY/baseRate, utils.FourBit)
	}

	loanAPY := redisInfo.Get(define.LoanAPYKey)
	if len(loanAPY) != 0 {
		_loanAPY, _ := strconv.ParseFloat(loanAPY, 64)
		sideResp.LoanAPY = utils.FloatAccurateBit(_loanAPY/baseRate, utils.FourBit)
	}

	financeUseRate := redisInfo.Get(define.FinanceUseRateKey)
	if len(financeUseRate) != 0 {
		_financeUseRate, _ := strconv.ParseFloat(financeUseRate, 64)
		sideResp.FinanceUseRate = utils.FloatAccurateBit(_financeUseRate/baseRate, utils.FourBit)
	}

	totalBalance := redisInfo.Get(define.TotalBalanceKey)
	if len(totalBalance) != 0 {
		_totalBalance, _ := strconv.ParseFloat(totalBalance, 64)
		sideResp.TotalBalance = utils.FloatAccurateBit(_totalBalance/baseRate, utils.TwoBit)
	}

	filValue := 3.72
	filPrice := redisInfo.Get(define.FilCoinPriceKey)
	if len(filPrice) != 0 {
		_filPrice, _ := strconv.ParseFloat(filPrice, 64)
		filValue = _filPrice
	}

	sideResp.FILValue = filValue

	if sideResp.TotalBalance != 0 {
		sideResp.TotalBalanceValue = utils.FloatAccurateBit(sideResp.TotalBalance*filValue, utils.TwoBit)
	}

	nodeBalanceSum, err := NodeInfoService.GetNodeBalanceSum()
	if err != nil {
		fmt.Println("Failed to obtain node count information from smart contract!", err)
	} else {
		sideResp.TotalLockValue = utils.FloatAccurateBit(nodeBalanceSum*filValue/baseRate, utils.TwoBit)
	}

	lastAvailableBalance := redisInfo.Get(define.LastAvailableBalanceKey)
	if len(lastAvailableBalance) != 0 {
		_lastAvailableBalance, _ := strconv.ParseFloat(lastAvailableBalance, 64)
		sideResp.LastAvailableBalance = utils.FloatAccurateBit(_lastAvailableBalance/baseRate, utils.TwoBit)
	}

	stakeNum := redisInfo.Get(define.StakeNumKey)
	if len(stakeNum) != 0 {
		_stakeNum, _ := strconv.ParseInt(stakeNum, 10, 64)
		sideResp.StakeNum = _stakeNum
	}

	nodeNum, err := NodeInfoService.GetNodeNum(config.NodeAppliedTrue)
	if err != nil {
		fmt.Println("Failed to obtain node count information from smart contract!", err)
	} else {
		sideResp.NodeNum = nodeNum
	}

	riskCoefficient := redisInfo.Get(define.RiskCoefficientKey)
	if len(riskCoefficient) != 0 {
		_riskCoefficient, _ := strconv.ParseFloat(riskCoefficient, 64)
		sideResp.RiskCoefficient = _riskCoefficient / 100
	}

	businessCoefficient := redisInfo.Get(define.BusinessCoefficientKey)
	if len(businessCoefficient) != 0 {
		_businessCoefficient, _ := strconv.ParseFloat(businessCoefficient, 64)
		sideResp.BusinessCoefficient = _businessCoefficient / 100
	}

	response.OkWithData(sideResp, c)
}

// GetTotalRateSide Get sidebar data for total interest rates
func (s *StakeInfoApi) GetTotalRateSide(c *gin.Context) {
	sideResp := &systemResponse.GetTotalRateSideResp{}
	dealTotalRateSide(sideResp)

	stakeRateAVG, err := stakeInfoService.GetStakeRate()
	if err != nil {
		fmt.Println("Failed to obtain the average value of pledge interest rate!", err)
	} else {
		if stakeRateAVG != 0 {
			sideResp.StakeQOQ = utils.FloatAccurateBit((sideResp.StakeAPY-stakeRateAVG)/stakeRateAVG, utils.FourBit)
		}
	}

	loanRateAVG, err := debtInfoService.GetLoanRate()
	if err != nil {
		fmt.Println("Failed to obtain the average loan interest rate!", err)
	} else {
		if loanRateAVG != 0 {
			sideResp.LoanQOQ = utils.FloatAccurateBit((sideResp.LoanAPY-loanRateAVG)/loanRateAVG, utils.FourBit)
		}
	}

	financeUseRateAVG, err := stakeInfoService.GetFinanceUseRate()
	if err != nil {
		fmt.Println("Failed to obtain the average fund utilization rate!", err)
	} else {
		if financeUseRateAVG != 0 {
			sideResp.FinanceUseQOQ = utils.FloatAccurateBit((sideResp.FinanceUseRate-financeUseRateAVG)/financeUseRateAVG, utils.FourBit)
		}
	}

	response.OkWithData(sideResp, c)
}

func dealTotalRateSide(sideResp *systemResponse.GetTotalRateSideResp) {
	baseRate := math.Pow(10, 18)
	redisInfo := redisClient.NewDefaultRedisStore()

	stakeAPY := redisInfo.Get(define.StakeAPYKey)
	if stakeAPY != "" {
		_stakeAPY, _ := strconv.ParseFloat(stakeAPY, 64)
		sideResp.StakeAPY = utils.FloatAccurateBit(_stakeAPY/baseRate, utils.FourBit)
	}

	loanAPY := redisInfo.Get(define.LoanAPYKey)
	if len(loanAPY) != 0 {
		_loanAPY, _ := strconv.ParseFloat(loanAPY, 64)
		sideResp.LoanAPY = utils.FloatAccurateBit(_loanAPY/baseRate, utils.FourBit)
	}

	financeUseRate := redisInfo.Get(define.FinanceUseRateKey)
	if len(financeUseRate) != 0 {
		_financeUseRate, _ := strconv.ParseFloat(financeUseRate, 64)
		sideResp.FinanceUseRate = utils.FloatAccurateBit(_financeUseRate/baseRate, utils.FourBit)
	}
}

// GetLoanRateSide Get sidebar data for borrowing and lending
func (s *StakeInfoApi) GetLoanRateSide(c *gin.Context) {
	sideResp := &systemResponse.GetLoanRateSideResp{}
	baseRate := math.Pow(10, 18)
	dealLoanRateSideResp(sideResp, baseRate)

	redisInfo := redisClient.NewDefaultRedisStore()
	lastAvailableBalance := redisInfo.Get(define.LastAvailableBalanceKey)
	if len(lastAvailableBalance) != 0 {
		_lastAvailableBalance, _ := strconv.ParseFloat(lastAvailableBalance, 64)
		sideResp.RemainFil = utils.FloatAccurateBit(_lastAvailableBalance/baseRate, utils.TwoBit)
	}

	loanRateAVG, err := debtInfoService.GetLoanRate()
	if err != nil {
		fmt.Println("Failed to obtain the average loan interest rate!", err)
	} else {
		if loanRateAVG != 0 {
			sideResp.LoanQOQ = utils.FloatAccurateBit((sideResp.LoanAPY-loanRateAVG)/loanRateAVG, utils.FourBit)
		}
	}

	totalLoanAVG, err := debtInfoService.GetTotalBalance()
	if err != nil {
		fmt.Println("Failed to obtain the average total loan amount!", err)
	} else {
		if totalLoanAVG != 0 {
			sideResp.TotalLoanQOQ = utils.FloatAccurateBit((sideResp.TotalLoan-totalLoanAVG)/totalLoanAVG, utils.FourBit)
		}
	}

	contractBalAVG, err := stakeInfoService.GetContractBalAVGData()
	if err != nil {
		fmt.Println("Failed to obtain the average remaining funds!", err)
	} else {
		if contractBalAVG != 0 {
			sideResp.RemainFilQOQ = utils.FloatAccurateBit((sideResp.RemainFil-contractBalAVG)/contractBalAVG, utils.FourBit)
		}
	}

	response.OkWithData(sideResp, c)
}

func dealLoanRateSideResp(sideResp *systemResponse.GetLoanRateSideResp, baseRate float64) {
	redisInfo := redisClient.NewDefaultRedisStore()
	loanAPY := redisInfo.Get(define.LoanAPYKey)
	if len(loanAPY) != 0 {
		_loanAPY, _ := strconv.ParseFloat(loanAPY, 64)
		sideResp.LoanAPY = utils.FloatAccurateBit(_loanAPY/baseRate, utils.FourBit)
	}

	totalLoan := redisInfo.Get(define.TotalLoanKey)
	if len(totalLoan) != 0 {
		_totalLoan, _ := strconv.ParseFloat(totalLoan, 64)
		sideResp.TotalLoan = utils.FloatAccurateBit(_totalLoan/baseRate, utils.TwoBit)
	}
}

// GetStackRateSide Obtain collateral sidebar data
func (s *StakeInfoApi) GetStackRateSide(c *gin.Context) {
	sideResp := &systemResponse.GetStakeRateSideResp{}
	dealStakeRateSideResp(sideResp)

	stakeRateAVG, err := stakeInfoService.GetStakeRate()
	if err != nil {
		fmt.Println("Failed to obtain the average value of pledge interest rate!", err)
	} else {
		if stakeRateAVG != 0 {
			sideResp.StakeQOQ = utils.FloatAccurateBit((sideResp.StakeAPY-stakeRateAVG)/stakeRateAVG, utils.FourBit)
		}
	}

	stakeTotalAVG, err := stakeInfoService.GetStakeTotalAVGData()
	if err != nil {
		fmt.Println("Failed to obtain the average total pledge value!", err)
	} else {
		if stakeTotalAVG != 0 {
			sideResp.TotalStakeQOQ = utils.FloatAccurateBit((sideResp.TotalStake-stakeTotalAVG)/stakeTotalAVG, utils.FourBit)
		}
	}

	response.OkWithData(sideResp, c)
}

func dealStakeRateSideResp(sideResp *systemResponse.GetStakeRateSideResp) {
	baseRate := math.Pow(10, 18)
	redisInfo := redisClient.NewDefaultRedisStore()

	stakeAPY := redisInfo.Get(define.StakeAPYKey)
	if stakeAPY != "" {
		_stakeAPY, _ := strconv.ParseFloat(stakeAPY, 64)
		sideResp.StakeAPY = utils.FloatAccurateBit(_stakeAPY/baseRate, utils.FourBit)
	}

	totalStake := redisInfo.Get(define.TotalStakeKey)
	if len(totalStake) != 0 {
		_totalStake, _ := strconv.ParseFloat(totalStake, 64)
		sideResp.TotalStake = utils.FloatAccurateBit(_totalStake/baseRate, utils.TwoBit)
	}
}

// GetFinanceUseRate Obtain funding utilization rate
func (s *StakeInfoApi) GetFinanceUseRate(c *gin.Context) {
	var useRate float64
	baseRate := math.Pow(10, 18)
	redisInfo := redisClient.NewDefaultRedisStore()

	financeUseRate := redisInfo.Get(define.FinanceUseRateKey)
	if len(financeUseRate) != 0 {
		_financeUseRate, _ := strconv.ParseFloat(financeUseRate, 64)
		useRate = utils.FloatAccurateBit(_financeUseRate/baseRate, utils.FourBit)
	}

	response.OkWithData(map[string]interface{}{"useRate": useRate}, c)
}

// GetStakeInfoList Obtain a line list of pledge interest rates and total pledge amounts
func (s *StakeInfoApi) GetStakeInfoList(c *gin.Context) {
	var req request.GetFoldLineListReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, err := stakeInfoService.GetSysStakeInfoList(req)
	if err != nil {
		global.ZC_LOG.Error("Getting information failure!", zap.Error(err))
		response.FailWithMessage("Getting information failure", c)
		return
	}

	totalFilResp := make([]systemResponse.StakeInfoResp, len(list))
	for key, val := range list {
		_createdAt := val.CreatedAt.Format("2006-01-02 15:00:00")
		dealTime, _ := time.ParseInLocation(config.TimeFormat, _createdAt, time.Local)
		totalFilResp[key].RecordTime = dealTime
		totalFilResp[key].StakeBalance = val.StakeBalance
		totalFilResp[key].StakeRate = val.StakeRate
	}

	sideResp := &systemResponse.GetStakeRateSideResp{}
	dealStakeRateSideResp(sideResp)
	currentTotalResp := systemResponse.StakeInfoResp{
		RecordTime:   time.Now(),
		StakeBalance: sideResp.TotalStake,
		StakeRate:    sideResp.StakeAPY,
	}

	resp := make([]systemResponse.StakeInfoResp, 0)
	resp = append(resp, currentTotalResp)
	resp = append(resp, totalFilResp...)

	response.OkWithDetailed(response.PageResult{
		List: resp,
	}, "Success", c)
}

// GetTotalityChangeList Obtain a list of fund utilization rate, loan interest rate, and pledge interest rate
func (s *StakeInfoApi) GetTotalityChangeList(c *gin.Context) {
	var req request.GetFoldLineListReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, err := stakeInfoService.GetSysStakeInfoList(req)
	if err != nil {
		global.ZC_LOG.Error("Getting information failure!", zap.Error(err))
		response.FailWithMessage("Getting information failure", c)
		return
	}

	totalityChangeResp := make([]systemResponse.TotalityChangeResp, len(list))
	for key, val := range list {
		_createdAt := val.CreatedAt.Format("2006-01-02 15:00:00")
		dealTime, _ := time.ParseInLocation(config.TimeFormat, _createdAt, time.Local)
		totalityChangeResp[key].RecordTime = dealTime
		totalityChangeResp[key].FinanceUseRate = val.FinanceUseRate
		totalityChangeResp[key].StakeRate = val.StakeRate
		debtInfo, err := debtInfoService.GetSysDebtInfo(dealTime)
		if err != nil {
			global.ZC_LOG.Error("Getting information failure!", zap.Error(err))
		} else {
			totalityChangeResp[key].DebtRate = debtInfo.DebtRate
		}
	}

	sideResp := &systemResponse.GetTotalRateSideResp{}
	dealTotalRateSide(sideResp)
	currentTotalFilResp := systemResponse.TotalityChangeResp{
		RecordTime:     time.Now(),
		FinanceUseRate: sideResp.FinanceUseRate,
		StakeRate:      sideResp.StakeAPY,
		DebtRate:       sideResp.LoanAPY,
	}

	resp := make([]systemResponse.TotalityChangeResp, 0)
	resp = append(resp, currentTotalFilResp)
	resp = append(resp, totalityChangeResp...)

	response.OkWithDetailed(response.PageResult{
		List: resp,
	}, "Success", c)
}
