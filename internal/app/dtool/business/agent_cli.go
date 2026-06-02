package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cast"
)

// normalizeAgentCliModels 清洗模型列表，去重并过滤空值。
// normalizeAgentCliModels trims model names, removes duplicates, and drops empty items.
func normalizeAgentCliModels(models []string) []string {
	result := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, rawModel := range models {
		modelName := strings.TrimSpace(rawModel)
		if modelName == "" {
			continue
		}
		if _, exists := seen[modelName]; exists {
			continue
		}
		seen[modelName] = struct{}{}
		result = append(result, modelName)
	}
	return result
}

// mergeAgentCliModels 合并当前模型与候选模型列表，保证当前模型优先展示。
// mergeAgentCliModels merges the current model into candidate models and keeps it first.
func mergeAgentCliModels(currentModel string, models []string) []string {
	normalizedModels := normalizeAgentCliModels(models)
	currentModel = strings.TrimSpace(currentModel)
	if currentModel == "" {
		return normalizedModels
	}
	if slices.Contains(normalizedModels, currentModel) {
		return normalizedModels
	}
	return append([]string{currentModel}, normalizedModels...)
}

// claude-mem 插件在 enabledPlugins 中的 key
const claudeMemPluginKey = "claude-mem@thedotmack"

// ReadAgentCliSettings 读取 settings.json 文件内容，若文件不存在返回空
func ReadAgentCliSettings(settingsPath string) (exists bool, content string, err error) {
	data, readErr := os.ReadFile(settingsPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("读取配置文件失败 %s: %w", settingsPath, readErr)
	}
	return true, string(data), nil
}

// WriteMcpServersToSettings 从 DB 读取全部 ChromeDevtools 端口写入 settings.json 的 mcpServers
func WriteMcpServersToSettings(settingsPath string) error {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_chrome_devtools_config ORDER BY id`,
	).All()
	if err != nil {
		return fmt.Errorf("查询端口配置失败: %w", err)
	}

	// 读取已有配置
	configData := make(map[string]any)
	content, readErr := os.ReadFile(settingsPath)
	if readErr == nil && len(content) > 0 {
		if err := json.Unmarshal(content, &configData); err != nil {
			return fmt.Errorf("解析配置文件失败 %s: %w", settingsPath, err)
		}
	}

	// 构建 mcpServers
	mcpServers := make(map[string]any)
	for i, row := range rows {
		port := cast.ToInt(row["port"])
		key := fmt.Sprintf("devtool_%d", i)
		mcpServers[key] = map[string]any{
			"command": "npx",
			"args": []string{
				"chrome-devtools-mcp@latest",
				fmt.Sprintf("--browser-url=http://127.0.0.1:%d", port),
			},
		}
	}
	configData["mcpServers"] = mcpServers

	// 确保目录存在
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 %s: %w", dir, err)
	}

	// 写回
	newContent, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	newContent = append(newContent, '\n')

	if err := os.WriteFile(settingsPath, newContent, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败 %s: %w", settingsPath, err)
	}
	return nil
}

// WriteDeepSeekToSettings 写入 DeepSeek 配置到 settings.json。
// modelList 为可选模型列表，首个模型会被写入 settings.json 顶层 model，兼容旧执行链路。
// WriteDeepSeekToSettings writes DeepSeek settings and persists modelList while keeping the first model as legacy default.
func WriteDeepSeekToSettings(settingsPath string, modelName string, modelList []string, apiKey string, baseUrl string) error {
	// 读取已有配置
	configData := make(map[string]any)
	content, readErr := os.ReadFile(settingsPath)
	if readErr == nil && len(content) > 0 {
		if err := json.Unmarshal(content, &configData); err != nil {
			return fmt.Errorf("解析配置文件失败 %s: %w", settingsPath, err)
		}
	}

	existingModel, existingBaseURL, existingAPIKey := GetAgentCliModelConfig(string(content))
	if strings.TrimSpace(modelName) == "" {
		modelName = existingModel
	}
	if strings.TrimSpace(apiKey) == "" {
		apiKey = existingAPIKey
	}
	if strings.TrimSpace(baseUrl) == "" {
		baseUrl = existingBaseURL
	}
	if baseUrl == "" {
		baseUrl = "https://api.deepseek.com/anthropic"
	}

	modelList = mergeAgentCliModels(modelName, modelList)
	if modelName == "" && len(modelList) > 0 {
		modelName = modelList[0]
	}

	// 设置 env
	configData["env"] = map[string]string{
		"ANTHROPIC_BASE_URL":                       baseUrl,
		"API_TIMEOUT_MS":                           "3000000",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":            modelName,
		"ANTHROPIC_DEFAULT_SONNET_MODEL":           modelName,
		"ANTHROPIC_DEFAULT_OPUS_MODEL":             modelName,
		"ANTHROPIC_AUTH_TOKEN":                     apiKey,
		"ANTHROPIC_API_KEY":                        "",
		"ANTHROPIC_REASONING_MODEL":                modelName,
	}

	// 设置 model
	configData["model"] = modelName
	if len(modelList) > 0 {
		configData["dtool_models"] = modelList
	} else {
		delete(configData, "dtool_models")
	}

	// 合并 enabledPlugins，保留已有启用的插件
	plugins := make(map[string]bool)
	if existing, ok := configData["enabledPlugins"]; ok {
		if existingMap, ok := existing.(map[string]any); ok {
			for k, v := range existingMap {
				plugins[k] = cast.ToBool(v)
			}
		}
	}
	plugins["skill-creator@claude-plugins-official"] = true
	configData["enabledPlugins"] = plugins

	// 设置 extraKnownMarketplaces
	configData["extraKnownMarketplaces"] = map[string]any{
		"anthropic-agent-skills": map[string]any{
			"source": map[string]any{
				"source": "git",
				"url":    "https://github.com/anthropics/skills.git",
			},
		},
	}

	// 确保目录存在
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 %s: %w", dir, err)
	}

	// 写回
	newContent, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	newContent = append(newContent, '\n')

	if err := os.WriteFile(settingsPath, newContent, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败 %s: %w", settingsPath, err)
	}
	return nil
}

// GetAgentCliSettingsSummary 从 settings.json 内容中提取摘要信息
func GetAgentCliSettingsSummary(content string) (model string, mcpCount int, claudeMemEnabled bool) {
	if content == "" {
		return "", 0, false
	}
	var configData map[string]any
	if err := json.Unmarshal([]byte(content), &configData); err != nil {
		return "", 0, false
	}
	if m, ok := configData["model"]; ok {
		model = cast.ToString(m)
	}
	if servers, ok := configData["mcpServers"]; ok {
		if serverMap, ok := servers.(map[string]any); ok {
			mcpCount = len(serverMap)
		}
	}
	claudeMemEnabled = IsClaudeMemEnabled(configData)
	return
}

// IsClaudeMemEnabled 判断 claude-mem 插件是否已启用
func IsClaudeMemEnabled(configData map[string]any) bool {
	if plugins, ok := configData["enabledPlugins"]; ok {
		if pluginMap, ok := plugins.(map[string]any); ok {
			if enabled, ok := pluginMap[claudeMemPluginKey]; ok {
				return cast.ToBool(enabled)
			}
		}
	}
	return false
}

// ToggleClaudeMem 启停 settings.json 中的 claude-mem 插件
func ToggleClaudeMem(settingsPath string, enable bool) error {
	configData := make(map[string]any)
	content, readErr := os.ReadFile(settingsPath)
	if readErr == nil && len(content) > 0 {
		if err := json.Unmarshal(content, &configData); err != nil {
			return fmt.Errorf("解析配置文件失败 %s: %w", settingsPath, err)
		}
	}

	plugins, ok := configData["enabledPlugins"]
	if !ok {
		configData["enabledPlugins"] = map[string]bool{
			claudeMemPluginKey: enable,
		}
	} else {
		pluginMap, ok := plugins.(map[string]any)
		if !ok {
			configData["enabledPlugins"] = map[string]bool{
				claudeMemPluginKey: enable,
			}
		} else {
			pluginMap[claudeMemPluginKey] = enable
		}
	}

	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 %s: %w", dir, err)
	}

	newContent, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	newContent = append(newContent, '\n')

	if err := os.WriteFile(settingsPath, newContent, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败 %s: %w", settingsPath, err)
	}
	return nil
}

// GetCodexCliConfig 从 DB config JSON 字段解析 Codex CLI 配置
func GetCodexCliConfig(configJson string) (*define.CodexCliConfig, error) {
	if configJson == "" {
		return nil, fmt.Errorf("Codex CLI 配置为空")
	}
	var cfg define.CodexCliConfig
	if err := json.Unmarshal([]byte(configJson), &cfg); err != nil {
		return nil, fmt.Errorf("解析 Codex CLI 配置失败: %w", err)
	}
	cfg.Models = mergeAgentCliModels(cfg.Model, cfg.Models)
	if cfg.Model == "" && len(cfg.Models) > 0 {
		cfg.Model = cfg.Models[0]
	}
	cfg.WireAPI = strings.TrimSpace(cfg.WireAPI)
	if cfg.WireAPI == "" {
		cfg.WireAPI = define.CodexCliDefaultWireAPI
	}
	return &cfg, nil
}

// GetCodexCliModelConfig 从 config JSON 中提取 Codex CLI 模型和请求地址。 // GetCodexCliModelConfig extracts the Codex CLI model and request URL from config JSON.
func GetCodexCliModelConfig(configJson string) (model string, baseURL string) {
	if configJson == "" {
		return "", ""
	}
	var cfg define.CodexCliConfig
	if err := json.Unmarshal([]byte(configJson), &cfg); err != nil {
		return "", ""
	}
	cfg.Models = mergeAgentCliModels(cfg.Model, cfg.Models)
	if cfg.Model == "" && len(cfg.Models) > 0 {
		cfg.Model = cfg.Models[0]
	}
	return cfg.Model, cfg.BaseURL
}

// GetCodexCliModelOptions 从 config JSON 中提取 Codex CLI 可选模型列表。
// GetCodexCliModelOptions extracts selectable model options from config JSON.
func GetCodexCliModelOptions(configJson string) []string {
	cfg, err := GetCodexCliConfig(configJson)
	if err != nil || cfg == nil {
		return nil
	}
	return append([]string{}, cfg.Models...)
}

// codexModelProviderKey Codex CLI config.toml 中自定义 API 提供商的 key。
// codexModelProviderKey is the custom provider key used in Codex config.toml.
const codexModelProviderKey = "myprovider"

// WriteCodexConfigToToml 将 Codex CLI 的 model/provider 写入 ~/.codex/config.toml。
// WriteCodexConfigToToml writes Codex CLI model/provider settings to ~/.codex/config.toml.
func WriteCodexConfigToToml(cfg *define.CodexCliConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}
	configPath := filepath.Join(homeDir, ".codex", "config.toml")

	// 读取已有配置
	content := ""
	data, readErr := os.ReadFile(configPath)
	if readErr == nil {
		content = string(data)
	}

	// 更新 model 行 / Update the active model.
	if cfg.Model != "" {
		content = setTomlTopLevelField(content, "model", cfg.Model)
	}

	// 处理自定义 provider 模式 / Configure the custom provider block.
	content = removeTomlTopLevelField(content, "openai_base_url")
	content = removeTomlModelProviderSection(content, codexModelProviderKey)
	if cfg.BaseURL != "" {
		content = setTomlTopLevelField(content, "model_provider", codexModelProviderKey)
		content = appendCodexModelProviderSection(content, codexModelProviderKey, cfg.BaseURL, cfg.ApiKey, cfg.WireAPI, cfg.SupportsWebsockets)
	} else {
		content = removeTomlTopLevelField(content, "model_provider")
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 %s: %w", dir, err)
	}

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入 config.toml 失败 %s: %w", configPath, err)
	}
	return nil
}

// appendCodexModelProviderSection 追加 Codex 自定义 provider 段。
// appendCodexModelProviderSection appends the Codex custom provider section.
// 使用 api_key 直接写入密钥而非 env_key，避免 Codex CLI 要求必须设置环境变量。
func appendCodexModelProviderSection(content, providerKey, baseURL, apiKey, wireAPI string, supportsWebsockets *bool) string {
	wireAPI = strings.TrimSpace(wireAPI)
	if wireAPI == "" {
		wireAPI = define.CodexCliDefaultWireAPI
	}
	section := fmt.Sprintf("\n[model_providers.%s]\nname = \"My Local Proxy\"\nbase_url = \"%s\"\nwire_api = \"%s\"\napi_key = \"%s\"\n", providerKey, baseURL, wireAPI, apiKey)
	if supportsWebsockets != nil {
		section += fmt.Sprintf("supports_websockets = %t\n", *supportsWebsockets)
	}
	return content + section
}

// WriteCodexAuthJson 将 API Key 写入 ~/.codex/auth.json 并切换认证模式为 api-key。
// 仅在 base_url 非空时调用，因为自定义 API 端点需要 API Key 认证而非 chatgpt OAuth。
func WriteCodexAuthJson(apiKey string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}
	authPath := filepath.Join(homeDir, ".codex", "auth.json")

	// 读取已有配置，保留原有字段（如 tokens）以便切换回 chatgpt 模式
	authData := make(map[string]any)
	content, readErr := os.ReadFile(authPath)
	if readErr == nil && len(content) > 0 {
		if err := json.Unmarshal(content, &authData); err != nil {
			return fmt.Errorf("解析 auth.json 失败 %s: %w", authPath, err)
		}
	}

	// 切换认证模式并写入 API Key（Codex CLI 认证模式枚举值：apikey / chatgpt / chatgptAuthTokens）
	authData["auth_mode"] = "apikey"
	authData["OPENAI_API_KEY"] = apiKey

	// 确保目录存在
	dir := filepath.Dir(authPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 %s: %w", dir, err)
	}

	newContent, err := json.MarshalIndent(authData, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 auth.json 失败: %w", err)
	}
	newContent = append(newContent, '\n')

	if err := os.WriteFile(authPath, newContent, 0644); err != nil {
		return fmt.Errorf("写入 auth.json 失败 %s: %w", authPath, err)
	}
	return nil
}

// setTomlTopLevelField 设置 TOML 顶层字符串字段，存在则更新，不存在则在第一个 [section] 前插入
func setTomlTopLevelField(content, key, value string) string {
	lines := strings.Split(content, "\n")
	found := false
	newLine := key + ` = "` + value + `"`

	for i, line := range lines {
		// 遇到 [section] 说明已过顶层，停止扫描
		if strings.HasPrefix(strings.TrimSpace(line), "[") {
			break
		}
		trimmed := strings.TrimSpace(line)
		afterKey := strings.TrimPrefix(trimmed, key)
		if afterKey != trimmed {
			afterKey = strings.TrimSpace(afterKey)
			if strings.HasPrefix(afterKey, "=") {
				lines[i] = newLine
				found = true
				break
			}
		}
	}

	if !found {
		// 在第一个 [section] 之前插入
		insertIdx := len(lines)
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "[") {
				insertIdx = i
				break
			}
		}
		lines = append(lines[:insertIdx], append([]string{newLine}, lines[insertIdx:]...)...)
	}

	return strings.Join(lines, "\n")
}

// removeTomlTopLevelField 移除 TOML 顶层字符串字段
func removeTomlTopLevelField(content, key string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "[") {
			break
		}
		trimmed := strings.TrimSpace(line)
		afterKey := strings.TrimPrefix(trimmed, key)
		if afterKey != trimmed {
			afterKey = strings.TrimSpace(afterKey)
			if strings.HasPrefix(afterKey, "=") {
				lines = append(lines[:i], lines[i+1:]...)
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

// setTomlModelProviderSection 设置 [model_providers.xxx] 段的 base_url 字段。
// 若段不存在则在文件末尾追加完整段。
func setTomlModelProviderSection(content, providerKey, baseURL string) string {
	sectionHeader := fmt.Sprintf("[model_providers.%s]", providerKey)
	lines := strings.Split(content, "\n")

	// 查找段起始位置
	sectionStart := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == sectionHeader {
			sectionStart = i
			break
		}
	}

	if sectionStart >= 0 {
		// 段已存在，查找并更新 base_url 行
		baseURLLine := fmt.Sprintf(`base_url = "%s"`, baseURL)
		found := false
		for i := sectionStart + 1; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])
			// 遇到下一个段，停止
			if strings.HasPrefix(trimmed, "[") {
				break
			}
			afterKey := strings.TrimPrefix(trimmed, "base_url")
			if afterKey != trimmed {
				afterKey = strings.TrimSpace(afterKey)
				if strings.HasPrefix(afterKey, "=") {
					lines[i] = baseURLLine
					found = true
					break
				}
			}
		}
		if !found {
			// 在段头之后插入 base_url
			lines = append(lines[:sectionStart+1], append([]string{baseURLLine}, lines[sectionStart+1:]...)...)
		}
		return strings.Join(lines, "\n")
	}

	// 段不存在，追加到末尾
	newSection := fmt.Sprintf("\n%s\nname = \"Custom API\"\nwire_api = \"responses\"\nbase_url = \"%s\"\n", sectionHeader, baseURL)
	return content + newSection
}

// removeTomlModelProviderSection 移除 [model_providers.xxx] 段（含段前空行）
func removeTomlModelProviderSection(content, providerKey string) string {
	sectionHeader := fmt.Sprintf("[model_providers.%s]", providerKey)
	lines := strings.Split(content, "\n")

	sectionStart := -1
	sectionEnd := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == sectionHeader {
			sectionStart = i
			continue
		}
		if sectionStart >= 0 && sectionEnd < 0 {
			if strings.HasPrefix(strings.TrimSpace(line), "[") {
				sectionEnd = i
				break
			}
		}
	}

	if sectionStart < 0 {
		return content // 段不存在，无需处理
	}
	if sectionEnd < 0 {
		sectionEnd = len(lines)
	}

	// 移除段前空行
	start := sectionStart
	if start > 0 && strings.TrimSpace(lines[start-1]) == "" {
		start--
	}

	newLines := append(lines[:start], lines[sectionEnd:]...)
	return strings.Join(newLines, "\n")
}

// WriteMcpServersToCodexConfig 从 DB 读取全部 ChromeDevtools 端口写入 ~/.codex/config.toml 的 [mcp_servers.*] 段
func WriteMcpServersToCodexConfig() error {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_chrome_devtools_config ORDER BY id`,
	).All()
	if err != nil {
		return fmt.Errorf("查询端口配置失败: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}
	configPath := filepath.Join(homeDir, ".codex", "config.toml")

	// 读取已有配置
	content := ""
	data, readErr := os.ReadFile(configPath)
	if readErr == nil {
		content = string(data)
	}

	// 移除所有已有的 devtool_ 开头的 mcp_servers 段
	content = removeAllTomlMcpDevtoolSections(content)

	// 追加新的 mcp_servers 段
	for i, row := range rows {
		port := cast.ToInt(row["port"])
		serverKey := fmt.Sprintf("devtool_%d", i)
		content = appendCodexMcpServerSection(content, serverKey, port)
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 %s: %w", dir, err)
	}

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入 config.toml 失败 %s: %w", configPath, err)
	}
	return nil
}

// appendCodexMcpServerSection 追加一个 Codex MCP server 段到 config.toml
func appendCodexMcpServerSection(content, serverKey string, port int) string {
	section := fmt.Sprintf("\n[mcp_servers.%s]\ncommand = \"npx\"\nargs = [\"chrome-devtools-mcp@latest\", \"--browser-url=http://127.0.0.1:%d\"]\n", serverKey, port)
	return content + section
}

// removeAllTomlMcpDevtoolSections 移除 config.toml 中所有 [mcp_servers.devtool_*] 段
func removeAllTomlMcpDevtoolSections(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))
	inDevtoolSection := false

	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		// 检测是否进入 devtool_ 的 mcp_servers 段
		if strings.HasPrefix(trimmed, "[mcp_servers.devtool_") && strings.HasSuffix(trimmed, "]") {
			inDevtoolSection = true
			// 移除段前空行
			if len(result) > 0 && strings.TrimSpace(result[len(result)-1]) == "" {
				result = result[:len(result)-1]
			}
			continue
		}
		// 遇到新的 section 头则退出 devtool 段
		if inDevtoolSection && strings.HasPrefix(trimmed, "[") {
			inDevtoolSection = false
		}
		if !inDevtoolSection {
			result = append(result, lines[i])
		}
	}

	return strings.Join(result, "\n")
}

// GetCodexMcpServerCount 读取 ~/.codex/config.toml 中 [mcp_servers.*] 段的数量
func GetCodexMcpServerCount() int {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0
	}
	configPath := filepath.Join(homeDir, ".codex", "config.toml")
	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		return 0
	}
	content := string(data)
	count := 0
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[mcp_servers.") && strings.HasSuffix(trimmed, "]") {
			count++
		}
	}
	return count
}

// GetAgentCliModelConfig 从 settings.json 内容中提取模型连接配置。
func GetAgentCliModelConfig(content string) (model string, baseURL string, apiKey string) {
	if content == "" {
		return "", "", ""
	}
	var configData map[string]any
	if err := json.Unmarshal([]byte(content), &configData); err != nil {
		return "", "", ""
	}
	if m, ok := configData["model"]; ok {
		model = cast.ToString(m)
	}
	if env, ok := configData["env"]; ok {
		if envMap, ok := env.(map[string]any); ok {
			if url, ok := envMap["ANTHROPIC_BASE_URL"]; ok {
				baseURL = cast.ToString(url)
			}
			if key, ok := envMap["ANTHROPIC_AUTH_TOKEN"]; ok {
				apiKey = cast.ToString(key)
			}
		}
	}
	return
}

// GetAgentCliModelOptions 从 settings.json 中提取可选模型列表，并兼容旧版仅保存单模型配置。
// GetAgentCliModelOptions extracts selectable model options from settings.json and remains compatible with legacy single-model config.
func GetAgentCliModelOptions(content string) []string {
	if content == "" {
		return nil
	}
	var configData map[string]any
	if err := json.Unmarshal([]byte(content), &configData); err != nil {
		return nil
	}
	models := make([]string, 0)
	if rawModels, ok := configData["dtool_models"]; ok {
		switch value := rawModels.(type) {
		case []any:
			for _, item := range value {
				models = append(models, cast.ToString(item))
			}
		case []string:
			models = append(models, value...)
		}
	}
	currentModel, _, _ := GetAgentCliModelConfig(content)
	return mergeAgentCliModels(currentModel, models)
}
