package request

import "zcfil-server/model/common/request"

type GetNodeRecordReq struct {
	NodeId string `json:"nodeId,required" form:"nodeId,required"` // Node ID
	OpType string `json:"opType,required" form:"opType,required"` // Node record types: Borrow Loan, Repayment Repayment, Withdrawal with Draw, NodeChange Node Change, Liquidation Clearing
	request.PageInfo
}
