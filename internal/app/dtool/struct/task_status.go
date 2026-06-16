package _struct

// TaskStatusSaveRequest 保存任务状态请求（新增/编辑）。
type TaskStatusSaveRequest struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

// TaskStatusDeleteRequest 删除任务状态请求。
type TaskStatusDeleteRequest struct {
	ID int `json:"id"`
}

// TaskStatusSortRequest 排序任务状态请求。
type TaskStatusSortRequest struct {
	IDs []int `json:"ids"`
}
