package controller

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

// fullpageSseConns 存储 Fullpage 业务 SSE 连接，key 为原始 clientID，value 为 *gsgin.Sse。
// 每个 clientID 只有一条连接，新连接会替换旧连接。
var fullpageSseConns sync.Map

// FullpageSseOpen 是 /sse/fullpage 的 SSE 连接建立回调，不绑定任何定时推送，仅保留通道供 ShellOutSetSeeId 使用。
func FullpageSseOpen(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
	clientID := strings.TrimSpace(urlValues.Get(`client_id`))
	if clientID == `` {
		return nil, fmt.Errorf(`client_id 不能为空`)
	}
	connID := fmt.Sprintf("fullpage_sse_%s_%d", clientID, time.Now().UnixNano())

	// 先关闭同 clientID 的旧连接，确保一 clientID 一连接
	if old, ok := fullpageSseConns.LoadAndDelete(clientID); ok {
		if oldSse, ok2 := old.(*gsgin.Sse); ok2 && oldSse != nil {
			gstool.FmtPrintlnLogTime(`[Fullpage-SSE] 替换旧连接 clientID=%s old_connID=%s`, clientID, oldSse.ClientId)
			oldSse.UnRegister()
		}
	}

	sse := gsgin.SseRegister(connID, stopC, c)
	fullpageSseConns.Store(clientID, sse)

	// 发送连接建立确认事件，便于前端确认通道已就绪
	_ = sse.SendToChan(`[CONNECT]`)

	gstool.FmtPrintlnLogTime(`[Fullpage-SSE] 连接已建立 clientID=%s connID=%s`, clientID, connID)
	return sse, nil
}

// FullpageSseClose 是 /sse/fullpage 的 SSE 连接关闭回调。
func FullpageSseClose(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	fullpageSseConns.Range(func(key, value any) bool {
		if value == sse {
			fullpageSseConns.Delete(key)
			gstool.FmtPrintlnLogTime(`[Fullpage-SSE] 连接关闭 connID=%s clientID=%s`, sse.ClientId, key)
			return false
		}
		return true
	})
	sse.UnRegister()
}

// GetFullpageSseByClientID 根据原始 clientID 获取对应的 SSE 连接。
func GetFullpageSseByClientID(clientID string) *gsgin.Sse {
	if v, ok := fullpageSseConns.Load(clientID); ok {
		if sse, ok2 := v.(*gsgin.Sse); ok2 {
			return sse
		}
	}
	return nil
}
