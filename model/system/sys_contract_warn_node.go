package system

import (
	"time"
	"zcfil-server/global"
)

type SysContractWarnNode struct {
	global.ZC_MODEL
	ContractAddress string    `gorm:"comment:Contract 0X Address ID" json:"contractAddress"`
	MinerId         string    `gorm:"comment:MinerId" json:"minerId"`
	NodeBalance     float64   `gorm:"comment:Total balance of nodes" json:"nodeBalance"`
	NodeDebt        float64   `gorm:"comment:Node debt" json:"nodeDebt"`
	NodeAvailable   float64   `gorm:"comment:Node available balance" json:"nodeAvailable"`
	DebtRate        float64   `gorm:"comment:Maximum debt ratio" json:"debtRate"`
	LiquidateRate   float64   `gorm:"comment:Settlement Rate " json:"liquidateRate"`
	NodeDebtRate    float64   `gorm:"comment:Node debt ratio" json:"nodeDebtRate"`
	Withdraw        float64   `gorm:"comment:Withdrawal threshold" json:"withdraw"`
	WarnBeginTime   time.Time `gorm:"comment:Alarm start time" json:"warnBeginTime"`
	WarnEndTime     time.Time `gorm:"comment:Alarm end time" json:"warnEndTime"`
	VoteBeginTime   time.Time `gorm:"comment:Voting start time" json:"voteBeginTime"`
	VoteEndTime     time.Time `gorm:"comment:Voting end time" json:"voteEndTime"`
	AgreeCount      float64   `gorm:"comment:Agree to vote counting" json:"agreeCount"`
	RejectCount     float64   `gorm:"comment:Oppose vote counting" json:"rejectCount"`
	AbstainCount    float64   `gorm:"comment:Abstention vote counting" json:"abstainCount"`
	TotalVote       float64   `gorm:"comment:Total number of votes" json:"totalVote"`
	VoteResults     int       `gorm:"comment:Voting results 1 agree, 2 oppose, and 3 abstain" json:"voteResults"`
	VoteStatus      int       `gorm:"comment:Proposal status 1 is effective, 2 is not effective" json:"voteStatus"`
	NodeStatus      int       `gorm:"comment:Node status 1 in alarm, 2 in voting, 3 in clearing, 4 closed" json:"nodeStatus"`
	ModifyControl   int       `gorm:"comment:Change Control 1 Yes, 0 No" json:"modifyControl"`
	Solidify        int       `gorm:"comment:Is it cured? 1 Yes, 0 No" json:"solidify"`
	Process         int       `gorm:"default:0;comment:Has it been cleared? Yes, 0 No" json:"process"`
}

func (SysContractWarnNode) TableName() string {
	return "sys_contract_warn_node"
}
