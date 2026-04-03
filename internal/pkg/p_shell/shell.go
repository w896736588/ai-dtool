package p_shell

import (
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"io"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"time"
	"unsafe"

	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
	"golang.org/x/crypto/ssh"
)

type Shell struct {
	ShellClientMap      map[string]*gsssh.SshTerminal
	ShellClientPoolMap  map[string][]*gsssh.SshTerminal
	ShellClientPoolNext map[string]int
	ShellClientStartMap map[*gsssh.SshTerminal]int64
	ShellClientLastUsed map[*gsssh.SshTerminal]int64
	lock                sync.Mutex
	LogPath             string
	log                 *gstool.GsSlog
}

const maxShellPoolSize = 20
const shellIdleTimeout = 3 * time.Minute
const shellIdleCleanTicker = 30 * time.Second
const getClientBusyWaitTimeout = 5 * time.Second
const getClientBusyWaitInterval = 100 * time.Millisecond

var terminalBusyInspector = isTerminalBusy

// canSendSse 判断当前 SSE 是否仍然可用，避免向空连接发送消息。
func canSendSse(sse *p_sse.SseShell) bool {
	return sse != nil && sse.Sse != nil
}

type receiveBinder interface {
	SetFuncReceiveMsg(func(string) string)
}

// makeReceiveHandler 根据格式化函数构造统一的终端消息转发逻辑。
func makeReceiveHandler(sse *p_sse.SseShell, formatStream func(string) []string) func(string) string {
	return func(msg string) string {
		if formatStream != nil {
			msgList := formatStream(msg)
			for _, line := range msgList {
				if canSendSse(sse) {
					sse.Send(line)
				}
			}
		} else if canSendSse(sse) {
			sse.Send(msg)
		}
		return msg
	}
}

// bindReceiveHandler 为终端重新绑定消息接收函数，便于连接复用时切换 SSE。
func bindReceiveHandler(target receiveBinder, sse *p_sse.SseShell, formatStream func(string) []string) {
	target.SetFuncReceiveMsg(makeReceiveHandler(sse, formatStream))
}

// splitPoolKey 提取复合 key 的前缀部分，兼容旧调用方按前缀删除连接的逻辑。
func splitPoolKey(uniqueKey string) string {
	if uniqueKey == "" {
		return ""
	}
	keyList := strings.SplitN(uniqueKey, "#", 2)
	return keyList[0]
}

// resolvePoolKey 返回连接池 key。
// 当前优先使用完整 shellClientId，避免同一 ssh 下不同业务面板误复用交互式终端。
func resolvePoolKey(sshConfig map[string]any, shellClientId string) string {
	if shellClientId != "" {
		return shellClientId
	}
	if sshId := cast.ToString(sshConfig["id"]); sshId != "" {
		return sshId
	}
	return shellClientId
}

// NewShell 初始化 Shell 管理器，并启动空闲连接清理协程。
func NewShell(logPath string) *Shell {
	log := gstool.NewSlog3(logPath, "shell")
	_ = log.CleanOldLogs(2)
	shell := &Shell{
		ShellClientMap:      make(map[string]*gsssh.SshTerminal),
		ShellClientPoolMap:  make(map[string][]*gsssh.SshTerminal),
		ShellClientPoolNext: make(map[string]int),
		ShellClientStartMap: make(map[*gsssh.SshTerminal]int64),
		ShellClientLastUsed: make(map[*gsssh.SshTerminal]int64),
		log:                 log,
		LogPath:             logPath,
	}
	// 启动连接池空闲清理协程：超过 3 分钟未使用的连接将自动断开并移除。
	go shell.startIdleCleaner()
	return shell
}

// removeClientFromPoolLocked 在已持锁前提下，从指定池中移除目标连接并关闭它。
func (h *Shell) removeClientFromPoolLocked(poolKey string, target *gsssh.SshTerminal) {
	pool, ok := h.ShellClientPoolMap[poolKey]
	if !ok || len(pool) == 0 {
		return
	}
	newPool := make([]*gsssh.SshTerminal, 0, len(pool))
	for _, item := range pool {
		if item == nil {
			continue
		}
		if item == target {
			item.CloseTerminal()
			delete(h.ShellClientStartMap, item)
			delete(h.ShellClientLastUsed, item)
			continue
		}
		newPool = append(newPool, item)
	}
	if len(newPool) == 0 {
		delete(h.ShellClientPoolMap, poolKey)
		delete(h.ShellClientPoolNext, poolKey)
		return
	}
	h.ShellClientPoolMap[poolKey] = newPool
	if h.ShellClientPoolNext[poolKey] >= len(newPool) {
		h.ShellClientPoolNext[poolKey] = 0
	}
}

// createShellClient 创建一个新的交互式 SSH 终端，并绑定断线与消息转发处理。
func (h *Shell) createShellClient(sshConfig map[string]any, poolKey string, sse *p_sse.SseShell,
	formatStream func(string) []string, promptKeywords []string, promptFunc func(string, io.WriteCloser, *ssh.Session) string) (*gsssh.SshTerminal, error) {
	start := time.Now()
	sshId := cast.ToString(sshConfig["id"])
	gstool.FmtPrintlnLogTime(`[Shell.createShellClient] enter ssh_id=%s pool_key=%s`, sshId, poolKey)
	gsShell := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
		Name:     "",
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
	}))

	gsShell.SetFuncBroken(func(msg string) {
		gstool.FmtPrintlnLogTime(`[Shell.createShellClient] broken ssh_id=%s pool_key=%s err=%s`, sshId, poolKey, msg)
		if canSendSse(sse) {
			sse.Send(" connection broken, will reconnect on next action: " + msg + "\n")
		}
		h.lock.Lock()
		defer h.lock.Unlock()
		h.removeClientFromPoolLocked(poolKey, gsShell)
	})

	gsShell.SetPtyConfig(gsssh.PtyConfig{Echo: 1})
	gsShell.SetMaxBufferSize(2 * 1024 * 1024)
	gstool.FmtPrintlnLogTime(`[Shell.createShellClient] run pwd begin ssh_id=%s pool_key=%s`, sshId, poolKey)
	pwdStart := time.Now()
	_, err := gsShell.RunCommandWait("pwd", 40*time.Second)
	gstool.FmtPrintlnLogTime(`[Shell.createShellClient] run pwd end ssh_id=%s pool_key=%s cost_ms=%d err=%v`,
		sshId, poolKey, time.Since(pwdStart).Milliseconds(), err)
	if err != nil {
		return nil, err
	}

	bindReceiveHandler(gsShell, sse, formatStream)

	if len(promptKeywords) == 0 {
		promptKeywords = []string{"Username for", "Password for", "passphrase", "Passphrase"}
	}
	gsShell.SetAuthPromptKeywords(promptKeywords)
	if promptFunc != nil {
		gsShell.SetFuncAuthPrompt(promptFunc)
	}
	gstool.FmtPrintlnLogTime(`[Shell.createShellClient] return success ssh_id=%s pool_key=%s total_cost_ms=%d`,
		sshId, poolKey, time.Since(start).Milliseconds())
	return gsShell, nil
}

// startIdleCleaner 周期性扫描并清理空闲连接。
func (h *Shell) startIdleCleaner() {
	ticker := time.NewTicker(shellIdleCleanTicker)
	defer ticker.Stop()
	for range ticker.C {
		h.cleanupIdleClients()
	}
}

// cleanupIdleClients 断开并移除超过空闲阈值的连接。
func (h *Shell) cleanupIdleClients() {
	now := time.Now().Unix()
	timeoutSeconds := int64(shellIdleTimeout / time.Second)
	idleClients := make([]*gsssh.SshTerminal, 0)

	h.lock.Lock()
	// 清理连接池中的空闲连接
	for poolKey, pool := range h.ShellClientPoolMap {
		newPool := make([]*gsssh.SshTerminal, 0, len(pool))
		for _, client := range pool {
			if client == nil {
				continue
			}
			// 正在执行命令的连接不清理，避免误断开长任务。
			// 注意：command 字段不会可靠清空，需用互斥锁状态判断是否忙碌。
			if terminalBusyInspector(client) {
				newPool = append(newPool, client)
				continue
			}
			lastUsed := h.ShellClientLastUsed[client]
			if lastUsed == 0 {
				lastUsed = h.ShellClientStartMap[client]
			}
			if lastUsed > 0 && now-lastUsed >= timeoutSeconds {
				idleClients = append(idleClients, client)
				delete(h.ShellClientStartMap, client)
				delete(h.ShellClientLastUsed, client)
				continue
			}
			newPool = append(newPool, client)
		}
		if len(newPool) == 0 {
			delete(h.ShellClientPoolMap, poolKey)
			delete(h.ShellClientPoolNext, poolKey)
			continue
		}
		h.ShellClientPoolMap[poolKey] = newPool
		if h.ShellClientPoolNext[poolKey] >= len(newPool) {
			h.ShellClientPoolNext[poolKey] = 0
		}
	}

	// 清理一对一缓存中的空闲连接
	for shellClientId, client := range h.ShellClientMap {
		if client == nil {
			delete(h.ShellClientMap, shellClientId)
			continue
		}
		// 正在执行命令的连接不清理，避免误断开长任务。
		// 注意：command 字段不会可靠清空，需用互斥锁状态判断是否忙碌。
		if terminalBusyInspector(client) {
			continue
		}
		lastUsed := h.ShellClientLastUsed[client]
		if lastUsed == 0 {
			lastUsed = h.ShellClientStartMap[client]
		}
		if lastUsed > 0 && now-lastUsed >= timeoutSeconds {
			idleClients = append(idleClients, client)
			delete(h.ShellClientMap, shellClientId)
			delete(h.ShellClientStartMap, client)
			delete(h.ShellClientLastUsed, client)
		}
	}
	h.lock.Unlock()

	for _, client := range idleClients {
		client.CloseTerminal()
	}
}

// GetClient 获取一个可复用的交互式终端。
// 同一 poolKey 下最多保留 maxShellPoolSize 个终端，若全部忙碌则短暂等待后返回错误。
func (h *Shell) GetClient(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell,
	formatStream func(string) []string, promptKeywords []string, promptFunc func(string, io.WriteCloser, *ssh.Session) string) (*gsssh.SshTerminal, error) {
	start := time.Now()
	sshId := cast.ToString(sshConfig["id"])
	if sshId == "" {
		return nil, errors.New("ssh config error, GetClient " + cast.ToString(debug.Stack()))
	}
	poolKey := resolvePoolKey(sshConfig, shellClientId)
	gstool.FmtPrintlnLogTime(`[Shell.GetClient] enter ssh_id=%s shell_client_id=%s pool_key=%s`,
		sshId, shellClientId, poolKey)

	for {
		h.lock.Lock()
		pool := h.ShellClientPoolMap[poolKey]
		if len(pool) > 0 {
			next := h.ShellClientPoolNext[poolKey]
			idleIndex := findIdleClientIndex(pool, next)
			if idleIndex >= 0 {
				chooseClient := pool[idleIndex]
				h.ShellClientPoolNext[poolKey] = (idleIndex + 1) % len(pool)
				h.ShellClientLastUsed[chooseClient] = time.Now().Unix()
				// pooled client may be reused by a new request; always rebind SSE receiver
				bindReceiveHandler(chooseClient, sse, formatStream)
				h.lock.Unlock()
				gstool.FmtPrintlnLogTime(`[Shell.GetClient] return success ssh_id=%s pool_key=%s choose_index=%d pool_size=%d total_cost_ms=%d`,
					sshId, poolKey, idleIndex, len(pool), time.Since(start).Milliseconds())
				return chooseClient, nil
			}
		}
		needCreate := len(pool) < maxShellPoolSize
		poolSize := len(pool)
		h.lock.Unlock()

		gstool.FmtPrintlnLogTime(`[Shell.GetClient] pool status ssh_id=%s pool_key=%s pool_size=%d need_create=%v`,
			sshId, poolKey, poolSize, needCreate)

		if needCreate {
			gstool.FmtPrintlnLogTime(`[Shell.GetClient] create begin ssh_id=%s pool_key=%s`, sshId, poolKey)
			createStart := time.Now()
			newClient, err := h.createShellClient(sshConfig, poolKey, sse, formatStream, promptKeywords, promptFunc)
			gstool.FmtPrintlnLogTime(`[Shell.GetClient] create end ssh_id=%s pool_key=%s cost_ms=%d err=%v`,
				sshId, poolKey, time.Since(createStart).Milliseconds(), err)
			if err != nil {
				return nil, err
			}
			h.lock.Lock()
			pool = h.ShellClientPoolMap[poolKey]
			if len(pool) < maxShellPoolSize {
				h.ShellClientPoolMap[poolKey] = append(pool, newClient)
				h.ShellClientStartMap[newClient] = time.Now().Unix()
				h.ShellClientLastUsed[newClient] = time.Now().Unix()
				if h.ShellClientPoolNext[poolKey] >= len(h.ShellClientPoolMap[poolKey]) {
					h.ShellClientPoolNext[poolKey] = 0
				}
				h.lock.Unlock()
			} else {
				h.lock.Unlock()
				newClient.CloseTerminal()
			}
			continue
		}

		if time.Since(start) >= getClientBusyWaitTimeout {
			gstool.FmtPrintlnLogTime(`[Shell.GetClient] return error ssh_id=%s pool_key=%s err=all ssh clients are busy total_cost_ms=%d`,
				sshId, poolKey, time.Since(start).Milliseconds())
			return nil, errors.New("all ssh clients are busy")
		}
		time.Sleep(getClientBusyWaitInterval)
	}
}

// findIdleClientIndex 从连接池中按轮询顺序挑选一个当前空闲的终端。
func findIdleClientIndex(pool []*gsssh.SshTerminal, next int) int {
	if len(pool) == 0 {
		return -1
	}
	if next < 0 || next >= len(pool) {
		next = 0
	}
	for i := 0; i < len(pool); i++ {
		idx := (next + i) % len(pool)
		client := pool[idx]
		if client == nil {
			continue
		}
		if !terminalBusyInspector(client) {
			return idx
		}
	}
	return -1
}

// isTerminalBusy 根据 gsssh.SshTerminal 内部互斥锁判断是否正在执行命令。
// command 字段不会在每次执行后清空，不能用于 busy 判定。
func isTerminalBusy(gsShell *gsssh.SshTerminal) bool {
	if gsShell == nil {
		return true
	}
	defer func() {
		_ = recover()
	}()
	val := reflect.ValueOf(gsShell)
	if !val.IsValid() || val.Kind() != reflect.Ptr || val.IsNil() {
		return true
	}
	elem := val.Elem()
	if !elem.IsValid() || elem.Kind() != reflect.Struct {
		return true
	}
	field := elem.FieldByName("lockCommand")
	if !field.IsValid() || field.Type() != reflect.TypeOf(sync.Mutex{}) || !field.CanAddr() {
		return false
	}
	mtx := (*sync.Mutex)(unsafe.Pointer(field.UnsafeAddr()))
	if mtx.TryLock() {
		mtx.Unlock()
		return false
	}
	return true
}

// GetClientMarkdown 获取旧版一对一终端连接，供 markdown/sftp 等独占场景使用。
func (h *Shell) GetClientMarkdown(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell) (*gsssh.SshTerminal, error) {
	defer h.lock.Unlock()
	h.lock.Lock()
	sshId := cast.ToString(sshConfig["id"])
	if sshId == "" {
		return nil, errors.New("ssh config error, GetClientMarkdown " + cast.ToString(debug.Stack()))
	}
	if shell, ok := h.ShellClientMap[shellClientId]; ok && shell != nil {
		h.SetSse(shell, sse)
		h.ShellClientLastUsed[shell] = time.Now().Unix()
		return shell, nil
	}
	gsShell := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
		Name:     "",
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
	}))
	gsShell.SetPtyConfig(gsssh.PtyConfig{Echo: 1})
	gsShell.SetFuncBroken(func(msg string) {
		if canSendSse(sse) {
			sse.Send(" connection broken, will reconnect on next action: " + msg + "\n")
		}
		h.RmClient(shellClientId)
	})
	gsShell.SetMaxBufferSize(2 * 1024 * 1024)
	_, err := gsShell.RunCommandWait("pwd", 40*time.Second)
	if err != nil {
		return nil, err
	}
	h.SetSse(gsShell, sse)
	gsShell.SetAuthPromptKeywords([]string{"Username for", "Password for", "passphrase", "Passphrase"})
	gsShell.SetFuncAuthPrompt(func(prompt string, stdin io.WriteCloser, session *ssh.Session) string {
		if session != nil {
			_ = session.Signal(ssh.SIGINT)
			if strings.Contains(strings.ToLower(prompt), "git") {
				_, _ = stdin.Write([]byte("git credential-cache exit; unset GIT_ASKPASS\n"))
			}
			if canSendSse(sse) {
				sse.Send("\nmanual auth prompt detected, please configure credentials and retry\n")
			}
			return prompt
		}
		return prompt
	})

	h.ShellClientMap[shellClientId] = gsShell
	h.ShellClientStartMap[gsShell] = time.Now().Unix()
	h.ShellClientLastUsed[gsShell] = time.Now().Unix()
	return gsShell, nil
}

// SetSse 为指定终端绑定新的 SSE 接收器。
func (h *Shell) SetSse(gsShell *gsssh.SshTerminal, sse *p_sse.SseShell) {
	bindReceiveHandler(gsShell, sse, nil)
}

// GetSshOnce 返回一次性 SSH 客户端，适合不需要终端复用的场景。
func (h *Shell) GetSshOnce(sshConfig map[string]any) (*gsssh.SshOnce, error) {
	sshId := cast.ToString(sshConfig["id"])
	if sshId == "" {
		return nil, errors.New("ssh config error, GetClientMarkdown " + cast.ToString(debug.Stack()))
	}

	return gsssh.NewSshOnce(gsssh.NewSsh(&gsssh.SshConfig{
		Name:     "",
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
	})), nil
}

// Exist 判断给定 key 对应的池连接或独占连接是否存在。
func (h *Shell) Exist(uniqueKey string) bool {
	defer h.lock.Unlock()
	h.lock.Lock()
	if pool, ok := h.ShellClientPoolMap[uniqueKey]; ok && len(pool) > 0 {
		return true
	}
	poolKey := splitPoolKey(uniqueKey)
	if poolKey != "" && poolKey != uniqueKey {
		if pool, ok := h.ShellClientPoolMap[poolKey]; ok && len(pool) > 0 {
			return true
		}
	}
	if _, ok := h.ShellClientMap[uniqueKey]; ok {
		return true
	}
	return false
}

// RmClient 按 key 移除连接，兼容删除整组池连接或单个独占连接。
func (h *Shell) RmClient(uniqueKey string) {
	defer h.lock.Unlock()
	h.lock.Lock()
	poolKey := uniqueKey
	if _, ok := h.ShellClientPoolMap[poolKey]; !ok {
		if k := splitPoolKey(uniqueKey); k != "" {
			poolKey = k
		}
	}
	if pool, ok := h.ShellClientPoolMap[poolKey]; ok {
		for _, sshCli := range pool {
			if sshCli != nil {
				sshCli.CloseTerminal()
				delete(h.ShellClientStartMap, sshCli)
				delete(h.ShellClientLastUsed, sshCli)
			}
		}
		delete(h.ShellClientPoolMap, poolKey)
		delete(h.ShellClientPoolNext, poolKey)
	}
	if sshCli, ok := h.ShellClientMap[uniqueKey]; ok {
		sshCli.CloseTerminal()
		delete(h.ShellClientStartMap, sshCli)
		delete(h.ShellClientLastUsed, sshCli)
		delete(h.ShellClientMap, uniqueKey)
	}
}

// WalkShellList 遍历当前所有活跃终端，供外部统一查看或回收。
func (h *Shell) WalkShellList(businessFunc func(uniqueKey string, gsShell *gsssh.SshTerminal)) {
	defer h.lock.Unlock()
	h.lock.Lock()
	for uniqueKey, pool := range h.ShellClientPoolMap {
		for i, gsShell := range pool {
			if gsShell == nil {
				continue
			}
			businessFunc(uniqueKey+"#pool"+cast.ToString(i), gsShell)
		}
	}
	for uniqueKey, gsShell := range h.ShellClientMap {
		if gsShell == nil {
			continue
		}
		businessFunc(uniqueKey, gsShell)
	}
}

// ConnectionInfo 描述终端连接在 UI 中展示所需的元数据。
type ConnectionInfo struct {
	ShellClientId  string `json:"shell_client_id"`
	CurrentCommand string `json:"current_command"`
	Status         string `json:"status"`
	ConnectTime    string `json:"connect_time"`
	ConnectSeconds int64  `json:"connect_seconds"`
	Type           string `json:"type"`
}

// getTerminalCurrentCommand 读取底层终端对象当前记录的命令文本。
func getTerminalCurrentCommand(gsShell *gsssh.SshTerminal) string {
	if gsShell == nil {
		return ""
	}
	defer func() {
		_ = recover()
	}()
	val := reflect.ValueOf(gsShell)
	if !val.IsValid() || val.Kind() != reflect.Ptr || val.IsNil() {
		return ""
	}
	elem := val.Elem()
	if !elem.IsValid() || elem.Kind() != reflect.Struct {
		return ""
	}
	field := elem.FieldByName("command")
	if !field.IsValid() || field.Kind() != reflect.String {
		return ""
	}
	return strings.TrimSpace(field.String())
}

// getTerminalConnectMeta 生成连接建立时间及连接时长信息。
func (h *Shell) getTerminalConnectMeta(gsShell *gsssh.SshTerminal, now int64) (string, int64) {
	if gsShell == nil {
		return "", 0
	}
	startTime, ok := h.ShellClientStartMap[gsShell]
	if !ok || startTime <= 0 {
		return "", 0
	}
	return gstool.TimeUnixToString(time.Unix(startTime, 0), "Y-m-d H:i:s"), now - startTime
}

// GetConnections 汇总当前池连接和独占连接，供连接面板展示。
func (h *Shell) GetConnections() []ConnectionInfo {
	defer h.lock.Unlock()
	h.lock.Lock()

	connections := make([]ConnectionInfo, 0)
	now := time.Now().Unix()
	for shellClientId, pool := range h.ShellClientPoolMap {
		for i, shellClient := range pool {
			connectTime, connectSeconds := h.getTerminalConnectMeta(shellClient, now)
			info := ConnectionInfo{
				ShellClientId:  shellClientId + "#pool" + cast.ToString(i),
				CurrentCommand: getTerminalCurrentCommand(shellClient),
				Status:         "active",
				ConnectTime:    connectTime,
				ConnectSeconds: connectSeconds,
				Type:           "shell",
			}
			connections = append(connections, info)
		}
	}
	for shellClientId, shellClient := range h.ShellClientMap {
		connectTime, connectSeconds := h.getTerminalConnectMeta(shellClient, now)
		info := ConnectionInfo{
			ShellClientId:  shellClientId,
			CurrentCommand: getTerminalCurrentCommand(shellClient),
			Status:         "active",
			ConnectTime:    connectTime,
			ConnectSeconds: connectSeconds,
			Type:           "shell",
		}
		connections = append(connections, info)
	}

	return connections
}
