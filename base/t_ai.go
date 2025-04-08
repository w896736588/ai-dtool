package base

import (
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

type TAi struct {
}

// ParseStream 解析标准流式数据
// 示例
// data: {"choices":[{"finish_reason":"stop","index":0,"delta":{"content":"","type":"text","role":"assistant"}}],"model":"","chunk_token_usage":0,"created":1743644846,"message_id":2,"parent_id":1}
func (h *TAi) ParseStream(msg string) []byte {
	msgList := strings.Split(msg, "\n")
	resBytes := make([]byte, 0)
	for _, msgVal := range msgList {
		if strings.HasPrefix(msgVal, `data: `) {
			msgVal = gstool.StringReplaces(msgVal, map[string]string{
				`data: `: ``,
			})
			if msgVal == "[DONE]" {
				gstool.FmtPrintlnLogTime(`返回结束`)
				return []byte("\n")
			}
			data := _struct.StreamData{}
			err := gstool.JsonDecode(msgVal, &data)
			if err != nil {
				gstool.FmtPrintlnLogTime(`报错了%s %s`, err.Error(), msgVal)
				return make([]byte, 0)
			}

			for _, choice := range data.Choices {
				resBytes = append(resBytes, []byte(choice.Delta.Content)...)
			}
		}
	}
	return resBytes
}
