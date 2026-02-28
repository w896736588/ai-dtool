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
	ShellClientPoolMap  map[string][]*gsssh.SshTerminal
	ShellClientPoolNext map[string]int
	ShellClientStartMap map[*gsssh.SshTerminal]int64
	lock                sync.Mutex
	LogPath             string
	log                 *gstool.GsSlog
}

const maxShellPoolSize = 10

func canSendSse(sse *p_sse.SseShell) bool {
	return sse != nil && sse.Sse != nil
}

type receiveBinder interface {
	SetFuncReceiveMsg(func(string) string)
}

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

func bindReceiveHandler(target receiveBinder, sse *p_sse.SseShell, formatStream func(string) []string) {
	target.SetFuncReceiveMsg(makeReceiveHandler(sse, formatStream))
}

func splitPoolKey(uniqueKey string) string {
	if uniqueKey == "" {
		return ""
	}
	keyList := strings.SplitN(uniqueKey, "#", 2)
	return keyList[0]
}

func resolvePoolKey(sshConfig map[string]any, shellClientId string) string {
	if sshId := cast.ToString(sshConfig["id"]); sshId != "" {
		return sshId
	}
	if key := splitPoolKey(shellClientId); key != "" {
		return key
	}
	return shellClientId
}

func NewShell(logPath string) *Shell {
	log := gstool.NewSlog3(logPath, "shell")
	_ = log.CleanOldLogs(2)
	return &Shell{
		ShellClientMap:      make(map[string]*gsssh.SshTerminal),
		ShellClientPoolMap:  make(map[string][]*gsssh.SshTerminal),
		ShellClientPoolNext: make(map[string]int),
		ShellClientStartMap: make(map[*gsssh.SshTerminal]int64),
		log:                 log,
		LogPath:             logPath,
	}
}

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

func (h *Shell) createShellClient(sshConfig map[string]any, poolKey string, sse *p_sse.SseShell,
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
		h.lock.Lock()
		defer h.lock.Unlock()
		h.removeClientFromPoolLocked(poolKey, gsShell)
	})

	gsShell.SetPtyConfig(gsssh.PtyConfig{Echo: 1})
	gsShell.SetMaxBufferSize(2 * 1024 * 1024)
	_, err := gsShell.RunCommandWait("pwd", 40*time.Second)
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
	return gsShell, nil
}

// GetClient returns a pooled shell client. Pool size is capped per sshId.
func (h *Shell) GetClient(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell,
	formatStream func(string) []string, promptKeywords []string, promptFunc func(string, io.WriteCloser, *ssh.Session) string) (*gsssh.SshTerminal, error) {
	sshId := cast.ToString(sshConfig["id"])
	if sshId == "" {
		return nil, errors.New("ssh config error, GetClient " + cast.ToString(debug.Stack()))
	}
	poolKey := resolvePoolKey(sshConfig, shellClientId)

	h.lock.Lock()
	pool := h.ShellClientPoolMap[poolKey]
	needCreate := len(pool) < maxShellPoolSize
	h.lock.Unlock()

	if needCreate {
		newClient, err := h.createShellClient(sshConfig, poolKey, sse, formatStream, promptKeywords, promptFunc)
		if err != nil {
			return nil, err
		}
		h.lock.Lock()
		h.ShellClientPoolMap[poolKey] = append(h.ShellClientPoolMap[poolKey], newClient)
		h.ShellClientStartMap[newClient] = time.Now().Unix()
		pool = h.ShellClientPoolMap[poolKey]
		if h.ShellClientPoolNext[poolKey] >= len(pool) {
			h.ShellClientPoolNext[poolKey] = 0
		}
		h.lock.Unlock()
	}

	h.lock.Lock()
	defer h.lock.Unlock()
	pool = h.ShellClientPoolMap[poolKey]
	if len(pool) == 0 {
		return nil, errors.New("ssh pool is empty")
	}
	next := h.ShellClientPoolNext[poolKey]
	if next >= len(pool) {
		next = 0
	}
	chooseClient := pool[next]
	h.ShellClientPoolNext[poolKey] = (next + 1) % len(pool)
	// pooled client may be reused by a new request; always rebind SSE receiver
	bindReceiveHandler(chooseClient, sse, formatStream)
	return chooseClient, nil
}

// GetClientMarkdown keeps old one-to-one key behavior for markdown/sftp paths.
func (h *Shell) GetClientMarkdown(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell) (*gsssh.SshTerminal, error) {
	defer h.lock.Unlock()
	h.lock.Lock()
	sshId := cast.ToString(sshConfig["id"])
	if sshId == "" {
		return nil, errors.New("ssh config error, GetClientMarkdown " + cast.ToString(debug.Stack()))
	}
	if shell, ok := h.ShellClientMap[shellClientId]; ok && shell != nil {
		h.SetSse(shell, sse)
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
	return gsShell, nil
}

func (h *Shell) SetSse(gsShell *gsssh.SshTerminal, sse *p_sse.SseShell) {
	bindReceiveHandler(gsShell, sse, nil)
}

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

// RmClient removes both pool clients and markdown client by key.
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
			}
		}
		delete(h.ShellClientPoolMap, poolKey)
		delete(h.ShellClientPoolNext, poolKey)
	}
	if sshCli, ok := h.ShellClientMap[uniqueKey]; ok {
		sshCli.CloseTerminal()
		delete(h.ShellClientStartMap, sshCli)
		delete(h.ShellClientMap, uniqueKey)
	}
}

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

// ConnectionInfo contains shell connection metadata for UI.
type ConnectionInfo struct {
	ShellClientId  string `json:"shell_client_id"`
	CurrentCommand string `json:"current_command"`
	Status         string `json:"status"`
	ConnectTime    string `json:"connect_time"`
	ConnectSeconds int64  `json:"connect_seconds"`
	Type           string `json:"type"`
}

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
