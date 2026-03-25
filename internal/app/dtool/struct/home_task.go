package _struct

// HomeTaskSaveRequest 保存首页任务请求。
type HomeTaskSaveRequest struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	TaskStatus string `json:"task_status"`
	Remark     string `json:"remark"`
	StartTime  int64  `json:"start_time"`
}

// HomeTaskListRequest 查询首页任务列表请求。
type HomeTaskListRequest struct {
	IsArchived int `json:"is_archived"`
}

// HomeTaskArchiveToggleRequest 切换首页任务归档状态请求。
type HomeTaskArchiveToggleRequest struct {
	ID         int `json:"id"`
	IsArchived int `json:"is_archived"`
}

// HomeTaskStatusQuickUpdateRequest 快捷切换首页任务状态请求。
type HomeTaskStatusQuickUpdateRequest struct {
	ID         int    `json:"id"`
	TaskStatus string `json:"task_status"`
}

// HomeTaskDeleteRequest 删除首页任务请求。
type HomeTaskDeleteRequest struct {
	ID int `json:"id"`
}
