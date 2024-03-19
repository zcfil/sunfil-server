// 自动生成模板SysHostInfo
package system

import (
	"zcfil-server/global"
)

type SysPledgeInfo struct {
	global.ZC_MODEL
	PledgeTime   string `json:"pledgeTime" gorm:"comment:Pledge time"`
	Profit       string `json:"profit"  gorm:"comment:Income"`
	Pledge       string `json:"pledge" gorm:"comment:Pledge"`
	TotalRevenue string `json:"totalRevenue" gorm:"comment:Total revenue"`
	TotalPower   string `json:"totalPower" gorm:"comment:Total computing power"`
}

func (SysPledgeInfo) TableName() string {
	return "sys_pledge_info"
}
