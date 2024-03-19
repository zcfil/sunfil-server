package response

import (
	"time"
)

type SysContractWarnNode struct {
	ContractAddress string  `json:"abiId"`
	MinerId         string  `json:"abiName"`
	NodeBalance     float64 `json:"nodeBalance"`
	NodeDebt        float64 `json:"nodeDebt"`
	NodeAvailable   float64 `json:"nodeAvailable"`
	DebtRate        float64 `json:"abiContent"`
	LiquidateRate   float64 `json:"liquidateRate"`
	NodeDebtRate    float64 `json:"nodeDebtRate"`
	WarnTime        string  `json:"WarnTime"`
	AgreeCount      float64 `json:"agreeCount"`
	RejectCount     float64 `json:"rejectCount"`
	AbstainCount    float64 `json:"abstainCount"`
	DurationTime    string  `json:"durationTime"`
}

type OldWarnNodeInfo struct {
	ContractAddress string
	WarnBeginTime   time.Time
	WarnEndTime     time.Time
	VoteBeginTime   time.Time
	VoteEndTime     time.Time
	AgreeCount      float64
	RejectCount     float64
	AbstainCount    float64
	TotalVote       float64
	NodeStatus      int // Node status 1 in alarm, 2 in voting, 3 ended
	ModifyControl   int // Change Control 1 Yes, 0 No
	Solidify        int // Is it cured? 1 Yes, 0 No
	VoteStatus      int
}
