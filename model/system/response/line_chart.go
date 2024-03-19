package response

import "time"

type TotalFilResp struct {
	RecordTime   time.Time `json:"recordTime"`
	TotalBalance float64   `json:"totalBalance"`
}

type StakeInfoResp struct {
	RecordTime   time.Time `json:"recordTime"`
	StakeBalance float64   `json:"stakeBalance"`
	StakeRate    float64   `json:"stakeRate"`
}

type TotalityChangeResp struct {
	RecordTime     time.Time `json:"recordTime"`
	FinanceUseRate float64   `json:"financeUseRate"`
	StakeRate      float64   `json:"stakeRate"`
	DebtRate       float64   `json:"debtRate"`
}

type GetTotalFilSideResp struct {
	StakeAPY             float64 `json:"stakeAPY"`
	LoanAPY              float64 `json:"loanAPY"`
	FinanceUseRate       float64 `json:"financeUseRate"`
	TotalBalance         float64 `json:"totalBalance"`
	FILValue             float64 `json:"filValue"`
	TotalBalanceValue    float64 `json:"totalBalanceValue"`
	TotalLockValue       float64 `json:"totalLockValue"`
	LastAvailableBalance float64 `json:"lastAvailableBalance"`
	StakeNum             int64   `json:"stakeNum"`
	NodeNum              int64   `json:"nodeNum"`
	BusinessCoefficient  float64 `json:"businessCoefficient"`
	RiskCoefficient      float64 `json:"riskCoefficient"`
}

type GetTotalRateSideResp struct {
	StakeAPY       float64 `json:"stakeAPY"`
	StakeQOQ       float64 `json:"stakeQOQ"`
	LoanAPY        float64 `json:"loanAPY"`
	LoanQOQ        float64 `json:"loanQOQ"`
	FinanceUseRate float64 `json:"financeUseRate"`
	FinanceUseQOQ  float64 `json:"financeUseQOQ"`
}

type GetLoanRateSideResp struct {
	LoanAPY      float64 `json:"loanAPY"`
	LoanQOQ      float64 `json:"loanQOQ"`
	TotalLoan    float64 `json:"totalLoan"`
	TotalLoanQOQ float64 `json:"totalLoanQOQ"`
	RemainFil    float64 `json:"remainFil"`
	RemainFilQOQ float64 `json:"remainFilQOQ"`
}

type GetStakeRateSideResp struct {
	StakeAPY      float64 `json:"stakeAPY"`
	StakeQOQ      float64 `json:"stakeQOQ"`
	TotalStake    float64 `json:"totalStake"`
	TotalStakeQOQ float64 `json:"totalStakeQOQ"`
}
