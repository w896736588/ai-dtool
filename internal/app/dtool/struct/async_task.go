package _struct

// AsyncTaskListRequest 查询异步任务列表请求。 // AsyncTaskListRequest defines the async task list request payload.
type AsyncTaskListRequest struct {
	Limit int `json:"limit"`
}

// AsyncTaskInfoRequest 查询异步任务详情请求。 // AsyncTaskInfoRequest defines the async task info request payload.
type AsyncTaskInfoRequest struct {
	ID int `json:"id"`
}

// AsyncTaskDeleteRequest 删除异步任务请求。 // AsyncTaskDeleteRequest defines the async task delete request payload.
type AsyncTaskDeleteRequest struct {
	ID int `json:"id"`
}

// AsyncTaskActionRequest 执行异步任务结果动作请求。 // AsyncTaskActionRequest defines the async task action request payload.
type AsyncTaskActionRequest struct {
	ID     int    `json:"id"`
	Action string `json:"action"`
}
