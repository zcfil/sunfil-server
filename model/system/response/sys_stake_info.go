package response

type StackRateAVGData struct {
	StakeRateAVG float64 `gorm:"column:stakeRateAVG" json:"stakeRateAVG"`
}

type FinanceUseRateAVGData struct {
	FinanceUseRateAVG float64 `gorm:"column:financeUseRateAVG" json:"financeUseRateAVG"`
}

type ContractBalData struct {
	ContractBalAVG float64 `gorm:"column:contractBalAVG" json:"contractBalAVG"`
}

type StakeTotalData struct {
	StakeTotalAVG float64 `gorm:"column:stakeTotalAVG" json:"stakeTotalAVG"`
}
