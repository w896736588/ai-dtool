package controller

import (
	"context"
	"dev_tool/base"
	"dev_tool/base_module"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gs"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
)

// BaseLogin 登录
func BaseLogin(c *gin.Context) {
	reqBody := &base_module.LoginStruct{}
	err := gsgin.GinPostBody(c, reqBody)
	if err != nil {
		gsgin.GinResponseSuccess(c, err.Error(), nil)
		return
	}
	userId, loginErr := base.Component.TSqlite.Login(reqBody.UserName, reqBody.Password)
	if loginErr != nil {
		gsgin.GinResponseError(c, `登录失败（`+loginErr.Error()+`）`, map[string]string{
			`NeedLogin`: `1`,
			`unikey`:    ``,
			`token`:     ``,
		})
		return
	}
	token, tokenErr := base.Component.AesGcm.Encrypt(gs.NewGs(userId).ToByte())
	if tokenErr != nil {
		gsgin.GinResponseError(c, `登录失败（`+tokenErr.Error()+`）`, map[string]string{
			`NeedLogin`: `1`,
			`unikey`:    ``,
			`token`:     ``,
		})
	}
	base_module.CreateGlobal(token)
	gsgin.GinResponseSuccess(c, `获取成功`, map[string]string{
		`unikey`: token,
		`token`:  token,
	})
}

// BaseCheckUnikeyExist 检查是否需要登录
func BaseCheckUnikeyExist(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseSuccess(c, err.Error(), nil)
		return
	}
	reqConsMap := gstool.ConsNewMap(reqMap)
	unikey := reqConsMap[`Unikey`]
	if unikey.IsEmpty() {
		gsgin.GinResponseSuccess(c, `Unikey不能为空`, nil)
		return
	}

	global, err := GetGlobal(reqConsMap)
	if global == nil {
		gsgin.GinResponseSuccess(c, ``, map[string]string{
			`NeedLogin`: `1`,
		})
		return
	}
	gsgin.GinResponseSuccess(c, `获取成功`, map[string]string{
		`NeedLogin`: `0`,
	})
}

// BaseRegisterService 注册各类服务
func BaseRegisterService(c *gin.Context) {
	reqBody := &base_module.RegisterStruct{}
	err := gsgin.GinPostBody(c, reqBody)
	if err != nil {
		gsgin.GinResponseSuccess(c, err.Error(), nil)
		return
	}
	global := base_module.GetGlobal(reqBody.Unikey)
	if global == nil {
		gsgin.GinResponseError(c, `请登录`, nil)
		return
	}
	base_module.Register(global, reqBody)
	gsgin.GinResponseSuccess(c, `ok`, nil)
}

// GetGlobalReqParams 拿到全局参数
func GetGlobalReqParams(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, nil, err
	}
	reqConsMap := gstool.ConsNewMap(reqMap)
	global, err := GetGlobal(reqConsMap)
	if err != nil {
		return nil, nil, err
	}
	return global, reqConsMap, nil
}

// GetGlobalReqParamsM 拿到全局参数 返回map
func GetGlobalReqParamsM(c *gin.Context) (*base_module.Global, map[string]interface{}, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, nil, err
	}
	global, err := GetGlobalM(reqMap)
	if err != nil {
		return nil, nil, err
	}
	return global, reqMap, nil
}

// BaseRedisGetReqDataRedis 基础方法
func BaseRedisGetReqDataRedis(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, *gsdb.GsRedis, error) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		return nil, nil, nil, err
	}
	redisName := reqMap[`RedisName`].ToStr()
	if redisName == `` {
		gsgin.GinResponseError(c, `缺少RedisName参数`, nil)
		return nil, nil, nil, errors.New(`缺少RedisName参数`)
	}
	client, err := global.RedisGetClient(redisName)
	if err != nil {
		return nil, nil, nil, err
	}

	return global, reqMap, client, nil
}

// BaseRedisCheckKeyExist 基础方法
func BaseRedisCheckKeyExist(redisCli *redis.Client, key string) error {
	//判断是否存在
	if existInt := redisCli.Exists(context.Background(), key).Val(); existInt <= 0 {
		return errors.New(fmt.Sprintf(`%s 不存在`, key))
	}
	return nil
}

func BaseResponseByError(c *gin.Context, err error) {
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), ``)
	} else {
		gsgin.GinResponseSuccess(c, ``, ``)
	}
}

// GetGlobalComponent 根据提交的参数获取各个组件 傻瓜化
func GetGlobalComponent(c *gin.Context) (base_module.Component, error) {
	component := base_module.Component{}
	global, reqMap, globalErr := GetGlobalReqParamsM(c)
	if globalErr != nil {
		gstool.FmtPrintlnLog(`获取参数失败 %s`, globalErr.Error())
		return component, globalErr
	}
	component.Global = global
	component.ReqMap = reqMap
	component.Encrypt = global.GetEncrypt()
	for requestParamKey, componentKey := range reqMap {
		switch requestParamKey {
		case `ShellName`:
			client, shellErr := global.ShellGet(cast.ToString(componentKey))
			if shellErr != nil {
				return component, shellErr
			}
			component.ShellClient = client
		case `RedisName`:
			redisCli, redisCliErr := global.RedisGetClient(cast.ToString(componentKey))
			if redisCliErr != nil {
				return component, redisCliErr
			}
			component.RedisClient = redisCli
		case `XkfMysqlName`, `mysqlName`:
			xkfMysqlCli, xkfMysqlCliErr := global.MysqlGetClient(cast.ToString(componentKey))
			if xkfMysqlCliErr != nil {
				return component, xkfMysqlCliErr
			}
			component.XkfMysqlClient = xkfMysqlCli
		case `AppUrlMysql`:
			appUrlMysqlCli, appUrlMysqlCliErr := global.MysqlGetClient(cast.ToString(componentKey))
			if appUrlMysqlCliErr != nil {
				return component, appUrlMysqlCliErr
			}
			component.AppUrlMysqlClient = appUrlMysqlCli
		}
	}
	component.Logger = global.GetLogger()
	return component, nil
}

func BaseSshList(c *gin.Context) {
	sshList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`ssh_list`: sshList,
	})
}
