package p_shell

import (
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
	"golang.org/x/crypto/ssh"
)

type Shell struct {
	ShellClientMap      map[string]*gsssh.SshTerminal
	ShellClientStartMap map[*gsssh.SshTerminal]int64
	ShellClientLastUsed map[*gsssh.SshTerminal]int64
	lock                sync.Mutex
	LogPath             string
	log                 *gstool.GsSlog
}

const shellIdleTimeout = 30 * time.Minute
const shellIdleCleanTicker = 30 * time.Second

// getSshConnectTimeout 从 SSH 配置中读取连接超时时间，未配置时默认 3 秒。
func getSshConnectTimeout(sshConfig map[string]any) time.Duration {
	timeout := cast.ToInt(sshConfig["connect_timeout"])
	if timeout <= 0 {
		return 3 * time.Second
	}
	return time.Duration(timeout) * time.Second
}

// shellConnectMaxAttempts 控制 SSH 建连失败后的最大重试次数。
const shellConnectMaxAttempts = 3

// canSendSse 判断当前 SSE 是否仍然可用，避免向空连接发送消息。
func canSendSse(sse *p_sse.SseShell) bool {
	return sse != nil && sse.Sse != nil
}

type receiveBinder interface {
	SetFuncReceiveMsg(func(string) string)
}

// terminalOutputFilter 复用底层 SSH 终端的输出清洗能力，优先过滤 shell prompt。
type terminalOutputFilter interface {
	FilterEndTip(string) string
}

// makeReceiveHandler 根据格式化函数构造统一的终端消息转发逻辑。
func makeReceiveHandler(sse *p_sse.SseShell, outputFilter terminalOutputFilter, formatStream func(string) []string) func(string) string {
	return func(msg string) string {
		displayMsg := filterTerminalOutputForDisplay(outputFilter, msg)
		if displayMsg == `` {
			return msg
		}
		if formatStream != nil {
			msgList := formatStream(displayMsg)
			for _, line := range msgList {
				if line != `` && canSendSse(sse) {
					sse.Send(line)
				}
			}
		} else if canSendSse(sse) {
			sse.Send(displayMsg)
		}
		return msg
	}
}

// bindReceiveHandler 为终端重新绑定消息接收函数，便于连接复用时切换 SSE。
func bindReceiveHandler(target receiveBinder, sse *p_sse.SseShell, formatStream func(string) []string) {
	outputFilter, _ := target.(terminalOutputFilter)
	target.SetFuncReceiveMsg(makeReceiveHandler(sse, outputFilter, formatStream))
}

// NewShell 初始化 Shell 管理器，并启动空闲连接清理协程。
func NewShell(logPath string) *Shell {
	log := gstool.NewSlog3(logPath, "shell")
	_ = log.CleanOldLogs(2)
	shell := &Shell{
		ShellClientMap:      make(map[string]*gsssh.SshTerminal),
		ShellClientStartMap: make(map[*gsssh.SshTerminal]int64),
		ShellClientLastUsed: make(map[*gsssh.SshTerminal]int64),
		log:                 log,
		LogPath:             logPath,
	}
	// 启动空闲连接清理协程：超过 3 分钟未使用的连接将自动断开并移除。
	go shell.startIdleCleaner()
	return shell
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
	idleKeys := make([]string, 0)

	h.lock.Lock()
	for shellClientId, client := range h.ShellClientMap {
		if client == nil {
			delete(h.ShellClientMap, shellClientId)
			continue
		}
		lastUsed := h.ShellClientLastUsed[client]
		if lastUsed == 0 {
			lastUsed = h.ShellClientStartMap[client]
		}
		if lastUsed > 0 && now-lastUsed >= timeoutSeconds {
			idleClients = append(idleClients, client)
			idleKeys = append(idleKeys, shellClientId)
			delete(h.ShellClientMap, shellClientId)
			delete(h.ShellClientStartMap, client)
			delete(h.ShellClientLastUsed, client)
		}
	}
	h.lock.Unlock()

	for _, client := range idleClients {
		h.log.Infof("清理空闲连接")
		client.CloseTerminal()
	}
}

// GetClient 获取一个交互式终端连接。
// 每个 shellClientId 对应一个独立的连接，不复用连接池。
func (h *Shell) GetClient(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell,
	formatStream func(string) []string, promptKeywords []string, promptFunc func(string, io.WriteCloser, *ssh.Session) string) (*gsssh.SshTerminal, error) {
	sshId := cast.ToString(sshConfig["id"])
	host := cast.ToString(sshConfig["host"])
	port := cast.ToString(sshConfig["port"])
	gstool.FmtPrintlnLogTime(`[Shell.GetClient][01] 开始获取SSH客户端 shell_client_id=%s ssh_id=%s host=%s port=%s can_send_sse=%v`, shellClientId, sshId, host, port, canSendSse(sse))
	if sshId == "" {
		gstool.FmtPrintlnLogTime(`[Shell.GetClient][02] SSH配置错误，ssh_id为空 shell_client_id=%s host=%s port=%s`, shellClientId, host, port)
		return nil, errors.New("ssh config error, GetClient " + cast.ToString(debug.Stack()))
	}

	h.lock.Lock()
	if client, ok := h.ShellClientMap[shellClientId]; ok && client != nil {
		h.ShellClientLastUsed[client] = time.Now().Unix()
		bindReceiveHandler(client, sse, formatStream)
		h.lock.Unlock()
		gstool.FmtPrintlnLogTime(`[Shell.GetClient][03] 命中已有SSH客户端 shell_client_id=%s`, shellClientId)
		if canSendSse(sse) {
			sse.Send(" [ssh] 复用已有连接: " + shellClientId + "\n")
		}
		return client, nil
	}
	h.lock.Unlock()

	gstool.FmtPrintlnLogTime(`[Shell.GetClient][04] 未命中已有SSH客户端，准备创建 shell_client_id=%s`, shellClientId)
	if canSendSse(sse) {
		sse.Send(" [ssh] 创建新连接: " + shellClientId + "\n")
	}
	client, err := h.createShellClient(sshConfig, shellClientId, sse, formatStream, promptKeywords, promptFunc)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[Shell.GetClient][05] 创建SSH客户端失败 shell_client_id=%s err=%s`, shellClientId, err.Error())
		return nil, err
	}

	h.lock.Lock()
	h.ShellClientMap[shellClientId] = client
	h.ShellClientStartMap[client] = time.Now().Unix()
	h.ShellClientLastUsed[client] = time.Now().Unix()
	activeCount := len(h.ShellClientMap)
	h.lock.Unlock()

	gstool.FmtPrintlnLogTime(`[Shell.GetClient][06] SSH客户端已加入连接表 shell_client_id=%s active_count=%d`, shellClientId, activeCount)
	return client, nil
}

// createShellClient 创建一个新的交互式 SSH 终端，并绑定断线与消息转发处理。
func (h *Shell) createShellClient(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell,
	formatStream func(string) []string, promptKeywords []string, promptFunc func(string, io.WriteCloser, *ssh.Session) string) (*gsssh.SshTerminal, error) {
	var lastErr error
	host := cast.ToString(sshConfig["host"])
	port := cast.ToString(sshConfig["port"])
	username := cast.ToString(sshConfig["username"])
	for attempt := 1; attempt <= shellConnectMaxAttempts; attempt++ {
		gstool.FmtPrintlnLogTime(`[Shell.createShellClient][01] 开始创建SSH终端 shell_client_id=%s host=%s port=%s username=%s attempt=%d/%d`, shellClientId, host, port, username, attempt, shellConnectMaxAttempts)
		gsShell := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
			Name:     "",
			Host:     host,
			Port:     port,
			UserName: username,
			Password: cast.ToString(sshConfig["password"]),
		}))

		gsShell.SetPtyConfig(gsssh.PtyConfig{Echo: 0})
		gsShell.SetMaxBufferSize(2 * 1024 * 1024)
		connectTimeout := getSshConnectTimeout(sshConfig)
		gstool.FmtPrintlnLogTime(`[Shell.createShellClient][02] 准备执行SSH连接后命令 shell_client_id=%s cmds=%q timeout=%s attempt=%d/%d`, shellClientId, cast.ToString(sshConfig["post_connect_cmds"]), connectTimeout.String(), attempt, shellConnectMaxAttempts)
		if canSendSse(sse) {
			sse.Send(fmt.Sprintf(" [ssh] 正在建立SSH连接 %s:%s（第%d/%d次）\n",
				host, port, attempt, shellConnectMaxAttempts))
		}
		pwdResult, err := runPostConnectCmds(gsShell, sshConfig)
		if err != nil {
			gstool.FmtPrintlnLogTime(`[Shell.createShellClient][03] SSH连接后命令执行失败 shell_client_id=%s attempt=%d/%d result_len=%d result=%q err=%s`, shellClientId, attempt, shellConnectMaxAttempts, len(pwdResult), pwdResult, err.Error())
			lastErr = err
			if canSendSse(sse) {
				sse.Send(fmt.Sprintf(" [ssh] 连接建立失败（第%d/%d次）: %s\n", attempt, shellConnectMaxAttempts, err.Error()))
			}
			gsShell.CloseTerminal()
			continue
		}
		gstool.FmtPrintlnLogTime(`[Shell.createShellClient][04] SSH连接后命令执行成功 shell_client_id=%s attempt=%d/%d result_len=%d result=%q`, shellClientId, attempt, shellConnectMaxAttempts, len(pwdResult), pwdResult)

		gsShell.SetFuncBroken(func(msg string) {
			gstool.FmtPrintlnLogTime(`[Shell.createShellClient][05] SSH连接断开 shell_client_id=%s msg=%s`, shellClientId, msg)
			if canSendSse(sse) {
				sse.Send(" connection broken, will reconnect on next action: " + msg + "\n")
			}
			h.RmClient(shellClientId)
		})
		gstool.FmtPrintlnLogTime(`[Shell.createShellClient][06] SSH终端初始化完成 shell_client_id=%s`, shellClientId)
		if canSendSse(sse) {
			sse.Send(" [ssh] SSH连接建立成功\n")
		}

		bindReceiveHandler(gsShell, sse, formatStream)

		if len(promptKeywords) == 0 {
			promptKeywords = []string{"Username for", "Password for", "passphrase", "Passphrase"}
		}
		gsShell.SetAuthPromptKeywords(promptKeywords)
		if promptFunc != nil {
			gsShell.SetFuncAuthPrompt(promptFunc)
		}
		return gsShell, nil
	}
	if lastErr == nil {
		lastErr = errors.New("ssh connect failed")
	}
	gstool.FmtPrintlnLogTime(`[Shell.createShellClient][07] SSH终端创建最终失败 shell_client_id=%s host=%s port=%s err=%s`, shellClientId, host, port, lastErr.Error())
	if canSendSse(sse) {
		sse.Send(" [ssh] 连接建立失败: " + lastErr.Error() + "\n")
	}
	return nil, lastErr
}

// GetClientMarkdown 获取旧版一对一终端连接，供 markdown/sftp 等独占场景使用。
// 现在与 GetClient 行为一致。
func (h *Shell) GetClientMarkdown(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell) (*gsssh.SshTerminal, error) {
	return h.GetClient(sshConfig, shellClientId, sse, nil, nil, nil)
}

// SetSse 为指定终端绑定新的 SSE 接收器。
func (h *Shell) SetSse(gsShell *gsssh.SshTerminal, sse *p_sse.SseShell) {
	bindReceiveHandler(gsShell, sse, nil)
}

// filterTerminalOutputForDisplay 过滤交互式终端提示符，避免 SSE 输出混入 user@host:path$。
func filterTerminalOutputForDisplay(outputFilter terminalOutputFilter, msg string) string {
	if outputFilter == nil || msg == `` {
		return msg
	}
	return outputFilter.FilterEndTip(msg)
}

// GetSshOnce 返回一次性 SSH 客户端，适合不需要终端复用的场景。
func (h *Shell) GetSshOnce(sshConfig map[string]any) (*gsssh.SshOnce, error) {
	sshId := cast.ToString(sshConfig["id"])
	if sshId == "" {
		return nil, errors.New("ssh config error, GetSshOnce " + cast.ToString(debug.Stack()))
	}

	return gsssh.NewSshOnce(gsssh.NewSsh(&gsssh.SshConfig{
		Name:     "",
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
	})), nil
}

// Exist 判断给定 key 对应的连接是否存在。
func (h *Shell) Exist(uniqueKey string) bool {
	defer h.lock.Unlock()
	h.lock.Lock()
	if _, ok := h.ShellClientMap[uniqueKey]; ok {
		return true
	}
	return false
}

// RmClient 按 key 移除连接。
func (h *Shell) RmClient(uniqueKey string) {
	defer h.lock.Unlock()
	h.lock.Lock()
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

// GetConnections 汇总当前所有连接，供连接面板展示。
func (h *Shell) GetConnections() []ConnectionInfo {
	defer h.lock.Unlock()
	h.lock.Lock()

	connections := make([]ConnectionInfo, 0)
	now := time.Now().Unix()
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

// getCmdTimeout 从 SSH 配置中读取命令执行超时，默认 3 秒。
func getCmdTimeout(sshConfig map[string]any) time.Duration {
	timeout := cast.ToInt(sshConfig["cmd_timeout"])
	if timeout <= 0 {
		return 3 * time.Second
	}
	return time.Duration(timeout) * time.Second
}

// runPostConnectCmds 连接成功后执行配置的命令列表，每行一条；空则回退 pwd；遇错停止。
func runPostConnectCmds(gsShell *gsssh.SshTerminal, sshConfig map[string]any) (string, error) {
	cmdsRaw := strings.TrimSpace(cast.ToString(sshConfig["post_connect_cmds"]))
	var cmds []string
	if cmdsRaw == "" {
		cmds = []string{"pwd"}
	} else {
		for _, line := range strings.Split(cmdsRaw, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				cmds = append(cmds, line)
			}
		}
	}
	if len(cmds) == 0 {
		cmds = []string{"pwd"}
	}
	timeout := getCmdTimeout(sshConfig)
	var lastResult string
	for _, cmd := range cmds {
		result, err := gsShell.RunCommandWait(cmd, timeout)
		if err != nil {
			return result, err
		}
		lastResult = result
	}
	return lastResult, nil
}
