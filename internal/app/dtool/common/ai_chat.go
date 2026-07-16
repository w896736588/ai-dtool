package common

import (
	"bufio"
	"bytes"
	"dev_tool/internal/pkg/p_common"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cast"
)

const (
	// aiChatRequestTimeout 统一限制 AI 普通与流式请求的最长等待时间为 5 分钟。 // aiChatRequestTimeout caps both standard and streaming AI requests at 5 minutes.
	aiChatRequestTimeout = 5 * time.Minute
)

// AiChatUsage 记录单次 AI 请求的 token 使用量。
type AiChatUsage struct {
	InputTokens  int
	OutputTokens int
}

// AIChatByModel 使用模型发起一次 AI 请求。
func (h *CSqlite) AIChatByModel(modelID int, systemPrompt, userPrompt string) (string, map[string]any, error) {
	modelInfo, requestURL, apiKey, providerType, err := h.aiChatBuildRequest(modelID)
	if err != nil {
		return ``, nil, err
	}
	bodyMap := buildProviderChatBody(providerType, cast.ToString(modelInfo[`model`]), systemPrompt, userPrompt)
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return ``, nil, err
	}
	setProviderAuthHeaders(request, providerType, apiKey)
	request.Header.Set(`Content-Type`, `application/json`)
	client := &http.Client{Timeout: aiChatRequestTimeout}
	startTime := time.Now()
	response, err := client.Do(request)
	costTimeMs := time.Since(startTime).Milliseconds()
	if err != nil {
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, 0, ``, err.Error(), costTimeMs)
		return ``, nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, response.StatusCode, ``, err.Error(), costTimeMs)
		return ``, nil, err
	}
	if response.StatusCode >= 300 {
		errMsg := `AI 请求失败: ` + string(responseBody)
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, response.StatusCode, string(responseBody), errMsg, costTimeMs)
		return ``, nil, errors.New(errMsg)
	}
	content := extractProviderContent(providerType, string(responseBody))
	if strings.TrimSpace(content) == `` {
		content = string(responseBody)
	}
	// 解析 token 使用量
	inputTokens, outputTokens := h.extractTokenUsage(providerType, string(responseBody))
	h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, response.StatusCode, string(responseBody), ``, costTimeMs, inputTokens, outputTokens)
	return strings.TrimSpace(content), modelInfo, nil
}

// AIChatByModelWithUsage 使用模型发起一次 AI 请求，同时返回 token 用量和耗时。
func (h *CSqlite) AIChatByModelWithUsage(modelID int, systemPrompt, userPrompt string) (string, map[string]any, *AiChatUsage, int64, error) {
	modelInfo, requestURL, apiKey, providerType, err := h.aiChatBuildRequest(modelID)
	if err != nil {
		return ``, nil, nil, 0, err
	}
	bodyMap := buildProviderChatBody(providerType, cast.ToString(modelInfo[`model`]), systemPrompt, userPrompt)
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return ``, nil, nil, 0, err
	}
	setProviderAuthHeaders(request, providerType, apiKey)
	request.Header.Set(`Content-Type`, `application/json`)
	client := &http.Client{Timeout: aiChatRequestTimeout}
	startTime := time.Now()
	response, err := client.Do(request)
	costTimeMs := time.Since(startTime).Milliseconds()
	if err != nil {
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, 0, ``, err.Error(), costTimeMs)
		return ``, nil, nil, costTimeMs, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, response.StatusCode, ``, err.Error(), costTimeMs)
		return ``, nil, nil, costTimeMs, err
	}
	if response.StatusCode >= 300 {
		errMsg := `AI 请求失败: ` + string(responseBody)
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, response.StatusCode, string(responseBody), errMsg, costTimeMs)
		return ``, nil, nil, costTimeMs, errors.New(errMsg)
	}
	content := extractProviderContent(providerType, string(responseBody))
	if strings.TrimSpace(content) == `` {
		content = string(responseBody)
	}
	inputTokens, outputTokens := h.extractTokenUsage(providerType, string(responseBody))
	h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, response.StatusCode, string(responseBody), ``, costTimeMs, inputTokens, outputTokens)
	usage := &AiChatUsage{InputTokens: inputTokens, OutputTokens: outputTokens}
	return strings.TrimSpace(content), modelInfo, usage, costTimeMs, nil
}

// AIChatStreamByModel 使用模型发起流式 AI 请求。
func (h *CSqlite) AIChatStreamByModel(modelID int, systemPrompt, userPrompt string, onChunk func(string)) (string, map[string]any, error) {
	modelInfo, requestURL, apiKey, providerType, err := h.aiChatBuildRequest(modelID)
	if err != nil {
		return ``, nil, err
	}
	bodyMap := buildProviderChatBodyStream(providerType, cast.ToString(modelInfo[`model`]), systemPrompt, userPrompt)
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return ``, nil, err
	}
	setProviderAuthHeaders(request, providerType, apiKey)
	request.Header.Set(`Content-Type`, `application/json`)
	client := &http.Client{Timeout: aiChatRequestTimeout}
	startTime := time.Now()
	response, err := client.Do(request)
	costTimeMs := time.Since(startTime).Milliseconds()
	if err != nil {
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, 0, ``, err.Error(), costTimeMs)
		return ``, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(response.Body)
		errMsg := `AI 请求失败: ` + string(responseBody)
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, response.StatusCode, string(responseBody), errMsg, costTimeMs)
		return ``, nil, errors.New(errMsg)
	}
	reader := bufio.NewReader(response.Body)
	contentBuilder := strings.Builder{}
	responseBodyBuilder := strings.Builder{}
	for {
		line, readErr := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `data:`) {
			payload := strings.TrimSpace(strings.TrimPrefix(line, `data:`))
			if payload == `[DONE]` {
				break
			}
			chunk := h.aiChatExtractStreamContent(providerType, payload)
			if chunk != `` {
				contentBuilder.WriteString(chunk)
				responseBodyBuilder.WriteString(payload + "\n")
				if onChunk != nil {
					onChunk(chunk)
				}
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return strings.TrimSpace(contentBuilder.String()), modelInfo, readErr
		}
	}
	// 流式响应不单独计算 token，留待后续扩展
	h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, response.StatusCode, responseBodyBuilder.String(), ``, costTimeMs, 0, 0)
	return strings.TrimSpace(contentBuilder.String()), modelInfo, nil
}

// AiModelInfo 查询 AI 模型配置。
func (h *CSqlite) AiModelInfo(id int) (map[string]any, error) {
	info, err := h.Client.QueryBySql(`
select m.*,p.name as provider_name,p.provider_type,p.base_url,p.api_key
from tbl_ai_model m
left join tbl_ai_provider p on p.id = m.provider_id
where m.id = ? and m.status = 1 and p.status = 1`, id).One()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`AI模型不存在或已停用`)
	}
	return info, nil
}

// aiChatBuildRequest 根据模型 ID 构建 AI 请求所需信息，返回模型信息、请求 URL、API Key 和服务商类型。
func (h *CSqlite) aiChatBuildRequest(modelID int) (modelInfo map[string]any, requestURL string, apiKey string, providerType string, err error) {
	modelInfo, err = h.AiModelInfo(modelID)
	if err != nil {
		return nil, ``, ``, ``, err
	}
	providerType = strings.ToLower(cast.ToString(modelInfo[`provider_type`]))
	if providerType == `` {
		providerType = `openai`
	}
	// 仅支持已知的服务商类型
	if providerType != `openai` && providerType != `deepseek` && providerType != `anthropic` && providerType != `openai-responses` {
		return nil, ``, ``, ``, errors.New(`不支持的服务商类型: ` + providerType)
	}
	baseURL := strings.TrimSpace(cast.ToString(modelInfo[`base_url`]))
	if baseURL == `` {
		return nil, ``, ``, ``, errors.New(`AI 服务商 base_url 不能为空`)
	}
	requestURI := strings.TrimSpace(cast.ToString(modelInfo[`uri`]))
	if requestURI == `` {
		requestURI = defaultURIForProviderType(providerType)
	}
	apiKey = strings.TrimSpace(cast.ToString(modelInfo[`api_key`]))
	if apiKey == `` {
		return nil, ``, ``, ``, errors.New(`AI 服务商 api_key 不能为空`)
	}
	requestURL = joinAIRequestURL(baseURL, requestURI)
	return modelInfo, requestURL, apiKey, providerType, nil
}

// defaultURIForProviderType 返回服务商类型的默认 URI。
func defaultURIForProviderType(providerType string) string {
	switch providerType {
	case `anthropic`:
		return `/v1/messages`
	case `openai-responses`:
		return `/v1/responses`
	default:
		return `/v1/chat/completions`
	}
}

// buildProviderChatBody 根据服务商类型构建请求体（非流式）。
func buildProviderChatBody(providerType string, model string, systemPrompt, userPrompt string) map[string]any {
	switch providerType {
	case `anthropic`:
		body := map[string]any{
			`model`:      model,
			`max_tokens`: 4096,
			`messages`: []map[string]string{
				{`role`: `user`, `content`: userPrompt},
			},
		}
		if systemPrompt != `` {
			body[`system`] = systemPrompt
		}
		return body
	case `openai-responses`:
		body := map[string]any{
			`model`: model,
			`input`: userPrompt,
		}
		if systemPrompt != `` {
			body[`instructions`] = systemPrompt
		}
		return body
	default:
		return map[string]any{
			`model`: model,
			`messages`: []map[string]string{
				{`role`: `system`, `content`: systemPrompt},
				{`role`: `user`, `content`: userPrompt},
			},
		}
	}
}

// buildProviderChatBodyStream 根据服务商类型构建流式请求体。
func buildProviderChatBodyStream(providerType string, model string, systemPrompt, userPrompt string) map[string]any {
	switch providerType {
	case `anthropic`:
		body := map[string]any{
			`model`:      model,
			`max_tokens`: 4096,
			`stream`:     true,
			`messages`: []map[string]string{
				{`role`: `user`, `content`: userPrompt},
			},
		}
		if systemPrompt != `` {
			body[`system`] = systemPrompt
		}
		return body
	case `openai-responses`:
		body := map[string]any{
			`model`:  model,
			`input`:  userPrompt,
			`stream`: true,
		}
		if systemPrompt != `` {
			body[`instructions`] = systemPrompt
		}
		return body
	default:
		return map[string]any{
			`model`:  model,
			`stream`: true,
			`messages`: []map[string]string{
				{`role`: `system`, `content`: systemPrompt},
				{`role`: `user`, `content`: userPrompt},
			},
		}
	}
}

// setProviderAuthHeaders 根据服务商类型设置请求认证头。
func setProviderAuthHeaders(request *http.Request, providerType, apiKey string) {
	switch providerType {
	case `anthropic`:
		request.Header.Set(`x-api-key`, apiKey)
		request.Header.Set(`anthropic-version`, `2023-06-01`)
	default:
		// openai, deepseek, openai-responses 使用 Bearer Token
		request.Header.Set(`Authorization`, `Bearer `+apiKey)
	}
}

// extractProviderContent 根据服务商类型从响应中提取文本内容。
func extractProviderContent(providerType, responseBody string) string {
	switch providerType {
	case `anthropic`:
		return extractAnthropicMessage(responseBody)
	case `openai-responses`:
		return extractOpenAiResponsesMessage(responseBody)
	default:
		// openai, deepseek
		return p_common.ExtractOpenAiMessage(responseBody)
	}
}

// extractAnthropicMessage 从 Anthropic 响应中提取文本内容。
func extractAnthropicMessage(responseBody string) string {
	if strings.TrimSpace(responseBody) == `` {
		return ``
	}
	dataMap := make(map[string]any)
	if err := json.Unmarshal([]byte(responseBody), &dataMap); err != nil {
		return ``
	}
	contentList, ok := dataMap[`content`].([]any)
	if !ok || len(contentList) == 0 {
		return ``
	}
	first, ok := contentList[0].(map[string]any)
	if !ok {
		return ``
	}
	return strings.TrimSpace(cast.ToString(first[`text`]))
}

// extractOpenAiResponsesMessage 从 OpenAI Responses API 响应中提取文本内容。
func extractOpenAiResponsesMessage(responseBody string) string {
	if strings.TrimSpace(responseBody) == `` {
		return ``
	}
	dataMap := make(map[string]any)
	if err := json.Unmarshal([]byte(responseBody), &dataMap); err != nil {
		return ``
	}
	outputRaw, ok := dataMap[`output`].([]any)
	if !ok || len(outputRaw) == 0 {
		return ``
	}
	firstOutput, ok := outputRaw[0].(map[string]any)
	if !ok {
		return ``
	}
	contentRaw, ok := firstOutput[`content`].([]any)
	if !ok || len(contentRaw) == 0 {
		return ``
	}
	firstContent, ok := contentRaw[0].(map[string]any)
	if !ok {
		return ``
	}
	return strings.TrimSpace(cast.ToString(firstContent[`text`]))
}

func (h *CSqlite) aiChatExtractStreamContent(providerType, payload string) string {
	if strings.TrimSpace(payload) == `` {
		return ``
	}
	// Anthropic 流式事件格式：event: content_block_delta / data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"..."}}
	if providerType == `anthropic` {
		return extractAnthropicStreamChunk(payload)
	}
	// openai-responses 流式 SSE 事件，delta 包含在 response.output_text_delta 中
	if providerType == `openai-responses` {
		return extractOpenAiResponsesStreamChunk(payload)
	}
	dataMap := make(map[string]any)
	if err := json.Unmarshal([]byte(payload), &dataMap); err != nil {
		return ``
	}
	choiceList, ok := dataMap[`choices`].([]any)
	if !ok || len(choiceList) == 0 {
		return ``
	}
	choiceMap, ok := choiceList[0].(map[string]any)
	if !ok {
		return ``
	}
	if deltaMap, ok := choiceMap[`delta`].(map[string]any); ok {
		if chunk := cast.ToString(deltaMap[`content`]); chunk != `` {
			return chunk
		}
	}
	if messageMap, ok := choiceMap[`message`].(map[string]any); ok {
		if chunk := cast.ToString(messageMap[`content`]); chunk != `` {
			return chunk
		}
	}
	return ``
}

// extractAnthropicStreamChunk 从 Anthropic SSE 事件中提取流式文本块。
func extractAnthropicStreamChunk(payload string) string {
	dataMap := make(map[string]any)
	if err := json.Unmarshal([]byte(payload), &dataMap); err != nil {
		return ``
	}
	// 只处理 content_block_delta 类型的事件
	if cast.ToString(dataMap[`type`]) != `content_block_delta` {
		return ``
	}
	delta, ok := dataMap[`delta`].(map[string]any)
	if !ok {
		return ``
	}
	if cast.ToString(delta[`type`]) == `text_delta` {
		return cast.ToString(delta[`text`])
	}
	return ``
}

// extractOpenAiResponsesStreamChunk 从 OpenAI Responses API 流式 SSE 事件中提取文本块。
func extractOpenAiResponsesStreamChunk(payload string) string {
	dataMap := make(map[string]any)
	if err := json.Unmarshal([]byte(payload), &dataMap); err != nil {
		return ``
	}
	// Responses API 流式事件：type 为 "response.output_text.delta"，delta 在顶层
	if cast.ToString(dataMap[`type`]) == `response.output_text.delta` {
		delta, ok := dataMap[`delta`].(map[string]any)
		if ok {
			return cast.ToString(delta[`text`])
		}
		return cast.ToString(dataMap[`delta`])
	}
	return ``
}

func joinAIRequestURL(baseURL, requestURI string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), `/`)
	requestURI = strings.TrimSpace(requestURI)
	if requestURI == `` {
		return baseURL
	}
	if !strings.HasPrefix(requestURI, `/`) {
		requestURI = `/` + requestURI
	}
	return baseURL + requestURI
}

// logAIRequest 记录 AI 请求日志到日志库。
func (h *CSqlite) logAIRequest(
	modelInfo map[string]any,
	requestURL, method string,
	requestParams map[string]any,
	requestHeaders map[string]string,
	statusCode int,
	responseBody, errMsg string,
	costTimeMs int64,
	inputTokens ...int,
) {
	// 提取可选的 token 参数
	inputTk := 0
	outputTk := 0
	if len(inputTokens) >= 1 {
		inputTk = inputTokens[0]
	}
	if len(inputTokens) >= 2 {
		outputTk = inputTokens[1]
	}

	// 构建请求头（脱敏）
	headers := make(map[string]string)
	if requestHeaders != nil {
		for k, v := range requestHeaders {
			if strings.ToLower(k) == `authorization` {
				// 脱敏 API Key
				if len(v) > 10 {
					v = v[:6] + `******` + v[len(v)-4:]
				}
			}
			headers[k] = v
		}
	}

	success := 1
	if errMsg != `` {
		success = 0
	}

	providerID := cast.ToInt(modelInfo[`provider_id`])
	providerName := cast.ToString(modelInfo[`provider_name`])
	modelID := cast.ToInt(modelInfo[`id`])
	modelName := cast.ToString(modelInfo[`name`])
	model := cast.ToString(modelInfo[`model`])
	modelType := cast.ToString(modelInfo[`model_type`])
	if modelType == `` {
		modelType = `llm`
	}
	requestFormat := cast.ToString(modelInfo[`provider_type`])
	if requestFormat == `` {
		requestFormat = `openai`
	}
	baseURL := cast.ToString(modelInfo[`base_url`])

	requestParamsJSON, _ := json.Marshal(requestParams)
	headersJSON, _ := json.Marshal(headers)

	logData := map[string]any{
		`provider_id`:          providerID,
		`provider_name`:        providerName,
		`model_id`:             modelID,
		`model_name`:           modelName,
		`model`:                model,
		`model_type`:           modelType,
		`request_format`:       requestFormat,
		`base_url`:             baseURL,
		`request_url`:          requestURL,
		`request_method`:       method,
		`request_params`:       string(requestParamsJSON),
		`request_headers`:      string(headersJSON),
		`response_status_code`: statusCode,
		`response_body`:        responseBody,
		`input_tokens`:         inputTk,
		`output_tokens`:        outputTk,
		`cost_time_ms`:         costTimeMs,
		`success`:              success,
		`error_message`:        errMsg,
		`create_time`:          time.Now().Unix(),
	}

	// 异步写入日志，避免阻塞主流程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ai_chat] logAIRequest panic: %v", r)
			}
		}()
		if DbLog != nil && DbLog.Client != nil {
			_, _ = DbLog.Client.QuickCreate(`tbl_ai_request_log`, logData).Exec()
		}
	}()
}

// extractTokenUsage 从响应中提取 token 使用量，根据服务商类型采用不同解析方式。
func (h *CSqlite) extractTokenUsage(providerType, responseBody string) (inputTokens, outputTokens int) {
	if strings.TrimSpace(responseBody) == `` {
		return 0, 0
	}
	switch providerType {
	case `anthropic`:
		return extractAnthropicTokenUsage(responseBody)
	default:
		return extractOpenAiTokenUsage(responseBody)
	}
}

// extractOpenAiTokenUsage 从 OpenAI 兼容响应中提取 token 使用量。
func extractOpenAiTokenUsage(responseBody string) (inputTokens, outputTokens int) {
	dataMap := make(map[string]any)
	if err := json.Unmarshal([]byte(responseBody), &dataMap); err != nil {
		return 0, 0
	}
	usage, ok := dataMap[`usage`].(map[string]any)
	if !ok {
		return 0, 0
	}
	inputTokens = cast.ToInt(usage[`prompt_tokens`])
	outputTokens = cast.ToInt(usage[`completion_tokens`])
	if inputTokens == 0 {
		inputTokens = cast.ToInt(usage[`total_tokens`]) - outputTokens
	}
	return inputTokens, outputTokens
}

// extractAnthropicTokenUsage 从 Anthropic 响应中提取 token 使用量。
func extractAnthropicTokenUsage(responseBody string) (inputTokens, outputTokens int) {
	dataMap := make(map[string]any)
	if err := json.Unmarshal([]byte(responseBody), &dataMap); err != nil {
		return 0, 0
	}
	usage, ok := dataMap[`usage`].(map[string]any)
	if !ok {
		return 0, 0
	}
	inputTokens = cast.ToInt(usage[`input_tokens`])
	outputTokens = cast.ToInt(usage[`output_tokens`])
	return inputTokens, outputTokens
}
