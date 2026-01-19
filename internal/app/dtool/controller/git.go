package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_shell"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/crypto/ssh"
)

var (
	cdCommand = `/var/www/`
)

// GitCurrentBranch 查询目录的git分支
func GitCurrentBranch(c *gin.Context) {
	reqMap, sshClient, _, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitChangeBranch 切换分支
func GitChangeBranch(c *gin.Context) {
	reqMap, sshClient, _, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	branchName := cast.ToString(reqMap[`BranchName`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	if branchName == `` {
		gsgin.GinResponseError(c, `切换的分支不能为空`, nil)
		return
	}
	command1 := p_shell.NewCommand()
	command1.Init()
	//command.Sudo()
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr(), 40*time.Second)
	gstool.FmtPrintlnLogTime(`获取当前分支为：%q`, currentBranch)
	currentBranch = CleanBranchName(currentBranch)
	gstool.FmtPrintlnLogTime(`当前分支 %#v`, currentBranch)

	command := p_shell.NewCommand()
	//command.Sudo()
	command.Cd(codePath)
	command.GitIgnoreAll()
	command.GitCleanAll()
	command.GitFetch()
	command.GitPull()
	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
	if currentBranch != branchName {
		//command.RemoteOriginBranch(branchName)
		command.GitCheckout(branchName)
	}
	command.GitPullOrigin(branchName)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitChangeBranchRemote 切换远程分支
func GitChangeBranchRemote(c *gin.Context) {
	reqMap, sshClient, _, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	branchName := cast.ToString(reqMap[`BranchName`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	if branchName == `` {
		gsgin.GinResponseError(c, `切换的分支不能为空`, nil)
		return
	}
	command1 := p_shell.NewCommand()
	command1.Init()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr(), 40*time.Second)
	currentBranch = CleanBranchName(currentBranch)

	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitIgnoreAll()
	command.GitFetch()
	command.GitPull()
	if !strings.Contains(currentBranch, branchName) {
		command.RemoteOriginBranch(branchName)
		command.GitCheckout(branchName)
	}
	command.GitPullOrigin(branchName)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitPullBranchOrigin 拉取当前分支最新代码
func GitPullBranchOrigin(c *gin.Context) {
	reqMap, sshClient, _, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	command1 := p_shell.NewCommand()
	command1.Init()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr(), 40*time.Second)
	currentBranch = sshClient.FilterEndTip(currentBranch)
	currentBranch = CleanBranchName(currentBranch)

	gstool.FmtPrintlnLogTime(`获取当前分支为：%q`, currentBranch)

	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitIgnoreAll()
	command.GitCleanAll()
	command.GitFetch()
	command.GitPull()
	command.GitPullOrigin(currentBranch)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, result)
}

func CleanBranchName(branchName string) string {
	branchName = p_common.TBaseClient.FilterTerminalChars(branchName)
	return strings.Replace(branchName, "\n", "", -1)
}

// QueryStatus 查询分支状态
func QueryStatus(c *gin.Context) {
	reqMap, sshClient, _, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}

	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitStatus()

	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitCommitLog 查询提交日志
func GitCommitLog(c *gin.Context) {
	reqMap, sshClient, _, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitCommitLog()

	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, result)
}

func GitConfigList(c *gin.Context) {
	gitGroupList, _ := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeGit,
	}).All()
	//id转为字符串
	for k, v := range gitGroupList {
		gitGroupList[k][`id`] = cast.ToString(v[`id`])
	}
	gitList, _ := common.DbMain.Client.QuickQuery(`tbl_git`, `*`, nil).All()
	//id转为字符串
	for k, v := range gitList {
		gitList[k][`id`] = cast.ToString(v[`id`])
		gitList[k][`git_group_id`] = cast.ToString(v[`git_group_id`])
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`git_group_list`: gitGroupList,
		`git_list`:       gitList,
	})
}

func CreateMerge(c *gin.Context) {
	reqMap, sshClient, _, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitCommitLog()

	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, result)
}

func getGitComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshTerminal, *p_sse.SseShell, error) {
	dataMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &dataMap)
	if err != nil {
		return nil, nil, nil, err
	}
	sshId := dataMap[`ssh_id`]
	if cast.ToString(sshId) == `` {
		return nil, nil, nil, errors.New(`缺少ssh_id参数`)
	}
	sseDistributeId := cast.ToString(dataMap[`sse_distribute_id`])
	sshConfig, _ := common.DbMain.GetSshConfig(sshId)
	uniqueKey := p_common.TBaseClient.GetCombineKey(sshId, sseDistributeId)
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: sseDistributeId,
	}
	globalMap, err := common.DbMain.AllGlobalMap()
	if err != nil {
		return nil, nil, nil, err
	}
	//输出格式化 去除特殊符号
	formatFunc := func(s string) []string {
		return []string{p_common.TBaseClient.FilterTerminalChars(s)}
	}
	//验证提示关键词
	promptKeywords := []string{"Username for", "Password for"}
	//遇到验证提示关键词时的回调处理
	promptFunc := func(prompt string, stdin io.WriteCloser, session *ssh.Session) string {
		if strings.Contains(prompt, `Username for`) {
			host := p_common.TBaseClient.GetGitPromptHosts(prompt)
			if len(host) == 0 {
				sse.Send(fmt.Sprintf(`未匹配到需要输入账号的来源 %s`, prompt) + "\n")
			} else {
				if input, exist := globalMap[host+`_username`]; exist {
					sse.Send(fmt.Sprintf(`输入git账号（%s）`, host+`_username`) + "\n")
					_, _ = stdin.Write([]byte(fmt.Sprintf("%s\n", input)))
					return ``
				} else {
					sse.Send(fmt.Sprintf(`未找到可以输入的git账号，请在全局变量中配置:%s`, host+`_username`) + "\n")
				}
			}
		}
		if strings.Contains(prompt, `Password for`) {
			host := p_common.TBaseClient.GetGitPromptHosts(prompt)
			if len(host) == 0 {
				sse.Send(fmt.Sprintf(`未匹配到需要输入账号的来源 %s`, prompt) + "\n")
				return ``
			} else {
				if input, exist := globalMap[host+`_password`]; exist {
					sse.Send(fmt.Sprintf("\n"+`输入git密码（%s）`, host+`_password`) + "\n")
					_, _ = stdin.Write([]byte(fmt.Sprintf("%s\n", input)))
					return ``
				} else {
					sse.Send(fmt.Sprintf(`未找到可以输入的git密码，请在全局变量中配置:%s`, host+`_password`) + "\n")
				}
			}
		}
		_ = session.Signal(ssh.SIGINT)
		//清除认证缓存
		if strings.Contains(strings.ToLower(prompt), `git`) {
			_, _ = stdin.Write([]byte("git credential-cache exit; unset GIT_ASKPASS\n"))
		}
		return "\n需要输入账号或密码，请按照提示在全局变量中设置后再次执行\n"
	}
	sshClient, sshClientErr := component.ShellClient.GetClient(sshConfig, uniqueKey, sse, formatFunc, promptKeywords, promptFunc)
	if sshClientErr != nil {
		return nil, nil, nil, sshClientErr
	}
	return dataMap, sshClient, sse, nil
}

// GitSetSafeLog 设置项目安全
func GitSetSafeLog(c *gin.Context) {
	reqMap, sshClient, _, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitSetSafe(codePath)

	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitSaveCredentials 设置项目git自动存储账号密码
func GitSaveCredentials(c *gin.Context) {
	reqMap, sshClient, sse, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.Cat(`.git/config`)
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 4*time.Second)
	if strings.Contains(result, `store`) && strings.Contains(result, `credential`) {
		sse.Send(`已存在设置，不再新增` + "\n")
	} else {
		command := p_shell.NewCommand()
		//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
		command.Cd(codePath)
		command.Append(`.git/config`, "[credential]\nhelper = store\n")
		_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), 4*time.Second)
		sse.Send(`写入成功` + "\n")
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}
