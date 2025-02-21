package ai_define

const RoleSystem = `system`
const RoleUser = `user`
const RoleAssistant = `assistant`

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type RequestBody struct {
	Model         string        `json:"model"`
	Messages      []Message     `json:"messages"`
	Tools         []Tool        `json:"tools,omitempty"`
	Stream        bool          `json:"stream"`
	StreamOptions StreamOptions `json:"stream_options"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string     `json:"name"`        //方法名
	Description string     `json:"description"` //做什么操作时调用这个function
	Parameters  Parameters `json:"parameters"`  //function 参数，不需要时不传
	Required    []string   `json:"required"`    //字段指定哪些参数为必填项
}

type Parameters struct {
	Type       string              `json:"type"`       //固定为object
	Properties map[string]Property `json:"properties"` //key为字段名称
}

type Property struct {
	Type        string `json:"type"`        //string
	Description string `json:"description"` //这个字段应该提取哪些值
}

type StreamData struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"delta"`
		Index        int         `json:"index"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason interface{} `json:"finish_reason"`
	} `json:"choices"`
	Object            string      `json:"object"`
	Usage             interface{} `json:"usage"`
	Created           int         `json:"created"`
	SystemFingerprint interface{} `json:"system_fingerprint"`
	Model             string      `json:"model"`
	Id                string      `json:"id"`
}
