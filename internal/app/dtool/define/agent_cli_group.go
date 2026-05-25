package define

// AgentCliGroupItem AgentCli 分组列表项
type AgentCliGroupItem struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// AgentCliGroupSaveRequest 新增/编辑分组请求
type AgentCliGroupSaveRequest struct {
	Id        int    `json:"id,omitempty"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order,omitempty"`
}

// AgentCliGroupDeleteRequest 删除分组请求
type AgentCliGroupDeleteRequest struct {
	Id int `json:"id"`
}

// AgentCliGroupRelSaveRequest 保存 AgentCli 与分组的关联关系请求
type AgentCliGroupRelSaveRequest struct {
	AgentCliId int   `json:"agent_cli_id"`
	GroupIds   []int `json:"group_ids"`
}
