package common

import (
	"bufio"
	"bytes"
	"dev_tool/internal/app/dtool/crawl4ai"
	"dev_tool/internal/pkg/p_common"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// InfoCrawlChatByModel 使用模型发起一次 AI 请求。
func (h *CSqlite) InfoCrawlChatByModel(modelID int, systemPrompt, userPrompt string) (string, map[string]any, error) {
	modelInfo, requestURL, apiKey, err := h.infoCrawlBuildChatRequest(modelID)
	if err != nil {
		return ``, nil, err
	}
	bodyMap := map[string]any{
		`model`: cast.ToString(modelInfo[`model`]),
		`messages`: []map[string]string{
			{
				`role`:    `system`,
				`content`: systemPrompt,
			},
			{
				`role`:    `user`,
				`content`: userPrompt,
			},
		},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return ``, nil, err
	}
	request.Header.Set(`Authorization`, `Bearer `+apiKey)
	request.Header.Set(`Content-Type`, `application/json`)
	client := &http.Client{Timeout: 120 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return ``, nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return ``, nil, err
	}
	if response.StatusCode >= 300 {
		return ``, nil, errors.New(`AI 请求失败: ` + string(responseBody))
	}
	content := p_common.ExtractOpenAiMessage(string(responseBody))
	if strings.TrimSpace(content) == `` {
		content = string(responseBody)
	}
	return strings.TrimSpace(content), modelInfo, nil
}

// InfoCrawlChatStreamByModel 使用模型发起流式 AI 请求。
func (h *CSqlite) InfoCrawlChatStreamByModel(modelID int, systemPrompt, userPrompt string, onChunk func(string)) (string, map[string]any, error) {
	modelInfo, requestURL, apiKey, err := h.infoCrawlBuildChatRequest(modelID)
	if err != nil {
		return ``, nil, err
	}
	bodyMap := map[string]any{
		`model`:  cast.ToString(modelInfo[`model`]),
		`stream`: true,
		`messages`: []map[string]string{
			{
				`role`:    `system`,
				`content`: systemPrompt,
			},
			{
				`role`:    `user`,
				`content`: userPrompt,
			},
		},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return ``, nil, err
	}
	request.Header.Set(`Authorization`, `Bearer `+apiKey)
	request.Header.Set(`Content-Type`, `application/json`)
	client := &http.Client{Timeout: 10 * time.Minute}
	response, err := client.Do(request)
	if err != nil {
		return ``, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(response.Body)
		return ``, nil, errors.New(`AI 请求失败: ` + string(responseBody))
	}
	reader := bufio.NewReader(response.Body)
	contentBuilder := strings.Builder{}
	for {
		line, readErr := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `data:`) {
			payload := strings.TrimSpace(strings.TrimPrefix(line, `data:`))
			if payload == `[DONE]` {
				break
			}
			chunk := h.infoCrawlExtractStreamContent(payload)
			if chunk != `` {
				contentBuilder.WriteString(chunk)
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
	return strings.TrimSpace(contentBuilder.String()), modelInfo, nil
}

// InfoCrawlSystemPrompt 返回统一系统提示词。
func (h *CSqlite) InfoCrawlSystemPrompt() string {
	return "你是一个信息抓取与整理助手。\n" +
		"请严格根据用户提示词和已抓取的网页材料完成整理。\n" +
		"不要假装自己联网，不要编造未在材料中出现的事实。\n" +
		"输出使用中文，内容尽量结构化，明确标注来源网址。"
}

// InfoCrawlBuildUserPrompt 构建信息采集 AI 用户提示词。
func (h *CSqlite) InfoCrawlBuildUserPrompt(taskInfo map[string]any, crawlResultList []crawl4ai.CrawlResult) string {
	builder := strings.Builder{}
	builder.WriteString("任务名称：")
	builder.WriteString(cast.ToString(taskInfo[`name`]))
	builder.WriteString("\n\n任务提示词：\n")
	builder.WriteString(cast.ToString(taskInfo[`prompt`]))
	builder.WriteString("\n\n已抓取网页材料：\n")
	for index, item := range crawlResultList {
		builder.WriteString("材料")
		builder.WriteString(cast.ToString(index + 1))
		builder.WriteString("：\nURL: ")
		builder.WriteString(item.URL)
		builder.WriteString("\n状态: ")
		if item.Success {
			builder.WriteString("成功")
		} else {
			builder.WriteString("失败")
		}
		if item.Title != `` {
			builder.WriteString("\n标题: ")
			builder.WriteString(item.Title)
		}
		if item.Error != `` {
			builder.WriteString("\n错误: ")
			builder.WriteString(item.Error)
		}
		if item.Markdown != `` {
			builder.WriteString("\n内容:\n")
			builder.WriteString(item.Markdown)
		}
		builder.WriteString("\n\n")
	}
	builder.WriteString("请基于以上材料完成信息整理，并明确引用对应网址。")
	return builder.String()
}

// infoCrawlBuildChatRequest 构建 AI 请求基础信息。
func (h *CSqlite) infoCrawlBuildChatRequest(modelID int) (map[string]any, string, string, error) {
	modelInfo, err := h.InfoCrawlAiModelInfo(modelID)
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

// infoCrawlExtractStreamContent 提取流式响应文本。
func (h *CSqlite) infoCrawlExtractStreamContent(payload string) string {
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
