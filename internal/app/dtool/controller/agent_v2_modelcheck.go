package controller

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dev_tool/internal/app/dtool/agent"
	"dev_tool/internal/app/dtool/common"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// AgentV2ModelTest 测试指定模型的可用性
// 通过启动 Pi Agent 发送简单提示词，验证 Provider + Model 是否可正常调用
func AgentV2ModelTest(c *gin.Context) {
	var req struct {
		ProviderId int `json:"provider_id"`
		ModelId    int `json:"model_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProviderId <= 0 || req.ModelId <= 0 {
		gsgin.GinResponseError(c, "参数错误：provider_id 和 model_id 必填", nil)
		return
	}

	// 查询 Provider（name 用于 Pi --provider 参数，provider_type 仅决定 api 类型）
	providerRow, err := common.DbMain.Client.QueryBySql(
		`SELECT name, provider_type, base_url, api_key FROM tbl_ai_provider WHERE id = ? AND status = 1`,
		req.ProviderId,
	).One()
	if err != nil || len(providerRow) == 0 {
		gsgin.GinResponseError(c, "Provider 不存在或已禁用", nil)
		return
	}

	// 查询 Model
	modelRow, err := common.DbMain.Client.QueryBySql(
		`SELECT model, name FROM tbl_ai_model WHERE id = ? AND provider_id = ? AND status = 1`,
		req.ModelId, req.ProviderId,
	).One()
	if err != nil || len(modelRow) == 0 {
		gsgin.GinResponseError(c, "模型不存在或已禁用", nil)
		return
	}

	providerName := cast.ToString(providerRow["name"])
	modelName := cast.ToString(modelRow["model"])

	// 同步 models.json（永久化配置，所有 Provider/Model 都注册）
	if err := syncPiModelsConfig(); err != nil {
		gsgin.GinResponseError(c, "同步 Pi 模型配置失败: "+err.Error(), nil)
		return
	}

	// 创建临时会话目录
	tmpDir, err := os.MkdirTemp("", "pi_model_test_*")
	if err != nil {
		gsgin.GinResponseError(c, "创建临时目录失败", nil)
		return
	}
	defer os.RemoveAll(tmpDir)

	adapter := agent.NewPiAdapter()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()

	// --provider 使用 provider name（唯一标识），models.json 包含 baseUrl/apiKey/api 全部信息
	if err := adapter.Start(ctx, agent.AgentStartConfig{
		SessionDir: tmpDir,
		Provider:   providerName,
		Model:      modelName,
	}); err != nil {
		gsgin.GinResponseError(c, "启动 Pi 失败: "+err.Error(), nil)
		return
	}
	defer adapter.Stop()

	// 发送测试提示词
	promptCmd, _ := json.Marshal(map[string]string{
		"type":    "prompt",
		"message": "1+1=? 只回答结果",
	})
	if err := adapter.SendCommand(promptCmd); err != nil {
		gsgin.GinResponseError(c, "发送测试提示词失败: "+err.Error(), nil)
		return
	}

	// 读取事件直到收到完整响应或超时
	var responseText string
	timeout := time.After(15 * time.Second)

	for {
		select {
		case evt, ok := <-adapter.Events():
			if !ok {
				if exitErr := adapter.ExitError(); exitErr != nil {
					gsgin.GinResponseError(c, "Pi 进程异常退出: "+exitErr.Error(), nil)
				} else if responseText != "" {
					gsgin.GinResponseSuccess(c, "ok", gin.H{
						"response": strings.TrimSpace(responseText),
						"model":    modelName,
						"provider": providerName,
					})
				} else {
					gsgin.GinResponseError(c, "未收到模型响应", nil)
				}
				return
			}

			var rawEvt map[string]interface{}
			if err := json.Unmarshal(evt.Raw, &rawEvt); err != nil {
				continue
			}

			evtType := cast.ToString(rawEvt["type"])

			switch evtType {
			case "message_update":
				msgEvt, _ := rawEvt["assistantMessageEvent"].(map[string]interface{})
				if msgEvt == nil {
					continue
				}
				subType := cast.ToString(msgEvt["type"])
				switch subType {
				case "text_delta":
					responseText += cast.ToString(msgEvt["delta"])
				case "thinking_delta":
					responseText += cast.ToString(msgEvt["delta"])
				case "error":
					errMsg := cast.ToString(msgEvt["reason"])
					if errMsg == "" {
						errMsg = cast.ToString(msgEvt["message"])
					}
					if errMsg == "" {
						errMsg = "未知错误"
					}
					gsgin.GinResponseError(c, "模型返回错误: "+errMsg, nil)
					return
				default:
					log.Printf("[agent-v2/model-test] unhandled message_update sub-type: %s", subType)
				}
			case "turn_end":
				if responseText != "" {
					gsgin.GinResponseSuccess(c, "ok", gin.H{
						"response": strings.TrimSpace(responseText),
						"model":    modelName,
						"provider": providerName,
					})
				} else {
					gsgin.GinResponseError(c, "模型返回了空响应", nil)
				}
				return
			case "error":
				errMsg := cast.ToString(rawEvt["error"])
				if evtErr, ok := rawEvt["error"].(map[string]interface{}); ok {
					errMsg = cast.ToString(evtErr["message"])
					if errMsg == "" {
						errMsg = cast.ToString(evtErr["type"])
					}
				}
				gsgin.GinResponseError(c, "调用失败: "+errMsg, nil)
				return
			}

		case <-timeout:
			if responseText != "" {
				gsgin.GinResponseSuccess(c, "ok", gin.H{
					"response": strings.TrimSpace(responseText),
					"model":    modelName,
					"provider": providerName,
				})
			} else {
				gsgin.GinResponseError(c, "测试超时（15秒内未收到完整响应）", nil)
			}
			return

		case <-ctx.Done():
			gsgin.GinResponseError(c, "测试已取消", nil)
			return
		}
	}
}

// piModelsJSON 顶层结构
type piModelsJSON struct {
	Providers map[string]piProviderConfig `json:"providers"`
}

type piProviderConfig struct {
	BaseURL string          `json:"baseUrl"`
	APIKey  string          `json:"apiKey"`
	API     string          `json:"api"`
	Compat  map[string]any  `json:"compat,omitempty"`
	Models  []piModelConfig `json:"models,omitempty"`
}

type piModelConfig struct {
	ID            string         `json:"id"`
	Name          string         `json:"name,omitempty"`
	ContextWindow int            `json:"contextWindow,omitempty"`
	Reasoning     bool           `json:"reasoning,omitempty"`
	Compat        map[string]any `json:"compat,omitempty"`
}

// piApiType 将 provider_type 映射为 Pi 的 api 字段值
// provider_type 同时充当"请求格式"的角色：
//   openai          → openai-completions (走 /v1/chat/completions)
//   openai-responses → openai-responses   (走 /v1/responses)
//   anthropic       → anthropic-messages  (走 /v1/messages)
//   deepseek        → openai-completions  (走 /v1/chat/completions，但需 compat.thinkingFormat=deepseek)
//   google          → google-generative-ai
func piApiType(providerType string) string {
	switch strings.ToLower(providerType) {
	case "anthropic":
		return "anthropic-messages"
	case "openai-responses":
		return "openai-responses"
	case "google":
		return "google-generative-ai"
	default:
		// openai, deepseek, 以及所有其他 OpenAI 兼容的 provider
		return "openai-completions"
	}
}

// applyProviderCompat 为特定 provider_type 自动生成 compat 配置
func applyProviderCompat(providerType string) map[string]any {
	switch strings.ToLower(providerType) {
	case "deepseek":
		return map[string]any{
			"thinkingFormat":          "deepseek",
			"supportsReasoningEffort": true,
		}
	default:
		return nil
	}
}

// applyModelCompat 为特定 provider_type + model 自动生成模型级 compat 和 reasoning 配置
func applyModelCompat(providerType, modelName string, mc piModelConfig) piModelConfig {
	switch strings.ToLower(providerType) {
	case "deepseek":
		// deepseek-reasoner 是推理模型
		if strings.Contains(strings.ToLower(modelName), "reasoner") || strings.Contains(strings.ToLower(modelName), "r1") {
			mc.Reasoning = true
			mc.Compat = map[string]any{
				"thinkingFormat":          "deepseek",
				"supportsReasoningEffort": true,
			}
		}
	}
	return mc
}

// syncPiModelsConfig 同步所有 Provider/Model 配置到 ~/.pi/agent/models.json（永久化）
// 从数据库读取所有启用的 providers 和 models，生成完整的 models.json
// 使用 provider name 作为 Pi provider key（唯一标识），provider_type 仅决定 api 字段
func syncPiModelsConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(homeDir, ".pi", "agent")
	modelsPath := filepath.Join(configDir, "models.json")

	// 查询所有启用的 Provider
	providers, err := common.DbMain.Client.QueryBySql(
		`SELECT id, name, provider_type, base_url, api_key FROM tbl_ai_provider WHERE status = 1`,
	).All()
	if err != nil {
		return err
	}

	cfg := piModelsJSON{Providers: make(map[string]piProviderConfig)}

	for _, p := range providers {
		providerId := cast.ToInt(p["id"])
		providerName := cast.ToString(p["name"])
		providerType := cast.ToString(p["provider_type"])
		baseURL := cast.ToString(p["base_url"])
		apiKey := cast.ToString(p["api_key"])

		// 查询该 provider 下所有启用的 LLM 模型
		modelRows, _ := common.DbMain.Client.QueryBySql(
			`SELECT model, name, uri, context_size FROM tbl_ai_model WHERE provider_id = ? AND model_type = 'llm' AND status = 1`,
			providerId,
		).All()

		// 获取第一个模型的 uri 用于计算 baseUrl 路径前缀
		firstUri := ""
		if len(modelRows) > 0 {
			firstUri = cast.ToString(modelRows[0]["uri"])
		}

		piBaseUrl := computePiBaseUrl(baseURL, firstUri, providerType)
		api := piApiType(providerType)
		providerCompat := applyProviderCompat(providerType)

		models := make([]piModelConfig, 0, len(modelRows))
		for _, m := range modelRows {
			mn := cast.ToString(m["model"])
			mc := piModelConfig{ID: mn}
			if name := cast.ToString(m["name"]); name != "" {
				mc.Name = name
			}
			if cs := cast.ToInt(m["context_size"]); cs > 0 {
				mc.ContextWindow = cs
			}
			mc = applyModelCompat(providerType, mn, mc)
			models = append(models, mc)
		}

		// provider key = name（唯一标识），避免同格式 provider 冲突和与 Pi 内置 provider 撞名
		cfg.Providers[providerName] = piProviderConfig{
			BaseURL: piBaseUrl,
			APIKey:  apiKey,
			API:     api,
			Compat:  providerCompat,
			Models:  models,
		}
	}

	newData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(modelsPath, newData, 0644); err != nil {
		return err
	}
	log.Printf("[pi-config] synced models.json: %d providers", len(cfg.Providers))
	return nil
}

// computePiBaseUrl 计算 Pi 的 baseUrl
// 数据库中 base_url 只有 scheme+host（不含 /v1），uri 包含完整路径（如 /v1/chat/completions）
// Pi 的 baseUrl 需要包含路径前缀（如 /v1），但不含 endpoint 部分（如 /chat/completions）
// Pi 根据 api 类型自动拼接 endpoint
//
// 示例：
//   baseURL=https://api.openai.com, uri=/v1/chat/completions → https://api.openai.com/v1
//   baseURL=https://api.anthropic.com, uri=/v1/messages → https://api.anthropic.com/v1
//   baseURL=https://api.deepseek.com, uri=/v1/chat/completions → https://api.deepseek.com/v1
func computePiBaseUrl(baseURL, uri, providerType string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if uri == "" {
		// uri 为空时，根据 provider_type 自动添加默认路径前缀
		switch strings.ToLower(providerType) {
		case "google":
			return baseURL + "/v1beta"
		default:
			return baseURL + "/v1"
		}
	}
	// 从 uri 中去掉 endpoint 部分，只保留路径前缀
	// Pi 的 baseUrl 需要包含路径前缀（如 /v1），Pi 根据 api 类型自动拼接 endpoint
	//
	// endpoint 是 Pi 自动拼接的最后一段路径：
	//   openai-completions → /chat/completions
	//   openai-responses   → /responses
	//   anthropic-messages → /messages
	//   google-generative-ai → 不需要拼接（baseUrl 直接到模型级别）
	//
	// 示例：uri=/v1/chat/completions → 去掉 /chat/completions → /v1 → baseUrl=https://api.openai.com/v1
	apiType := piApiType(providerType)
	var endpointSuffixes []string
	switch apiType {
	case "openai-completions":
		endpointSuffixes = []string{"/chat/completions"}
	case "openai-responses":
		endpointSuffixes = []string{"/responses"}
	case "anthropic-messages":
		endpointSuffixes = []string{"/messages"}
	case "google-generative-ai":
		// Google 的 baseUrl 不需要去掉 endpoint，uri 可能是 /v1beta/models 或 /v1beta/models:model
		// 对于测试请求，baseUrl 应包含到 /v1beta
		endpointSuffixes = []string{"/models", "/models:"}
	}
	for _, ep := range endpointSuffixes {
		if strings.HasSuffix(uri, ep) {
			return baseURL + strings.TrimRight(uri[:len(uri)-len(ep)], "/")
		}
	}
	// fallback：去掉最后一个路径段（保守策略）
	parts := strings.Split(strings.Trim(uri, "/"), "/")
	if len(parts) >= 2 {
		pathPrefix := "/" + strings.Join(parts[:len(parts)-1], "/")
		return baseURL + pathPrefix
	}
	// 最差 fallback：只加 /v1
	return baseURL + "/v1"
}
