package controller

import (
	"dev_tool/base"
	"errors"
	"time"

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
	id, err := base.Component.TSqlite.Client.QuickCreate(`tbl_shell_out`, map[string]any{
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
	id, err := base.Component.TSqlite.Client.QuickUpdate(`tbl_shell_out`, map[string]any{
		`id`: reqMap[`id`],
	}, map[string]any{
		`name`:        cast.ToString(reqMap[`name`]),
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
	_, err = base.Component.TSqlite.Client.QuickUpdate(`tbl_shell_out`, map[string]any{
		`id`: reqMap[`id`],
	}, map[string]any{
		`is_run`:      1,
		`update_time`: time.Now().Unix(),
	}).Exec()
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

func GetShellOuts(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	list, err := base.Component.TSqlite.Client.QuickQuery(`tbl_shell_out`, `*`, nil).Order(`id asc`).All()
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
	sc := base.Component.TSqlite.Client.QuickDelete(`tbl_shell_out`, map[string]any{
		`id`: reqMap[`id`],
	})
	_, err = sc.Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), map[string]any{
			`sql`: sc.GetSql()})
		return
	}
	shellClientId := cast.ToString(reqMap[`shell_client_id`])
	base.Component.TShellOut.Delete(shellClientId)
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
	base.Component.TShellOut.Delete(shellClientId)
	gsgin.GinResponseSuccess(c, ``, nil)
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
