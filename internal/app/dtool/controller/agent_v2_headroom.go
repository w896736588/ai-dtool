package controller

import (
	"dev_tool/internal/app/dtool/agent"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"log"
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
	cmdLine := agent.BuildHeadroomProxyCommand(cfg)
	key := headroomProcessKey(req.AgentId)

	dataMap := map[string]any{
		"key":          key,
		"name":         "headroom-agent-" + cast.ToString(req.AgentId),
		"command_line": cmdLine,
	}

	syncManagedProcessLogDir()

	var err error
	var status *managedProcessStatus

	switch req.Action {
	case "start":
		status, err = toolManagedProcessClient.Start(dataMap, time.Now())
	case "stop":
		status, err = toolManagedProcessClient.Stop(dataMap, time.Now())
	case "restart":
		status, err = toolManagedProcessClient.Restart(dataMap, time.Now())
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

// ======================== 内部辅助函数 ========================

// loadHeadroomConfig 从 agent config JSON 中读取 headroom 子配置
func loadHeadroomConfig(agentId int) define.AgentV2HeadroomConfig {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT config FROM tbl_agent_v2 WHERE id = ?`, agentId,
	).All()
	if err != nil || len(rows) == 0 {
		return define.AgentV2HeadroomConfig{Port: agent.DefaultHeadroomPort}
	}

	raw := cast.ToString(rows[0]["config"])
	if raw == "" || raw == "{}" {
		return define.AgentV2HeadroomConfig{Port: agent.DefaultHeadroomPort}
	}

	var configMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &configMap); err != nil {
		return define.AgentV2HeadroomConfig{Port: agent.DefaultHeadroomPort}
	}

	headroomRaw, ok := configMap["headroom"]
	if !ok {
		return define.AgentV2HeadroomConfig{Port: agent.DefaultHeadroomPort}
	}

	var cfg define.AgentV2HeadroomConfig
	if err := json.Unmarshal(headroomRaw, &cfg); err != nil {
		return define.AgentV2HeadroomConfig{Port: agent.DefaultHeadroomPort}
	}

	if cfg.Port <= 0 {
		cfg.Port = agent.DefaultHeadroomPort
	}
	return cfg
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
