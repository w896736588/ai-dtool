package controller

import (
	"dev_tool/base_module"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
)

//DockerRestart 重启
func DockerRestart(c *gin.Context) {
	_, reqMap, shell, err := getDockerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	dockerName := reqMap[`DockerName`].ToStr()
	if dockerName == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, `dockerName不能为空`, nil)
		return
	}
	restartCommand := base_module.NewCommand().Sudo().Cd(fmt.Sprintf(`cd /var/www/dockerfiles/dev_test/app/%s/`, dockerName))
	restartCommand = restartCommand.DockerRestart()
	ret, err := shell.RunShell3(restartCommand.GetCommand().ToByte())
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseError, ret, nil)
}

//DockerShowCompose 查看配置文件
func DockerShowCompose(c *gin.Context) {
	_, reqMap, shell, err := getDockerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	dockerName := reqMap[`DockerName`].ToStr()
	if dockerName == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, `dockerName不能为空`, nil)
		return
	}
	dockerComposeFilePath := fmt.Sprintf(`/var/www/dockerfiles/dev_test/app/%s/docker-compose.yml`, dockerName)
	showCommand := base_module.NewCommand().Sudo().ConsumerConfigCat(dockerComposeFilePath, dockerName)
	ret, err := shell.RunShell3(showCommand.GetCommand().ToByte())
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseError, ret, nil)
}

//DockerExec 执行命令
func DockerExec(c *gin.Context) {
	_, reqMap, shell, err := getDockerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	dockerName := reqMap[`DockerName`].ToStr()
	if dockerName == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, `dockerName不能为空`, nil)
		return
	}
	dockerCommand := reqMap[`DockerCommand`].ToStr()
	if dockerName == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, `DockerCommand不能为空`, nil)
		return
	}
	execCommand := base_module.NewCommand().Sudo().DockerExec(dockerName, dockerCommand)
	ret, err := shell.RunShell3(execCommand.GetCommand().ToByte())
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseError, ret, nil)
}

//DockerPs ps
func DockerPs(c *gin.Context) {
	_, _, shell, err := getDockerReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	psCommand := base_module.NewCommand().Sudo().DockerPs()
	ret, err := shell.RunShell3(psCommand.GetCommand().ToByte())
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseError, ret, nil)
}

func getDockerReqData(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, *gstool.GsShell, error) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		return nil, nil, nil, err
	}
	shellName := reqMap[`ShellName`]
	if shellName == nil {
		return nil, nil, nil, errors.New(`缺少ShellName参数`)
	}
	client, err := global.ShellGetClient(shellName.ToStr())
	if err != nil {
		return nil, nil, nil, err
	}
	return global, reqMap, client, nil
}
