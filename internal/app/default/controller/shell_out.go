package controller

import (
	"dev_tool/base"
	"errors"

	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gsssh"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func ShellOut(c *gin.Context) {
	reqMap, client, shellClientId, err := getShellOutComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	command := cast.ToString(reqMap[`command`])
	_ = client.RunCommand(command)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`shell_client_id`: shellClientId,
	})
	return
}

func ShellOutSetSeeId(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	sseId := cast.ToString(reqMap[`sse_id`])
	sshId := cast.ToString(reqMap[`ssh_id`])
	command := cast.ToString(reqMap[`command`])
	err = base.Component.TShellOut.SetClientSseId(shellClientId, sshId, sseId, command, nil)
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
	base.Component.TShellOut.CleanErrors(shellClientId)
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
	return
}

func getShellOutComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshConfig, string, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, nil, ``, err
	}
	sshId := reqMap[`ssh_id`]
	if cast.ToString(sshId) == `` {
		return nil, nil, ``, errors.New(`缺少ssh_id参数`)
	}
	sseId := reqMap[`sse_id`]
	sshConfig, _ := base.Component.TSqlite.GetSshConfig(sshId)
	shellClientId := base.Component.TBase.GetUnique(`shell_out_`)
	shellOut, _, sshClientErr := base.Component.TShellOut.GetClient(sshConfig, shellClientId, cast.ToString(sseId), nil)
	if sshClientErr != nil {
		return nil, nil, ``, sshClientErr
	}
	return reqMap, shellOut.Client, shellClientId, nil
}
