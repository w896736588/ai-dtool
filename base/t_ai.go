package base

import (
	"dev_tool/internal/pkg/ai/ai_define"
	"gitee.com/Sxiaobai/gs/gstool"
)

type TAi struct {
}

// ParseStream 解析标准流式数据
// 示例
// data: {"choices":[{"finish_reason":"stop","index":0,"delta":{"content":"","type":"text","role":"assistant"}}],"model":"","chunk_token_usage":0,"created":1743644846,"message_id":2,"parent_id":1}
func (h *TAi) ParseStream(msg string) []byte {
	msg = gstool.StringReplaces(msg, map[string]string{
		`data: `: ``,
	})
	if msg == `[DONE]` {
		return make([]byte, 0)
	}
	data := ai_define.StreamData{}
	err := gstool.JsonDecode(msg, &data)
	if err != nil {
		return make([]byte, 0)
	}
	resBytes := make([]byte, 0)
	for _, choice := range data.Choices {
		resBytes = append(resBytes, []byte(choice.Delta.Content)...)
	}
	return resBytes
}
