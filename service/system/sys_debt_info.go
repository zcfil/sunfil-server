package system

import (
	"time"
	"zcfil-server/config"
	"zcfil-server/global"
	"zcfil-server/model/common/request"
	"zcfil-server/model/system"
	"zcfil-server/model/system/response"
)

//@function: CreateSysDebtInfo
//@description: Create total fil data
//@param: sysDebtInfo model.SysDebtInfo
//@return: err error

type DebtInfoService struct{}

func (debtInfoService *DebtInfoService) CreateSysDebtInfo(sysDebtInfo system.SysDebtInfo) (err error) {
	err = global.ZC_DB.Create(&sysDebtInfo).Error
	return err
}

//@function: GetSysDebtInfo
//@description: Obtain individual data of total fil information based on the corresponding time
//@param: id int
//@return: sysDebtInfo system.SysDebtInfo, err error

func (debtInfoService *DebtInfoService) GetSysDebtInfo(selectTime time.Time) (sysDebtInfo system.SysDebtInfo, err error) {
	err = global.ZC_DB.Where("created_at > ? AND created_at < ? ", selectTime.Add(time.Hour*-1), selectTime.Add(time.Hour)).First(&sysDebtInfo).Error
	return
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: DeleteSysDebtInfoByIds
//@description: Batch deletion of records
//@param: ids request.IdsReq
//@return: err error

func (debtInfoService *DebtInfoService) DeleteSysDebtInfoByIds(ids request.IdsReq) (err error) {
	err = global.ZC_DB.Delete(&[]system.SysDebtInfo{}, "id in (?)", ids.Ids).Unscoped().Error
	return err
}

// @function: GetSysDebtInfoList
// @description: Paging to obtain the total fil list
// @param: info request.PageInfo
// @return: list interface{}, total int64, err error

func (debtInfoService *DebtInfoService) GetSysDebtInfoList(info request.GetFoldLineListReq) (list []system.SysDebtInfo, err error) {
	db := global.ZC_DB.Model(&system.SysDebtInfo{})
	var entities []system.SysDebtInfo
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

// GetLoanRate Obtain the average loan interest rate from the previous day
func (debtInfoService *DebtInfoService) GetLoanRate() (float64, error) {
	sql := `select AVG(debt_rate) as 'debtRateAVG' from sys_debt_info where created_at >= DATE_SUB(date(now()),INTERVAL 1 DAY) and created_at < date(now())`
	data := response.DebtRateAVGData{}
	if err := global.ZC_DB.Raw(sql).Scan(&data).Error; err != nil {
		return 0, err
	}
	return data.DebtRateAVG, nil
}

// GetTotalBalance Obtain the average total borrowing and lending amount from the previous day
func (debtInfoService *DebtInfoService) GetTotalBalance() (float64, error) {
	sql := `select AVG(debt_balance) as 'debtTotalAVG' from sys_debt_info where created_at >= DATE_SUB(date(now()),INTERVAL 1 DAY) and created_at < date(now())`
	data := response.DebtTotalAVGData{}
	if err := global.ZC_DB.Raw(sql).Scan(&data).Error; err != nil {
		return 0, err
	}
	return data.DebtTotalAVG, nil
}
