package base

import (
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

type TAi struct {
}

// ParseStream 解析流式数据
// 支持纳米AI
// 支持deepseek
func (h *TAi) ParseStream(url, msg string) []byte {
	msgList := strings.Split(msg, "\n")
	resBytes := make([]byte, 0)
	Component.GsLog.Errof(`收到消息---%v---`, msg)
	for _, msgVal := range msgList {
		if !strings.HasPrefix(msgVal, `data: `) {
			continue
		}
		if strings.Contains(url, `/api/assistant/chat`) { //纳米AI
			h.ParseBot(msgVal, &resBytes)
		} else if gstool.SContains(url, []string{`/api/v0/chat/completion`, `/api/GitLab`, `basic`}) { //git日志
			h.ParseDeepseek(msgVal, &resBytes)
		}
	}
	return resBytes
}

// ParseBot 纳米AI格式
func (h *TAi) ParseBot(msg string, resBytes *[]byte) {
	msg = gstool.StringReplaces(msg, map[string]string{
		`data: `: ``,
	})
	Component.GsLog.Errof(`解析：---%v---`, msg)
	if strings.HasPrefix(msg, `MESSAGEID`) {
		*resBytes = append(*resBytes, []byte(msg+"\n")...)
	} else {
		if msg == `` { //纳米AI 空可能表示这个消息结束
			*resBytes = append(*resBytes, []byte("\n")...)
		} else {
			*resBytes = append(*resBytes, []byte(msg)...)
		}
	}
}

func (h *TAi) ParseDeepseek(msg string, resBytes *[]byte) {
	msg = gstool.StringReplaces(msg, map[string]string{
		`data: `: ``,
	})
	if msg == "[DONE]" {
		*resBytes = append(*resBytes, []byte("\n")...)
		return
	}
	data := _struct.StreamData{}
	err := gstool.JsonDecode(msg, &data)
	if err != nil {
		Component.GsLog.Errof(`解析deepseek内容失败 --%s--`, msg)
	} else {
		for _, choice := range data.Choices {
			*resBytes = append(*resBytes, []byte(choice.Delta.Content)...)
		}
	}
}
