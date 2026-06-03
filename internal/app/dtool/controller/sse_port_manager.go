package controller

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"

	"github.com/gin-gonic/gin"
	"github.com/w896736588/go-tool/gsgin"
)

// 每个 SSE 端口允许的最大连接数
const MaxSseConnectionsPerPort = 5

// ssePortConnMap 记录每个 SSE 端口上的当前连接数
var (
	ssePortConnMap   = make(map[string]int)
	ssePortConnMutex sync.Mutex
)

// sseClientIdToPort 记录每个 clientId 对应的 SSE 端口（用于断开时减计数）
var sseClientIdToPort = make(map[string]string)

// IsSsePort 判断指定端口是否是 SSE 端口
func IsSsePort(port string) bool {
	for _, p := range component.EnvClient.SsePorts {
		if p == port {
			return true
		}
	}
	return false
}

// ssePortIncrement 增加 clientId 对应端口的连接计数
func ssePortIncrement(clientId, port string) {
	ssePortConnMutex.Lock()
	defer ssePortConnMutex.Unlock()
	ssePortConnMap[port]++
	sseClientIdToPort[clientId] = port
}

// ssePortDecrement 减少 clientId 对应端口的连接计数
func ssePortDecrement(clientId string) {
	ssePortConnMutex.Lock()
	defer ssePortConnMutex.Unlock()
	port, ok := sseClientIdToPort[clientId]
	if !ok {
		return
	}
	ssePortConnMap[port]--
	if ssePortConnMap[port] < 0 {
		ssePortConnMap[port] = 0
	}
	delete(sseClientIdToPort, clientId)
}

// ssePortConnCount 返回指定端口的当前连接数
func ssePortConnCount(port string) int {
	ssePortConnMutex.Lock()
	defer ssePortConnMutex.Unlock()
	return ssePortConnMap[port]
}

// ssePortIsAvailable 判断指定端口是否还有连接配额
func ssePortIsAvailable(port string) bool {
	return ssePortConnCount(port) < MaxSseConnectionsPerPort
}

// SseAvailablePort 接口：返回每个 SSE 端口的连接数信息
func SseAvailablePort(c *gin.Context) {
	ssePorts := component.EnvClient.SsePorts
	items := make([]map[string]any, 0, len(ssePorts))
	ssePortConnMutex.Lock()
	for _, port := range ssePorts {
		count := ssePortConnMap[port]
		items = append(items, map[string]any{
			`port`:      port,
			`count`:     count,
			`available`: count < MaxSseConnectionsPerPort,
		})
	}
	ssePortConnMutex.Unlock()
	gsgin.GinResponseSuccess(c, `获取成功`, map[string]any{
		`sse_ports`: items,
	})
}

// BuildSseOpenFunc 返回通用 SSE 连接的 openFunc，包含每端口连接数限制
func BuildSseOpenFunc(ssePort string) func(url.Values, chan int, *gin.Context) (*gsgin.Sse, error) {
	return func(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
		if !ssePortIsAvailable(ssePort) {
			return nil, errors.New(fmt.Sprintf(`SSE端口 %s 连接数已满（上限 %d）`, ssePort, MaxSseConnectionsPerPort))
		}
		clientId := urlValues.Get(`client_id`)
		sseC := gsgin.SseGetByClientId(clientId)
		if sseC != nil {
			return nil, errors.New(`已存在链接`)
		}
		sse := gsgin.SseRegister(clientId, stopC, c)
		// 记录该端口连接数+1
		ssePortIncrement(clientId, ssePort)
		// 发送连接建立事件
		_ = sse.SendToChan(define.SseConnect)
		BindShellConnectionsSSE(sse, stopC, 5*time.Second)
		BindAsyncTasksSSE(sse, stopC, 5*time.Second)
		BindMemoryFragmentStatusSSE(sse, stopC, 10*time.Second)
		BindGitPendingStatusSSE(sse, stopC, 5*time.Second)
		BindWorkflowUnreadSnapshotSSE(sse, stopC, 3*time.Second)
		BindConnectionCountSSE(sse, stopC, 5*time.Second)
		return sse, nil
	}
}

// BuildSseCloseFunc 返回通用 SSE 连接的 closeFunc，断开时减计数
func BuildSseCloseFunc() func(sse *gsgin.Sse) {
	return func(sse *gsgin.Sse) {
		if sse != nil {
			ssePortDecrement(sse.ClientId)
		}
		sse.UnRegister()
	}
}
