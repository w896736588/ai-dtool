package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cast"
)

// mcpConfigSyncMu 防止对同一目标智能体配置文件的并发写入
var mcpConfigSyncMu sync.Mutex

// McpServerEntry 配置文件中 mcpServers 的一个条目
type McpServerEntry struct {
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// SyncMcpConfig 重建目标智能体配置文件的 mcpServers 段。
// 在添加/移除绑定操作后调用。
func SyncMcpConfig(agentTargetId int) error {
	mcpConfigSyncMu.Lock()
	defer mcpConfigSyncMu.Unlock()

	// 查询目标智能体信息
	targetRow, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_mcp_agent_target WHERE id = ?`, agentTargetId,
	).One()
	if err != nil || len(targetRow) == 0 {
		return fmt.Errorf("agent target not found: %d", agentTargetId)
	}
	configDir := cast.ToString(targetRow["config_dir"])
	configFilename := cast.ToString(targetRow["config_filename"])
	if configDir == "" || configFilename == "" {
		return fmt.Errorf("config path not configured for agent target %d", agentTargetId)
	}

	// 查询该目标智能体下所有绑定，关联回目录映射表
	rows, err := common.DbMain.Client.QueryBySql(`
		SELECT b.mcp_type, m.mapping_key, m.user_data_index
		FROM tbl_mcp_binding b
		INNER JOIN tbl_smart_link_directory_mapping m ON b.mapping_id = m.id
		WHERE b.agent_target_id = ?
		ORDER BY b.id
	`, agentTargetId).All()
	if err != nil {
		return fmt.Errorf("query bindings failed: %w", err)
	}

	// 构建 mcpServers 映射
	mcpServers := make(map[string]McpServerEntry)
	for _, row := range rows {
		mcpType := cast.ToString(row["mcp_type"])
		mappingKey := cast.ToString(row["mapping_key"])
		userDataIndex := cast.ToInt(row["user_data_index"])

		typeDef, ok := define.McpTypeDefs[mcpType]
		if !ok {
			continue
		}

		userDataDir := ""
		if common.DbMain.Env != nil {
			userDataDir = filepath.Join(common.DbMain.Env.WebkitDataPath, cast.ToString(userDataIndex))
		}

		mcpServers[mappingKey] = McpServerEntry{
			Type:    "stdio",
			Command: typeDef.Command,
			Args:    []string{typeDef.PackageName},
			Env: map[string]string{
				"CHROME_USER_DATA_DIR": userDataDir,
			},
		}
	}

	// 读取已有配置文件（不存在则创建空对象）
	configPath := filepath.Join(configDir, configFilename)
	return writeMcpServers(configPath, mcpServers)
}

// writeMcpServers 读取 JSON 配置文件，更新 mcpServers 字段后写回
func writeMcpServers(configPath string, mcpServers map[string]McpServerEntry) error {
	configData := make(map[string]any)

	// 尝试读取已有文件
	content, readErr := os.ReadFile(configPath)
	if readErr == nil && len(content) > 0 {
		if err := json.Unmarshal(content, &configData); err != nil {
			return fmt.Errorf("parse config file failed %s: %w", configPath, err)
		}
	}

	// 更新 mcpServers 字段
	configData["mcpServers"] = mcpServers

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir failed %s: %w", dir, err)
	}

	// 格式化写回
	newContent, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config failed: %w", err)
	}
	newContent = append(newContent, '\n')

	return os.WriteFile(configPath, newContent, 0644)
}

// GetMcpConfigPreview 返回配置文件的当前内容和即将写入的新内容，用于用户确认
func GetMcpConfigPreview(agentTargetId int) (oldContent string, newContent string, err error) {
	targetRow, dbErr := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_mcp_agent_target WHERE id = ?`, agentTargetId,
	).One()
	if dbErr != nil || len(targetRow) == 0 {
		err = fmt.Errorf("agent target not found: %d", agentTargetId)
		return
	}
	configDir := cast.ToString(targetRow["config_dir"])
	configFilename := cast.ToString(targetRow["config_filename"])
	configPath := filepath.Join(configDir, configFilename)

	// 读取旧内容
	content, readErr := os.ReadFile(configPath)
	if readErr == nil {
		oldContent = string(content)
	}

	// 构建新内容
	rows, qErr := common.DbMain.Client.QueryBySql(`
		SELECT b.mcp_type, m.mapping_key, m.user_data_index
		FROM tbl_mcp_binding b
		INNER JOIN tbl_smart_link_directory_mapping m ON b.mapping_id = m.id
		WHERE b.agent_target_id = ?
		ORDER BY b.id
	`, agentTargetId).All()
	if qErr != nil {
		err = qErr
		return
	}

	mcpServers := make(map[string]McpServerEntry)
	for _, row := range rows {
		mcpType := cast.ToString(row["mcp_type"])
		mappingKey := cast.ToString(row["mapping_key"])
		userDataIndex := cast.ToInt(row["user_data_index"])

		typeDef, ok := define.McpTypeDefs[mcpType]
		if !ok {
			continue
		}
		userDataDir := ""
		if common.DbMain.Env != nil {
			userDataDir = filepath.Join(common.DbMain.Env.WebkitDataPath, cast.ToString(userDataIndex))
		}
		mcpServers[mappingKey] = McpServerEntry{
			Type:    "stdio",
			Command: typeDef.Command,
			Args:    []string{typeDef.PackageName},
			Env: map[string]string{
				"CHROME_USER_DATA_DIR": userDataDir,
			},
		}
	}

	configData := make(map[string]any)
	if oldContent != "" {
		_ = json.Unmarshal([]byte(oldContent), &configData)
	}
	configData["mcpServers"] = mcpServers

	newBytes, marshalErr := json.MarshalIndent(configData, "", "  ")
	if marshalErr != nil {
		err = marshalErr
		return
	}
	newContent = string(newBytes) + "\n"
	return
}
