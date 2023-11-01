package zhima

import (
	"context"
	"dev_tool/internal/app/xkf_tool"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdefine"
	"gitee.com/Sxiaobai/gs/gstool"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"net/url"
	"strings"
	"time"
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
	SupervisorStopAllCommand     string
	SupervisorStatusCommand      string
	SupervisorConfigShowCommand  string
	GitStatusCommand             string
	WkSupervisorConfListCommand  string
	XkfSupervisorConfListCommand string
	QueryWechatKefuExistCommand  string
	DockerPsCommand              string
}

// WechatKefuStatus
// @auth frog
// @date 2022-12-07 11:20:36
// @param reqBody
// @param cliConf
func (command *Command) WechatKefuStatus(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell, wkCliTerConf *gstool.GsShell) string {
	cli := cliConf
	if reqBody.SshName == `wk` {
		cli = wkCliTerConf
	}
	//查询微信客服所在的环境
	retMsgList := make([]string, 0)
	//拿到appid
	appInfo := xkf_tool.QueryWechatAppid(reqBody.WechatKefuAppid)
	if appInfo.Appid == `` || appInfo.AppType != `wechat_kefu` {
		retMsgList = append(retMsgList, `找不到该应用`)
		return `找不到该应用`
	}

	shellCommand := fmt.Sprintf(command.QueryWechatKefuExistCommand, appInfo.Appid)
	fmt.Println(`执行的命令 ` + shellCommand)
	RunResultMsg, err := cli.RunShell3([]byte(shellCommand))
	returnMsg := ``
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, shellCommand, err.Error())
		returnMsg += fmt.Sprintf(`执行命令失败%s %s`, shellCommand, err.Error()) + xkf_tool.ENTER
	}
	appMsg := fmt.Sprintf(`所属管理员ID %s %s %s`, appInfo.UserId, appInfo.Appid, xkf_tool.ENTER)
	if returnMsg != `` {
		return appMsg + returnMsg
	} else {
		return appMsg + RunResultMsg
	}
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

	runCommand := command.cdCommand + `;` + command.showCurrentBranchCommand
	currentBranch, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, runCommand, err.Error())
	}
	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
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
	runCommand1 := strings.Join(runCommandList, `;`)
	log.Debug(`指定命令 ` + runCommand1)
	ret, err := cliConf.RunShell3([]byte(runCommand1))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, runCommand1, err.Error())
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

	runCommand := command.cdCommand + `;` + command.showCurrentBranchCommand
	currentBranch, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, runCommand, err.Error())
	}

	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
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
	runCommand1 := strings.Join(runCommandList, `;`)
	ret, err := cliConf.RunShell3([]byte(runCommand1))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, runCommand1, err.Error())
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
	runCommand := strings.Join(runCommandList, `;`)
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, runCommand, err.Error())
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
	runCommand := strings.Join(runCommandList, `;`)
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, runCommand, err.Error())
	}
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
	command.SupervisorStopAllCommand = `supervisorctl stop all`
	command.SupervisorConfigShowCommand = `cat %s`
	command.GitStatusCommand = `git status`
	command.XkfSupervisorConfListCommand = `cd /var/www/dockerfiles/dev_test/docker_volumes/supervisor/etc/supervisor/conf.d/; ls | grep '\.conf$' | awk '{printf ""$1"---"; system("head -n 1 "$1)}'`
	command.WkSupervisorConfListCommand = `cd /etc/supervisor/conf.d/; ls | grep '\.conf$' | awk '{printf ""$1"---"; system("head -n 1 "$1)}'`
	command.QueryWechatKefuExistCommand = `sudo docker ps |awk '{print $NF}'|grep -v NAMES|xargs -I {} bash -c " echo {} && sudo docker exec {} ps -ef | grep -i %s "`
	command.DockerPsCommand = `sudo docker stats --no-stream`
}

// WechatKefuChange 切换微信客服到当前环境
// @auth frog
// @date 2022-12-07 11:27:59
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) WechatKefuChange(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell, wkCliTerConf *gstool.GsShell) []string {
	cli := cliConf
	if reqBody.SshName == `wk` {
		cli = wkCliTerConf
	}
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
		if value.SshName != reqBody.SshName {
			continue
		}
		runCommand = fmt.Sprintf(command.queryDockerProcessByName, value.Id, appInfo.Appid)
		runResultMsg, err := cli.RunShell3([]byte(runCommand))
		if err != nil {
			xkf_tool.Logger.Errorf(`执行失败 %s`, err.Error())
		}
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
				_, err := cli.RunShell3([]byte(killCommand))
				if err != nil {
					xkf_tool.Logger.Errorf(`执行命令失败 %s %s`, killCommand, err.Error())
				}
			}
		}
	}
	//丢一个topic
	time.Sleep(time.Second)
	//host := reqBody.SshConfig.Host
	//if reqBody.SshName == `wk` {
	//	host = reqBody.WkSshConfig.Host
	//}
	//producer := xkf_tool.GetProducer(host, `4150`, `wechat_kefu_open_`+appInfo.Appid)
	//if producer != nil {
	//	err := producer.PublishMsg(`0`)
	//	if err != nil {
	//		xkf_tool.Logger.Errorf(`推送消息失败 %s`, err.Error())
	//	}
	//}
	phpRunCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + fmt.Sprintf(command.runPhpCommand, reqBody.DockerCodePath, appInfo.Appid)
	gstool.FmtPrintlnLog(`执行命令 %s`, phpRunCommand)
	_, err := cli.RunShell3([]byte(phpRunCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败 %s %s`, phpRunCommand, err.Error())
	}
	//查询是否成功
	runCommand2 := fmt.Sprintf(command.queryDockerProcessByName, reqBody.DockerId, appInfo.Appid)
	_, err = cli.RunShell3([]byte(runCommand2))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败 %s %s`, runCommand2, err.Error())
	}
	//查询结果
	returnMsgList := make([]string, 0)
	returnMsgList = append(returnMsgList, command.WechatKefuStatus(reqBody, cliConf, wkCliTerConf))
	return returnMsgList
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

// SupervisorRestartAll 消费者管理
// @auth frog
// @date 2022-12-08 11:41:31
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) SupervisorRestartAll(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//消费者
	command.cdCommand += reqBody.CodePath
	runCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + ` ` + command.SupervisorRestartAllCommand
	_, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, runCommand, err.Error())
	}
	return command.SupervisorStatusList(reqBody, cliConf)
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
	runCommand := strings.Join(runCommandList, `;`)
	xkf_tool.Logger.Errorf(`执行命令 %s`, runCommand)
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败%s %s`, runCommand, err.Error())
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
	runCommand := strings.Join(runCommandList, `;`)
	xkf_tool.Logger.Errorf(`执行命令 %s`, runCommand)
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败 %s %s`, runCommand, err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// SupervisorStopAll 停止所有消费者
// @auth frog
// @date 2023-01-14 10:03:07
func (command *Command) SupervisorStopAll(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList := make([]string, 0)
	if reqBody.ParentType == `wk` {
		runCommandList = append(runCommandList, `sudo `+command.SupervisorStopAllCommand, reqBody.SupervisorRestartName)
		runCommandList = append(runCommandList, `sudo `+command.SupervisorStatusCommand)
	} else {
		runCommandList = append(runCommandList,
			fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+command.SupervisorStopAllCommand,
			fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+command.SupervisorStopAllCommand,
		)
	}
	runCommand := strings.Join(runCommandList, `;`)
	xkf_tool.Logger.Errorf(`执行命令 %s`, runCommand)
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行命令失败 %s %s`, runCommand, err.Error())
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
		ret, err := cliConf.RunShell3([]byte(runCommand))
		if err != nil {
			xkf_tool.Logger.Errorf(`RunShell3失败 %s`, err.Error())
		}
		retMsgList = append(retMsgList, ret)
	} else {
		command.cdCommand += reqBody.CodePath
		runCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + ` ` + command.SupervisorStatusCommand
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
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行失败 %s %s`, runCommand, err.Error())
	}
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

	runCommand := strings.Join(runCommandList, `;`)
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行失败 %s %s`, runCommand, err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// DockerPs 查看docker状态
// @auth frog
// @date 2023-01-14 10:03:26
func (command *Command) DockerPs(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell, wkCliConf *gstool.GsShell) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList := make([]string, 0)
	runCommandList = append(runCommandList,
		command.DockerPsCommand,
	)

	runCommand := strings.Join(runCommandList, `;`)
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行失败 %s %s`, runCommand, err.Error())
	}
	retMsgList = append(retMsgList, ret)

	retMsgList = append(retMsgList, gsdefine.Enter)

	//执行另一个
	ret, err = wkCliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行失败 %s %s`, runCommand, err.Error())
	}
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
	_ = xkf_tool.UpdateVip(userInfo.Id, cast.ToString(reqBody.ExpiredDay), cast.ToString(reqBody.SystemType), cast.ToString(reqBody.VipLevel))
	//移除缓存
	for _, value := range reqBody.RedisConfigList {
		if xkf_tool.RedisRunList[value.Name] != nil {
			xkf_tool.RedisRunList[value.Name].Client.HDel(context.Background(), `wechatapp.vip.info.v20220308..`+cast.ToString(cast.ToInt(userInfo.Id)%10), userInfo.Id)
			xkf_tool.RedisRunList[value.Name].Client.HDel(context.Background(), `wechatapp.kefu.vip.info.v20220308..`+cast.ToString(cast.ToInt(userInfo.Id)%10), userInfo.Id)
		}
	}
	return command.QueryVipType(reqBody)
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
	//拿到一个应用ID和一个渠道ID
	wechatAppId, channelId := xkf_tool.QueryOneWechatAppIdChannelId(cast.ToInt(userInfo.Id))

	redirectUrl := reqBody.LoginUrl
	redirectUrl = strings.Replace(redirectUrl, `{wechatapp_id}`, wechatAppId, -1)
	redirectUrl = strings.Replace(redirectUrl, `{channel_id}`, channelId, -1)
	token := gstool.JsonEncode(map[string]string{
		`login_type`: `1`,
		`user_id`:    cast.ToString(userInfo.Id),
		`param`: gstool.JsonEncode(map[string]string{
			`uri`: redirectUrl,
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
	runCommand := strings.Join(runCommandList, `;`)
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行失败 %s %s`, runCommand, err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

func (command *Command) ShowCompose(reqBody *xkf_tool.SshExec, cliConf *gstool.GsShell) []string {
	retMsgList := make([]string, 0)
	runCommandList := make([]string, 0)
	runCommandList = append(runCommandList, fmt.Sprintf(`cat /var/www/dockerfiles/dev_test/app/%s/docker-compose.yml`, reqBody.DockerCodeName))
	runCommand := strings.Join(runCommandList, `;`)
	ret, err := cliConf.RunShell3([]byte(runCommand))
	if err != nil {
		xkf_tool.Logger.Errorf(`执行失败 %s %s`, runCommand, err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

func (command *Command) QueryEnvWechatKefuList(reqBody *xkf_tool.SshExec) string {
	userInfo := xkf_tool.GetAdminUserId(reqBody.Account)
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
	ret, err := cliConf.RunShell3([]byte(strings.Join(runCommandList, `;`)))
	if err != nil {
		xkf_tool.Logger.Errorf(`获取配置失败 %s`, err.Error())
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

//QueryWechatQrCdeList 微信客服二维码列表
func (command *Command) QueryWechatQrCdeList(reqBody *xkf_tool.SshExec) string {
	appInfo := xkf_tool.QueryWechatAppid(reqBody.WechatKefuAppid)
	channelList, err := xkf_tool.XkfDevMysql.GetAll(`select _id,channel_name from tbl_channel where wechatapp_id = ? `, appInfo.Id)
	if err != nil {
		return `获取渠道列表失败`
	}
	staffList, err := xkf_tool.XkfDevMysql.GetAll(`select name,user_id from tbl_staff where parent_user_id = ? `, appInfo.UserId)
	if err != nil {
		return `获取客服列表失败`
	}
	returnMap := make([]map[string]interface{}, 0)
	if err != nil {
		xkf_tool.Logger.Errorf(`获取渠道列表失败`)
		return `获取渠道列表失败`
	} else {
		for _, channelInfo := range channelList {
			tempMap := make(map[string]interface{})
			channelRelList, err := xkf_tool.XkfDevMysql.GetAll(`select user_id,short_code from tbl_channel_user_rel where wechatapp_id = ? and channel_id = ? and status = 1`, appInfo.Id, cast.ToInt(channelInfo[`_id`]))
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
		return gstool.JsonEncode(returnMap)
	}
}
