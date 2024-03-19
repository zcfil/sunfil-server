package response

import (
	"time"
)

type NodeInfo struct {
	Available     string `json:"available"`
	Vesting       string `json:"vesting"`
	Pledge        string `json:"pledge"`
	Active        uint64 `json:"active"`
	Faulty        uint64 `json:"faulty"`
	Live          uint64 `json:"live"`
	TerminateFile string `json:"terminateFile"`
}

type WarnNodeInfo struct {
	Id              uint      `json:"id"`
	MinerId         string    `json:"minerId"`
	ContractAddress string    `json:"contractAddress"`
	NodeBalance     string    `json:"nodeBalance"`
	NodeDebt        string    `json:"nodeDebt"`
	NodeAvailable   string    `json:"nodeAvailable"`
	DebtRate        string    `json:"debtRate"`
	LiquidateRate   string    `json:"liquidateRate"`
	NodeDebtRate    string    `json:"nodeDebtRate"`
	WarnBeginTime   time.Time `json:"warnBeginTime"`
	VoteBeginTime   time.Time `json:"voteBeginTime"`
	VoteEndTime     time.Time `json:"voteEndTime"`
	RiskDuration    string    `json:"riskDuration"`
	AgreeCount      string    `json:"agreeCount"`
	RejectCount     string    `json:"rejectCount"`
	AbstainCount    string    `json:"abstainCount"`
	AgreeRate       string    `json:"agreeRate"`
	RejectRate      string    `json:"rejectRate"`
	AbstainRate     string    `json:"abstainRate"`
	DiffRate        string    `json:"diffRate"`
	VoteRate        string    `json:"voteRate"`
	VoteResults     int       `json:"voteResults"`
	ResultsRate     string    `json:"resultsRate"`
	Withdraw        string    `json:"withdraw"`
	VoteExpired     string    `json:"voteExpired"`
}

type WarnCount struct {
	Total int `json:"total"`
}
