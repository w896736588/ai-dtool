package controller

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"

	"github.com/gin-gonic/gin"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
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

	// 调试日志：输出当前所有活跃SSE客户端
	activeClients := gsgin.SseStatus()
	gstool.FmtPrintlnLogTime(`[SSE-Available] 所有活跃SSE客户端=%v 端口连接=%v`, activeClients, items)

	gsgin.GinResponseSuccess(c, `获取成功`, map[string]any{
		`sse_ports`: items,
	})
}

// BuildSseOpenFunc 返回通用 SSE 连接的 openFunc，包含每端口连接数限制
func BuildSseOpenFunc(ssePort string) func(url.Values, chan int, *gin.Context) (*gsgin.Sse, error) {
	return func(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
		currentCount := ssePortConnCount(ssePort)
		gstool.FmtPrintlnLogTime(`[SSE-Open] 开始建立SSE连接 port=%s current_count=%d max=%d`, ssePort, currentCount, MaxSseConnectionsPerPort)

		if !ssePortIsAvailable(ssePort) {
			gstool.FmtPrintlnLogTime(`[SSE-Open] 端口连接数已满 port=%s count=%d`, ssePort, currentCount)
			return nil, errors.New(fmt.Sprintf(`SSE端口 %s 连接数已满（上限 %d）`, ssePort, MaxSseConnectionsPerPort))
		}
		gstool.FmtPrintlnLogTime(`[SSE-Open] 端口连接数检查通过 port=%s`, ssePort)

		clientId := urlValues.Get(`client_id`)
		gstool.FmtPrintlnLogTime(`[SSE-Open] 获取client_id=%s port=%s`, clientId, ssePort)

		sseC := gsgin.SseGetByClientId(clientId)
		if sseC != nil {
			gstool.FmtPrintlnLogTime(`[SSE-Open] client_id已存在，拒绝重复连接 client_id=%s port=%s`, clientId, ssePort)
			return nil, errors.New(`已存在链接`)
		}
		gstool.FmtPrintlnLogTime(`[SSE-Open] client_id检查通过，无重复 client_id=%s port=%s`, clientId, ssePort)

		sse := gsgin.SseRegister(clientId, stopC, c)
		gstool.FmtPrintlnLogTime(`[SSE-Open] SseRegister完成 client_id=%s port=%s`, clientId, ssePort)

		// 记录该端口连接数+1
		ssePortIncrement(clientId, ssePort)
		gstool.FmtPrintlnLogTime(`[SSE-Open] 端口计数+1 port=%s new_count=%d`, ssePort, ssePortConnCount(ssePort))

		// 发送连接建立事件
		sendErr := sse.SendToChan(define.SseConnect)
		gstool.FmtPrintlnLogTime(`[SSE-Open] 发送CONNECT事件 client_id=%s port=%s err=%v`, clientId, ssePort, sendErr)

		BindShellConnectionsSSE(sse, stopC, 5*time.Second)
		gstool.FmtPrintlnLogTime(`[SSE-Open] BindShellConnectionsSSE完成 client_id=%s`, clientId)

		BindAsyncTasksSSE(sse, stopC, 5*time.Second)
		gstool.FmtPrintlnLogTime(`[SSE-Open] BindAsyncTasksSSE完成 client_id=%s`, clientId)

		BindMemoryFragmentStatusSSE(sse, stopC, 10*time.Second)
		gstool.FmtPrintlnLogTime(`[SSE-Open] BindMemoryFragmentStatusSSE完成 client_id=%s`, clientId)

		BindGitPendingStatusSSE(sse, stopC, 5*time.Second)
		gstool.FmtPrintlnLogTime(`[SSE-Open] BindGitPendingStatusSSE完成 client_id=%s`, clientId)

		BindWorkflowUnreadSnapshotSSE(sse, stopC, 3*time.Second)
		gstool.FmtPrintlnLogTime(`[SSE-Open] BindWorkflowUnreadSnapshotSSE完成 client_id=%s`, clientId)

		BindConnectionCountSSE(sse, stopC, 5*time.Second)
		gstool.FmtPrintlnLogTime(`[SSE-Open] BindConnectionCountSSE完成 client_id=%s 所有Bind完成`, clientId)

		return sse, nil
	}
}

// BuildSseCloseFunc 返回通用 SSE 连接的 closeFunc，断开时减计数
func BuildSseCloseFunc() func(sse *gsgin.Sse) {
	return func(sse *gsgin.Sse) {
		if sse != nil {
			gstool.FmtPrintlnLogTime(`[SSE-Close] SSE连接关闭 client_id=%s`, sse.ClientId)
			ssePortDecrement(sse.ClientId)
		}
		sse.UnRegister()
		gstool.FmtPrintlnLogTime(`[SSE-Close] UnRegister完成`)
	}
}

// sseClientIdStatusPrefix gsgin.SseStatus 返回值中 clientId 的前缀
const sseClientIdStatusPrefix = `ClientId:`

// SseConnectionDetails 返回服务端所有活跃 SSE 连接详情，供前端弹窗展示
func SseConnectionDetails(c *gin.Context) {
	statusList := gsgin.SseStatus()
	conns := make([]map[string]any, 0, len(statusList))
	for _, status := range statusList {
		connID := strings.TrimSpace(strings.TrimPrefix(status, sseClientIdStatusPrefix))
		if connID == `` {
			continue
		}
		connType, displayClientID := classifySseConnID(connID)
		conns = append(conns, map[string]any{
			`client_id`: displayClientID,
			`type`:      connType,
		})
	}
	gsgin.GinResponseSuccess(c, `获取成功`, map[string]any{
		`connections`: conns,
	})
}

// classifySseConnID 根据 connID 判断 SSE 连接类型并返回展示用的 clientID
// connID 规则：
//   - general:    原始 clientID（如 sse_client_id_xxx）
//   - agent_cli:  格式 "agent_cli_sse_<clientID>_<timestamp>"
//   - task_workflow: 格式 "task_workflow_sse_<clientID>_<timestamp>"
func classifySseConnID(connID string) (connType string, displayClientID string) {
	if strings.HasPrefix(connID, `agent_cli_sse_`) {
		return `Agent 客户端`, extractBusinessClientID(connID, `agent_cli_sse_`)
	}
	if strings.HasPrefix(connID, `task_workflow_sse_`) {
		return `Task Workflow`, extractBusinessClientID(connID, `task_workflow_sse_`)
	}
	if strings.HasPrefix(connID, `fullpage_sse_`) {
		return `Fullpage`, extractBusinessClientID(connID, `fullpage_sse_`)
	}
	return `通用 SSE`, connID
}

// extractBusinessClientID 从业务 SSE 的 connID 中提取原始 clientID
// connID 格式: "<prefix><clientID>_<timestamp>"
func extractBusinessClientID(connID string, prefix string) string {
	remain := strings.TrimPrefix(connID, prefix)
	lastIdx := strings.LastIndex(remain, `_`)
	if lastIdx > 0 {
		return remain[:lastIdx]
	}
	return remain
}
