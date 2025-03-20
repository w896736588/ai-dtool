package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"errors"
	"gitee.com/Sxiaobai/gs/gsdefine"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gsssh"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"strings"
)

// SupervisorRestartAll 重启所有
func SupervisorRestartAll(c *gin.Context) {
	reqMap, sshClient, err := getSupervisorComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dockerName := cast.ToString(reqMap[`docker_name`])
	restartCommand := base.NewCommand().Sudo()
	if dockerName == `` { //非docker环境
		restartCommand.ConsumerRestartAll()
	} else {
		restartCommand.DockerExecConsumerRestartAll(dockerName)
	}
	_, _ = sshClient.RunCommandWait(restartCommand.GetCommand().ToStr())
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
	restartCommand := base.NewCommand().Sudo()
	if dockerName == `` { //非docker环境
		restartCommand.ConsumerStopAll()
	} else {
		restartCommand.DockerExecConsumerStopAll(dockerName)
	}
	_, _ = sshClient.RunCommandWait(restartCommand.GetCommand().ToStr())
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
func getConsumerStatus(dockerName string, sshClient *gsssh.SshConfig) ([]string, error) {
	//消费者
	retMsgList := make([]string, 0)
	if dockerName == `` { //非docker环境
		statusCommand := base.NewCommand().Sudo().ConsumerStatus()
		statusRet, _ := sshClient.RunCommandWait(statusCommand.GetCommand().ToStr())
		retMsgList = append(retMsgList, statusRet)
	} else {
		xkfStatusCommand := base.NewCommand().Sudo()
		xkfStatusCommand.DockerExecConsumerStatus(dockerName)
		xkfStatusRet, _ := sshClient.RunCommandWait(xkfStatusCommand.GetCommand().ToStr())
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
	catCommand := base.NewCommand().Sudo().ConsumerConfigCat(configPath, dockerName)
	ret, _ := sshClient.RunCommandWait(catCommand.GetCommand().ToStr())
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
	restartCommand := base.NewCommand().Sudo()
	restartCommand.ConsumerRestart(dockerName, supervisorName)
	restartCommand.ConsumerStatusGrep(dockerName, supervisorName)
	ret, _ := sshClient.RunCommandWait(restartCommand.GetCommand().ToStr())
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
	stopCommand := base.NewCommand().Sudo()
	stopCommand.ConsumerStop(dockerName, supervisorName)
	stopCommand.ConsumerStatusGrep(dockerName, supervisorName)
	ret, _ := sshClient.RunCommandWait(stopCommand.GetCommand().ToStr())
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
	configListCommand := base.NewCommand()
	configListCommand.ConsumerConfigList(dockerName)
	ret, _ := sshClient.RunCommandWait(configListCommand.GetCommand().ToStr())
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponseSuccess(c, ``, strings.Join(retMsgList, gsdefine.Enter))
	return
}

func SupervisorConfigList(c *gin.Context) {
	supervisorList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_supervisor`, `*`, nil).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`supervisor_list`: supervisorList,
	})
}

func getSupervisorComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshConfig, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, nil, err
	}
	sshId := reqMap[`ssh_id`]
	if cast.ToString(sshId) == `` {
		return nil, nil, errors.New(`缺少ssh_id参数`)
	}
	sshConfig, _ := base.Component.TSqlite.GetSshConfig(sshId)
	uniqueKey := base.Component.TBase.GetCombineKey(sshId, `supervisor`)
	sshClient, sshClientErr := base.Component.TShell.GetClient(sshConfig, uniqueKey, define.SseSupervisor)
	if sshClientErr != nil {
		return nil, nil, err
	}
	return reqMap, sshClient, nil
}
