package response

type DebtRateAVGData struct {
	DebtRateAVG float64 `gorm:"column:debtRateAVG" json:"debtRateAVG"`
}

type DebtTotalAVGData struct {
	DebtTotalAVG float64 `gorm:"column:debtTotalAVG" json:"debtTotalAVG"`
}
