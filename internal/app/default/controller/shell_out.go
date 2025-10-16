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
	reqMap, client, uniqueKey, err := getShellOutComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	command := cast.ToString(reqMap[`command`])
	_ = client.RunCommand(command)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`conn_unique_key`: uniqueKey,
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
	connUniqueKey := cast.ToString(reqMap[`conn_unique_key`])
	sseId := cast.ToString(reqMap[`sse_id`])
	sshId := cast.ToString(reqMap[`ssh_id`])
	command := cast.ToString(reqMap[`command`])
	err = base.Component.TShellOut.SetClientSseId(connUniqueKey, sshId, sseId, command, nil)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
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
	uniqueKey := base.Component.TBase.GetCombineKey(sshId, sseId)
	sshClient, _, sshClientErr := base.Component.TShellOut.GetClient(sshConfig, uniqueKey, cast.ToString(sseId), nil)
	if sshClientErr != nil {
		return nil, nil, ``, sshClientErr
	}
	return reqMap, sshClient, uniqueKey, nil
}
