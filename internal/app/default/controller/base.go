package controller

import (
	"context"
	"dev_tool/base"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"strings"

	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
)

// BaseLogin 登录
func BaseLogin(c *gin.Context) {
	reqBody := &_struct.LoginStruct{}
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
	token, tokenErr := base.Component.AesGcm.Encrypt([]byte(cast.ToString(userId)))
	if tokenErr != nil {
		gsgin.GinResponseError(c, `登录失败（`+tokenErr.Error()+`）`, map[string]string{
			`NeedLogin`: `1`,
			`unikey`:    ``,
			`token`:     ``,
		})
	}
	gsgin.GinResponseSuccess(c, `获取成功`, map[string]any{
		`unikey`: token,
		`token`:  token,
		`ports`:  strings.Split(base.Component.ConfigViper.GetString(`run.ports`), `,`),
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

	gsgin.GinResponseSuccess(c, `获取成功`, map[string]string{
		`NeedLogin`: `0`,
	})
}

// BaseRegisterService 注册各类服务
func BaseRegisterService(c *gin.Context) {
	gsgin.GinResponseSuccess(c, `ok`, nil)
}

// GetGlobalReqParamsM 拿到全局参数 返回map
func GetGlobalReqParamsM(c *gin.Context) (map[string]interface{}, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, err
	}
	return reqMap, nil
}

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

func BaseSshList(c *gin.Context) {
	sshList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`ssh_list`: sshList,
	})
}

// Ip 外网IP
func Ip(c *gin.Context) {
	ip, _ := base.Component.TBase.GetPublicIPWithSTUN()
	gsgin.GinResponseSuccess(c, `获取成功`, map[string]string{
		`ip`: ip,
	})
}
