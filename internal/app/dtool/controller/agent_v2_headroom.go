package controller

import (
	"dev_tool/internal/app/dtool/agent"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// ======================== Headroom 代理管理 ========================

// AgentV2HeadroomStatus 获取 Headroom 完整状态（配置 + 运行时状态）
func AgentV2HeadroomStatus(c *gin.Context) {
	var req struct {
		AgentId int `json:"agent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.AgentId <= 0 {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	// 读取保存的配置
	savedConfig := loadHeadroomConfig(req.AgentId)

	// 检测 headroom 安装状态
	installed, version := agent.DetectHeadroom()

	// 检测进程运行状态
	running, pid, startedAt := detectHeadroomProcess(req.AgentId)

	status := define.AgentV2HeadroomStatus{
		AgentV2HeadroomConfig: savedConfig,
		Installed:             installed,
		Version:               version,
		Running:               running,
		Pid:                   pid,
		StartedAt:             startedAt,
	}

	gsgin.GinResponseSuccess(c, "", status)
}

// AgentV2HeadroomConfigSave 保存 Headroom 配置
func AgentV2HeadroomConfigSave(c *gin.Context) {
	var req define.AgentV2HeadroomSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.AgentId <= 0 {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	// 默认端口
	if req.Config.Port <= 0 {
		req.Config.Port = agent.DefaultHeadroomPort
	}

	// 读取 agent 现有 config JSON
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT config FROM tbl_agent_v2 WHERE id = ?`, req.AgentId,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, "读取 Agent 配置失败: "+err.Error(), nil)
		return
	}
	if len(rows) == 0 {
		gsgin.GinResponseError(c, "Agent 不存在", nil)
		return
	}

	configMap := make(map[string]interface{})
	if raw := cast.ToString(rows[0]["config"]); raw != "" && raw != "{}" {
		if err := json.Unmarshal([]byte(raw), &configMap); err != nil {
			log.Printf("[headroom] parse agent config failed: %v", err)
		}
	}

	// 将 headroom 配置写入子键
	configMap["headroom"] = req.Config
	newConfig, err := json.Marshal(configMap)
	if err != nil {
		gsgin.GinResponseError(c, "序列化配置失败: "+err.Error(), nil)
		return
	}

	now := time.Now().Unix()
	_, err = common.DbMain.Client.ExecBySql(
		`UPDATE tbl_agent_v2 SET config = ?, updated_at = ? WHERE id = ?`,
		string(newConfig), now, req.AgentId,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, "保存失败: "+err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// AgentV2HeadroomProcess 控制 Headroom 代理进程（启动/停止/重启）
func AgentV2HeadroomProcess(c *gin.Context) {
	var req define.AgentV2HeadroomProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.AgentId <= 0 || req.Action == "" {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	// 读取 headroom 配置
	cfg := loadHeadroomConfig(req.AgentId)

	var err error
	var status *managedProcessStatus

	switch req.Action {
	case "start":
		status, err = headroomStart(req.AgentId, cfg)
	case "stop":
		status, err = headroomStop(req.AgentId, cfg)
	case "restart":
		status, err = headroomRestart(req.AgentId, cfg)
	default:
		gsgin.GinResponseError(c, "不支持的操作: "+req.Action, nil)
		return
	}

	if err != nil {
		gsgin.GinResponseError(c, "操作失败: "+err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", status)
}

// ======================== Headroom 升级 ========================

// AgentV2HeadroomUpgrade 执行 headroom update
func AgentV2HeadroomUpgrade(c *gin.Context) {
	var req define.AgentV2EnvToolUpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Key == "" {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	var output string
	var err error

	switch req.Key {
	case "headroom":
		output, err = agent.HeadroomUpgrade(req.Check, req.Pre)
	case "rtk":
		output, err = runRTKUpgrade(req.Check)
	default:
		gsgin.GinResponseError(c, "不支持的工具: "+req.Key, nil)
		return
	}

	if err != nil {
		// CombinedOutput 已同时捕获 stdout+stderr，直接展示给用户
		msg := output
		if msg == "" {
			msg = err.Error()
		}
		gsgin.GinResponseSuccess(c, "", define.AgentV2EnvToolUpgradeResponse{
			Output:  msg,
			Success: false,
			Check:   req.Check,
		})
		return
	}

	gsgin.GinResponseSuccess(c, "", define.AgentV2EnvToolUpgradeResponse{
		Output:  output,
		Success: true,
		Check:   req.Check,
	})
}

// ======================== Headroom 统计 ========================

// AgentV2HeadroomStats 获取 Headroom 统计信息
func AgentV2HeadroomStats(c *gin.Context) {
	var req struct {
		AgentId int `json:"agent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.AgentId <= 0 {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	cfg := loadHeadroomConfig(req.AgentId)
	port := cfg.Port
	if port <= 0 {
		port = agent.DefaultHeadroomPort
	}

	stats, err := agent.HeadroomFetchStats(port)
	if err != nil {
		gsgin.GinResponseError(c, "获取统计信息失败: "+err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", stats)
}

// ======================== Headroom 日志 ========================

// AgentV2HeadroomLogList 列出 Headroom 日志文件
func AgentV2HeadroomLogList(c *gin.Context) {
	rootPath := component.EnvClient.RootPath
	items, err := agent.ListHeadroomLogFiles(rootPath)
	if err != nil {
		gsgin.GinResponseError(c, "获取日志列表失败: "+err.Error(), nil)
		return
	}
	if items == nil {
		items = make([]define.AgentV2HeadroomLogItem, 0)
	}
	gsgin.GinResponseSuccess(c, "", gin.H{"list": items})
}

// AgentV2HeadroomLogRead 读取 Headroom 日志内容
func AgentV2HeadroomLogRead(c *gin.Context) {
	var req define.AgentV2HeadroomActionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.LogFile == "" {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	rootPath := component.EnvClient.RootPath
	content, err := agent.ReadHeadroomLogFile(rootPath, req.LogFile)
	if err != nil {
		gsgin.GinResponseError(c, "读取日志失败: "+err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", define.AgentV2HeadroomLogContentResponse{
		Content:  content,
		FileName: req.LogFile,
	})
}

// ======================== 自动启动 ========================

// AutoStartHeadroom 程序启动时自动检测并启动所有启用了 auto_start 的 Headroom
func AutoStartHeadroom() {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT id, config FROM tbl_agent_v2 WHERE type = 'pi'`,
	).All()
	if err != nil {
		log.Printf("[headroom] auto-start: query agents failed: %v", err)
		return
	}

	for _, row := range rows {
		agentId := cast.ToInt(row["id"])
		raw := cast.ToString(row["config"])

		var configMap map[string]json.RawMessage
		if err := json.Unmarshal([]byte(raw), &configMap); err != nil {
			continue
		}
		headroomRaw, ok := configMap["headroom"]
		if !ok {
			continue
		}

		var cfg define.AgentV2HeadroomConfig
		if err := json.Unmarshal(headroomRaw, &cfg); err != nil {
			continue
		}

		// auto_start 默认为 true（仅当显式设置为 false 时才跳过）
		// go 的 bool 零值为 false，但我们通过 JSON 序列化来判断用户是否明确设置了 false
		// 由于 config 可能来自不同版本的保存，我们用原始 JSON 来判断是否有 auto_start 字段
		var rawCfg map[string]interface{}
		if err := json.Unmarshal(headroomRaw, &rawCfg); err != nil {
			continue
		}

		autoStart := true // 默认开
		if val, exists := rawCfg["auto_start"]; exists {
			if v, ok := val.(bool); ok {
				autoStart = v
			}
		}

		if !autoStart {
			log.Printf("[headroom] auto-start: agent %d auto_start=false, skip", agentId)
			continue
		}

		// 检测是否已安装
		installed, _ := agent.DetectHeadroom()
		if !installed {
			log.Printf("[headroom] auto-start: agent %d headroom not installed, skip", agentId)
			continue
		}

		// 检测是否已在运行
		running, _, _ := detectHeadroomProcess(agentId)
		if running {
			log.Printf("[headroom] auto-start: agent %d headroom already running, skip", agentId)
			continue
		}

		// 启动
		log.Printf("[headroom] auto-start: starting headroom for agent %d", agentId)
		if _, err := headroomStart(agentId, cfg); err != nil {
			log.Printf("[headroom] auto-start: agent %d start failed: %v", agentId, err)
		} else {
			log.Printf("[headroom] auto-start: agent %d started successfully", agentId)
		}
	}
}

// ======================== Headroom 启动/停止辅助函数 ========================

// headroomStart 启动 Headroom 代理进程
func headroomStart(agentId int, cfg define.AgentV2HeadroomConfig) (*managedProcessStatus, error) {
	cmdLine := agent.BuildHeadroomProxyCommand(cfg)

	// 确保日志目录存在（供后续日志 API 使用，不嵌入命令）
	_ = agent.EnsureHeadroomLogDir(component.EnvClient.RootPath)
	_ = agent.GetHeadroomLogFilePath(component.EnvClient.RootPath, agentId) // 生成但不使用

	key := headroomProcessKey(agentId)
	dataMap := map[string]any{
		"key":          key,
		"name":         "headroom-agent-" + cast.ToString(agentId),
		"command_line": cmdLine,
	}

	syncManagedProcessLogDir()
	return toolManagedProcessClient.Start(dataMap, time.Now())
}

// headroomStop 停止 Headroom 代理进程
func headroomStop(agentId int, cfg define.AgentV2HeadroomConfig) (*managedProcessStatus, error) {
	cmdLine := agent.BuildHeadroomProxyCommand(cfg)
	key := headroomProcessKey(agentId)
	dataMap := map[string]any{
		"key":          key,
		"name":         "headroom-agent-" + cast.ToString(agentId),
		"command_line": cmdLine,
	}
	syncManagedProcessLogDir()
	return toolManagedProcessClient.Stop(dataMap, time.Now())
}

// headroomRestart 重启 Headroom 代理进程
func headroomRestart(agentId int, cfg define.AgentV2HeadroomConfig) (*managedProcessStatus, error) {
	cmdLine := agent.BuildHeadroomProxyCommand(cfg)

	_ = agent.EnsureHeadroomLogDir(component.EnvClient.RootPath)
	_ = agent.GetHeadroomLogFilePath(component.EnvClient.RootPath, agentId)

	key := headroomProcessKey(agentId)
	dataMap := map[string]any{
		"key":          key,
		"name":         "headroom-agent-" + cast.ToString(agentId),
		"command_line": cmdLine,
	}
	syncManagedProcessLogDir()
	return toolManagedProcessClient.Restart(dataMap, time.Now())
}

// ======================== RTK 升级辅助 ========================

// runRTKUpgrade 执行 RTK 升级
func runRTKUpgrade(check bool) (string, error) {
	// RTK 没有统一的 update 命令，根据安装方式决定
	// 先尝试 rtk --version 确认安装
	_, err := exec.LookPath("rtk")
	if err != nil {
		return "", fmt.Errorf("RTK 未安装")
	}

	if check {
		out, err := runCmd("rtk", "--version")
		if err != nil {
			return "", err
		}
		return "当前版本: " + strings.TrimSpace(out) + "\n请在终端手动检查新版本: https://github.com/rtk-ai/rtk/releases", nil
	}

	// 尝试不同升级方式
	var upgradeHint string
	switch runtime.GOOS {
	case "darwin":
		upgradeHint = "brew upgrade rtk"
	case "linux":
		upgradeHint = "curl -fsSL https://raw.githubusercontent.com/rtk-ai/rtk/refs/heads/master/install.sh | sh"
	case "windows":
		upgradeHint = "从 https://github.com/rtk-ai/rtk/releases 下载最新版本，替换 rtk.exe"
	}

	// 尝试执行 brew upgrade（macOS）
	if runtime.GOOS == "darwin" {
		brewOut, brewErr := runCmd("brew", "upgrade", "rtk")
		if brewErr == nil {
			return "升级成功:\n" + brewOut, nil
		}
	}

	return fmt.Sprintf("RTK 不支持自动升级，请手动执行:\n  %s", upgradeHint), nil
}

// runCmd 执行命令并返回输出
func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ======================== 内部辅助函数 ========================

// loadHeadroomConfig 从 agent config JSON 中读取 headroom 子配置
func loadHeadroomConfig(agentId int) define.AgentV2HeadroomConfig {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT config FROM tbl_agent_v2 WHERE id = ?`, agentId,
	).All()
	if err != nil || len(rows) == 0 {
		return define.AgentV2HeadroomConfig{Port: agent.DefaultHeadroomPort, AutoStart: true}
	}

	raw := cast.ToString(rows[0]["config"])
	if raw == "" || raw == "{}" {
		return define.AgentV2HeadroomConfig{Port: agent.DefaultHeadroomPort, AutoStart: true}
	}

	var configMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &configMap); err != nil {
		return define.AgentV2HeadroomConfig{Port: agent.DefaultHeadroomPort, AutoStart: true}
	}

	headroomRaw, ok := configMap["headroom"]
	if !ok {
		return define.AgentV2HeadroomConfig{Port: agent.DefaultHeadroomPort, AutoStart: true}
	}

	var cfg define.AgentV2HeadroomConfig
	if err := json.Unmarshal(headroomRaw, &cfg); err != nil {
		return define.AgentV2HeadroomConfig{Port: agent.DefaultHeadroomPort, AutoStart: true}
	}

	if cfg.Port <= 0 {
		cfg.Port = agent.DefaultHeadroomPort
	}

	// auto_start 默认 true
	if !hasAutoStart(headroomRaw) {
		cfg.AutoStart = true
	}

	return cfg
}

// hasAutoStart 检查原始 JSON 中是否显式设置了 auto_start
func hasAutoStart(raw json.RawMessage) bool {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	_, exists := m["auto_start"]
	return exists
}

// headroomProcessKey 生成 headroom 进程的唯一 key
func headroomProcessKey(agentId int) string {
	return "headroom-agent-" + cast.ToString(agentId)
}

// detectHeadroomProcess 检测 headroom 代理进程是否在运行
func detectHeadroomProcess(agentId int) (running bool, pid int32, startedAt int64) {
	cfg := loadHeadroomConfig(agentId)
	cmdLine := agent.BuildHeadroomProxyCommand(cfg)
	key := headroomProcessKey(agentId)

	syncManagedProcessLogDir()

	dataMap := map[string]any{
		"key":          key,
		"name":         "headroom-agent-" + cast.ToString(agentId),
		"command_line": cmdLine,
	}

	status, err := toolManagedProcessClient.Status(dataMap, time.Now())
	if err != nil {
		log.Printf("[headroom] process status check failed: %v", err)
		return false, 0, 0
	}

	if status == nil || !status.Running {
		return false, 0, 0
	}

	return true, status.PID, status.StartedAt
}

// syncManagedProcessLogDir 同步日志目录到 managedProcessManager
func syncManagedProcessLogDir() {
	toolManagedProcessClient.syncLogDir(getManagedProcessLogDir())
}

// ======================== 日志文件管理工具 ========================

// CleanupOrphanedHeadroomLogs 清理孤立日志文件（清理 48 小时前的日志）
func CleanupOrphanedHeadroomLogs() {
	if component.EnvClient == nil {
		return
	}
	rootPath := component.EnvClient.RootPath
	dir := filepath.Join(rootPath, "logs", agent.HeadroomLogSubDir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-48 * time.Hour)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			filePath := filepath.Join(dir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("[headroom] cleanup old log %s failed: %v", entry.Name(), err)
			} else {
				log.Printf("[headroom] cleaned old log: %s", entry.Name())
			}
		}
	}
}
