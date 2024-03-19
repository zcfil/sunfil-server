package request

import "zcfil-server/model/common/request"

type NodeInfoReq struct {
	NodeId           string `json:"nodeId" form:"nodeId"`                     // Node ID
	Operator         string `json:"operator" form:"operator"`                 // Little Fox Address
	NodeDebt         string `json:"nodeDebt" form:"nodeDebt"`                 // Node debt
	MaxWithdrawLimit string `json:"maxWithdrawLimit" form:"maxWithdrawLimit"` // Withdrawal threshold (maximum withdrawable balance)
}

type NodeListReq struct {
	NodeId   string `json:"nodeId" form:"nodeId"`     // Node Name
	Status   int64  `json:"status" form:"status"`     // Node status: 1. Joined, 2. Resigned
	Operator string `json:"operator" form:"operator"` // Little Fox Address
	request.PageInfo
}

type NodeAddressReq struct {
	NodeId string `json:"nodeId" form:"nodeId"` // Node Name
}
