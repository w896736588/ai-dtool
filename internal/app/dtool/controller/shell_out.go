package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_define"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gsssh"
	"github.com/w896736588/go-tool/gstool"
)

var (
	// getShellOutComponentFunc 允许测试替换 SSH 初始化，聚焦持久化断言。 // Allows tests to bypass SSH setup and focus on persistence assertions.
	getShellOutComponentFunc = getShellOutComponent
	// shellOutRunCommandFunc 允许测试跳过真实终端执行。 // Allows tests to skip real terminal execution during controller coverage.
	shellOutRunCommandFunc = func(client any, command string) error {
		if client == nil {
			return nil
		}
		terminalClient, ok := client.(*gsssh.SshTerminal)
		if !ok || terminalClient == nil {
			return nil
		}
		return terminalClient.RunCommand(command)
	}
)

type ShellConnectionView struct {
	ShellClientId  string `json:"shell_client_id"`
	CurrentCommand string `json:"current_command"`
	Status         string `json:"status"`
	ConnectTime    string `json:"connect_time"`
	ConnectSeconds int64  `json:"connect_seconds"`
	LastReceive    string `json:"last_receive"`
	IdleSeconds    int64  `json:"idle_seconds"`
	Type           string `json:"type"`
}

// ShellConnectionsBroadcaster Shell连接状态广播器
type ShellConnectionsBroadcaster struct {
	ticker *time.Ticker
	stopC  chan struct{}
}

var ShellConnectionsBroadcasterInstance *ShellConnectionsBroadcaster

// NewShellConnectionsBroadcaster 创建广播器并启动定时推送
func NewShellConnectionsBroadcaster(interval time.Duration) *ShellConnectionsBroadcaster {
	b := &ShellConnectionsBroadcaster{
		ticker: time.NewTicker(interval),
		stopC:  make(chan struct{}),
	}
	go b.run()
	return b
}

// run 定时广播连接状态
func (b *ShellConnectionsBroadcaster) run() {
	for {
		select {
		case <-b.ticker.C:
			// 中文注释：连接状态改为跟随普通 SSE client_id 推送，这里保留空轮询以兼容旧初始化流程。
			// English comment: Shell connection events now ride on normal SSE client IDs, so global broadcast is a no-op.
		case <-b.stopC:
			return
		}
	}
}

// Stop 停止广播
func (b *ShellConnectionsBroadcaster) Stop() {
	close(b.stopC)
	b.ticker.Stop()
}

// Broadcast 广播当前所有Shell连接状态
func (b *ShellConnectionsBroadcaster) Broadcast() {
}

// buildShellConnectionsPayload 构造当前所有 Shell 连接状态快照。
// buildShellConnectionsPayload builds the current shell connection snapshot payload.
func buildShellConnectionsPayload() map[string]any {
	// 获取p_shell.Shell类型的连接
	shellConnections := component.ShellClient.GetConnections()

	// 合并两种类型的连接
	allConnections := make([]ShellConnectionView, 0, len(shellConnections))
	for _, conn := range shellConnections {
		allConnections = append(allConnections, ShellConnectionView{
			ShellClientId:  conn.ShellClientId,
			CurrentCommand: conn.CurrentCommand,
			Status:         conn.Status,
			ConnectTime:    conn.ConnectTime,
			ConnectSeconds: conn.ConnectSeconds,
			Type:           conn.Type,
		})
	}
	sort.Slice(allConnections, func(i, j int) bool {
		if allConnections[i].ConnectSeconds == allConnections[j].ConnectSeconds {
			return allConnections[i].ShellClientId < allConnections[j].ShellClientId
		}
		return allConnections[i].ConnectSeconds < allConnections[j].ConnectSeconds
	})

	return map[string]any{
		`connections`: allConnections,
		`total`:       len(allConnections),
	}
}

// sendShellConnectionsSnapshot 向指定 SSE 连接发送一次连接状态快照。
// sendShellConnectionsSnapshot sends one shell-connections snapshot to the provided SSE client.
func sendShellConnectionsSnapshot(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	data := buildShellConnectionsPayload()
	err := sse.SendToChan(gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseShellConnections,
		Data:            data,
		Type:            p_define.SseContentTypeConnections,
	}))
	if err != nil {
		gstool.FmtPrintlnLogTime(`ShellConnections广播错误 %s`, err.Error())
	}
}

// BindShellConnectionsSSE 为普通 SSE client 绑定连接状态推送，无需单独 shell_connections 连接。
// BindShellConnectionsSSE attaches shell-connections events to a normal SSE client_id stream.
func BindShellConnectionsSSE(sse *gsgin.Sse, stopC chan int, interval time.Duration) {
	if sse == nil {
		return
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	// 中文注释：建连后立即推一次，避免前端初次打开时要等下一个周期。
	// English comment: Push once immediately so the UI does not wait for the first ticker tick.
	sendShellConnectionsSnapshot(sse)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sendShellConnectionsSnapshot(sse)
			case <-stopC:
				return
			}
		}
	}()
}

func ShellOut(c *gin.Context) {
	reqMap, client, shellClientId, err := getShellOutComponentFunc(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	command := cast.ToString(reqMap[`command`])
	_ = shellOutRunCommandFunc(client, command)
	id, err := common.DbMain.Client.QuickCreate(`tbl_shell_out`, map[string]any{
		`command`:         command,
		`shell_client_id`: shellClientId,
		`name`:            cast.ToString(reqMap[`name`]),
		`group_id`:        reqMap[`group_id`],
		`rule_set_id`:     cast.ToInt(reqMap[`rule_set_id`]),
		`is_run`:          1,
		`ssh_id`:          cast.ToString(reqMap[`ssh_id`]),
		`create_time`:     time.Now().Unix(),
		`update_time`:     time.Now().Unix(),
	}).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`shell_client_id`: shellClientId,
		`id`:              cast.ToString(id),
	})
	return
}

func ShellOutEdit(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	id, err := common.DbMain.Client.QuickUpdate(`tbl_shell_out`, map[string]any{
		`id`: reqMap[`id`],
	}, map[string]any{
		`name`:        cast.ToString(reqMap[`name`]),
		`command`:     cast.ToString(reqMap[`command`]),
		`ssh_id`:      cast.ToInt(reqMap[`ssh_id`]),
		`group_id`:    reqMap[`group_id`],
		`rule_set_id`: cast.ToInt(reqMap[`rule_set_id`]),
		`update_time`: time.Now().Unix(),
	}).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`id`: cast.ToString(id),
	})
	return
}

func ShellOutErrorContext(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	errorLine := cast.ToString(reqMap[`error_line`])
	lines, _ := component.ShellOutClient.ErrorContext(shellClientId, errorLine, 10)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`lines`: lines,
	})
	return
}

func ShellOutSearchContent(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	searchContent := cast.ToString(reqMap[`search_content`])
	searchContents := strings.Split(searchContent, "##")
	allLines := make([]common.Search, 0)
	allNumber := 0
	for _, searchContent := range searchContents {
		if searchContent == `` {
			continue
		}
		lines, number := component.ShellOutClient.ShellOutSearchContent(shellClientId, searchContent, 1000)
		allLines = append(allLines, lines...)
		allNumber += number
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`lines`:  allLines,
		`number`: allNumber,
	})
	return
}

// sendFullpageSSEStep 通过 SSE 通道向前端推送步骤进度消息（仅 fullpage 场景）
func sendFullpageSSEStep(sse *gsgin.Sse, sseDistributeId string, msg string) {
	if sse == nil {
		return
	}
	_ = sse.SendToChan(gstool.JsonEncode(p_define.SseData{
		SseDistributeId: sseDistributeId,
		Data:            msg,
		Type:            p_define.SseContentTypeMsg,
	}))
}

func ShellOutSetSeeId(c *gin.Context) {
	dataMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &dataMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shellClientId := cast.ToString(dataMap[`shell_client_id`])
	sshId := cast.ToString(dataMap[`ssh_id`])
	command := cast.ToString(dataMap[`command`])
	groupId := cast.ToInt(dataMap[`group_id`])
	ruleSetID := cast.ToInt(dataMap[`rule_set_id`])
	sseDistributeId := cast.ToString(dataMap[`sse_distribute_id`])
	// 优先从 body 参数取 sse_client_id（兼容独立 SSE 场景），兼容 Header 中的 SseClientId
	sseClientId := cast.ToString(dataMap[`sse_client_id`])
	if sseClientId == `` {
		sseClientId = c.GetHeader(`SseClientId`)
	}

	gstool.FmtPrintlnLogTime(`[ShellOutSetSeeId] 收到请求 sse_distribute_id=%s shell_client_id=%s SseClientId=%s command=%s group_id=%d`,
		sseDistributeId, shellClientId, sseClientId, command, groupId)

	if groupId == 0 {
		gsgin.GinResponseError(c, `组id不能为空`, nil)
		return
	}

	// 优先从 Fullpage 专用 SSE 中查找，找不到则回退到通用 SSE
	sseConn := GetFullpageSseByClientID(sseClientId)
	if sseConn != nil {
		gstool.FmtPrintlnLogTime(`[ShellOutSetSeeId] 使用Fullpage专用SSE sse_distribute_id=%s clientID=%s`, sseDistributeId, sseClientId)
		sendFullpageSSEStep(sseConn, sseDistributeId, "[系统] Fullpage SSE 通道已就绪 → 开始绑定 Shell 输出流\n")
	} else {
		sseConn = gsgin.SseGetByClientId(sseClientId)
		if sseConn != nil {
			gstool.FmtPrintlnLogTime(`[ShellOutSetSeeId] 使用通用SSE sse_distribute_id=%s clientID=%s`, sseDistributeId, sseClientId)
			sendFullpageSSEStep(sseConn, sseDistributeId, "[系统] 通用 SSE 通道已就绪 → 开始绑定 Shell 输出流\n")
		} else {
			gstool.FmtPrintlnLogTime(`[ShellOutSetSeeId] SSE连接不存在 sse_distribute_id=%s clientID=%s`, sseDistributeId, sseClientId)
			gsgin.GinResponseError(c, `SSE连接未建立，无法绑定Shell输出流（请刷新页面后重试）`, nil)
			return
		}
	}

	sse := &p_sse.SseShell{
		Sse:             sseConn,
		SseDistributeId: sseDistributeId,
	}

	// 推送步骤详情：查询 SSH 配置
	sendFullpageSSEStep(sseConn, sseDistributeId, fmt.Sprintf("[系统] ① 查询 SSH 配置 (ssh_id=%s)...\n", sshId))
	sshConfig, getSshErr := component.DbMain.GetSshConfig(sshId)
	if getSshErr != nil {
		sendFullpageSSEStep(sseConn, sseDistributeId, fmt.Sprintf("[系统] ❌ SSH 配置查询失败: %s\n", getSshErr.Error()))
		gsgin.GinResponseError(c, getSshErr.Error(), nil)
		return
	}
	sendFullpageSSEStep(sseConn, sseDistributeId, fmt.Sprintf("[系统] ② SSH 目标: %s:%s (user=%s)\n",
		cast.ToString(sshConfig[`host`]), cast.ToString(sshConfig[`port`]), cast.ToString(sshConfig[`username`])))

	// 推送步骤详情：即将执行命令
	sendFullpageSSEStep(sseConn, sseDistributeId, fmt.Sprintf("[系统] ③ 即将执行命令: %s\n", command))

	err = component.ShellOutClient.SetClientSseId(shellClientId, sshId, sse, command, groupId, ruleSetID, func(s string) []string {
		return []string{p_common.TBaseClient.FilterTerminalChars(s)}
	})
	if err != nil {
		gstool.FmtPrintlnLogTime(`[ShellOutSetSeeId] SetClientSseId失败 sse_distribute_id=%s err=%s`, sseDistributeId, err.Error())
		sendFullpageSSEStep(sseConn, sseDistributeId, fmt.Sprintf("[系统] ❌ 绑定失败: %s\n", err.Error()))
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gstool.FmtPrintlnLogTime(`[ShellOutSetSeeId] 绑定成功 sse_distribute_id=%s shell_client_id=%s`, sseDistributeId, shellClientId)
	sendFullpageSSEStep(sseConn, sseDistributeId, "[系统] ✅ Shell 输出流绑定完成，等待命令输出...\n")
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
	return
}

func ShellOutCleanErrors(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	component.ShellOutClient.CleanErrors(shellClientId)
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
	return
}

func GetShellOuts(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	list, err := common.DbMain.Client.QuickQuery(`tbl_shell_out`, `*`, nil).Order(`id asc`).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
	return
}

func ShellOutDelete(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	sc := common.DbMain.Client.QuickDelete(`tbl_shell_out`, map[string]any{
		`id`: reqMap[`id`],
	})
	_, err = sc.Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), map[string]any{
			`sql`: sc.GetSql()})
		return
	}
	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	component.ShellOutClient.Delete(shellClientId)
	gsgin.GinResponseSuccess(c, ``, nil)
	return
}

func ShellOutStop(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	component.ShellOutClient.Delete(shellClientId)
	_, err = common.DbMain.Client.QuickUpdate(`tbl_shell_out`, map[string]any{
		`id`: reqMap[`id`],
	}, map[string]any{
		`is_run`:      0,
		`update_time`: time.Now().Unix(),
	}).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
	return
}

func ShellOutCleanLog(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	component.ShellOutClient.CleanLog(shellClientId)
	gsgin.GinResponseSuccess(c, ``, nil)
	return
}

// ShellOutGetFilter 获取指定 shell 会话当前的 SSE 推送过滤关键词，用于刷新页面后回填。
func ShellOutGetFilter(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	if shellClientId == `` {
		gsgin.GinResponseError(c, `shell_client_id不能为空`, nil)
		return
	}
	keywords := component.ShellOutClient.GetFilterKeywords(shellClientId)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`keywords`: keywords,
	})
	return
}

// ShellOutSetFilter 设置指定 shell 会话的 SSE 推送过滤关键词。
// filter_content 按空格分割，消息需包含所有关键词才推送；空字符串表示取消过滤。
func ShellOutSetFilter(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	if shellClientId == `` {
		gsgin.GinResponseError(c, `shell_client_id不能为空`, nil)
		return
	}
	filterContent := cast.ToString(reqMap[`filter_content`])
	var keywords []string
	if strings.TrimSpace(filterContent) != `` {
		keywords = strings.Fields(filterContent)
	}
	component.ShellOutClient.SetFilterKeywords(shellClientId, keywords)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`keywords`: keywords,
	})
	return
}

func ShellOutReconnect(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	if shellClientId == `` {
		gsgin.GinResponseError(c, `shell_client_id不能为空`, nil)
		return
	}

	component.ShellOutClient.RmClient(shellClientId)
	component.ShellClient.RmClient(shellClientId)
	gsgin.GinResponseSuccess(c, `重连成功`, nil)
	return
}

func getShellOutComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshTerminal, string, error) {
	dataMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &dataMap)
	if err != nil {
		return nil, nil, ``, err
	}
	sshId := dataMap[`ssh_id`]
	if cast.ToString(sshId) == `` {
		return nil, nil, ``, errors.New(`缺少ssh_id参数`)
	}
	groupId := cast.ToInt(dataMap[`group_id`])
	ruleSetID := cast.ToInt(dataMap[`rule_set_id`])
	sshConfig, _ := common.DbMain.GetSshConfig(sshId)
	shellClientId := p_common.TBaseClient.GetUnique(`shell_out_`)
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: cast.ToString(dataMap[`sse_distribute_id`]),
	}

	shellOut, _, sshClientErr := component.ShellOutClient.GetClient(sshConfig, shellClientId, sse, groupId, ruleSetID, nil)
	if sshClientErr != nil {
		return nil, nil, ``, sshClientErr
	}
	return dataMap, shellOut.Client, shellClientId, nil
}
