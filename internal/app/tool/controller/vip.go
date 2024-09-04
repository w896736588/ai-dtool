package controller

import (
	"context"
	"dev_tool/internal/app/default/controller"
	"dev_tool/internal/app/zhima/service"
	"errors"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var VipMap = map[string]string{
	`0`: `免费版`,
	`1`: `专业版`,
	`2`: `企业版`,
	`3`: `标准版`,
	`4`: `平台版`,
}

// VipChange vip版本切换
func VipChange(c *gin.Context) {
	reqMap, redisCli, mysqlCli, err := getVipReqData(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	account := reqMap[`Account`]
	if account == nil {
		gsgin.GinResponseError(c, `账号不能为空`, nil)
		return
	}
	userInfo := service.GetAdminUserId(mysqlCli, cast.ToString(account))
	if userInfo[`_id`] == nil {
		gsgin.GinResponseError(c, `找不到该账号`, nil)
		return
	}
	_, upErr := service.UpdateVip(mysqlCli, cast.ToString(userInfo[`_id`]), cast.ToString(reqMap[`ExpireDay`]), cast.ToString(reqMap[`SystemType`]), cast.ToString(reqMap[`VipLevel`]))
	if upErr != nil {
		gsgin.GinResponseError(c, upErr.Error(), nil)
		return
	}
	//移除缓存
	adminUserId := cast.ToInt(userInfo[`_id`])
	number := cast.ToString(adminUserId % 10)
	redisCli.Client.HDel(context.Background(), `wechatapp.vip.info.v20220308..`+number, cast.ToString(adminUserId))
	redisCli.Client.HDel(context.Background(), `wechatapp.kefu.vip.info.v20220308..`+number, cast.ToString(adminUserId))
	result, resultErr := queryVipType(reqMap, mysqlCli)
	if resultErr != nil {
		gsgin.GinResponseError(c, resultErr.Error(), nil)
		return
	} else {
		gsgin.GinResponseSuccess(c, ``, result)
		return
	}
}

// VipQuery vip版本查询
func VipQuery(c *gin.Context) {
	reqMap, _, mysqlCli, err := getVipReqData(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	result, resultErr := queryVipType(reqMap, mysqlCli)
	if resultErr != nil {
		gsgin.GinResponseError(c, resultErr.Error(), nil)
		return
	} else {
		gsgin.GinResponseSuccess(c, ``, result)
		return
	}
}

// QueryVipType 查询VIP版本
func queryVipType(reqMap map[string]interface{}, mysqlCli *gsdb.GsMysql) (string, error) {

	account := reqMap[`Account`]
	if account == nil {
		return ``, errors.New(`账号不能为空`)
	}
	userInfo := service.GetAdminUserId(mysqlCli, cast.ToString(account))

	if userInfo[`_id`] == nil {
		return ``, errors.New(`找不到该账号`)
	}
	adminUserIdStr := cast.ToString(userInfo[`_id`])
	vipInfo, queryErr := service.QueryVip(mysqlCli, cast.ToString(userInfo[`_id`]), cast.ToString(reqMap[`SystemType`]))
	if queryErr != nil {
		return ``, queryErr
	}
	if len(vipInfo) == 0 {
		return `管理员ID：` + adminUserIdStr + `未查到vip信息`, nil
	}
	return `管理员ID：` + adminUserIdStr + `，vip版本：` + VipMap[cast.ToString(vipInfo[`vip_type`])] + `，过期时间：` + cast.ToString(vipInfo[`expired_time`]), nil
}

// 拿到各类句柄
func getVipReqData(c *gin.Context) (map[string]interface{}, *gsdb.GsRedis, *gsdb.GsMysql, error) {
	component, componentErr := controller.GetGlobalComponent(c)
	if componentErr != nil {
		return nil, nil, nil, componentErr
	}
	if component.RedisClient == nil {
		return nil, nil, nil, errors.New(`redis客户端为空`)
	}
	if component.XkfMysqlClient == nil {
		return nil, nil, nil, errors.New(`mysql客户端为空`)
	}
	return component.ReqMap, component.RedisClient, component.XkfMysqlClient, nil
}
