package controller

import (
	"dev_tool/base_module"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
)

// DockerRestart 重启
func DockerRestart(c *gin.Context) {
	_, reqMap, shell, err := getDockerReqData(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dockerName := reqMap[`DockerName`].ToStr()
	if dockerName == `` {
		gsgin.GinResponseError(c, `dockerName不能为空`, nil)
		return
	}
	restartCommand := base_module.NewCommand().Sudo().Cd(fmt.Sprintf(`cd /var/www/dockerfiles/dev_test/app/%s/`, dockerName))
	restartCommand = restartCommand.DockerRestart()
	ret, err := shell.RunCommandWait(restartCommand.GetCommand().ToStr())
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseError(c, ret, nil)
}

// DockerShowCompose 查看配置文件
func DockerShowCompose(c *gin.Context) {
	_, reqMap, shell, err := getDockerReqData(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dockerName := reqMap[`DockerName`].ToStr()
	if dockerName == `` {
		gsgin.GinResponseError(c, `dockerName不能为空`, nil)
		return
	}
	dockerComposeFilePath := fmt.Sprintf(`/var/www/dockerfiles/dev_test/app/%s/docker-compose.yml`, dockerName)
	showCommand := base_module.NewCommand().Sudo().ConsumerConfigCat(dockerComposeFilePath, dockerName)
	ret, err := shell.RunCommandWait(showCommand.GetCommand().ToStr())
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseError(c, ret, nil)
}

// DockerExec 执行命令
func DockerExec(c *gin.Context) {
	_, reqMap, shell, err := getDockerReqData(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dockerName := reqMap[`DockerName`].ToStr()
	if dockerName == `` {
		gsgin.GinResponseError(c, `dockerName不能为空`, nil)
		return
	}
	dockerCommand := reqMap[`DockerCommand`].ToStr()
	execCommand := base_module.NewCommand().Sudo().DockerExec(dockerName, dockerCommand)
	ret, err := shell.RunCommandWait(execCommand.GetCommand().ToStr())
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseError(c, ret, nil)
}

// DockerPs ps
func DockerPs(c *gin.Context) {
	_, _, shell, err := getDockerReqData(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	psCommand := base_module.NewCommand().Sudo().DockerPs()
	ret, err := shell.RunCommandWait(psCommand.GetCommand().ToStr())
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseError(c, ret, nil)
}

func getDockerReqData(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, *gsssh.SshConfig, error) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		return nil, nil, nil, err
	}
	shellName := reqMap[`ShellName`]
	if shellName == nil {
		return nil, nil, nil, errors.New(`缺少ShellName参数`)
	}
	client, err := global.ShellGet(shellName.ToStr())
	if err != nil {
		return nil, nil, nil, err
	}
	return global, reqMap, client, nil
}
