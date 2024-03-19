package system

import (
	"time"
	"zcfil-server/config"
	"zcfil-server/global"
	"zcfil-server/model/common/request"
	"zcfil-server/model/system"
	"zcfil-server/model/system/response"
)

//@function: CreateSysStakeInfo
//@description: Create total fil data
//@param: sysStakeInfo model.SysStakeInfo
//@return: err error

type StakeInfoService struct{}

func (hostGroupService *StakeInfoService) CreateSysStakeInfo(sysStakeInfo system.SysStakeInfo) (err error) {
	err = global.ZC_DB.Create(&sysStakeInfo).Error
	return err
}

//@function: GetSysStakeInfo
//@description: Obtain individual data of total fil information based on ID
//@param: id int
//@return: sysStakeInfo system.SysStakeInfo, err error

func (hostGroupService *StakeInfoService) GetSysStakeInfo(groupId int) (sysStakeInfo system.SysStakeInfo, err error) {
	err = global.ZC_DB.Where("id = ?", groupId).First(&sysStakeInfo).Error
	return
}

//@function: DeleteSysStakeInfoByIds
//@description: Batch deletion of records
//@param: ids request.IdsReq
//@return: err error

func (hostGroupService *StakeInfoService) DeleteSysStakeInfoByIds(ids request.IdsReq) (err error) {
	err = global.ZC_DB.Delete(&[]system.SysStakeInfo{}, "id in (?)", ids.Ids).Unscoped().Error
	return err
}

// @function: GetSysStakeInfoList
// @description: Paging to obtain the total fil list
// @param: info request.PageInfo
// @return: list interface{}, total int64, err error

func (hostGroupService *StakeInfoService) GetSysStakeInfoList(info request.GetFoldLineListReq) (list []system.SysStakeInfo, err error) {
	db := global.ZC_DB.Model(&system.SysStakeInfo{})
	var entities []system.SysStakeInfo
	var beginTime time.Time
	var endTime time.Time

	// Processing start and end time
	switch info.TimeInterval {
	case config.D1TimeInterval:
		beginTime = time.Now().Add(time.Hour * -24)
		endTime = time.Now()
	case config.D7TimeInterval:
		beginTime = time.Now().Add(time.Hour * -24 * 7)
		endTime = time.Now()
	case config.M1TimeInterval:
		beginTime = time.Now().Add(time.Hour * -24 * 30)
		endTime = time.Now()
	case config.M3TimeInterval:
		beginTime = time.Now().Add(time.Hour * -24 * 90)
		endTime = time.Now()
	default:
		beginTime = time.Now().Add(time.Hour * -24)
		endTime = time.Now()
	}
	db = db.Where("created_at >= ? and created_at <= ?", beginTime, endTime)

	err = db.Order("id desc").Find(&entities).Error
	return entities, err
}

// GetStakeRate Obtain the average pledge interest rate from the previous day
func (hostGroupService *StakeInfoService) GetStakeRate() (float64, error) {
	sql := `select AVG(stake_rate) as 'stakeRateAVG' from sys_stake_info where created_at >= DATE_SUB(date(now()),INTERVAL 1 DAY) and created_at < date(now())`
	data := response.StackRateAVGData{}
	if err := global.ZC_DB.Raw(sql).Scan(&data).Error; err != nil {
		return 0, err
	}
	return data.StakeRateAVG, nil
}

// GetFinanceUseRate Obtain the previous day's fund utilization rate
func (hostGroupService *StakeInfoService) GetFinanceUseRate() (float64, error) {
	sql := `select AVG(finance_use_rate) as 'financeUseRateAVG' from sys_stake_info where created_at >= DATE_SUB(date(now()),INTERVAL 1 DAY) and created_at < date(now())`
	data := response.FinanceUseRateAVGData{}
	if err := global.ZC_DB.Raw(sql).Scan(&data).Error; err != nil {
		return 0, err
	}
	return data.FinanceUseRateAVG, nil
}

// GetContractBalAVGData Obtain the total amount of currency from the previous day and
func (hostGroupService *StakeInfoService) GetContractBalAVGData() (float64, error) {
	sql := `select AVG(contract_balance) as 'contractBalAVG' from sys_stake_info where created_at >= DATE_SUB(date(now()),INTERVAL 1 DAY) and created_at < date(now())`
	data := response.ContractBalData{}
	if err := global.ZC_DB.Raw(sql).Scan(&data).Error; err != nil {
		return 0, err
	}
	return data.ContractBalAVG, nil
}

// GetStakeTotalAVGData Obtain the average of the previous day's total pledged amount
func (hostGroupService *StakeInfoService) GetStakeTotalAVGData() (float64, error) {
	sql := `select AVG(stake_balance) as 'stakeTotalAVG' from sys_stake_info where created_at >= DATE_SUB(date(now()),INTERVAL 1 DAY) and created_at < date(now())`
	data := response.StakeTotalData{}
	if err := global.ZC_DB.Raw(sql).Scan(&data).Error; err != nil {
		return 0, err
	}
	return data.StakeTotalAVG, nil
}
