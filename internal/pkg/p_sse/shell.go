package p_sse

import (
	"dev_tool/internal/pkg/p_define"

	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

// SseShell 通过ssh转发的消息
type SseShell struct {
	Sse             *gsgin.Sse //sse
	SseDistributeId string     //具体业务接收的id
	StopKey         any
	StopCall        func(any) bool
}

func (h *SseShell) Send(msg any, typs ...string) {
	if h == nil || h.Sse == nil {
		return
	}
	if h.StopCall != nil && h.StopCall(h.StopKey) {
		h.Sse.CleanMsg()
		return
	}
	typ := p_define.SseContentTypeMsg
	if len(typs) > 0 && typs[0] != `` {
		typ = typs[0]
	}
	sendData := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: h.SseDistributeId,
		Data:            msg,
		Type:            typ,
	})
	err := h.Sse.SendToChan(sendData)
	if err != nil {
		gstool.FmtPrintlnLogTime(`发送sse错误 %s`, err.Error())
	}
}

func (h *SseShell) CleanMsg() {
	if h == nil || h.Sse == nil {
		return
	}
	h.Sse.CleanMsg()
}
