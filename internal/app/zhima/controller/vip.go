package controller

import (
	"context"
	"dev_tool/base_module"
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

//VipChange vip版本切换
func VipChange(c *gin.Context) {
	_, reqMap, redisCli, mysqlCli, err := getVipReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	account := reqMap[`Account`]
	if account == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `账号不能为空`, nil)
		return
	}
	userInfo := service.GetAdminUserId(mysqlCli, cast.ToString(account))
	if userInfo[`_id`] == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `找不到该账号`, nil)
		return
	}
	_, upErr := service.UpdateVip(mysqlCli, cast.ToString(userInfo[`_id`]), cast.ToString(reqMap[`ExpireDay`]), cast.ToString(reqMap[`SystemType`]), cast.ToString(reqMap[`VipLevel`]))
	if upErr != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, upErr.Error(), nil)
		return
	}
	//移除缓存
	adminUserId := cast.ToInt(userInfo[`_id`])
	number := cast.ToString(adminUserId % 10)
	redisCli.Client.HDel(context.Background(), `wechatapp.vip.info.v20220308..`+number, cast.ToString(adminUserId))
	redisCli.Client.HDel(context.Background(), `wechatapp.kefu.vip.info.v20220308..`+number, cast.ToString(adminUserId))
	result, resultErr := queryVipType(c)
	if resultErr != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, resultErr.Error(), nil)
		return
	} else {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, result, nil)
		return
	}
}

//VipQuery vip版本查询
func VipQuery(c *gin.Context) {
	result, resultErr := queryVipType(c)
	if resultErr != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, resultErr.Error(), nil)
		return
	} else {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, result, nil)
		return
	}
}

// QueryVipType 查询VIP版本
func queryVipType(c *gin.Context) (string, error) {
	_, reqMap, _, mysqlCli, err := getVipReqData(c)
	if err != nil {
		return ``, err
	}
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

func getVipReqData(c *gin.Context) (*base_module.Global, map[string]interface{}, *gsdb.GsRedis, *gsdb.GsMysql, error) {
	global, reqMap, err := GetGlobalReqParamsM(c)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	redisName := cast.ToString(reqMap[`redisName`])
	redisCli, err := global.RedisGetClient(redisName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	mysqlName := cast.ToString(reqMap[`mysqlName`])
	mysqlCli, err := global.MysqlGetClient(mysqlName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return global, reqMap, redisCli, mysqlCli, nil
}
