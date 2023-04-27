package xkf_tool_gin

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"net/url"
	"strings"
	"time"
	"xkf_tool/internal/app/xkf_tool"
)

// Command
// @auth frog
// @date 2022-12-07 12:25:06
type Command struct {
	cdCommand                    string //cd 命令
	showCurrentBranchCommand     string
	ignoreAllCommand             string
	cleanAllCommand              string
	fetchCommand                 string
	checkoutCommand              string
	pullCommand                  string
	pullOriginCommand            string
	queryDockerProcessByName     string
	dockerExecCommand            string
	dockerKillCommand            string
	runPhpCommand                string
	showLogCommand               string
	SupervisorRestartAllCommand  string
	SupervisorRestartCommand     string
	SupervisorStopCommand        string
	SupervisorStatusCommand      string
	SupervisorConfigShowCommand  string
	GitStatusCommand             string
	WkSupervisorConfListCommand  string
	XkfSupervisorConfListCommand string
}

// WechatKefuStatus
// @auth frog
// @date 2022-12-07 11:20:36
// @param reqBody
// @param cliConf
func (command *Command) WechatKefuStatus(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//查询微信客服所在的环境
	retMsgList := make([]string, 0)
	//拿到appid
	appInfo := xkf_tool.QueryWechatAppid(reqBody.WechatKefuAppid)
	if appInfo.Appid == `` || appInfo.AppType != `wechat_kefu` {
		retMsgList = append(retMsgList, `找不到该应用`)
		return retMsgList
	}
	retMsgList = append(retMsgList, fmt.Sprintf(`所属管理员ID %s %s %s`, appInfo.UserId, appInfo.Appid, xkf_tool.ENTER))
	var runCommand string
	for _, value := range reqBody.DockerList {
		runCommand = fmt.Sprintf(command.queryDockerProcessByName, value.Id, appInfo.Appid)
		xkf_tool.Logger.Debugf(`执行` + runCommand)
		runResultMsg := cliConf.RunShell(runCommand)
		if strings.Contains(runResultMsg, `Process exited with status 1`) {
			runResultMsg = `not find 
`
		}
		retMsgList = append(retMsgList, value.Name+` `+runResultMsg)
	}
	return retMsgList
}

// PullBranchOrigin 拉取当前分支最新代码
// @auth frog
// @date 2022-12-07 11:27:59
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) PullBranchOrigin(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//更新当前分支
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	command.checkoutCommand += reqBody.BranchName
	currentBranch := cliConf.RunShell(command.cdCommand + `;` + command.showCurrentBranchCommand)
	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
	xkf_tool.Logger.Debugf(`当前分支 ` + currentBranch)
	command.pullOriginCommand += currentBranch
	runCommandList := make([]string, 0)
	//切换分支
	runCommandList = append(runCommandList,
		command.cdCommand,
		command.ignoreAllCommand,
		command.cleanAllCommand,
		command.fetchCommand,
		command.pullCommand,
		command.pullOriginCommand,
		command.showCurrentBranchCommand,
	)
	runCommandList = gstool.ArrayFilterEmptyString(&runCommandList)
	log.Debug(`指定命令 ` + strings.Join(runCommandList, `;`))
	ret, err := cliConf.RunShell3([]byte(strings.Join(runCommandList, `;`)))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, strings.Join(runCommandList, `;`), err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// ChangeBranch 切换分支
// @auth frog
// @date 2022-12-07 11:29:35
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) ChangeBranch(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//拿到当前分支
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	command.checkoutCommand += reqBody.BranchName
	command.pullOriginCommand += reqBody.BranchName

	currentBranch := cliConf.RunShell(command.cdCommand + `;` + command.showCurrentBranchCommand)
	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
	xkf_tool.Logger.Debugf(`当前分支 ` + currentBranch)
	//如果已经包含了此分支 那么不再处理
	if strings.Contains(currentBranch, reqBody.BranchName) {
		command.checkoutCommand = ``
	}
	runCommandList := make([]string, 0)
	//切换分支
	runCommandList = append(runCommandList,
		command.cdCommand,
		command.ignoreAllCommand,
		command.cleanAllCommand,
		command.fetchCommand,
		command.pullCommand,
		command.checkoutCommand,
		command.pullOriginCommand,
		command.showCurrentBranchCommand,
	)
	runCommandList = gstool.ArrayFilterEmptyString(&runCommandList)
	ret, err := cliConf.RunShell3([]byte(strings.Join(runCommandList, `;`)))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, strings.Join(runCommandList, `;`), err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// QueryCurrentBranch 查询最新分支
// @auth frog
// @date 2022-12-07 11:31:24
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) QueryCurrentBranch(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//查询当前分支
	retMsgList := make([]string, 0)
	runCommandList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList = append(runCommandList,
		command.cdCommand,
		command.showCurrentBranchCommand,
	)
	xkf_tool.Logger.Debugf(`执行的命令 ` + strings.Join(runCommandList, `;`))
	ret, err := cliConf.RunShell3([]byte(strings.Join(runCommandList, `;`)))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, strings.Join(runCommandList, `;`), err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// QueryStatus 状态
// @auth frog
// @date 2022-12-07 11:31:24
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) QueryStatus(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//查询当前分支
	retMsgList := make([]string, 0)
	runCommandList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList = append(runCommandList,
		command.cdCommand,
		command.GitStatusCommand,
	)
	xkf_tool.Logger.Debugf(`执行的命令 ` + strings.Join(runCommandList, `;`))
	ret := cliConf.RunShell(strings.Join(runCommandList, `;`))
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// Filter 清理全局数据
// @auth frog
// @date 2022-12-07 12:22:04
func (command *Command) Filter() {
	command.cdCommand = `cd /var/www/`
	//查询当前分支
	command.showCurrentBranchCommand = `sudo git symbolic-ref --short -q HEAD;`
	command.ignoreAllCommand = `sudo git checkout .` //忽略所有变更
	command.cleanAllCommand = `sudo git clean -df`   //清理所有新增文件
	command.fetchCommand = `sudo git fetch`
	command.checkoutCommand = `sudo git checkout `
	command.pullCommand = `sudo git pull`
	command.pullOriginCommand = `sudo git pull origin `
	command.queryDockerProcessByName = `sudo docker exec %s ps -ef | grep -i %s`
	command.dockerExecCommand = `sudo docker exec %s `
	command.dockerKillCommand = `kill -9 %s`
	command.runPhpCommand = ` php /var/www/%s/scan/protected/yiic OpenPushWechatKefuOpen %s & `
	command.showLogCommand = `/var/www/%s/scan/protected/runtime/%s.log`
	command.SupervisorRestartAllCommand = ` supervisorctl restart all`
	command.SupervisorRestartCommand = ` supervisorctl restart %s`
	command.SupervisorStatusCommand = `supervisorctl status `
	command.SupervisorStopCommand = `supervisorctl stop %s`
	command.SupervisorConfigShowCommand = `cat %s`
	command.GitStatusCommand = `git status`
	command.XkfSupervisorConfListCommand = `cd /var/www/dockerfiles/dev_test/docker_volumes/supervisor/etc/supervisor/conf.d/; ls | grep '\.conf$' | awk '{printf ""$1"---"; system("head -n 1 "$1)}'`
	command.WkSupervisorConfListCommand = `cd /etc/supervisor/conf.d/; ls | grep '\.conf$' | awk '{printf ""$1"---"; system("head -n 1 "$1)}'`
}

// WechatKefuChange 切换微信客服到当前环境
// @auth frog
// @date 2022-12-07 11:27:59
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) WechatKefuChange(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//查询微信客服所在的环境
	retMsgList := make([]string, 0)
	//拿到appid
	appInfo := xkf_tool.QueryWechatAppid(reqBody.WechatKefuAppid)
	if appInfo.Appid == `` || appInfo.AppType != `wechat_kefu` {
		retMsgList = append(retMsgList, `找不到该应用`)
		return retMsgList
	}
	var runCommand string
	for _, value := range reqBody.DockerList {
		runCommand = fmt.Sprintf(command.queryDockerProcessByName, value.Id, appInfo.Appid)
		xkf_tool.Logger.Debugf(`执行` + runCommand)
		runResultMsg := cliConf.RunShell(runCommand)
		if strings.Contains(runResultMsg, `Process exited with status 1`) {
			runResultMsg = `not find
`
			retMsgList = append(retMsgList, value.Name+` `+runResultMsg)
		} else {
			retMsgList = append(retMsgList, value.Name+` `+runResultMsg)
			//找到了进程 那么找到pid kill掉进程
			pid := getPsPid(runResultMsg)
			if cast.ToInt(pid) > 0 {
				killCommand := fmt.Sprintf(command.dockerExecCommand, value.Id) + fmt.Sprintf(command.dockerKillCommand, pid)
				retMsgList = append(retMsgList, value.Name+` `+killCommand+xkf_tool.ENTER)
				cliConf.RunShell(killCommand)
			}
		}
	}
	//丢一个topic
	time.Sleep(time.Second)
	producer := xkf_tool.GetProducer(reqBody.SshConfig.Host, `4150`, `wechat_kefu_open_`+appInfo.Appid)
	if producer != nil {
		err := producer.PublishMsg(`0`)
		if err != nil {
			xkf_tool.Logger.Errorf(`推送消息失败 %s`, err.Error())
		}
	}

	phpRunCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + fmt.Sprintf(command.runPhpCommand, reqBody.DockerCodePath, appInfo.Appid)
	cliConf.RunShell(phpRunCommand)
	//查询是否成功
	runRetMsg := cliConf.RunShell(fmt.Sprintf(command.queryDockerProcessByName, reqBody.DockerId, appInfo.Appid))
	retMsgList = append(retMsgList, `result：`+runRetMsg)
	return retMsgList
}

func getPsPid(runResultMsg string) string {
	xkf_tool.Logger.Debugf(`分割前字符串 ` + runResultMsg)
	for i := 0; i < 20; i++ {
		runResultMsg = strings.Replace(runResultMsg, `  `, ` `, 100)
	}
	splitResultList := strings.Split(runResultMsg, ` `)
	xkf_tool.Logger.Debugf(`分割结果 %#v`, splitResultList)
	if len(splitResultList) >= 1 {
		return cast.ToString(splitResultList[1])
	}
	return ``
}

// SupervisorRestartAll 消费者管理
// @auth frog
// @date 2022-12-08 11:41:31
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) SupervisorRestartAll(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + ` ` + command.SupervisorRestartAllCommand
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, runCommand, err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// SupervisorRestart 重启消费者
// @auth frog
// @date 2022-12-26 09:21:38
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) SupervisorRestart(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList := make([]string, 0)
	if reqBody.ParentType == `wk` {
		runCommandList = append(runCommandList, fmt.Sprintf(`sudo `+command.SupervisorRestartCommand, reqBody.SupervisorRestartName))
		runCommandList = append(runCommandList, `sudo `+command.SupervisorStatusCommand+` | grep `+reqBody.SupervisorRestartName)
	} else {
		runCommandList = append(runCommandList,
			fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+fmt.Sprintf(command.SupervisorRestartCommand, reqBody.SupervisorRestartName),
			fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+command.SupervisorStatusCommand+` | grep `+reqBody.SupervisorRestartName,
		)
	}
	xkf_tool.Logger.Debugf(`重启命令 %s`, strings.Join(runCommandList, `;`))
	ret, err := cliConf.RunShell3([]byte(strings.Join(runCommandList, `;`)))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, strings.Join(runCommandList, `;`), err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// SupervisorStop 停止消费者
// @auth frog
// @date 2023-01-14 10:03:07
func (command *Command) SupervisorStop(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList := make([]string, 0)
	if reqBody.ParentType == `wk` {
		runCommandList = append(runCommandList, fmt.Sprintf(`sudo `+command.SupervisorStopCommand, reqBody.SupervisorRestartName))
		runCommandList = append(runCommandList, `sudo `+command.SupervisorStatusCommand+` | grep `+reqBody.SupervisorRestartName)
	} else {
		runCommandList = append(runCommandList,
			fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+fmt.Sprintf(command.SupervisorStopCommand, reqBody.SupervisorRestartName),
			fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+command.SupervisorStatusCommand+` | grep `+reqBody.SupervisorRestartName,
		)
	}
	ret, err := cliConf.RunShell3([]byte(strings.Join(runCommandList, `;`)))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败 %s %s`, strings.Join(runCommandList, `;`), err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// SupervisorStatusList 消费者列表
// @auth frog
// @date 2022-12-19 12:22:38
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) SupervisorStatusList(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//消费者
	retMsgList := make([]string, 0)
	if reqBody.ParentType == `wk` {
		runCommand := `sudo ` + command.SupervisorStatusCommand
		xkf_tool.Logger.Debugf(`执行的命令1 ` + runCommand)
		ret, err := cliConf.RunShell3([]byte(runCommand))
		if err != nil {
			xkf_tool.Logger.Errorf(`RunShell3失败 %s`, err.Error())
		}
		retMsgList = append(retMsgList, ret)
	} else {
		command.cdCommand += reqBody.CodePath
		runCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + ` ` + command.SupervisorStatusCommand
		xkf_tool.Logger.Debugf(`执行的命令1 ` + runCommand)
		ret, err := cliConf.RunShell3([]byte(runCommand))
		if err != nil {
			xkf_tool.Logger.Errorf(`RunShell3失败 %s`, err.Error())
		}
		retMsgList = append(retMsgList, ret)
	}

	return retMsgList
}

// SupervisorConfigShow 查看supervisor配置内容
// @auth frog
// @date 2022-12-20 09:26:38
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) SupervisorConfigShow(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//消费者
	retMsgList := make([]string, 0)
	runCommand := fmt.Sprintf(command.SupervisorConfigShowCommand, reqBody.SupervisorConfigPath)
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`获取配置失败 %s`, err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// ShowLog 查看日志
// @auth frog
// @date 2023-01-10 12:04:20
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) ShowLog(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommand := `tail -f -n 1000 ` + fmt.Sprintf(command.showLogCommand, reqBody.CodePath) + `/` + reqBody.LogFile
	xkf_tool.Logger.Debugf(`执行的命令 ` + runCommand)
	ret := cliConf.RunShell(runCommand)
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// DockerExec 执行docker内命令
// @auth frog
// @date 2023-01-14 10:03:26
func (command *Command) DockerExec(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList := make([]string, 0)
	runCommandList = append(runCommandList,
		fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+reqBody.DockerExecCommand,
	)

	xkf_tool.Logger.Debugf(`执行的命令 ` + strings.Join(runCommandList, `;`))
	ret := cliConf.RunShell(strings.Join(runCommandList, `;`))
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// ChangeVipType 变更VIP版本
// @auth frog
// @date 2023-01-17 15:41:29
func (command *Command) ChangeVipType(reqBody *xkf_tool.SshExec) []string {
	//查询微信客服所在的环境
	retMsgList := make([]string, 0)
	//拿到appid
	userInfo := xkf_tool.GetAdminUserId(reqBody.Account)
	if userInfo.Id == `` {
		retMsgList = append(retMsgList, `找不到该账号`)
		return retMsgList
	}
	ret := xkf_tool.UpdateVip(userInfo.Id, cast.ToString(reqBody.ExpiredDay), cast.ToString(reqBody.SystemType), cast.ToString(reqBody.VipLevel))
	//移除缓存
	for _, value := range reqBody.RedisConfigList {
		if xkf_tool.RedisRunList[value.Name] != nil {
			xkf_tool.RedisRunList[value.Name].Client.HDel(`wechatapp.vip.info.v20220308..`+cast.ToString(cast.ToInt(userInfo.Id)%10), userInfo.Id)
			xkf_tool.RedisRunList[value.Name].Client.HDel(`wechatapp.kefu.vip.info.v20220308..`+cast.ToString(cast.ToInt(userInfo.Id)%10), userInfo.Id)
		}
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// GetLoginUrl 获取登录地址
// @auth frog
// @date 2023-03-15 10:10:26
func (command *Command) GetLoginUrl(reqBody *xkf_tool.SshExec) []string {
	retMsgList := make([]string, 0)
	//拿到appid
	userInfo := xkf_tool.GetAdminUserId(reqBody.Account)
	if userInfo.Id == `` {
		retMsgList = append(retMsgList, `找不到该账号`)
		return retMsgList
	}
	token := gstool.JsonEncode(map[string]string{
		`login_type`: `1`,
		`user_id`:    cast.ToString(userInfo.Id),
		`param`: gstool.JsonEncode(map[string]string{
			`uri`: reqBody.LoginUrl,
		}),
		`time`: cast.ToString(time.Now().Unix()), //仅10秒内有效
	})
	data, err := xkf_tool.EncryptMain.EncryptDataDesCBC(token)
	if err != nil {
		retMsgList = append(retMsgList, `加密失败 `+err.Error())
		return retMsgList
	}
	token = url.QueryEscape(data)
	retMsgList = append(retMsgList, reqBody.LoginHost+`index/LoginRedirect?token=`+token)
	return retMsgList
}

// QueryVipType 查询VIP版本
// @auth frog
// @date 2023-03-16 09:30:15
func (command *Command) QueryVipType(reqBody *xkf_tool.SshExec) []string {
	//查询微信客服所在的环境
	retMsgList := make([]string, 0)
	//拿到appid
	userInfo := xkf_tool.GetAdminUserId(reqBody.Account)
	if userInfo.Id == `` {
		retMsgList = append(retMsgList, `找不到该账号`)
		return retMsgList
	}
	vipInfo := xkf_tool.QueryVip(userInfo.Id, cast.ToString(reqBody.SystemType))
	retMsgList = append(retMsgList, `管理员ID：`+userInfo.Id+`，vip版本：`+xkf_tool.VipMap[vipInfo.VipType]+`，过期时间：`+vipInfo.ExpiredTime)
	return retMsgList
}

// CheckAllDockerStatus 检查所有docker状态
// @auth frog
// @date 2023-03-27 14:30:00
func (command *Command) CheckAllDockerStatus(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//查询微信客服所在的环境
	retMsgList := make([]string, 0)
	//	var runCommand string
	//	//检查所有docker状态
	//	for _, value := range reqBody.DockerList {
	//		runCommand = fmt.Sprintf(command.queryDockerProcessByName, value.Id, appInfo.Appid)
	//		xkf_tool.Logger.Debugf(`执行` + runCommand)
	//		runResultMsg := cliConf.RunShell(runCommand)
	//		if strings.Contains(runResultMsg, `Process exited with status 1`) {
	//			runResultMsg = `not find
	//`
	//		}
	//		retMsgList = append(retMsgList, value.Name+` `+runResultMsg)
	//	}
	return retMsgList
}

// RestartDocker 重启docker
// @auth frog
// @date 2023-03-27 14:34:01
func (command *Command) RestartDocker(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	retMsgList := make([]string, 0)
	runCommandList := make([]string, 0)
	runCommandList = append(runCommandList, fmt.Sprintf(`cd /var/www/dockerfiles/dev_test/app/%s/`, reqBody.DockerCodeName))
	runCommandList = append(runCommandList, `sudo docker-compose restart`)
	xkf_tool.Logger.Debugf(`执行的命令 ` + strings.Join(runCommandList, `;`))
	ret := cliConf.RunShell(strings.Join(runCommandList, `;`))
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

func (command *Command) ShowCompose(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	retMsgList := make([]string, 0)
	runCommandList := make([]string, 0)
	runCommandList = append(runCommandList, fmt.Sprintf(`cat /var/www/dockerfiles/dev_test/app/%s/docker-compose.yml`, reqBody.DockerCodeName))
	xkf_tool.Logger.Debugf(`执行的命令 ` + strings.Join(runCommandList, `;`))
	ret := cliConf.RunShell(strings.Join(runCommandList, `;`))
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

func (command *Command) QueryEnvWechatKefuList(reqBody *xkf_tool.SshExec) string {
	userInfo := xkf_tool.GetAdminUserId(reqBody.Account)
	xkf_tool.Logger.Debugf(`userInfo %#v Account %s XkfDevDbConfig %#v`, userInfo, reqBody.Account, reqBody.XkfDevDbConfig)
	if userInfo.Id == `` {
		return ``
	}
	return xkf_tool.QueryEnvWechatKefuList(userInfo.Id)
}

// SupervisorConfList 配置列表
func (command *Command) SupervisorConfList(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	retMsgList := make([]string, 0)
	runCommandList := make([]string, 0)
	if reqBody.ParentType == `wk` {
		runCommandList = append(runCommandList, command.WkSupervisorConfListCommand)
	} else {
		runCommandList = append(runCommandList, command.XkfSupervisorConfListCommand)
	}
	xkf_tool.Logger.Debugf(`执行的命令 ` + reqBody.ParentType + ` ` + strings.Join(runCommandList, `;`))
	ret, err := cliConf.RunShell3([]byte(strings.Join(runCommandList, `;`)))
	if err != nil {
		xkf_tool.Logger.Errorf(`获取配置失败 %s`, err.Error())
	}
	retMsgList = append(retMsgList, ret)
	xkf_tool.Logger.Debugf(`结果 ` + ret)
	return retMsgList
}
