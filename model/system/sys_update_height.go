package system

import "zcfil-server/global"

type SysUpdateHeight struct {
	global.ZC_MODEL
	Height int64 `json:"height" form:"height" gorm:"column:height;comment:height"`
}

func (SysUpdateHeight) TableName() string {
	return "sys_update_height"
}
