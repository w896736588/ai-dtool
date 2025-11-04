package base

import (
	"dev_tool/base/define"
	"errors"
	"fmt"
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
const MaxLength = 500000

/*  等待式输出 ssh 不重复使用，持续等待 ssh 返回结果 */

// ShellOut 单个 ssh 会话
type ShellOut struct {
	Client           *gsssh.SshConfig
	sseClientId      string
	errorList        []ErrorBlock        // 最终归档的错误块
	errorContent     string              // 错误检测内容
	remainContent    string              // 保留的内容(替换后的)
	sourceContent    string              // 原本内容
	seen             map[string]struct{} // 去重表：key=错误行
	errorRegex       *regexp.Regexp      // 错误行正则
	mu               sync.Mutex          // 保护 errorContent / errorList / seen
	regexFiltersTips map[string]int      //过滤正则数量统计
	startTime        int64               //启动时间
	groupId          int                 //分组id
	extractTimer     *time.Timer         // 延迟提取计时器
}

// ErrorBlock 错误块
type ErrorBlock struct {
	Lines      []string `json:"lines"`      // 最多 11 行
	ErrorLine  string   `json:"error_line"` // 用于去重
	LineNumber int      `json:"line_no"`    // 错误行在快照里的行号（从 0 起）
}

// TShellOut 管理多个 ShellOut
type TShellOut struct {
	ShellOutMap       map[string]*ShellOut
	lock              sync.Mutex
	log               *gstool.GsSlog
	GroupRegexFilters map[int][]string
	GroupConfigLock   sync.Mutex
}

// NewTShellOut 构造函数
func NewTShellOut() *TShellOut {
	log := gstool.NewSlog3(Component.Env.LogPath, `shell_wait`)
	_ = log.CleanOldLogs(2)
	shellOut := &TShellOut{
		ShellOutMap:       make(map[string]*ShellOut),
		log:               log,
		GroupRegexFilters: make(map[int][]string),
	}
	return shellOut
}

func (h *TShellOut) InitGroupConfigs() {
	h.GroupConfigLock.Lock()
	defer h.GroupConfigLock.Unlock()
	all, allErr := Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeShellOut,
	}).All()
	if allErr != nil {
		gstool.FmtPrintlnLogTime(`获取ssh配置错误 %s`, allErr.Error())
		return
	}
	for _, item := range all {
		groupId := cast.ToInt(item[`id`])
		extra1 := cast.ToString(item[`extra_1`])
		h.GroupRegexFilters[groupId] = strings.Split(extra1, "\n")
	}
	gstool.FmtPrintlnLogTime(`初始化正则 %s`, gstool.JsonEncode(h.GroupRegexFilters))
}

// GetClient 获取或新建 ssh 客户端
func (h *TShellOut) GetClient(sshConfig map[string]any, shellClientId, sseClientId string, groupId int,
	formatStream func(string) []string) (*ShellOut, bool, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	sshId := cast.ToString(sshConfig[`id`])
	if sshId == `` {
		return nil, false, errors.New(`ssh配置错误，GetClient ` + cast.ToString(debug.Stack()))
	}
	if shellOut, ok := h.ShellOutMap[shellClientId]; ok && shellOut != nil {
		shellOut.groupId = groupId
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
		_ = Component.TSse.SendMsg(sseClientId, define.SseContentTypeMsg, sseClientId+` 注意：连接已中断，下次动作时进行链接`+"\n", 0)
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
		Client:           gsShell,
		seen:             make(map[string]struct{}),
		sseClientId:      sseClientId,
		errorRegex:       regexp.MustCompile(ErrRegex),
		regexFiltersTips: map[string]int{},
		startTime:        time.Now().Unix(),
		groupId:          groupId,
	}
	h.SetReceiveMsg(shellOut, formatStream)
	h.ShellOutMap[shellClientId] = shellOut
	return shellOut, false, nil
}

// SetClientSseId 设置 sse 推送 & 错误检测
func (h *TShellOut) SetClientSseId(shellClientId, sshId, sseClientId, command string, groupId int,
	formatStream func(string) []string) error {

	sshConfig, _ := Component.TSqlite.GetSshConfig(sshId)
	shellOut, exist, err := h.GetClient(sshConfig, shellClientId, sseClientId, groupId, formatStream)
	if err != nil {
		return err
	}
	shellOut.sseClientId = sseClientId
	h.SetReceiveMsg(shellOut, formatStream)
	if !exist {
		go func() {
			err := shellOut.Client.RunCommand(command)
			if err != nil {
				fmt.Println(fmt.Sprintf(`执行错误 %s`, err.Error()))
			}
		}()
		return nil
	} else {
		//最多展示10000个字符
		h.SendMsg(shellOut, StringLastRunes(shellOut.remainContent, 10000))
		h.SendErrList(shellOut)
	}
	return nil
}

func (h *TShellOut) GetRegexFilters(shellOut *ShellOut) []string {
	h.GroupConfigLock.Lock()
	defer h.GroupConfigLock.Unlock()
	return h.GroupRegexFilters[shellOut.groupId]
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

func (h *TShellOut) Delete(shellClientId string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	shellOut := h.ShellOutMap[shellClientId]
	if shellOut == nil {
		return
	}
	shellOut.Client.CloseTerminal()
	delete(h.ShellOutMap, shellClientId)

}

func (h *TShellOut) SetReceiveMsg(shellOut *ShellOut, formatStream func(string) []string) {
	shellOut.Client.SetFuncStreamReceive(func(msg string) {
		//原内容处理
		shellOut.sourceContent += msg
		shellOut.sourceContent = StringLastRunes(shellOut.sourceContent, MaxLength*2)
		//保留内容处理
		shellOut.remainContent += msg
		shellOut.remainContent = StringLastRunes(shellOut.remainContent, MaxLength)
		//过滤内容处理
		boolFilter := h.RegexFilter(shellOut, msg)
		if boolFilter {
			return
		}
		//错误检测
		h.CheckError(shellOut, msg)
		//推送
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

func (h *TShellOut) CheckError(shellOut *ShellOut, msg string) {
	if !shellOut.errorRegex.MatchString(msg) {
		return
	}
	// 去重
	if _, ok := shellOut.seen[msg]; ok {
		return
	}
	shellOut.seen[msg] = struct{}{}

	block := ErrorBlock{
		Lines:     []string{},
		ErrorLine: msg,
	}
	shellOut.errorList = append(shellOut.errorList, block)
	h.SendErr(shellOut, block)
}

func (h *TShellOut) RegexFilter(shellOut *ShellOut, msg string) bool {
	boolFilter := false
	for _, regexFilter := range h.GetRegexFilters(shellOut) {
		if strings.TrimSpace(regexFilter) == `` {
			continue
		}
		name := ``
		regexParams := strings.Split(regexFilter, `#`)
		if len(regexParams) == 2 {
			regexFilter = regexParams[1]
			name = regexParams[0]
		}
		var re = regexp.MustCompile(regexFilter)
		if re.MatchString(msg) {
			boolFilter = true
			if gstool.MapKeyExist(&shellOut.regexFiltersTips, regexFilter) {
				shellOut.regexFiltersTips[regexFilter] += 1
			} else {
				shellOut.regexFiltersTips[regexFilter] = 1
			}
			if shellOut.regexFiltersTips[regexFilter]%10 == 0 {
				if name != `` {
					h.SendMsg(shellOut, fmt.Sprintf(`%s 过滤输出：%s,已过滤：%d次`+"\n", gstool.TimeNowUnixToString(``), name, shellOut.regexFiltersTips[regexFilter]))
				} else {
					h.SendMsg(shellOut, fmt.Sprintf(`%s 过滤输出：%s,已过滤：%d次`+"\n", gstool.TimeNowUnixToString(``), regexFilter, shellOut.regexFiltersTips[regexFilter]))
				}
			}
			break
		}
	}
	return boolFilter
}

func (h *TShellOut) SendMsg(shellOut *ShellOut, msg string) {
	msg = strings.Replace(msg, `\n`, "\n", -1)
	_ = Component.TSse.SendMsg(shellOut.sseClientId, define.SseContentTypeMsg, msg, 0)
}

func (h *TShellOut) SendEvent(shellOut *ShellOut, eventType, msg string) {
	msg = strings.Replace(msg, `\n`, "\n", -1)
	_ = Component.TSse.SendMsg(shellOut.sseClientId, eventType, msg, 0)
}

func (h *TShellOut) SendErrList(shellOut *ShellOut) {
	_ = Component.TSse.SendMsg(shellOut.sseClientId, define.SseContentTypeErrorList, shellOut.errorList, 0)
}

func (h *TShellOut) SendErr(shellOut *ShellOut, err ErrorBlock) {
	_ = Component.TSse.SendMsg(shellOut.sseClientId, define.SseContentTypeError, err, 0)
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
