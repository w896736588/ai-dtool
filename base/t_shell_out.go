package base

import (
	"errors"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

const ErrRegex = `(?i)\b(error|exception|fatal|panic|err|错误|报错|fail)\b`

/*  等待式输出 ssh 不重复使用，持续等待 ssh 返回结果 */

// ShellOut 单个 ssh 会话
type ShellOut struct {
	Client        *gsssh.SshConfig
	sseClientId   string
	errorList     []ErrorBlock        // 最终归档的错误块
	errorContent  string              // 错误检测内容
	remainContent string              // 保留的内容（最后 10 000 字符）
	seen          map[string]struct{} // 去重表：key=错误行
	errorRegex    *regexp.Regexp      // 错误行正则
	mu            sync.Mutex          // 保护 errorContent / errorList / seen

	extractTimer *time.Timer // 延迟提取计时器
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
		Client:      gsShell,
		seen:        make(map[string]struct{}),
		sseClientId: sseClientId,
		errorRegex:  regexp.MustCompile(ErrRegex),
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
	shellOut.sseClientId = sseClientId
	h.SetReceiveMsg(shellOut, sseClientId, formatStream)
	if !exist {
		return shellOut.Client.RunCommand(command)
	} else {
		h.SendMsg(shellOut, shellOut.remainContent)
		h.SendErrList(shellOut)
	}
	return nil
}

func (h *TShellOut) CleanErrors(shellClientId string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	shellOut := h.ShellOutMap[shellClientId]
	if shellOut == nil {
		return
	}
	shellOut.errorList = make([]ErrorBlock, 0)
}

func (h *TShellOut) SetReceiveMsg(shellOut *ShellOut, sseClientId string, formatStream func(string) []string) {
	shellOut.Client.SetFuncStreamReceive(func(msg string) {
		// 1. 追加内容
		shellOut.mu.Lock()
		shellOut.remainContent += msg
		shellOut.remainContent = StringLastRunes(shellOut.remainContent, 50000)
		shellOut.errorContent += msg
		shellOut.mu.Unlock()

		// 2. 重置/启动 1 秒延迟提取器
		h.resetExtractTimer(shellOut)

		// 3. SSE 推送
		if formatStream != nil {
			msgList := formatStream(msg)
			for _, msg := range msgList {
				h.SendMsg(shellOut, msg)
			}
		} else {
			h.SendMsg(shellOut, msg)
		}
	})
}

// resetExtractTimer 保证 1 秒内没新消息才真正去提取
func (h *TShellOut) resetExtractTimer(so *ShellOut) {
	so.mu.Lock()
	defer so.mu.Unlock()

	// 如果已经有一个计时器，先停掉
	if so.extractTimer != nil {
		so.extractTimer.Stop()
	}
	// 重新计时 1 秒
	so.extractTimer = time.AfterFunc(time.Second, func() {
		h.doExtractErrorBlocks(so)
	})
}

func (h *TShellOut) SendMsg(shellOut *ShellOut, msg string) {
	send := map[string]any{
		`type`: `msg`,
		`data`: msg,
	}
	_ = Component.TSse.SendMsg(shellOut.sseClientId, gstool.JsonEncode(send)+"\n", 0)
}

func (h *TShellOut) SendErrList(shellOut *ShellOut) {
	send := map[string]any{
		`type`: `error_list`,
		`data`: shellOut.errorList,
	}
	_ = Component.TSse.SendMsg(shellOut.sseClientId, gstool.JsonEncode(send)+"\n", 0)
}

func (h *TShellOut) SendErr(shellOut *ShellOut, err ErrorBlock) {
	send := map[string]any{
		`type`: `error`,
		`data`: err,
	}
	_ = Component.TSse.SendMsg(shellOut.sseClientId, gstool.JsonEncode(send)+"\n", 0)
}

// 真正的提取逻辑，运行在计时器回调里，无锁竞争
func (h *TShellOut) doExtractErrorBlocks(shellOut *ShellOut) {
	shellOut.mu.Lock()
	defer shellOut.mu.Unlock()

	lines := strings.Split(shellOut.errorContent, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !shellOut.errorRegex.MatchString(line) {
			continue
		}
		// 去重
		if _, ok := shellOut.seen[line]; ok {
			continue
		}
		shellOut.seen[line] = struct{}{}

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
		shellOut.errorList = append(shellOut.errorList, block)
		h.SendErr(shellOut, block)

		// 清理：扔掉该错误行及之前的内容
		remainLines := lines[i+1:]
		shellOut.errorContent = strings.Join(remainLines, "\n")
		if len(remainLines) > 0 {
			shellOut.errorContent += "\n"
		}
		return // 一次只处理第一个错误，避免嵌套
	}
}

func (h *TShellOut) RmClient(uniqueKey string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if sh, ok := h.ShellOutMap[uniqueKey]; ok {
		// 停掉可能存在的计时器，防止闭包内访问已释放对象
		sh.mu.Lock()
		if sh.extractTimer != nil {
			sh.extractTimer.Stop()
		}
		sh.mu.Unlock()

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
