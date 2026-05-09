package controller

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_define"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
)

// sendConnectionCountSnapshot 向指定 SSE 连接发送一次连接数快照（包含已用数和总数）
func sendConnectionCountSnapshot(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	count := len(gsgin.SseStatus())
	total := len(component.EnvClient.SsePorts) * MaxSseConnectionsPerPort
	err := sse.SendToChan(gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseConnectionCount,
		Data: map[string]any{
			`count`: count,
			`total`: total,
		},
		Type: p_define.SseContentTypeMsg,
	}))
	if err != nil {
		gstool.FmtPrintlnLogTime(`SseConnectionCount推送错误 %s`, err.Error())
	}
}

// BindConnectionCountSSE 为普通 SSE client 绑定连接数定时推送
func BindConnectionCountSSE(sse *gsgin.Sse, stopC chan int, interval time.Duration) {
	if sse == nil {
		return
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	// 建连后立即推一次
	sendConnectionCountSnapshot(sse)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sendConnectionCountSnapshot(sse)
			case <-stopC:
				return
			}
		}
	}()
}
