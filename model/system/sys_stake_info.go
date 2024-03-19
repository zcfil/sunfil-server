// 自动生成模板SysHostInfo
package system

import (
	"zcfil-server/global"
)

type SysStakeInfo struct {
	global.ZC_MODEL
	TotalBalance    float64 `json:"totalBalance" form:"totalBalance" gorm:"column:total_balance;comment:Total amount of currency"`
	StakeBalance    float64 `json:"stakeBalance" form:"stakeBalance" gorm:"column:stake_balance;comment:Pledged currency quantity"`
	StakeRate       float64 `json:"stakeRate" form:"stakeRate" gorm:"column:stake_rate;comment:Pledge interest rate"`
	FinanceUseRate  float64 `json:"financeUseRate" form:"financeUseRate" gorm:"column:finance_use_rate;comment:Fund utilization rate"`
	ContractBalance float64 `json:"contractBalance" form:"contractBalance" gorm:"column:contract_balance;comment:Pond contract balance"`
}

func (SysStakeInfo) TableName() string {
	return "sys_stake_info"
}
