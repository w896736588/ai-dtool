package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_shell"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"regexp"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsdefine"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var supervisorConfLineReg = regexp.MustCompile(`^[^\s]+\.conf---.*$`)

// SupervisorRestartAll 重启所有
func SupervisorRestartAll(c *gin.Context) {
	reqMap, sshClient, err := getSupervisorComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dockerName := cast.ToString(reqMap[`docker_name`])
	restartCommand := p_shell.NewCommand().Sudo()
	if dockerName == `` { //非docker环境
		restartCommand.ConsumerRestartAll()
	} else {
		restartCommand.DockerExecConsumerRestartAll(dockerName)
	}
	_, _ = sshClient.RunCommandWait(restartCommand.GetCommand().ToStr(), 40*time.Second)
	statusRet, statusRetErr := getConsumerStatus(dockerName, sshClient)
	if statusRetErr != nil {
		gsgin.GinResponseError(c, statusRetErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, strings.Join(statusRet, gsdefine.Enter))
}

// SupervisorStopAll 停止所有
func SupervisorStopAll(c *gin.Context) {
	reqMap, sshClient, err := getSupervisorComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dockerName := cast.ToString(reqMap[`docker_name`])
	restartCommand := p_shell.NewCommand().Sudo()
	if dockerName == `` { //非docker环境
		restartCommand.ConsumerStopAll()
	} else {
		restartCommand.DockerExecConsumerStopAll(dockerName)
	}
	_, _ = sshClient.RunCommandWait(restartCommand.GetCommand().ToStr(), 40*time.Second)
	statusRet, statusRetErr := getConsumerStatus(dockerName, sshClient)
	if statusRetErr != nil {
		gsgin.GinResponseError(c, statusRetErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, strings.Join(statusRet, gsdefine.Enter))
}

func SupervisorStatusList(c *gin.Context) {
	reqMap, sshClient, err := getSupervisorComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dockerName := cast.ToString(reqMap[`docker_name`])
	statusRet, statusRetErr := getConsumerStatus(dockerName, sshClient)
	if statusRetErr != nil {
		gsgin.GinResponseError(c, statusRetErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, strings.Join(statusRet, gsdefine.Enter))
}

// 拿到消费者状态 支持docker与非docker
func getConsumerStatus(dockerName string, sshClient *gsssh.SshTerminal) ([]string, error) {
	//消费者
	retMsgList := make([]string, 0)
	if dockerName == `` { //非docker环境
		statusCommand := p_shell.NewCommand().Sudo().ConsumerStatus()
		statusRet, _ := sshClient.RunCommandWait(statusCommand.GetCommand().ToStr(), 40*time.Second)
		retMsgList = append(retMsgList, statusRet)
	} else {
		xkfStatusCommand := p_shell.NewCommand().Sudo()
		xkfStatusCommand.DockerExecConsumerStatus(dockerName)
		xkfStatusRet, _ := sshClient.RunCommandWait(xkfStatusCommand.GetCommand().ToStr(), 40*time.Second)
		retMsgList = append(retMsgList, xkfStatusRet)
	}
	return retMsgList, nil
}

// SupervisorConfigShow 查看supervisor配置内容
func SupervisorConfigShow(c *gin.Context) {
	reqMap, sshClient, err := getSupervisorComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	configPath := cast.ToString(reqMap[`config_path`])
	retMsgList := make([]string, 0)
	catCommand := p_shell.NewCommand().Sudo().Cat(configPath)
	ret, _ := sshClient.RunCommandWait(catCommand.GetCommand().ToStr(), 40*time.Second)
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponseSuccess(c, ``, strings.Join(retMsgList, gsdefine.Enter))
	return
}

// SupervisorRestart 重启某一个消费者
func SupervisorRestart(c *gin.Context) {
	reqMap, sshClient, err := getSupervisorComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	retMsgList := make([]string, 0)
	dockerName := cast.ToString(reqMap[`docker_name`])
	supervisorName := cast.ToString(reqMap[`supervisor_name`])
	// only_current_status=true 时，仅查询当前重启服务状态；默认查询全部服务状态（兼容其他页面刷新列表）。
	onlyCurrentStatus := cast.ToBool(reqMap[`only_current_status`])
	if supervisorName == `` {
		gsgin.GinResponseError(c, `消费者name不能为空`, nil)
		return
	}
	restartCommand := p_shell.NewCommand().Sudo()
	restartCommand.ConsumerRestart(dockerName, supervisorName)
	if onlyCurrentStatus {
		restartCommand.ConsumerStatusGrep(dockerName, supervisorName)
	} else {
		if dockerName == `` {
			restartCommand.ConsumerStatus()
		} else {
			restartCommand.DockerExecConsumerStatus(dockerName)
		}
	}
	ret, _ := sshClient.RunCommandWait(restartCommand.GetCommand().ToStr(), 40*time.Second)
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponseSuccess(c, ``, strings.Join(retMsgList, gsdefine.Enter))
	return
}

// SupervisorStop 停止某一个消费者
func SupervisorStop(c *gin.Context) {
	reqMap, sshClient, err := getSupervisorComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	retMsgList := make([]string, 0)
	dockerName := cast.ToString(reqMap[`docker_name`])
	supervisorName := cast.ToString(reqMap[`supervisor_name`])
	if supervisorName == `` {
		gsgin.GinResponseError(c, `消费者name不能为空`, nil)
		return
	}
	stopCommand := p_shell.NewCommand().Sudo()
	stopCommand.ConsumerStop(dockerName, supervisorName)
	stopCommand.ConsumerStatusGrep(dockerName, supervisorName)
	ret, _ := sshClient.RunCommandWait(stopCommand.GetCommand().ToStr(), 40*time.Second)
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponseSuccess(c, ``, strings.Join(retMsgList, gsdefine.Enter))
	return
}

// SupervisorConfList 获取配置列表
func SupervisorConfList(c *gin.Context) {
	allStart := time.Now()
	gstool.FmtPrintlnLogTime(`[SupervisorConfList] start sse_client_id=%s`, c.GetHeader(`SseClientId`))
	reqMap, sshClient, err := getSupervisorComponent(c)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[SupervisorConfList] getSupervisorComponent error=%s`, err.Error())
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	retMsgList := make([]string, 0)
	dockerName := cast.ToString(reqMap[`docker_name`])
	sshId := cast.ToString(reqMap[`ssh_id`])
	configListCommand := p_shell.NewCommand()
	configListCommand.ConsumerConfigList(dockerName)
	commandText := configListCommand.GetCommand().ToStr()
	gstool.FmtPrintlnLogTime(`[SupervisorConfList] run begin ssh_id=%s docker_name=%s command=%s`, sshId, dockerName, commandText)
	runStart := time.Now()
	ret, runErr := sshClient.RunCommandWait(commandText, 40*time.Second)
	gstool.FmtPrintlnLogTime(`[SupervisorConfList] run end ssh_id=%s cost_ms=%d ret_len=%d err=%v`,
		sshId, time.Since(runStart).Milliseconds(), len(ret), runErr)
	if runErr != nil {
		gsgin.GinResponseError(c, runErr.Error(), nil)
		return
	}
	confLines := parseSupervisorConfListOutput(ret)
	if len(confLines) == 0 {
		gstool.FmtPrintlnLogTime(`[SupervisorConfList] warn ssh_id=%s no valid conf lines, raw_ret=%q`, sshId, ret)
	}
	retMsgList = append(retMsgList, strings.Join(confLines, "\n"))
	gstool.FmtPrintlnLogTime(`[SupervisorConfList] success ssh_id=%s total_cost_ms=%d`, sshId, time.Since(allStart).Milliseconds())
	gsgin.GinResponseSuccess(c, ``, strings.Join(retMsgList, gsdefine.Enter))
	return
}

// parseSupervisorConfListOutput 过滤 ssh 回显噪音，仅保留 supervisor 配置行（*.conf---...）。
func parseSupervisorConfListOutput(raw string) []string {
	lines := strings.Split(raw, "\n")
	ret := make([]string, 0, len(lines))
	seen := make(map[string]struct{})
	for _, line := range lines {
		clean := strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))
		if clean == "" {
			continue
		}
		if !supervisorConfLineReg.MatchString(clean) {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		ret = append(ret, clean)
	}
	return ret
}

func SupervisorConfigList(c *gin.Context) {
	supervisorList, _ := common.DbMain.Client.QuickQuery(`tbl_supervisor`, `*`, nil).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`supervisor_list`: supervisorList,
	})
}

func getSupervisorComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshTerminal, error) {
	start := time.Now()
	sseClientId := c.GetHeader(`SseClientId`)
	dataMap := make(map[string]interface{})
	parseStart := time.Now()
	err := gsgin.GinPostBody(c, &dataMap)
	if err != nil {
		return nil, nil, err
	}
	gstool.FmtPrintlnLogTime(`[getSupervisorComponent] body parsed sse_client_id=%s cost_ms=%d`,
		sseClientId, time.Since(parseStart).Milliseconds())
	sshId := dataMap[`ssh_id`]
	if cast.ToString(sshId) == `` {
		return nil, nil, errors.New(`缺少ssh_id参数`)
	}
	sseDistributeId := cast.ToString(dataMap[`sse_distribute_id`])
	dbStart := time.Now()
	sshConfig, sshConfigErr := common.DbMain.GetSshConfig(sshId)
	gstool.FmtPrintlnLogTime(`[getSupervisorComponent] query ssh config ssh_id=%s sse_distribute_id=%s cost_ms=%d err=%v`,
		cast.ToString(sshId), sseDistributeId, time.Since(dbStart).Milliseconds(), sshConfigErr)
	if sshConfigErr != nil {
		return nil, nil, sshConfigErr
	}
	uniqueKey := p_common.TBaseClient.GetCombineKey(sshId, sseDistributeId)
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(sseClientId),
		SseDistributeId: sseDistributeId,
	}
	clientStart := time.Now()
	sshClient, sshClientErr := component.ShellClient.GetClient(sshConfig, uniqueKey, sse, func(s string) []string {
		return []string{p_common.TBaseClient.FilterTerminalChars(s)}
	}, nil, nil)
	gstool.FmtPrintlnLogTime(`[getSupervisorComponent] get ssh client ssh_id=%s unique_key=%s cost_ms=%d err=%v`,
		cast.ToString(sshId), uniqueKey, time.Since(clientStart).Milliseconds(), sshClientErr)
	if sshClientErr != nil {
		return nil, nil, sshClientErr
	}
	gstool.FmtPrintlnLogTime(`[getSupervisorComponent] success ssh_id=%s total_cost_ms=%d`,
		cast.ToString(sshId), time.Since(start).Milliseconds())
	return dataMap, sshClient, nil
}
