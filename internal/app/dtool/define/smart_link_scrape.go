package define

// SmartLinkScrapeConfig 自定义网页抓取 Markdown 的运行配置。
// 该结构仍被服务端抓取流程复用，因此不能随 dtool-agent 一起删除。
type SmartLinkScrapeConfig struct {
	JumpURL     string `json:"jump_url"`
	CssSelector string `json:"css_selector"`
	WaitSeconds int    `json:"wait_seconds"`
}
