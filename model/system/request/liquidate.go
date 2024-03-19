package request

import "zcfil-server/model/common/request"

type WarnNodeReq struct {
	Id         int `json:"id" form:"id"`
	NodeStatus int `json:"nodeStatus" form:"nodeStatus"`
	request.PageInfo
}
