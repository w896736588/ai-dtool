package gin

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"redis_manager/base"
	"redis_manager/define"
	"redis_manager/helper"
	"strings"
	"time"
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
	wechatKefuLogCommand        string
	SupervisorRestartAllCommand string
	SupervisorRestartCommand    string
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
	var runCommand string
	for _, value := range reqBody.DockerList {
		runCommand = fmt.Sprintf(command.queryDockerProcessByName, value.Id, reqBody.WechatKefuAppid)
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
	//输出到指定日志
	//command.runPhpCommand = `nohup php /var/www/%s/scan/protected/yiic OpenPushWechatKefuOpen %s > %s 2>&1 & `
	command.wechatKefuLogCommand = `/var/www/%s/scan/protected/runtime/%s.log`
	command.SupervisorRestartAllCommand = ` supervisorctl restart all`
	command.SupervisorRestartCommand = ` supervisorctl restart %s`
	command.SupervisorStatusCommand = `supervisorctl status `
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
	var runCommand string
	for _, value := range reqBody.DockerList {
		runCommand = fmt.Sprintf(command.queryDockerProcessByName, value.Id, reqBody.WechatKefuAppid)
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
				retMsgList = append(retMsgList, value.Name+` `+killCommand+`
`)
				cliConf.RunShell(killCommand)
			}
		}

	}
	//日志路径
	//logFilePath := fmt.Sprintf(command.wechatKefuLogCommand , reqBody.DockerCodePath , reqBody.WechatKefuAppid)
	//先往日志文件写入一行日志
	//echoLogCommand := fmt.Sprintf(command.dockerExecCommand , reqBody.DockerId) + ` touch ` + logFilePath
	//log.Debugf(`写入日志命令 ` + echoLogCommand)
	//发布一条消息
	base.XkfPublishMsg(`0`, fmt.Sprintf(`wechat_kefu_open_%s`, reqBody.WechatKefuAppid))
	time.Sleep(time.Second)
	//cliConf.RunShell(echoLogCommand)
	phpRunCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + fmt.Sprintf(command.runPhpCommand, reqBody.DockerCodePath, reqBody.WechatKefuAppid)
	log.Debugf(`执行进程命令 ` + phpRunCommand)
	cliConf.RunShell(phpRunCommand)
	//查询是否成功
	runRetMsg := cliConf.RunShell(fmt.Sprintf(command.queryDockerProcessByName, reqBody.DockerId, reqBody.WechatKefuAppid))
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
	runCommand := fmt.Sprintf(command.dockerExecCommand, reqBody.DockerId) + ` ` + fmt.Sprintf(command.SupervisorRestartCommand, reqBody.SupervisorRestartName)
	log.Debugf(`执行的命令 ` + runCommand)
	ret := cliConf.RunShell(runCommand)
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
