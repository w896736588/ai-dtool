package ai_model

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"gitee.com/Sxiaobai/gs/v2/gshttp"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

type Bailian struct {
	apiKey      string
	turnsChat   bool //是否开启多轮会话，这里简单的使用每次对话传递上下文，不用多轮对话缓存空间
	messageList []_struct.Message
	model       string
	streamFunc  func(s string, err error)
}

func NewBailian(model, apiKey string, turnsChat bool, streamFunc func(s string, err error)) *Bailian {
	return &Bailian{
		apiKey:     apiKey,
		turnsChat:  turnsChat,
		model:      model,
		streamFunc: streamFunc,
	}
}

func (h *Bailian) Api(messageList []_struct.Message, tools []_struct.Tool) (string, error) {
	if h.turnsChat {
		h.messageList = append(h.messageList, messageList...)
	} else {
		h.messageList = messageList
	}
	base.Component.GsLog.Debugf("message count %d %s", len(h.messageList), gstool.JsonEncode(h.messageList))
	requestBody := _struct.RequestBody{
		Model:         h.model, //通义千问2.5-Coder-3B 模型列表：https://help.aliyun.com/zh/model-studio/getting-started/models
		Messages:      h.messageList,
		Tools:         tools,
		Stream:        true,
		StreamOptions: _struct.StreamOptions{IncludeUsage: false},
	}
	jsonData := gstool.JsonEncode(requestBody)
	if jsonData == `` {
		return ``, errors.New(`json encode error`)
	}
	cli := gshttp.PostJson(`https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions`).
		BodyStr(jsonData).
		Headers(map[string]string{
			"Authorization": "Bearer " + h.apiKey,
			"Content-Type":  "application/json",
		})
	var res []byte
	var resErr error
	if h.streamFunc != nil {
		res, resErr = cli.OpenStreamBytesEnd([]byte("\n\n"), h.streamFunc, func(bytes []byte) []byte {
			return bytes
		}).Request(200).Result()
	} else {
		res, resErr = cli.Request(200).Result()
	}
	base.Component.GsLog.Debugf(`结束对话后 %v`, resErr)
	if resErr == nil {
		base.Component.GsLog.Debugf(`结束对话后追加结果 %t`, h.turnsChat)
		if h.turnsChat {
			base.Component.GsLog.Debugf(`结束对话后添加`)
			h.messageList = append(h.messageList, _struct.Message{
				Role:    define.RoleAssistant,
				Content: cast.ToString(res),
			})
		}
	}
	return cast.ToString(res), resErr
}

func (h *Bailian) MessageList() []_struct.Message {
	return h.messageList
}
