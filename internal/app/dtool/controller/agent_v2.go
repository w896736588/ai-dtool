package controller

import (
	"dev_tool/internal/app/dtool/agent"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// ======================== Agent V2 CRUD ========================

// AgentV2List 列出所有 Agent V2 配置
func AgentV2List(c *gin.Context) {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_v2 ORDER BY id`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	items := make([]define.AgentV2StatusItem, 0, len(rows))
	for _, row := range rows {
		item := define.AgentV2StatusItem{
			AgentV2Item: define.AgentV2Item{
				Id:        cast.ToInt(row["id"]),
				Name:      cast.ToString(row["name"]),
				Type:      cast.ToString(row["type"]),
				Config:    cast.ToString(row["config"]),
				Enabled:   cast.ToInt(row["enabled"]),
				CreatedAt: cast.ToInt64(row["created_at"]),
				UpdatedAt: cast.ToInt64(row["updated_at"]),
			},
		}

		// 检测是否已安装
		adapter := getAdapterForType(item.Type)
		item.Installed = adapter.IsInstalled()
		item.InstallHint = adapter.InstallHint()

		// 统计会话数
		sessionRows, err := common.DbMain.Client.QueryBySql(
			`SELECT COUNT(*) as cnt FROM tbl_agent_v2_session WHERE agent_id = ?`, item.Id,
		).All()
		if err != nil {
			log.Printf("[agent-v2] count sessions for agent %d error: %v", item.Id, err)
		} else if len(sessionRows) > 0 {
			item.SessionCount = cast.ToInt(sessionRows[0]["cnt"])
		}

		items = append(items, item)
	}

	gsgin.GinResponseSuccess(c, "", gin.H{"list": items})
}

// AgentV2Save 新增/编辑 Agent V2 配置
func AgentV2Save(c *gin.Context) {
	var req define.AgentV2SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	now := time.Now().Unix()
	if req.Id > 0 {
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_v2 SET name = ?, type = ?, config = ?, updated_at = ? WHERE id = ?`,
			req.Name, req.Type, req.Config, now, req.Id,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		gsgin.GinResponseSuccess(c, "", nil)
	} else {
		name := req.Name
		if name == "" {
			name = req.Type
		}
		lastId, err := common.DbMain.Client.InsertBySql(
			`INSERT INTO tbl_agent_v2 (name, type, config, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
			name, req.Type, req.Config, 0, now, now,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		gsgin.GinResponseSuccess(c, "", gin.H{"id": lastId})
	}
}

// AgentV2Delete 删除 Agent V2 配置
func AgentV2Delete(c *gin.Context) {
	var req struct {
		Id int `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	if req.Id <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}

	// 事务包裹级联删除
	if _, err := common.DbMain.Client.ExecBySql("BEGIN TRANSACTION").Exec(); err != nil {
		gsgin.GinResponseError(c, "事务启动失败: "+err.Error(), nil)
		return
	}
	rolledBack := false
	defer func() {
		if !rolledBack {
			return
		}
		common.DbMain.Client.ExecBySql("ROLLBACK").Exec()
	}()
	rollback := func(msg string) {
		log.Printf("[agent-v2] delete agent %d: %s", req.Id, msg)
		rolledBack = true
		gsgin.GinResponseError(c, msg, nil)
	}

	if _, err := common.DbMain.Client.ExecBySql(`DELETE FROM tbl_agent_v2_workspace WHERE agent_id = ?`, req.Id).Exec(); err != nil {
		rollback("删除关联工作空间失败: " + err.Error())
		return
	}
	if _, err := common.DbMain.Client.ExecBySql(`DELETE FROM tbl_agent_v2_session WHERE agent_id = ?`, req.Id).Exec(); err != nil {
		rollback("删除关联会话失败: " + err.Error())
		return
	}
	if _, err := common.DbMain.Client.ExecBySql(`DELETE FROM tbl_agent_v2_skill WHERE agent_id = ?`, req.Id).Exec(); err != nil {
		rollback("删除关联 Skill 失败: " + err.Error())
		return
	}
	if _, err := common.DbMain.Client.ExecBySql(`DELETE FROM tbl_agent_v2 WHERE id = ?`, req.Id).Exec(); err != nil {
		rollback("删除 Agent 失败: " + err.Error())
		return
	}

	if _, err := common.DbMain.Client.ExecBySql("COMMIT").Exec(); err != nil {
		rollback("提交事务失败: " + err.Error())
		return
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// AgentV2CheckInstall 检测 Agent 是否已安装
func AgentV2CheckInstall(c *gin.Context) {
	var req struct {
		Type string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	adapter := getAdapterForType(req.Type)
	gsgin.GinResponseSuccess(c, "", gin.H{
		"installed":    adapter.IsInstalled(),
		"install_hint": adapter.InstallHint(),
	})
}

// ======================== 工作空间管理 ========================

// AgentV2WorkspaceList 列出工作空间
func AgentV2WorkspaceList(c *gin.Context) {
	var req struct {
		AgentId int `json:"agent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_v2_workspace WHERE agent_id = ? ORDER BY id`, req.AgentId,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	items := make([]define.AgentV2Workspace, 0, len(rows))
	for _, row := range rows {
		items = append(items, define.AgentV2Workspace{
			Id:        cast.ToInt(row["id"]),
			AgentId:   cast.ToInt(row["agent_id"]),
			Name:      cast.ToString(row["name"]),
			Path:      cast.ToString(row["path"]),
			CreatedAt: cast.ToInt64(row["created_at"]),
		})
	}

	gsgin.GinResponseSuccess(c, "", gin.H{"list": items})
}

// AgentV2WorkspaceSave 新增/编辑工作空间
func AgentV2WorkspaceSave(c *gin.Context) {
	var req define.AgentV2WorkspaceSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	now := time.Now().Unix()
	if req.Id > 0 {
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_v2_workspace SET name = ?, path = ? WHERE id = ?`,
			req.Name, req.Path, req.Id,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	} else {
		_, err := common.DbMain.Client.InsertBySql(
			`INSERT INTO tbl_agent_v2_workspace (agent_id, name, path, created_at) VALUES (?, ?, ?, ?)`,
			req.AgentId, req.Name, req.Path, now,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// AgentV2WorkspaceDelete 删除工作空间
func AgentV2WorkspaceDelete(c *gin.Context) {
	var req struct {
		Id int `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_v2_workspace WHERE id = ?`, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// ======================== Skills/Tools 管理 ========================

// AgentV2SkillList 列出 Skills
func AgentV2SkillList(c *gin.Context) {
	var req struct {
		AgentId int `json:"agent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_v2_skill WHERE agent_id = ? ORDER BY id`, req.AgentId,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	items := make([]define.AgentV2Skill, 0, len(rows))
	for _, row := range rows {
		items = append(items, define.AgentV2Skill{
			Id:        cast.ToInt(row["id"]),
			AgentId:   cast.ToInt(row["agent_id"]),
			Name:      cast.ToString(row["name"]),
			SkillType: cast.ToString(row["skill_type"]),
			Config:    cast.ToString(row["config"]),
			Enabled:   cast.ToInt(row["enabled"]),
			CreatedAt: cast.ToInt64(row["created_at"]),
			UpdatedAt: cast.ToInt64(row["updated_at"]),
		})
	}

	gsgin.GinResponseSuccess(c, "", gin.H{"list": items})
}

// AgentV2SkillSave 新增/编辑 Skill
func AgentV2SkillSave(c *gin.Context) {
	var req define.AgentV2SkillSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	now := time.Now().Unix()
	if req.Id > 0 {
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_v2_skill SET name = ?, skill_type = ?, config = ?, enabled = ?, updated_at = ? WHERE id = ?`,
			req.Name, req.SkillType, req.Config, req.Enabled, now, req.Id,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	} else {
		_, err := common.DbMain.Client.InsertBySql(
			`INSERT INTO tbl_agent_v2_skill (agent_id, name, skill_type, config, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			req.AgentId, req.Name, req.SkillType, req.Config, req.Enabled, now, now,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// AgentV2SkillDelete 删除 Skill
func AgentV2SkillDelete(c *gin.Context) {
	var req struct {
		Id int `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_v2_skill WHERE id = ?`, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// ======================== 内置工具列表 ========================

// AgentV2BuiltinToolList 读取 data/ 目录下的内置工具
func AgentV2BuiltinToolList(c *gin.Context) {
	dataDir := define.DefaultBuiltinToolsDir

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		gsgin.GinResponseError(c, "读取内置工具目录失败: "+err.Error(), nil)
		return
	}

	tools := make([]define.AgentV2BuiltinTool, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(dataDir, entry.Name())
		metaPath := filepath.Join(dirPath, "meta.json")
		metaData, err := os.ReadFile(metaPath)
		if err != nil {
			continue // 跳过没有 meta.json 的目录
		}

		var meta struct {
			Name            string                        `json:"name"`
			ToolName        string                        `json:"tool_name"`
			Description     string                        `json:"description"`
			ToolDescription string                        `json:"tool_description"`
			Parameters      []define.AgentV2ToolParameter `json:"parameters"`
		}
		if err := json.Unmarshal(metaData, &meta); err != nil {
			continue
		}

		// 读取脚本文件（优先 index.ts，否则第一个 .ts 文件）
		scriptContent := ""
		tsFiles, _ := filepath.Glob(filepath.Join(dirPath, "*.ts"))
		if len(tsFiles) > 0 {
			for _, f := range tsFiles {
				if filepath.Base(f) == "index.ts" || strings.HasSuffix(f, ".ts") {
					data, err := os.ReadFile(f)
					if err == nil {
						scriptContent = string(data)
						break
					}
				}
			}
		}

		tools = append(tools, define.AgentV2BuiltinTool{
			DirName:         entry.Name(),
			Name:            meta.Name,
			ToolName:        meta.ToolName,
			Description:     meta.Description,
			ToolDescription: meta.ToolDescription,
			Parameters:      meta.Parameters,
			ScriptContent:   scriptContent,
		})
	}

	gsgin.GinResponseSuccess(c, "", gin.H{"list": tools})
}

// ======================== 模型配置 ========================

// AgentV2ProviderModels 获取所有 Provider 及其 LLM 模型（供 Agent 模块选择模型使用）
func AgentV2ProviderModels(c *gin.Context) {
	providers, err := common.DbMain.Client.QueryBySql(
		`SELECT id, name, provider_type, base_url
		 FROM tbl_ai_provider WHERE status = 1 ORDER BY id`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	type ModelItem struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Model       string `json:"model"`
		Uri         string `json:"uri"`
		ContextSize int    `json:"context_size"`
	}
	type ProviderWithModels struct {
		Id           int         `json:"id"`
		Name         string      `json:"name"`
		ProviderType string      `json:"provider_type"`
		BaseUrl      string      `json:"base_url"`
		Models       []ModelItem `json:"models"`
	}

	result := make([]ProviderWithModels, 0, len(providers))
	for _, p := range providers {
		pid := cast.ToInt(p["id"])
		modelRows, _ := common.DbMain.Client.QueryBySql(
			`SELECT id, name, model, uri, context_size FROM tbl_ai_model
			 WHERE provider_id = ? AND model_type = 'llm' AND status = 1 ORDER BY id`,
			pid,
		).All()

		models := make([]ModelItem, 0, len(modelRows))
		for _, m := range modelRows {
			models = append(models, ModelItem{
				Id:          cast.ToInt(m["id"]),
				Name:        cast.ToString(m["name"]),
				Model:       cast.ToString(m["model"]),
				Uri:         cast.ToString(m["uri"]),
				ContextSize: cast.ToInt(m["context_size"]),
			})
		}
		result = append(result, ProviderWithModels{
			Id:           pid,
			Name:         cast.ToString(p["name"]),
			ProviderType: cast.ToString(p["provider_type"]),
			BaseUrl:      cast.ToString(p["base_url"]),
			Models:       models,
		})
	}

	gsgin.GinResponseSuccess(c, "", gin.H{"providers": result})
}

// ======================== 辅助函数 ========================

// getAdapterForType 根据类型获取适配器实例
func getAdapterForType(agentType string) agent.AgentAdapter {
	switch agentType {
	case define.AgentV2TypePi:
		return agent.NewPiAdapter()
	// TODO: Codex 和 Claude Code 适配器待实现
	case define.AgentV2TypeCodex, define.AgentV2TypeClaudeCode:
		log.Printf("[agent-v2] unsupported agent type: %s", agentType)
		return agent.NewPiAdapter() // 临时回退
	default:
		log.Printf("[agent-v2] unknown agent type: %s, fallback to pi", agentType)
		return agent.NewPiAdapter()
	}
}
