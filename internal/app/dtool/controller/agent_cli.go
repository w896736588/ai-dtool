package controller

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// AgentCliList 返回所有 Agent Cli 实例列表（含状态摘要）
func AgentCliList(c *gin.Context) {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_cli ORDER BY id`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	items := make([]define.AgentCliStatusItem, 0, len(rows))
	for _, row := range rows {
		item := define.AgentCliStatusItem{
			AgentCliItem: define.AgentCliItem{
				Id:                cast.ToInt(row["id"]),
				Name:              cast.ToString(row["name"]),
				Type:              cast.ToString(row["type"]),
				SettingsPath:      cast.ToString(row["settings_path"]),
				ThinkingCollapsed: cast.ToInt(row["thinking_collapsed"]),
				CreatedAt:         cast.ToInt64(row["created_at"]),
				UpdatedAt:         cast.ToInt64(row["updated_at"]),
			},
		}

		exists, content, _ := business.ReadAgentCliSettings(item.SettingsPath)
		item.SettingsExists = exists
		if exists {
			item.CurrentModel, item.McpServerCount, item.ClaudeMemEnabled = business.GetAgentCliSettingsSummary(content)
			if item.CurrentModel == "" {
				item.CurrentModel = "-"
			}
		}

		items = append(items, item)
	}

	gsgin.GinResponseSuccess(c, "", gin.H{"list": items})
}

// AgentCliSave 新增/编辑 Agent Cli 实例
func AgentCliSave(c *gin.Context) {
	var req define.AgentCliSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	if req.SettingsPath == "" {
		gsgin.GinResponseError(c, "settings.json 路径不能为空", nil)
		return
	}

	now := time.Now().Unix()

	if req.Id > 0 {
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_cli SET name = ?, type = ?, settings_path = ?, thinking_collapsed = ?, updated_at = ? WHERE id = ?`,
			req.Name, req.Type, req.SettingsPath, req.ThinkingCollapsed, now, req.Id,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		savedItem := define.AgentCliItem{
			Id:                req.Id,
			Name:              req.Name,
			Type:              req.Type,
			SettingsPath:      req.SettingsPath,
			ThinkingCollapsed: req.ThinkingCollapsed,
			CreatedAt:         0,
			UpdatedAt:         now,
		}
		gsgin.GinResponseSuccess(c, "", savedItem)
		return
	}

	name := req.Name
	if name == "" {
		name = "Claude Code CLI"
	}
	if req.Type == "" {
		req.Type = define.AgentCliTypeClaudeCodeCli
	}
	lastId, err := common.DbMain.Client.InsertBySql(
		`INSERT INTO tbl_agent_cli (name, type, settings_path, thinking_collapsed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		name, req.Type, req.SettingsPath, req.ThinkingCollapsed, now, now,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	savedItem := define.AgentCliItem{
		Id:                int(lastId),
		Name:              name,
		Type:              req.Type,
		SettingsPath:      req.SettingsPath,
		ThinkingCollapsed: req.ThinkingCollapsed,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	gsgin.GinResponseSuccess(c, "", savedItem)
}

// AgentCliDelete 删除 Agent Cli 实例
func AgentCliDelete(c *gin.Context) {
	var req define.AgentCliDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	if req.Id <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}

	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_cli WHERE id = ?`, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// AgentCliReadSettings 读取指定实例的 settings.json 内容
func AgentCliReadSettings(c *gin.Context) {
	var req define.AgentCliReadSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	row, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_cli WHERE id = ?`, req.Id,
	).One()
	if err != nil || len(row) == 0 {
		gsgin.GinResponseError(c, "Agent Cli 实例不存在", nil)
		return
	}

	settingsPath := cast.ToString(row["settings_path"])
	exists, content, err := business.ReadAgentCliSettings(settingsPath)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", define.AgentCliReadSettingsResponse{
		Exists:       exists,
		Content:      content,
		SettingsPath: settingsPath,
	})
}

// AgentCliWriteMcpServers 将全部 ChromeDevtools 端口写入 settings.json
func AgentCliWriteMcpServers(c *gin.Context) {
	var req define.AgentCliWriteMcpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	row, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_cli WHERE id = ?`, req.Id,
	).One()
	if err != nil || len(row) == 0 {
		gsgin.GinResponseError(c, "Agent Cli 实例不存在", nil)
		return
	}

	settingsPath := cast.ToString(row["settings_path"])
	if err := business.WriteMcpServersToSettings(settingsPath); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	now := time.Now().Unix()
	common.DbMain.Client.ExecBySql(
		`UPDATE tbl_agent_cli SET updated_at = ? WHERE id = ?`, now, req.Id,
	).Exec()

	gsgin.GinResponseSuccess(c, "", nil)
}

// AgentCliWriteDeepSeek 写入 DeepSeek 配置到 settings.json
func AgentCliWriteDeepSeek(c *gin.Context) {
	var req define.AgentCliWriteDeepSeekRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	if req.ModelName == "" || req.ApiKey == "" {
		gsgin.GinResponseError(c, "模型名和 API Key 不能为空", nil)
		return
	}

	row, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_cli WHERE id = ?`, req.Id,
	).One()
	if err != nil || len(row) == 0 {
		gsgin.GinResponseError(c, "Agent Cli 实例不存在", nil)
		return
	}

	settingsPath := cast.ToString(row["settings_path"])
	if err := business.WriteDeepSeekToSettings(settingsPath, req.ModelName, req.ApiKey, req.BaseUrl); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	now := time.Now().Unix()
	common.DbMain.Client.ExecBySql(
		`UPDATE tbl_agent_cli SET updated_at = ? WHERE id = ?`, now, req.Id,
	).Exec()

	gsgin.GinResponseSuccess(c, "", nil)
}

// AgentCliToggleClaudeMem 切换 claude-mem 插件启停
func AgentCliToggleClaudeMem(c *gin.Context) {
	var req define.AgentCliToggleClaudeMemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	row, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_cli WHERE id = ?`, req.Id,
	).One()
	if err != nil || len(row) == 0 {
		gsgin.GinResponseError(c, "Agent Cli 实例不存在", nil)
		return
	}

	settingsPath := cast.ToString(row["settings_path"])
	if err := business.ToggleClaudeMem(settingsPath, req.Enable); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	now := time.Now().Unix()
	common.DbMain.Client.ExecBySql(
		`UPDATE tbl_agent_cli SET updated_at = ? WHERE id = ?`, now, req.Id,
	).Exec()

	gsgin.GinResponseSuccess(c, "", nil)
}
