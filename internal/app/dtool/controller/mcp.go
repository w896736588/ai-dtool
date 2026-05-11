package controller

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// McpTypeList 返回所有 MCP 类型定义及各目标智能体的绑定数统计
func McpTypeList(c *gin.Context) {
	// 查询所有目标智能体
	targetRows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_mcp_agent_target ORDER BY id`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 构建绑定数统计: agent_target_id -> count, 按 mcp_type 分组
	type bindCountKey struct {
		McpType       string
		AgentTargetId int
	}
	countMap := make(map[bindCountKey]int)
	if len(targetRows) > 0 {
		rows, err := common.DbMain.Client.QueryBySql(
			`SELECT mcp_type, agent_target_id, COUNT(*) as cnt FROM tbl_mcp_binding GROUP BY mcp_type, agent_target_id`,
		).All()
		if err == nil {
			for _, row := range rows {
				key := bindCountKey{
					McpType:       cast.ToString(row["mcp_type"]),
					AgentTargetId: cast.ToInt(row["agent_target_id"]),
				}
				countMap[key] = cast.ToInt(row["cnt"])
			}
		}
	}

	// 构建响应列表
	items := make([]define.McpTypeItem, 0, len(define.McpTypeDefs))
	for mcpType, typeDef := range define.McpTypeDefs {
		bindCountPerTarget := make(map[string]int)
		for _, tRow := range targetRows {
			targetId := cast.ToInt(tRow["id"])
			key := bindCountKey{McpType: mcpType, AgentTargetId: targetId}
			k := cast.ToString(targetId)
			bindCountPerTarget[k] = countMap[key]
		}
		items = append(items, define.McpTypeItem{
			McpType:      mcpType,
			Name:         typeDef.Name,
			PackageName:  typeDef.PackageName,
			Description:  typeDef.Description,
			BindCountMap: bindCountPerTarget,
		})
	}
	gsgin.GinResponseSuccess(c, "", items)
}

// McpBindingList 返回指定 mcp_type + agent_target_id 下的所有目录映射及绑定状态
func McpBindingList(c *gin.Context) {
	var req define.McpRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	if req.McpType == "" || req.AgentTargetId <= 0 {
		gsgin.GinResponseError(c, "mcp_type 和 agent_target_id 不能为空", nil)
		return
	}

	// 查询所有目录映射
	mappingRows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_smart_link_directory_mapping ORDER BY id`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 查询当前 mcp_type + agent_target_id 的绑定记录
	bindingRows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_mcp_binding WHERE mcp_type = ? AND agent_target_id = ?`,
		req.McpType, req.AgentTargetId,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 构建绑定查找表: mapping_id -> binding row
	bindingMap := make(map[int]map[string]any)
	for _, bRow := range bindingRows {
		mappingId := cast.ToInt(bRow["mapping_id"])
		bindingMap[mappingId] = bRow
	}

	// 构建响应
	items := make([]define.McpBindingItem, 0, len(mappingRows))
	for _, mRow := range mappingRows {
		mappingId := cast.ToInt(mRow["id"])
		userDataIndex := cast.ToInt(mRow["user_data_index"])
		userDataDir := ""
		if common.DbMain.Env != nil {
			userDataDir = filepath.Join(common.DbMain.Env.WebkitDataPath, cast.ToString(userDataIndex))
		}

		item := define.McpBindingItem{
			MappingId:     mappingId,
			MappingKey:    cast.ToString(mRow["mapping_key"]),
			Label:         cast.ToString(mRow["label"]),
			UserDataIndex: userDataIndex,
			UserDataDir:   userDataDir,
		}

		if bRow, ok := bindingMap[mappingId]; ok {
			item.IsBound = true
			item.BindingId = cast.ToInt(bRow["id"])
			item.Instruction = mcpBuildInstruction(req.McpType, cast.ToString(mRow["mapping_key"]), userDataDir)
		}

		items = append(items, item)
	}
	gsgin.GinResponseSuccess(c, "", items)
}

// McpBindingAdd 添加绑定并同步配置文件
func McpBindingAdd(c *gin.Context) {
	var req define.McpRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	if req.McpType == "" || req.MappingId <= 0 || req.AgentTargetId <= 0 {
		gsgin.GinResponseError(c, "参数不完整", nil)
		return
	}

	// 校验 mcp_type
	if _, ok := define.McpTypeDefs[req.McpType]; !ok {
		gsgin.GinResponseError(c, "不支持的 MCP 类型", nil)
		return
	}

	// 校验目录映射存在
	_, err := common.DbMain.Client.QueryBySql(
		`SELECT id FROM tbl_smart_link_directory_mapping WHERE id = ?`, req.MappingId,
	).One()
	if err != nil {
		gsgin.GinResponseError(c, "目录映射不存在", nil)
		return
	}

	// 校验目标智能体存在
	_, err = common.DbMain.Client.QueryBySql(
		`SELECT id FROM tbl_mcp_agent_target WHERE id = ?`, req.AgentTargetId,
	).One()
	if err != nil {
		gsgin.GinResponseError(c, "目标智能体不存在", nil)
		return
	}

	now := time.Now().Unix()
	_, err = common.DbMain.Client.QuickCreate(`tbl_mcp_binding`, map[string]any{
		`mcp_type`:        req.McpType,
		`mapping_id`:      req.MappingId,
		`agent_target_id`: req.AgentTargetId,
		`create_time`:     now,
		`update_time`:     now,
	}).Exec()
	if err != nil {
		gsgin.GinResponseError(c, "添加绑定失败: "+err.Error(), nil)
		return
	}

	// 同步配置文件
	if syncErr := business.SyncMcpConfig(req.AgentTargetId); syncErr != nil {
		gsgin.GinResponseError(c, "绑定已添加，但配置文件同步失败: "+syncErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// McpBindingRemove 移除绑定并同步配置文件
func McpBindingRemove(c *gin.Context) {
	var req define.McpRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	if req.BindingId <= 0 {
		gsgin.GinResponseError(c, "binding_id 不能为空", nil)
		return
	}

	// 删除前获取绑定信息
	bindRow, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_mcp_binding WHERE id = ?`, req.BindingId,
	).One()
	if err != nil || len(bindRow) == 0 {
		gsgin.GinResponseError(c, "绑定记录不存在", nil)
		return
	}
	agentTargetId := cast.ToInt(bindRow["agent_target_id"])

	_, err = common.DbMain.Client.QueryBySql(
		`DELETE FROM tbl_mcp_binding WHERE id = ?`, req.BindingId,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, "删除绑定失败: "+err.Error(), nil)
		return
	}

	// 同步配置文件
	if syncErr := business.SyncMcpConfig(agentTargetId); syncErr != nil {
		gsgin.GinResponseError(c, "绑定已删除，但配置文件同步失败: "+syncErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// McpBindingInstruction 返回绑定对应的 AI 使用说明文本
func McpBindingInstruction(c *gin.Context) {
	var req define.McpRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	if req.McpType == "" || req.MappingId <= 0 {
		gsgin.GinResponseError(c, "参数不完整", nil)
		return
	}

	mRow, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_smart_link_directory_mapping WHERE id = ?`, req.MappingId,
	).One()
	if err != nil || len(mRow) == 0 {
		gsgin.GinResponseError(c, "目录映射不存在", nil)
		return
	}

	userDataIndex := cast.ToInt(mRow["user_data_index"])
	userDataDir := ""
	if common.DbMain.Env != nil {
		userDataDir = filepath.Join(common.DbMain.Env.WebkitDataPath, cast.ToString(userDataIndex))
	}

	instruction := mcpBuildInstruction(req.McpType, cast.ToString(mRow["mapping_key"]), userDataDir)
	gsgin.GinResponseSuccess(c, "", instruction)
}

// McpAgentTargetList 返回所有目标智能体
func McpAgentTargetList(c *gin.Context) {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_mcp_agent_target ORDER BY id`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	items := make([]define.McpAgentTargetItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, define.McpAgentTargetItem{
			Id:             cast.ToInt(row["id"]),
			AgentName:      cast.ToString(row["agent_name"]),
			ConfigFilename: cast.ToString(row["config_filename"]),
			ConfigDir:      cast.ToString(row["config_dir"]),
			CreateTime:     cast.ToInt64(row["create_time"]),
			UpdateTime:     cast.ToInt64(row["update_time"]),
		})
	}
	gsgin.GinResponseSuccess(c, "", items)
}

// McpAgentTargetSave 新增或编辑目标智能体
func McpAgentTargetSave(c *gin.Context) {
	var req define.McpAgentTargetRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	if strings.TrimSpace(req.AgentName) == "" || strings.TrimSpace(req.ConfigFilename) == "" || strings.TrimSpace(req.ConfigDir) == "" {
		gsgin.GinResponseError(c, "参数不完整", nil)
		return
	}

	now := time.Now().Unix()
	if req.Id > 0 {
		// 编辑
		_, err := common.DbMain.Client.QuickUpdate(`tbl_mcp_agent_target`, map[string]any{
			`id`: req.Id,
		}, map[string]any{
			`agent_name`:      req.AgentName,
			`config_filename`: req.ConfigFilename,
			`config_dir`:      req.ConfigDir,
			`update_time`:     now,
		}).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	} else {
		// 新增
		_, err := common.DbMain.Client.QuickCreate(`tbl_mcp_agent_target`, map[string]any{
			`agent_name`:      req.AgentName,
			`config_filename`: req.ConfigFilename,
			`config_dir`:      req.ConfigDir,
			`create_time`:     now,
			`update_time`:     now,
		}).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// McpAgentTargetDelete 删除目标智能体（仅当无绑定时允许）
func McpAgentTargetDelete(c *gin.Context) {
	var req define.McpAgentTargetRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	if req.Id <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}

	// 检查是否存在绑定关系
	bindRow, _ := common.DbMain.Client.QueryBySql(
		`SELECT COUNT(*) as cnt FROM tbl_mcp_binding WHERE agent_target_id = ?`, req.Id,
	).One()
	if cast.ToInt(bindRow["cnt"]) > 0 {
		gsgin.GinResponseError(c, "该目标智能体下存在绑定关系，请先移除所有绑定", nil)
		return
	}

	_, err := common.DbMain.Client.QueryBySql(
		`DELETE FROM tbl_mcp_agent_target WHERE id = ?`, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// McpConfigPreview 返回配置文件的前后对比内容，用于用户确认
func McpConfigPreview(c *gin.Context) {
	var req define.McpRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	if req.AgentTargetId <= 0 {
		gsgin.GinResponseError(c, "agent_target_id 不能为空", nil)
		return
	}

	oldContent, newContent, err := business.GetMcpConfigPreview(req.AgentTargetId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", map[string]string{
		"old_content": oldContent,
		"new_content": newContent,
	})
}

// mcpBuildInstruction 生成 AI 使用说明文本
func mcpBuildInstruction(mcpType, mappingKey, userDataDir string) string {
	typeDef, ok := define.McpTypeDefs[mcpType]
	if !ok {
		return ""
	}
	return typeDef.Name + " MCP: " + mappingKey + "\n用户数据目录: " + userDataDir + "\n\n通过 MCP 的 " + strings.ReplaceAll(mcpType, "-", "_") + " 工具操作此浏览器，登录态独立隔离。"
}

const chromeDevtoolsConfigTable = `tbl_chrome_devtools_config`

// McpChromeDevtoolsConfigList 返回所有 Chrome DevTools 调试端口配置
func McpChromeDevtoolsConfigList(c *gin.Context) {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM ` + chromeDevtoolsConfigTable + ` ORDER BY id`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	items := make([]define.McpChromeDevtoolsConfigItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, define.McpChromeDevtoolsConfigItem{
			Id:         cast.ToInt(row["id"]),
			Name:       cast.ToString(row["name"]),
			Port:       cast.ToInt(row["port"]),
			Remark:     cast.ToString(row["remark"]),
			IsUsed:     cast.ToInt(row["is_used"]),
			CreateTime: cast.ToInt64(row["create_time"]),
			UpdateTime: cast.ToInt64(row["update_time"]),
		})
	}
	gsgin.GinResponseSuccess(c, "", items)
}

// McpChromeDevtoolsConfigSave 新增或编辑 Chrome DevTools 调试端口配置
func McpChromeDevtoolsConfigSave(c *gin.Context) {
	var req define.McpChromeDevtoolsConfigRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	if strings.TrimSpace(req.Name) == "" || req.Port <= 0 {
		gsgin.GinResponseError(c, "名称和端口不能为空", nil)
		return
	}

	now := time.Now().Unix()
	if req.Id > 0 {
		// 编辑：检查端口是否被其他记录占用
		existRow, _ := common.DbMain.Client.QueryBySql(
			`SELECT id FROM `+chromeDevtoolsConfigTable+` WHERE port = ? AND id != ?`, req.Port, req.Id,
		).One()
		if len(existRow) > 0 {
			gsgin.GinResponseError(c, "端口已被其他配置使用", nil)
			return
		}
		_, err := common.DbMain.Client.QuickUpdate(chromeDevtoolsConfigTable, map[string]any{
			`id`: req.Id,
		}, map[string]any{
			`name`:        req.Name,
			`port`:        req.Port,
			`remark`:      req.Remark,
			`update_time`: now,
		}).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	} else {
		// 新增：检查端口唯一性
		existRow, _ := common.DbMain.Client.QueryBySql(
			`SELECT id FROM `+chromeDevtoolsConfigTable+` WHERE port = ?`, req.Port,
		).One()
		if len(existRow) > 0 {
			gsgin.GinResponseError(c, "端口已存在", nil)
			return
		}
		_, err := common.DbMain.Client.QuickCreate(chromeDevtoolsConfigTable, map[string]any{
			`name`:        req.Name,
			`port`:        req.Port,
			`remark`:      req.Remark,
			`create_time`: now,
			`update_time`: now,
		}).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// McpChromeDevtoolsConfigDelete 删除 Chrome DevTools 调试端口配置
func McpChromeDevtoolsConfigDelete(c *gin.Context) {
	var req define.McpChromeDevtoolsConfigRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	if req.Id <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}
	_, err := common.DbMain.Client.QueryBySql(
		`DELETE FROM `+chromeDevtoolsConfigTable+` WHERE id = ?`, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// GetUnusedChromeDevtoolsPort 获取第一个未使用的调试端口并标记为已使用，返回该配置
func GetUnusedChromeDevtoolsPort() (*define.McpChromeDevtoolsConfigItem, error) {
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM ` + chromeDevtoolsConfigTable + ` WHERE is_used = 0 ORDER BY id`,
	).One()
	if err != nil {
		return nil, err
	}
	if len(row) == 0 {
		return nil, fmt.Errorf("没有可用的调试端口配置")
	}

	now := time.Now().Unix()
	_, err = common.DbMain.Client.QuickUpdate(chromeDevtoolsConfigTable, map[string]any{
		`id`: cast.ToInt(row["id"]),
	}, map[string]any{
		`is_used`:     cast.ToInt(1),
		`update_time`: now,
	}).Exec()
	if err != nil {
		return nil, fmt.Errorf("标记端口使用状态失败: %w", err)
	}

	return &define.McpChromeDevtoolsConfigItem{
		Id:     cast.ToInt(row["id"]),
		Name:   cast.ToString(row["name"]),
		Port:   cast.ToInt(row["port"]),
		Remark: cast.ToString(row["remark"]),
		IsUsed: cast.ToInt(1),
	}, nil
}

// McpChromeDevtoolsConfigToggleUsed 切换端口使用状态
func McpChromeDevtoolsConfigToggleUsed(c *gin.Context) {
	var req define.McpChromeDevtoolsConfigRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	if req.Id <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}

	// 查询当前状态
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT is_used FROM `+chromeDevtoolsConfigTable+` WHERE id = ?`, req.Id,
	).One()
	if err != nil || len(row) == 0 {
		gsgin.GinResponseError(c, "配置不存在", nil)
		return
	}

	currentUsed := cast.ToInt(row["is_used"])
	newUsed := cast.ToInt(0)
	if currentUsed == 0 {
		newUsed = cast.ToInt(1)
	}

	now := time.Now().Unix()
	_, err = common.DbMain.Client.QuickUpdate(chromeDevtoolsConfigTable, map[string]any{
		`id`: req.Id,
	}, map[string]any{
		`is_used`:     newUsed,
		`update_time`: now,
	}).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", map[string]int{
		"is_used": newUsed,
	})
}

// ReleaseChromeDevtoolsPort 释放调试端口
func ReleaseChromeDevtoolsPort(port int) {
	now := time.Now().Unix()
	_, _ = common.DbMain.Client.QuickUpdate(chromeDevtoolsConfigTable, map[string]any{
		`port`: port,
	}, map[string]any{
		`is_used`:     cast.ToInt(0),
		`update_time`: now,
	}).Exec()
}
