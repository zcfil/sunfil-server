package response

type PageResult struct {
	Main     interface{} `json:"main"`
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

type DetailsResult struct {
	HeadInfo    interface{} `json:"headInfo"`
	PageDetails PageResult  `json:"pageDetails"`
}
