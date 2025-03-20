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
	"path"
	"strings"
)

func DockerComposeList(c *gin.Context) {
	_, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	all, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`status`: 1,
	}).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	for k, v := range all {
		composeYmlPath := v[`compose_yml_path`].(string)
		command := base.NewCommand()
		command.Sudo()
		command.Cd(path.Dir(composeYmlPath))
		command.DockerComposePs()
		result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr())
		all[k][`result`] = strings.Join(strings.Split(result, "\n"), "<br/>")
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: all,
	})
}

func DockerComposeConfigShow(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	catCommand := base.NewCommand().Sudo().ConsumerConfigCat(cast.ToString(data[`config_path`]), ``)
	ret, _ := sshClient.RunCommandWait(catCommand.GetCommand().ToStr())
	retMsgList := make([]string, 0)
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponseSuccess(c, ``, strings.Join(retMsgList, gsdefine.Enter))
}

func DockerComposeRestart(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	id := cast.ToInt(data[`id`])
	if id == 0 {
		gsgin.GinResponseError(c, `id is empty`, nil)
		return
	}
	one, oneErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	composeYmlPath := one[`compose_yml_path`].(string)
	command := base.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	command.DockerComposeRestart()
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

func DockerComposeStop(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	id := cast.ToInt(data[`id`])
	if id == 0 {
		gsgin.GinResponseError(c, `id is empty`, nil)
		return
	}
	one, oneErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	composeYmlPath := one[`compose_yml_path`].(string)
	command := base.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	command.DockerComposeStop()
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

func DockerComposeStart(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	id := cast.ToInt(data[`id`])
	if id == 0 {
		gsgin.GinResponseError(c, `id is empty`, nil)
		return
	}
	one, oneErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	composeYmlPath := one[`compose_yml_path`].(string)
	command := base.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	command.DockerComposeStart()
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

func getDockerComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshConfig, error) {
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
	uniqueKey := base.Component.TBase.GetCombineKey(sshId, `compose`)
	sshClient, sshClientErr := base.Component.TShell.GetClient(sshConfig, uniqueKey, define.SseDocker)
	if sshClientErr != nil {
		return nil, nil, err
	}
	return reqMap, sshClient, nil
}
