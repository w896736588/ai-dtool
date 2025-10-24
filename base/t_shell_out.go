package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

/*  等待式输出 ssh 不重复使用，持续等待 ssh 返回结果 */

const ErrRegex = `(?i)\b(error|exception|fatal|panic|err|错误|报错|fail)\b`

// ShellOut 单个 ssh 会话
type ShellOut struct {
	Client        *gsssh.SshConfig
	errorList     []ErrorBlock        // 最终归档的错误块
	errorContent  string              // 错误检测内容
	remainContent string              // 保留的内容（最后 10 000 字符）
	seen          map[string]struct{} // 去重表：key=错误行
	errorRegex    *regexp.Regexp      // 错误行正则
	mu            sync.Mutex          // 保护 errorContent / errorList / seen
}

// ErrorBlock 错误块
type ErrorBlock struct {
	Lines      []string `json:"lines"`      // 最多 11 行
	ErrorLine  string   `json:"error_line"` // 用于去重
	LineNumber int      `json:"line_no"`    // 错误行在快照里的行号（从 0 起）
}

// TShellOut 管理多个 ShellOut
type TShellOut struct {
	ShellOutMap map[string]*ShellOut
	lock        sync.Mutex
	log         *gstool.GsSlog
}

// NewTShellOut 构造函数
func NewTShellOut() *TShellOut {
	log := gstool.NewSlog3(Component.Env.LogPath, `shell_wait`)
	_ = log.CleanOldLogs(2)
	return &TShellOut{
		ShellOutMap: make(map[string]*ShellOut),
		log:         log,
	}
}

// GetClient 获取或新建 ssh 客户端
func (h *TShellOut) GetClient(sshConfig map[string]any, shellClientId, sseClientId string,
	formatStream func(string) []string) (*ShellOut, bool, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	sshId := cast.ToString(sshConfig[`id`])
	if sshId == `` {
		return nil, false, errors.New(`ssh配置错误，GetClient ` + cast.ToString(debug.Stack()))
	}
	if shellOut, ok := h.ShellOutMap[shellClientId]; ok && shellOut != nil {
		return shellOut, true, nil
	}

	gsShell := gsssh.NewSshAuthPassword(
		cast.ToString(sshConfig["host"]),
		cast.ToString(sshConfig["port"]),
		cast.ToString(sshConfig["username"]),
		cast.ToString(sshConfig["password"]))
	gsShell.GsSlog = Component.GsLog

	// 断开回调
	gsShell.SetFuncBroken(func() {
		_ = Component.TSse.SendMsg(sseClientId, sseClientId+` 注意：连接已中断，下次动作时进行链接`+"\n", 0)
		h.RmClient(shellClientId)
	})
	gsShell.SetMaxRunSecond(40)

	if err := gsShell.ConnectAuthPassword(); err != nil {
		return nil, false, err
	}
	if _, err := gsShell.RunCommandWait(`pwd`); err != nil {
		return nil, false, err
	}

	// 新建 ShellOut
	shellOut := &ShellOut{
		Client:     gsShell,
		seen:       make(map[string]struct{}),
		errorRegex: regexp.MustCompile(ErrRegex),
	}
	h.SetReceiveMsg(shellOut, sseClientId, formatStream)
	h.ShellOutMap[shellClientId] = shellOut
	return shellOut, false, nil
}

// SetClientSseId 设置 sse 推送 & 错误检测
func (h *TShellOut) SetClientSseId(shellClientId, sshId, sseClientId, command string,
	formatStream func(string) []string) error {

	sshConfig, _ := Component.TSqlite.GetSshConfig(sshId)
	shellOut, exist, err := h.GetClient(sshConfig, shellClientId, sseClientId, formatStream)
	if err != nil {
		return err
	}
	h.SetReceiveMsg(shellOut, sseClientId, formatStream)
	if !exist {
		return shellOut.Client.RunCommand(command)
	}
	return nil
}

func (h *TShellOut) SetReceiveMsg(shellOut *ShellOut, sseClientId string, formatStream func(string) []string) {
	shellOut.Client.SetFuncStreamReceive(func(msg string) {
		// 1. 追加内容
		shellOut.mu.Lock()
		shellOut.remainContent += msg
		shellOut.remainContent = StringLastRunes(shellOut.remainContent, 10000)
		shellOut.errorContent += msg
		shellOut.mu.Unlock()

		// 2. 提取错误块（内部会清理已处理部分）
		shellOut.extractErrorBlocks()

		// 3. SSE 推送
		if formatStream != nil {
			msgList := formatStream(msg)
			_ = Component.TSse.SendMsgChunkList(sseClientId, msgList, 10)
		} else {
			_ = Component.TSse.SendMsgChunk(sseClientId, msg, _struct.Chunk{
				Type: define.ChunkNum,
				Num:  50,
			}, 10)
		}
	})
}

// 核心：正则提取错误块 + 清理已扫描内容
func (so *ShellOut) extractErrorBlocks() {
	so.mu.Lock()
	defer so.mu.Unlock()

	lines := strings.Split(so.errorContent, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !so.errorRegex.MatchString(line) {
			continue
		}
		// 去重
		if _, ok := so.seen[line]; ok {
			continue
		}
		so.seen[line] = struct{}{}

		// 取前后 5 行
		start := i - 5
		if start < 0 {
			start = 0
		}
		end := i + 5 + 1
		if end > len(lines) {
			end = len(lines)
		}
		block := ErrorBlock{
			Lines:      lines[start:end],
			ErrorLine:  line,
			LineNumber: i,
		}
		Component.GsLog.Debugf(`提取到错误 %s`, gstool.JsonFormat(block))
		so.errorList = append(so.errorList, block)

		// 清理：扔掉该错误行及之前的内容
		remainLines := lines[i+1:]
		so.errorContent = strings.Join(remainLines, "\n")
		if len(remainLines) > 0 {
			so.errorContent += "\n"
		}
		return // 一次只处理第一个错误，避免嵌套
	}
}

// SetErrorPattern 运行时替换正则
func (so *ShellOut) SetErrorPattern(expr string) error {
	re, err := regexp.Compile(expr)
	if err != nil {
		return err
	}
	so.mu.Lock()
	so.errorRegex = re
	so.mu.Unlock()
	return nil
}

// GetErrorList 并发安全获取已归档错误
func (so *ShellOut) GetErrorList() []ErrorBlock {
	so.mu.Lock()
	defer so.mu.Unlock()
	dst := make([]ErrorBlock, len(so.errorList))
	copy(dst, so.errorList)
	return dst
}

// 以下原有方法保持不变
func (h *TShellOut) Exist(uniqueKey string) bool {
	h.lock.Lock()
	defer h.lock.Unlock()
	_, ok := h.ShellOutMap[uniqueKey]
	return ok
}

func (h *TShellOut) RmClient(uniqueKey string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if sh, ok := h.ShellOutMap[uniqueKey]; ok {
		sh.Client.CloseTerminal()
	}
	delete(h.ShellOutMap, uniqueKey)
}

func (h *TShellOut) WalkShellList(businessFunc func(uniqueKey string, gsShell *gsssh.SshConfig)) {
	h.lock.Lock()
	defer h.lock.Unlock()
	for k, v := range h.ShellOutMap {
		businessFunc(k, v.Client)
	}
}

func StringLastRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > n {
		runes = runes[len(runes)-n:]
	}
	return string(runes)
}
