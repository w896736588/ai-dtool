package common

import (
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_define"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsssh"
	"github.com/w896736588/go-tool/gstool"
)

const MaxSourceLength = 100000 //原始内容最多保留多少行 用于搜索
const MaxRemainLength = 10000  //过滤后内容最多保留多少行
const MaxSendLength = 1000     //刷新页面后最多发送给前端多少行

const ExistTip = `command is exit`

/*  等待式输出 ssh 不重复使用，持续等待 ssh 返回结果 */

// ShellOut 单个 ssh 会话
type ShellOut struct {
	ShellClientId      string
	Client             *gsssh.SshTerminal
	Sse                *p_sse.SseShell
	errorList          []ErrorBlock // 最终归档的错误块
	alertEvents        []ShellOutAlertEvent
	remainContents     []string       // 保留的内容(替换后的)
	sourceContents     []SourceLine   // 原本内容（带行号）
	searchReadContents map[string]any //已经搜索过的内容
	regexFiltersTips   map[string]int //过滤正则数量统计
	startTime          int64          //启动时间
	groupId            int            //分组id
	ruleSetId          int            //规则集id
	ruleItems          []ShellOutRuleItem
	breakTimer         *time.Ticker
	lastReceiveTime    int64
	pendingMsg         string       // Buffer for incomplete log lines
	filterKeywords     []string     // SSE推送过滤关键词（空格分割，需全部包含才推送）
	filterLock         sync.RWMutex // filterKeywords 的读写锁
}

// ErrorBlock 错误块
type ErrorBlock struct {
	ErrorLine  string `json:"error_line"`
	LineNumber int64  `json:"line_number"` // 错误行的行号
	Time       string `json:"time"`
	RuleName   string `json:"rule_name"`
	Level      string `json:"level"`
	Category   string `json:"category"`
}

// LineContext 行上下文（带行号）
type LineContext struct {
	LineNumber int64  `json:"line_number"` // 行号
	Content    string `json:"content"`     // 内容
}

// SourceLine 原始内容行（带行号）
type SourceLine struct {
	LineNumber int64  // 行号
	Content    string // 内容
}

// TShellOut 管理多个 ShellOut
type TShellOut struct {
	ShellOutMap map[string]*ShellOut
	lock        sync.Mutex
	log         *gstool.GsSlog
}

// NewTShellOut 构造函数
func NewTShellOut(logPath string) *TShellOut {
	log := gstool.NewSlog3(logPath, `shell_wait`)
	_ = log.CleanOldLogs(2)
	shellOut := &TShellOut{
		ShellOutMap: make(map[string]*ShellOut),
		log:         log,
	}
	return shellOut
}

func (h *TShellOut) InitGroupConfigs() {
	// 壳分组不再承载日志规则，新架构下这里保留空实现以避免初始化链路报错。
	// Shell-out groups no longer own log rules, so keep this as a no-op during the migration window.
}

// GetClient 获取或新建 ssh 客户端
func (h *TShellOut) GetClient(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell, groupId int, ruleSetID int,
	formatStream func(string) []string) (*ShellOut, bool, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	sshId := cast.ToString(sshConfig[`id`])
	if sshId == `` {
		return nil, false, errors.New(`ssh配置错误，GetClient ` + cast.ToString(debug.Stack()))
	}
	if shellOut, ok := h.ShellOutMap[shellClientId]; ok && shellOut != nil {
		if sse != nil {
			sse.Send(" [shell_out] 复用已有连接: " + shellClientId + "\n")
		}
		return shellOut, true, nil
	}
	if sse != nil {
		sse.Send(" [shell_out] 创建新连接: " + shellClientId + " → " + cast.ToString(sshConfig["host"]) + ":" + cast.ToString(sshConfig["port"]) + "\n")
	}
	gsShell := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
		Name:     "",
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
	}))
	// 断开回调
	gsShell.SetFuncBroken(func(msg string) {
		sse.Send(` 注意：连接已中断，下次动作时进行链接,` + msg + "\n")
		h.RmClient(shellClientId)
	})
	gsShell.SetMaxBufferSize(2 * 1024 * 1024) //最大允许2M的输出
	gsShell.SetPtyConfig(gsssh.PtyConfig{
		Width: 1000,
		Echo:  1,
	})

	if err := runShellOutPostConnectCmds(gsShell, sshConfig); err != nil {
		gstool.FmtPrintlnLogTime(`shell out 连接后命令执行失败 %s`, err.Error())
		if sse != nil {
			sse.Send(" [shell_out] 连接建立失败: " + err.Error() + "\n")
		}
		return nil, false, err
	}
	if sse != nil {
		sse.Send(" [shell_out] SSH连接建立成功\n")
	}

	// 新建 ShellOut
	shellOut := &ShellOut{
		ShellClientId:      shellClientId,
		Client:             gsShell,
		Sse:                sse,
		regexFiltersTips:   map[string]int{},
		startTime:          time.Now().Unix(),
		groupId:            groupId,
		ruleSetId:          ruleSetID,
		ruleItems:          []ShellOutRuleItem{},
		alertEvents:        []ShellOutAlertEvent{},
		errorList:          make([]ErrorBlock, 0),
		remainContents:     make([]string, 0),
		sourceContents:     make([]SourceLine, 0),
		searchReadContents: map[string]any{},
		breakTimer:         time.NewTicker(time.Second * 30),
		lastReceiveTime:    time.Now().Unix(),
		pendingMsg:         ``,
	}
	ruleItems, loadRuleErr := h.LoadRuleItems(ruleSetID)
	if loadRuleErr != nil {
		return nil, false, loadRuleErr
	}
	shellOut.ruleItems = ruleItems
	h.SetReceiveMsg(shellOut, formatStream)
	h.ShellOutMap[shellClientId] = shellOut
	return shellOut, false, nil
}

// SetClientSseId 设置 sse 推送 & 错误检测
func (h *TShellOut) SetClientSseId(shellClientId, sshId string, sse *p_sse.SseShell, command string, groupId int, ruleSetID int,
	formatStream func(string) []string) error {

	sshConfig, _ := DbMain.GetSshConfig(sshId)
	shellOut, exist, err := h.GetClient(sshConfig, shellClientId, sse, groupId, ruleSetID, formatStream)
	if err != nil {
		return err
	}
	shellOut.groupId = groupId
	shellOut.ruleSetId = ruleSetID
	shellOut.Sse = sse
	ruleItems, loadRuleErr := h.LoadRuleItems(ruleSetID)
	if loadRuleErr != nil {
		return loadRuleErr
	}
	shellOut.ruleItems = ruleItems
	h.SetReceiveMsg(shellOut, formatStream)
	if !exist {
		go func() {
			err := shellOut.Client.RunCommand(command + fmt.Sprintf(";echo '%s'", ExistTip))
			if err != nil {
				fmt.Println(fmt.Sprintf(`执行错误 %s`, err.Error()))
			}
		}()
		return nil
	} else {
		remainLen := len(shellOut.remainContents)
		startIdx := 0
		if remainLen > MaxSendLength {
			startIdx = remainLen - MaxSendLength
		}
		// 逐行推送，使 SSE 过滤关键词逐行生效：不满足过滤条件的行不会被推送
		for i := startIdx; i < remainLen; i++ {
			h.SendMsg(shellOut, shellOut.remainContents[i])
		}
		h.SendErrList(shellOut)
		h.SendFilterList(shellOut)
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
	shellOut.alertEvents = make([]ShellOutAlertEvent, 0)
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

func (h *TShellOut) CleanLog(shellClientId string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	shellOut := h.ShellOutMap[shellClientId]
	if shellOut == nil {
		return
	}
	shellOut.remainContents = []string{}
	shellOut.pendingMsg = ``
}

// SetFilterKeywords 设置指定 shell 会话的 SSE 推送过滤关键词。
// keywords 为空切片时表示取消过滤，所有消息正常推送。
func (h *TShellOut) SetFilterKeywords(shellClientId string, keywords []string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	shellOut := h.ShellOutMap[shellClientId]
	if shellOut == nil {
		return
	}
	shellOut.filterLock.Lock()
	shellOut.filterKeywords = keywords
	shellOut.filterLock.Unlock()
}

// GetFilterKeywords 获取指定 shell 会话当前的 SSE 推送过滤关键词。
func (h *TShellOut) GetFilterKeywords(shellClientId string) []string {
	h.lock.Lock()
	defer h.lock.Unlock()
	shellOut := h.ShellOutMap[shellClientId]
	if shellOut == nil {
		return nil
	}
	shellOut.filterLock.RLock()
	defer shellOut.filterLock.RUnlock()
	return shellOut.filterKeywords
}

func (h *TShellOut) SetReceiveMsg(shellOut *ShellOut, formatStream func(string) []string) {
	go h.timeBreakSsh(shellOut)
	shellOut.Client.SetFuncReceiveMsg(func(msg string) string {
		shellOut.lastReceiveTime = time.Now().Unix()
		if strings.Contains(msg, ExistTip) {
			if !strings.Contains(msg, fmt.Sprintf(`;echo '%s'`, ExistTip)) {
				h.SendMsg(shellOut, msg)
				h.SendMsg(shellOut, `监听到命令已中断，刷新后再次链接`)
				h.RmClient(shellOut.ShellClientId)
			}
			return msg
		}
		msg = gstool.StringFilterANSI(msg)
		msg = strings.Replace(msg, "\u001B", "", -1)

		// Combine with pending message and split by newlines
		fullMsg := shellOut.pendingMsg + msg
		lines := strings.Split(fullMsg, "\n")

		// The last part is incomplete (no trailing newline), save it for next message
		if len(lines) > 0 {
			shellOut.pendingMsg = lines[len(lines)-1]
			lines = lines[:len(lines)-1]
		}

		// Process each complete line
		for _, line := range lines {
			if line == "" {
				continue
			}

			// Store original content with line number
			lineNumberStr := p_common.TBaseClient.GetUnique(`line`)
			lineNumber := cast.ToInt64(lineNumberStr[4:]) // Remove "line" prefix, keep number
			shellOut.sourceContents = append(shellOut.sourceContents, SourceLine{
				LineNumber: lineNumber,
				Content:    gstool.TimeNowUnixToString(``) + ` ` + line,
			})
			if len(shellOut.sourceContents) > MaxSourceLength {
				shellOut.sourceContents = shellOut.sourceContents[MaxSourceLength:]
			}

			ruleResult := h.ApplyRuleItems(shellOut, line, lineNumber)
			if ruleResult.Dropped {
				continue
			}

			// Store filtered content
			shellOut.remainContents = append(shellOut.remainContents, line)
			if len(shellOut.remainContents) > MaxRemainLength {
				shellOut.remainContents = shellOut.remainContents[MaxRemainLength:]
			}

			// Push to frontend
			if formatStream != nil {
				msgList := formatStream(line)
				for _, formattedMsg := range msgList {
					h.SendMsg(shellOut, formattedMsg)
				}
			} else {
				h.SendMsg(shellOut, line)
			}
		}
		return msg
	})
}

func (h *TShellOut) timeBreakSsh(shellOut *ShellOut) {
	for {
		select {
		case <-shellOut.breakTimer.C:
			second := time.Now().Unix() - shellOut.lastReceiveTime
			if second > 30 {
				if shellOut.Sse != nil {
					shellOut.Sse.Send(`warning:` + cast.ToString(second) + `秒未收到任何内容返回,链接可能已断开,尝试重新启动` + "\n")
				}
			}
		}
	}
}

func (h *TShellOut) SendMsg(shellOut *ShellOut, msg string) {
	msg = strings.Replace(msg, `\n`, "\n", -1)
	if shellOut == nil || shellOut.Sse == nil {
		return
	}
	// 检查 SSE 推送过滤关键词：消息必须包含所有关键词才推送，空列表表示不过滤
	shellOut.filterLock.RLock()
	keywords := shellOut.filterKeywords
	shellOut.filterLock.RUnlock()
	if len(keywords) > 0 {
		for _, kw := range keywords {
			if !strings.Contains(msg, kw) {
				return
			}
		}
	}
	shellOut.Sse.Send(msg)
}

func (h *TShellOut) SendEvent(shellOut *ShellOut, eventType, msg string) {
	msg = strings.Replace(msg, `\n`, "\n", -1)
	if shellOut == nil || shellOut.Sse == nil {
		return
	}
	shellOut.Sse.Send(msg)
}

func (h *TShellOut) SendErrList(shellOut *ShellOut) {
	if shellOut == nil || shellOut.Sse == nil {
		return
	}
	shellOut.Sse.Send(shellOut.errorList, p_define.SseContentTypeErrorList)
}

func (h *TShellOut) SendFilterList(shellOut *ShellOut) {
	if shellOut == nil || shellOut.Sse == nil {
		return
	}
	shellOut.Sse.Send(shellOut.regexFiltersTips, p_define.SseContentTypeFilterList)
}

func (h *TShellOut) SendFilter(shellOut *ShellOut, msg string) {
	if shellOut == nil || shellOut.Sse == nil {
		return
	}
	shellOut.Sse.Send(msg, p_define.SseContentTypeFilter)
}

func (h *TShellOut) SendErr(shellOut *ShellOut, err ErrorBlock) {
	if shellOut == nil || shellOut.Sse == nil {
		return
	}
	shellOut.Sse.Send(err, p_define.SseContentTypeError)
}

func (h *TShellOut) RmClient(uniqueKey string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if sh, ok := h.ShellOutMap[uniqueKey]; ok {
		sh.Client.CloseTerminal()
	} else {
		return
	}
	h.ShellOutMap[uniqueKey].breakTimer.Stop()
	delete(h.ShellOutMap, uniqueKey)
}

func (h *TShellOut) WalkShellList(businessFunc func(uniqueKey string, gsShell *gsssh.SshTerminal)) {
	h.lock.Lock()
	defer h.lock.Unlock()
	for k, v := range h.ShellOutMap {
		businessFunc(k, v.Client)
	}
}

// ErrorContext 返回 错误行的上下文
func (h *TShellOut) ErrorContext(shellClientId string, errorLine string, n int) ([]LineContext, int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	shellOut, ok := h.ShellOutMap[shellClientId]
	if !ok || n < 0 {
		return []LineContext{}, 0
	}

	// 将传入的errorLine（行号）转换为int64
	targetLineNumber := cast.ToInt64(errorLine)
	if targetLineNumber == 0 {
		return []LineContext{}, 0
	}

	src := shellOut.sourceContents
	// 按行号匹配
	for i, sourceLine := range src {
		if sourceLine.LineNumber != targetLineNumber {
			continue
		}

		// 计算合法区间
		start := i - n
		if start < 0 {
			start = 0
		}
		end := i + n + 1 // 切片右边界开区间
		if end > len(src) {
			end = len(src)
		}

		// 提取上下文内容（带行号）
		lines := make([]LineContext, end-start)
		for j := start; j < end; j++ {
			lines[j-start] = LineContext{
				LineNumber: src[j].LineNumber,
				Content:    src[j].Content,
			}
		}
		firstLineNo := int(src[start].LineNumber) // 使用实际行号
		return lines, firstLineNo
	}
	return []LineContext{}, 0
}

type Search struct {
	LineNumber int64  // 行号
	Content    string // 匹配的内容
	IsRead     bool   // true 已经搜索过
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	ShellClientId  string `json:"shell_client_id"`
	CurrentCommand string `json:"current_command"` // 当前命令
	Status         string `json:"status"`          // active: 活跃, idle: 闲置, broken: 断开
	ConnectTime    string `json:"connect_time"`    // 连接时间
	ConnectSeconds int64  `json:"connect_seconds"` // 连接时长(秒)
	LastReceive    string `json:"last_receive"`    // 最后接收时间
	IdleSeconds    int64  `json:"idle_seconds"`    // 闲置时长(秒)
	Type           string `json:"type"`            // shell_out
}

func getTerminalCurrentCommand(gsShell *gsssh.SshTerminal) string {
	if gsShell == nil {
		return ``
	}
	defer func() {
		_ = recover()
	}()
	val := reflect.ValueOf(gsShell)
	if !val.IsValid() || val.Kind() != reflect.Ptr || val.IsNil() {
		return ``
	}
	elem := val.Elem()
	if !elem.IsValid() || elem.Kind() != reflect.Struct {
		return ``
	}
	field := elem.FieldByName(`command`)
	if !field.IsValid() || field.Kind() != reflect.String {
		return ``
	}
	return strings.TrimSpace(field.String())
}

// GetConnections 获取所有连接状态
func (h *TShellOut) GetConnections() []ConnectionInfo {
	h.lock.Lock()
	defer h.lock.Unlock()

	connections := make([]ConnectionInfo, 0)
	now := time.Now().Unix()

	for shellClientId, shellOut := range h.ShellOutMap {
		connectTime := ``
		connectSeconds := int64(0)
		if shellOut.startTime > 0 {
			connectTime = gstool.TimeUnixToString(time.Unix(shellOut.startTime, 0), `Y-m-d H:i:s`)
			connectSeconds = now - shellOut.startTime
		}
		lastReceive := ``
		idleSeconds := int64(0)
		if shellOut.lastReceiveTime > 0 {
			lastReceive = gstool.TimeUnixToString(time.Unix(shellOut.lastReceiveTime, 0), `Y-m-d H:i:s`)
			idleSeconds = now - shellOut.lastReceiveTime
		}
		info := ConnectionInfo{
			ShellClientId:  shellClientId,
			CurrentCommand: getTerminalCurrentCommand(shellOut.Client),
			Status:         "active",
			ConnectTime:    connectTime,
			ConnectSeconds: connectSeconds,
			LastReceive:    lastReceive,
			IdleSeconds:    idleSeconds,
			Type:           "shell_out",
		}

		// 判断连接状态
		if idleSeconds > 30 {
			info.Status = "idle"
		}

		connections = append(connections, info)
	}

	return connections
}

// Reconnect 重连
func (h *TShellOut) Reconnect(shellClientId string) error {
	// 先查询数据库获取连接参数
	dbRecords, err := DbMain.Client.QuickQuery(`tbl_shell_out`, `*`, map[string]any{
		`shell_client_id`: shellClientId,
	}).Limit(1).All()
	if err != nil {
		return fmt.Errorf("查询数据库失败: %s", err.Error())
	}
	if len(dbRecords) == 0 {
		return fmt.Errorf("连接记录不存在")
	}

	record := dbRecords[0]
	sshId := cast.ToString(record[`ssh_id`])
	command := cast.ToString(record[`command`])
	groupId := cast.ToInt(record[`group_id`])

	// 获取SSH配置
	sshConfig, err := DbMain.GetSshConfig(sshId)
	if err != nil {
		return fmt.Errorf("获取SSH配置失败: %s", err.Error())
	}

	// 移除旧连接
	h.RmClient(shellClientId)

	// 创建新的SSE对象（这里使用空的SSE，因为重连后需要前端重新连接）
	sse := &p_sse.SseShell{}

	// 获取新的客户端
	shellOut, _, err := h.GetClient(sshConfig, shellClientId, sse, groupId, 0, nil)
	if err != nil {
		return fmt.Errorf("重新建立连接失败: %s", err.Error())
	}

	// 重新执行命令
	go func() {
		err := shellOut.Client.RunCommand(command + fmt.Sprintf(";echo '%s'", ExistTip))
		if err != nil {
			fmt.Println(fmt.Sprintf(`重连后执行命令错误 %s`, err.Error()))
		}
	}()

	return nil
}

// ShellOutSearchContent 匹配所有
func (h *TShellOut) ShellOutSearchContent(shellClientId string, searchContent string, maxNum int) ([]Search, int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	shellOut, ok := h.ShellOutMap[shellClientId]
	if !ok {
		return []Search{}, 0
	}
	searchs := make([]Search, 0)
	gstool.ArrayWalkDesc(shellOut.sourceContents, func(line SourceLine) bool {
		if !strings.Contains(line.Content, searchContent) {
			return true
		}
		if _, ok := shellOut.searchReadContents[line.Content]; ok {
			searchs = append(searchs, Search{
				LineNumber: line.LineNumber,
				Content:    line.Content,
				IsRead:     true,
			})
		} else {
			searchs = append(searchs, Search{
				LineNumber: line.LineNumber,
				Content:    line.Content,
				IsRead:     false,
			})
			shellOut.searchReadContents[line.Content] = nil
		}
		if len(searchs) > maxNum {
			return false
		}
		return true
	})
	return searchs, len(searchs)
}

// runShellOutPostConnectCmds 从 SSH 配置中读取连接后命令列表，每行一条，空则回退 pwd，遇错停止。
func runShellOutPostConnectCmds(gsShell *gsssh.SshTerminal, sshConfig map[string]any) error {
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

	for _, cmd := range cmds {
		if err := gsShell.RunCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}
