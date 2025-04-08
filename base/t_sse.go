package base

import "C"
import (
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
)

type TSse struct {
	Sse *gsgin.TSse
}

func (h *TSse) SendMsg(sseClient, msg string) error {
	data := _struct.StreamData{
		Choices: []struct {
			Delta struct {
				Content string `json:"content"`
				Role    string `json:"role"`
			} `json:"delta"`
		}{
			{
				Delta: struct {
					Content string `json:"content"`
					Role    string `json:"role"`
				}{
					Content: msg,
					Role:    "",
				},
			},
		},
	}
	return h.Sse.Send(sseClient, `data: `+gstool.JsonEncode(data))
}
