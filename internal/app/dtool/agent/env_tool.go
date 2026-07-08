package agent

import (
	"dev_tool/internal/app/dtool/define"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetPiExtensionsDir 返回 Pi 扩展目录路径
func GetPiExtensionsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".pi", "agent", "extensions")
}

// ScanInstalledTools 扫描已安装到 .pi/extensions/ 的 .ts 文件
func ScanInstalledTools() []define.AgentV2InstalledTool {
	extDir := GetPiExtensionsDir()
	if extDir == "" {
		return nil
	}

	entries, err := os.ReadDir(extDir)
	if err != nil {
		return nil
	}

	tools := make([]define.AgentV2InstalledTool, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".ts") {
			continue
		}
		tools = append(tools, define.AgentV2InstalledTool{
			Name:     strings.TrimSuffix(name, ".ts"),
			FilePath: filepath.Join(extDir, name),
		})
	}
	return tools
}

// RemoveInstalledTool 删除 .pi/extensions/ 下的指定扩展文件
func RemoveInstalledTool(name string) error {
	extDir := GetPiExtensionsDir()
	filePath := filepath.Join(extDir, name+".ts")
	return os.Remove(filePath)
}

// CheckExtensionFile 检查指定扩展的 .ts 文件是否存在于 .pi/extensions/
func CheckExtensionFile(filename string) (bool, string) {
	extDir := GetPiExtensionsDir()
	filePath := filepath.Join(extDir, filename+".ts")
	if _, err := os.Stat(filePath); err == nil {
		return true, filePath
	}
	return false, ""
}

// BuiltinEnvToolDefs 内置环境工具定义列表
var BuiltinEnvToolDefs = []define.AgentV2EnvToolItem{
	{
		Key:             "headroom",
		Name:            "Headroom (API Proxy & Context Compressor)",
		Description:     "LLM 上下文压缩层，60-95% token 节省（JSON），15-20% token 节省（代码）。代理模式零代码改动，支持 Anthropic/OpenAI/Gemini/Vertex/CloudCode/Bedrock 六个大模型服务商。",
		Icon:            "🧠",
		Homepage:        "https://github.com/chopratejas/headroom",
		InstallCmdHint:  GetHeadroomInstallHint(),
		ActivateCmdHint: "", // Headroom 不需要 "激活" — 直接启动代理进程
		DetectBin:       HeadroomDetectBin,
	},
	{
		Key:             "rtk",
		Name:            "RTK (Rust Token Killer)",
		Description:     "CLI 代理工具，自动过滤和压缩命令输出，为 LLM 节省 60-90% token。支持 100+ 常用命令，<10ms 开销。安装后执行激活命令即可对 Pi Agent 生效。",
		Icon:            "⚡",
		Homepage:        "https://github.com/rtk-ai/rtk",
		InstallCmdHint:  getRTKInstallHint(),
		ActivateCmdHint: "rtk init -g --agent pi",
		DetectBin:       "rtk",
	},
}

func getRTKInstallHint() string {
	switch runtime.GOOS {
	case "windows":
		return "从 https://github.com/rtk-ai/rtk/releases 下载 rtk-x86_64-pc-windows-msvc.zip，解压后将 rtk.exe 放入 PATH 目录"
	case "darwin":
		return "brew install rtk"
	default:
		return "curl -fsSL https://raw.githubusercontent.com/rtk-ai/rtk/refs/heads/master/install.sh | sh"
	}
}

// DetectEnvToolStatus 检测单个环境工具的安装状态
func DetectEnvToolStatus(def define.AgentV2EnvToolItem) define.AgentV2EnvToolStatus {
	st := define.AgentV2EnvToolStatus{
		AgentV2EnvToolItem: def,
		Status:             "not_installed",
	}

	// 检测二进制是否在 PATH 中
	_, err := exec.LookPath(def.DetectBin)
	if err != nil {
		return st
	}

	st.Installed = true

	// 尝试获取版本号
	checkCmd := getCheckCmd(def)
	if checkCmd != "" {
		out, err := runShellCmd(strings.Fields(checkCmd))
		if err == nil {
			st.Version = strings.TrimSpace(out)
			if st.Version != "" {
				if idx := strings.Index(st.Version, "\n"); idx > 0 {
					st.Version = st.Version[:idx]
				}
			}
		}
	}

	// headroom 不走 Pi 扩展文件机制，直接标记为 installed
	if def.Key == "headroom" {
		st.Status = "installed"
		return st
	}

	// 检测 Pi 扩展文件是否已安装到 .pi/extensions/
	if found, path := CheckExtensionFile(def.Key); found {
		st.ExtensionInstalled = true
		st.ExtensionFilePath = path
		st.Status = "activated"
	} else {
		// 回退：检测 shell hook 是否激活
		st.Activated = checkRTKHookActivated(def.Key)
		if st.Activated {
			st.Status = "activated"
		} else {
			st.Status = "installed"
		}
	}

	return st
}

// checkRTKHookActivated 检测环境工具 hook 是否已激活
func checkRTKHookActivated(key string) bool {
	switch key {
	case "rtk":
		// 尝试执行 rtk status 或检测 hook 注入情况
		out, err := runShellCmd(strings.Fields("rtk status"))
		if err == nil && strings.Contains(strings.ToLower(out), "active") {
			return true
		}
		// 备用检测：检查 shell rc 文件是否包含 rtk init
		return checkShellRCForRTK()
	}
	return false
}

// checkShellRCForRTK 快速检测常见 shell rc 文件中是否包含 rtk init 注入
func checkShellRCForRTK() bool {
	homeDir := ""
	switch runtime.GOOS {
	case "windows":
		// Windows: 检测 PowerShell profile
		cmd := exec.Command("powershell", "-NoProfile", "-Command", "echo $PROFILE")
		out, err := cmd.Output()
		if err == nil {
			profile := strings.TrimSpace(string(out))
			if profile != "" {
				data, err := exec.Command("powershell", "-NoProfile", "-Command",
					"if (Test-Path '"+profile+"') { Get-Content '"+profile+"' -Raw }").Output()
				if err == nil && strings.Contains(string(data), "rtk init") {
					return true
				}
			}
		}
	default:
		// Unix: 检测 ~/.bashrc / ~/.zshrc
		for _, rc := range []string{homeDir + "/.bashrc", homeDir + "/.zshrc", homeDir + "/.bash_profile"} {
			cmd := exec.Command("sh", "-c", "grep -l 'rtk init' "+rc+" 2>/dev/null")
			out, _ := cmd.Output()
			if strings.TrimSpace(string(out)) != "" {
				return true
			}
		}
	}
	return false
}

// getCheckCmd 生成检测版本号的命令
func getCheckCmd(t define.AgentV2EnvToolItem) string {
	switch t.Key {
	case "rtk":
		return "rtk --version"
	case "headroom":
		return "headroom --version"
	default:
		return t.DetectBin + " --version"
	}
}

// GetUninstallCmd 返回指定环境工具的卸载命令
func GetUninstallCmd(key string) string {
	switch key {
	case "rtk":
		return "rtk init -g --uninstall"
	default:
		return ""
	}
}

// runShellCmd 执行 shell 命令并返回输出
func runShellCmd(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[agent-v2/envtool] cmd failed: %v, err=%v", args, err)
		return "", err
	}
	return string(out), nil
}
