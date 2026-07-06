package controller

import (
	"context"
	"encoding/json"
	"os"
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

	// 查询 Provider
	providerRow, err := common.DbMain.Client.QueryBySql(
		`SELECT provider_type, base_url, api_key FROM tbl_ai_provider WHERE id = ? AND status = 1`,
		req.ProviderId,
	).One()
	if err != nil || len(providerRow) == 0 {
		gsgin.GinResponseError(c, "Provider 不存在或已禁用", nil)
		return
	}

	// 查询 Model
	modelRow, err := common.DbMain.Client.QueryBySql(
		`SELECT model, uri FROM tbl_ai_model WHERE id = ? AND provider_id = ? AND status = 1`,
		req.ModelId, req.ProviderId,
	).One()
	if err != nil || len(modelRow) == 0 {
		gsgin.GinResponseError(c, "模型不存在或已禁用", nil)
		return
	}

	providerType := cast.ToString(providerRow["provider_type"])
	modelName := cast.ToString(modelRow["model"])
	baseURL := cast.ToString(providerRow["base_url"])
	apiKey := cast.ToString(providerRow["api_key"])
	uri := cast.ToString(modelRow["uri"])

	modelAddr := baseURL
	if uri != "" {
		modelAddr = strings.TrimRight(baseURL, "/") + uri
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

	if err := adapter.Start(ctx, agent.AgentStartConfig{
		SessionDir: tmpDir,
		Provider:   providerType,
		Model:      modelName,
		ModelAddr:  modelAddr,
		ApiKey:     apiKey,
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
				if responseText != "" {
					gsgin.GinResponseSuccess(c, "ok", gin.H{
						"response": strings.TrimSpace(responseText),
						"model":    modelName,
						"provider": providerType,
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
				if msgEvt, ok := rawEvt["assistantMessageEvent"].(map[string]interface{}); ok {
					if cast.ToString(msgEvt["type"]) == "text_delta" {
						responseText += cast.ToString(msgEvt["delta"])
					}
				}
			case "turn_end":
				if responseText != "" {
					gsgin.GinResponseSuccess(c, "ok", gin.H{
						"response": strings.TrimSpace(responseText),
						"model":    modelName,
						"provider": providerType,
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
					"provider": providerType,
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
