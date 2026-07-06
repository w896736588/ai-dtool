package controller

import (
	"dev_tool/internal/app/dtool/agent"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// sessionProc 维护一个 Agent 会话的子进程实例
type sessionProc struct {
	mu      sync.Mutex
	adapter agent.AgentAdapter
	conn    *websocket.Conn
	ctx     chan struct{}
}

// sessionRegistry 全局会话注册表
var (
	sessionRegistry   = make(map[int]*sessionProc)
	sessionRegistryMu sync.Mutex
)

// parsePiConfig 解析 Agent config JSON 中的 Pi 配置
func parsePiConfig(configStr string) (provider, model, modelAddr, apiKey, sessionDir, extraArgs string) {
	if configStr == "" {
		return "", "", "", "", "", ""
	}
	var cfg struct {
		Provider   string `json:"provider"`
		Model      string `json:"model"`
		ModelAddr  string `json:"model_addr"`
		ApiKey     string `json:"api_key"`
		SessionDir string `json:"session_dir"`
		ExtraArgs  string `json:"extra_args"`
	}
	if err := json.Unmarshal([]byte(configStr), &cfg); err != nil {
		return "", "", "", "", "", ""
	}
	return cfg.Provider, cfg.Model, cfg.ModelAddr, cfg.ApiKey, cfg.SessionDir, cfg.ExtraArgs
}

// computeSessionDir 计算会话持久化目录
func computeSessionDir(agentCfgSessionDir string, agentId, sessionId int) string {
	if agentCfgSessionDir != "" {
		return filepath.Join(agentCfgSessionDir, "s"+cast.ToString(sessionId))
	}
	// 默认目录：{DefaultPiSessionDir}/{agent_id}/s{session_id}
	dir := filepath.Join(define.DefaultPiSessionDir, cast.ToString(agentId), "s"+cast.ToString(sessionId))
	os.MkdirAll(dir, 0755)
	return dir
}

// AgentV2WS 处理 Agent V2 WebSocket 连接
func AgentV2WS(c *gin.Context) {
	agentId := cast.ToInt(c.Query("agent_id"))
	sessionId := cast.ToInt(c.Query("session_id"))

	if agentId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id required"})
		return
	}

	// 查询 Agent 配置
	agentRow, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_v2 WHERE id = ?`, agentId,
	).One()
	if err != nil || len(agentRow) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	agentType := cast.ToString(agentRow["type"])
	adapter := getAdapterForType(agentType)

	// 检查是否已安装
	if !adapter.IsInstalled() {
		c.JSON(http.StatusBadRequest, gin.H{"error": adapter.InstallHint()})
		return
	}

	// 升级为 WebSocket
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[agent-v2-ws] upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// 获取工作空间目录
	workDir := ""
	if sessionId > 0 {
		sessionRow, _ := common.DbMain.Client.QueryBySql(
			`SELECT s.workspace_id, w.path FROM tbl_agent_v2_session s
			 LEFT JOIN tbl_agent_v2_workspace w ON s.workspace_id = w.id
			 WHERE s.id = ?`, sessionId,
		).One()
		if sessionRow != nil {
			workDir = cast.ToString(sessionRow["path"])
		}
	}

	// 解析 Pi 配置
	configStr := cast.ToString(agentRow["config"])
	provider, model, modelAddr, apiKey, cfgSessionDir, extraArgs := parsePiConfig(configStr)
	sessionDir := computeSessionDir(cfgSessionDir, agentId, sessionId)

	// 解析额外参数
	var extraArgsList []string
	if extraArgs != "" {
		extraArgsList = strings.Fields(extraArgs)
	}

	// 启动 Agent 子进程
	startCfg := agent.AgentStartConfig{
		WorkDir:    workDir,
		SessionDir: sessionDir,
		Provider:   provider,
		Model:      model,
		ModelAddr:  modelAddr,
		ApiKey:     apiKey,
		ExtraArgs:  extraArgsList,
	}

	procCtx := make(chan struct{})
	sp := &sessionProc{
		adapter: adapter,
		conn:    conn,
		ctx:     procCtx,
	}

	// 注册会话（sessionId 必须 > 0，避免 0 键冲突）
	if sessionId > 0 {
		sessionRegistryMu.Lock()
		sessionRegistry[sessionId] = sp
		sessionRegistryMu.Unlock()

		defer func() {
			sessionRegistryMu.Lock()
			delete(sessionRegistry, sessionId)
			sessionRegistryMu.Unlock()
		}()
	}

	// 打开事件持久化文件（追加模式，跨 WS 连接保留历史）
	var eventsFile *os.File
	if sessionDir != "" && sessionId > 0 {
		os.MkdirAll(sessionDir, 0755)
		eventsFilePath := filepath.Join(sessionDir, "dtool_events.jsonl")
		eventsFile, _ = os.OpenFile(eventsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}

	defer func() {
		close(procCtx)
		adapter.Stop()
		if eventsFile != nil {
			eventsFile.Close()
		}
	}()

	if err := adapter.Start(c.Request.Context(), startCfg); err != nil {
		conn.WriteJSON(gin.H{"type": "error", "error": "启动 Agent 失败: " + err.Error()})
		return
	}

	log.Printf("[agent-v2-ws] Agent started, agent_id=%d session_id=%d provider=%s model=%s session_dir=%s",
		agentId, sessionId, provider, model, sessionDir)

	// 更新会话的 session_dir
	if sessionId > 0 {
		common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_v2_session SET session_dir = ? WHERE id = ?`,
			sessionDir, sessionId,
		).Exec()
	}

	// 发送就绪状态（包含已读历史消息，使用统一的消息转换逻辑）
	historyMessages := readSessionMessagesList(sessionDir)
	conn.WriteJSON(gin.H{
		"type":  "state",
		"state": gin.H{"status": "ready", "agent_id": agentId, "session_id": sessionId, "session_dir": sessionDir, "model": model, "provider": provider},
	})
	if len(historyMessages) > 0 {
		conn.WriteJSON(gin.H{
			"type":     "history",
			"messages": historyMessages,
		})
	}

	// 启动两个 goroutine：读取 WebSocket（前端→后端），读取 Pi stdout（Pi→前端）
	wsDone := make(chan struct{})
	piDone := make(chan struct{})

	// WebSocket → Pi stdin
	go func() {
		defer func() {
			close(wsDone)
		}()
		for {
			select {
			case <-procCtx:
				return
			default:
			}

			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[agent-v2-ws] ws read error: %v", err)
				return
			}

			var wsMsg define.AgentV2WSMessage
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				log.Printf("[agent-v2-ws] invalid ws message: %v", err)
				continue
			}

			switch wsMsg.Type {
			case "command":
				if wsMsg.Command != nil {
					// 持久化用户 prompt 事件 + 更新会话标题
					cmdMap, ok := wsMsg.Command.(map[string]interface{})
					if ok {
						cmdType := cast.ToString(cmdMap["type"])
						if cmdType == "prompt" {
							userMsg := cast.ToString(cmdMap["message"])
							if userMsg != "" {
								// 持久化到 JSONL
								if eventsFile != nil {
									entry, _ := json.Marshal(map[string]interface{}{
										"type":    "user_text",
										"message": userMsg,
									})
									fmt.Fprintf(eventsFile, "%s\n", entry)
								}
								// 更新会话标题为最新用户提问（截断到 50 字符）
								title := userMsg
								if len(title) > 50 {
									title = title[:50] + "..."
								}
								if sessionId > 0 {
									now := time.Now().Unix()
									common.DbMain.Client.ExecBySql(
										`UPDATE tbl_agent_v2_session SET name = ?, updated_at = ? WHERE id = ?`,
										title, now, sessionId,
									).Exec()
								}
							}
						}
					}
					cmdBytes, _ := json.Marshal(wsMsg.Command)
					if err := adapter.SendCommand(cmdBytes); err != nil {
						log.Printf("[agent-v2-ws] send command error: %v", err)
						conn.WriteJSON(gin.H{"type": "error", "error": err.Error()})
					}
				}
			case "get_state":
				// 前端请求获取状态，后端转发到 Pi
				cmdBytes, _ := json.Marshal(map[string]string{"type": "get_state"})
				adapter.SendCommand(cmdBytes)
			case "get_session_stats":
				cmdBytes, _ := json.Marshal(map[string]string{"type": "get_session_stats"})
				adapter.SendCommand(cmdBytes)
			}
		}
	}()

	// Pi stdout → WebSocket
	go func() {
		defer close(piDone)
		for evt := range adapter.Events() {
			select {
			case <-procCtx:
				return
			default:
			}

			var rawEvt map[string]interface{}
			if err := json.Unmarshal(evt.Raw, &rawEvt); err != nil {
				log.Printf("[agent-v2-ws] parse event error: %v raw=%s", err, string(evt.Raw))
				continue
			}

			// 持久化 Pi 事件到 JSONL 文件
			if eventsFile != nil {
				fmt.Fprintf(eventsFile, "%s\n", string(evt.Raw))
			}

			// 根据事件类型补充元数据
			evtType := cast.ToString(rawEvt["type"])
			log.Printf("[agent-v2-ws] pi event → ws, type=%s raw=%s", evtType, string(evt.Raw))

			// 对 response 类型事件，标记命令名方便前端过滤
			if evtType == "response" {
				rawEvt["_command"] = cast.ToString(rawEvt["command"])
			}

			if err := conn.WriteJSON(gin.H{
				"type":  "event",
				"event": rawEvt,
			}); err != nil {
				log.Printf("[agent-v2-ws] ws write error: %v", err)
				return
			}
		}
	}()

	// 等待任一端结束
	select {
	case <-wsDone:
	case <-piDone:
	}
}
