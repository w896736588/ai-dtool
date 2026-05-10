package define

// McpType 常量定义支持的 MCP 类型
const (
	McpTypeChromeDevTools = "chrome-devtools"
)

// McpTypeDef 描述一个已知 MCP 类型
type McpTypeDef struct {
	Name        string
	PackageName string
	Command     string
	Description string
}

// McpTypeDefs 将 mcp_type 映射到其定义
var McpTypeDefs = map[string]McpTypeDef{
	McpTypeChromeDevTools: {
		Name:        "ChromeDevTools",
		PackageName: "chrome-devtools-mcp@latest",
		Command:     "npx",
		Description: "通过 Chrome DevTools Protocol 操作浏览器",
	},
}

// McpTypeItem MCP 类型列表中的一项
type McpTypeItem struct {
	McpType      string         `json:"mcp_type"`
	Name         string         `json:"name"`
	PackageName  string         `json:"package_name"`
	Description  string         `json:"description"`
	BindCountMap map[string]int `json:"bind_count_map"` // key=agent_target_id(string), value=绑定数
}

// McpBindingItem 绑定列表中的一项
type McpBindingItem struct {
	Id            int    `json:"id"`
	MappingId     int    `json:"mapping_id"`
	MappingKey    string `json:"mapping_key"`
	Label         string `json:"label"`
	UserDataIndex int    `json:"user_data_index"`
	IsBound       bool   `json:"is_bound"`
	BindingId     int    `json:"binding_id"`
	Instruction   string `json:"instruction,omitempty"`
	UserDataDir   string `json:"user_data_dir,omitempty"`
}

// McpAgentTargetItem 目标智能体列表中的一项
type McpAgentTargetItem struct {
	Id             int    `json:"id"`
	AgentName      string `json:"agent_name"`
	ConfigFilename string `json:"config_filename"`
	ConfigDir      string `json:"config_dir"`
	CreateTime     int64  `json:"create_time"`
	UpdateTime     int64  `json:"update_time"`
}

// McpRequest MCP 相关接口的通用请求体
type McpRequest struct {
	McpType       string `json:"mcp_type,omitempty"`
	AgentTargetId int    `json:"agent_target_id,omitempty"`
	MappingId     int    `json:"mapping_id,omitempty"`
	BindingId     int    `json:"binding_id,omitempty"`
}

// McpAgentTargetRequest 目标智能体 CRUD 的请求体
type McpAgentTargetRequest struct {
	Id             int    `json:"id,omitempty"`
	AgentName      string `json:"agent_name,omitempty"`
	ConfigFilename string `json:"config_filename,omitempty"`
	ConfigDir      string `json:"config_dir,omitempty"`
}
