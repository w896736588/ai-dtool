package controller

import (
	"context"
	"dev_tool/base_module"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"time"
)

//BaseLogin 登录
func BaseLogin(c *gin.Context) {
	reqBody := &base_module.LoginStruct{}
	err := gsgin.GinPostBody(c, reqBody)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, err.Error(), nil)
		return
	}
	unikey := gstool.Md5(reqBody.UserName + reqBody.Password + cast.ToString(time.Now().UnixMicro()))
	base_module.CreateGlobal(unikey)
	gsgin.GinResponse(c, gsgin.ResponseSuccess, `获取成功`, map[string]string{
		`unikey`: unikey,
	})
}

//BaseCheckUnikeyExist 检查是否需要登录
func BaseCheckUnikeyExist(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, err.Error(), nil)
		return
	}
	reqConsMap := gstool.ConsNewMap(reqMap)
	unikey := reqConsMap[`Unikey`]
	if unikey.IsEmpty() {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, `Unikey不能为空`, nil)
		return
	}

	global, err := GetGlobal(reqConsMap)
	if global == nil {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, map[string]string{
			`NeedLogin`: `1`,
		})
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, `获取成功`, map[string]string{
		`NeedLogin`: `0`,
	})
}

//BaseRegisterService 注册各类服务
func BaseRegisterService(c *gin.Context) {
	reqBody := &base_module.RegisterStruct{}
	err := gsgin.GinPostBody(c, reqBody)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, err.Error(), nil)
		return
	}
	global := base_module.GetGlobal(reqBody.Unikey)
	if global == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `请登录`, nil)
		return
	}
	base_module.Register(global, reqBody)
	gsgin.GinResponse(c, gsgin.ResponseSuccess, `ok`, nil)
}

//GetGlobalReqParams 拿到全局参数
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

//GetGlobalReqParamsM 拿到全局参数 返回map
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

//BaseRedisGetReqDataRedis 基础方法
func BaseRedisGetReqDataRedis(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, *gsdb.GsRedis, error) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		return nil, nil, nil, err
	}
	redisName := reqMap[`RedisName`].ToStr()
	gstool.FmtPrintlnLog(`初始化 %#v`, reqMap)
	if redisName == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少RedisName参数`, nil)
		return nil, nil, nil, errors.New(`缺少RedisName参数`)
	}
	client, err := global.RedisGetClient(redisName)
	if err != nil {
		return nil, nil, nil, err
	}

	return global, reqMap, client, nil
}

//BaseRedisGetReqDataRedisM 基础方法
func BaseRedisGetReqDataRedisM(c *gin.Context) (*base_module.Global, map[string]interface{}, *gsdb.GsRedis, error) {
	global, reqMap, err := GetGlobalReqParamsM(c)
	if err != nil {
		return nil, nil, nil, err
	}
	redisName := cast.ToString(reqMap[`RedisName`])
	gstool.FmtPrintlnLog(`初始化 %#v`, reqMap)
	if redisName == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少RedisName参数`, nil)
		return nil, nil, nil, errors.New(`缺少RedisName参数`)
	}
	client, err := global.RedisGetClient(redisName)
	if err != nil {
		return nil, nil, nil, err
	}

	return global, reqMap, client, nil
}

//BaseRedisCheckKeyExist 基础方法
func BaseRedisCheckKeyExist(redisCli *redis.Client, key string) error {
	//判断是否存在
	if existInt := redisCli.Exists(context.Background(), key).Val(); existInt <= 0 {
		return errors.New(fmt.Sprintf(`%s 不存在`, key))
	}
	return nil
}

func BaseResponseByError(c *gin.Context, err error) {
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), ``)
	} else {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, ``)
	}
}
