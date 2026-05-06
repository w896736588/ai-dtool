package _struct

// HomeTaskSaveRequest 保存首页任务请求。
type HomeTaskSaveRequest struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	TaskStatus       string `json:"task_status"`
	StartTime        int64  `json:"start_time"`
	MemoryFragmentID any    `json:"memory_fragment_id"`
	TapdUrl          string `json:"tapd_url"`
	GitID            int    `json:"git_id"`
	ApiDevEnabled    int    `json:"api_dev_enabled"`
	ApiCollectionID  int    `json:"api_collection_id"`
	ApiDirID         int    `json:"api_dir_id"`
	ApiHost          string `json:"api_host"`
	ApiToken         string `json:"api_token"`
	MysqlID          int    `json:"mysql_id"`
	GitIds           string `json:"git_ids"`
	ApiDevEntries    string `json:"api_dev_entries"`
	DevConfigs       string `json:"dev_configs"`
}

// DevConfig 开发配置条目，组合了 Git 仓库、接口集合/文件夹、Docker、MySQL 配置、自定义网页。
type DevConfig struct {
	GitID            int    `json:"git_id"`
	CollectionID     int    `json:"collection_id"`
	DirID            int    `json:"dir_id"`
	DockerID         int    `json:"docker_id"`
	MysqlID          int    `json:"mysql_id"`
	LocalDir         string `json:"local_dir"`
	ParentBranch     string `json:"parent_branch"`
	SmartLinkID      int    `json:"smart_link_id"`
	SmartLinkLabel   string `json:"smart_link_label"`
	SmartLinkAccount string `json:"smart_link_account"`
}

// ApiDevEntry 接口开发条目，对应一个集合+文件夹组合。
type ApiDevEntry struct {
	CollectionID int `json:"collection_id"`
	DirID        int `json:"dir_id"`
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
