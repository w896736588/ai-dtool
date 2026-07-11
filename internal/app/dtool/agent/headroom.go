package agent

import (
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	// DefaultHeadroomPort Headroom 代理默认端口
	DefaultHeadroomPort = 8787
	// HeadroomDetectBin headroom 二进制检测名称
	HeadroomDetectBin = "headroom"
	// HeadroomLogSubDir  headroom 日志子目录
	HeadroomLogSubDir = "headroom"
)

// DetectHeadroom 检测 headroom CLI 是否安装并返回版本号
// 通过 exec.LookPath 检测二进制 + --version 获取版本
func DetectHeadroom() (installed bool, version string) {
	binPath, err := exec.LookPath(HeadroomDetectBin)
	if err != nil {
		return false, ""
	}

	installed = true
	_ = binPath // 已确认二进制存在

	// 获取版本号
	out, err := runShellCmd([]string{HeadroomDetectBin, "--version"})
	if err != nil {
		log.Printf("[headroom] version check failed: %v", err)
		return true, ""
	}
	version = strings.TrimSpace(out)
	return true, version
}

// BuildHeadroomProxyCommand 根据配置构建 headroom proxy 命令行
// 返回完整的命令字符串，供 managedProcessManager 执行
func BuildHeadroomProxyCommand(cfg define.AgentV2HeadroomConfig) string {
	port := cfg.Port
	if port <= 0 {
		port = DefaultHeadroomPort
	}

	var sb strings.Builder
	sb.WriteString("headroom proxy")
	fmt.Fprintf(&sb, " --port %d", port)

	if cfg.AnthropicApiUrl != "" {
		fmt.Fprintf(&sb, " --anthropic-api-url %s", cfg.AnthropicApiUrl)
	}
	if cfg.OpenaiApiUrl != "" {
		fmt.Fprintf(&sb, " --openai-api-url %s", cfg.OpenaiApiUrl)
	}
	if cfg.GeminiApiUrl != "" {
		fmt.Fprintf(&sb, " --gemini-api-url %s", cfg.GeminiApiUrl)
	}
	if cfg.CloudcodeApiUrl != "" {
		fmt.Fprintf(&sb, " --cloudcode-api-url %s", cfg.CloudcodeApiUrl)
	}
	if cfg.VertexApiUrl != "" {
		fmt.Fprintf(&sb, " --vertex-api-url %s", cfg.VertexApiUrl)
	}

	return sb.String()
}

// GetHeadroomInstallHint 返回 headroom 安装提示（跨平台）
func GetHeadroomInstallHint() string {
	switch runtime.GOOS {
	case "windows":
		return "pip install headroom-ai[all]   （需 Python 3.10+）"
	case "darwin":
		return "pip install headroom-ai[all]   （需 Python 3.10+）"
	default:
		return "pip install headroom-ai[all]   （需 Python 3.10+）"
	}
}

// HeadroomUpgrade 执行 headroom update 命令
// check: 仅检查新版本（--check），不实际升级
// pre: 包含预发布版本（--pre）
func HeadroomUpgrade(check, pre bool) (string, error) {
	args := []string{"update"}
	if check {
		args = append(args, "--check")
	}
	if pre {
		args = append(args, "--pre")
	}

	cmd := exec.Command(HeadroomDetectBin, args...)
	out, err := cmd.CombinedOutput() // 同时捕获 stdout + stderr
	return strings.TrimSpace(string(out)), err
}

// HeadroomFetchStats 从 headroom 代理拉取统计信息
func HeadroomFetchStats(port int) (*define.AgentV2HeadroomStatsResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/stats", port)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("无法连接 Headroom 代理 (%s): %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	rawJSON := string(body)
	items := parseStatsToItems(rawJSON)

	return &define.AgentV2HeadroomStatsResponse{
		Items:   items,
		RawJSON: rawJSON,
	}, nil
}

// statsLabelMap 统计字段中文翻译映射表（含嵌套 key）
var statsLabelMap = map[string]string{
	// === 顶层 ===
	"summary":                  "概览",
	"tokens":                   "Token 统计",
	"savings":                  "节省统计",
	"cost":                     "费用统计",
	"latency":                  "延迟统计",
	"overhead":                 "代理开销",
	"ttfb":                     "首字节延迟",
	"throughput":               "吞吐量",
	"requests":                 "请求统计",
	"compression":              "压缩缓存",
	"compression_cache":        "压缩缓存模式",
	"cache":                    "缓存配置",
	"rate_limiter":             "限流配置",
	"config":                   "代理配置",
	"proxy_inbound":            "入站代理",
	"router":                   "路由",
	"telemetry":                "遥测",
	"toin":                     "Toin 优化",
	"feedback_loop":            "反馈循环",
	"context_tool":             "命令行过滤 (RTK)",
	"cli_filtering":            "命令行过滤 (RTK)",
	"prefix_cache":             "前缀缓存",
	"pipeline_timing":          "管道耗时",
	"codex_ws":                 "Codex WebSocket",
	"otel":                     "OpenTelemetry",
	"langfuse":                 "Langfuse",
	"agent_usage":              "Agent 用量",
	"display_session":          "当前会话",
	"persistent_savings":       "历史节省",
	"savings_history":          "节省历史",
	"waste_signals":            "浪费信号",
	"compressions_by_strategy": "策略压缩统计",
	"tokens_saved_by_strategy": "策略节省 Token",
	"request_logs":             "请求日志",
	"recent_requests":          "最近请求",
	"subscription_window":      "订阅窗口",
	"log_full_messages":        "记录完整消息",
	"anon_telemetry_shipping":  "匿名遥测上报",
	"update_available":         "有新版本",

	// === summary 子字段 ===
	"summary.mode":                                               "代理模式",
	"summary.api_requests":                                       "API 请求数",
	"summary.primary_model":                                      "主要模型",
	"summary.compression.requests_compressed":                    "已压缩请求",
	"summary.compression.avg_compression_pct":                    "平均压缩率",
	"summary.compression.best_compression_pct":                   "最佳压缩率",
	"summary.compression.best_detail":                            "最佳详情",
	"summary.compression.total_tokens_removed":                   "累计移除 Token",
	"summary.compression.cli_filtering_tokens_avoided":           "CLI 过滤避免 Token",
	"summary.compression.total_tokens_saved_with_cli_filtering":  "含 CLI 过滤总节省 Token",
	"summary.compression.total_tokens_before_with_cli_filtering": "CLI 过滤前总 Token",
	"summary.compression.rtk_tokens_avoided":                     "RTK 避免 Token",
	"summary.compression.total_tokens_saved_with_rtk":            "含 RTK 总节省 Token",
	"summary.compression.total_tokens_before_with_rtk":           "RTK 前总 Token",
	"summary.mcp.compressions":                                   "MCP 压缩次数",
	"summary.mcp.tokens_removed":                                 "MCP 移除 Token",
	"summary.mcp.retrievals":                                     "MCP 检索次数",
	"summary.cost.without_headroom_usd":                          "无代理费用(USD)",
	"summary.cost.with_headroom_usd":                             "含代理费用(USD)",
	"summary.cost.total_saved_usd":                               "累计节省(USD)",
	"summary.cost.savings_pct":                                   "节省百分比",
	"summary.cost.breakdown.cache_savings_usd":                   "缓存节省(USD)",
	"summary.cost.breakdown.compression_savings_usd":             "压缩节省(USD)",

	// === tokens 子字段 ===
	"tokens.input":                      "输入 Token",
	"tokens.output":                     "输出 Token",
	"tokens.saved":                      "已节省 Token",
	"tokens.output_saved":               "输出节省 Token",
	"tokens.output_reduction_percent":   "输出减少率",
	"tokens.proxy_compression_saved":    "代理压缩节省 Token",
	"tokens.cli_filtering_saved":        "CLI 过滤节省 Token",
	"tokens.rtk_saved":                  "RTK 节省 Token",
	"tokens.lean_ctx_saved":             "精简上下文节省 Token",
	"tokens.cli_tokens_avoided":         "CLI 避免 Token",
	"tokens.all_layers_saved":           "全层节省 Token",
	"tokens.savings_percent":            "节省百分比",
	"tokens.active_savings_percent":     "活跃节省百分比",
	"tokens.proxy_savings_percent":      "代理节省百分比",
	"tokens.all_layers_savings_percent": "全层节省百分比",

	// === savings 子字段 ===
	"savings.total_tokens":                                  "累计 Token",
	"savings.by_layer.cli_filtering.tokens":                 "RTK 当前 Token",
	"savings.by_layer.cli_filtering.tokens_saved":           "RTK 已节省 Token",
	"savings.by_layer.cli_filtering.lifetime_savings_pct":   "RTK 历史节省率",
	"savings.by_layer.cli_filtering.lifetime.commands":      "RTK 历史命令数",
	"savings.by_layer.cli_filtering.lifetime.input_tokens":  "RTK 历史输入 Token",
	"savings.by_layer.cli_filtering.lifetime.output_tokens": "RTK 历史输出 Token",
	"savings.by_layer.cli_filtering.lifetime.tokens_saved":  "RTK 历史节省 Token",
	"savings.by_layer.compression.tokens":                   "压缩当前 Token",
	"savings.by_layer.compression.proxy_tokens":             "代理压缩 Token",
	"savings.by_layer.compression.all_layers_tokens":        "全层压缩 Token",

	// === cost 子字段 ===
	"cost.total_tokens_saved":           "累计节省 Token",
	"cost.total_input_tokens":           "累计输入 Token",
	"cost.total_input_cost_usd":         "累计费用(USD)",
	"cost.cost_with_headroom_usd":       "含代理费用(USD)",
	"cost.savings_usd":                  "节省(USD)",
	"cost.compression_savings_usd":      "压缩节省(USD)",
	"cost.cache_savings_usd":            "缓存节省(USD)",
	"cost.cache_write_5m_tokens":        "5分钟缓存写入 Token",
	"cost.cache_write_1h_tokens":        "1小时缓存写入 Token",
	"cost.cli_tokens_avoided":           "CLI 避免 Token",
	"cost.cli_filtering_tokens_avoided": "CLI 过滤避免 Token",
	"cost.budget_period":                "预算周期",

	// === latency ===
	"latency.average_ms":     "平均延迟(ms)",
	"latency.min_ms":         "最小延迟(ms)",
	"latency.max_ms":         "最大延迟(ms)",
	"latency.total_requests": "统计请求数",

	// === overhead ===
	"overhead.average_ms": "平均开销(ms)",
	"overhead.min_ms":     "最小开销(ms)",
	"overhead.max_ms":     "最大开销(ms)",

	// === ttfb ===
	"ttfb.average_ms": "平均首字节(ms)",
	"ttfb.min_ms":     "最小首字节(ms)",
	"ttfb.max_ms":     "最大首字节(ms)",

	// === throughput ===
	"throughput.rolling.input_wall_clock": "滚动输入耗时",
	"throughput.rolling.input_active_p50": "滚动输入P50",
	"throughput.rolling.input_active_p95": "滚动输入P95",
	"throughput.rolling.compression_p50":  "滚动压缩P50",
	"throughput.rolling.forward_p50":      "滚动转发P50",
	"throughput.rolling.generation_p50":   "滚动生成P50",
	"throughput.current.input_wall_clock": "当前输入耗时",
	"throughput.current.compression_p50":  "当前压缩P50",

	// === compression_cache ===
	"compression_cache.mode": "压缩模式",

	// === cache ===
	"cache.entries":     "当前条目",
	"cache.max_entries": "最大条目",
	"cache.total_hits":  "累计命中",
	"cache.ttl_seconds": "TTL(秒)",

	// === rate_limiter ===
	"rate_limiter.requests_per_minute": "每分钟请求限制",
	"rate_limiter.tokens_per_minute":   "每分钟Token限制",
	"rate_limiter.active_keys":         "活跃 Key 数",

	// === proxy_inbound ===
	"proxy_inbound.total":     "总连接数",
	"proxy_inbound.completed": "已完成",
	"proxy_inbound.active":    "活跃连接",

	// === config ===
	"config.compress_system_messages": "压缩系统消息",
	"config.compress_user_messages":   "压缩用户消息",
	"config.force_kompress":           "强制压缩",
	"config.min_tokens_to_crush":      "压缩最小Token",
	"config.max_items_after_crush":    "压缩后最大条目",

	// === display_session ===
	"display_session.requests":                "请求数",
	"display_session.tokens_saved":            "节省 Token",
	"display_session.compression_savings_usd": "压缩节省(USD)",
	"display_session.total_input_tokens":      "输入 Token",
	"display_session.total_input_cost_usd":    "费用(USD)",
	"display_session.savings_percent":         "节省率",

	// === persistent_savings ===
	"persistent_savings.lifetime.requests":                "历史请求数",
	"persistent_savings.lifetime.tokens_saved":            "历史节省 Token",
	"persistent_savings.lifetime.total_input_tokens":      "历史输入 Token",
	"persistent_savings.lifetime.compression_savings_usd": "历史压缩节省(USD)",
	"persistent_savings.history_points":                   "历史数据点",
	"persistent_savings.storage_path":                     "存储路径",

	// === context_tool / cli_filtering ===
	"context_tool.configured":                            "已配置工具",
	"context_tool.label":                                 "工具标签",
	"context_tool.available":                             "可用",
	"context_tool.stats.installed":                       "已安装",
	"context_tool.stats.scope":                           "作用域",
	"context_tool.stats.avg_savings_pct":                 "平均节省率",
	"context_tool.stats.lifetime_avg_savings_pct":        "历史平均节省率",
	"context_tool.stats.lifetime_total_commands":         "历史命令数",
	"context_tool.stats.lifetime_input_tokens":           "历史输入 Token",
	"context_tool.stats.lifetime_output_tokens":          "历史输出 Token",
	"context_tool.stats.lifetime_tokens_saved":           "历史节省 Token",
	"context_tool.stats.lifetime_total_time_ms":          "历史总耗时(ms)",
	"context_tool.stats.lifetime_savings_pct":            "历史节省率",
	"context_tool.stats.total_commands":                  "当前命令数",
	"context_tool.stats.input_tokens":                    "当前输入 Token",
	"context_tool.stats.output_tokens":                   "当前输出 Token",
	"context_tool.stats.tokens_saved":                    "当前节省 Token",
	"context_tool.stats.total_time_ms":                   "总耗时(ms)",
	"context_tool.stats.savings_pct":                     "节省率",
	"context_tool.stats.avg_savings_pct_scope":           "节省率统计范围",
	"context_tool.stats.sample_ttl_seconds":              "采样 TTL(秒)",
	"context_tool.stats.refresh_interval_seconds":        "刷新间隔(秒)",
	"context_tool.stats.sampled_at":                      "采样时间",
	"context_tool.stats.counter_reset_detected":          "检测到计数器重置",
	"context_tool.stats.baseline.commands":               "基线命令数",
	"context_tool.stats.baseline.input_tokens":           "基线输入 Token",
	"context_tool.stats.baseline.output_tokens":          "基线输出 Token",
	"context_tool.stats.baseline.tokens_saved":           "基线节省 Token",
	"context_tool.stats.baseline.total_time_ms":          "基线总耗时",
	"context_tool.stats.baseline.captured_at":            "基线采集时间",
	"context_tool.stats.session_baseline_total_commands": "会话基线命令数",
	"context_tool.stats.session_baseline_input_tokens":   "会话基线输入 Token",
	"context_tool.stats.session_baseline_output_tokens":  "会话基线输出 Token",
	"context_tool.stats.session_baseline_tokens_saved":   "会话基线节省 Token",
	"context_tool.stats.session_baseline_total_time_ms":  "会话基线总耗时",
	"context_tool.stats.session_baseline_captured_at":    "会话基线采集时间",
	"context_tool.stats.session.commands":                "会话命令数",
	"context_tool.stats.session.input_tokens":            "会话输入 Token",
	"context_tool.stats.session.output_tokens":           "会话输出 Token",
	"context_tool.stats.session.tokens_saved":            "会话节省 Token",
	"context_tool.stats.session.total_time_ms":           "会话总耗时",
	"context_tool.stats.session.avg_time_ms":             "会话平均耗时",
	"context_tool.stats.session.savings_pct":             "会话节省率",
	"context_tool.stats.lifetime.savings_pct":            "历史会话节省率",
	"context_tool.stats.tool":                            "工具名",

	// === cli_filtering (same structure as context_tool.stats) ===
	"cli_filtering.installed":                "已安装",
	"cli_filtering.scope":                    "作用域",
	"cli_filtering.avg_savings_pct":          "平均节省率",
	"cli_filtering.lifetime_avg_savings_pct": "历史平均节省率",
	"cli_filtering.lifetime_total_commands":  "历史命令数",
	"cli_filtering.lifetime_input_tokens":    "历史输入 Token",
	"cli_filtering.lifetime_output_tokens":   "历史输出 Token",
	"cli_filtering.lifetime_tokens_saved":    "历史节省 Token",
	"cli_filtering.lifetime_total_time_ms":   "历史总耗时(ms)",
	"cli_filtering.lifetime_savings_pct":     "历史节省率",
	"cli_filtering.total_commands":           "当前命令数",
	"cli_filtering.input_tokens":             "当前输入 Token",
	"cli_filtering.output_tokens":            "当前输出 Token",
	"cli_filtering.tokens_saved":             "当前节省 Token",
	"cli_filtering.total_time_ms":            "总耗时(ms)",
	"cli_filtering.savings_pct":              "节省率",
	"cli_filtering.avg_savings_pct_scope":    "节省率统计范围",
	"cli_filtering.sample_ttl_seconds":       "采样 TTL(秒)",
	"cli_filtering.refresh_interval_seconds": "刷新间隔(秒)",
	"cli_filtering.sampled_at":               "采样时间",
	"cli_filtering.counter_reset_detected":   "检测到计数器重置",
	"cli_filtering.baseline.commands":        "基线命令数",
	"cli_filtering.baseline.input_tokens":    "基线输入 Token",
	"cli_filtering.baseline.output_tokens":   "基线输出 Token",
	"cli_filtering.baseline.tokens_saved":    "基线节省 Token",
	"cli_filtering.baseline.total_time_ms":   "基线总耗时",
	"cli_filtering.baseline.captured_at":     "基线采集时间",
	"cli_filtering.tool":                     "工具名",

	// === telemetry ===
	"telemetry.enabled":               "已启用",
	"telemetry.total_compressions":    "累计压缩",
	"telemetry.total_retrievals":      "累计检索",
	"telemetry.global_retrieval_rate": "全局检索率",

	// === toin ===
	"toin.enabled":               "已启用",
	"toin.patterns_tracked":      "模式追踪数",
	"toin.total_compressions":    "累计压缩",
	"toin.total_retrievals":      "累计检索",
	"toin.global_retrieval_rate": "全局检索率",

	// === agent_usage ===
	"agent_usage.totals.requests":          "请求数",
	"agent_usage.totals.before_tokens":     "压缩前 Token",
	"agent_usage.totals.after_tokens":      "压缩后 Token",
	"agent_usage.totals.tokens_saved":      "节省 Token",
	"agent_usage.totals.savings_percent":   "节省率",
	"agent_usage.coverage.logged_requests": "记录请求数",
	"agent_usage.coverage.mode":            "覆盖模式",

	// === prefix_cache ===
	"prefix_cache.totals.cache_read_tokens":  "缓存读取 Token",
	"prefix_cache.totals.cache_write_tokens": "缓存写入 Token",
	"prefix_cache.totals.requests":           "请求数",
	"prefix_cache.totals.hit_requests":       "命中请求",
	"prefix_cache.totals.savings_usd":        "节省(USD)",
	"prefix_cache.totals.net_savings_usd":    "净节省(USD)",
	"prefix_cache.totals.hit_rate":           "命中率",
	"prefix_cache.totals.request_hit_rate":   "请求命中率",
	"prefix_cache.totals.bust_count":         "缓存失效次数",

	// === compression (CCR) ===
	"compression.ccr_entries":              "CCR 条目",
	"compression.ccr_max_entries":          "CCR 最大条目",
	"compression.original_tokens_cached":   "缓存原始 Token",
	"compression.compressed_tokens_cached": "缓存压缩 Token",
	"compression.ccr_retrievals":           "CCR 检索",

	// === codex_ws ===
	"codex_ws.units_total":             "单元总数",
	"codex_ws.units_modified_total":    "修改单元数",
	"codex_ws.frames_attempted_total":  "尝试帧数",
	"codex_ws.frames_compressed_total": "压缩帧数",
	"codex_ws.frames_failed_total":     "失败帧数",

	// === 其他 ===
	"requests.total":                      "总请求",
	"requests.cached":                     "缓存命中",
	"requests.rate_limited":               "被限流",
	"requests.failed":                     "失败",
	"feedback_loop.tools_tracked":         "追踪工具数",
	"feedback_loop.total_compressions":    "累计压缩",
	"feedback_loop.total_retrievals":      "累计检索",
	"feedback_loop.global_retrieval_rate": "全局检索率",
}

// parseStatsToItems 将 stats JSON 解析为结构化统计项（递归展平嵌套）
func parseStatsToItems(raw string) []define.AgentV2HeadroomStatsItem {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return []define.AgentV2HeadroomStatsItem{{
			Label: "原始数据",
			Key:   "raw",
			Value: raw,
		}}
	}

	return recursiveFlatten(data, "", 0)
}

// recursiveFlatten 递归展平嵌套 JSON，只保留有意义的值
func recursiveFlatten(data map[string]interface{}, prefix string, depth int) []define.AgentV2HeadroomStatsItem {
	var items []define.AgentV2HeadroomStatsItem

	for _, kv := range sortedKeys(data) {
		key := kv.key
		val := kv.val
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := val.(type) {
		case nil:
			continue // 跳过 nil

		case bool:
			items = append(items, statItem(fullKey, v))

		case float64:
			items = append(items, statItem(fullKey, v))

		case string:
			if v == "" {
				continue
			}
			items = append(items, statItem(fullKey, v))

		case map[string]interface{}:
			if len(v) == 0 {
				continue // 跳过空对象
			}
			// 深度限制防止无限递归
			if depth >= 3 {
				// 太深了，显示总结
				label := statsLabelMap[fullKey]
				if label == "" {
					label = fullKey
				}
				items = append(items, define.AgentV2HeadroomStatsItem{
					Label: label,
					Key:   fullKey,
					Value: fmt.Sprintf("[%d 个子字段]", len(v)),
				})
				continue
			}
			// 加分隔标题
			if depth == 0 {
				sectionLabel := statsLabelMap[key]
				if sectionLabel == "" {
					sectionLabel = key
				}
				items = append(items, define.AgentV2HeadroomStatsItem{
					Label: "── " + sectionLabel + " ──",
					Key:   "_group_",
					Value: "",
				})
			}
			// 递归展平
			children := recursiveFlatten(v, fullKey, depth+1)
			items = append(items, children...)

		case []interface{}:
			if len(v) == 0 {
				continue // 跳过空数组
			}
			items = append(items, statItem(fullKey, fmt.Sprintf("[%d 条记录]", len(v))))
		}
	}

	return items
}

// kvPair 键值对（用于排序）
type kvPair struct {
	key string
	val interface{}
}

// sortedKeys 返回按 key 字母排序的键值对
func sortedKeys(data map[string]interface{}) []kvPair {
	pairs := make([]kvPair, 0, len(data))
	for k, v := range data {
		pairs = append(pairs, kvPair{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].key < pairs[j].key
	})
	return pairs
}

// statItem 构建单个统计项（自动翻译和格式化）
func statItem(fullKey string, value interface{}) define.AgentV2HeadroomStatsItem {
	label := statsLabelMap[fullKey]
	if label == "" {
		// 兜底：取最后一个 . 之后的 key 名，并尝试智能翻译
		lastSeg := fullKey
		if idx := strings.LastIndex(fullKey, "."); idx >= 0 {
			lastSeg = fullKey[idx+1:]
		}
		label = smartTranslateKey(lastSeg)
		if label == lastSeg {
			// 真的没匹配，保留原始 key 并标记为灰色
			label = "[未翻译] " + lastSeg
		}
	}
	return define.AgentV2HeadroomStatsItem{
		Label: label,
		Key:   fullKey,
		Value: formatStatsValue(value),
	}
}

// smartTranslateKey 对未翻译的 key 名做后缀/前缀规则匹配
func smartTranslateKey(key string) string {
	// 精确简短 key
	switch key {
	case "total":
		return "总数"
	case "count":
		return "数量"
	case "size":
		return "大小"
	case "name":
		return "名称"
	case "type":
		return "类型"
	case "mode":
		return "模式"
	case "key":
		return "键"
	case "id":
		return "ID"
	case "url":
		return "地址"
	}

	// 后缀匹配
	suffixes := []struct{ suffix, zh string }{
		{"_usd", "(USD)"},
		{"_pct", "(%)"},
		{"_ms", "(ms)"},
		{"_percent", "(%)"},
		{"_tokens", " Token"},
		{"_tokens_saved", "节省 Token"},
		{"_tokens_avoided", "避免 Token"},
		{"_seconds", "(秒)"},
		{"_requests", "请求数"},
		{"_commands", "命令数"},
		{"_count", "次数"},
		{"_total", "总数"},
		{"_entries", "条目数"},
		{"_hits", "命中数"},
		{"_rate", "率"},
		{"_ratio", "比率"},
		{"_limit", "限制"},
		{"_bytes", "(字节)"},
		{"_enabled", "启用"},
		{"_available", "可用"},
		{"_installed", "已安装"},
		{"_active", "活跃"},
		{"_time", "时间"},
		{"_at", "时间"},
		{"_dir", "目录"},
		{"_path", "路径"},
		{"_version", "版本"},
		{"_period", "周期"},
		{"_level", "级别"},
		{"_key", " Key"},
		{"_sum", "总数"},
		{"_max", "最大值"},
		{"_min", "最小值"},
		{"_avg", "平均值"},
		{"_p50", " P50"},
		{"_p95", " P95"},
		{"_window", "窗口"},
		{"_points", "数据点"},
		{"_errors", "错误数"},
		{"_failed", "失败数"},
		{"_cached", "缓存数"},
		{"_completed", "完成数"},
	}

	for _, s := range suffixes {
		if strings.HasSuffix(key, s.suffix) {
			prefix := key[:len(key)-len(s.suffix)]
			return camelToWords(prefix) + s.zh
		}
	}

	// 通用 camel_case → 中文词
	return camelToWords(key)
}

// camelToWords 将 camel_case / snake_case 转为空格分隔的词
func camelToWords(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	// 不大写首字母，保持可读
	return s
}

// formatStatsValue 格式化值（数字智能缩写，bool转是/否，科学计数法格式化）
func formatStatsValue(v interface{}) string {
	switch val := v.(type) {
	case float64:
		return formatStatsNumber(val)
	case string:
		return val
	case bool:
		if val {
			return "是"
		}
		return "否"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatStatsNumber 智能格式化数值：大数缩写，科学计数法转换，比率转百分比
func formatStatsNumber(val float64) string {
	switch {
	case val >= 1e9:
		return fmt.Sprintf("%.2fB", val/1e9)
	case val >= 1e7:
		return fmt.Sprintf("%.1fM", val/1e6)
	case val >= 1e6:
		return fmt.Sprintf("%.2fM", val/1e6)
	case val >= 1e4:
		return fmt.Sprintf("%.0f", val)
	case val >= 10:
		return fmt.Sprintf("%.1f", val)
	case val > 0 && val < 1:
		return fmt.Sprintf("%.1f%%", val*100)
	default:
		return fmt.Sprintf("%.2f", val)
	}
}

// GetHeadroomLogDir 返回 headroom 日志目录（基于 rootPath 下的 logs/headroom/）
func GetHeadroomLogDir(rootPath string) string {
	return filepath.Join(rootPath, "logs", HeadroomLogSubDir)
}

// EnsureHeadroomLogDir 确保日志目录存在
func EnsureHeadroomLogDir(rootPath string) error {
	dir := GetHeadroomLogDir(rootPath)
	return os.MkdirAll(dir, 0755)
}

// GetHeadroomLogFilePath 生成 headroom 日志文件路径
func GetHeadroomLogFilePath(rootPath string, agentId int) string {
	dir := GetHeadroomLogDir(rootPath)
	dateStr := time.Now().Format("20060102")
	return filepath.Join(dir, fmt.Sprintf("headroom-agent-%d-%s.log", agentId, dateStr))
}

// ListHeadroomLogFiles 列出所有 headroom 日志文件
func ListHeadroomLogFiles(rootPath string) ([]define.AgentV2HeadroomLogItem, error) {
	dir := GetHeadroomLogDir(rootPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var items []define.AgentV2HeadroomLogItem
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, define.AgentV2HeadroomLogItem{
			Name:    entry.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
		})
	}

	// 按修改时间倒序
	sort.Slice(items, func(i, j int) bool {
		return items[i].ModTime > items[j].ModTime
	})

	return items, nil
}

// ReadHeadroomLogFile 读取指定日志文件内容（最多 200KB）
func ReadHeadroomLogFile(rootPath, fileName string) (string, error) {
	// 安全检查：防止路径遍历
	fileName = filepath.Base(fileName)
	if !strings.HasPrefix(fileName, "headroom-agent-") || !strings.HasSuffix(fileName, ".log") {
		return "", fmt.Errorf("无效的日志文件名: %s", fileName)
	}

	dir := GetHeadroomLogDir(rootPath)
	filePath := filepath.Join(dir, fileName)

	fi, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("日志文件不存在: %s", fileName)
	}

	maxSize := int64(200 * 1024) // 200KB
	offset := int64(0)
	if fi.Size() > maxSize {
		offset = fi.Size() - maxSize
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return "", err
		}
	}

	data := make([]byte, maxSize)
	n, err := f.Read(data)
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(data[:n]), nil
}
