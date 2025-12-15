package common

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_define"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const MaxSourceLength = 20000 //原始内容最多保留多少行 用于搜索
const MaxRemainLength = 10000 //过滤后内容最多保留多少行
const MaxSendLength = 1000    //刷新页面后最多发送给前端多少行

/*  等待式输出 ssh 不重复使用，持续等待 ssh 返回结果 */

// ShellOut 单个 ssh 会话
type ShellOut struct {
	Client           *gsssh.SshTerminal
	Sse              *p_sse.SseShell
	errorList        []ErrorBlock   // 最终归档的错误块
	remainContents   []string       // 保留的内容(替换后的)
	sourceContents   []string       // 原本内容
	regexFiltersTips map[string]int //过滤正则数量统计
	startTime        int64          //启动时间
	groupId          int            //分组id
}

// ErrorBlock 错误块
type ErrorBlock struct {
	Lines     []string `json:"lines"`
	ErrorLine string   `json:"error_line"`
	Time      string   `json:"time"`
}

// TShellOut 管理多个 ShellOut
type TShellOut struct {
	ShellOutMap       map[string]*ShellOut
	lock              sync.Mutex
	log               *gstool.GsSlog
	GroupRegexFilters map[int][]string //过滤规则
	GroupRegexErrors  map[int][]string //错误规则
	GroupNoErrors     map[int][]string //错误再次排除规则
	GroupConfigLock   sync.Mutex
}

var ShellOutClient *TShellOut

// NewTShellOut 构造函数
func NewTShellOut() *TShellOut {
	log := gstool.NewSlog3(component.EnvClient.LogPath, `shell_wait`)
	_ = log.CleanOldLogs(2)
	shellOut := &TShellOut{
		ShellOutMap:       make(map[string]*ShellOut),
		log:               log,
		GroupRegexFilters: make(map[int][]string),
		GroupRegexErrors:  make(map[int][]string),
		GroupNoErrors:     make(map[int][]string),
	}
	return shellOut
}

func (h *TShellOut) InitGroupConfigs() {
	h.GroupConfigLock.Lock()
	defer h.GroupConfigLock.Unlock()
	all, allErr := DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeShellOut,
	}).All()
	if allErr != nil {
		gstool.FmtPrintlnLogTime(`获取ssh配置错误 %s`, allErr.Error())
		return
	}
	for _, item := range all {
		groupId := cast.ToInt(item[`id`])
		extra1 := cast.ToString(item[`extra_1`])
		extra2 := cast.ToString(item[`extra_2`])
		extra3 := cast.ToString(item[`extra_3`])
		h.GroupRegexFilters[groupId] = strings.Split(extra1, "\n")
		h.GroupRegexErrors[groupId] = strings.Split(extra2, "\n")
		h.GroupNoErrors[groupId] = strings.Split(extra3, "\n")
	}
}

// GetClient 获取或新建 ssh 客户端
func (h *TShellOut) GetClient(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell, groupId int,
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
	gsShell := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
		Name:     "",
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
	}))
	// 断开回调
	gsShell.SetFuncBroken(func() {
		sse.Send(` 注意：连接已中断，下次动作时进行链接` + "\n")
		h.RmClient(shellClientId)
	})
	gsShell.SetCombineNum(1)
	gsShell.SetMaxBufferSize(2 * 1024 * 1024) //最大允许2M的输出
	gsShell.SetPtyConfig(gsssh.PtyConfig{
		Width: 1000,
	})

	if err := gsShell.RunCommand(`pwd`); err != nil {
		gstool.FmtPrintlnLogTime(`shell out 执行失败 %s`, err.Error())
		return nil, false, err
	}

	// 新建 ShellOut
	shellOut := &ShellOut{
		Client:           gsShell,
		Sse:              sse,
		regexFiltersTips: map[string]int{},
		startTime:        time.Now().Unix(),
		groupId:          groupId,
		errorList:        make([]ErrorBlock, 0),
		remainContents:   make([]string, 0),
		sourceContents:   make([]string, 0),
	}
	h.SetReceiveMsg(shellOut, formatStream)
	h.ShellOutMap[shellClientId] = shellOut
	return shellOut, false, nil
}

// SetClientSseId 设置 sse 推送 & 错误检测
func (h *TShellOut) SetClientSseId(shellClientId, sshId string, sse *p_sse.SseShell, command string, groupId int,
	formatStream func(string) []string) error {

	sshConfig, _ := DbMain.GetSshConfig(sshId)
	shellOut, exist, err := h.GetClient(sshConfig, shellClientId, sse, groupId, formatStream)
	if err != nil {
		return err
	}
	shellOut.groupId = groupId
	shellOut.Sse = sse
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
		remainLen := len(shellOut.remainContents)
		if remainLen > MaxSendLength {
			h.SendMsg(shellOut, strings.Join(shellOut.remainContents[(remainLen-MaxSendLength):], "\n"))
		} else {
			h.SendMsg(shellOut, strings.Join(shellOut.remainContents, "\n"))
		}
		h.SendErrList(shellOut)
		h.SendFilterList(shellOut)
	}
	return nil
}

func (h *TShellOut) GetRegexFilters(shellOut *ShellOut) []string {
	h.GroupConfigLock.Lock()
	defer h.GroupConfigLock.Unlock()
	return h.GroupRegexFilters[shellOut.groupId]
}

func (h *TShellOut) GetRegexErrors(shellOut *ShellOut) []string {
	h.GroupConfigLock.Lock()
	defer h.GroupConfigLock.Unlock()
	return h.GroupRegexErrors[shellOut.groupId]
}

func (h *TShellOut) GetNoErrors(shellOut *ShellOut) []string {
	h.GroupConfigLock.Lock()
	defer h.GroupConfigLock.Unlock()
	return h.GroupNoErrors[shellOut.groupId]
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

func (h *TShellOut) CleanLog(shellClientId string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	shellOut := h.ShellOutMap[shellClientId]
	if shellOut == nil {
		return
	}
	shellOut.remainContents = []string{}
}

func (h *TShellOut) SetReceiveMsg(shellOut *ShellOut, formatStream func(string) []string) {
	shellOut.Client.SetFuncStreamReceive(func(msg string) {
		msg = gstool.StringFilterANSI(msg)
		msg = strings.Replace(msg, "\u001B", "", -1)
		//原内容处理
		shellOut.sourceContents = append(shellOut.sourceContents, gstool.TimeNowUnixToString(``)+` `+msg)
		if len(shellOut.sourceContents) > MaxSourceLength {
			shellOut.sourceContents = shellOut.sourceContents[MaxSourceLength:]
		}
		//过滤内容处理
		boolFilter := h.RegexFilter(shellOut, msg)
		if boolFilter {
			return
		}
		//保留内容处理
		shellOut.remainContents = append(shellOut.remainContents, msg)
		if len(shellOut.remainContents) > MaxRemainLength {
			shellOut.remainContents = shellOut.remainContents[MaxRemainLength:]
		}
		//错误检测
		h.RegexError(shellOut, msg)
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

func (h *TShellOut) RegexError(shellOut *ShellOut, msg string) {
	noErrors := h.GetNoErrors(shellOut)
	for _, regexError := range h.GetRegexErrors(shellOut) {
		if regexError == `` {
			continue
		}
		if strings.TrimSpace(regexError) == `` {
			continue
		}
		regexParams := strings.Split(regexError, `#`)
		if len(regexParams) == 2 {
			regexError = regexParams[1]
		}
		var re = regexp.MustCompile(regexError)
		if re.MatchString(msg) {
			//再次过滤
			for _, noError := range noErrors {
				if strings.Contains(msg, noError) {
					continue
				}
			}
			block := ErrorBlock{
				Lines:     []string{},
				ErrorLine: msg,
				Time:      gstool.TimeNowUnixToString(``),
			}
			shellOut.errorList = append(shellOut.errorList, block)
			h.SendErr(shellOut, block)
		}
	}
}

func (h *TShellOut) RegexFilter(shellOut *ShellOut, msg string) bool {
	boolFilter := false
	split := `#`
	for _, regexFilter := range h.GetRegexFilters(shellOut) {
		if regexFilter == `` {
			continue
		}
		if strings.TrimSpace(regexFilter) == `` {
			continue
		}
		name := ``
		regexParams := strings.Split(regexFilter, split)
		if len(regexParams) == 2 {
			regexFilter = regexParams[1]
			name = regexParams[0]
		}
		var re = regexp.MustCompile(regexFilter)
		if re.MatchString(msg) {
			boolFilter = true
			unikey := name + split + regexFilter
			if gstool.MapKeyExist(&shellOut.regexFiltersTips, regexFilter) {
				shellOut.regexFiltersTips[unikey] += 1
			} else {
				shellOut.regexFiltersTips[unikey] = 1
			}
			h.SendFilter(shellOut, unikey)
			break
		}
	}
	return boolFilter
}

func (h *TShellOut) SendMsg(shellOut *ShellOut, msg string) {
	msg = strings.Replace(msg, `\n`, "\n", -1)
	shellOut.Sse.Send(msg)
}

func (h *TShellOut) SendEvent(shellOut *ShellOut, eventType, msg string) {
	msg = strings.Replace(msg, `\n`, "\n", -1)
	shellOut.Sse.Send(msg)
}

func (h *TShellOut) SendErrList(shellOut *ShellOut) {
	shellOut.Sse.Send(shellOut.errorList, p_define.SseContentTypeErrorList)
}

func (h *TShellOut) SendFilterList(shellOut *ShellOut) {
	shellOut.Sse.Send(shellOut.regexFiltersTips, p_define.SseContentTypeFilterList)
}

func (h *TShellOut) SendFilter(shellOut *ShellOut, msg string) {
	shellOut.Sse.Send(msg, p_define.SseContentTypeFilter)
}

func (h *TShellOut) SendErr(shellOut *ShellOut, err ErrorBlock) {
	shellOut.Sse.Send(err, p_define.SseContentTypeError)
}

func (h *TShellOut) RmClient(uniqueKey string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if sh, ok := h.ShellOutMap[uniqueKey]; ok {
		sh.Client.CloseTerminal()
	}
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
func (h *TShellOut) ErrorContext(shellClientId string, errorLine string, n int) (lines []string, firstLineNo int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	shellOut, ok := h.ShellOutMap[shellClientId]
	if !ok || n < 0 {
		return []string{}, 0
	}
	src := shellOut.sourceContents
	for i, line := range src {
		if !strings.Contains(line, errorLine) {
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

		lines = make([]string, end-start)
		copy(lines, src[start:end])
		firstLineNo = start + 1 // 行号从 1 开始
		return
	}
	return []string{}, 0
}

// ShellOutSearchContent 匹配所有
func (h *TShellOut) ShellOutSearchContent(shellClientId string, searchContent string, maxNum int) ([]string, int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	shellOut, ok := h.ShellOutMap[shellClientId]
	if !ok {
		return []string{}, 0
	}
	searchs := make([]string, 0)
	gstool.ArrayWalkDesc(shellOut.sourceContents, func(line string) bool {
		if !strings.Contains(line, searchContent) {
			return true
		}
		searchs = append(searchs, line)
		if len(searchs) > maxNum {
			return false
		}
		return true
	})
	return searchs, len(searchs)
}
