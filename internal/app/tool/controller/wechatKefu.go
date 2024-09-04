package controller

import (
	"dev_tool/base_module"
	"dev_tool/internal/app/default/controller"
	"dev_tool/internal/app/zhima/service"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsdefine"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gsnsq"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"strings"
	"time"
)

// WechatKefuStatus 查询微信客服应用的状态
func WechatKefuStatus(c *gin.Context) {
	reqMap, shell, xkfMysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	appInfo := service.QueryWechatAppid(xkfMysqlCli, cast.ToString(reqMap[`WechatKefuAppid`]))
	wechatappId := cast.ToString(appInfo[`_id`])
	appIdStr := cast.ToString(appInfo[`app_id`])
	userIdStr := cast.ToString(appInfo[`user_id`])
	appTypeStr := cast.ToString(appInfo[`app_type`])
	if appIdStr == `` || appTypeStr != `wechat_kefu` {
		controller.BaseResponseByError(c, errors.New(`找不到微信客服应用`))
		return
	}
	//按app_id查找
	command := base_module.NewCommand().Sudo().WechatKefuStatus(wechatappId)
	RunResultMsg, _ := shell.RunCommandWait(command.GetCommand().ToStr())
	appMsg := fmt.Sprintf(`所属管理员ID %s %s %s %s`, userIdStr, wechatappId, appIdStr, gsdefine.Enter)

	gsgin.GinResponseSuccess(c, ``, appMsg+RunResultMsg)
}

// WechatKefuChange 切换微信客服环境
func WechatKefuChange(c *gin.Context) {
	reqMap, shell, xkfMysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	appInfo := service.QueryWechatAppid(xkfMysqlCli, cast.ToString(reqMap[`WechatKefuAppid`]))
	wechatappId := cast.ToString(appInfo[`_id`])
	appIdStr := cast.ToString(appInfo[`app_id`])
	appTypeStr := cast.ToString(appInfo[`app_type`])
	if appIdStr == `` || appTypeStr != `wechat_kefu` {
		controller.BaseResponseByError(c, errors.New(`找不到微信客服应用`))
		return
	}

	cDockerList := base_module.NewCommand().Sudo().DockerNameList()
	dockerLists, _ := shell.RunCommandWait(cDockerList.GetCommand().ToStr())
	dockerList := strings.Split(dockerLists, "\n")
	for _, dockerName := range dockerList {
		if dockerName == `NAMES` || dockerName == `` || !strings.Contains(dockerName, `xkf`) {
			continue
		}
		//根据app_id 找到pid
		WechatKefuSearchAndKillPid(dockerName, appIdStr, shell)
		//按照wechatapp_id 找到pid
		WechatKefuSearchAndKillPid(dockerName, wechatappId, shell)
	}
	//丢一个topic
	time.Sleep(time.Second)
	nsqProducer := gsnsq.NsqStruct{
		Topic:   `wechat_kefu_open_` + appIdStr,
		Channel: `wechat_kefu_open_` + appIdStr + `_channel`,
		Config: gsnsq.NsqConfig{
			PubMsgHost: shell.Host + `:4150`,
		},
		ConsumerList: nil,
		Producer:     nil,
	}
	nsqProErr := nsqProducer.CreateProducer()
	if nsqProErr != nil {
		gsgin.GinResponseError(c, nsqProErr.Error(), nil)
		return
	}
	errProducer := nsqProducer.PublishMsg(`0`)
	if errProducer != nil {
		gsgin.GinResponseError(c, errProducer.Error(), nil)
		return
	}
	dockerName := cast.ToString(reqMap[`DockerName`])
	if dockerName == `` {
		gsgin.GinResponseError(c, `DockerName不能为空`, nil)
		return
	}
	//执行脚本
	cPhpCommand := base_module.NewCommand().Init().Sudo().DockerExecPhpWechatKefu(dockerName, cast.ToString(reqMap[`DockerCodePath`]), wechatappId)
	_, _ = shell.RunCommandWait(cPhpCommand.GetCommand().ToStr())
	//查询是否成功
	cProcess := base_module.NewCommand().WechatKefuProcess(dockerName, wechatappId)
	result, _ := shell.RunCommandWait(cProcess.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, result)
}

func WechatKefuSearchAndKillPid(dockerName, processName string, shell *gsssh.SshConfig) {
	searchIdPidCommand := base_module.NewCommand().Sudo().DockerSearchPidList(dockerName, processName)
	searchIdPidLists, _ := shell.RunCommandWait(searchIdPidCommand.GetCommand().ToStr())
	searchIdPidList := strings.Split(searchIdPidLists, "\n")
	for _, pid := range searchIdPidList {
		gstool.FmtPrintlnLogTime(`找到进程ID ##%s##`, pid)
		pid = strings.ReplaceAll(pid, ` `, ``)
		if pid == `` {
			continue
		}
		if cast.ToString(cast.ToInt(pid)) != pid {
			continue
		}
		gstool.FmtPrintlnLogTime(`删除进程id ##%s##`, pid)
		dockerKillPid := base_module.NewCommand().Sudo().DockerExecKill(dockerName, pid)
		_, _ = shell.RunCommandWait(dockerKillPid.GetCommand().ToStr())
	}
}

// WechatKefuQueryAppList 查询微信客服列表
func WechatKefuQueryAppList(c *gin.Context) {
	reqMap, _, xkfMysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	userInfo := service.GetAdminUserId(xkfMysqlCli, cast.ToString(reqMap[`Account`]))
	if userInfo[`_id`] == `` {
		gsgin.GinResponseError(c, ``, nil)
		return
	}
	dataList := service.QueryEnvWechatKefuList(xkfMysqlCli, cast.ToString(userInfo[`_id`]))
	gsgin.GinResponseSuccess(c, ``, dataList)
}

// WechatKefuQueryQrCdeList 微信客服二维码列表
func WechatKefuQueryQrCdeList(c *gin.Context) {
	reqMap, _, xkfMysqlCli, err := getWechatKefuReqData(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	appInfo := service.QueryWechatAppid(xkfMysqlCli, cast.ToString(reqMap[`WechatKefuAppid`]))
	channelList, err := xkfMysqlCli.QueryBySql(`select _id,channel_name from tbl_channel where wechatapp_id = ? `, cast.ToString(appInfo[`_id`])).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	staffList, err := xkfMysqlCli.QueryBySql(`select name,user_id from tbl_staff where parent_user_id = ? `, cast.ToString(appInfo[`user_id`])).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	returnMap := make([]map[string]interface{}, 0)

	for _, channelInfo := range channelList {
		tempMap := make(map[string]interface{})
		channelRelList, err := xkfMysqlCli.QueryBySql(`select user_id,short_code from tbl_channel_user_rel where wechatapp_id = ? and channel_id = ? and status = 1`, cast.ToString(appInfo[`_id`]), cast.ToInt(channelInfo[`_id`])).All()
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
	gsgin.GinResponseSuccess(c, ``, returnMap)
}

// 获取组件
func getWechatKefuReqData(c *gin.Context) (map[string]any, *gsssh.SshConfig, *gsdb.GsMysql, error) {
	component, componentErr := controller.GetGlobalComponent(c)
	if componentErr != nil {
		return nil, nil, nil, componentErr
	}
	if component.ShellClient == nil {
		return nil, nil, nil, errors.New(`缺少shell client参数`)
	}
	if component.XkfMysqlClient == nil {
		return nil, nil, nil, errors.New(`缺少Mysql client参数`)
	}
	return component.ReqMap, component.ShellClient, component.XkfMysqlClient, nil
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
