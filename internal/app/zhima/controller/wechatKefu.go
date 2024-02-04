package controller

import (
	"dev_tool/base_module"
	"dev_tool/internal/app/zhima/service"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsdefine"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gsnsq"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"strings"
	"time"
)

// WechatKefuStatus 查询微信客服应用的状态
func WechatKefuStatus(c *gin.Context) {
	_, reqMap, shell, _, _, xkfMysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	appInfo := service.QueryWechatAppid(xkfMysqlCli, reqMap[`WechatKefuAppid`].ToStr())
	wechatappId := cast.ToString(appInfo[`_id`])
	appIdStr := cast.ToString(appInfo[`app_id`])
	userIdStr := cast.ToString(appInfo[`user_id`])
	appTypeStr := cast.ToString(appInfo[`app_type`])
	if appIdStr == `` || appTypeStr != `wechat_kefu` {
		BaseResponseByError(c, errors.New(`找不到微信客服应用`))
		return
	}
	//按app_id查找
	command := base_module.NewCommand().Sudo().WechatKefuStatus(wechatappId)
	RunResultMsg := shell.RunShell(command.GetCommand().ToByte())
	appMsg := fmt.Sprintf(`所属管理员ID %s %s %s %s`, userIdStr, wechatappId, appIdStr, gsdefine.Enter)

	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, appMsg+RunResultMsg)
}

//WechatKefuChange 切换微信客服环境
func WechatKefuChange(c *gin.Context) {
	_, reqMap, shell, _, _, xkfMysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	appInfo := service.QueryWechatAppid(xkfMysqlCli, reqMap[`WechatKefuAppid`].ToStr())
	wechatappId := cast.ToString(appInfo[`_id`])
	appIdStr := cast.ToString(appInfo[`app_id`])
	appTypeStr := cast.ToString(appInfo[`app_type`])
	if appIdStr == `` || appTypeStr != `wechat_kefu` {
		BaseResponseByError(c, errors.New(`找不到微信客服应用`))
		return
	}

	cDockerList := base_module.NewCommand().Sudo().DockerNameList()
	dockerLists := shell.RunShell(cDockerList.GetCommand().ToByte())
	gstool.FmtPrintlnLog(`获取所有的docker %#v`, dockerLists)
	dockerList := strings.Split(dockerLists, "\n")
	gstool.FmtPrintlnLog(`dockerList %#v`, dockerList)
	for _, dockerName := range dockerList {
		if dockerName == `NAMES` || dockerName == `` || !strings.Contains(dockerName, `xkf`) {
			continue
		}
		//按照app_id
		dockerKillProcess := base_module.NewCommand().Sudo().DockerKill9(dockerName, appIdStr)
		shell.RunShell(dockerKillProcess.GetCommand().ToByte())
		//按wechatapp_id
		dockerIdKillProcess := base_module.NewCommand().Sudo().DockerKill9(dockerName, wechatappId)
		shell.RunShell(dockerIdKillProcess.GetCommand().ToByte())
	}
	//丢一个topic
	time.Sleep(time.Second)
	nsqProducer := gsnsq.NsqStruct{
		Topic:   `wechat_kefu_open_` + appIdStr,
		Channel: `wechat_kefu_open_` + appIdStr + `_channel`,
		PConfig: gsnsq.NsqConfig{
			Host: shell.Config.Host,
			Port: cast.ToString(4150),
		},
		CConfig:      gsnsq.NsqConfig{},
		ConsumerList: nil,
		Producer:     nil,
		ProducerChan: gstool.ChanStruct{},
	}
	nsqProErr := nsqProducer.CreateProducer(nsqProducer.PConfig)
	if nsqProErr != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, nsqProErr.Error(), nil)
		return
	}
	errProducer := nsqProducer.PublishMsg(`0`)
	if errProducer != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, errProducer.Error(), nil)
		return
	}
	dockerName := reqMap[`DockerName`].ToStr()
	if dockerName == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, `DockerName不能为空`, nil)
		return
	}
	//执行脚本
	cPhpCommand := base_module.NewCommand().Init().Sudo().DockerExecPhpWechatKefu(dockerName, reqMap[`DockerCodePath`].ToStr(), wechatappId)
	_ = shell.RunShell(cPhpCommand.GetCommand().ToByte())
	//查询是否成功
	cProcess := base_module.NewCommand().WechatKefuProcess(dockerName, wechatappId)
	result := shell.RunShell(cProcess.GetCommand().ToByte())
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, result)
}

//WechatKefuQueryAppList 查询微信客服列表
func WechatKefuQueryAppList(c *gin.Context) {
	_, reqMap, _, _, appUrlMysqlCli, xkfMysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	userInfo := service.GetAdminUserId(appUrlMysqlCli, reqMap[`Account`].ToStr())
	if userInfo[`_id`] == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, ``, nil)
		return
	}
	dataList := service.QueryEnvWechatKefuList(xkfMysqlCli, cast.ToString(userInfo[`_id`]))
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, dataList)
}

//WechatKefuQueryQrCdeList 微信客服二维码列表
func WechatKefuQueryQrCdeList(c *gin.Context) {
	_, reqMap, _, _, _, xkfMysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	appInfo := service.QueryWechatAppid(xkfMysqlCli, reqMap[`WechatKefuAppid`].ToStr())
	channelList, err := xkfMysqlCli.GetAll(`select _id,channel_name from tbl_channel where wechatapp_id = ? `, cast.ToString(appInfo[`_id`]))
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	staffList, err := xkfMysqlCli.GetAll(`select name,user_id from tbl_staff where parent_user_id = ? `, cast.ToString(appInfo[`user_id`]))
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	returnMap := make([]map[string]interface{}, 0)

	for _, channelInfo := range channelList {
		tempMap := make(map[string]interface{})
		channelRelList, err := xkfMysqlCli.GetAll(`select user_id,short_code from tbl_channel_user_rel where wechatapp_id = ? and channel_id = ? and status = 1`, cast.ToString(appInfo[`_id`]), cast.ToInt(channelInfo[`_id`]))
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
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, returnMap)
}

func getWechatKefuReqData(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, *gstool.GsShellPush, *gsdb.GsRedis, *gsdb.GsMysql, *gsdb.GsMysql, error) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	shellName := reqMap[`ShellName`]
	if shellName == nil {
		return nil, nil, nil, nil, nil, nil, errors.New(`缺少ShellName参数`)
	}
	client, err := global.ShellPushGetClient(shellName.ToStr())
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	redisName := reqMap[`RedisName`]
	redisCli, err := global.RedisGetClient(redisName.ToStr())
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	xkfMysqlName := reqMap[`XkfMysqlName`]
	xkfMysqlCli, err := global.MysqlGetClient(xkfMysqlName.ToStr())
	appUrlMysqlName := reqMap[`AppUrlMysqlName`]
	appUrlMysqlCli, err := global.MysqlGetClient(appUrlMysqlName.ToStr())
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	return global, reqMap, client, redisCli, appUrlMysqlCli, xkfMysqlCli, nil
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
