package define

// AppName 管家应用名，用于配置目录定位与日志标识。
const AppName = `dtool-butler`

// 消息角色常量
const (
	RoleUser      = `user`
	RoleAssistant = `assistant`
	RoleSystem    = `system`
)

// 任务状态常量
const (
	TaskStatusPending   = `pending`
	TaskStatusExecuting = `executing`
	TaskStatusVerifying = `verifying`
	TaskStatusDone      = `done`
	TaskStatusFailed    = `failed`
)

// BotConfigItem 钉钉机器人配置项，从共用库 tbl_butler_bot_config 读取。
type BotConfigItem struct {
	Id         int    `json:"id"`
	Platform   string `json:"platform"`
	Name       string `json:"name"`
	AppKey     string `json:"app_key"`
	AppSecret  string `json:"app_secret"`
	RobotCode  string `json:"robot_code"`
	WebhookUrl string `json:"webhook_url"`
	Secret     string `json:"secret"`
	Status     int    `json:"status"`
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

// HistoryMessage 历史消息记录，对应 tbl_butler_message。
type HistoryMessage struct {
	Id        int
	SessionId string
	Role      string
	Content   string
	Topic     string
	CreatedAt int64
}

// Env 管家运行时环境，从 dtool config.ini 读取数据库与记忆库路径。
type Env struct {
	RootPath      string
	ConfigPath    string
	ConfigFile    string
	DbPath        string
	DbName        string
	LogDbPath     string
	MemoryDbPath  string
	DatabaseUpDir string
	LogPath       string
}
