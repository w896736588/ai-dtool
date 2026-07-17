package controller

import (
	"context"
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

func init() {
	initSessionCleanup()
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// sessionProc owns one backend Agent process. A browser WebSocket can attach,
// detach, and reattach without controlling the process lifetime.
type sessionProc struct {
	mu           sync.Mutex
	wsWriteMu    sync.Mutex
	adapter      agent.AgentAdapter
	conn         *websocket.Conn
	ctx          chan struct{}
	createdAt    time.Time
	lastActiveAt time.Time
	agentID      int
	sessionID    int
	sessionDir   string
	configStr    string
	currentModel string
	eventWriteCh chan string
	writerDone   chan struct{}
	forwardDone  chan struct{}
	taskRunning  bool
	// 执行耗时累计（毫秒）：跨轮次累加；execTurnStartAt 为当前轮起始时间（零值表示未在运行）
	execAccumulatedMs int64
	execTurnStartAt   time.Time
}

var (
	sessionRegistry   = make(map[int]*sessionProc)
	sessionRegistryMu sync.Mutex
)

func initSessionCleanup() {
	go func() {
		for {
			time.Sleep(30 * time.Minute)
			cleanupStaleSessions(1 * time.Hour)
		}
	}()
}

func cleanupStaleSessions(maxAge time.Duration) {
	sessionRegistryMu.Lock()
	defer sessionRegistryMu.Unlock()

	now := time.Now()
	for id, sp := range sessionRegistry {
		sp.mu.Lock()
		taskRunning := sp.taskRunning
		connOpen := sp.conn != nil
		lastActiveAt := sp.lastActiveAt
		sp.mu.Unlock()

		if !taskRunning && !connOpen && now.Sub(lastActiveAt) > maxAge {
			log.Printf("[agent-v2/ws] cleanup stale session %d (idle since %s)", id, lastActiveAt.Format(time.RFC3339))
			sp.stop()
			delete(sessionRegistry, id)
		}
	}
}

// parsePiConfig resolves the provider/model/session configuration for a Pi run.
func parsePiConfig(configStr string, runtimeModel string) (providerName, model, sessionDir, extraArgs, runtimeDir string) {
	if configStr == "" {
		return "", "", "", "", ""
	}
	var cfg struct {
		Provider   string `json:"provider"`
		Model      string `json:"model"`
		SessionDir string `json:"session_dir"`
		ExtraArgs  string `json:"extra_args"`
		RuntimeDir string `json:"runtime_dir"`
		ProviderId int    `json:"provider_id"`
		ModelId    int    `json:"model_id"`
	}
	if err := json.Unmarshal([]byte(configStr), &cfg); err != nil {
		return "", "", "", "", ""
	}

	sessionDir = cfg.SessionDir
	extraArgs = cfg.ExtraArgs
	runtimeDir = cfg.RuntimeDir

	if runtimeModel != "" {
		idx := strings.LastIndex(runtimeModel, "/")
		if idx >= 0 {
			pName := runtimeModel[:idx]
			mName := runtimeModel[idx+1:]

			providerRow, err := common.DbMain.Client.QueryBySql(
				`SELECT id FROM tbl_ai_provider WHERE name = ? AND status = 1`,
				pName,
			).One()
			if err == nil && len(providerRow) > 0 {
				providerId := cast.ToInt(providerRow["id"])
				modelRow, err := common.DbMain.Client.QueryBySql(
					`SELECT model FROM tbl_ai_model WHERE model = ? AND provider_id = ? AND status = 1`,
					mName, providerId,
				).One()
				if err == nil && len(modelRow) > 0 {
					return pName, cast.ToString(modelRow["model"]), sessionDir, extraArgs, runtimeDir
				}
			}
		}
		log.Printf("[agent-v2/ws] runtime model %q not found, falling back to agent config", runtimeModel)
	}

	if cfg.ProviderId > 0 && cfg.ModelId > 0 {
		providerRow, err := common.DbMain.Client.QueryBySql(
			`SELECT name FROM tbl_ai_provider WHERE id = ? AND status = 1`,
			cfg.ProviderId,
		).One()
		if err == nil && len(providerRow) > 0 {
			cfg.Provider = cast.ToString(providerRow["name"])
		} else {
			log.Printf("[agent-v2/ws] provider_id=%d not found or disabled", cfg.ProviderId)
		}
		modelRow, err := common.DbMain.Client.QueryBySql(
			`SELECT model FROM tbl_ai_model WHERE id = ? AND status = 1`,
			cfg.ModelId,
		).One()
		if err == nil && len(modelRow) > 0 {
			cfg.Model = cast.ToString(modelRow["model"])
		} else {
			log.Printf("[agent-v2/ws] model_id=%d not found or disabled", cfg.ModelId)
		}
	}

	return cfg.Provider, cfg.Model, sessionDir, extraArgs, runtimeDir
}

func computeSessionDir(agentCfgSessionDir string, agentId, sessionId int) string {
	var dir string
	if agentCfgSessionDir != "" {
		dir = filepath.Join(agentCfgSessionDir, "s"+cast.ToString(sessionId))
	} else {
		dir = filepath.Join(define.DefaultPiSessionDir, cast.ToString(agentId), "s"+cast.ToString(sessionId))
	}
	if dir != "" {
		os.MkdirAll(dir, 0755)
	}
	return dir
}

func computePiSessionID(agentId, sessionId int) string {
	if sessionId <= 0 {
		return ""
	}
	return fmt.Sprintf("agent-%d-session-%d", agentId, sessionId)
}

func updateAgentV2SessionStatus(sessionId int, status string) {
	if sessionId <= 0 || status == "" {
		return
	}
	now := time.Now().Unix()
	if _, err := common.DbMain.Client.ExecBySql(
		`UPDATE tbl_agent_v2_session SET status = ?, updated_at = ? WHERE id = ?`,
		status, now, sessionId,
	).Exec(); err != nil {
		log.Printf("[agent-v2/ws] update session %d status=%s error: %v", sessionId, status, err)
	}
}

func stopAgentV2SessionProc(sessionId int) {
	if sessionId <= 0 {
		return
	}
	sessionRegistryMu.Lock()
	sp := sessionRegistry[sessionId]
	delete(sessionRegistry, sessionId)
	sessionRegistryMu.Unlock()
	if sp != nil {
		sp.stop()
	}
}

func getAgentV2SessionProc(sessionId int) *sessionProc {
	if sessionId <= 0 {
		return nil
	}
	sessionRegistryMu.Lock()
	defer sessionRegistryMu.Unlock()
	return sessionRegistry[sessionId]
}

func (sp *sessionProc) isProcessRunning() bool {
	if sp == nil {
		return false
	}
	sp.mu.Lock()
	adapter := sp.adapter
	sp.mu.Unlock()
	return adapter != nil && adapter.IsRunning()
}

func getOrStartAgentV2SessionProc(agentId, sessionId int, adapter agent.AgentAdapter, startCfg agent.AgentStartConfig, sessionDir, configStr, currentModel string) (*sessionProc, bool, error) {
	now := time.Now()

	sessionRegistryMu.Lock()
	if existing := sessionRegistry[sessionId]; existing != nil {
		existing.mu.Lock()
		taskRunning := existing.taskRunning
		existingModel := existing.currentModel
		processRunning := existing.adapter != nil && existing.adapter.IsRunning()
		existing.mu.Unlock()
		if processRunning && (taskRunning || existingModel == currentModel || currentModel == "") {
			sessionRegistryMu.Unlock()
			return existing, false, nil
		}
		delete(sessionRegistry, sessionId)
		sessionRegistryMu.Unlock()
		existing.stop()
	} else {
		sessionRegistryMu.Unlock()
	}

	eventsFilePath := ""
	if sessionDir != "" && sessionId > 0 {
		os.MkdirAll(sessionDir, 0755)
		eventsFilePath = filepath.Join(sessionDir, "dtool_events.jsonl")
	}

	sp := &sessionProc{
		agentID:      agentId,
		sessionID:    sessionId,
		sessionDir:   sessionDir,
		configStr:    configStr,
		currentModel: currentModel,
		eventWriteCh: make(chan string, 256),
		writerDone:   make(chan struct{}),
		forwardDone:  make(chan struct{}),
		lastActiveAt: now,
		adapter:      adapter,
		ctx:          make(chan struct{}),
		createdAt:    now,
		taskRunning:  false,
	}
	sp.startEventWriter(eventsFilePath)

	if err := adapter.Start(context.Background(), startCfg); err != nil {
		sp.stopContext()
		<-sp.writerDone
		return nil, false, err
	}

	sessionRegistryMu.Lock()
	sessionRegistry[sessionId] = sp
	sessionRegistryMu.Unlock()

	go sp.forwardPiEvents()
	return sp, true, nil
}

func (sp *sessionProc) stopContext() {
	select {
	case <-sp.ctx:
	default:
		close(sp.ctx)
	}
}

func (sp *sessionProc) stop() {
	sp.mu.Lock()
	sp.stopContext()
	conn := sp.conn
	sp.conn = nil
	sp.taskRunning = false
	sp.lastActiveAt = time.Now()
	sp.mu.Unlock()

	if conn != nil {
		conn.Close()
	}
	if sp.adapter != nil {
		sp.adapter.Stop()
	}
	if sp.forwardDone != nil {
		<-sp.forwardDone
	}
	if sp.writerDone != nil {
		<-sp.writerDone
	}
	// 进程结束时兜底落库执行耗时（防止 agent_end 未收到而丢失当前轮计时）
	sp.persistExecDuration()
	updateAgentV2SessionStatus(sp.sessionID, "active")
}

func (sp *sessionProc) startEventWriter(eventsFilePath string) {
	go func() {
		defer close(sp.writerDone)
		var eventsFile *os.File
		if eventsFilePath != "" {
			var err error
			eventsFile, err = os.OpenFile(eventsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("[agent-v2/ws] open events file %s error: %v", eventsFilePath, err)
			}
		}
		defer func() {
			if eventsFile != nil {
				eventsFile.Sync()
				eventsFile.Close()
			}
		}()

		writeLine := func(line string) {
			if eventsFile != nil && line != "" {
				fmt.Fprintf(eventsFile, "%s\n", line)
			}
		}
		for {
			select {
			case <-sp.ctx:
				for {
					select {
					case line := <-sp.eventWriteCh:
						writeLine(line)
					default:
						return
					}
				}
			case line := <-sp.eventWriteCh:
				writeLine(line)
			}
		}
	}()
}

func (sp *sessionProc) writeEventLine(line string) bool {
	if sp == nil || line == "" {
		return false
	}
	select {
	case <-sp.ctx:
		return false
	case sp.eventWriteCh <- line:
		return true
	}
}

func (sp *sessionProc) attachConn(conn *websocket.Conn) {
	sp.mu.Lock()
	oldConn := sp.conn
	sp.conn = conn
	sp.lastActiveAt = time.Now()
	sp.mu.Unlock()
	if oldConn != nil && oldConn != conn {
		oldConn.Close()
	}
}

func (sp *sessionProc) detachConn(conn *websocket.Conn) {
	sp.mu.Lock()
	if sp.conn == conn {
		sp.conn = nil
		sp.lastActiveAt = time.Now()
	}
	sp.mu.Unlock()
	if conn != nil {
		conn.Close()
	}
}

func (sp *sessionProc) currentConn() *websocket.Conn {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return sp.conn
}

func (sp *sessionProc) writeConn(conn *websocket.Conn, payload any) bool {
	if conn == nil {
		return false
	}
	sp.wsWriteMu.Lock()
	err := conn.WriteJSON(payload)
	sp.wsWriteMu.Unlock()
	if err != nil {
		log.Printf("[agent-v2/ws] ws write error: %v", err)
		sp.detachConn(conn)
		return false
	}
	return true
}

func (sp *sessionProc) markTaskRunning(running bool) {
	sp.mu.Lock()
	changed := sp.taskRunning != running
	sp.taskRunning = running
	sp.lastActiveAt = time.Now()
	sp.mu.Unlock()
	if changed {
		if running {
			updateAgentV2SessionStatus(sp.sessionID, "running")
		} else {
			updateAgentV2SessionStatus(sp.sessionID, "active")
		}
	}
}

func (sp *sessionProc) isTaskRunning() bool {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return sp.taskRunning
}

// execTotalMs 返回当前累计执行耗时（含正在运行轮的实时时长）
func (sp *sessionProc) execTotalMs() int64 {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	total := sp.execAccumulatedMs
	if !sp.execTurnStartAt.IsZero() {
		total += time.Since(sp.execTurnStartAt).Milliseconds()
	}
	return total
}

// persistExecDuration 将累计耗时落库；若当前轮仍在运行则先并入，随后清零轮起始（幂等）
func (sp *sessionProc) persistExecDuration() {
	sp.mu.Lock()
	if !sp.execTurnStartAt.IsZero() {
		sp.execAccumulatedMs += time.Since(sp.execTurnStartAt).Milliseconds()
		sp.execTurnStartAt = time.Time{}
	}
	acc := sp.execAccumulatedMs
	sp.mu.Unlock()
	if sp.sessionID > 0 {
		if _, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_v2_session SET exec_duration_ms = ? WHERE id = ?`,
			acc, sp.sessionID,
		).Exec(); err != nil {
			log.Printf("[agent-v2/ws] persist exec_duration_ms session=%d error: %v", sp.sessionID, err)
		}
	}
}

// pushExecProgress 通过 WS 向前端推送当前执行耗时（工具/思考完成等事件触发，或定时刷新）
func (sp *sessionProc) pushExecProgress() {
	sp.mu.Lock()
	running := !sp.execTurnStartAt.IsZero()
	sp.mu.Unlock()
	if !running && sp.execTotalMs() == 0 {
		return // 全新会话尚未开始，无需推送
	}
	conn := sp.currentConn()
	if conn == nil {
		return
	}
	sp.writeConn(conn, gin.H{
		"type": "event",
		"event": gin.H{
			"type":     "exec_progress",
			"total_ms": sp.execTotalMs(),
			"running":  running,
		},
	})
}

func AgentV2WS(c *gin.Context) {
	agentId := cast.ToInt(c.Query("agent_id"))
	sessionId := cast.ToInt(c.Query("session_id"))

	if agentId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id required"})
		return
	}
	if sessionId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
		return
	}

	agentRow, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_v2 WHERE id = ?`, agentId,
	).One()
	if err != nil || len(agentRow) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	agentType := cast.ToString(agentRow["type"])
	adapter := getAdapterForType(agentType)
	if !adapter.IsInstalled() {
		c.JSON(http.StatusBadRequest, gin.H{"error": adapter.InstallHint()})
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[agent-v2/ws] upgrade error: %v", err)
		return
	}

	workDir := ""
	sessionRow, _ := common.DbMain.Client.QueryBySql(
		`SELECT s.workspace_id, w.path FROM tbl_agent_v2_session s
		 LEFT JOIN tbl_agent_v2_workspace w ON s.workspace_id = w.id
		 WHERE s.id = ? AND s.agent_id = ?`, sessionId, agentId,
	).One()
	if sessionRow != nil {
		workDir = cast.ToString(sessionRow["path"])
	}

	configStr := cast.ToString(agentRow["config"])
	runtimeModel := c.Query("model")
	providerName, model, cfgSessionDir, extraArgs, runtimeDir := parsePiConfig(configStr, runtimeModel)
	sessionDir := computeSessionDir(cfgSessionDir, agentId, sessionId)

	if err := syncPiModelsConfig(); err != nil {
		log.Printf("[agent-v2/ws] syncPiModelsConfig error: %v", err)
	}

	var extraArgsList []string
	if extraArgs != "" {
		extraArgsList = strings.Fields(extraArgs)
	}

	startCfg := agent.AgentStartConfig{
		WorkDir:    workDir,
		SessionDir: sessionDir,
		SessionID:  computePiSessionID(agentId, sessionId),
		Provider:   providerName,
		Model:      model,
		ExtraArgs:  extraArgsList,
		Env:        map[string]string{},
	}
	// 运行目录：指定时通过 PI_CODING_AGENT_DIR 让该 Pi 实例使用独立的数据/配置目录
	if runtimeDir != "" {
		startCfg.Env["PI_CODING_AGENT_DIR"] = expandHome(runtimeDir)
	}

	attachOnly := c.Query("attach_only") == "1"
	historyMessages := readSessionMessagesList(sessionDir)
	writeStateAndHistory := func(running bool, status string) {
		sp := getAgentV2SessionProc(sessionId)
		spSessionDir := sessionDir
		if sp != nil && sp.sessionDir != "" {
			spSessionDir = sp.sessionDir
		}
		conn.WriteJSON(gin.H{
			"type": "state",
			"state": gin.H{
				"status":      status,
				"running":     running,
				"agent_id":    agentId,
				"session_id":  sessionId,
				"session_dir": spSessionDir,
				"model":       model,
				"provider":    providerName,
			},
		})
		if len(historyMessages) > 0 {
			conn.WriteJSON(gin.H{
				"type":     "history",
				"messages": historyMessages,
			})
		}
	}

	var sp *sessionProc
	started := false
	if attachOnly {
		sp = getAgentV2SessionProc(sessionId)
		if sp == nil || !sp.isProcessRunning() || !sp.isTaskRunning() {
			updateAgentV2SessionStatus(sessionId, "active")
			writeStateAndHistory(false, "stale")
			conn.Close()
			return
		}
	} else {
		var err error
		sp, started, err = getOrStartAgentV2SessionProc(agentId, sessionId, adapter, startCfg, sessionDir, configStr, model)
		if err != nil {
			conn.WriteJSON(gin.H{"type": "error", "error": "启动 Agent 失败: " + err.Error()})
			conn.Close()
			return
		}
	}
	sp.attachConn(conn)
	defer sp.detachConn(conn)

	// 连接建立即推送一次当前执行耗时，刷新页面后能立即显示进行中的计时
	sp.pushExecProgress()

	action := "attached"
	if started {
		action = "started"
	}
	log.Printf("[agent-v2/ws] Agent %s, agent_id=%d session_id=%d provider=%s model=%s session_dir=%s",
		action, agentId, sessionId, providerName, model, sessionDir)

	common.DbMain.Client.ExecBySql(
		`UPDATE tbl_agent_v2_session SET session_dir = ? WHERE id = ?`,
		sessionDir, sessionId,
	).Exec()

	writeStateAndHistory(sp.isTaskRunning(), "ready")

	sp.readWSCommands(conn, sessionId, sessionDir, configStr, model)
}

func (sp *sessionProc) readWSCommands(conn *websocket.Conn, sessionId int, sessionDir, configStr, currentModel string) {
	for {
		select {
		case <-sp.ctx:
			return
		default:
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[agent-v2/ws] ws read error: %v", err)
			return
		}

		var wsMsg define.AgentV2WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Printf("[agent-v2/ws] invalid ws message: %v", err)
			continue
		}

		switch wsMsg.Type {
		case "command":
			if wsMsg.Command == nil {
				continue
			}
			isPromptCommand := false
			cmdMap, ok := wsMsg.Command.(map[string]interface{})
			if ok && cast.ToString(cmdMap["type"]) == "prompt" {
				userMsg := cast.ToString(cmdMap["message"])
				if userMsg != "" {
					isPromptCommand = true
					entry, _ := json.Marshal(map[string]interface{}{
						"type":    "user_text",
						"message": userMsg,
					})
					sp.writeEventLine(string(entry))
					sp.markTaskRunning(true)

					title := userMsg
					// 按 rune（字符）截断，避免按字节切 UTF-8 多字节中文产生乱码（如 "你当前的任务是…" 被切成 "�"）
					runes := []rune(userMsg)
					if len(runes) > 50 {
						title = string(runes[:50]) + "..."
					}
					if sessionId > 0 {
						now := time.Now().Unix()
						common.DbMain.Client.ExecBySql(
							`UPDATE tbl_agent_v2_session SET name = ?, updated_at = ?, model_name = ?, status = ? WHERE id = ?`,
							title, now, currentModel, "running", sessionId,
						).Exec()
					}
				}
			}

			cmdBytes, _ := json.Marshal(wsMsg.Command)
			if err := sp.adapter.SendCommand(cmdBytes); err != nil {
				log.Printf("[agent-v2/ws] send command error: %v", err)
				if isPromptCommand {
					sp.markTaskRunning(false)
				}
				sp.writeConn(conn, gin.H{"type": "error", "error": err.Error()})
			}
		case "get_state":
			cmdBytes, _ := json.Marshal(map[string]string{"type": "get_state"})
			sp.adapter.SendCommand(cmdBytes)
		case "get_session_stats":
			modelsCtx := parseModelsCtx(configStr)
			stats := computeSessionStats(sessionDir, modelsCtx)
			sp.writeConn(conn, gin.H{
				"type": "event",
				"event": gin.H{
					"type":     "response",
					"command":  "get_session_stats",
					"_command": "get_session_stats",
					"success":  true,
					"data":     stats,
				},
			})
		}
	}
}

// isExecCompletionEvent 判断是否为“完成类”事件（工具调用完成/思考完成/消息完成等），
// 此类事件触发一次执行耗时推送，使前端在每步完成时即时刷新。
func isExecCompletionEvent(evtType string) bool {
	switch evtType {
	case "tool_execution_end", "thinking_end", "message_end", "agent_end", "step_end", "turn_end":
		return true
	}
	return strings.HasSuffix(evtType, "_end") || strings.Contains(evtType, "complete")
}

func (sp *sessionProc) forwardPiEvents() {
	defer close(sp.forwardDone)
	defer sp.markTaskRunning(false)

	// 定时推送执行耗时，保证长工具执行等“无事件间隙”也能实时显示
	progressTicker := time.NewTicker(2 * time.Second)
	defer progressTicker.Stop()
	go func() {
		for {
			select {
			case <-sp.ctx:
				return
			case <-progressTicker.C:
				sp.mu.Lock()
				running := !sp.execTurnStartAt.IsZero()
				sp.mu.Unlock()
				if running {
					sp.pushExecProgress()
				}
			}
		}
	}()

	for evt := range sp.adapter.Events() {
		select {
		case <-sp.ctx:
			return
		default:
		}

		var rawEvt map[string]interface{}
		if err := json.Unmarshal(evt.Raw, &rawEvt); err != nil {
			log.Printf("[agent-v2/ws] parse event error: %v raw=%s", err, string(evt.Raw))
			continue
		}

		sp.writeEventLine(string(evt.Raw))

		evtType := cast.ToString(rawEvt["type"])
		log.Printf("[agent-v2/ws] pi event -> ws, type=%s raw=%s", evtType, string(evt.Raw))
		switch evtType {
		case "agent_start":
			sp.mu.Lock()
			sp.execTurnStartAt = time.Now()
			sp.mu.Unlock()
			sp.markTaskRunning(true)
			sp.pushExecProgress()
		case "agent_end":
			sp.persistExecDuration()
			sp.markTaskRunning(false)
			sp.pushExecProgress()
		case "response":
			rawEvt["_command"] = cast.ToString(rawEvt["command"])
		}

		// 完成类事件（工具/思考/消息完成）即时推送最新耗时
		if isExecCompletionEvent(evtType) {
			sp.pushExecProgress()
		}

		conn := sp.currentConn()
		if conn == nil {
			continue
		}
		sp.writeConn(conn, gin.H{
			"type":  "event",
			"event": rawEvt,
		})
	}
}
