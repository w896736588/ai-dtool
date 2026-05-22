package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Proxy HTTP 代理转发器
type Proxy struct {
	target *url.URL
	store  *Store
	client *http.Client
	format string // API 格式：openai 或 anthropic，从规则配置中获取
	ruleID string // 所属规则的 ID
}

// NewProxy 创建代理实例
func NewProxy(targetURL string, store *Store, format string, ruleID string) (*Proxy, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("解析目标URL失败: %w", err)
	}
	return &Proxy{
		target: u,
		store:  store,
		client: &http.Client{
			Timeout: 0, // 流式请求不设超时
		},
		format: format,
		ruleID: ruleID,
	}, nil
}

// ServeHTTP 实现 http.Handler 接口
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 创建记录
	rec := &Record{
		ID:           uuid.New().String(),
		Time:         startTime,
		Method:       r.Method,
		Path:         r.URL.Path,
		Headers:      flattenHeaders(r.Header),
		QueryParams:  flattenValues(r.URL.Query()),
		RuleID:       p.ruleID,
		IsStream:     false,
		StreamChunks: make([]StreamChunk, 0),
	}

	// 读取请求 body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求body失败", http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	rec.Body = string(bodyBytes)

	// 判断是否流式请求
	rec.IsStream = isStreamRequest(bodyBytes, r.Header)
	// 提取 model
	rec.Model = extractModel(rec.Body)

	// 构建上游 URL
	upstreamURL := p.buildUpstreamURL(r.URL)

	// 创建上游请求
	upstreamReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "创建上游请求失败", http.StatusInternalServerError)
		return
	}
	copyHeadersExclude(upstreamReq.Header, r.Header)

	// 发送上游请求
	resp, err := p.client.Do(upstreamReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("上游请求失败: %s", err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	rec.StatusCode = resp.StatusCode

	// 复制响应头
	copyHeaders(w.Header(), resp.Header)

	if rec.IsStream {
		// 流式处理：先 Add 初始记录，再增量更新
		p.store.Add(rec)
		p.handleStream(w, resp, rec, startTime)
	} else {
		// 非流式处理
		p.handleRegular(w, resp, rec, startTime)
	}
}

// buildUpstreamURL 构建上游请求 URL
func (p *Proxy) buildUpstreamURL(reqURL *url.URL) string {
	// 拼接 target + 原始 path + query string
	result := *p.target
	result.Path = strings.TrimSuffix(result.Path, "/") + "/" + strings.TrimPrefix(reqURL.Path, "/")
	result.Path = strings.TrimRight(result.Path, "/")
	if result.Path == "" {
		result.Path = "/"
	}
	result.RawQuery = reqURL.RawQuery
	result.Fragment = reqURL.Fragment
	return result.String()
}

// handleRegular 处理非流式响应
func (p *Proxy) handleRegular(w http.ResponseWriter, resp *http.Response, rec *Record, startTime time.Time) {
	// 读取完整响应 body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		rec.ResponseBody = fmt.Sprintf("读取响应失败: %s", err.Error())
	} else {
		rec.ResponseBody = string(bodyBytes)
	}

	// 记录完成时间和耗时
	rec.CompleteTime = time.Now()
	rec.Duration = rec.CompleteTime.Sub(startTime).Milliseconds()
	rec.Completed = true

	// 使用规则配置的格式解析 token 信息
	rec.InputTokens, rec.OutputTokens, rec.CacheReadInputTokens = parseTokens(p.format, rec.ResponseBody)

	// 保存记录
	p.store.Add(rec)

	// 返回响应给客户端
	w.WriteHeader(resp.StatusCode)
	if bodyBytes != nil {
		w.Write(bodyBytes)
	}
}

// handleStream 处理流式（SSE）响应，增量更新记录
func (p *Proxy) handleStream(w http.ResponseWriter, resp *http.Response, rec *Record, startTime time.Time) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "不支持流式响应", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)

	var chunksBuf bytes.Buffer
	scanner := bufio.NewScanner(resp.Body)
	// 增大 scanner buffer 以处理较长的 SSE 行
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	chunkCount := 0
	// 每 N 个 chunk 推送一次 UI 更新，避免 SSE 数据太频繁
	const updateInterval = 5

	for scanner.Scan() {
		line := scanner.Text()
		// 写入客户端
		fmt.Fprintf(w, "%s\n", line)
		flusher.Flush()

		// 记录 chunk
		rec.StreamChunks = append(rec.StreamChunks, StreamChunk{
			Timestamp: time.Now(),
			Data:      line,
		})
		// 累积到 buffer 用于后续 token 解析
		chunksBuf.WriteString(line)
		chunksBuf.WriteString("\n")

		chunkCount++
		if chunkCount%updateInterval == 0 {
			p.store.UpdateRecord(rec)
		}
	}

	if err := scanner.Err(); err != nil {
		rec.ResponseBody = fmt.Sprintf("流读取错误: %s", err.Error())
	}

	chunksStr := chunksBuf.String()
	rec.CompleteTime = time.Now()
	rec.Duration = rec.CompleteTime.Sub(startTime).Milliseconds()

	// 从 chunksBuf 解析 token（不保留 ResponseBody，前端从 StreamChunks 拼接）
	rec.InputTokens, rec.OutputTokens, rec.CacheReadInputTokens = parseTokens(p.format, chunksStr)

	// 标记完成，最终更新
	rec.Completed = true
	p.store.UpdateRecord(rec)
}

// isStreamRequest 判断是否为流式请求
func isStreamRequest(body []byte, headers http.Header) bool {
	// 1. 检查请求 body 中是否包含 "stream": true
	if len(body) > 0 {
		var req map[string]interface{}
		if err := json.Unmarshal(body, &req); err == nil {
			if stream, ok := req["stream"]; ok {
				if streamBool, ok := stream.(bool); ok && streamBool {
					return true
				}
			}
		}
	}

	// 2. 检查 Accept header 是否为 SSE
	accept := headers.Get("Accept")
	if strings.Contains(accept, "text/event-stream") {
		return true
	}

	return false
}

// flattenHeaders 将 http.Header 扁平化为 map[string]string
func flattenHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		result[k] = strings.Join(v, ", ")
	}
	return result
}

// flattenValues 将 url.Values 扁平化为 map[string]string
func flattenValues(values url.Values) map[string]string {
	result := make(map[string]string)
	for k, v := range values {
		result[k] = strings.Join(v, ", ")
	}
	return result
}

// copyHeaders 复制 HTTP headers
func copyHeaders(dst, src http.Header) {
	for k, v := range src {
		for _, vv := range v {
			dst.Add(k, vv)
		}
	}
}

// headersToExclude 代理转发时需要排除的 headers
var headersToExclude = map[string]bool{
	"Host":              true,
	"Transfer-Encoding": true,
	"Te":                true,
	"Trailer":           true,
}

// copyHeadersExclude 复制 headers，排除代理相关 headers
func copyHeadersExclude(dst, src http.Header) {
	for k, v := range src {
		if headersToExclude[k] {
			continue
		}
		for _, vv := range v {
			dst.Add(k, vv)
		}
	}
}
