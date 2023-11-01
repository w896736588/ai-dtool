package controller

import (
	"dev_tool/base_module"
	"dev_tool/internal/app/xkf_tool"
	"dev_tool/internal/app/zhima/service"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"strings"
)

// WechatKefuStatus 查询微信客服应用的状态
func WechatKefuStatus(c *gin.Context) {
	_, reqMap, shell, _, mysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	appInfo := service.QueryWechatAppid(mysqlCli, reqMap[`WechatKefuAppid`].ToStr())
	appIdStr := cast.ToString(appInfo[`app_id`])
	userIdStr := cast.ToString(appInfo[`user_id`])
	appTypeStr := cast.ToString(appInfo[`app_type`])
	if appIdStr == `` || appTypeStr != `wechat_kefu` {
		BaseResponseByError(c, errors.New(`找不到微信客服应用`))
		return
	}
	command := base_module.NewCommand().Sudo().WechatKefuStatus(appIdStr)

	RunResultMsg, err := shell.RunShell3(command.GetCommand().ToByte())
	if err != nil {
		BaseResponseByError(c, err)
		return
	}
	appMsg := fmt.Sprintf(`所属管理员ID %s %s %s`, userIdStr, appIdStr, xkf_tool.ENTER)
	gsgin.GinResponse(c, gsgin.ResponseSuccess, appMsg+RunResultMsg, ``)
}

//WechatKefuChange 切换微信客服环境
func WechatKefuChange(c *gin.Context) {
	_, reqMap, shell, _, mysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	appInfo := service.QueryWechatAppid(mysqlCli, reqMap[`WechatKefuAppid`].ToStr())
	appIdStr := cast.ToString(appInfo[`app_id`])
	appTypeStr := cast.ToString(appInfo[`app_type`])
	if appIdStr == `` || appTypeStr != `wechat_kefu` {
		BaseResponseByError(c, errors.New(`找不到微信客服应用`))
		return
	}

	cDockerList := base_module.NewCommand().DockerNameList()
	dockerLists, err := shell.RunShell3(cDockerList.GetCommand().ToByte())
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	dockerList := strings.Split(dockerLists, `\n`)
	retMsgList := make([]string, 0)
	for _, dockerName := range dockerList {
		cProcess := base_module.NewCommand().WechatKefuProcess(dockerName, appIdStr)
		runResultMsg, err := shell.RunShell3(cProcess.GetCommand().ToByte())
		if err != nil {
			gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
			return
		}
		if strings.Contains(runResultMsg, `Process exited with status 1`) {
			runResultMsg = `not find
`
			retMsgList = append(retMsgList, dockerName+` `+runResultMsg)
		} else {
			retMsgList = append(retMsgList, dockerName+` `+runResultMsg)
			//找到了进程 那么找到pid kill掉进程
			pid := getPsPid(runResultMsg)
			if cast.ToInt(pid) > 0 {
				cKill := base_module.NewCommand().DockerExecKill(dockerName, pid)
				_, err := shell.RunShell3(cKill.GetCommand().ToByte())
				if err != nil {
					gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
					return
				}
			}
		}
	}
	//丢一个topic
	//time.Sleep(time.Second)
	//producer := xkf_tool.GetProducer(`172.16.0.185`, `4150`, `wechat_kefu_open_`+appIdStr)
	//if producer != nil {
	//	err := producer.PublishMsg(`0`)
	//	if err != nil {
	//		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
	//		return
	//	}
	//}
	dockerName := reqMap[`DockerName`].ToStr()
	if dockerName == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, `DockerName不能为空`, nil)
		return
	}
	//执行脚本
	cPhpCommand := base_module.NewCommand().Init().Sudo().DockerExecPhpWechatKefu(dockerName, reqMap[`DockerCodePath`].ToStr(), appIdStr)
	_, err = shell.RunShell3(cPhpCommand.GetCommand().ToByte())
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	//查询是否成功
	cProcess := base_module.NewCommand().WechatKefuProcess(dockerName, appIdStr)
	result, err := shell.RunShell3(cProcess.GetCommand().ToByte())
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, result, ``)
}

//WechatKefuQueryAppList 查询微信客服列表
func WechatKefuQueryAppList(c *gin.Context) {
	_, reqMap, _, _, mysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	userInfo := service.GetAdminUserId(mysqlCli, reqMap[`Account`].ToStr())
	if userInfo[`_id`] == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, ``, nil)
		return
	}
	dataList := service.QueryEnvWechatKefuList(mysqlCli, cast.ToString(userInfo[`id`]))
	gsgin.GinResponse(c, gsgin.ResponseError, gstool.JsonEncode(dataList), nil)
}

//WechatKefuQueryQrCdeList 微信客服二维码列表
func WechatKefuQueryQrCdeList(c *gin.Context) {
	_, reqMap, _, _, mysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	appInfo := service.QueryWechatAppid(mysqlCli, reqMap[`WechatKefuAppid`].ToStr())
	channelList, err := mysqlCli.GetAll(`select _id,channel_name from tbl_channel where wechatapp_id = ? `, cast.ToString(appInfo[`_id`]))
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	staffList, err := mysqlCli.GetAll(`select name,user_id from tbl_staff where parent_user_id = ? `, cast.ToString(appInfo[`user_id`]))
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	returnMap := make([]map[string]interface{}, 0)

	for _, channelInfo := range channelList {
		tempMap := make(map[string]interface{})
		channelRelList, err := mysqlCli.GetAll(`select user_id,short_code from tbl_channel_user_rel where wechatapp_id = ? and channel_id = ? and status = 1`, cast.ToString(appInfo[`_id`]), cast.ToInt(channelInfo[`_id`]))
		if err != nil {
			continue
		}
		tempMap[`_id`] = cast.ToString(channelInfo[`_id`])
		tempMap[`channel_name`] = cast.ToString(channelInfo[`channel_name`])
		linkList := make([]map[string]string, 0)
		for _, channelRel := range channelRelList {
			staffName := ``
			for _, staffInfo := range staffList {
				if cast.ToString(staffInfo[`user_id`]) == cast.ToString(channelRel[`user_id`]) {
					staffName = cast.ToString(staffInfo[`name`])
					break
				}
			}
			linkList = append(linkList, map[string]string{
				`staff_name`: staffName,
				`short_code`: cast.ToString(channelRel[`short_code`]),
			})
		}
		tempMap[`link_list`] = linkList
		returnMap = append(returnMap, tempMap)
	}
	gsgin.GinResponse(c, gsgin.ResponseError, gstool.JsonEncode(returnMap), nil)
}

func getWechatKefuReqData(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, *gstool.GsShell, *gsdb.GsRedis, *gsdb.GsMysql, error) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	shellName := reqMap[`ShellName`]
	if shellName == nil {
		return nil, nil, nil, nil, nil, errors.New(`缺少ShellName参数`)
	}
	client, err := global.ShellGetClient(shellName.ToStr())
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	redisName := reqMap[`redisName`]
	redisCli, err := global.RedisGetClient(redisName.ToStr())
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	mysqlName := reqMap[`mysqlName`]
	mysqlCli, err := global.MysqlGetClient(mysqlName.ToStr())
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return global, reqMap, client, redisCli, mysqlCli, nil
}

func getPsPid(runResultMsg string) string {
	for i := 0; i < 20; i++ {
		runResultMsg = strings.Replace(runResultMsg, `  `, ` `, 100)
	}
	splitResultList := strings.Split(runResultMsg, ` `)
	if len(splitResultList) >= 1 {
		return cast.ToString(splitResultList[1])
	}
	return ``
}
