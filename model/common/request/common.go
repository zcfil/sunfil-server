package request

// PageInfo Paging common input parameter structure
type PageInfo struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Keyword  string `json:"keyword" form:"keyword"`
}

// GetById Find by id structure
type GetById struct {
	ID int `json:"id" form:"id"`
}

type IdStatus struct {
	ID     int `json:"id" form:"id"`
	Status int `json:"status" form:"status"`
}

type ActorIdStatus struct {
	Actor  string `json:"actor" form:"actor"`
	ID     uint   `json:"id" form:"id"`
	Status int    `json:"status" form:"status"`
	Value  string `json:"value" form:"value"`
}

type IdCount struct {
	ID    int `json:"id" form:"id"`
	Count int `json:"count" form:"count"`
}

func (r *GetById) Uint() uint {
	return uint(r.ID)
}

type IdsReq struct {
	Ids []int `json:"ids" form:"ids"`
}

type RoomIdsReq struct {
	RoomIds []string `json:"roomIds" form:"roomIds"`
}

type UUIDsReq struct {
	UUIDs []string `json:"uuids" form:"uuids"`
}

// GetAuthorityId Get role by id structure
type GetAuthorityId struct {
	AuthorityId uint `json:"authorityId" form:"authorityId"`
}

type Empty struct{}

type UserInfo struct {
	UserId   string
	UserName string
}

type GetFoldLineListReq struct {
	TimeInterval string `json:"timeInterval" form:"timeInterval"`
}
