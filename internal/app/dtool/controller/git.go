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
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gsssh"
	"github.com/w896736588/go-tool/gstool"
	"golang.org/x/crypto/ssh"
)

var (
	cdCommand = `/var/www/`
)

// sendGitSse 安全发送 Git 操作 SSE 消息。
func sendGitSse(sse *p_sse.SseShell, msg string) {
	if sse != nil && sse.Sse != nil {
		sse.Send(msg + "\n")
	}
}

type gitSwitchStreamEvent struct {
	Type    string         `json:"type"`
	Status  string         `json:"status,omitempty"`
	Message string         `json:"message,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

func sendGitSwitchStreamEvent(sse *gsgin.Sse, event gitSwitchStreamEvent) {
	if sse == nil {
		return
	}
	_ = sse.SendToChan(gstool.JsonEncode(event))
}

func sendGitSwitchStreamLine(sse *gsgin.Sse, line string) {
	line = strings.TrimRight(line, "\r\n")
	if sse == nil || line == `` {
		return
	}
	sendGitSwitchStreamEvent(sse, gitSwitchStreamEvent{
		Type:    `line`,
		Message: line,
	})
}

func gitCleanupAndSwitchRunLocalStep(sse *gsgin.Sse, localDir string, stepName string, args ...string) error {
	sendGitSwitchStreamEvent(sse, gitSwitchStreamEvent{
		Type:    `step`,
		Status:  `running`,
		Message: stepName,
		Data: map[string]any{
			`local_dir`: localDir,
			`command`:   `git ` + strings.Join(args, ` `),
		},
	})
	cmd := exec.Command(`git`, append([]string{`-C`, localDir}, args...)...)
	output, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(output))
	if trimmed != `` {
		for _, line := range strings.Split(trimmed, "\n") {
			sendGitSwitchStreamLine(sse, strings.TrimRight(line, "\r"))
		}
	}
	if err != nil {
		msg := trimmed
		if msg == `` {
			msg = err.Error()
		}
		sendGitSwitchStreamEvent(sse, gitSwitchStreamEvent{
			Type:    `step`,
			Status:  `error`,
			Message: stepName + `失败`,
			Data: map[string]any{
				`local_dir`: localDir,
				`command`:   `git ` + strings.Join(args, ` `),
				`error`:     msg,
			},
		})
		return fmt.Errorf(msg)
	}
	sendGitSwitchStreamEvent(sse, gitSwitchStreamEvent{
		Type:    `step`,
		Status:  `done`,
		Message: stepName + `完成`,
		Data: map[string]any{
			`local_dir`: localDir,
			`command`:   `git ` + strings.Join(args, ` `),
		},
	})
	return nil
}

func GitCleanupAndSwitchBranchByIdStream(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
	localDir := strings.TrimSpace(urlValues.Get(`local_dir`))
	baseBranch := strings.TrimSpace(urlValues.Get(`base_branch`))
	branchName := strings.TrimSpace(urlValues.Get(`branch_name`))
	if localDir == `` {
		return nil, fmt.Errorf(`local_dir不能为空`)
	}
	if baseBranch == `` {
		return nil, fmt.Errorf(`base_branch不能为空`)
	}
	if branchName == `` {
		return nil, fmt.Errorf(`branch_name不能为空`)
	}
	if !isSafeGitBranchInput(baseBranch) || !isSafeGitBranchInput(branchName) {
		return nil, fmt.Errorf(`分支名格式不合法`)
	}
	info, err := os.Stat(localDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf(`本地目录不存在: %s`, localDir)
	}
	clientID := `git_cleanup_switch_` + cast.ToString(time.Now().UnixNano())
	sse := gsgin.SseRegister(clientID, stopC, c)
	go func() {
		defer close(stopC)
		sendGitSwitchStreamEvent(sse, gitSwitchStreamEvent{
			Type:    `meta`,
			Status:  `running`,
			Message: `开始本地清理并切换分支`,
			Data: map[string]any{
				`mode`:        `local`,
				`local_dir`:   localDir,
				`base_branch`: baseBranch,
				`branch_name`: branchName,
			},
		})
		steps := []struct {
			name string
			args []string
		}{
			{name: `检查当前仓库`, args: []string{`rev-parse`, `--show-toplevel`}},
			{name: `清理索引区和工作区`, args: []string{`reset`, `--hard`, `HEAD`}},
			{name: `删除未跟踪文件`, args: []string{`clean`, `-fd`}},
			{name: `拉取远端引用`, args: []string{`fetch`, `--all`, `--prune`}},
			{name: `切换到基线分支`, args: []string{`checkout`, baseBranch}},
			{name: `同步基线分支`, args: []string{`pull`, `origin`, baseBranch}},
		}
		for _, step := range steps {
			if runErr := gitCleanupAndSwitchRunLocalStep(sse, localDir, step.name, step.args...); runErr != nil {
				sendGitSwitchStreamEvent(sse, gitSwitchStreamEvent{
					Type:    `done`,
					Status:  `error`,
					Message: runErr.Error(),
				})
				return
			}
		}
		if branchName != baseBranch {
			checkCmd := exec.Command(`git`, `-C`, localDir, `show-ref`, `--verify`, `--quiet`, `refs/heads/`+branchName)
			if err := checkCmd.Run(); err == nil {
				if runErr := gitCleanupAndSwitchRunLocalStep(sse, localDir, `删除本地旧分支`, `branch`, `-D`, branchName); runErr != nil {
					sendGitSwitchStreamEvent(sse, gitSwitchStreamEvent{
						Type:    `done`,
						Status:  `error`,
						Message: runErr.Error(),
					})
					return
				}
			}
			if runErr := gitCleanupAndSwitchRunLocalStep(sse, localDir, `创建并切换目标分支`, `checkout`, `-b`, branchName); runErr != nil {
				sendGitSwitchStreamEvent(sse, gitSwitchStreamEvent{
					Type:    `done`,
					Status:  `error`,
					Message: runErr.Error(),
				})
				return
			}
		}
		currentCmd := exec.Command(`git`, `-C`, localDir, `rev-parse`, `--abbrev-ref`, `HEAD`)
		currentOutput, currentErr := currentCmd.CombinedOutput()
		currentBranch := strings.TrimSpace(string(currentOutput))
		if currentErr != nil {
			sendGitSwitchStreamEvent(sse, gitSwitchStreamEvent{
				Type:    `done`,
				Status:  `error`,
				Message: strings.TrimSpace(string(currentOutput)),
			})
			return
		}
		sendGitSwitchStreamEvent(sse, gitSwitchStreamEvent{
			Type:    `done`,
			Status:  `success`,
			Message: `分支切换完成`,
			Data: map[string]any{
				`local_dir`:      localDir,
				`base_branch`:    baseBranch,
				`branch_name`:    branchName,
				`current_branch`: currentBranch,
			},
		})
	}()
	return sse, nil
}

func GitCleanupAndSwitchBranchByIdStreamClose(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	sse.UnRegister()
}

// GitCurrentBranch 查询目录的git分支
func GitCurrentBranch(c *gin.Context) {
	reqMap, sshClient, sse, err := getGitComponent(c)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[GitCurrentBranch][01] 获取Git组件失败 err=%s`, err.Error())
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	if codePath == `` {
		gstool.FmtPrintlnLogTime(`[GitCurrentBranch][02] code_path为空 req=%#v`, reqMap)
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	gstool.FmtPrintlnLogTime(`[GitCurrentBranch][03] 开始前置环境处理 code_path=%s ssh_id=%s sse_distribute_id=%s`, codePath, cast.ToString(reqMap[`ssh_id`]), cast.ToString(reqMap[`sse_distribute_id`]))
	// 所有通过 SSH 的 Git 操作前，默认先执行"目录安全 + 保存账号密码"。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gstool.FmtPrintlnLogTime(`[GitCurrentBranch][04] 前置环境处理失败 code_path=%s err=%s`, codePath, prepareErr.Error())
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}
	gstool.FmtPrintlnLogTime(`[GitCurrentBranch][05] 前置环境处理完成，准备查询当前分支 code_path=%s`, codePath)
	sendGitSse(sse, fmt.Sprintf("[ssh] 查询 %s 当前分支...", codePath))
	command := p_shell.NewCommand()
	//command.Sudo()
	command.Cd(codePath)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	commandText := command.GetCommand().ToStr()
	gstool.FmtPrintlnLogTime(`[GitCurrentBranch][06] 执行当前分支查询命令 timeout=%s command=%s`, (time.Second * 4).String(), commandText)
	result, runErr := sshClient.RunCommandWait(commandText, time.Second*4)
	if runErr != nil {
		gstool.FmtPrintlnLogTime(`[GitCurrentBranch][07] 当前分支查询失败 result_len=%d result=%q err=%s`, len(result), result, runErr.Error())
		sendGitSse(sse, fmt.Sprintf("[ssh] 查询失败 %s", runErr.Error()))
	} else {
		gstool.FmtPrintlnLogTime(`[GitCurrentBranch][08] 当前分支查询完成 result_len=%d result=%q`, len(result), result)
		sendGitSse(sse, "[ssh] 查询完成")
	}
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitChangeBranch 切换分支
func GitChangeBranch(c *gin.Context) {
	reqMap, sshClient, sse, err := getGitComponent(c)
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
	// 所有通过 SSH 的 Git 操作前，默认先执行"目录安全 + 保存账号密码"。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}
	sendGitSse(sse, fmt.Sprintf("[ssh] 切换 %s 到分支 %s...", codePath, branchName))
	command1 := p_shell.NewCommand()
	command1.Init()
	//command.Sudo()
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr(), getGitOperationTimeout(gitOperationBranchChange))
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
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), getGitOperationTimeout(gitOperationBranchChange))
	sendGitSse(sse, fmt.Sprintf("[ssh] 切换到分支 %s 完成", branchName))
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitChangeBranchRemote 切换远程分支
func GitChangeBranchRemote(c *gin.Context) {
	reqMap, sshClient, sse, err := getGitComponent(c)
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
	// 所有通过 SSH 的 Git 操作前，默认先执行"目录安全 + 保存账号密码"。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}
	command1 := p_shell.NewCommand()
	command1.Init()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr(), getGitOperationTimeout(gitOperationBranchChange))
	currentBranch = CleanBranchName(currentBranch)

	sendGitSse(sse, fmt.Sprintf("[ssh] 切换 %s 到远程分支 %s...", codePath, branchName))
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
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), getGitOperationTimeout(gitOperationBranchChange))
	sendGitSse(sse, fmt.Sprintf("[ssh] 切换到远程分支 %s 完成", branchName))
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitPullBranchOrigin 拉取当前分支最新代码
func GitPullBranchOrigin(c *gin.Context) {
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
	// 所有通过 SSH 的 Git 操作前，默认先执行"目录安全 + 保存账号密码"。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}

	sendGitSse(sse, fmt.Sprintf("[ssh] 拉取 %s 项目代码...", codePath))
	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitIgnoreAll()
	command.GitCleanAll()
	command.GitFetch()
	command.GitPull()
	// 通过命令替换动态取当前分支，避免分支探测输出中的 prompt/命令残留污染后续拉取命令。
	command.GitPullOriginCurrentBranch()
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, runErr := sshClient.RunCommandWait(command.GetCommand().ToStr(), getGitOperationTimeout(gitOperationPull))
	if runErr != nil {
		sendGitSse(sse, fmt.Sprintf("[ssh] 拉取失败: %s", runErr.Error()))
		gsgin.GinResponseError(c, runErr.Error(), result)
		return
	}
	sendGitSse(sse, "[ssh] 拉取完成")
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitRemoteBranchList 查询指定仓库的全部远程分支
func GitRemoteBranchList(c *gin.Context) {
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

	// 所有通过 SSH 的 Git 操作前，默认先执行"目录安全 + 保存账号密码"。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}

	sendGitSse(sse, fmt.Sprintf("[ssh] 查询 %s 远程分支列表...", codePath))
	command := p_shell.NewCommand()
	command.Cd(codePath)
	command.GitFetch()
	command.GitShowAllOriginBranches()
	result, runErr := sshClient.RunCommandWait(command.GetCommand().ToStr(), getGitOperationTimeout(gitOperationPull))
	if runErr != nil {
		gsgin.GinResponseError(c, runErr.Error(), nil)
		return
	}
	branchList := parseAllRemoteBranches(result)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: branchList,
	})
}

// GitQuickCreateBranch 快捷创建并推送业务分支
func GitQuickCreateBranch(c *gin.Context) {
	reqMap, sshClient, sse, err := getGitComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(reqMap[`code_path`])
	baseBranch := cast.ToString(reqMap[`base_branch`])
	branchType := strings.ToLower(cast.ToString(reqMap[`branch_type`]))
	businessEN := cast.ToString(reqMap[`business_en`])
	if codePath == `` {
		gsgin.GinResponseError(c, `git未配置目录`, nil)
		return
	}
	if baseBranch == `` {
		gsgin.GinResponseError(c, `基于分支不能为空`, nil)
		return
	}
	if !isSafeGitBranchInput(baseBranch) {
		gsgin.GinResponseError(c, `基于分支格式不合法`, nil)
		return
	}
	if branchType != `feature` && branchType != `hotfix` {
		gsgin.GinResponseError(c, `分支类型仅支持 feature/hotfix`, nil)
		return
	}
	if !isValidBusinessEnglish(businessEN) {
		gsgin.GinResponseError(c, `业务英文仅允许英文、数字、下划线`, nil)
		return
	}
	// 所有通过 SSH 的 Git 操作前，默认先执行"目录安全 + 保存账号密码"。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}

	globalMap, mapErr := common.DbMain.AllGlobalMap()
	if mapErr != nil {
		gsgin.GinResponseError(c, mapErr.Error(), nil)
		return
	}
	// 兼容两种全局变量写法：{global_user_name}
	userName := cast.ToString(globalMap[`{global_user_name}`])
	if userName == `` {
		gsgin.GinResponseError(c, `全局变量 {global_user_name}为空或不合法`, nil)
		return
	}

	branchDate := time.Now().Format(`20060102`)
	newBranchName := fmt.Sprintf(`%s_%s_%s_%s`, branchType, userName, businessEN, branchDate)

	sendGitSse(sse, fmt.Sprintf("[ssh] 在 %s 创建分支 %s...", codePath, newBranchName))
	command := p_shell.NewCommand()
	command.Cd(codePath)
	// 按约定顺序执行：pull -> fetch -> checkout . -> clean
	command.GitPull()
	command.GitFetch()
	command.GitIgnoreAll()
	command.GitCleanAll()
	command.GitCheckout(baseBranch)
	command.GitPullOrigin(baseBranch)
	command.GitCheckoutNewBranch(newBranchName)
	command.GitPushOriginSetUpstream(newBranchName)
	command.Echo(`新分支：`)
	command.Echo(newBranchName)

	result, runErr := sshClient.RunCommandWait(command.GetCommand().ToStr(), getGitOperationTimeout(gitOperationQuickCreate))
	if runErr != nil {
		sendGitSse(sse, fmt.Sprintf("[ssh] 创建分支失败: %s", runErr.Error()))
		gsgin.GinResponseError(c, runErr.Error(), nil)
		return
	}
	sendGitSse(sse, fmt.Sprintf("[ssh] 创建分支 %s 完成", newBranchName))
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`branch_name`: newBranchName,
		`result`:      result,
	})
}

func CleanBranchName(branchName string) string {
	branchName = p_common.TBaseClient.FilterTerminalChars(branchName)
	return strings.Replace(branchName, "\n", "", -1)
}

// QueryStatus 查询分支状态
func QueryStatus(c *gin.Context) {
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
	// 所有通过 SSH 的 Git 操作前，默认先执行"目录安全 + 保存账号密码"。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}

	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitStatus()

	sendGitSse(sse, fmt.Sprintf("[ssh] 查询 %s 项目状态...", codePath))
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	sendGitSse(sse, "[ssh] 查询状态完成")
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitCommitLog 查询提交日志
func GitCommitLog(c *gin.Context) {
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
	// 所有通过 SSH 的 Git 操作前，默认先执行"目录安全 + 保存账号密码"。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}
	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitCommitLog()

	sendGitSse(sse, fmt.Sprintf("[ssh] 查询 %s 提交日志...", codePath))
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	sendGitSse(sse, "[ssh] 查询提交日志完成")
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
	gitList = filterGitListByExistingGroups(gitGroupList, gitList)
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

// filterGitListByExistingGroups 仅保留仍然绑定到有效 Git 分组的仓库配置。
func filterGitListByExistingGroups(gitGroupList, gitList []map[string]any) []map[string]any {
	validGroupMap := make(map[string]struct{}, len(gitGroupList))
	for _, gitGroup := range gitGroupList {
		groupID := strings.TrimSpace(cast.ToString(gitGroup[`id`]))
		if groupID == `` {
			continue
		}
		validGroupMap[groupID] = struct{}{}
	}

	filteredList := make([]map[string]any, 0, len(gitList))
	for _, gitItem := range gitList {
		groupID := strings.TrimSpace(cast.ToString(gitItem[`git_group_id`]))
		if groupID == `` {
			continue
		}
		if _, ok := validGroupMap[groupID]; !ok {
			continue
		}
		filteredList = append(filteredList, gitItem)
	}
	return filteredList
}

// 推送一行
func pushTableLine(name, codePath, localBranch, remoteBranch, use string, writeSummary func(string)) {
	tableLine := fmt.Sprintf(
		"| %s | %s | %s | %s |",
		escapeMarkdownTableCell(name),
		escapeMarkdownTableCell(codePath),
		escapeMarkdownTableCell(localBranch),
		escapeMarkdownTableCell(remoteBranch),
	)
	writeSummary(tableLine + "\n")
}

const (
	// gitGroupBranchListConcurrency Git 分组分支查询的最大并发数。
	// gitGroupBranchListConcurrency caps Git group branch queries to 5 concurrent SSH sessions.
	gitGroupBranchListConcurrency = 5
)

// gitBranchRunner 抽象单次 SSH 执行能力，避免交互式终端缓冲造成串线。
// gitBranchRunner abstracts one-shot SSH execution to avoid interactive terminal buffer bleed.
type gitBranchRunner interface {
	RunCommandOnce(command string) (string, error)
	Close()
}

// gitBranchOnceRunner 包装 SshOnce，统一成可关闭的单次执行接口。
// gitBranchOnceRunner wraps SshOnce into a closeable one-shot runner interface.
type gitBranchOnceRunner struct {
	client *gsssh.SshOnce
}

func (h *gitBranchOnceRunner) RunCommandOnce(command string) (string, error) {
	return h.client.RunCommandOnce(command)
}

func (h *gitBranchOnceRunner) Close() {}

// gitBranchRunnerFactory 创建单次查询使用的一次性 SSH 执行器。
// gitBranchRunnerFactory builds a one-shot SSH runner for every repository query.
type gitBranchRunnerFactory func(sshConfig map[string]any) (gitBranchRunner, error)

// gitBranchRunnerRelease 在单次执行结束后做额外清理；默认无需动作。
// gitBranchRunnerRelease performs optional cleanup after one-shot execution completes.
type gitBranchRunnerRelease func()

var (
	// gitGroupBranchRunnerFactory 默认创建一次性 SSH 执行器；测试中可替换。
	// gitGroupBranchRunnerFactory creates a one-shot SSH runner and can be swapped in tests.
	gitGroupBranchRunnerFactory gitBranchRunnerFactory = func(sshConfig map[string]any) (gitBranchRunner, error) {
		client, err := component.ShellClient.GetSshOnce(sshConfig)
		if err != nil {
			return nil, err
		}
		return &gitBranchOnceRunner{client: client}, nil
	}
	// gitGroupBranchRunnerRelease 一次性执行器默认无需额外释放；测试中可替换。
	// gitGroupBranchRunnerRelease is a no-op by default because SshOnce is not pooled.
	gitGroupBranchRunnerRelease gitBranchRunnerRelease = func() {
	}
)

// runGitGroupBranchQueries 并发查询 Git 分组下所有仓库分支。
// runGitGroupBranchQueries uses fresh SSH connections per repository and limits concurrency.
func runGitGroupBranchQueries(
	gitList []map[string]any,
	sshConfig map[string]any,
	writeSummary func(string),
	factory gitBranchRunnerFactory,
	release gitBranchRunnerRelease,
) []map[string]any {
	if len(gitList) == 0 {
		return []map[string]any{}
	}
	if factory == nil {
		factory = gitGroupBranchRunnerFactory
	}
	if release == nil {
		release = gitGroupBranchRunnerRelease
	}

	results := make([]map[string]any, len(gitList))
	// 中文注释：使用信号量限制同时活跃的 SSH 连接数，避免同一 SSH 被瞬时打满。
	// English comment: A semaphore keeps active SSH sessions under the configured concurrency cap.
	limiter := make(chan struct{}, gitGroupBranchListConcurrency)
	var waitGroup sync.WaitGroup

	for index, item := range gitList {
		index := index
		item := item
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			limiter <- struct{}{}
			defer func() {
				<-limiter
			}()
			results[index] = queryGitGroupBranchItem(item, sshConfig, writeSummary, factory, release)
		}()
	}

	waitGroup.Wait()
	return results
}

// queryGitGroupBranchItem 查询单个仓库的本地/远程分支，并保证连接即用即释放。
// queryGitGroupBranchItem queries one repository and always closes/releases the fresh SSH client.
func queryGitGroupBranchItem(
	item map[string]any,
	sshConfig map[string]any,
	writeSummary func(string),
	factory gitBranchRunnerFactory,
	release gitBranchRunnerRelease,
) map[string]any {
	name := cast.ToString(item[`name`])
	codePath := cast.ToString(item[`code_path`])
	command := p_shell.NewCommand()
	command.Cd(codePath)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()

	runner, cliErr := factory(sshConfig)
	if cliErr != nil {
		pushTableLine(name, codePath, cliErr.Error(), ``, ``, writeSummary)
		return map[string]any{
			`name`:         name,
			`local_branch`: cliErr.Error(),
		}
	}
	defer func() {
		runner.Close()
		release()
	}()

	result, getErr := runner.RunCommandOnce(command.GetCommand().ToStr())
	if getErr != nil {
		pushTableLine(name, codePath, getErr.Error(), ``, ``, writeSummary)
		return map[string]any{
			`name`:         name,
			`local_branch`: getErr.Error(),
		}
	}

	localBranch, remoteBranch := parseGitGroupBranchOutput(name, result)
	pushTableLine(name, codePath, localBranch, remoteBranch, ``, writeSummary)
	return map[string]any{
		`name`:          name,
		`local_branch`:  localBranch,
		`remote_branch`: remoteBranch,
	}
}

// parseGitGroupBranchOutput 解析分支查询输出中的本地/远程分支名称。
// parseGitGroupBranchOutput extracts local and remote branch names from one-shot command output.
func parseGitGroupBranchOutput(name string, result string) (string, string) {
	// 中文注释：一次性 SSH 输出不存在交互式终端 prompt，只需清理终端控制字符并按标签提取。
	// English comment: One-shot SSH output has no terminal prompt state; clean terminal chars and parse labels.
	ret := p_common.TBaseClient.FilterTerminalChars(result)

	splitRet := strings.Split(ret, "\n")
	localBranch := `-`
	remoteBranch := `-`
	for indexSplit, split := range splitRet {
		if split == `当前分支：` && len(splitRet) > indexSplit+1 {
			localBranch = splitRet[indexSplit+1]
		}
		if split == `远程分支：` && len(splitRet) > indexSplit+1 {
			remoteBranch = splitRet[indexSplit+1]
		}
	}
	gstool.FmtPrintlnLogTime(`%s 运行结果 %#v`, name, splitRet)
	gstool.FmtPrintlnLogTime(`%s 结果 本地：%s 远程：%s`, name, localBranch, remoteBranch)
	return localBranch, remoteBranch
}

// 查询某个组当前的git分支和使用情况
func GitGroupBranchList(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gitGroupId := cast.ToString(reqMap[`git_group_id`])
	if gitGroupId == `` {
		gsgin.GinResponseError(c, `缺少git_group_id参数`, nil)
		return
	}

	groupInfo, _ := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeGit,
		`id`:   gitGroupId,
	}).One()
	if len(groupInfo) == 0 {
		gsgin.GinResponseError(c, `未找到对应Git分组`, nil)
		return
	}

	gitList, _ := common.DbMain.Client.QuickQuery(`tbl_git`, `*`, map[string]any{
		`git_group_id`: gitGroupId,
	}).All()
	if len(gitList) == 0 {
		gsgin.GinResponseSuccess(c, ``, map[string]any{
			`git_group_id`: gitGroupId,
			`group_name`:   cast.ToString(groupInfo[`name`]),
			`list`:         []map[string]any{},
			`summary_text`: fmt.Sprintf("Git分组 [%s] 下暂无项目\n", cast.ToString(groupInfo[`name`])),
		})
		return
	}
	//sse分发id
	sseDistributeId := cast.ToString(reqMap[`sse_distribute_id`])
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: sseDistributeId,
	}

	// 统一发送汇总信息（串行流程下无需并发锁）
	writeSummary := func(msg string) {
		if sse != nil && sse.Sse != nil {
			sse.Send(msg)
		}
	}

	totalCount := len(gitList)
	gstool.FmtPrintlnLogTime(`本次查询仓库数量%d`, totalCount)
	summaryLines := make([]string, 0, totalCount+4)
	writeSummary("\n" + fmt.Sprintf("### Git分组 `%s` 分支总览", cast.ToString(groupInfo[`name`])))
	writeSummary(fmt.Sprintf("本次查询仓库数量 %d", totalCount))

	summaryLines = append(summaryLines, "| 名称 | 路径 | 当前分支 | 远程分支 |")
	summaryLines = append(summaryLines, "| --- | --- | --- | --- | ")
	writeSummary("\n" + strings.Join(summaryLines, "\n") + "\n")
	// 中文注释：同组仓库共用一个 SSH 配置，但每个仓库查询都重新建连。
	// English comment: Repositories in the group share one SSH config, but each query uses a fresh connection.
	sshId := gitList[0][`ssh_id`]
	sshConfig, sshErr := common.DbMain.GetSshConfig(sshId)
	if sshErr != nil || len(sshConfig) == 0 {
		gsgin.GinResponseSuccess(c, ``, map[string]any{
			`git_group_id`: gitGroupId,
			`group_name`:   cast.ToString(groupInfo[`name`]),
			`list`:         []map[string]any{},
			`summary_text`: fmt.Sprintf("Git(%s)未配置ssh连接\n", gitList[0][`name`]),
		})
		return
	}
	resultList := runGitGroupBranchQueries(gitList, sshConfig, writeSummary, nil, nil)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`git_group_id`:  gitGroupId,
		`group_name`:    cast.ToString(groupInfo[`name`]),
		`list`:          resultList,
		`summary_lines`: summaryLines,
	})
	return

}

type GitCurrentBranchInfo struct {
	LocalBranch  string
	RemoteBranch string
	RawOutput    string
}

type GitBranchUsageInfo struct {
	UsageDisplay string
	Owners       []string
}

const (
	gitDefaultCommandTimeout    = 40 * time.Second
	gitBranchChangeTimeout      = 10 * time.Minute
	gitOperationBranchChange    = `branch_change`
	gitOperationPull            = `pull`
	gitOperationQuickCreate     = `quick_create_branch`
	gitRemoteBranchTimeout      = 10 * time.Second
	gitRemoteBranchRetryTimeout = 3 * time.Second
	// gitBranchUsageUsedDisplay 表示检测到当前分支有人使用。
	gitBranchUsageUsedDisplay = "有人使用"
	// gitBranchUsageNoneDisplay 统一表示无人使用，或本地分支不存在。
	gitBranchUsageNoneDisplay = "-"
)

// getGitOperationTimeout 根据Git操作类型返回对应的命令超时时间
func getGitOperationTimeout(operation string) time.Duration {
	switch operation {
	case gitOperationBranchChange:
		return gitBranchChangeTimeout
	case gitOperationPull:
		return gitBranchChangeTimeout
	case gitOperationQuickCreate:
		return gitBranchChangeTimeout
	default:
		return gitDefaultCommandTimeout
	}
}

func isBranchParseNoise(line string) bool {
	if line == `` {
		return true
	}
	if strings.Contains(line, `%d\n`) || strings.Contains(line, `%d\\n`) || strings.Contains(line, `"$?"`) || strings.Contains(line, `$?"`) {
		return true
	}
	if strings.HasPrefix(line, `cd `) {
		return true
	}
	if isLikelyShellPrompt(line) {
		return true
	}
	return false
}

func isLikelyShellPrompt(line string) bool {
	if !(strings.HasSuffix(line, `$`) || strings.HasSuffix(line, `#`) || strings.HasSuffix(line, `%`)) {
		return false
	}
	if !strings.Contains(line, `@`) || !strings.Contains(line, `:`) {
		return false
	}
	return true
}

// parseAllRemoteBranches 解析 git ls-remote --heads origin 输出
func parseAllRemoteBranches(output string) []string {
	text := p_common.TBaseClient.FilterTerminalChars(output)
	lines := strings.Split(text, "\n")
	set := make(map[string]struct{})
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == `` {
			continue
		}
		if isBranchParseNoise(trimmed) {
			continue
		}
		if !strings.Contains(trimmed, "refs/heads/") {
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) == 0 {
			continue
		}
		lastPart := parts[len(parts)-1]
		if !strings.HasPrefix(lastPart, `refs/heads/`) {
			continue
		}
		branchName := strings.TrimPrefix(lastPart, `refs/heads/`)
		if branchName == `` {
			continue
		}
		if _, exist := set[branchName]; exist {
			continue
		}
		set[branchName] = struct{}{}
		result = append(result, branchName)
	}
	sort.Strings(result)
	return result
}

// isValidBusinessEnglish 校验业务英文（仅英文、数字、下划线）
func isValidBusinessEnglish(value string) bool {
	ok, _ := regexp.MatchString(`^[A-Za-z0-9_]+$`, strings.TrimSpace(value))
	return ok
}

// isSafeGitBranchInput 校验用户输入的基于分支名，避免命令注入
func isSafeGitBranchInput(value string) bool {
	ok, _ := regexp.MatchString(`^[A-Za-z0-9._/\-]+$`, strings.TrimSpace(value))
	return ok
}

// normalizeBranchNamePart 统一处理分支名中的人员字段，保留字母/数字（含中文）/下划线/中划线
func normalizeBranchNamePart(value string) string {
	v := strings.TrimSpace(value)
	if v == `` {
		return ``
	}
	re := regexp.MustCompile(`[^\p{L}\p{N}_-]+`)
	v = re.ReplaceAllString(v, `_`)
	v = strings.Trim(v, `_-`)
	return v
}

func escapeMarkdownTableCell(value string) string {
	v := strings.TrimSpace(value)
	v = strings.ReplaceAll(v, "|", "\\|")
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "\r", " ")
	if v == "" {
		return " "
	}
	return v
}

func CreateMerge(c *gin.Context) {
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
	// 所有通过 SSH 的 Git 操作前，默认先执行"目录安全 + 保存账号密码"。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}
	command := p_shell.NewCommand()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command.Cd(codePath)
	command.GitCommitLog()

	sendGitSse(sse, fmt.Sprintf("[ssh] 查询 %s 提交日志...", codePath))
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	sendGitSse(sse, "[ssh] 查询提交日志完成")
	gsgin.GinResponseSuccess(c, ``, result)
}

func getGitComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshTerminal, *p_sse.SseShell, error) {
	sseClientId := c.GetHeader(`SseClientId`)
	gstool.FmtPrintlnLogTime(`[getGitComponent][01] 开始解析Git请求 sse_client_id=%s`, sseClientId)
	dataMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &dataMap)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[getGitComponent][02] 解析body失败 err=%s`, err.Error())
		return nil, nil, nil, err
	}
	sshId := dataMap[`ssh_id`]
	if cast.ToString(sshId) == `` {
		gstool.FmtPrintlnLogTime(`[getGitComponent][03] 缺少ssh_id req=%#v`, dataMap)
		return nil, nil, nil, errors.New(`缺少ssh_id参数`)
	}
	sseDistributeId := cast.ToString(dataMap[`sse_distribute_id`])
	sshConfig, sshConfigErr := common.DbMain.GetSshConfig(sshId)
	if sshConfigErr != nil {
		gstool.FmtPrintlnLogTime(`[getGitComponent][04] 获取SSH配置失败 ssh_id=%s err=%s`, cast.ToString(sshId), sshConfigErr.Error())
	} else if len(sshConfig) == 0 {
		gstool.FmtPrintlnLogTime(`[getGitComponent][04] SSH配置为空 ssh_id=%s`, cast.ToString(sshId))
	}
	uniqueKey := p_common.TBaseClient.GetCombineKey(sshId, sseDistributeId)
	gstool.FmtPrintlnLogTime(`[getGitComponent][05] 请求参数解析完成 ssh_id=%s unique_key=%s sse_distribute_id=%s code_path=%s`, cast.ToString(sshId), uniqueKey, sseDistributeId, cast.ToString(dataMap[`code_path`]))
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(sseClientId),
		SseDistributeId: sseDistributeId,
	}
	gstool.FmtPrintlnLogTime(`[getGitComponent][06] SSE绑定状态 sse_client_id=%s sse_distribute_id=%s can_send=%v`, sseClientId, sseDistributeId, sse.Sse != nil)
	globalMap, err := common.DbMain.AllGlobalMap()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[getGitComponent][07] 获取global map失败 err=%s`, err.Error())
		return nil, nil, nil, err
	}
	gstool.FmtPrintlnLogTime(`[getGitComponent][08] global map加载完成 count=%d`, len(globalMap))
	//输出格式化 去除特殊符号
	formatFunc := func(s string) []string {
		return []string{p_common.TBaseClient.FilterTerminalChars(s)}
	}
	//验证提示关键词
	promptKeywords := []string{"Username for", "Password for"}
	promptFunc := func(prompt string, stdin io.WriteCloser, session *ssh.Session) string {
		gstool.FmtPrintlnLogTime(`prompt %s`, prompt)
		if strings.Contains(prompt, `Username for`) {
			host := p_common.TBaseClient.GetGitPromptHosts(prompt)
			if len(host) == 0 {
				gstool.FmtPrintlnLogTime(`未匹配到需要输入账号的来源 %s`+"\n", prompt)
				sse.Send(fmt.Sprintf(`未匹配到需要输入账号的来源 %s`, prompt) + "\n")
			} else {
				if input, exist := globalMap[host+`_username`]; exist {
					sse.Send(fmt.Sprintf(`输入git账号（%s）`, host+`_username`) + "\n")
					gstool.FmtPrintlnLogTime(`输入git账号（%s）,%s`, host+`_username`, input)
					_, _ = stdin.Write([]byte(fmt.Sprintf("%s\n", input)))
					return p_common.TBaseClient.FilterGitPromptHosts(prompt, `Username for`)
				} else {
					gstool.FmtPrintlnLogTime(`未找到可以输入的git账号，请在全局变量中配置:%s`, host+`_username`)
					sse.Send(fmt.Sprintf(`未找到可以输入的git账号，请在全局变量中配置:%s`, host+`_username`) + "\n")
				}
			}
		}
		if strings.Contains(prompt, `Password for`) {
			host := p_common.TBaseClient.GetGitPromptHosts(prompt)
			if len(host) == 0 {
				gstool.FmtPrintlnLogTime(`未匹配到需要输入账号的来源 %s`+"\n", prompt)
				sse.Send(fmt.Sprintf(`未匹配到需要输入账号的来源 %s`, prompt) + "\n")
			} else {
				if input, exist := globalMap[host+`_password`]; exist {
					gstool.FmtPrintlnLogTime(`输入git密码（%s）,%s`, host+`_password`, input)
					sse.Send(fmt.Sprintf("\n"+`输入git密码（%s）`, host+`_password`) + "\n")
					_, _ = stdin.Write([]byte(fmt.Sprintf("%s\n", input)))
					return p_common.TBaseClient.FilterGitPromptHosts(prompt, `Password for`)
				} else {
					gstool.FmtPrintlnLogTime(`未找到可以输入的git密码，请在全局变量中配置:%s`, host+`_password`)
					sse.Send(fmt.Sprintf(`未找到可以输入的git密码，请在全局变量中配置:%s`, host+`_password`) + "\n")
				}
			}
		}
		// 只有在未处理任何认证信息时才发送中断信号
		_ = session.Signal(ssh.SIGINT)
		//清除认证缓存
		if strings.Contains(strings.ToLower(prompt), `git`) {
			_, _ = stdin.Write([]byte("git credential-cache exit; unset GIT_ASKPASS\n"))
		}
		sse.Send("\n需要输入账号或密码，请按照提示在全局变量中设置后再次执行\n")
		return prompt
	}
	sshClient, sshClientErr := component.ShellClient.GetClient(sshConfig, uniqueKey, sse, formatFunc, promptKeywords, promptFunc)
	if sshClientErr != nil {
		gstool.FmtPrintlnLogTime(`[getGitComponent][09] 获取SSH客户端失败 unique_key=%s err=%s`, uniqueKey, sshClientErr.Error())
		return nil, nil, nil, sshClientErr
	}
	gstool.FmtPrintlnLogTime(`[getGitComponent][10] 获取SSH客户端成功 unique_key=%s`, uniqueKey)
	return dataMap, sshClient, sse, nil
}

// GitSetSafeLog 设置项目安全
func GitSetSafeLog(c *gin.Context) {
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
	command.GitSetSafe(codePath)

	sendGitSse(sse, fmt.Sprintf("[ssh] 设置 %s 安全目录...", codePath))
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	sendGitSse(sse, "[ssh] 设置安全目录完成")
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
	sendGitSse(sse, fmt.Sprintf("[ssh] 配置 %s git账号密码存储...", codePath))
	command.SetCommand(`grep -i -E '^\[credential\]|^[[:space:]]*helper[[:space:]]*=[[:space:]]*store' .git/config`)
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

// prepareGitOperationEnv 统一执行 Git 的 SSH 前置环境处理。
// 先设置 safe.directory，再确保 .git/config 存在 credential.store，避免操作进入交互认证。
// prepareGitOperationEnv 设置safe.directory并确保credential store已配置（合并为单次SSH命令）
func prepareGitOperationEnv(sshClient *gsssh.SshTerminal, codePath string) error {
	if sshClient == nil {
		return errors.New(`ssh client 为空`)
	}
	if codePath == `` {
		return errors.New(`git未配置目录`)
	}

	// 合并为单次SSH命令：设置safe.directory + 检查credential store，不存在则追加
	cmd := p_shell.NewCommand()
	cmd.Cd(codePath)
	cmd.GitSetSafe(codePath)
	cmd.SetCommand(`grep -qi '\[credential\]' .git/config 2>/dev/null && grep -qi 'helper.*=.*store' .git/config 2>/dev/null || printf '[credential]\nhelper = store\n' >> .git/config`)
	_, err := sshClient.RunCommandWait(cmd.GetCommand().ToStr(), 6*time.Second)
	return err
}

const (
	gitUploadFileMaxSize = 10 * 1024 * 1024 // 文件上传最大10MB
)

// GitUploadFile 上传本地文件到远程Git项目目录
// 参数：git_id(Git配置ID), local_file_paths(文件路径数组,每项含full_file_path和relative_file_path), code_path(远程代码目录,为空则从tbl_git获取)
func GitUploadFile(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, nil)
		return
	}

	gitId := cast.ToString(reqMap[`git_id`])

	// 解析 local_file_paths 为 [{full_file_path, relative_file_path}]
	type fileItem struct {
		FullFilePath     string
		RelativeFilePath string
	}
	var fileItems []fileItem
	if paths, ok := reqMap[`local_file_paths`]; ok {
		if arr, ok := paths.([]interface{}); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok {
					fileItems = append(fileItems, fileItem{
						FullFilePath:     strings.TrimSpace(cast.ToString(m[`full_file_path`])),
						RelativeFilePath: strings.TrimSpace(cast.ToString(m[`relative_file_path`])),
					})
				}
			}
		}
	}

	// 参数校验
	if gitId == `` {
		gsgin.GinResponseError(c, `git_id不能为空`, nil)
		return
	}
	if len(fileItems) == 0 {
		gsgin.GinResponseError(c, `local_file_paths不能为空`, nil)
		return
	}
	for _, item := range fileItems {
		if item.FullFilePath == `` {
			gsgin.GinResponseError(c, `full_file_path不能为空`, nil)
			return
		}
		if item.RelativeFilePath == `` {
			gsgin.GinResponseError(c, `relative_file_path不能为空`, nil)
			return
		}
		if strings.Contains(item.RelativeFilePath, `..`) {
			gsgin.GinResponseError(c, `relative_file_path不合法: `+item.RelativeFilePath, nil)
			return
		}
	}

	// 根据git_id查询tbl_git配置
	gitInfo, queryErr := common.DbMain.Client.QuickQuery(`tbl_git`, `*`, map[string]any{
		`id`: gitId,
	}).One()
	if queryErr != nil || len(gitInfo) == 0 {
		gsgin.GinResponseError(c, `未找到id为 "`+gitId+`" 的Git配置`, nil)
		return
	}

	// 优先使用参数传入的 code_path，为空则使用 tbl_git 配置的 code_path
	codePath := strings.TrimSpace(cast.ToString(reqMap[`code_path`]))
	if codePath == `` {
		codePath = cast.ToString(gitInfo[`code_path`])
	}
	if codePath == `` {
		gsgin.GinResponseError(c, `Git项目未配置code_path`, nil)
		return
	}

	sshId := gitInfo[`ssh_id`]
	sshConfig, sshConfigErr := common.DbMain.GetSshConfig(sshId)
	if sshConfigErr != nil {
		gsgin.GinResponseError(c, `获取SSH配置失败: `+sshConfigErr.Error(), nil)
		return
	}
	if len(sshConfig) == 0 {
		gsgin.GinResponseError(c, `SSH配置为空`, nil)
		return
	}

	// 创建一次性SSH连接
	gstool.FmtPrintlnLogTime(`[GitUploadFile] 创建SSH连接 ssh_id=%s host=%s port=%s`, sshId, sshConfig[`host`], sshConfig[`port`])
	sshOnce, sshErr := component.ShellClient.GetSshOnce(sshConfig)
	if sshErr != nil {
		gstool.FmtPrintlnLogTime(`[GitUploadFile] 创建SSH连接失败: %s`, sshErr.Error())
		gsgin.GinResponseError(c, `创建SSH连接失败: `+sshErr.Error(), nil)
		return
	}

	// 提前创建所有远程目标目录，避免因目录不存在导致上传失败
	targetDirs := make(map[string]bool)
	for _, item := range fileItems {
		relativePath := strings.ReplaceAll(item.RelativeFilePath, `\`, `/`)
		relativePath = strings.TrimLeft(relativePath, `/`)
		targetPath := codePath + `/` + relativePath
		targetDirs[path.Dir(targetPath)] = true
	}
	gstool.FmtPrintlnLogTime(`[GitUploadFile] 需要创建的远程目录: %v`, targetDirs)
	for dir := range targetDirs {
		gstool.FmtPrintlnLogTime(`[GitUploadFile] mkdir -p %s`, dir)
		if out, mkdirErr := sshOnce.RunCommandOnce(fmt.Sprintf(`mkdir -p %s`, dir)); mkdirErr != nil {
			gstool.FmtPrintlnLogTime(`[GitUploadFile] mkdir失败: dir=%s err=%s out=%s`, dir, mkdirErr.Error(), out)
			gsgin.GinResponseError(c, `创建远程目录失败: `+mkdirErr.Error(), nil)
			return
		}
		gstool.FmtPrintlnLogTime(`[GitUploadFile] mkdir成功: %s`, dir)
	}

	// 上传每个文件
	results := make([]map[string]any, 0, len(fileItems))
	gstool.FmtPrintlnLogTime(`[GitUploadFile] 开始上传 %d 个文件, code_path=%s`, len(fileItems), codePath)
	for _, item := range fileItems {
		gstool.FmtPrintlnLogTime(`[GitUploadFile] 上传文件: local=%s remote_relative=%s`, item.FullFilePath, item.RelativeFilePath)
		// 校验本地文件
		fileInfo, statErr := os.Stat(item.FullFilePath)
		if statErr != nil {
			gsgin.GinResponseError(c, `本地文件不存在或无法访问: `+statErr.Error(), nil)
			return
		}
		if fileInfo.IsDir() {
			gsgin.GinResponseError(c, `full_file_path不能是目录: `+item.FullFilePath, nil)
			return
		}
		if fileInfo.Size() > gitUploadFileMaxSize {
			gsgin.GinResponseError(c, fmt.Sprintf(`文件 %s 大小超过限制(最大%dMB)`, item.FullFilePath, gitUploadFileMaxSize/(1024*1024)), nil)
			return
		}

		// 标准化 relative_file_path 并拼接目标路径
		relativePath := strings.ReplaceAll(item.RelativeFilePath, `\`, `/`)
		relativePath = strings.TrimLeft(relativePath, `/`)
		targetPath := codePath + `/` + relativePath

		// SCP传输到临时文件
		fileName := filepath.Base(item.FullFilePath)
		targetTempFile := targetPath + p_common.TBaseClient.GetUnique(`_upload`)
		gstool.FmtPrintlnLogTime(`[GitUploadFile] SCP上传开始: local=%s remote=%s fileSize=%d`, item.FullFilePath, targetTempFile, fileInfo.Size())
		if uploadErr := sshOnce.UploadFileProcessScp(targetTempFile, item.FullFilePath, func(int64, int64) {}); uploadErr != nil {
			gstool.FmtPrintlnLogTime(`[GitUploadFile] SCP上传失败: local=%s remote=%s err=%s`, item.FullFilePath, targetTempFile, uploadErr.Error())
			gsgin.GinResponseError(c, fmt.Sprintf(`文件上传失败[%s]: %s`, fileName, uploadErr.Error()), nil)
			return
		}

		// mv临时文件到最终路径（已存在则覆盖），复用已有的sshOnce连接
		gstool.FmtPrintlnLogTime(`[GitUploadFile] mv临时文件: %s -> %s`, targetTempFile, targetPath)
		if _, mvErr := sshOnce.RunCommandOnce(fmt.Sprintf(`mv %s %s`, targetTempFile, targetPath)); mvErr != nil {
			gstool.FmtPrintlnLogTime(`[GitUploadFile] mv失败: %s`, mvErr.Error())
			gsgin.GinResponseError(c, fmt.Sprintf(`移动文件失败[%s]: %s`, fileName, mvErr.Error()), nil)
			return
		}

		results = append(results, map[string]any{
			`remote_path`: targetPath,
			`file_name`:   fileName,
			`file_size`:   fileInfo.Size(),
			`git_id`:      gitId,
		})
	}

	gstool.FmtPrintlnLogTime(`[GitUploadFile] 全部上传完成, 共%d个文件`, len(results))
	gsgin.GinResponseSuccess(c, `文件上传成功`, map[string]any{
		`list`: results,
	})
}

// getGitInfoByGitId 通过 git_id 查询 tbl_git 获取信息，创建 SshTerminal 连接
func getGitInfoByGitId(gitId string) (map[string]any, *gsssh.SshTerminal, error) {
	if gitId == `` {
		return nil, nil, errors.New(`git_id不能为空`)
	}
	gitInfo, queryErr := common.DbMain.Client.QuickQuery(`tbl_git`, `*`, map[string]any{
		`id`: gitId,
	}).One()
	if queryErr != nil || len(gitInfo) == 0 {
		return nil, nil, fmt.Errorf(`未找到id为 "%s" 的Git配置`, gitId)
	}
	codePath := cast.ToString(gitInfo[`code_path`])
	if codePath == `` {
		return nil, nil, errors.New(`Git项目未配置code_path`)
	}
	sshId := gitInfo[`ssh_id`]
	sshConfig, sshConfigErr := common.DbMain.GetSshConfig(sshId)
	if sshConfigErr != nil {
		return nil, nil, fmt.Errorf(`获取SSH配置失败: %s`, sshConfigErr.Error())
	}
	if len(sshConfig) == 0 {
		return nil, nil, errors.New(`SSH配置为空`)
	}
	uniqueKey := p_common.TBaseClient.GetCombineKey(sshId, gitId)
	sshClient, sshErr := component.ShellClient.GetClient(sshConfig, uniqueKey, nil, nil, nil, nil)
	if sshErr != nil {
		return nil, nil, fmt.Errorf(`创建SSH连接失败: %s`, sshErr.Error())
	}
	return gitInfo, sshClient, nil
}

// GitCurrentBranchById 通过 git_id 查询当前分支
func GitCurrentBranchById(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, nil)
		return
	}
	gitId := cast.ToString(reqMap[`git_id`])
	gitInfo, sshClient, err := getGitInfoByGitId(gitId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(gitInfo[`code_path`])
	gstool.FmtPrintlnLogTime(`[GitCurrentBranchById] 开始查询远程分支 git_id=%s code_path=%s`, gitId, codePath)

	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gstool.FmtPrintlnLogTime(`[GitCurrentBranchById] 前置环境处理失败 err=%s`, prepareErr.Error())
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}

	command := p_shell.NewCommand()
	command.Cd(codePath)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, runErr := sshClient.RunCommandWait(command.GetCommand().ToStr(), time.Second*4)
	if runErr != nil {
		gstool.FmtPrintlnLogTime(`[GitCurrentBranchById] 查询失败 err=%s result=%q`, runErr.Error(), result)
		gsgin.GinResponseError(c, runErr.Error(), nil)
		return
	}
	gstool.FmtPrintlnLogTime(`[GitCurrentBranchById] 查询完成 git_id=%s result=%q`, gitId, result)
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitPull 通过 git_id 拉取当前分支最新代码
func GitPull(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, nil)
		return
	}
	gitId := cast.ToString(reqMap[`git_id`])
	gitInfo, sshClient, err := getGitInfoByGitId(gitId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(gitInfo[`code_path`])
	gstool.FmtPrintlnLogTime(`[GitPull] 开始拉取代码 git_id=%s code_path=%s`, gitId, codePath)

	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gstool.FmtPrintlnLogTime(`[GitPull] 前置环境处理失败 err=%s`, prepareErr.Error())
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}

	command := p_shell.NewCommand()
	command.Cd(codePath)
	command.GitIgnoreAll()
	command.GitCleanAll()
	command.GitFetch()
	command.GitPull()
	command.GitPullOriginCurrentBranch()
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, runErr := sshClient.RunCommandWait(command.GetCommand().ToStr(), getGitOperationTimeout(gitOperationPull))
	if runErr != nil {
		gstool.FmtPrintlnLogTime(`[GitPull] 拉取失败 err=%s result=%q`, runErr.Error(), result)
		gsgin.GinResponseError(c, runErr.Error(), result)
		return
	}
	gstool.FmtPrintlnLogTime(`[GitPull] 拉取完成 git_id=%s`, gitId)
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitChangeBranchById 通过 git_id 切换分支
func GitChangeBranchById(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, nil)
		return
	}
	gitId := cast.ToString(reqMap[`git_id`])
	branchName := cast.ToString(reqMap[`branch_name`])
	if branchName == `` {
		gsgin.GinResponseError(c, `branch_name不能为空`, nil)
		return
	}
	gitInfo, sshClient, err := getGitInfoByGitId(gitId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	codePath := cast.ToString(gitInfo[`code_path`])
	gstool.FmtPrintlnLogTime(`[GitChangeBranchById] 开始切换分支 git_id=%s code_path=%s branch=%s`, gitId, codePath, branchName)

	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gstool.FmtPrintlnLogTime(`[GitChangeBranchById] 前置环境处理失败 err=%s`, prepareErr.Error())
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}

	// 先查询当前分支
	command1 := p_shell.NewCommand()
	command1.Init()
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr(), getGitOperationTimeout(gitOperationBranchChange))
	currentBranch = CleanBranchName(currentBranch)

	command := p_shell.NewCommand()
	command.Cd(codePath)
	command.GitIgnoreAll()
	command.GitCleanAll()
	command.GitFetch()
	command.GitPull()
	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
	if currentBranch != branchName {
		command.GitCheckout(branchName)
	}
	command.GitPullOrigin(branchName)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, runErr := sshClient.RunCommandWait(command.GetCommand().ToStr(), getGitOperationTimeout(gitOperationBranchChange))
	if runErr != nil {
		gstool.FmtPrintlnLogTime(`[GitChangeBranchById] 切换分支失败 err=%s result=%q`, runErr.Error(), result)
		gsgin.GinResponseError(c, runErr.Error(), result)
		return
	}
	gstool.FmtPrintlnLogTime(`[GitChangeBranchById] 切换分支完成 git_id=%s branch=%s`, gitId, branchName)
	gsgin.GinResponseSuccess(c, ``, result)
}
