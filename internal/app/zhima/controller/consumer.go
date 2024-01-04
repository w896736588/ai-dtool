package controller

import (
	"dev_tool/base_module"
	"errors"
	"gitee.com/Sxiaobai/gs/gsdefine"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"strings"
)

//ConsumerRestartAll 重启所有
func ConsumerRestartAll(c *gin.Context) {
	_, reqMap, shell, err := getConsumerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	dockerName := cast.ToString(reqMap[`DockerName`])
	restartCommand := base_module.NewCommand().Sudo()
	if dockerName == `` { //非docker环境
		restartCommand.ConsumerRestartAll()
	} else {
		restartCommand.DockerExecConsumerRestartAll(dockerName)
	}
	_ = shell.RunShell(restartCommand.GetCommand().ToByte())
	statusRet, err := getConsumerStatus(dockerName, reqMap, shell)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, strings.Join(statusRet, gsdefine.Enter))
}

//ConsumerStopAll 停止所有
func ConsumerStopAll(c *gin.Context) {
	_, reqMap, shell, err := getConsumerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	dockerName := cast.ToString(reqMap[`DockerName`])
	restartCommand := base_module.NewCommand().Sudo().Cd(cast.ToString(reqMap[`CodePath`]))
	if dockerName == `` { //非docker环境
		restartCommand.ConsumerStopAll()
	} else {
		restartCommand.DockerExecConsumerStopAll(dockerName)
	}
	_ = shell.RunShell(restartCommand.GetCommand().ToByte())
	statusRet, err := getConsumerStatus(dockerName, reqMap, shell)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, strings.Join(statusRet, gsdefine.Enter))
}

// ConsumerStatusList 消费者列表 通过supervisorctl status获取
func ConsumerStatusList(c *gin.Context) {
	_, reqMap, shell, err := getConsumerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	dockerName := cast.ToString(reqMap[`DockerName`])
	statusRet, err := getConsumerStatus(dockerName, reqMap, shell)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, strings.Join(statusRet, gsdefine.Enter))
}

//拿到消费者状态 支持docker与非docker
func getConsumerStatus(dockerName string, reqMap map[string]interface{}, shell *gstool.GsShellPush) ([]string, error) {
	//消费者
	retMsgList := make([]string, 0)
	if dockerName == `` { //非docker环境
		statusCommand := base_module.NewCommand().Sudo().ConsumerStatus()
		statusRet := shell.RunShell(statusCommand.GetCommand().ToByte())
		retMsgList = append(retMsgList, statusRet)
	} else {
		xkfStatusCommand := base_module.NewCommand().Sudo()
		xkfStatusCommand.DockerExecConsumerStatus(dockerName)
		xkfStatusRet := shell.RunShell(xkfStatusCommand.GetCommand().ToByte())
		retMsgList = append(retMsgList, xkfStatusRet)
	}
	return retMsgList, nil
}

// ConsumerConfigShow 查看supervisor配置内容
func ConsumerConfigShow(c *gin.Context) {
	_, reqMap, shell, err := getConsumerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	retMsgList := make([]string, 0)
	catCommand := base_module.NewCommand().Sudo().ConsumerConfigCat(cast.ToString(reqMap[`SupervisorConfigPath`]), cast.ToString(reqMap[`DockerName`]))
	ret := shell.RunShell(catCommand.GetCommand().ToByte())
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, strings.Join(retMsgList, gsdefine.Enter))
	return
}

//ConsumerRestart 重启某一个消费者
func ConsumerRestart(c *gin.Context) {
	_, reqMap, shell, err := getConsumerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	retMsgList := make([]string, 0)
	dockerName := cast.ToString(reqMap[`DockerName`])
	consumerName := cast.ToString(reqMap[`ConsumerName`])
	restartCommand := base_module.NewCommand().Sudo()
	restartCommand.ConsumerRestart(dockerName, consumerName)
	restartCommand.ConsumerStatusGrep(dockerName, consumerName)
	ret := shell.RunShell(restartCommand.GetCommand().ToByte())
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, strings.Join(retMsgList, gsdefine.Enter))
	return
}

//ConsumerStop 停止某一个消费者
func ConsumerStop(c *gin.Context) {
	_, reqMap, shell, err := getConsumerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	retMsgList := make([]string, 0)
	dockerName := cast.ToString(reqMap[`DockerName`])
	consumerName := cast.ToString(reqMap[`ConsumerName`])
	restartCommand := base_module.NewCommand().Sudo()
	restartCommand.ConsumerStop(dockerName, consumerName)
	restartCommand.ConsumerStatusGrep(dockerName, consumerName)
	ret := shell.RunShell(restartCommand.GetCommand().ToByte())
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, strings.Join(retMsgList, gsdefine.Enter))
	return
}

//ConsumerConfigList 获取配置列表
func ConsumerConfigList(c *gin.Context) {
	_, reqMap, shell, err := getConsumerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	retMsgList := make([]string, 0)
	dockerName := cast.ToString(reqMap[`DockerName`])
	configListCommand := base_module.NewCommand()
	configListCommand.ConsumerConfigList(dockerName)
	ret := shell.RunShell(configListCommand.GetCommand().ToByte())
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, strings.Join(retMsgList, gsdefine.Enter))
	return
}

func getConsumerReqData(c *gin.Context) (*base_module.Global, map[string]interface{}, *gstool.GsShellPush, error) {
	global, reqMap, err := GetGlobalReqParamsM(c)
	if err != nil {
		return nil, nil, nil, err
	}
	shellName := cast.ToString(reqMap[`ShellName`])
	if shellName == `` {
		return nil, nil, nil, errors.New(`缺少ShellName参数`)
	}
	client, err := global.ShellPushGetClient(shellName)
	if err != nil {
		return nil, nil, nil, err
	}
	return global, reqMap, client, nil
}
