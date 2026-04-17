package main

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_js"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

var defaultServerURL = "http://localhost:17170"

// Config Agent 配置
type Config struct {
	ServerURL     string
	ClientID      string
	ClientVersion string
}

func main() {
	// 从环境变量读取配置
	serverURL := os.Getenv("DTOOL_SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}
	clientID := os.Getenv("DTOOL_CLIENT_ID")
	if clientID == "" {
		hostname, _ := os.Hostname()
		clientID = fmt.Sprintf("client_%s_%d", hostname, time.Now().Unix())
	}
	clientVersion := os.Getenv("DTOOL_CLIENT_VERSION")
	if clientVersion == "" {
		clientVersion = "2.0.0"
	}

	cfg := Config{
		ServerURL:     serverURL,
		ClientID:      clientID,
		ClientVersion: clientVersion,
	}

	fmt.Printf("dtool-agent 启动\n")
	fmt.Printf("服务端地址: %s\n", cfg.ServerURL)
	fmt.Printf("客户端ID: %s\n", cfg.ClientID)
	fmt.Printf("版本: %s\n", cfg.ClientVersion)

	// 初始化最小组件
	if err := initComponents(); err != nil {
		fmt.Printf("组件初始化失败: %s\n", err.Error())
		return
	}

	// 创建 WebSocket 客户端
	wsClient := NewWsClient(cfg)
	// 创建任务执行器
	taskRunner := NewTaskRunner(wsClient)

	wsClient.SetTaskHandler(taskRunner.HandleTask)

	// 注册
	registerResp, err := wsClient.Register()
	if err != nil {
		fmt.Printf("注册失败: %s\n", err.Error())
		return
	}
	// 从 Data 中提取详细信息
	if data, ok := registerResp["Data"].(map[string]any); ok {
		fmt.Printf("注册成功，服务端版本要求: %v\n", data["required_client_version"])
	} else {
		fmt.Printf("注册成功（原始响应）: %+v\n", registerResp)
	}

	// 连接 WebSocket
	if err := wsClient.Connect(); err != nil {
		fmt.Printf("WebSocket连接失败: %s\n", err.Error())
		return
	}

	// 环境准备（异步）
	go prepareRuntime(wsClient)

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("正在关闭...")
	wsClient.Close()
}

// initComponents 初始化 Agent 最小组件
func initComponents() error {
	// Agent 根目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}
	agentRootDir := filepath.Join(homeDir, ".dtool", "agent")

	// 初始化 EnvClient
	component.EnvClient = &define.Env{
		RootPath:           agentRootDir,
		AppName:            "dtool-agent",
		LogPath:            filepath.Join(agentRootDir, "logs"),
		NodePath:           "node",
		WebkitDriverPath:   filepath.Join(agentRootDir, "webkit_driver"),
		WebkitDataPath:     filepath.Join(agentRootDir, "webkit_data"),
		WebkitDownloadPath: filepath.Join(agentRootDir, "webkit_download"),
	}

	// 创建必要目录
	dirs := []string{
		component.EnvClient.LogPath,
		component.EnvClient.WebkitDriverPath,
		component.EnvClient.WebkitDataPath,
		component.EnvClient.WebkitDownloadPath,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
	}

	// 初始化 TBaseClient
	p_common.TBaseClient = &p_common.TBase{
		StartMillUnix: time.Now().UnixMilli(),
		LogPath:       component.EnvClient.LogPath,
	}

	// 初始化 PlaywrightClient
	component.PlaywrightClient = component.NewTPlaywright()
	component.PlaywrightClient.LockFileFullPath = filepath.Join(agentRootDir, "playwright.RunLock")

	// 初始化 TJasClient（通过嵌入的 JS 文件）
	p_common.TJasClient = &p_common.TJas{
		JsData: map[string]string{
			"p_js/tip.js":    p_js.TipJS,
			"p_js/info.js":   p_js.InfoJS,
			"p_js/delete.js": p_js.DeleteJS,
		},
	}

	gstool.FmtPrintlnLogTime(`Agent组件初始化完成`)
	return nil
}

// prepareRuntime 异步准备运行环境
func prepareRuntime(wsClient *WsClient) {
	// 检测 Node.js
	if !component.PlaywrightClient.EnsureNodeRuntime() {
		gstool.FmtPrintlnLogTime(`未检测到 Node.js，请安装后重启 Agent`)
		wsClient.SendStatus("preparing_runtime")
		return
	}
	gstool.FmtPrintlnLogTime(`Node.js 检测通过: %s`, component.EnvClient.NodePath)

	// 检查并更新 Playwright 核心
	// Agent 模式下使用空 SSE（不需要向前端输出安装进度）
	wsClient.SendStatus("preparing_runtime")
	component.PlaywrightClient.SmartCheckAndUpdate(nil)

	// 等待浏览器核心就绪
	for i := 0; i < 60; i++ {
		if component.PlaywrightClient.Pw != nil {
			gstool.FmtPrintlnLogTime(`Playwright 浏览器核心已就绪`)
			wsClient.SendStatus("online")
			return
		}
		time.Sleep(2 * time.Second)
	}
	gstool.FmtPrintlnLogTime(`Playwright 浏览器核心初始化超时`)
	wsClient.SendStatus("error")
}

func getHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func getUsername() string {
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	return username
}

func getOs() string {
	return runtime.GOOS
}

func getArch() string {
	return runtime.GOARCH
}
