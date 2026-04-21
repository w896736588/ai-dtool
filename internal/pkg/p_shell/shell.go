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

const shellIdleTimeout = 3 * time.Minute
const shellIdleCleanTicker = 30 * time.Second

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
	if sshId == "" {
		return nil, errors.New("ssh config error, GetClient " + cast.ToString(debug.Stack()))
	}

	h.lock.Lock()
	if client, ok := h.ShellClientMap[shellClientId]; ok && client != nil {
		h.ShellClientLastUsed[client] = time.Now().Unix()
		bindReceiveHandler(client, sse, formatStream)
		h.lock.Unlock()
		if canSendSse(sse) {
			sse.Send(" [ssh] 复用已有连接: " + shellClientId + "\n")
		}
		return client, nil
	}
	h.lock.Unlock()

	if canSendSse(sse) {
		sse.Send(" [ssh] 创建新连接: " + shellClientId + "\n")
	}
	client, err := h.createShellClient(sshConfig, shellClientId, sse, formatStream, promptKeywords, promptFunc)
	if err != nil {
		return nil, err
	}

	h.lock.Lock()
	h.ShellClientMap[shellClientId] = client
	h.ShellClientStartMap[client] = time.Now().Unix()
	h.ShellClientLastUsed[client] = time.Now().Unix()
	h.lock.Unlock()

	return client, nil
}

// createShellClient 创建一个新的交互式 SSH 终端，并绑定断线与消息转发处理。
func (h *Shell) createShellClient(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell,
	formatStream func(string) []string, promptKeywords []string, promptFunc func(string, io.WriteCloser, *ssh.Session) string) (*gsssh.SshTerminal, error) {
	gsShell := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
		Name:     "",
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
	}))

	gsShell.SetFuncBroken(func(msg string) {
		if canSendSse(sse) {
			sse.Send(" connection broken, will reconnect on next action: " + msg + "\n")
		}
		h.RmClient(shellClientId)
	})

	gsShell.SetPtyConfig(gsssh.PtyConfig{Echo: 1})
	gsShell.SetMaxBufferSize(2 * 1024 * 1024)
	if canSendSse(sse) {
		sse.Send(" [ssh] 正在建立SSH连接 " + cast.ToString(sshConfig["host"]) + ":" + cast.ToString(sshConfig["port"]) + "\n")
	}
	_, err := gsShell.RunCommandWait("pwd", 40*time.Second)
	if err != nil {
		if canSendSse(sse) {
			sse.Send(" [ssh] 连接建立失败: " + err.Error() + "\n")
		}
		return nil, err
	}
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

// GetClientMarkdown 获取旧版一对一终端连接，供 markdown/sftp 等独占场景使用。
// 现在与 GetClient 行为一致。
func (h *Shell) GetClientMarkdown(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell) (*gsssh.SshTerminal, error) {
	return h.GetClient(sshConfig, shellClientId, sse, nil, nil, nil)
}

// SetSse 为指定终端绑定新的 SSE 接收器。
func (h *Shell) SetSse(gsShell *gsssh.SshTerminal, sse *p_sse.SseShell) {
	bindReceiveHandler(gsShell, sse, nil)
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
