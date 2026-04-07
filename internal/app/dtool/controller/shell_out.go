package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_define"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"sort"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
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
			b.Broadcast()
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
	sse := gsgin.SseGetByClientId(define.SseShellConnections)
	if sse == nil {
		return
	}

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

	data := map[string]any{
		`connections`: allConnections,
		`total`:       len(allConnections),
	}

	err := sse.SendToChan(gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseShellConnections,
		Data:            data,
		Type:            p_define.SseContentTypeConnections,
	}))
	if err != nil {
		gstool.FmtPrintlnLogTime(`ShellConnections广播错误 %s`, err.Error())
	}
}

func ShellOut(c *gin.Context) {
	reqMap, client, shellClientId, err := getShellOutComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	command := cast.ToString(reqMap[`command`])
	_ = client.RunCommand(command)
	id, err := common.DbMain.Client.QuickCreate(`tbl_shell_out`, map[string]any{
		`command`:         command,
		`shell_client_id`: shellClientId,
		`name`:            cast.ToString(reqMap[`name`]),
		`group_id`:        reqMap[`group_id`],
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
	if groupId == 0 {
		gsgin.GinResponseError(c, `组id不能为空`, nil)
		return
	}
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: cast.ToString(dataMap[`sse_distribute_id`]),
	}
	err = component.ShellOutClient.SetClientSseId(shellClientId, sshId, sse, command, groupId, func(s string) []string {
		return []string{p_common.TBaseClient.FilterTerminalChars(s)}
	})
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
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
	sshConfig, _ := common.DbMain.GetSshConfig(sshId)
	shellClientId := p_common.TBaseClient.GetUnique(`shell_out_`)
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: cast.ToString(dataMap[`sse_distribute_id`]),
	}

	shellOut, _, sshClientErr := component.ShellOutClient.GetClient(sshConfig, shellClientId, sse, groupId, nil)
	if sshClientErr != nil {
		return nil, nil, ``, sshClientErr
	}
	return dataMap, shellOut.Client, shellClientId, nil
}
