package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"errors"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"strings"
)

var (
	cdCommand = `/var/www/`
)

// GitCurrentBranch 查询目录的git分支
func GitCurrentBranch(c *gin.Context) {
	reqMap, sshClient, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	command := base.NewCommand()
	command.Sudo()
	command.Cd(codePath)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitChangeBranch 切换分支
func GitChangeBranch(c *gin.Context) {
	reqMap, sshClient, err := getGitComponent(c)
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
	command1 := base.NewCommand()
	command1.Init()
	command1.Sudo()
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr())

	command := base.NewCommand()
	command.Sudo()
	command.Cd(codePath)
	command.GitIgnoreAll()
	command.GitCleanAll()
	command.GitFetch()
	command.GitPull()
	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
	if !strings.Contains(currentBranch, branchName) {
		//command.RemoteOriginBranch(branchName)
		command.GitCheckout(branchName)
	}
	command.GitPullOrigin(branchName)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitPullBranchOrigin 拉取当前分支最新代码
func GitPullBranchOrigin(c *gin.Context) {
	reqMap, sshClient, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	command1 := base.NewCommand()
	command1.Init()
	command1.Sudo()
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr())
	currentBranch = strings.Replace(currentBranch, "\n", "", -1)

	command := base.NewCommand()
	command.Sudo()
	command.Cd(codePath)
	command.GitIgnoreAll()
	command.GitCleanAll()
	command.GitFetch()
	command.GitPull()
	command.GitPullOrigin(strings.Replace(currentBranch, "\n", "", -1))
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, result)
}

// QueryStatus 查询分支状态
func QueryStatus(c *gin.Context) {
	reqMap, sshClient, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}

	command := base.NewCommand()
	command.Sudo()
	command.Cd(codePath)
	command.GitStatus()

	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitCommitLog 查询提交日志
func GitCommitLog(c *gin.Context) {
	reqMap, sshClient, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	command := base.NewCommand()
	command.Sudo()
	command.Cd(codePath)
	command.GitCommitLog()

	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, result)
}

func GitConfigList(c *gin.Context) {
	gitGroupList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeGit,
	}).All()
	gitList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_git`, `*`, nil).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`git_group_list`: gitGroupList,
		`git_list`:       gitList,
	})
}

func getGitComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshConfig, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, nil, err
	}
	sshId := reqMap[`ssh_id`]
	if cast.ToString(sshId) == `` {
		return nil, nil, errors.New(`缺少ssh_id参数`)
	}
	sshConfig, _ := base.Component.TSqlite.GetSshConfig(sshId)
	uniqueKey := base.Component.TBase.GetCombineKey(sshId, `git`)
	sshClient, sshClientErr := base.Component.TShell.GetClient(sshConfig, uniqueKey, define.SseGit, func(s string) []string {
		if gstool.SContains(s, []string{
			`Receiving objects:`,
			`remote: Counting objects:`,
			`Resolving deltas:`,
			`remote: Compressing objects:`,
			`Checking out files:`,
			`Unpacking objects:`}) {
			msgList := strings.Split(s, "\r")
			msgList = append(msgList, "\n")
			return msgList
		} else {
			return []string{s}
		}
	})
	if sshClientErr != nil {
		return nil, nil, err
	}
	return reqMap, sshClient, nil
}
