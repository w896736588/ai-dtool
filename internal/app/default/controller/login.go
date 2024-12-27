package controller

import (
	"dev_tool/base"
	"dev_tool/internal/app/zhima/service"
	"gitee.com/Sxiaobai/gs/gsencrypt"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"net/url"
	"strings"
	"time"
)

// LoginLink 登录地址
func LoginLink(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	mysqlConfig, mysqlConfigErr := base.Component.TSqlite.GetMysqlConfig(8)
	if mysqlConfigErr != nil {
		gsgin.GinResponseError(c, mysqlConfigErr.Error(), nil)
		return
	}
	mysqlCliXkf, mysqlCliXkfErr := base.Component.TMysql.GetClient(mysqlConfig)
	if mysqlCliXkfErr != nil {
		gsgin.GinResponseError(c, mysqlCliXkfErr.Error(), nil)
		return
	}
	account := reqMap[`Account`]
	if account == nil {
		gsgin.GinResponseError(c, `账号不能为空`, nil)
		return
	}
	userInfo := service.GetAdminUserId(mysqlCliXkf, cast.ToString(account))
	if userInfo[`_id`] == nil {
		gsgin.GinResponseError(c, `找不到该账号`, nil)
		return
	}
	loginHost := cast.ToString(reqMap[`LoginHost`])
	encrypt := &gsencrypt.DesCbc{
		Key: `zima`,
		Iv:  `14554552`,
	}
	//拿到一个应用ID和一个渠道ID
	wechatAppId, channelId, errQuery := service.QueryOneWechatAppIdChannelId(mysqlCliXkf, cast.ToInt(userInfo[`_id`]))
	if errQuery != nil {
		gsgin.GinResponseError(c, errQuery.Error(), nil)
		return
	}
	redirectUrl := cast.ToString(reqMap[`LoginUrl`])
	redirectUrl = strings.Replace(redirectUrl, `{wechatapp_id}`, wechatAppId, -1)
	redirectUrl = strings.Replace(redirectUrl, `{channel_id}`, channelId, -1)
	token := gstool.JsonEncode(map[string]string{
		`login_type`: `1`,
		`user_id`:    cast.ToString(userInfo[`_id`]),
		`param`: gstool.JsonEncode(map[string]string{
			`uri`: redirectUrl,
		}),
		`time`: cast.ToString(time.Now().Unix()), //仅10秒内有效
	})
	data, err := encrypt.Encrypt(token)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	token = url.QueryEscape(data)
	gsgin.GinResponseSuccess(c, ``, loginHost+`index/LoginRedirect?token=`+token)
	return
}
