// 自动生成模板SysHostInfo
package system

import (
	"zcfil-server/global"
)

type SysNodeInfo struct {
	global.ZC_MODEL
	NodeName      uint64  `json:"nodeName" form:"nodeName" gorm:"index;unique;column:node_name;comment:Node Name"`
	Status        int     `json:"status" form:"status" gorm:"column:status;comment:Node status: 1. Joined, 2. Resigned"`
	DebtBalance   string  `json:"debtBalance" form:"debtBalance" gorm:"column:debt_balance;comment:debt"`
	NodeAddress   string  `json:"nodeAddress" form:"nodeAddress" gorm:"column:node_address;comment:Node address"`
	Owner         string  `json:"owner" form:"owner" gorm:"column:owner;comment:Node owner"`
	Applied       bool    `json:"applied" form:"applied" gorm:"column:applied;comment:Effective or not"`
	Balance       float64 `json:"balance" form:"balance" gorm:"column:balance;comment:Node quota"`
	Operator      string  `json:"operator" form:"operator" gorm:"column:operator;comment:Operator"`
	MaxDebtRate   float64 `json:"maxDebtRate" form:"maxDebtRate" gorm:"column:max_debt_rate;comment:Maximum debt ratio"`
	LiquidateRate float64 `json:"liquidateRate" form:"liquidateRate" gorm:"column:liquidate_rate;comment:Clearing threshold"`
}

func (SysNodeInfo) TableName() string {
	return "sys_node_info"
}
