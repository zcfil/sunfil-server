package system

import "zcfil-server/global"

type SysSectorRecords struct {
	global.ZC_MODEL
	MinerId      string `gorm:"comment:Node number" json:"minerId"`
	Sectors      string `gorm:"type:text;comment:Sectors" json:"sectors"`
	SectorCount  int    `gorm:"comment:Sector count" json:"sectorCount"`
	PledgeAmount string `gorm:"comment:Sector pledge amount" json:"PledgeAmount"`
}

func (SysSectorRecords) TableName() string {
	return "sys_sector_records"
}
