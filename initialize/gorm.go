package initialize

import (
	adapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"zcfil-server/global"
	"zcfil-server/model/system"
	"zcfil-server/service"
)

// Gorm Initialize the database and generate global database variables
func Gorm() *gorm.DB {
	switch global.ZC_CONFIG.System.DbType {
	case "mysql":
		return GormMysql()
	default:
		return GormMysql()
	}
}

// RegisterTables Register database tables exclusively
func RegisterTables(db *gorm.DB) {
	err := db.AutoMigrate(
		// System Module Table
		adapter.CasbinRule{},
		system.SysNodeInfo{},
		system.SysNodeRecords{},
		system.SysUpdateHeight{},
		system.SysContractAbi{},
		system.SysContractWarnNode{},
		system.SysStakeInfo{},
		system.SysDebtInfo{},
		system.SysPledgeInfo{},
		system.SysRewardsInfo{},
		system.SysSectorRecords{},
	)
	if err != nil {
		global.ZC_LOG.Error("register table failed", zap.Error(err))
		os.Exit(0)
	}
	global.ZC_LOG.Info("register table success")
}

// InitMysqlData Initialize data
func InitMysqlData() error {
	return service.ServiceGroupApp.SystemServiceGroup.InitDBService.InitData("mysql")
}
