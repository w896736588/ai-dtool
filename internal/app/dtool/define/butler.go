package define

// ButlerAppName 管家应用名，用于配置目录定位与日志标识。
const ButlerAppName = `dtool-butler`

// 消息角色常量
const (
	ButlerRoleUser      = `user`
	ButlerRoleAssistant = `assistant`
	ButlerRoleSystem    = `system`
)

// 任务状态常量
const (
	ButlerTaskStatusPending   = `pending`
	ButlerTaskStatusExecuting = `executing`
	ButlerTaskStatusVerifying = `verifying`
	ButlerTaskStatusDone      = `done`
	ButlerTaskStatusFailed    = `failed`
)

// 机器人连接状态常量
const (
	ConnStatusUnknown      = 0 // 未知/未连接
	ConnStatusConnected    = 1 // 已连接
	ConnStatusFailed       = 2 // 连接失败
	ConnStatusDisconnected = 3 // 连接断开
)

// BotConfigItem 钉钉机器人配置项，从共用库 tbl_butler_bot_config 读取。
// 纯流式机器人模式，不需要 webhook_url/secret。
type BotConfigItem struct {
	Id        int    `json:"id"`
	Platform  string `json:"platform"`
	Name      string `json:"name"`
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
	RobotCode string `json:"robot_code"`
	Status    int    `json:"status"`
}

// RoleItem 管家角色配置项，从 tbl_butler_role 读取。
type RoleItem struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Persona      string `json:"persona"`
	Tone         string `json:"tone"`
	SystemPrompt string `json:"system_prompt"`
	InitGreeting string `json:"init_greeting"`
	Status       int    `json:"status"`
}

// ButlerConfigItem 管家运行参数，从 tbl_butler_config 读取。
type ButlerConfigItem struct {
	Id                   int    `json:"id"`
	Name                 string `json:"name"`
	RoleId               int    `json:"role_id"`
	ModelId              int    `json:"model_id"`
	FcModelId            int    `json:"fc_model_id"`
	AgentCliId           int    `json:"agent_cli_id"`
	BotConfigId          int    `json:"bot_config_id"`
	ActiveTimeoutMinutes int    `json:"active_timeout_minutes"`
	MaxHistory           int    `json:"max_history"`
	AutoCleanOnNewTopic  int    `json:"auto_clean_on_new_topic"`
	IndexDocPath         string `json:"index_doc_path"`
	AutoInitOnStart      int    `json:"auto_init_on_start"`
	Status               int    `json:"status"`
}

// ButlerHistoryMessage 历史消息记录，对应 tbl_butler_message。
type ButlerHistoryMessage struct {
	Id        int
	SessionId string
	Role      string
	Content   string
	Topic     string
	CreatedAt int64
}

// ButlerEnv 管家运行时环境，从 dtool config.ini 读取数据库与记忆库路径。
type ButlerEnv struct {
	RootPath      string
	ConfigPath    string
	ConfigFile    string
	DbPath        string
	DbName        string
	LogDbPath     string
	MemoryDbPath  string
	DatabaseUpDir string
	LogPath       string
	DtoolBaseURL  string // dtool API 基地址（如 http://localhost:17170），供 http_call 工具使用
}
