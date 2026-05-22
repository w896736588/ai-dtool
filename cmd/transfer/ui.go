package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"embed"
)

//go:embed templates/index.html
var templateFS embed.FS

// adminHandler 管理服务器的路由处理器
type adminHandler struct {
	store   *Store
	config  *ConfigStore
	manager *ServerManager
}

// ServeHTTP 实现 http.Handler 接口
func (h *adminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/":
		h.serveUI(w, r)
	case path == "/api/events":
		h.handleSSE(w, r)
	case path == "/api/records":
		h.handleRecords(w, r)
	case path == "/api/records/clear":
		h.handleRecordsClear(w, r)
	case path == "/api/rules":
		h.handleRules(w, r)
	default:
		http.NotFound(w, r)
	}
}

// writeJSONError 写入 JSON 格式的错误响应，自动处理特殊字符转义
func writeJSONError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// serveUI 返回 UI 页面（直接输出嵌入的 HTML，避免 template 引擎干扰 Vue 语法）
func (h *adminHandler) serveUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data, err := templateFS.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, "读取页面失败", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// handleSSE SSE 事件流端点，实时推送请求记录更新
func (h *adminHandler) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "不支持 SSE", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// 订阅事件
	subID, ch := h.store.Subscribe()
	defer h.store.Unsubscribe(subID)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(evt)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// handleRecords 处理记录相关请求
func (h *adminHandler) handleRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	records := h.store.List()
	if records == nil {
		records = make([]*Record, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// handleRecordsClear 清空所有记录
func (h *adminHandler) handleRecordsClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	h.store.Clear()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleRules 处理规则 CRUD 请求
func (h *adminHandler) handleRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// 获取规则列表
		rules := h.config.List()
		if rules == nil {
			rules = make([]*ProxyRule, 0)
		}
		json.NewEncoder(w).Encode(rules)

	case http.MethodPost:
		// 新建规则
		var rule ProxyRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			http.Error(w, `{"error":"请求body解析失败"}`, http.StatusBadRequest)
			return
		}
		// 校验必填字段
		if rule.Port <= 0 || rule.TargetURL == "" || rule.Format == "" {
			http.Error(w, `{"error":"端口、目标地址、格式为必填项"}`, http.StatusBadRequest)
			return
		}
		if rule.Format != formatOpenAI && rule.Format != formatAnthropic {
			http.Error(w, `{"error":"格式必须是 openai 或 anthropic"}`, http.StatusBadRequest)
			return
		}
		if rule.Host == "" {
			rule.Host = "localhost"
		}
		if err := h.config.Add(&rule); err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 启动代理服务
		if err := h.manager.StartRule(&rule); err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "rule": rule})

	case http.MethodPut:
		// 更新规则（通过 query string ?id=xxx）
		ruleID := r.URL.Query().Get("id")
		if ruleID == "" {
			http.Error(w, `{"error":"缺少 id 参数"}`, http.StatusBadRequest)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var rule ProxyRule
		if err := json.Unmarshal(body, &rule); err != nil {
			http.Error(w, `{"error":"请求body解析失败"}`, http.StatusBadRequest)
			return
		}
		rule.ID = ruleID
		if rule.Port <= 0 || rule.TargetURL == "" || rule.Format == "" {
			http.Error(w, `{"error":"端口、目标地址、格式为必填项"}`, http.StatusBadRequest)
			return
		}
		if rule.Host == "" {
			rule.Host = "localhost"
		}
		// 先停止旧服务
		h.manager.StopRule(ruleID)
		// 更新配置
		if err := h.config.Update(&rule); err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 启动新服务
		if err := h.manager.StartRule(&rule); err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "rule": rule})

	case http.MethodDelete:
		// 删除规则（通过 query string ?id=xxx）
		ruleID := r.URL.Query().Get("id")
		if ruleID == "" {
			http.Error(w, `{"error":"缺少 id 参数"}`, http.StatusBadRequest)
			return
		}
		// 先停止代理服务
		h.manager.StopRule(ruleID)
		// 删除配置
		if err := h.config.Delete(ruleID); err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
