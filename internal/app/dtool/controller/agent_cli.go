package controller

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"log"
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

	// 预加载 webhook 配置名称映射
	webhookNameMap := make(map[int]string)
	webhookRows, _ := common.DbMain.Client.QueryBySql(
		`SELECT id, name FROM tbl_webhook_config`,
	).All()
	for _, wr := range webhookRows {
		webhookNameMap[cast.ToInt(wr["id"])] = cast.ToString(wr["name"])
	}

	// 预加载分组关联映射（agent_cli_id -> []group_id）
	groupRelMap := make(map[int][]int)
	groupRelRows, _ := common.DbMain.Client.QueryBySql(
		`SELECT agent_cli_id, group_id FROM tbl_agent_cli_group_rel`,
	).All()
	for _, gr := range groupRelRows {
		cliId := cast.ToInt(gr["agent_cli_id"])
		groupId := cast.ToInt(gr["group_id"])
		groupRelMap[cliId] = append(groupRelMap[cliId], groupId)
	}

	for _, row := range rows {
		item := define.AgentCliStatusItem{
			AgentCliItem: define.AgentCliItem{
				Id:                cast.ToInt(row["id"]),
				Name:              cast.ToString(row["name"]),
				Type:              cast.ToString(row["type"]),
				SettingsPath:      cast.ToString(row["settings_path"]),
				Enabled:           cast.ToInt(row["enabled"]),
				ThinkingCollapsed: cast.ToInt(row["thinking_collapsed"]),
				WebhookConfigId:   cast.ToInt(row["webhook_config_id"]),
				CreatedAt:         cast.ToInt64(row["created_at"]),
				UpdatedAt:         cast.ToInt64(row["updated_at"]),
			},
		}

		if item.WebhookConfigId > 0 {
			item.WebhookConfigName = webhookNameMap[item.WebhookConfigId]
		}

		// 填充分组关联
		if gids, ok := groupRelMap[item.Id]; ok {
			item.GroupIds = gids
		} else {
			item.GroupIds = []int{}
		}

		// 根据 CLI 类型获取不同的状态摘要
		if item.Type == define.AgentCliTypeCodexCli {
			configJson := cast.ToString(row["config"])
			item.Config = configJson
			codexModel, codexBaseURL := business.GetCodexCliModelConfig(configJson)
			item.ModelOptions = business.GetCodexCliModelOptions(configJson)
			if codexModel != "" {
				item.CurrentModel = codexModel
			} else {
				item.CurrentModel = "-"
			}
			item.RequestURL = codexBaseURL
			item.SettingsExists = configJson != ""
			item.DisplayedEnabled = item.Enabled == define.AgentCliEnabled && item.SettingsExists
		} else {
			exists, content, _ := business.ReadAgentCliSettings(item.SettingsPath)
			item.SettingsExists = exists
			if exists {
				item.CurrentModel, item.McpServerCount, item.ClaudeMemEnabled = business.GetAgentCliSettingsSummary(content)
				item.ModelOptions = business.GetAgentCliModelOptions(content)
				_, item.RequestURL, _ = business.GetAgentCliModelConfig(content)
				if item.CurrentModel == "" {
					item.CurrentModel = "-"
				}
			}
			item.DisplayedEnabled = item.Enabled == define.AgentCliEnabled && item.SettingsExists
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

	// 根据 CLI 类型验证必填字段
	var codexCfg *define.CodexCliConfig
	if req.Type == define.AgentCliTypeCodexCli {
		if req.Config == "" {
			gsgin.GinResponseError(c, "Codex CLI 配置不能为空", nil)
			return
		}
		// 验证 config JSON 中 api_key 必填
		var cfgErr error
		codexCfg, cfgErr = business.GetCodexCliConfig(req.Config)
		if cfgErr != nil {
			gsgin.GinResponseError(c, cfgErr.Error(), nil)
			return
		}
		if codexCfg.ApiKey == "" {
			gsgin.GinResponseError(c, "Codex CLI API Key 不能为空", nil)
			return
		}
	}

	now := time.Now().Unix()
	var savedItem define.AgentCliItem
	enabled := req.Enabled
	if req.Id <= 0 {
		if req.Type == define.AgentCliTypeCodexCli {
			enabled = define.AgentCliDisabled
		} else {
			enabled = define.AgentCliEnabled
		}
	}
	if enabled != define.AgentCliEnabled {
		enabled = define.AgentCliDisabled
	}

	if req.Id > 0 {
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_cli SET name = ?, type = ?, settings_path = ?, config = ?, enabled = ?, thinking_collapsed = ?, webhook_config_id = ?, updated_at = ? WHERE id = ?`,
			req.Name, req.Type, req.SettingsPath, req.Config, enabled, req.ThinkingCollapsed, req.WebhookConfigId, now, req.Id,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		if enabled == define.AgentCliEnabled && req.Type == define.AgentCliTypeCodexCli {
			disableOtherAgentCliByType(req.Type, req.Id)
		}
		savedItem = define.AgentCliItem{
			Id:                req.Id,
			Name:              req.Name,
			Type:              req.Type,
			SettingsPath:      req.SettingsPath,
			Config:            req.Config,
			Enabled:           enabled,
			ThinkingCollapsed: req.ThinkingCollapsed,
			WebhookConfigId:   req.WebhookConfigId,
			CreatedAt:         0,
			UpdatedAt:         now,
		}
	} else {
		name := req.Name
		if name == "" {
			if req.Type == define.AgentCliTypeCodexCli {
				name = "Codex CLI"
			} else {
				name = "Claude Code CLI"
			}
		}
		if req.Type == "" {
			req.Type = define.AgentCliTypeClaudeCodeCli
		}
		lastId, err := common.DbMain.Client.InsertBySql(
			`INSERT INTO tbl_agent_cli (name, type, settings_path, config, enabled, thinking_collapsed, webhook_config_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			name, req.Type, req.SettingsPath, req.Config, enabled, req.ThinkingCollapsed, req.WebhookConfigId, now, now,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		if enabled == define.AgentCliEnabled && req.Type == define.AgentCliTypeCodexCli {
			disableOtherAgentCliByType(req.Type, int(lastId))
		}
		savedItem = define.AgentCliItem{
			Id:                int(lastId),
			Name:              name,
			Type:              req.Type,
			SettingsPath:      req.SettingsPath,
			Config:            req.Config,
			Enabled:           enabled,
			ThinkingCollapsed: req.ThinkingCollapsed,
			WebhookConfigId:   req.WebhookConfigId,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
	}

	// 仅当当前 Codex CLI 实例处于启用态时，才同步全局 ~/.codex 配置，避免编辑其它实例时串改当前活动实例。
	if codexCfg != nil && savedItem.Enabled == define.AgentCliEnabled {
		if writeErr := business.WriteCodexConfigToToml(codexCfg); writeErr != nil {
			log.Printf("[agent-cli] 写入 Codex config.toml 失败: %v", writeErr)
		}
		if codexCfg.BaseURL != "" && codexCfg.ApiKey != "" {
			if writeErr := business.WriteCodexAuthJson(codexCfg.ApiKey); writeErr != nil {
				log.Printf("[agent-cli] 写入 Codex auth.json 失败: %v", writeErr)
			}
		}
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

	// 清理分组关联
	common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_cli_group_rel WHERE agent_cli_id = ?`, req.Id,
	).Exec()

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
	if err := business.WriteDeepSeekToSettings(settingsPath, req.ModelName, req.ModelList, req.ApiKey, req.BaseUrl); err != nil {
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

// AgentCliToggleEnabled 切换 Agent CLI 启停；同类型同一时刻仅允许一个启用。
func AgentCliToggleEnabled(c *gin.Context) {
	var req define.AgentCliToggleEnabledRequest
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

	now := time.Now().Unix()
	enabled := define.AgentCliDisabled
	if req.Enable {
		enabled = define.AgentCliEnabled
	}
	_, err = common.DbMain.Client.ExecBySql(
		`UPDATE tbl_agent_cli SET enabled = ?, updated_at = ? WHERE id = ?`,
		enabled, now, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	rowType := cast.ToString(row["type"])
	if enabled == define.AgentCliEnabled {
		// 启用 Codex CLI 时同步写入全局 codex 配置 / Sync global codex config when enabling a codex CLI.
		if rowType == define.AgentCliTypeCodexCli {
			disableOtherAgentCliByType(rowType, req.Id)
			if err := applyCodexCliConfigByRow(row); err != nil {
				gsgin.GinResponseError(c, err.Error(), nil)
				return
			}
		}
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// disableOtherAgentCliByType 关闭同类型的其它 Agent CLI / disable other enabled Agent CLIs of same type.
func disableOtherAgentCliByType(cliType string, currentID int) {
	if cliType == "" || currentID <= 0 {
		return
	}
	_, _ = common.DbMain.Client.ExecBySql(
		`UPDATE tbl_agent_cli SET enabled = ?, updated_at = ? WHERE type = ? AND id != ?`,
		define.AgentCliDisabled, time.Now().Unix(), cliType, currentID,
	).Exec()
}

// applyCodexCliConfigByRow 将当前 Agent CLI 行中的 Codex 配置写入全局 ~/.codex 文件 / apply current row codex config to global ~/.codex files.
func applyCodexCliConfigByRow(row map[string]any) error {
	configJson := cast.ToString(row["config"])
	codexCfg, err := business.GetCodexCliConfig(configJson)
	if err != nil {
		return err
	}
	if err := business.WriteCodexConfigToToml(codexCfg); err != nil {
		log.Printf("[agent-cli] 启用时写入 Codex config.toml 失败: %v", err)
		return err
	}
	// 仅自定义 base_url 模式写 auth.json / Only custom base_url mode requires auth.json.
	if codexCfg.BaseURL != "" && codexCfg.ApiKey != "" {
		if err := business.WriteCodexAuthJson(codexCfg.ApiKey); err != nil {
			log.Printf("[agent-cli] 启用时写入 Codex auth.json 失败: %v", err)
			return err
		}
	}
	return nil
}
