package initialize

import (
	"gorm.io/gorm"
	"zcfil-server/config"
	"zcfil-server/global"
)

const sys = "system"

func DBList() {
	dbMap := make(map[string]*gorm.DB)
	for _, info := range global.ZC_CONFIG.DBList {
		if info.Disable {
			continue
		}
		switch info.Type {
		case "mysql":
			dbMap[info.AliasName] = GormMysqlByConfig(config.Mysql{GeneralDB: info.GeneralDB})
		default:
			continue
		}
	}
	// Make a special judgment to determine if there is a migration
	// Adapt to low version migration of multiple database versions
	if sysDB, ok := dbMap[sys]; ok {
		global.ZC_DB = sysDB
	}
	global.ZC_DBList = dbMap
}
