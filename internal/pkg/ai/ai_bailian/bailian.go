package ai_bailian

import (
	"dev_tool/internal/pkg/ai/ai_define"
	"errors"
	"gitee.com/Sxiaobai/gs/gshttp"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

type Bailian struct {
	apiKey      string
	openBatch   bool //是否开启多轮会话，这里简单的使用每次对话传递上下文，不用多轮对话缓存空间
	messageList []ai_define.Message
	model       string
	streamFunc  func(s string, err error)
}

func NewBailian(model, apiKey string, openBatch bool, streamFunc func(s string, err error)) *Bailian {
	return &Bailian{
		apiKey:     apiKey,
		openBatch:  openBatch,
		model:      model,
		streamFunc: streamFunc,
	}
}

func (h *Bailian) Api(messageList []ai_define.Message, tools []ai_define.Tool) (string, error) {
	if h.openBatch {
		h.messageList = append(h.messageList, messageList...)
	} else {
		h.messageList = messageList
	}
	requestBody := ai_define.RequestBody{
		Model:         h.model, //通义千问2.5-Coder-3B 模型列表：https://help.aliyun.com/zh/model-studio/getting-started/models
		Messages:      h.messageList,
		Tools:         tools,
		Stream:        true,
		StreamOptions: ai_define.StreamOptions{IncludeUsage: true},
	}
	jsonData := gstool.JsonEncode(requestBody)
	if jsonData == `` {
		return ``, errors.New(`json encode error`)
	}
	res, resErr := gshttp.PostJson(`https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions`).
		BodyStr(jsonData).
		Headers(map[string]any{
			"Authorization": "Bearer " + h.apiKey,
			"Content-Type":  "application/json",
		}).OpenStream(h.streamFunc).Request(30).Result()
	if resErr != nil {
		if h.openBatch {
			h.messageList = append(h.messageList, ai_define.Message{
				Role:    ai_define.RoleAssistant,
				Content: cast.ToString(res),
			})
		}
	}
	return gstool.StringReplaces(cast.ToString(res), map[string]string{
		`\n`: "\n",
	}), resErr
}
