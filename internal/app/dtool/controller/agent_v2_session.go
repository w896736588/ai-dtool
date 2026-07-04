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

	// 从 agent config 中获取 session_dir 配置
	sessionDir := computeSessionDirFromAgent(req.AgentId, 0)

	lastId, err := common.DbMain.Client.InsertBySql(
		`INSERT INTO tbl_agent_v2_session (agent_id, workspace_id, name, session_dir, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'active', ?, ?)`,
		req.AgentId, req.WorkspaceId, name, sessionDir, now, now,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// sessionDir 在 WS 连接时由 agent_v2_ws.go 根据实际 sessionId 更新
	gsgin.GinResponseSuccess(c, "", gin.H{
		"id":          cast.ToInt(lastId),
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
func readSessionMessagesList(sessionDir string) []map[string]interface{} {
	if sessionDir == "" {
		return []map[string]interface{}{}
	}

	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return []map[string]interface{}{}
	}

	var messages []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
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

			// 转换为前端消息格式
			msg := convertPiEventToMessage(raw)
			if msg != nil {
				messages = append(messages, msg)
			}
		}
	}
	return messages
}

// convertPiEventToMessage 将 Pi 事件转换为前端消息格式
func convertPiEventToMessage(raw map[string]interface{}) map[string]interface{} {
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
