package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	"dev_tool/internal/pkg/p_define"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Agent 和服务端在同一信任域内
	},
}

// AgentWsConnection 管理 Agent 的 WebSocket 连接
type AgentWsConnection struct {
	ClientID  string
	Conn      *websocket.Conn
	mu        sync.Mutex
	connected bool
}

// AgentWsManager 管理所有 Agent WebSocket 连接
type AgentWsManager struct {
	connections map[string]*AgentWsConnection
	mu          sync.RWMutex
}

var GlobalAgentWsManager = &AgentWsManager{
	connections: make(map[string]*AgentWsConnection),
}

// Register 注册新连接，踢掉同 clientID 的旧连接
func (m *AgentWsManager) Register(clientID string, conn *websocket.Conn) *AgentWsConnection {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 踢掉旧连接
	if old, ok := m.connections[clientID]; ok {
		old.Close()
	}

	c := &AgentWsConnection{
		ClientID:  clientID,
		Conn:      conn,
		connected: true,
	}
	m.connections[clientID] = c
	return c
}

// Unregister 移除连接
func (m *AgentWsManager) Unregister(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, clientID)
}

// Send 向指定 Agent 发送消息
func (m *AgentWsManager) Send(clientID string, msg define.AgentWsMessage) error {
	m.mu.RLock()
	c, ok := m.connections[clientID]
	m.mu.RUnlock()

	if !ok {
		return nil // 连接不存在
	}
	return c.Send(msg)
}

// GetConnection 获取指定 Agent 的连接
func (m *AgentWsManager) GetConnection(clientID string) *AgentWsConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connections[clientID]
}

// Send 向连接发送消息（线程安全）
func (c *AgentWsConnection) Send(msg define.AgentWsMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return nil
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.Conn.WriteMessage(websocket.TextMessage, data)
}

// Close 关闭连接
func (c *AgentWsConnection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.connected {
		c.connected = false
		_ = c.Conn.Close()
	}
}

// AgentWs 处理 Agent WebSocket 连接
func AgentWs(c *gin.Context) {
	clientID := c.Query("client_id")
	agentToken := c.Query("agent_token")

	gstool.FmtPrintlnLogTime(`AgentWs 请求 client_id=%s agent_token=%s remote=%s`, clientID, agentToken, c.Request.RemoteAddr)

	if clientID == "" {
		gstool.FmtPrintlnLogTime(`AgentWs 拒绝: client_id为空`)
		c.JSON(http.StatusOK, map[string]any{"ErrCode": 1, "ErrMsg": "client_id不能为空"})
		return
	}

	// 校验 agent_token（第一阶段简单校验：token 非空且在数据库中能找到该客户端）
	if agentToken == "" {
		gstool.FmtPrintlnLogTime(`AgentWs 拒绝: agent_token为空 client_id=%s`, clientID)
		c.JSON(http.StatusOK, map[string]any{"ErrCode": 1, "ErrMsg": "agent_token不能为空"})
		return
	}

	// 验证客户端是否已注册
	client, dbErr := common.DbMain.Client.QueryBySql(
		"SELECT * FROM tbl_smart_link_client WHERE client_id = ?", clientID,
	).One()
	if dbErr != nil {
		gstool.FmtPrintlnLogTime(`AgentWs DB查询错误 client_id=%s err=%s`, clientID, dbErr.Error())
	}
	if len(client) == 0 {
		gstool.FmtPrintlnLogTime(`AgentWs 拒绝: 客户端未注册 client_id=%s`, clientID)
		c.JSON(http.StatusOK, map[string]any{"ErrCode": 1, "ErrMsg": "客户端未注册"})
		return
	}

	// 验证 token（第一阶段：简单校验 token 字段等于 client_id + 时间戳 hash）
	storedToken := cast.ToString(client["agent_token"])
	if storedToken != "" && storedToken != agentToken {
		gstool.FmtPrintlnLogTime(`AgentWs 拒绝: token不匹配 stored=%s provided=%s`, storedToken, agentToken)
		c.JSON(http.StatusOK, map[string]any{"ErrCode": 1, "ErrMsg": "agent_token无效"})
		return
	}

	gstool.FmtPrintlnLogTime(`AgentWs 校验通过，准备升级 client_id=%s stored_token=%s`, clientID, storedToken)

	// 升级为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		gstool.FmtPrintlnLogTime(`WebSocket升级失败 %s`, err.Error())
		return
	}

	// 注册连接
	GlobalAgentWsManager.Register(clientID, conn)
	gstool.FmtPrintlnLogTime(`Agent WebSocket连接建立 client_id=%s`, clientID)

	// 更新数据库中的客户端状态
	now := time.Now().Unix()
	_, _ = common.DbMain.Client.QuickUpdate("tbl_smart_link_client", map[string]any{
		"client_id": clientID,
	}, map[string]any{
		"status":         define.SmartLinkClientStatusOnline,
		"last_seen_time": now,
		"update_time":    now,
	}).Exec()
	go BroadcastSmartLinkClientStatusUpdate()

	// 读消息循环
	defer func() {
		GlobalAgentWsManager.Unregister(clientID)
		conn.Close()
		gstool.FmtPrintlnLogTime(`Agent WebSocket连接断开 client_id=%s`, clientID)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				gstool.FmtPrintlnLogTime(`Agent WebSocket读错误 client_id=%s err=%s`, clientID, err.Error())
			}
			break
		}

		var msg define.AgentWsMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			gstool.FmtPrintlnLogTime(`Agent WebSocket消息解析失败 %s`, err.Error())
			continue
		}

		msg.ClientID = clientID // 确保使用已验证的 client_id

		switch msg.Type {
		case define.AgentWsMsgHello:
			handleAgentHello(clientID, msg)
		case define.AgentWsMsgHeartbeat:
			handleAgentHeartbeat(clientID, msg)
		case define.AgentWsMsgTaskLog:
			handleAgentTaskLog(msg)
		case define.AgentWsMsgTaskStatus:
			handleAgentTaskStatus(msg)
		case define.AgentWsMsgTaskResult:
			handleAgentTaskResult(msg)
		default:
			gstool.FmtPrintlnLogTime(`未知消息类型 %s`, msg.Type)
		}
	}
}

// handleAgentHello 处理 agent_hello 消息
func handleAgentHello(clientID string, msg define.AgentWsMessage) {
	gstool.FmtPrintlnLogTime(`Agent hello client_id=%s data=%v`, clientID, msg.Data)

	// 更新心跳时间
	now := time.Now().Unix()
	_, _ = common.DbMain.Client.QuickUpdate("tbl_smart_link_client", map[string]any{
		"client_id": clientID,
	}, map[string]any{
		"status":         define.SmartLinkClientStatusOnline,
		"last_seen_time": now,
		"update_time":    now,
	}).Exec()
	go BroadcastSmartLinkClientStatusUpdate()
}

// handleAgentHeartbeat 处理心跳消息
func handleAgentHeartbeat(clientID string, msg define.AgentWsMessage) {
	// 更新心跳时间
	now := time.Now().Unix()
	updateData := map[string]any{
		"last_seen_time": now,
		"update_time":    now,
	}

	// 如果有心跳数据，更新运行时状态
	if data, ok := msg.Data.(map[string]any); ok {
		if runtimeReady, ok := data["runtime_ready"].(bool); ok && runtimeReady {
			updateData["status"] = define.SmartLinkClientStatusOnline
		}
		if currentTaskID, ok := data["current_task_id"].(string); ok && currentTaskID != "" {
			updateData["status"] = define.SmartLinkClientStatusRunning
		}
	}

	_, _ = common.DbMain.Client.QuickUpdate("tbl_smart_link_client", map[string]any{
		"client_id": clientID,
	}, updateData).Exec()
	go BroadcastSmartLinkClientStatusUpdate()
}

// handleAgentTaskLog 处理 Agent 上报的日志，转发到 SSE
func handleAgentTaskLog(msg define.AgentWsMessage) {
	if msg.SseDistributeId == "" {
		return
	}

	logData, _ := json.Marshal(msg.Data)

	// 转发到前端 SSE
	sseMsg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: msg.SseDistributeId,
		Data:            string(logData),
		Type:            p_define.SseContentTypeMsg,
	})

	// 向所有 SSE 客户端广播
	for _, item := range gsgin.SseStatus() {
		clientID := strings.TrimSpace(strings.TrimPrefix(item, "ClientId:"))
		if clientID == "" || clientID == item {
			continue
		}
		sse := gsgin.SseGetByClientId(clientID)
		if sse != nil {
			_ = sse.SendToChan(sseMsg)
		}
	}
}

// handleAgentTaskStatus 处理 Agent 上报的状态
func handleAgentTaskStatus(msg define.AgentWsMessage) {
	if msg.TaskID == "" {
		return
	}

	// 更新数据库中的任务状态
	now := time.Now().Unix()
	statusData, _ := msg.Data.(map[string]any)
	status := ""
	if statusData != nil {
		status = cast.ToString(statusData["status"])
	}

	updateData := map[string]any{
		"status":      status,
		"update_time": now,
	}

	_, _ = common.DbMain.Client.QuickUpdate("tbl_smart_link_task", map[string]any{
		"task_id": msg.TaskID,
	}, updateData).Exec()

	// 同时通过 SSE 转发状态到前端
	if msg.SseDistributeId != "" {
		statusMsg := gstool.JsonEncode(p_define.SseData{
			SseDistributeId: msg.SseDistributeId,
			Data:            msg.Data,
			Type:            p_define.SseContentTypeMsg,
		})
		for _, item := range gsgin.SseStatus() {
			clientID := strings.TrimSpace(strings.TrimPrefix(item, "ClientId:"))
			if clientID == "" || clientID == item {
				continue
			}
			sse := gsgin.SseGetByClientId(clientID)
			if sse != nil {
				_ = sse.SendToChan(statusMsg)
			}
		}
	}
}

// handleAgentTaskResult 处理 Agent 上报的最终结果
func handleAgentTaskResult(msg define.AgentWsMessage) {
	if msg.TaskID == "" {
		return
	}

	now := time.Now().Unix()
	resultData, _ := msg.Data.(map[string]any)

	updateData := map[string]any{
		"status":        cast.ToString(resultData["status"]),
		"error_message": cast.ToString(resultData["error_message"]),
		"update_time":   now,
		"finish_time":   now,
	}

	_, _ = common.DbMain.Client.QuickUpdate("tbl_smart_link_task", map[string]any{
		"task_id": msg.TaskID,
	}, updateData).Exec()

	// 转发结果到 SSE
	if msg.SseDistributeId != "" {
		resultMsg := gstool.JsonEncode(p_define.SseData{
			SseDistributeId: msg.SseDistributeId,
			Data:            msg.Data,
			Type:            p_define.SseContentTypeMsg,
		})
		for _, item := range gsgin.SseStatus() {
			clientID := strings.TrimSpace(strings.TrimPrefix(item, "ClientId:"))
			if clientID == "" || clientID == item {
				continue
			}
			sse := gsgin.SseGetByClientId(clientID)
			if sse != nil {
				_ = sse.SendToChan(resultMsg)
			}
		}
	}

	// 更新客户端状态为在线
	GlobalAgentWsManager.Send(msg.ClientID, define.AgentWsMessage{
		Type: "task_complete_ack",
	})
}

// BuildAgentRunParams 从 PlaywrightRunParams 构造可序列化的 AgentRunParams（服务端使用）
func BuildAgentRunParams(runParams *plw.PlaywrightRunParams) define.AgentRunParams {
	return define.AgentRunParams{
		Id:                  runParams.Id,
		Link:                runParams.Link,
		LinkIdLabel:         runParams.LinkIdLabel,
		OpenNum:             runParams.OpenNum,
		Cookie:              runParams.Cookie,
		Headers:             runParams.Headers,
		OpenType:            int(runParams.OpenType),
		CombineType:         runParams.CombineType,
		ProcessList:         runParams.ProcessList,
		ReplaceList:         runParams.ReplaceList,
		BrowserAuthUsername: runParams.BrowserAuthUsername,
		BrowserAuthPassword: runParams.BrowserAuthPassword,
		Domain:              runParams.Domain,
		Scheme:              runParams.Scheme,
		LocatorTimeout:      runParams.LocatorTimeout,
		GetPageTimeout:      runParams.GetPageTimeout,
		LastIndexLabel:      runParams.LastIndexLabel,
		LinkId:              runParams.LinkId,
		DownloadFinds:       runParams.DownloadFinds,
		AutoCloseSecond:     runParams.AutoCloseSecond,
		Channel:             runParams.Channel,
		FilterUris:          runParams.FilterUris,
		ShowCookies:         runParams.ShowCookies,
	}
}
