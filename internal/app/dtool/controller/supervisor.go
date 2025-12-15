package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_shell"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsdefine"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

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
	dockerName := cast.ToString(reqMap[`docker_name`])
	configPath := cast.ToString(reqMap[`config_path`])
	retMsgList := make([]string, 0)
	catCommand := p_shell.NewCommand().Sudo().ConsumerConfigCat(configPath, dockerName)
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
	if supervisorName == `` {
		gsgin.GinResponseError(c, `消费者name不能为空`, nil)
		return
	}
	restartCommand := p_shell.NewCommand().Sudo()
	restartCommand.ConsumerRestart(dockerName, supervisorName)
	restartCommand.ConsumerStatusGrep(dockerName, supervisorName)
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
	reqMap, sshClient, err := getSupervisorComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	retMsgList := make([]string, 0)
	dockerName := cast.ToString(reqMap[`docker_name`])
	configListCommand := p_shell.NewCommand()
	configListCommand.ConsumerConfigList(dockerName)
	ret, _ := sshClient.RunCommandWait(configListCommand.GetCommand().ToStr(), 40*time.Second)
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponseSuccess(c, ``, strings.Join(retMsgList, gsdefine.Enter))
	return
}

func SupervisorConfigList(c *gin.Context) {
	supervisorList, _ := common.DbMain.Client.QuickQuery(`tbl_supervisor`, `*`, nil).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`supervisor_list`: supervisorList,
	})
}

func getSupervisorComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshTerminal, error) {
	dataMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &dataMap)
	if err != nil {
		return nil, nil, err
	}
	sshId := dataMap[`ssh_id`]
	if cast.ToString(sshId) == `` {
		return nil, nil, errors.New(`缺少ssh_id参数`)
	}
	sseId := dataMap[`sse_id`]
	sshConfig, _ := common.DbMain.GetSshConfig(sshId)
	uniqueKey := p_common.TBaseClient.GetCombineKey(sshId, sseId)
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: cast.ToString(dataMap[`sse_id`]),
	}
	sshClient, sshClientErr := component.ShellClient.GetClient(sshConfig, uniqueKey, sse, nil)
	if sshClientErr != nil {
		return nil, nil, sshClientErr
	}
	return dataMap, sshClient, nil
}
