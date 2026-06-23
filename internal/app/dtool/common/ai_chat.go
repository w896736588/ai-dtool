package common

import (
	"bufio"
	"bytes"
	"dev_tool/internal/pkg/p_common"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gshttp"
)

const (
	// aiChatRequestTimeout 统一限制 AI 普通与流式请求的最长等待时间为 5 分钟。 // aiChatRequestTimeout caps both standard and streaming AI requests at 5 minutes.
	aiChatRequestTimeout        = 5 * time.Minute
	aiChatRequestTimeoutSeconds = 300 // 5 分钟对应 gshttp.Request() 的秒数参数
)

// AiChatUsage 记录单次 AI 请求的 token 使用量。
type AiChatUsage struct {
	InputTokens          int
	OutputTokens         int
	CacheReadInputTokens int // 输入缓存命中 token 数
}

// AIChatByModel 使用模型发起一次 AI 请求。
func (h *CSqlite) AIChatByModel(modelID int, systemPrompt, userPrompt string) (string, map[string]any, error) {
	modelInfo, requestURL, apiKey, err := h.aiChatBuildRequest(modelID)
	if err != nil {
		return ``, nil, err
	}
	bodyMap := map[string]any{
		`model`: cast.ToString(modelInfo[`model`]),
		`messages`: []map[string]string{
			{`role`: `system`, `content`: systemPrompt},
			{`role`: `user`, `content`: userPrompt},
		},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	headers := map[string]string{
		`Authorization`: `Bearer ` + apiKey,
		`Content-Type`:  `application/json`,
	}
	startTime := time.Now()
	responseBytes, err := gshttp.PostJson(requestURL).
		BodyStr(string(bodyBytes)).
		Headers(headers).
		Request(aiChatRequestTimeoutSeconds).
		Result()
	costTimeMs := time.Since(startTime).Milliseconds()
	if err != nil {
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, 0, ``, err.Error(), costTimeMs)
		return ``, nil, err
	}
	responseBody := string(responseBytes)
	content := p_common.ExtractOpenAiMessage(responseBody)
	if strings.TrimSpace(content) == `` {
		content = responseBody
	}
	// 解析 token 使用量
	inputTokens, outputTokens, _ := h.extractTokenUsage(responseBody)
	h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, 200, responseBody, ``, costTimeMs, inputTokens, outputTokens)
	return strings.TrimSpace(content), modelInfo, nil
}

// AIChatByModelWithUsage 使用模型发起一次 AI 请求，同时返回 token 用量和耗时。
func (h *CSqlite) AIChatByModelWithUsage(modelID int, systemPrompt, userPrompt string) (string, map[string]any, *AiChatUsage, int64, error) {
	modelInfo, requestURL, apiKey, err := h.aiChatBuildRequest(modelID)
	if err != nil {
		return ``, nil, nil, 0, err
	}
	bodyMap := map[string]any{
		`model`: cast.ToString(modelInfo[`model`]),
		`messages`: []map[string]string{
			{`role`: `system`, `content`: systemPrompt},
			{`role`: `user`, `content`: userPrompt},
		},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	headers := map[string]string{
		`Authorization`: `Bearer ` + apiKey,
		`Content-Type`:  `application/json`,
	}
	startTime := time.Now()
	responseBytes, err := gshttp.PostJson(requestURL).
		BodyStr(string(bodyBytes)).
		Headers(headers).
		Request(aiChatRequestTimeoutSeconds).
		Result()
	costTimeMs := time.Since(startTime).Milliseconds()
	if err != nil {
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, 0, ``, err.Error(), costTimeMs)
		return ``, nil, nil, costTimeMs, err
	}
	responseBody := string(responseBytes)
	content := p_common.ExtractOpenAiMessage(responseBody)
	if strings.TrimSpace(content) == `` {
		content = responseBody
	}
	inputTokens, outputTokens, cacheReadInputTokens := h.extractTokenUsage(responseBody)
	h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, 200, responseBody, ``, costTimeMs, inputTokens, outputTokens)
	usage := &AiChatUsage{
		InputTokens:          inputTokens,
		OutputTokens:         outputTokens,
		CacheReadInputTokens: cacheReadInputTokens,
	}
	return strings.TrimSpace(content), modelInfo, usage, costTimeMs, nil
}

// AIChatStreamByModel 使用模型发起流式 AI 请求。
func (h *CSqlite) AIChatStreamByModel(modelID int, systemPrompt, userPrompt string, onChunk func(string)) (string, map[string]any, error) {
	modelInfo, requestURL, apiKey, err := h.aiChatBuildRequest(modelID)
	if err != nil {
		return ``, nil, err
	}
	bodyMap := map[string]any{
		`model`:  cast.ToString(modelInfo[`model`]),
		`stream`: true,
		`messages`: []map[string]string{
			{`role`: `system`, `content`: systemPrompt},
			{`role`: `user`, `content`: userPrompt},
		},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return ``, nil, err
	}
	request.Header.Set(`Authorization`, `Bearer `+apiKey)
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
			chunk := h.aiChatExtractStreamContent(payload)
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

// AIChatStreamByModelWithMessages 使用模型发起流式 AI 请求，支持多轮对话（传入完整 messages 列表）。
func (h *CSqlite) AIChatStreamByModelWithMessages(modelID int, messages []map[string]string, onChunk func(string)) (string, map[string]any, error) {
	modelInfo, requestURL, apiKey, err := h.aiChatBuildRequest(modelID)
	if err != nil {
		return ``, nil, err
	}
	bodyMap := map[string]any{
		`model`:    cast.ToString(modelInfo[`model`]),
		`stream`:   true,
		`messages`: messages,
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return ``, nil, err
	}
	request.Header.Set(`Authorization`, `Bearer `+apiKey)
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
			chunk := h.aiChatExtractStreamContent(payload)
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
	h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, response.StatusCode, responseBodyBuilder.String(), ``, costTimeMs, 0, 0)
	return strings.TrimSpace(contentBuilder.String()), modelInfo, nil
}

// AIChatByModelWithTools 使用模型发起 AI 请求，支持 Function Calling。
// messages 使用 []map[string]any 以支持 tool 角色（需 tool_call_id 字段）。
// tools 为 OpenAI 格式的工具定义列表，为空时不传 tools 字段。
// 返回 AI 回复内容、tool_calls 原始列表、token 用量、模型信息、错误。
func (h *CSqlite) AIChatByModelWithTools(modelID int, messages []map[string]any, tools []map[string]any) (string, []any, *AiChatUsage, map[string]any, error) {
	modelInfo, requestURL, apiKey, err := h.aiChatBuildRequest(modelID)
	if err != nil {
		return ``, nil, nil, nil, err
	}
	bodyMap := map[string]any{
		`model`:    cast.ToString(modelInfo[`model`]),
		`messages`: messages,
	}
	if len(tools) > 0 {
		bodyMap[`tools`] = tools
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	headers := map[string]string{
		`Authorization`: `Bearer ` + apiKey,
		`Content-Type`:  `application/json`,
	}
	startTime := time.Now()
	responseBytes, err := gshttp.PostJson(requestURL).
		BodyStr(string(bodyBytes)).
		Headers(headers).
		Request(aiChatRequestTimeoutSeconds).
		Result()
	costTimeMs := time.Since(startTime).Milliseconds()
	if err != nil {
		h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, 0, ``, err.Error(), costTimeMs)
		return ``, nil, nil, nil, err
	}
	responseBody := string(responseBytes)
	content, toolCalls := h.extractContentAndToolCalls(responseBody)
	inputTokens, outputTokens, cacheReadInputTokens := h.extractTokenUsage(responseBody)
	h.logAIRequest(modelInfo, requestURL, http.MethodPost, bodyMap, nil, 200, responseBody, ``, costTimeMs, inputTokens, outputTokens)
	usage := &AiChatUsage{
		InputTokens:          inputTokens,
		OutputTokens:         outputTokens,
		CacheReadInputTokens: cacheReadInputTokens,
	}
	return content, toolCalls, usage, modelInfo, nil
}

// extractContentAndToolCalls 从 AI 响应中提取文本内容和 tool_calls 列表。
func (h *CSqlite) extractContentAndToolCalls(responseBody string) (string, []any) {
	dataMap := make(map[string]any)
	if err := json.Unmarshal([]byte(responseBody), &dataMap); err != nil {
		return ``, nil
	}
	choiceList, ok := dataMap[`choices`].([]any)
	if !ok || len(choiceList) == 0 {
		return ``, nil
	}
	choiceMap, ok := choiceList[0].(map[string]any)
	if !ok {
		return ``, nil
	}
	messageMap, ok := choiceMap[`message`].(map[string]any)
	if !ok {
		return ``, nil
	}
	content := cast.ToString(messageMap[`content`])
	toolCalls, _ := messageMap[`tool_calls`].([]any)
	return content, toolCalls
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

func (h *CSqlite) aiChatBuildRequest(modelID int) (map[string]any, string, string, error) {
	modelInfo, err := h.AiModelInfo(modelID)
	if err != nil {
		return nil, ``, ``, err
	}
	if strings.ToLower(cast.ToString(modelInfo[`provider_type`])) != `openai` {
		return nil, ``, ``, errors.New(`当前仅支持 openai 兼容服务商`)
	}
	baseURL := strings.TrimSpace(cast.ToString(modelInfo[`base_url`]))
	if baseURL == `` {
		return nil, ``, ``, errors.New(`AI 服务商 base_url 不能为空`)
	}
	requestURI := strings.TrimSpace(cast.ToString(modelInfo[`uri`]))
	if requestURI == `` {
		requestURI = `/v1/chat/completions`
	}
	apiKey := strings.TrimSpace(cast.ToString(modelInfo[`api_key`]))
	if apiKey == `` {
		return nil, ``, ``, errors.New(`AI 服务商 api_key 不能为空`)
	}
	return modelInfo, joinAIRequestURL(baseURL, requestURI), apiKey, nil
}

func (h *CSqlite) aiChatExtractStreamContent(payload string) string {
	if strings.TrimSpace(payload) == `` {
		return ``
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
		if DbLog != nil && DbLog.Client != nil {
			_, _ = DbLog.Client.QuickCreate(`tbl_ai_request_log`, logData).Exec()
		}
	}()
}

// extractTokenUsage 从 OpenAI 响应中提取 token 使用量，包含缓存命中 token。
func (h *CSqlite) extractTokenUsage(responseBody string) (inputTokens, outputTokens, cacheReadInputTokens int) {
	if strings.TrimSpace(responseBody) == `` {
		return 0, 0, 0
	}
	dataMap := make(map[string]any)
	if err := json.Unmarshal([]byte(responseBody), &dataMap); err != nil {
		return 0, 0, 0
	}
	usage, ok := dataMap[`usage`].(map[string]any)
	if !ok {
		return 0, 0, 0
	}
	inputTokens = cast.ToInt(usage[`prompt_tokens`])
	outputTokens = cast.ToInt(usage[`completion_tokens`])
	// 如果 prompt_tokens 为 0，尝试使用 total_tokens
	if inputTokens == 0 {
		inputTokens = cast.ToInt(usage[`total_tokens`]) - outputTokens
	}
	// 提取缓存命中 token：支持 Anthropic 风格 (cache_read_input_tokens) 和 OpenAI 风格 (prompt_tokens_details.cached_tokens)
	cacheReadInputTokens = cast.ToInt(usage[`cache_read_input_tokens`])
	if cacheReadInputTokens == 0 {
		if details, ok := usage[`prompt_tokens_details`].(map[string]any); ok {
			cacheReadInputTokens = cast.ToInt(details[`cached_tokens`])
		}
	}
	return inputTokens, outputTokens, cacheReadInputTokens
}
