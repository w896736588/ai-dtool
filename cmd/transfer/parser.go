package main

import (
	"encoding/json"
	"strings"
)

// usageOpenAI OpenAI 格式的 usage 结构
type usageOpenAI struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// respOpenAI OpenAI 非流式响应结构（仅解析需要的字段）
type respOpenAI struct {
	Model string      `json:"model"`
	Usage usageOpenAI `json:"usage"`
}

// chunkOpenAI OpenAI 流式 chunk 结构
type chunkOpenAI struct {
	Model string      `json:"model"`
	Usage usageOpenAI `json:"usage"`
}

// usageAnthropic Anthropic 格式的 usage 结构
type usageAnthropic struct {
	InputTokens            int `json:"input_tokens"`
	OutputTokens           int `json:"output_tokens"`
	CacheCreateInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens   int `json:"cache_read_input_tokens"`
}

// respAnthropic Anthropic 非流式响应结构
type respAnthropic struct {
	Model string         `json:"model"`
	Usage usageAnthropic `json:"usage"`
}

// sseEventAnthropic Anthropic SSE 事件中的 usage（message_start / message_delta）
// message_start 的 usage 嵌套在 message 字段下，message_delta 则在顶层
type sseEventAnthropic struct {
	Type    string         `json:"type"`
	Usage   usageAnthropic `json:"usage"`
	Model   string         `json:"model"`
	Message *struct {
		Usage usageAnthropic `json:"usage"`
	} `json:"message,omitempty"`
}

// API 格式常量
const (
	formatOpenAI    = "openai"
	formatAnthropic = "anthropic"
)

// extractModel 从请求 body 中提取 model 字段
func extractModel(body string) string {
	var req struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return ""
	}
	return req.Model
}

// parseTokens 解析响应中的 token 信息
// format: "openai" 或 "anthropic"
// responseBody: 非流式完整响应 body 或流式累积的 chunks
func parseTokens(format, responseBody string) (inputTokens, outputTokens, cacheReadTokens int) {
	if format == formatOpenAI {
		inputTokens, outputTokens = parseOpenAITokens(responseBody)
		return inputTokens, outputTokens, 0
	}
	return parseAnthropicTokens(responseBody)
}

// parseOpenAITokens 解析 OpenAI 格式的 token 信息
func parseOpenAITokens(responseBody string) (inputTokens, outputTokens int) {
	// 优先尝试解析为完整响应（非流式）
	var fullResp respOpenAI
	if err := json.Unmarshal([]byte(responseBody), &fullResp); err == nil && fullResp.Usage.TotalTokens > 0 {
		return fullResp.Usage.PromptTokens, fullResp.Usage.CompletionTokens
	}

	// 流式响应：从最后一条含 usage 的 JSON 行提取
	return parseOpenAIStreamChunks(responseBody)
}

// parseOpenAIStreamChunks 从流式 SSE data 中提取 OpenAI token 信息
func parseOpenAIStreamChunks(chunksData string) (inputTokens, outputTokens int) {
	lines := strings.Split(chunksData, "\n")
	var lastUsage usageOpenAI

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || line == "[DONE]" {
			continue
		}
		// 去掉 "data: " 前缀
		data := strings.TrimPrefix(line, "data: ")
		if data == line {
			data = strings.TrimPrefix(line, "data:")
		}

		var chunk chunkOpenAI
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Usage.TotalTokens > 0 {
			lastUsage = chunk.Usage
			break
		}
	}

	return lastUsage.PromptTokens, lastUsage.CompletionTokens
}

// parseAnthropicTokens 解析 Anthropic 格式的 token 信息
func parseAnthropicTokens(responseBody string) (inputTokens, outputTokens, cacheReadTokens int) {
	// 优先尝试解析为完整响应（非流式）
	var fullResp respAnthropic
	if err := json.Unmarshal([]byte(responseBody), &fullResp); err == nil {
		return fullResp.Usage.InputTokens, fullResp.Usage.OutputTokens, fullResp.Usage.CacheReadInputTokens
	}

	// 流式响应：从 message_start 和 message_delta 事件提取
	return parseAnthropicStreamChunks(responseBody)
}

// parseAnthropicStreamChunks 从流式 SSE data 中提取 Anthropic token 信息
func parseAnthropicStreamChunks(chunksData string) (inputTokens, outputTokens, cacheReadTokens int) {
	lines := strings.Split(chunksData, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 去掉 SSE event/data 前缀
		data := strings.TrimPrefix(line, "data: ")
		if data == line {
			data = strings.TrimPrefix(line, "data:")
		}
		if data == "" {
			continue
		}

		var event sseEventAnthropic
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "message_start":
			// message_start 的 usage 嵌套在 message 字段下
			if event.Message != nil && event.Message.Usage.InputTokens > 0 {
				inputTokens = event.Message.Usage.InputTokens
				cacheReadTokens = event.Message.Usage.CacheReadInputTokens
			}
		case "message_delta":
			// message_delta 的 usage 在顶层，包含 input_tokens + output_tokens + cache 信息
			if event.Usage.InputTokens > 0 {
				inputTokens = event.Usage.InputTokens
			}
			if event.Usage.OutputTokens > 0 {
				outputTokens = event.Usage.OutputTokens
			}
			if event.Usage.CacheReadInputTokens > 0 {
				cacheReadTokens = event.Usage.CacheReadInputTokens
			}
		}
	}

	return inputTokens, outputTokens, cacheReadTokens
}
