package api

// ListReq 分页请求参数
type ListReq struct {
	// 指定的游标位置 从0开始 (比如 1、（获取前一百条 after: 0 first: 100）2、（获取第101条到200条 after: 100 first: 100）)
	After int64 `json:"after"`
	// 从指定游标开始，获取多少个数据
	First int64 `json:"first"`
}
