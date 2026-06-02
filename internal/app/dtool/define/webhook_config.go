package define

// Webhook 配置类型常量
const (
	WebhookTypeDingtalk = "dingtalk"
	WebhookTypeFeishu   = "feishu"
	WebhookTypeWecom    = "wecom"
)

// WebhookConfigItem 列表项
type WebhookConfigItem struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	WebhookUrl string `json:"webhook_url"`
	Secret     string `json:"secret"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

// WebhookConfigSaveRequest 新增/编辑请求
type WebhookConfigSaveRequest struct {
	Id         int    `json:"id,omitempty"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	WebhookUrl string `json:"webhook_url"`
	Secret     string `json:"secret,omitempty"`
}

// WebhookConfigDeleteRequest 删除请求
type WebhookConfigDeleteRequest struct {
	Id int `json:"id"`
}

// WebhookConfigTestRequest 测试发送请求
type WebhookConfigTestRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	WebhookUrl string `json:"webhook_url"`
	Secret     string `json:"secret,omitempty"`
	SingleURL  string `json:"single_url,omitempty"`
}
