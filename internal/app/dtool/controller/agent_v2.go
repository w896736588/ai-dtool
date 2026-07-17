package controller

import (
	"dev_tool/internal/app/dtool/agent"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// ======================== Agent V2 CRUD ========================

// defaultPiAgentDir 返回 Pi 默认数据/配置目录（~/.pi/agent）
func defaultPiAgentDir() string {
	if h, err := os.UserHomeDir(); err == nil && h != "" {
		return filepath.Join(h, ".pi", "agent")
	}
	return filepath.Join(".pi", "agent")
}

// expandHome 将路径开头的 ~ 展开为用户主目录
func expandHome(p string) string {
	if p == "~" {
		if h, err := os.UserHomeDir(); err == nil && h != "" {
			return h
		}
	}
	if strings.HasPrefix(p, "~/") {
		if h, err := os.UserHomeDir(); err == nil && h != "" {
			return filepath.Join(h, p[2:])
		}
	}
	return p
}

// resolveRuntimeDir 解析运行目录：空 -> Pi 默认目录；否则展开 ~ 后返回
func resolveRuntimeDir(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultPiAgentDir()
	}
	return expandHome(raw)
}

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

	// 运行目录唯一性校验：解析后的有效目录不能与其他 Agent 重复
	var cfg struct {
		RuntimeDir string `json:"runtime_dir"`
	}
	_ = json.Unmarshal([]byte(req.Config), &cfg)
	newRuntimeDir := resolveRuntimeDir(cfg.RuntimeDir)
	rows, _ := common.DbMain.Client.QueryBySql(
		`SELECT id, config FROM tbl_agent_v2`,
	).All()
	for _, row := range rows {
		if cast.ToInt(row["id"]) == req.Id {
			continue
		}
		var ocfg struct {
			RuntimeDir string `json:"runtime_dir"`
		}
		_ = json.Unmarshal([]byte(cast.ToString(row["config"])), &ocfg)
		if resolveRuntimeDir(ocfg.RuntimeDir) == newRuntimeDir {
			gsgin.GinResponseError(c, "运行目录 "+newRuntimeDir+" 已被其他 Agent 占用，请指定不同的目录", nil)
			return
		}
	}

	now := time.Now().Unix()
	if req.Id > 0 {
		// 合并配置：保留原有 config 中本接口不管理的子键（如 headroom），
		// 否则保存基础配置时会整体覆盖，导致 Headroom 上游地址等被清空。
		mergedConfig, err := mergeAgentV2Config(req.Id, req.Config)
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		_, err = common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_v2 SET name = ?, type = ?, config = ?, updated_at = ? WHERE id = ?`,
			req.Name, req.Type, mergedConfig, now, req.Id,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		// 保存 Agent 基础配置（可能更新 runtime_dir）后同步一次模型列表，
		// 保证该 Agent 运行目录下 models.json 与最新 Provider/Model 配置一致。
		syncPiModelsConfigAllAgents("after agent update")
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
		// 新建 Agent 后立即同步一次模型列表，确保自定义 runtime_dir 的 Agent
		// 运行目录下写入 models.json（新建后可能首次启动 pi，避免 Unknown provider）。
		syncPiModelsConfigAllAgents("after agent create")
		gsgin.GinResponseSuccess(c, "", gin.H{"id": lastId})
	}
}

// mergeAgentV2Config 将新的配置 JSON 合并到现有 config 中，保留现有配置里
// 本接口不直接管理的子键（如 headroom），避免保存基础 Agent 配置时整体覆盖导致丢失。
// 新配置中的同名 key 覆盖旧值；若现有/新配置无法解析则回退为直接覆盖，保证不报错。
func mergeAgentV2Config(agentID int, newConfig string) (string, error) {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT config FROM tbl_agent_v2 WHERE id = ?`, agentID,
	).All()
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return newConfig, nil
	}

	existingRaw := cast.ToString(rows[0]["config"])
	var existing map[string]interface{}
	if existingRaw != "" && existingRaw != "{}" {
		if err := json.Unmarshal([]byte(existingRaw), &existing); err != nil {
			// 现有配置损坏，直接用新配置覆盖
			return newConfig, nil
		}
	}
	if existing == nil {
		return newConfig, nil
	}

	var incoming map[string]interface{}
	if err := json.Unmarshal([]byte(newConfig), &incoming); err != nil || incoming == nil {
		// 新配置无法解析，保留现有配置
		return existingRaw, nil
	}

	for k, v := range incoming {
		existing[k] = v
	}

	b, err := json.Marshal(existing)
	if err != nil {
		return newConfig, nil
	}
	return string(b), nil
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
		`SELECT * FROM tbl_agent_v2_workspace WHERE agent_id = ? ORDER BY sort_order ASC, id ASC`, req.AgentId,
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
			SortOrder: cast.ToInt(row["sort_order"]),
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
		// 新工作空间追加到末尾：sort_order 取当前最大值 + 1
		mxVal, err := common.DbMain.Client.QueryBySql(
			`SELECT COALESCE(MAX(sort_order), 0) AS mx FROM tbl_agent_v2_workspace WHERE agent_id = ?`, req.AgentId,
		).Value(`mx`)
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		sortOrder := cast.ToInt(mxVal) + 1
		_, err = common.DbMain.Client.InsertBySql(
			`INSERT INTO tbl_agent_v2_workspace (agent_id, name, path, sort_order, created_at) VALUES (?, ?, ?, ?, ?)`,
			req.AgentId, req.Name, req.Path, sortOrder, now,
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

// AgentV2WorkspaceReorder 批量保存工作空间顺序（按传入 id 顺序写入 sort_order）
func AgentV2WorkspaceReorder(c *gin.Context) {
	var req struct {
		AgentId    int   `json:"agent_id"`
		OrderedIds []int `json:"ordered_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.AgentId <= 0 {
		gsgin.GinResponseError(c, "缺少 agent_id", nil)
		return
	}

	for i, id := range req.OrderedIds {
		if id <= 0 {
			continue
		}
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_v2_workspace SET sort_order = ? WHERE id = ? AND agent_id = ?`,
			i, id, req.AgentId,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
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

// AgentV2SkillSave 新增/编辑 Skill/Tool
// 说明：当 skill_type == "tool" 且 config 中包含 script_content 时，
// 会把脚本物化到 Pi 扩展目录 ~/.pi/agent/extensions/<name>.ts，
// 这样 Pi 启动时才能真正加载该工具（此前只写 DB 导致扩展目录为空）。
func AgentV2SkillSave(c *gin.Context) {
	var req define.AgentV2SkillSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	now := time.Now().Unix()

	// 新建时，若同名同类型已存在则改为更新（幂等，便于重复安装内置工具）
	if req.Id == 0 {
		rows, _ := common.DbMain.Client.QueryBySql(
			`SELECT id FROM tbl_agent_v2_skill WHERE agent_id = ? AND name = ? AND skill_type = ?`,
			req.AgentId, req.Name, req.SkillType,
		).All()
		if len(rows) > 0 {
			req.Id = cast.ToInt(rows[0]["id"])
		}
	}

	if req.Id > 0 {
		// 改名时清理旧扩展文件
		var oldName string
		rows, _ := common.DbMain.Client.QueryBySql(
			`SELECT name FROM tbl_agent_v2_skill WHERE id = ?`, req.Id,
		).All()
		if len(rows) > 0 {
			oldName = cast.ToString(rows[0]["name"])
		}
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_v2_skill SET name = ?, skill_type = ?, config = ?, enabled = ?, updated_at = ? WHERE id = ?`,
			req.Name, req.SkillType, req.Config, req.Enabled, now, req.Id,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		if req.SkillType == "tool" {
			if oldName != "" && oldName != req.Name {
				_ = agent.RemoveToolExtension(oldName)
			}
			materializeToolExtension(req.Name, req.Config)
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
		if req.SkillType == "tool" {
			materializeToolExtension(req.Name, req.Config)
		}
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// materializeToolExtension 将 tool 的 script_content 写入 Pi 扩展目录
func materializeToolExtension(name, configStr string) {
	if name == "" || configStr == "" {
		return
	}
	var cfg struct {
		ScriptContent string `json:"script_content"`
	}
	if err := json.Unmarshal([]byte(configStr), &cfg); err != nil {
		return
	}
	if cfg.ScriptContent == "" {
		return
	}
	if err := agent.WriteToolExtension(name, cfg.ScriptContent); err != nil {
		log.Printf("[agent-v2] 物化工具扩展失败 %q: %v", name, err)
	}
}

// AgentV2SkillDelete 删除 Skill/Tool
// 工具类需同时清理 ~/.pi/agent/extensions/ 下的对应 .ts 文件
func AgentV2SkillDelete(c *gin.Context) {
	var req struct {
		Id int `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	// 删除前取出名称与类型，工具类需清理扩展文件
	var toolName string
	rows, _ := common.DbMain.Client.QueryBySql(
		`SELECT name, skill_type FROM tbl_agent_v2_skill WHERE id = ?`, req.Id,
	).All()
	if len(rows) > 0 && cast.ToString(rows[0]["skill_type"]) == "tool" {
		toolName = cast.ToString(rows[0]["name"])
	}

	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_v2_skill WHERE id = ?`, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	if toolName != "" {
		_ = agent.RemoveToolExtension(toolName)
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

		// 读取脚本文件：优先 index.ts；排除 *.d.ts 类型声明文件（如 env.d.ts）
		scriptContent := ""
		indexFile := filepath.Join(dirPath, "index.ts")
		if data, err := os.ReadFile(indexFile); err == nil {
			scriptContent = string(data)
		} else {
			// 回退：取目录中第一个非 *.d.ts 的 .ts 文件
			tsFiles, _ := filepath.Glob(filepath.Join(dirPath, "*.ts"))
			for _, f := range tsFiles {
				if strings.HasSuffix(f, ".d.ts") {
					continue
				}
				if data, err := os.ReadFile(f); err == nil {
					scriptContent = string(data)
					break
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

// ======================== 已安装扩展扫描（.pi/extensions/） ========================

// AgentV2InstalledToolList 扫描 .pi/extensions/ 目录下的已安装扩展
func AgentV2InstalledToolList(c *gin.Context) {
	tools := agent.ScanInstalledTools()
	gsgin.GinResponseSuccess(c, "", gin.H{"list": tools})
}

// AgentV2InstalledToolRemove 删除 .pi/extensions/ 下的扩展文件
func AgentV2InstalledToolRemove(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	if err := agent.RemoveInstalledTool(req.Name); err != nil {
		gsgin.GinResponseError(c, "删除失败: "+err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// ======================== 环境工具管理 ========================

// AgentV2EnvToolList 列出所有环境工具及其安装状态
func AgentV2EnvToolList(c *gin.Context) {
	var req struct {
		AgentId int `json:"agent_id"`
	}
	_ = c.ShouldBindJSON(&req)

	// 计划模式（Pi 内置扩展）的安装/启用状态取决于 Agent 启动参数（--plan / --extension）
	extraArgs := ""
	if req.AgentId > 0 {
		if ea, err := loadAgentV2ExtraArgs(req.AgentId); err == nil {
			extraArgs = ea
		}
	}

	tools := make([]define.AgentV2EnvToolStatus, 0, len(agent.BuiltinEnvToolDefs))
	for _, def := range agent.BuiltinEnvToolDefs {
		var st define.AgentV2EnvToolStatus
		if def.Key == "plan-mode" {
			st = detectPlanModeStatus(def, extraArgs)
		} else {
			st = agent.DetectEnvToolStatus(def)
		}
		tools = append(tools, st)
	}
	gsgin.GinResponseSuccess(c, "", gin.H{"list": tools})
}

// detectPlanModeStatus 根据 Agent 启动参数判断 Pi 内置计划模式的启用情况
func detectPlanModeStatus(def define.AgentV2EnvToolItem, extraArgs string) define.AgentV2EnvToolStatus {
	st := define.AgentV2EnvToolStatus{AgentV2EnvToolItem: def}
	hasPlan, hasExt := parsePlanModeFlags(extraArgs)
	switch {
	case hasPlan && hasExt:
		st.Installed = true
		st.ExtensionInstalled = true
		st.Activated = true
		st.Status = "activated"
	case hasExt:
		st.Installed = true
		st.ExtensionInstalled = true
		st.Status = "installed"
	default:
		st.Status = "not_installed"
	}
	return st
}

// parseAgentExtraArgs 将保存为字符串的额外参数解析为参数数组。
// 支持单引号和双引号，确保 Windows 上带空格的扩展路径不会被拆开。
func parseAgentExtraArgs(raw string) []string {
	args := make([]string, 0)
	var current strings.Builder
	var quote rune
	tokenStarted := false
	runes := []rune(raw)

	flush := func() {
		if tokenStarted {
			args = append(args, current.String())
			current.Reset()
			tokenStarted = false
		}
	}

	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if quote != 0 {
			if ch == quote {
				quote = 0
				tokenStarted = true
				continue
			}
			if ch == '\\' && i+1 < len(runes) && runes[i+1] == quote {
				current.WriteRune(quote)
				tokenStarted = true
				i++
				continue
			}
			current.WriteRune(ch)
			tokenStarted = true
			continue
		}

		switch {
		case ch == '\'' || ch == '"':
			quote = ch
			tokenStarted = true
		case unicode.IsSpace(ch):
			flush()
		default:
			current.WriteRune(ch)
			tokenStarted = true
		}
	}
	flush()
	return args
}

func quoteAgentExtraArg(arg string) string {
	if arg == "" {
		return `""`
	}
	if !strings.ContainsAny(arg, " \t\r\n'\"") {
		return arg
	}
	if !strings.ContainsRune(arg, '\'') {
		return "'" + arg + "'"
	}
	return `"` + strings.ReplaceAll(arg, `"`, `\"`) + `"`
}

func joinAgentExtraArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, quoteAgentExtraArg(arg))
	}
	return strings.Join(quoted, " ")
}

// parsePlanModeFlags 解析启动参数，判断是否包含 --plan 与指向 plan-mode 的 --extension
func parsePlanModeFlags(extraArgs string) (hasPlan, hasExt bool) {
	tokens := parseAgentExtraArgs(extraArgs)
	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case "--plan":
			hasPlan = true
		case "--extension":
			if i+1 < len(tokens) && strings.Contains(tokens[i+1], "plan-mode") {
				hasExt = true
			}
		}
	}
	return
}

// loadAgentV2ExtraArgs 读取指定 Agent 的启动参数（extra_args）
func loadAgentV2ExtraArgs(agentID int) (string, error) {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT config FROM tbl_agent_v2 WHERE id = ?`, agentID,
	).All()
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("agent 不存在")
	}
	var cfg struct {
		ExtraArgs string `json:"extra_args"`
	}
	if err := json.Unmarshal([]byte(cast.ToString(rows[0]["config"])), &cfg); err != nil {
		return "", err
	}
	return cfg.ExtraArgs, nil
}

// saveAgentV2ExtraArgs 写回指定 Agent 的启动参数（保留其他配置字段）
func saveAgentV2ExtraArgs(agentID int, extraArgs string) error {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT config FROM tbl_agent_v2 WHERE id = ?`, agentID,
	).All()
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return fmt.Errorf("agent 不存在")
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(cast.ToString(rows[0]["config"])), &cfg); err != nil || cfg == nil {
		cfg = map[string]interface{}{}
	}
	cfg["extra_args"] = extraArgs
	b, _ := json.Marshal(cfg)
	_, err = common.DbMain.Client.ExecBySql(
		`UPDATE tbl_agent_v2 SET config = ?, updated_at = ? WHERE id = ?`,
		string(b), time.Now().Unix(), agentID,
	).Exec()
	return err
}

// enablePlanModeFlags 在 Agent 启动参数中加入 --plan 与 --extension <plan-mode 源文件>
func enablePlanModeFlags(agentID int, extPath string) error {
	extra, err := loadAgentV2ExtraArgs(agentID)
	if err != nil {
		return err
	}
	tokens := parseAgentExtraArgs(extra)
	hasPlan, hasExt := parsePlanModeFlags(extra)
	if !hasPlan {
		tokens = append(tokens, "--plan")
	}
	if !hasExt {
		tokens = append(tokens, "--extension", extPath)
	}
	return saveAgentV2ExtraArgs(agentID, joinAgentExtraArgs(tokens))
}

// disablePlanModeFlags 从 Agent 启动参数中移除 --plan 与指向 plan-mode 的 --extension
func disablePlanModeFlags(agentID int) error {
	extra, err := loadAgentV2ExtraArgs(agentID)
	if err != nil {
		return err
	}
	tokens := parseAgentExtraArgs(extra)
	out := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "--plan" {
			continue
		}
		if tokens[i] == "--extension" && i+1 < len(tokens) && strings.Contains(tokens[i+1], "plan-mode") {
			i++ // 跳过紧跟的扩展路径
			continue
		}
		out = append(out, tokens[i])
	}
	return saveAgentV2ExtraArgs(agentID, joinAgentExtraArgs(out))
}

// AgentV2EnvToolAction 执行环境工具操作（安装/激活/停用/移除）
func AgentV2EnvToolAction(c *gin.Context) {
	var req define.AgentV2EnvToolActionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Key == "" || req.Action == "" {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	// 查找工具定义
	var def *define.AgentV2EnvToolItem
	for i := range agent.BuiltinEnvToolDefs {
		if agent.BuiltinEnvToolDefs[i].Key == req.Key {
			def = &agent.BuiltinEnvToolDefs[i]
			break
		}
	}
	if def == nil {
		gsgin.GinResponseError(c, "未知的环境工具: "+req.Key, nil)
		return
	}

	switch req.Action {
	case "install":
		gsgin.GinResponseSuccess(c, "", gin.H{
			"action":  "install",
			"command": def.InstallCmdHint,
			"message": "请在终端执行以下命令完成安装，然后刷新页面检测状态",
			"success": false,
		})
	case "activate":
		// Pi 内置计划模式：一键加载扩展并加入启动参数，无需手动复制文件
		if def.Key == "plan-mode" {
			src := agent.FindPiPlanModeSource()
			if src == "" {
				gsgin.GinResponseError(c, "未找到 Pi 自带的 plan-mode 扩展源文件，请确认 pi-coding-agent 已正确安装", nil)
				return
			}
			if err := enablePlanModeFlags(req.AgentId, src); err != nil {
				gsgin.GinResponseError(c, "启用计划模式失败: "+err.Error(), nil)
				return
			}
			gsgin.GinResponseSuccess(c, "计划模式已启用（已加入 --plan 与 --extension 启动参数），重启 Agent 会话后生效", nil)
			return
		}
		cmd := def.ActivateCmdHint
		if cmd == "" {
			gsgin.GinResponseError(c, "该工具无需激活", nil)
			return
		}
		gsgin.GinResponseSuccess(c, "", gin.H{
			"action":  "activate",
			"command": cmd,
			"message": "请在终端执行以下命令激活，完成后刷新页面检测状态",
			"success": false,
		})
	case "deactivate":
		cmd := agent.GetUninstallCmd(req.Key)
		if cmd == "" {
			gsgin.GinResponseError(c, "该工具不支持停用操作", nil)
			return
		}
		gsgin.GinResponseSuccess(c, "", gin.H{
			"action":  "deactivate",
			"command": cmd,
			"message": "请在终端执行以下命令停用 hook",
			"success": false,
		})
	case "remove":
		// Pi 内置计划模式：仅移除启动参数（未复制文件到扩展目录）
		if def.Key == "plan-mode" {
			if err := disablePlanModeFlags(req.AgentId); err != nil {
				gsgin.GinResponseError(c, "关闭计划模式失败: "+err.Error(), nil)
				return
			}
			gsgin.GinResponseSuccess(c, "计划模式已关闭（已移除启动参数），重启 Agent 会话后生效", nil)
			return
		}
		// 直接从 .pi/extensions/ 删除文件
		if err := agent.RemoveInstalledTool(req.Key); err != nil {
			gsgin.GinResponseError(c, "移除失败: "+err.Error(), nil)
			return
		}
		gsgin.GinResponseSuccess(c, "", nil)
	default:
		gsgin.GinResponseError(c, "不支持的操作: "+req.Action, nil)
	}
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
