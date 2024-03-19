package system

import (
	"zcfil-server/service"
)

type ApiGroup struct {
	NodeInfoApi
	LiquidateApi
	MinerApi
	StakeInfoApi
	DebtInfoApi
	NodeRecordApi
}

var (
	NodeInfoService    = service.ServiceGroupApp.SystemServiceGroup.NodeInfoService
	liquidateService   = service.ServiceGroupApp.SystemServiceGroup.LiquidateService
	nodeQueryService   = service.ServiceGroupApp.SystemServiceGroup.NodeQueryService
	NodeRecordsService = service.ServiceGroupApp.SystemServiceGroup.NodeRecordsService
	stakeInfoService   = service.ServiceGroupApp.SystemServiceGroup.StakeInfoService
	debtInfoService    = service.ServiceGroupApp.SystemServiceGroup.DebtInfoService
)
