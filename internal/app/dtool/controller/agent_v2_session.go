package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// AgentV2SessionList 列出会话
func AgentV2SessionList(c *gin.Context) {
	var req struct {
		AgentId     int `json:"agent_id"`
		WorkspaceId int `json:"workspace_id"`
	}
	c.ShouldBindJSON(&req)

	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_v2_session WHERE agent_id = ? AND workspace_id = ? ORDER BY updated_at DESC`, req.AgentId, req.WorkspaceId,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	items := make([]define.AgentV2Session, 0, len(rows))
	for _, row := range rows {
		items = append(items, define.AgentV2Session{
			Id:          cast.ToInt(row["id"]),
			AgentId:     cast.ToInt(row["agent_id"]),
			WorkspaceId: cast.ToInt(row["workspace_id"]),
			Name:        cast.ToString(row["name"]),
			SessionDir:  cast.ToString(row["session_dir"]),
			Status:      cast.ToString(row["status"]),
			CreatedAt:   cast.ToInt64(row["created_at"]),
			UpdatedAt:   cast.ToInt64(row["updated_at"]),
		})
	}

	gsgin.GinResponseSuccess(c, "", gin.H{"list": items})
}

// AgentV2SessionCreate 创建会话
func AgentV2SessionCreate(c *gin.Context) {
	var req define.AgentV2SessionSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	now := time.Now().Unix()
	name := req.Name
	if name == "" {
		name = time.Now().Format("2006-01-02 15:04:05")
	}

	// 先插入会话（session_dir 留空，后续根据实际 sessionId 更新）
	lastId, err := common.DbMain.Client.InsertBySql(
		`INSERT INTO tbl_agent_v2_session (agent_id, workspace_id, name, session_dir, status, created_at, updated_at)
		 VALUES (?, ?, ?, '', 'active', ?, ?)`,
		req.AgentId, req.WorkspaceId, name, now, now,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	newId := cast.ToInt(lastId)

	// 根据实际 sessionId 计算正确的 session_dir 并立即更新
	sessionDir := computeSessionDirFromAgent(req.AgentId, newId)
	if sessionDir != "" {
		if _, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_v2_session SET session_dir = ? WHERE id = ?`,
			sessionDir, newId,
		).Exec(); err != nil {
			log.Printf("[agent-v2] update session_dir error: %v", err)
		}
	}

	gsgin.GinResponseSuccess(c, "", gin.H{
		"id":          newId,
		"name":        name,
		"session_dir": sessionDir,
	})
}

// computeSessionDirFromAgent 从 Agent 配置中读取 session_dir 基础路径
func computeSessionDirFromAgent(agentId, sessionId int) string {
	agentRow, err := common.DbMain.Client.QueryBySql(
		`SELECT config FROM tbl_agent_v2 WHERE id = ?`, agentId,
	).One()
	if err != nil || len(agentRow) == 0 {
		return ""
	}

	var cfg struct {
		SessionDir string `json:"session_dir"`
	}
	configStr := cast.ToString(agentRow["config"])
	if configStr != "" {
		json.Unmarshal([]byte(configStr), &cfg)
	}

	if cfg.SessionDir != "" {
		if sessionId > 0 {
			return filepath.Join(cfg.SessionDir, "s"+cast.ToString(sessionId))
		}
		return cfg.SessionDir
	}

	// 默认目录
	dir := filepath.Join(define.DefaultPiSessionDir, cast.ToString(agentId))
	if sessionId > 0 {
		dir = filepath.Join(dir, "s"+cast.ToString(sessionId))
	}
	return dir
}

// AgentV2SessionDelete 删除会话
func AgentV2SessionDelete(c *gin.Context) {
	var req struct {
		Id int `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	// 先查 session_dir，删除持久化文件
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT session_dir FROM tbl_agent_v2_session WHERE id = ?`, req.Id,
	).One()
	if err != nil {
		log.Printf("[agent-v2] query session_dir before delete session %d error: %v", req.Id, err)
	}
	if row != nil {
		sessionDir := cast.ToString(row["session_dir"])
		if sessionDir != "" {
			if err := os.RemoveAll(sessionDir); err != nil {
				log.Printf("[agent-v2] remove session dir %s error: %v", sessionDir, err)
			}
		}
	}

	_, err = common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_v2_session WHERE id = ?`, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// AgentV2SessionRename 重命名会话
func AgentV2SessionRename(c *gin.Context) {
	var req struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	now := time.Now().Unix()
	_, err := common.DbMain.Client.ExecBySql(
		`UPDATE tbl_agent_v2_session SET name = ?, updated_at = ? WHERE id = ?`,
		req.Name, now, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// AgentV2SessionMessages 获取会话消息历史（从持久化的 JSONL 文件读取）
func AgentV2SessionMessages(c *gin.Context) {
	var req struct {
		SessionId int `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	// 查询会话信息
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT session_dir FROM tbl_agent_v2_session WHERE id = ?`, req.SessionId,
	).One()
	if err != nil || len(row) == 0 {
		gsgin.GinResponseError(c, "会话不存在", nil)
		return
	}

	sessionDir := cast.ToString(row["session_dir"])
	messages := readSessionMessagesList(sessionDir)
	gsgin.GinResponseSuccess(c, "", gin.H{"messages": messages})
}

// readSessionMessagesList 从 JSONL 文件读取并转换为前端友好格式
// 支持两种格式：
// 1. 简化格式（旧版兼容）：{"type":"user","message":"..."}, {"type":"assistant","message":"..."}
// 2. Pi 原始事件流：agent_start, message_update, message_end, agent_end 等
func readSessionMessagesList(sessionDir string) []map[string]interface{} {
	if sessionDir == "" {
		return []map[string]interface{}{}
	}

	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return []map[string]interface{}{}
	}

	// 先收集所有 dtool_*.jsonl 事件行（按文件名排序保证顺序）
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	var allEvents []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "dtool_") || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(sessionDir, entry.Name()))
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			var raw map[string]interface{}
			if err := json.Unmarshal([]byte(line), &raw); err != nil {
				continue
			}
			allEvents = append(allEvents, raw)
		}
	}

	// 检测格式：如果全部事件都是简化格式（user/assistant/tool），直接返回
	allSimple := true
	for _, raw := range allEvents {
		t := cast.ToString(raw["type"])
		if t != "user" && t != "user_text" && t != "assistant" && t != "assistant_text" && t != "tool_call" && t != "tool_result" {
			allSimple = false
			break
		}
	}
	if allSimple {
		var messages []map[string]interface{}
		for _, raw := range allEvents {
			if msg := convertSimpleFormat(raw); msg != nil {
				messages = append(messages, msg)
			}
		}
		return messages
	}

	// 原始 Pi 事件流：重建消息
	return reconstructMessagesFromPiEvents(allEvents)
}

// convertSimpleFormat 处理简化格式（旧版兼容）
func convertSimpleFormat(raw map[string]interface{}) map[string]interface{} {
	msgType := cast.ToString(raw["type"])
	switch msgType {
	case "user_text", "user":
		return map[string]interface{}{
			"role":    "user",
			"content": cast.ToString(raw["message"]),
		}
	case "assistant_text", "assistant":
		content := cast.ToString(raw["message"])
		thinking := cast.ToString(raw["thinking"])
		return map[string]interface{}{
			"role":     "assistant",
			"content":  content,
			"thinking": thinking,
		}
	case "tool_call", "tool_result":
		return map[string]interface{}{
			"role":        "tool",
			"tool_name":   cast.ToString(raw["name"]),
			"tool_input":  raw["input"],
			"tool_output": raw["output"],
			"status":      cast.ToString(raw["status"]),
		}
	}
	return nil
}

// reconstructMessagesFromPiEvents 从 Pi 原始事件流重建用户/助手消息
// 模拟前端 handlePiEvent 的重建逻辑
func reconstructMessagesFromPiEvents(events []map[string]interface{}) []map[string]interface{} {
	var messages []map[string]interface{}
	var streamingText, streamingThinking string
	var pendingToolCalls []map[string]interface{}
	pendingToolCallMap := make(map[string]map[string]interface{})
	needPushAssistant := false

	for _, raw := range events {
		evtType := cast.ToString(raw["type"])
		switch evtType {
		case "user_text", "user":
			messages = append(messages, map[string]interface{}{
				"role":    "user",
				"content": cast.ToString(raw["message"]),
			})

		case "agent_start":
			streamingText = ""
			streamingThinking = ""
			pendingToolCalls = nil
			pendingToolCallMap = make(map[string]map[string]interface{})
			needPushAssistant = false

		case "message_update":
			needPushAssistant = handlePiMessageUpdate(raw, &streamingText, &streamingThinking,
				&pendingToolCalls, pendingToolCallMap) || needPushAssistant

		case "message_end":
			streamingText, streamingThinking, pendingToolCalls, pendingToolCallMap, needPushAssistant =
				handlePiMessageEnd(raw, &messages, streamingText, streamingThinking,
					pendingToolCalls, pendingToolCallMap, needPushAssistant)

		case "agent_end":
			streamingText, streamingThinking, pendingToolCalls, pendingToolCallMap, needPushAssistant =
				handlePiAgentEnd(&messages, streamingText, streamingThinking,
					pendingToolCalls, pendingToolCallMap, needPushAssistant)

		case "tool_execution_start":
			handlePiToolExecStart(raw, pendingToolCallMap)

		case "tool_execution_update":
			handlePiToolExecUpdate(raw, pendingToolCallMap)

		case "tool_execution_end":
			handlePiToolExecEnd(raw, messages, pendingToolCallMap)
		}
	}

	return messages
}

// handlePiMessageUpdate 处理 message_update 事件，更新流式文本/思考和工具调用
func handlePiMessageUpdate(raw map[string]interface{}, streamingText, streamingThinking *string,
	pendingToolCalls *[]map[string]interface{}, pendingToolCallMap map[string]map[string]interface{}) bool {

	msgEvt, _ := raw["assistantMessageEvent"].(map[string]interface{})
	if msgEvt == nil {
		return false
	}

	deltaType := cast.ToString(msgEvt["type"])
	switch deltaType {
	case "text_delta":
		*streamingText += cast.ToString(msgEvt["delta"])
		return true
	case "thinking_delta":
		*streamingThinking += cast.ToString(msgEvt["delta"])
		return true
	case "toolcall_start", "toolcall_delta", "toolcall_end":
		// 格式1: Anthropic — msgEvt.toolCall 直接携带
		tc, _ := msgEvt["toolCall"].(map[string]interface{})
		if tc != nil {
			tcId := cast.ToString(tc["id"])
			if tcId != "" {
				if existing, ok := pendingToolCallMap[tcId]; ok {
					if args := cast.ToString(tc["arguments"]); args != "" {
						if _, isStr := tc["arguments"].(string); isStr {
							existing["input"] = args
						} else {
							existing["input"] = tc["arguments"]
						}
					}
				} else {
					ti := map[string]interface{}{
						"id":     tcId,
						"name":   cast.ToString(tc["name"]),
						"status": "running",
					}
					if args := tc["arguments"]; args != nil {
						if argsStr, ok := args.(string); ok && argsStr != "" {
							ti["input"] = argsStr
						} else {
							ti["input"] = args
						}
					}
					pendingToolCallMap[tcId] = ti
					*pendingToolCalls = append(*pendingToolCalls, ti)
				}
			}
		}
		// 格式2: DeepSeek/OpenAI — partial.content 数组中的 toolCall 块
		partial, _ := msgEvt["partial"].(map[string]interface{})
		if partial != nil {
			contentBlocks, _ := partial["content"].([]interface{})
			for _, blockRaw := range contentBlocks {
				block, _ := blockRaw.(map[string]interface{})
				if block == nil || cast.ToString(block["type"]) != "toolCall" {
					continue
				}
				blockId := cast.ToString(block["id"])
				if blockId == "" {
					continue
				}
				if existing, ok := pendingToolCallMap[blockId]; ok {
					args := block["arguments"]
					if args != nil {
						if argsStr, ok := args.(string); ok && argsStr != "" {
							existing["input"] = argsStr
						} else if argsObj, ok := args.(map[string]interface{}); ok && len(argsObj) > 0 {
							existing["input"] = argsObj
						}
					}
					if partialArgs := cast.ToString(block["partialArgs"]); partialArgs != "" {
						existing["input"] = partialArgs
					}
				} else {
					ti := map[string]interface{}{
						"id":     blockId,
						"name":   cast.ToString(block["name"]),
						"status": "running",
					}
					args := block["arguments"]
					if args != nil {
						if argsStr, ok := args.(string); ok && argsStr != "" {
							ti["input"] = argsStr
						} else if argsObj, ok := args.(map[string]interface{}); ok && len(argsObj) > 0 {
							ti["input"] = argsObj
						}
					}
					if partialArgs := cast.ToString(block["partialArgs"]); partialArgs != "" && ti["input"] == nil {
						ti["input"] = partialArgs
					}
					pendingToolCallMap[blockId] = ti
					*pendingToolCalls = append(*pendingToolCalls, ti)
				}
			}
		}
		return true
	}
	return false
}

// handlePiMessageEnd 处理 message_end 事件，将完成的助手消息推入消息列表
func handlePiMessageEnd(raw map[string]interface{}, messages *[]map[string]interface{},
	streamingText, streamingThinking string,
	pendingToolCalls []map[string]interface{}, pendingToolCallMap map[string]map[string]interface{},
	needPushAssistant bool,
) (string, string, []map[string]interface{}, map[string]map[string]interface{}, bool) {

	msg, _ := raw["message"].(map[string]interface{})
	if msg == nil {
		return streamingText, streamingThinking, pendingToolCalls, pendingToolCallMap, needPushAssistant
	}
	if cast.ToString(msg["role"]) != "assistant" {
		return streamingText, streamingThinking, pendingToolCalls, pendingToolCallMap, needPushAssistant
	}

	content := extractPiContentFromEvent(msg["content"])
	errorMsg := cast.ToString(msg["errorMessage"])
	if content != "" || errorMsg != "" || streamingThinking != "" || len(pendingToolCalls) > 0 {
		msgObj := map[string]interface{}{"role": "assistant"}
		if errorMsg != "" {
			msgObj["content"] = "**Error:** " + errorMsg
		} else {
			msgObj["content"] = content
		}
		if streamingThinking != "" {
			msgObj["thinking"] = streamingThinking
		}
		if len(pendingToolCalls) > 0 {
			msgObj["toolCalls"] = pendingToolCalls
		}
		*messages = append(*messages, msgObj)
		return "", "", nil, make(map[string]map[string]interface{}), false
	}
	return streamingText, streamingThinking, pendingToolCalls, pendingToolCallMap, needPushAssistant
}

// handlePiAgentEnd 处理 agent_end 事件，兜底推送未完成的助手消息
func handlePiAgentEnd(messages *[]map[string]interface{},
	streamingText, streamingThinking string,
	pendingToolCalls []map[string]interface{}, pendingToolCallMap map[string]map[string]interface{},
	needPushAssistant bool,
) (string, string, []map[string]interface{}, map[string]map[string]interface{}, bool) {

	if needPushAssistant && (streamingText != "" || len(pendingToolCalls) > 0) {
		msgObj := map[string]interface{}{
			"role":    "assistant",
			"content": streamingText,
		}
		if streamingThinking != "" {
			msgObj["thinking"] = streamingThinking
		}
		if len(pendingToolCalls) > 0 {
			msgObj["toolCalls"] = pendingToolCalls
		}
		*messages = append(*messages, msgObj)
	}
	return "", "", nil, make(map[string]map[string]interface{}), false
}

// handlePiToolExecStart 处理 tool_execution_start 事件
func handlePiToolExecStart(raw map[string]interface{}, pendingToolCallMap map[string]map[string]interface{}) {
	tcId := cast.ToString(raw["toolCallId"])
	if tcId == "" {
		tcId = cast.ToString(raw["id"])
	}
	if tc, ok := pendingToolCallMap[tcId]; ok {
		tc["status"] = "running"
	}
}

// handlePiToolExecUpdate 处理 tool_execution_update 事件
func handlePiToolExecUpdate(raw map[string]interface{}, pendingToolCallMap map[string]map[string]interface{}) {
	tcId := cast.ToString(raw["toolCallId"])
	if tcId == "" {
		tcId = cast.ToString(raw["id"])
	}
	if tc, ok := pendingToolCallMap[tcId]; ok {
		if output, ok := raw["output"]; ok {
			tc["output"] = cast.ToString(tc["output"]) + cast.ToString(output)
		}
	}
}

// handlePiToolExecEnd 处理 tool_execution_end 事件
func handlePiToolExecEnd(raw map[string]interface{}, messages []map[string]interface{},
	pendingToolCallMap map[string]map[string]interface{}) {

	tcId := cast.ToString(raw["toolCallId"])
	if tcId == "" {
		tcId = cast.ToString(raw["id"])
	}
	var output interface{}
	if o, ok := raw["output"]; ok {
		output = o
	} else if r, ok := raw["result"]; ok {
		output = r
	}
	if tc, ok := pendingToolCallMap[tcId]; ok {
		tc["status"] = "done"
		if output != nil {
			tc["output"] = output
		}
		syncToolCallToMessages(messages, tcId, tc)
	} else {
		updateTc := map[string]interface{}{"status": "done"}
		if output != nil {
			updateTc["output"] = output
		}
		syncToolCallToMessages(messages, tcId, updateTc)
	}
}

// extractPiContentFromEvent 从 Pi content 数组提取文本
// content 格式：[{"type":"text","text":"..."}]
func extractPiContentFromEvent(content interface{}) string {
	blocks, ok := content.([]interface{})
	if !ok {
		return ""
	}
	var parts []string
	for _, block := range blocks {
		b, ok := block.(map[string]interface{})
		if !ok {
			continue
		}
		if cast.ToString(b["type"]) == "text" {
			parts = append(parts, cast.ToString(b["text"]))
		}
	}
	return strings.Join(parts, "")
}

// syncToolCallToMessages 从后往前找到最近的包含此 toolCall 的助手消息，同步 status/output
func syncToolCallToMessages(messages []map[string]interface{}, tcId string, tc map[string]interface{}) {
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if cast.ToString(msg["role"]) != "assistant" {
			continue
		}
		toolCalls, _ := msg["toolCalls"].([]map[string]interface{})
		if toolCalls == nil {
			continue
		}
		for j := range toolCalls {
			if cast.ToString(toolCalls[j]["id"]) == tcId {
				toolCalls[j]["status"] = tc["status"]
				toolCalls[j]["output"] = tc["output"]
				return
			}
		}
	}
}

// 常见模型的默认上下文窗口大小（兜底）
// 优先使用 tbl_ai_model 表中的 context_size（由 parseModelsCtx 提供）
var defaultModelContextSizes = map[string]int{
	"claude-sonnet-4-20250514": 200000,
	"claude-haiku-4-20250514":  200000,
	"claude-opus-4-20250514":   200000,
	"gpt-4o":                   128000,
	"gpt-4o-mini":              128000,
	"deepseek-v4-flash":        128000,
	"deepseek-v3":              128000,
	"gemini-2.5-pro":           1048576,
	"gemini-2.0-flash":         1048576,
}

// lookupContextTotal 根据模型名查找上下文窗口大小
// 优先级：传入的 modelsCtx > 默认表 > 128000
func lookupContextTotal(model string, modelsCtx map[string]int) int {
	if modelsCtx != nil {
		if ctx, ok := modelsCtx[model]; ok && ctx > 0 {
			return ctx
		}
	}
	if ctx, ok := defaultModelContextSizes[model]; ok {
		return ctx
	}
	return 128000
}

// parseModelsCtx 从 Agent config JSON 中解析模型上下文大小映射
// 优先读 config.models_ctx（旧格式），否则从 tbl_ai_model 表查询
func parseModelsCtx(configStr string) map[string]int {
	result := make(map[string]int)

	// 先尝试从 config JSON 中解析（旧格式兼容）
	if configStr != "" {
		var cfg struct {
			ModelsCtx map[string]int `json:"models_ctx"`
		}
		if err := json.Unmarshal([]byte(configStr), &cfg); err == nil && len(cfg.ModelsCtx) > 0 {
			return cfg.ModelsCtx
		}
	}

	// 从 DB 查询所有 LLM 模型的 context_size
	if common.DbMain != nil && common.DbMain.Client != nil {
		rows, err := common.DbMain.Client.QueryBySql(
			`SELECT model, context_size FROM tbl_ai_model WHERE model_type = 'llm' AND status = 1`,
		).All()
		if err == nil {
			for _, row := range rows {
				model := cast.ToString(row["model"])
				ctx := cast.ToInt(row["context_size"])
				if model != "" && ctx > 0 {
					result[model] = ctx
				}
			}
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// computeSessionStats 从会话事件文件中提取真实 token 用量
// 不再用字符数估算，而是从 Pi events 的 usage 数据获取
// 返回前端 tokenStats 所需的字段：input_tokens, output_tokens, cached_input_tokens, context_used, context_total, total_cost
func computeSessionStats(sessionDir string, modelsCtx map[string]int) map[string]interface{} {
	stats := map[string]interface{}{
		"input_tokens":        0,
		"output_tokens":       0,
		"cached_input_tokens": 0,
		"context_used":        0,
		"context_total":       128000,
		"total_cost":          0,
	}
	if sessionDir == "" {
		return stats
	}

	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return stats
	}

	var totalInputTokens, totalOutputTokens, totalCachedTokens float64
	var totalCost float64
	var latestInputTokens int
	currentModel := ""

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "dtool_") || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(sessionDir, entry.Name()))
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			var raw map[string]interface{}
			if err := json.Unmarshal([]byte(line), &raw); err != nil {
				continue
			}

			// 只处理 message_end 事件
			if cast.ToString(raw["type"]) != "message_end" {
				continue
			}

			msg, ok := raw["message"].(map[string]interface{})
			if !ok {
				continue
			}

			// 提取当前使用的模型名和 provider
			if modelName := cast.ToString(msg["model"]); modelName != "" {
				currentModel = modelName
			}
			provider := cast.ToString(msg["provider"])

			// 只处理 assistant 角色（包含真实 usage 数据）
			if cast.ToString(msg["role"]) != "assistant" {
				continue
			}

			usage, ok := msg["usage"].(map[string]interface{})
			if !ok {
				continue
			}

			inputTokens := cast.ToFloat64(usage["input"])
			outputTokens := cast.ToFloat64(usage["output"])
			cacheRead := cast.ToFloat64(usage["cacheRead"])
			cacheWrite := cast.ToFloat64(usage["cacheWrite"])

			totalInputTokens += inputTokens
			totalOutputTokens += outputTokens
			totalCachedTokens += cacheRead
			// DeepSeek 的 usage.input 只计未缓存部分，需加上 cacheRead/cacheWrite 才是真实上下文用量
			var contextUsed int
			if strings.EqualFold(provider, "deepseek") {
				contextUsed = int(inputTokens + cacheRead + cacheWrite)
			} else {
				contextUsed = int(inputTokens)
			}
			if contextUsed > 0 {
				latestInputTokens = contextUsed
			}

			if costMap, ok := usage["cost"].(map[string]interface{}); ok {
				totalCost += cast.ToFloat64(costMap["total"])
			}
		}
	}

	// 确定 context_total
	contextTotal := lookupContextTotal(currentModel, modelsCtx)

	stats["input_tokens"] = int(totalInputTokens)
	stats["output_tokens"] = int(totalOutputTokens)
	stats["cached_input_tokens"] = int(totalCachedTokens)
	stats["context_used"] = latestInputTokens
	stats["context_total"] = contextTotal
	stats["total_cost"] = totalCost
	return stats
}
