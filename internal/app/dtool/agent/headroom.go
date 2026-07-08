package agent

import (
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

const (
	// DefaultHeadroomPort Headroom 代理默认端口
	DefaultHeadroomPort = 8787
	// HeadroomDetectBin headroom 二进制检测名称
	HeadroomDetectBin = "headroom"
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
