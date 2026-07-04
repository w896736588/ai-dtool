package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// AgentV2SessionList 列出会话
func AgentV2SessionList(c *gin.Context) {
	var req struct {
		AgentId int `json:"agent_id"`
	}
	c.ShouldBindJSON(&req)

	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_v2_session WHERE agent_id = ? ORDER BY updated_at DESC`, req.AgentId,
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
		common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_v2_session SET session_dir = ? WHERE id = ?`,
			sessionDir, newId,
		).Exec()
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
	dir := filepath.Join("data", "agent_sessions", cast.ToString(agentId))
	if sessionId > 0 {
		dir = filepath.Join(dir, "s"+cast.ToString(sessionId))
	}
	os.MkdirAll(dir, 0755)
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
	row, _ := common.DbMain.Client.QueryBySql(
		`SELECT session_dir FROM tbl_agent_v2_session WHERE id = ?`, req.Id,
	).One()
	if row != nil {
		sessionDir := cast.ToString(row["session_dir"])
		if sessionDir != "" {
			os.RemoveAll(sessionDir)
		}
	}

	_, err := common.DbMain.Client.ExecBySql(
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
			// user_text 事件已提供用户消息，agent_start 不再重复添加
			streamingText = ""
			streamingThinking = ""
			pendingToolCalls = nil
			pendingToolCallMap = make(map[string]map[string]interface{})
			needPushAssistant = false

		case "message_update":
			msgEvt, _ := raw["assistantMessageEvent"].(map[string]interface{})
			if msgEvt == nil {
				continue
			}
			deltaType := cast.ToString(msgEvt["type"])
			switch deltaType {
			case "text_delta":
				streamingText += cast.ToString(msgEvt["delta"])
				needPushAssistant = true
			case "thinking_delta":
				streamingThinking += cast.ToString(msgEvt["delta"])
				needPushAssistant = true
			case "toolcall_delta":
				tc, _ := msgEvt["toolCall"].(map[string]interface{})
				if tc != nil {
					tcId := cast.ToString(tc["id"])
					if tcId == "" {
						continue
					}
					if existing, ok := pendingToolCallMap[tcId]; ok {
						// 更新已有 toolCall 的参数
						if args := cast.ToString(tc["arguments"]); args != "" {
							_, isStr := tc["arguments"].(string)
							if isStr {
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
						pendingToolCalls = append(pendingToolCalls, ti)
					}
				}
				needPushAssistant = true
			}

		case "message_end":
			msg, _ := raw["message"].(map[string]interface{})
			if msg == nil {
				continue
			}
			role := cast.ToString(msg["role"])
			if role == "assistant" {
				content := extractPiContentFromEvent(msg["content"])
				errorMsg := cast.ToString(msg["errorMessage"])
				// 如果 message_end 已推送（含 content/error），agent_end 不再重复推送
				if content != "" || errorMsg != "" || streamingThinking != "" {
					msgObj := map[string]interface{}{
						"role": "assistant",
					}
					if errorMsg != "" {
						msgObj["content"] = "**Error:** " + errorMsg
					} else {
						msgObj["content"] = content
					}
					if streamingThinking != "" {
						msgObj["thinking"] = streamingThinking
						streamingThinking = ""
					}
					messages = append(messages, msgObj)
					needPushAssistant = false
					streamingText = ""
				}
			}

		case "agent_end":
			// 如果 message_end 未推送，用 streamingText + toolCalls 组装
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
				messages = append(messages, msgObj)
			}
			streamingText = ""
			streamingThinking = ""
			pendingToolCalls = nil
			pendingToolCallMap = make(map[string]map[string]interface{})
			needPushAssistant = false

		case "tool_execution_start":
			tcId := cast.ToString(raw["toolCallId"])
			if tcId == "" {
				tcId = cast.ToString(raw["id"])
			}
			if tc, ok := pendingToolCallMap[tcId]; ok {
				tc["status"] = "running"
			}

		case "tool_execution_update":
			tcId := cast.ToString(raw["toolCallId"])
			if tcId == "" {
				tcId = cast.ToString(raw["id"])
			}
			if tc, ok := pendingToolCallMap[tcId]; ok {
				if output, ok := raw["output"]; ok {
					tc["output"] = cast.ToString(tc["output"]) + cast.ToString(output)
				}
			}

		case "tool_execution_end":
			tcId := cast.ToString(raw["toolCallId"])
			if tcId == "" {
				tcId = cast.ToString(raw["id"])
			}
			if tc, ok := pendingToolCallMap[tcId]; ok {
				tc["status"] = "done"
				if output, ok := raw["output"]; ok {
					tc["output"] = output
				} else if result, ok := raw["result"]; ok {
					tc["output"] = result
				}
			}
		}
	}

	return messages
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
