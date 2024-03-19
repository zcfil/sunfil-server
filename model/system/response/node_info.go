package response

type NodeRepayInfo struct {
	OutstandingDebt string `json:"outstandingDebt"`
	MaxLoanLimit    string `json:"maxLoanLimit"`
}

type NodeWithdrawalInfo struct {
	MaxWithdrawBalance string `json:"maxWithdrawBalance"`
}

type NodeDetailInfo struct {
	NodeTotalBalance     float64 `json:"nodeTotalBalance"`
	DebtValue            float64 `json:"debtValue"`
	DebtRatio            float64 `json:"debtRatio"`
	WithdrawalThreshold  float64 `json:"withdrawalThreshold"`
	NodeOwnershipBalance float64 `json:"nodeOwnershipBalance"`
	MaxDebtRatio         float64 `json:"maxDebtRatio"`
	SettlementThreshold  float64 `json:"settlementThreshold"`
	NodeAvailableBalance float64 `json:"nodeAvailableBalance"`
	LockInRewards        float64 `json:"lockInRewards"`
	SectorPledge         float64 `json:"sectorPledge"`
	IsOperator           bool    `json:"isOperator"`
}

type NodeListResp struct {
	ID                   uint    `json:"id" gorm:"primarykey"`
	NodeName             int64   `json:"-" gorm:"column:node_name"`
	NodeId               string  `json:"nodeId" gorm:"-"`
	Status               int     `json:"status" gorm:"column:status"`
	Balance              float64 `json:"-" gorm:"column:balance"`
	NodeTotalBalance     float64 `json:"nodeTotalBalance" gorm:"-"`
	DebtBalance          float64 `json:"-" gorm:"column:debt_balance"`
	DebtValue            float64 `json:"debtValue" gorm:"-"`
	NodeOwnershipBalance float64 `json:"nodeOwnershipBalance" gorm:"-"`
	MaxDebtRatio         float64 `json:"maxDebtRatio" gorm:"column:max_debt_rate"`
	SettlementThreshold  float64 `json:"settlementThreshold" gorm:"column:liquidate_rate"`
	DebtRatio            float64 `json:"debtRatio" gorm:"-"`
	Operator             string  `json:"-" gorm:"column:operator"`
	IsOperator           bool    `json:"isOperator" gorm:"-"`
}
