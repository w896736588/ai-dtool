package controller

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cast"
)

// normalizeAgentCliRuntimeType 统一运行态短值与 AgentCli 配置类型枚举，避免前后端使用不同命名时误判类型不一致。
func normalizeAgentCliRuntimeType(cliType string) string {
	switch strings.TrimSpace(cliType) {
	case `codex`, define.AgentCliTypeCodexCli:
		return define.AgentCliTypeCodexCli
	case `claude`, ``, define.AgentCliTypeClaudeCodeCli:
		return define.AgentCliTypeClaudeCodeCli
	default:
		return strings.TrimSpace(cliType)
	}
}

// validateAgentCliRuntimeConfig 校验当前执行所用配置文件与 AgentCli 记录是否一致。
func validateAgentCliRuntimeConfig(agentCliID int, cliType string) error {
	if agentCliID <= 0 {
		return fmt.Errorf(`agent_cli_id不能为空`)
	}
	normalizedCliType := normalizeAgentCliRuntimeType(cliType)
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_cli WHERE id = ?`, agentCliID,
	).One()
	if err != nil || len(row) == 0 {
		return fmt.Errorf(`Agent Cli 实例不存在`)
	}
	if strings.TrimSpace(cast.ToString(row[`type`])) != normalizedCliType {
		return fmt.Errorf(`Agent Cli 类型与当前执行器不一致`)
	}
	switch normalizedCliType {
	case define.AgentCliTypeCodexCli:
		return validateCodexAgentCliRuntimeConfig(row)
	case define.AgentCliTypeClaudeCodeCli:
		return validateClaudeAgentCliRuntimeConfig(row)
	default:
		return fmt.Errorf(`不支持的 Agent Cli 类型: %s`, cliType)
	}
}

func validateCodexAgentCliRuntimeConfig(row map[string]any) error {
	cfg, err := business.GetCodexCliConfig(cast.ToString(row[`config`]))
	if err != nil {
		return err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf(`获取用户目录失败: %w`, err)
	}
	configPath := filepath.Join(homeDir, `.codex`, `config.toml`)
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf(`读取 Codex 配置失败 %s: %w`, configPath, err)
	}
	var tomlData map[string]any
	if err := toml.Unmarshal(content, &tomlData); err != nil {
		return fmt.Errorf(`解析 Codex 配置失败 %s: %w`, configPath, err)
	}
	if cast.ToString(tomlData[`model`]) != cfg.Model {
		return fmt.Errorf(`Codex 配置模型与当前 Agent Cli 不一致`)
	}
	if cfg.BaseURL != `` {
		providerName := cast.ToString(tomlData[`model_provider`])
		if providerName == `` {
			return fmt.Errorf(`Codex 配置未启用自定义 provider`)
		}
		providers, ok := tomlData[`model_providers`].(map[string]any)
		if !ok {
			return fmt.Errorf(`Codex 配置缺少 model_providers`)
		}
		provider, ok := providers[providerName].(map[string]any)
		if !ok {
			return fmt.Errorf(`Codex 配置缺少当前 provider 段`)
		}
		if cast.ToString(provider[`base_url`]) != cfg.BaseURL {
			return fmt.Errorf(`Codex 配置 base_url 与当前 Agent Cli 不一致`)
		}
	}
	if strings.TrimSpace(cfg.ApiKey) != `` {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf(`获取用户目录失败: %w`, err)
		}
		authPath := filepath.Join(homeDir, `.codex`, `auth.json`)
		authContent, err := os.ReadFile(authPath)
		if err != nil {
			return fmt.Errorf(`读取 Codex 认证文件失败 %s: %w`, authPath, err)
		}
		var authData map[string]any
		if err := json.Unmarshal(authContent, &authData); err != nil {
			return fmt.Errorf(`解析 Codex 认证文件失败 %s: %w`, authPath, err)
		}
		if cast.ToString(authData[`OPENAI_API_KEY`]) != cfg.ApiKey {
			return fmt.Errorf(`Codex API Key 与当前 Agent Cli 不一致`)
		}
	}
	return nil
}

func validateClaudeAgentCliRuntimeConfig(row map[string]any) error {
	settingsPath := cast.ToString(row[`settings_path`])
	if strings.TrimSpace(settingsPath) == `` {
		return fmt.Errorf(`settings.json 路径不能为空`)
	}
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf(`读取 settings.json 失败 %s: %w`, settingsPath, err)
	}
	if len(content) == 0 {
		return fmt.Errorf(`settings.json 为空: %s`, settingsPath)
	}
	var settingsData map[string]any
	if err := json.Unmarshal(content, &settingsData); err != nil {
		return fmt.Errorf(`解析 settings.json 失败 %s: %w`, settingsPath, err)
	}
	return nil
}
