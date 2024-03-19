package response

import "time"

type NodeRecordResp struct {
	Cid          string    `json:"cid" gorm:"column:cid"`
	FromAddr     string    `json:"fromAddr" gorm:"column:from_addr"`
	ToAddr       string    `json:"toAddr" gorm:"column:to_addr"`
	NodeId       uint64    `json:"-" gorm:"column:actor_id"`
	ActorId      string    `json:"actorId" gorm:"-"`
	Amount       string    `json:"amount" gorm:"-"`
	OpContent    string    `json:"-"  gorm:"column:op_content" `
	OpType       int       `json:"opType" gorm:"column:op_type"` // Operation types: 1. onboarding, 2. resignation, 3. modifying operator, 4. withdrawal, 5. borrowing, 6. repayment, 7. modifying worker, 8. modifying control, 9. beneficiary onboarding, 10. beneficiary resignation
	UpdatedAt    time.Time `json:"-" gorm:"column:updated_at"`
	TimeDuration string    `json:"timeDuration" gorm:"-"`
	NodeAction   int       `json:"nodeAction" gorm:"column:node_action"` // Node change: 0 operator operation, 1. Join, 2. Change, 3. Exit
}
