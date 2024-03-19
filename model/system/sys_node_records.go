package system

import "zcfil-server/global"

type SysNodeRecords struct {
	global.ZC_MODEL
	Cid            string `json:"cid" form:"cid" gorm:"index;unique;column:cid;comment:Transaction Hash"`
	FromAddr       string `json:"fromAddr" form:"fromAddr" gorm:"column:from_addr;comment:Transfer address"`
	ToAddr         string `json:"toAddr" form:"toAddr" gorm:"column:to_addr;comment:Transfer address"`
	ActorId        uint64 `json:"actorId" form:"actorId" gorm:"column:actor_id;comment:Node number"`
	Amount         string `json:"amount" form:"amount" gorm:"column:amount;comment:Amount"`
	ToContract     string `json:"toContract" form:"toContract" gorm:"column:to_contract;comment:Internal transaction address"`
	AmountContract string `json:"amountContract" form:"amountContract" gorm:"column:amount_contract;comment:Internal transaction amount"`
	OpType         int    `json:"opType" form:"opType" gorm:"column:op_type;comment:Operation types: 1. onboarding, 2. resignation, 3. modifying operator, 4. withdrawal, 5. borrowing, 6. repayment, 7. modifying worker, 8. modifying control, 9. beneficiary onboarding, 10. beneficiary resignation, 11. decision liquidation, 12. protection liquidation"`
	OpContent      string `json:"opContent"  form:"opContent" gorm:"column:op_content;type:text;comment:Operation content" `
	NodeAction     int    `json:"nodeAction" form:"nodeAction" gorm:"default:0;column:node_action;comment:Node change: 0 operator operation, 1. Join, 2. Change, 3. Exit"`
	IsMultisig     bool   `json:"isMultisig" form:"isMultisig" gorm:"default:0;column:is_multisig;comment:Wallet type: true multi signature wallet, false regular wallet"`
	TxId           int64  `json:"txId" form:"txId" gorm:"column:tx_id;comment:Multiple transaction ID"`
	Applied        bool   `json:"applied" form:"applied" gorm:"column:applied;comment:Effective or not: true has taken effect, false has not taken effect"`
}

func (SysNodeRecords) TableName() string {
	return "sys_node_records"
}
