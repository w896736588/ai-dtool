package gin

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"redis_manager/base"
	"redis_manager/define"
	"redis_manager/helper"
	"strings"
)

// Command
// @auth frog
// @date 2022-12-07 12:25:06
type Command struct {
	cdCommand                   string //cd 命令
	showCurrentBranchCommand    string
	ignoreAllCommand            string
	cleanAllCommand             string
	fetchCommand                string
	checkoutCommand             string
	pullCommand                 string
	pullOriginCommand           string
	queryDockerProcessByName    string
	dockerExecCommand           string
	dockerKillCommand           string
	runPhpCommand               string
	showLogCommand              string
	SupervisorRestartAllCommand string
	SupervisorRestartCommand    string
	SupervisorStopCommand       string
	SupervisorStatusCommand     string
	SupervisorConfigShowCommand string
	GitStatusCommand            string
}

// WechatKefuStatus
// @auth frog
// @date 2022-12-07 11:20:36
// @param reqBody
// @param cliConf
func (command *Command) WechatKefuStatus(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//查询微信客服所在的环境
	retMsgList := make([]string, 0)
	//拿到appid
	appInfo := base.QueryWechatAppid(reqBody.WechatKefuAppid, reqBody.XkfDevDbConfig)
	if appInfo.Appid == `` || appInfo.AppType != `wechat_kefu` {
		retMsgList = append(retMsgList, `找不到该应用`)
		return retMsgList
	}
	retMsgList = append(retMsgList, fmt.Sprintf(`所属管理员ID %s %s`, appInfo.UserId, define.ENTER))
	var runCommand string
	for _, value := range reqBody.DockerList {
		runCommand = fmt.Sprintf(command.queryDockerProcessByName, value.Id, appInfo.Appid)
		log.Debugf(`执行` + runCommand)
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
func (command *Command) PullBranchOrigin(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//更新当前分支
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	command.checkoutCommand += reqBody.BranchName
	currentBranch := cliConf.RunShell(command.cdCommand + `;` + command.showCurrentBranchCommand)
	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
	log.Debugf(`当前分支 ` + currentBranch)
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
	runCommandList = helper.FilterEmptyString(&runCommandList)
	log.Debug(`指定命令 ` + strings.Join(runCommandList, `;`))
	ret := cliConf.RunShell(strings.Join(runCommandList, `;`))
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// ChangeBranch 切换分支
// @auth frog
// @date 2022-12-07 11:29:35
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) ChangeBranch(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//拿到当前分支
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	command.checkoutCommand += reqBody.BranchName
	command.pullOriginCommand += reqBody.BranchName

	currentBranch := cliConf.RunShell(command.cdCommand + `;` + command.showCurrentBranchCommand)
	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
	log.Debugf(`当前分支 ` + currentBranch)
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
	runCommandList = helper.FilterEmptyString(&runCommandList)
	log.Debug(`指定命令 ` + strings.Join(runCommandList, `;`))
	ret := cliConf.RunShell(strings.Join(runCommandList, `;`))
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// QueryCurrentBranch 查询最新分支
// @auth frog
// @date 2022-12-07 11:31:24
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) QueryCurrentBranch(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//查询当前分支
	retMsgList := make([]string, 0)
	runCommandList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList = append(runCommandList,
		command.cdCommand,
		command.showCurrentBranchCommand,
	)
	log.Debugf(`执行的命令 ` + strings.Join(runCommandList, `;`))
	ret := cliConf.RunShell(strings.Join(runCommandList, `;`))
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// QueryStatus 状态
// @auth frog
// @date 2022-12-07 11:31:24
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) QueryStatus(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//查询当前分支
	retMsgList := make([]string, 0)
	runCommandList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList = append(runCommandList,
		command.cdCommand,
		command.GitStatusCommand,
	)
	log.Debugf(`执行的命令 ` + strings.Join(runCommandList, `;`))
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
}

// WechatKefuChange 切换微信客服到当前环境
// @auth frog
// @date 2022-12-07 11:27:59
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) WechatKefuChange(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//查询微信客服所在的环境
	retMsgList := make([]string, 0)
	//拿到appid
	appInfo := base.QueryWechatAppid(reqBody.WechatKefuAppid, reqBody.XkfDevDbConfig)
	if appInfo.Appid == `` || appInfo.AppType != `wechat_kefu` {
		retMsgList = append(retMsgList, `找不到该应用`)
		return retMsgList
	}
	var runCommand string
	for _, value := range reqBody.DockerList {
		runCommand = fmt.Sprintf(command.queryDockerProcessByName, value.Id, appInfo.Appid)
		log.Debugf(`执行` + runCommand)
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
				retMsgList = append(retMsgList, value.Name+` `+killCommand+define.ENTER)
				cliConf.RunShell(killCommand)
			}
		}
	}
	phpRunCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + fmt.Sprintf(command.runPhpCommand, reqBody.DockerCodePath, appInfo.Appid)
	log.Debugf(`执行进程命令 ` + phpRunCommand)
	cliConf.RunShell(phpRunCommand)
	//查询是否成功
	runRetMsg := cliConf.RunShell(fmt.Sprintf(command.queryDockerProcessByName, reqBody.DockerId, appInfo.Appid))
	retMsgList = append(retMsgList, `result：`+runRetMsg)
	return retMsgList
}

func getPsPid(runResultMsg string) string {
	log.Debugf(`分割前字符串 ` + runResultMsg)
	for i := 0; i < 20; i++ {
		runResultMsg = strings.Replace(runResultMsg, `  `, ` `, 100)
	}
	splitResultList := strings.Split(runResultMsg, ` `)
	log.Debugf(`分割结果 %#v`, splitResultList)
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
func (command *Command) SupervisorRestartAll(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + ` ` + command.SupervisorRestartAllCommand
	log.Debugf(`执行的命令 ` + runCommand)
	ret := cliConf.RunShell(runCommand)
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// SupervisorRestart 重启消费者
// @auth frog
// @date 2022-12-26 09:21:38
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) SupervisorRestart(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList := make([]string, 0)
	runCommandList = append(runCommandList,
		fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+fmt.Sprintf(command.SupervisorRestartCommand, reqBody.SupervisorRestartName),
		fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+command.SupervisorStatusCommand+` | grep `+reqBody.SupervisorRestartName,
	)

	log.Debugf(`执行的命令 ` + strings.Join(runCommandList, `;`))
	ret := cliConf.RunShell(strings.Join(runCommandList, `;`))
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// SupervisorStop 停止消费者
// @auth frog
// @date 2023-01-14 10:03:07
func (command *Command) SupervisorStop(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList := make([]string, 0)
	runCommandList = append(runCommandList,
		fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+fmt.Sprintf(command.SupervisorStopCommand, reqBody.SupervisorRestartName),
		fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+command.SupervisorStatusCommand+` | grep `+reqBody.SupervisorRestartName,
	)

	log.Debugf(`执行的命令 ` + strings.Join(runCommandList, `;`))
	ret := cliConf.RunShell(strings.Join(runCommandList, `;`))
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// SupervisorStatusList 消费者列表
// @auth frog
// @date 2022-12-19 12:22:38
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) SupervisorStatusList(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + ` ` + command.SupervisorStatusCommand
	log.Debugf(`执行的命令 ` + runCommand)
	ret := cliConf.RunShell(runCommand)
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// SupervisorConfigShow 查看supervisor配置内容
// @auth frog
// @date 2022-12-20 09:26:38
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) SupervisorConfigShow(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//消费者
	retMsgList := make([]string, 0)
	runCommand := fmt.Sprintf(command.SupervisorConfigShowCommand, reqBody.SupervisorConfigPath)
	log.Debugf(`执行的命令 ` + runCommand)
	ret := cliConf.RunShell(runCommand)
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// ShowLog 查看日志
// @auth frog
// @date 2023-01-10 12:04:20
// @param reqBody
// @param cliConf
// @return []string
func (command *Command) ShowLog(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommand := `tail -f -n 1000 ` + fmt.Sprintf(command.showLogCommand, reqBody.CodePath) + `/` + reqBody.LogFile
	log.Debugf(`执行的命令 ` + runCommand)
	ret := cliConf.RunShell(runCommand)
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// DockerExec 执行docker内命令
// @auth frog
// @date 2023-01-14 10:03:26
func (command *Command) DockerExec(reqBody *define.SshExec, cliConf base.ClientConfig) []string {
	//消费者
	retMsgList := make([]string, 0)
	command.cdCommand += reqBody.CodePath
	runCommandList := make([]string, 0)
	runCommandList = append(runCommandList,
		fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId)+` `+reqBody.DockerExecCommand,
	)

	log.Debugf(`执行的命令 ` + strings.Join(runCommandList, `;`))
	ret := cliConf.RunShell(strings.Join(runCommandList, `;`))
	retMsgList = append(retMsgList, ret)
	return retMsgList
}

// ChangeVipType 变更VIP版本
// @auth frog
// @date 2023-01-17 15:41:29
func (command *Command) ChangeVipType(reqBody *define.SshExec) []string {
	//查询微信客服所在的环境
	retMsgList := make([]string, 0)
	//拿到appid
	userInfo := base.GetAdminUserId(reqBody.Account, reqBody.XkfDevDbConfig)
	if userInfo.Id == `` {
		retMsgList = append(retMsgList, `找不到该账号`)
		return retMsgList
	}
	ret := base.UpdateVip(userInfo.Id, cast.ToString(reqBody.SystemType), cast.ToString(reqBody.VipLevel), reqBody.XkfDevDbConfig)
	//移除缓存
	for _, value := range reqBody.RedisConfigList {
		if base.RedisRunList[value.UniKey] != nil {
			base.RedisRunList[value.UniKey].HDel(`wechatapp.vip.info.v20220308..`+cast.ToString(cast.ToInt(userInfo.Id)%10), userInfo.Id)
			base.RedisRunList[value.UniKey].HDel(`wechatapp.kefu.vip.info.v20220308..`+cast.ToString(cast.ToInt(userInfo.Id)%10), userInfo.Id)
		}
	}
	retMsgList = append(retMsgList, ret)
	return retMsgList
}
