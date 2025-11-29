package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"errors"
	"regexp"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
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
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
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
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr())
	currentBranch = CleanBranchName(currentBranch)

	command := base.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
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

// GitChangeBranchRemote 切换远程分支
func GitChangeBranchRemote(c *gin.Context) {
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
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr())
	currentBranch = CleanBranchName(currentBranch)

	command := base.NewCommand()
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
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr())
	currentBranch = CleanBranchName(currentBranch)

	gstool.FmtPrintlnLogTime(`获取当前分支为：%q`, currentBranch)

	command := base.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
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

// 1. 先把非 Git 合法字符全部删掉
// 2. 再把开头结尾的 "/""." 去掉，防止出现 ".foo" 或 "bar/"
var cleanBranchRe = regexp.MustCompile(`[^A-Za-z0-9._/-]+`)

func CleanBranchName(s string) string {
	s = gstool.StringFilterANSI(s)
	s = cleanBranchRe.ReplaceAllString(s, "") // 只留合法字符
	s = strings.TrimPrefix(s, ".")            // 不能以 . 开头
	s = strings.TrimSuffix(s, "/")            // 不能以 / 结尾
	return s
}

// 过滤所有常见的 bracketed-paste 序列
// 兼容日志里 ^[ 或真实 ESC 两种情况
var reESC = regexp.MustCompile(`(\x1b|\^\[)\[?2004[hl]`)

func stripESC(s string) string { return reESC.ReplaceAllString(s, "") }

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
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
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
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitCommitLog()

	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, result)
}

func GitConfigList(c *gin.Context) {
	gitGroupList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeGit,
	}).All()
	//id转为字符串
	for k, v := range gitGroupList {
		gitGroupList[k][`id`] = cast.ToString(v[`id`])
	}
	gitList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_git`, `*`, nil).All()
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
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitCommitLog()

	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr())
	gsgin.GinResponseSuccess(c, ``, result)
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
	sseId := reqMap[`sse_id`]
	sshConfig, _ := base.Component.TSqlite.GetSshConfig(sshId)
	uniqueKey := base.Component.TBase.GetCombineKey(sshId, sseId)
	sshClient, sshClientErr := base.Component.TShell.GetClient(sshConfig, uniqueKey, cast.ToString(sseId), nil)
	if sshClientErr != nil {
		return nil, nil, sshClientErr
	}
	return reqMap, sshClient, nil
}
