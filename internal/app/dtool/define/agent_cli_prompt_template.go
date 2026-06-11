package define

// AgentCliPromptTemplateItem AgentCli 提示词模板条目。
type AgentCliPromptTemplateItem struct {
	Id           int      `json:"id"`
	Name         string   `json:"name"`
	Content      string   `json:"content"`
	ApplyAllDirs bool     `json:"apply_all_dirs"`
	SortOrder    int      `json:"sort_order"`
	LocalDirs    []string `json:"local_dirs"`
	CreatedAt    int64    `json:"created_at"`
	UpdatedAt    int64    `json:"updated_at"`
}

// AgentCliPromptTemplateSaveRequest 保存提示词模板请求。
type AgentCliPromptTemplateSaveRequest struct {
	Id           int      `json:"id,omitempty"`
	Name         string   `json:"name"`
	Content      string   `json:"content"`
	ApplyAllDirs bool     `json:"apply_all_dirs"`
	SortOrder    int      `json:"sort_order,omitempty"`
	LocalDirs    []string `json:"local_dirs"`
}

// AgentCliPromptTemplateDeleteRequest 删除提示词模板请求。
type AgentCliPromptTemplateDeleteRequest struct {
	Id int `json:"id"`
}
