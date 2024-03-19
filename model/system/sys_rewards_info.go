package system

import (
	"zcfil-server/global"
)

type SysRewardsInfo struct {
	global.ZC_MODEL
	HeightRewards string `json:"heightRewards" gorm:"comment:High block revenue"`
	BlockRewards  string `json:"blockRewards" gorm:"comment:Block revenue"`
	BlockCount    int    `json:"blockCount" gorm:"comment:Number of blocks"`
	BlockHeight   int    `json:"blockHeight" gorm:"comment:Block height "`
	BlockTime     string `json:"blockTime" gorm:"comment:Block time"`
}

func (SysRewardsInfo) TableName() string {
	return "sys_rewards_info"
}
