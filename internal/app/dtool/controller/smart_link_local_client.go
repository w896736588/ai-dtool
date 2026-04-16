package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	"dev_tool/internal/pkg/p_define"
	"fmt"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// SmartLinkRuntimeConfig 获取自定义网页运行时配置
func SmartLinkRuntimeConfig(c *gin.Context) {
	cfg := component.EnvClient.SmartLinkConfig
	if cfg == nil {
		cfg = &define.SmartLinkConfig{
			RunMode:       define.SmartLinkRunModeServer,
			ClientVersion: "1.0.0",
		}
	}

	baseURL := buildAgentDefaultServerURL(c.Request)
	downloadURLs := map[string]string{
		"windows": baseURL + "/api/agent/download?os=windows",
		"darwin":  baseURL + "/api/agent/download?os=darwin",
		"linux":   baseURL + "/api/agent/download?os=linux",
	}

	gsgin.GinResponseSuccess(c, "", map[string]any{
		"run_mode":                cfg.RunMode,
		"required_client_version": cfg.ClientVersion,
		"download_urls":           downloadURLs,
	})
}

// SmartLinkClientStatus 获取本地客户端状态
func SmartLinkClientStatus(c *gin.Context) {
	cfg := component.EnvClient.SmartLinkConfig
	if cfg == nil {
		cfg = &define.SmartLinkConfig{
			RunMode:       define.SmartLinkRunModeServer,
			ClientVersion: "1.0.0",
		}
	}

	client, queryErr := common.DbMain.Client.QueryBySql(`
		SELECT * FROM tbl_smart_link_client 
		ORDER BY last_seen_time DESC 
	`).One()
	if queryErr != nil {
		gstool.FmtPrintlnLogTime(`SmartLinkClientStatus 查询失败: %s`, queryErr.Error())
	}

	if len(client) == 0 {
		gsgin.GinResponseSuccess(c, "", map[string]any{
			"client_connected":     false,
			"client_status":        define.SmartLinkClientStatusOffline,
			"client_name":          "",
			"client_version":       "",
			"client_version_match": false,
			"client_last_seen_at":  0,
			"client_os":            "",
			"client_arch":          "",
		})
		return
	}

	lastSeen := cast.ToInt64(client["last_seen_time"])
	clientVersion := cast.ToString(client["client_version"])
	now := time.Now().Unix()
	isConnected := (now - lastSeen) < 30
	versionMatch := clientVersion == cfg.ClientVersion
	clientStatus := define.SmartLinkClientStatus(cast.ToString(client["status"]))

	if !isConnected {
		clientStatus = define.SmartLinkClientStatusOffline
	} else if !versionMatch {
		clientStatus = define.SmartLinkClientStatusVersionMismatch
	}

	gsgin.GinResponseSuccess(c, "", map[string]any{
		"client_connected":     isConnected,
		"client_status":        clientStatus,
		"client_name":          cast.ToString(client["client_name"]),
		"client_version":       clientVersion,
		"client_version_match": versionMatch,
		"client_last_seen_at":  lastSeen,
		"client_os":            cast.ToString(client["os"]),
		"client_arch":          cast.ToString(client["arch"]),
	})
}

// buildSmartLinkClientStatusPayload 构建客户端状态快照数据。
func buildSmartLinkClientStatusPayload() map[string]any {
	cfg := component.EnvClient.SmartLinkConfig
	if cfg == nil {
		cfg = &define.SmartLinkConfig{
			RunMode:       define.SmartLinkRunModeServer,
			ClientVersion: "1.0.0",
		}
	}

	client, queryErr := common.DbMain.Client.QueryBySql(`
		SELECT * FROM tbl_smart_link_client 
		ORDER BY last_seen_time DESC 
	`).One()
	if queryErr != nil {
		gstool.FmtPrintlnLogTime(`buildSmartLinkClientStatusPayload 查询失败: %s`, queryErr.Error())
	}

	if len(client) == 0 {
		return map[string]any{
			"client_connected":     false,
			"client_status":        define.SmartLinkClientStatusOffline,
			"client_name":          "",
			"client_version":       "",
			"client_version_match": false,
			"client_last_seen_at":  0,
			"client_os":            "",
			"client_arch":          "",
		}
	}

	lastSeen := cast.ToInt64(client["last_seen_time"])
	clientVersion := cast.ToString(client["client_version"])
	now := time.Now().Unix()
	isConnected := (now - lastSeen) < 30
	versionMatch := clientVersion == cfg.ClientVersion
	clientStatus := define.SmartLinkClientStatus(cast.ToString(client["status"]))

	if !isConnected {
		clientStatus = define.SmartLinkClientStatusOffline
	} else if !versionMatch {
		clientStatus = define.SmartLinkClientStatusVersionMismatch
	}

	return map[string]any{
		"client_connected":     isConnected,
		"client_status":        clientStatus,
		"client_name":          cast.ToString(client["client_name"]),
		"client_version":       clientVersion,
		"client_version_match": versionMatch,
		"client_last_seen_at":  lastSeen,
		"client_os":            cast.ToString(client["os"]),
		"client_arch":          cast.ToString(client["arch"]),
	}
}

// sendSmartLinkClientStatusSnapshot 向指定 SSE 连接发送一次客户端状态快照。
func sendSmartLinkClientStatusSnapshot(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	data := buildSmartLinkClientStatusPayload()
	err := sse.SendToChan(gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseSmartLinkClientStatus,
		Data:            data,
		Type:            p_define.SseContentTypeMsg,
	}))
	if err != nil {
		gstool.FmtPrintlnLogTime(`SmartLinkClientStatus广播错误 %s`, err.Error())
	}
}

// BindSmartLinkClientStatusSSE 为普通 SSE client 绑定本地客户端状态推送。
func BindSmartLinkClientStatusSSE(sse *gsgin.Sse, stopC chan int, interval time.Duration) {
	if sse == nil {
		return
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	// 建连后立即推一次，避免前端初次打开时要等下一个周期。
	sendSmartLinkClientStatusSnapshot(sse)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sendSmartLinkClientStatusSnapshot(sse)
			case <-stopC:
				return
			}
		}
	}()
}

// BroadcastSmartLinkClientStatusUpdate 主动广播客户端状态更新。
func BroadcastSmartLinkClientStatusUpdate() {
	msg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseSmartLinkClientStatus,
		Data:            buildSmartLinkClientStatusPayload(),
		Type:            p_define.SseContentTypeMsg,
	})
	for _, item := range gsgin.SseStatus() {
		clientID := strings.TrimSpace(strings.TrimPrefix(item, "ClientId:"))
		if clientID == "" || clientID == item {
			continue
		}
		sse := gsgin.SseGetByClientId(clientID)
		if sse == nil {
			continue
		}
		_ = sse.SendToChan(msg)
	}
}

// AgentRegister 客户端注册
func AgentRegister(c *gin.Context) {
	var req map[string]any
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误: "+err.Error(), nil)
		return
	}

	cfg := component.EnvClient.SmartLinkConfig
	if cfg == nil {
		cfg = &define.SmartLinkConfig{
			RunMode:       define.SmartLinkRunModeServer,
			ClientVersion: "1.0.0",
		}
	}

	clientID := cast.ToString(req["client_id"])
	clientVersion := cast.ToString(req["client_version"])
	now := time.Now().Unix()

	clientData := map[string]any{
		"client_id":        clientID,
		"client_name":      cast.ToString(req["hostname"]),
		"client_version":   clientVersion,
		"required_version": cfg.ClientVersion,
		"status":           define.SmartLinkClientStatusOnline,
		"host_name":        cast.ToString(req["hostname"]),
		"os":               cast.ToString(req["os"]),
		"arch":             cast.ToString(req["arch"]),
		"user_name":        cast.ToString(req["user_name"]),
		"last_seen_time":   now,
		"update_time":      now,
	}

	existing, _ := common.DbMain.Client.QuickQuery("tbl_smart_link_client", "*", map[string]any{
		"client_id": clientID,
	}).One()

	// 生成 agent_token（第一阶段：简单使用 client_id + 时间戳）
	agentToken := fmt.Sprintf("%s_%d", clientID, now)

	if len(existing) == 0 {
		clientData["create_time"] = now
		clientData["agent_token"] = agentToken
		_, _ = common.DbMain.Client.QuickCreate("tbl_smart_link_client", clientData).Exec()
	} else {
		clientData["agent_token"] = agentToken
		_, _ = common.DbMain.Client.QuickUpdate("tbl_smart_link_client", map[string]any{
			"client_id": clientID,
		}, clientData).Exec()
	}

	// 客户端注册后主动推送状态变更
	go BroadcastSmartLinkClientStatusUpdate()

	gsgin.GinResponseSuccess(c, "", map[string]any{
		"accepted":                true,
		"required_client_version": cfg.ClientVersion,
		"server_time":             now,
		"version_match":           clientVersion == cfg.ClientVersion,
		"agent_token":             agentToken,
	})
}

// AgentHeartbeat 客户端心跳
func AgentHeartbeat(c *gin.Context) {
	var req map[string]any
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}

	now := time.Now().Unix()
	updateData := map[string]any{
		"client_version": cast.ToString(req["client_version"]),
		"status":         cast.ToString(req["status"]),
		"host_name":      cast.ToString(req["hostname"]),
		"last_seen_time": now,
		"update_time":    now,
	}

	_, _ = common.DbMain.Client.QuickUpdate("tbl_smart_link_client", map[string]any{
		"client_id": cast.ToString(req["client_id"]),
	}, updateData).Exec()

	// 心跳后主动推送状态变更
	go BroadcastSmartLinkClientStatusUpdate()

	gsgin.GinResponseSuccess(c, "", nil)
}

// SmartLinkTaskCreate 创建本地执行任务（通过 WebSocket 下发给 Agent）
func SmartLinkTaskCreate(c *gin.Context) {
	var req map[string]any
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}

	cfg := component.EnvClient.SmartLinkConfig
	if cfg == nil {
		cfg = &define.SmartLinkConfig{
			RunMode:       define.SmartLinkRunModeServer,
			ClientVersion: "1.0.0",
		}
	}

	// 检查运行模式
	if cfg.RunMode != define.SmartLinkRunModeLocalClient {
		gsgin.GinResponseError(c, "当前运行模式不是本地客户端模式", nil)
		return
	}

	// 检查客户端状态
	client, _ := common.DbMain.Client.QueryBySql(`
		SELECT * FROM tbl_smart_link_client 
		ORDER BY last_seen_time DESC 
		LIMIT 1
	`).One()

	if len(client) == 0 {
		gsgin.GinResponseError(c, "SMART_LINK_CLIENT_OFFLINE", nil)
		return
	}

	lastSeen := cast.ToInt64(client["last_seen_time"])
	clientVersion := cast.ToString(client["client_version"])
	now := time.Now().Unix()

	if (now - lastSeen) >= 30 {
		gsgin.GinResponseError(c, "SMART_LINK_CLIENT_OFFLINE", nil)
		return
	}

	if clientVersion != cfg.ClientVersion {
		gsgin.GinResponseError(c, "SMART_LINK_CLIENT_VERSION_MISMATCH", nil)
		return
	}

	clientStatus := define.SmartLinkClientStatus(cast.ToString(client["status"]))
	if clientStatus == define.SmartLinkClientStatusPreparingRuntime {
		gsgin.GinResponseError(c, "SMART_LINK_CLIENT_PREPARING_RUNTIME", nil)
		return
	}

	clientID := cast.ToString(client["client_id"])

	// 检查 WebSocket 连接是否存在
	conn := GlobalAgentWsManager.GetConnection(clientID)
	if conn == nil {
		gsgin.GinResponseError(c, "SMART_LINK_CLIENT_WS_NOT_CONNECTED", nil)
		return
	}

	// 构建 PlaywrightRunParams（服务端查数据库构造完整参数）
	id := cast.ToInt(req["smart_link_id"])
	label := cast.ToString(req["label"])
	userName := cast.ToString(req["user_name"])
	password := cast.ToString(req["password"])
	openType := cast.ToInt(req["open_type"])
	openNum := cast.ToInt(req["open_num"])
	replaceList := make(map[string]string)

	runParams, runParamsErr := plw.GetRunParams(id, label, userName, password, openType, openNum, replaceList)
	if runParamsErr != nil {
		gsgin.GinResponseError(c, "构建运行参数失败: "+runParamsErr.Error(), nil)
		return
	}

	// 生成任务 ID 和 SSE 分发 ID
	taskID := "task_" + cast.ToString(now) + "_" + cast.ToString(id)
	sseDistributeId := cast.ToString(req["sse_distribute_id"])
	if sseDistributeId == "" {
		sseDistributeId = "smart_link_run_" + cast.ToString(now)
	}

	// 创建任务记录到数据库（用于状态追踪）
	_, createErr := common.DbMain.Client.QuickCreate("tbl_smart_link_task", map[string]any{
		"task_id":       taskID,
		"client_id":     clientID,
		"smart_link_id": id,
		"label":         label,
		"status":        define.SmartLinkTaskStatusPending,
		"run_mode":      define.SmartLinkRunModeLocalClient,
		"create_time":   now,
		"update_time":   now,
	}).Exec()
	if createErr != nil {
		gsgin.GinResponseError(c, "创建任务失败: "+createErr.Error(), nil)
		return
	}

	// 通过 WebSocket 下发任务给 Agent
	agentRunParams := BuildAgentRunParams(runParams)
	wsMsg := define.AgentWsMessage{
		Type:            define.AgentWsMsgTaskExecute,
		ClientID:        clientID,
		TaskID:          taskID,
		SseDistributeId: sseDistributeId,
		Data: define.AgentTaskExecuteData{
			TaskID:          taskID,
			SseDistributeId: sseDistributeId,
			ClientID:        clientID,
			RunParams:       agentRunParams,
		},
	}

	if sendErr := GlobalAgentWsManager.Send(clientID, wsMsg); sendErr != nil {
		gsgin.GinResponseError(c, "下发任务到Agent失败: "+sendErr.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", map[string]any{
		"task_id":           taskID,
		"client_id":         clientID,
		"status":            define.SmartLinkTaskStatusPending,
		"sse_distribute_id": sseDistributeId,
	})
}
