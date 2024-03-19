// 自动生成模板SysHostInfo
package system

import (
	"zcfil-server/global"
)

// 如果含有time.Time 请自行import time包
type SysDebtInfo struct {
	global.ZC_MODEL
	DebtBalance float64 `json:"debtBalance" form:"debtBalance" gorm:"column:debt_balance;comment:Lending currency quantity"`
	DebtRate    float64 `json:"debtRate" form:"debtRate" gorm:"column:debt_rate;comment:Loan interest rate"`
}

func (SysDebtInfo) TableName() string {
	return "sys_debt_info"
}
