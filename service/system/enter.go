package system

type ServiceGroup struct {
	InitDBService
	NodeInfoService
	LiquidateService
	StakeInfoService
	NodeQueryService
	NodeRecordsService
	DebtInfoService
	PledgeInfoService
}
