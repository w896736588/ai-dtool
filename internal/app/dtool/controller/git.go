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
	"sort"
	"strings"
	"sync"
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
	result, runErr := queryCurrentBranchInfo(sshClient, codePath, 40*time.Second)
	if runErr != nil {
		gsgin.GinResponseError(c, runErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, result.RawOutput)
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
	currentBranch = sshClient.FilterCommand(currentBranch)
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

	sseDistributeId := cast.ToString(reqMap[`sse_distribute_id`])
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: sseDistributeId,
	}

	writeSummary := func(msg string) {
		if sse != nil && sse.Sse != nil {
			sse.Send(msg)
		}
	}

	writeSummary(fmt.Sprintf("开始并发查询分组 [%s] 下 %d 个项目（每项最多5秒）...\n", cast.ToString(groupInfo[`name`]), len(gitList)))

	type gitItemResult struct {
		Index int
		Data  map[string]any
		Line  string
	}

	resultChan := make(chan gitItemResult, len(gitList))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 8)

	for idx, gitConfig := range gitList {
		wg.Add(1)
		go func(i int, item map[string]any) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			itemName := cast.ToString(item[`name`])
			itemPath := cast.ToString(item[`code_path`])
			sshId := cast.ToString(item[`ssh_id`])
			itemResult := map[string]any{
				`id`:            cast.ToString(item[`id`]),
				`name`:          itemName,
				`code_path`:     itemPath,
				`ssh_id`:        sshId,
				`local_branch`:  `N/A`,
				`remote_branch`: `N/A`,
				`ok`:            false,
				`error`:         ``,
			}

			if itemPath == `` || sshId == `` {
				itemResult[`error`] = `缺少 code_path 或 ssh_id 配置`
				resultChan <- gitItemResult{Index: i, Data: itemResult, Line: fmt.Sprintf("%s | %s | 失败: %s", itemName, itemPath, itemResult[`error`])}
				return
			}

			sshConfig, sshErr := common.DbMain.GetSshConfig(sshId)
			if sshErr != nil || len(sshConfig) == 0 {
				itemResult[`error`] = `SSH配置不存在`
				resultChan <- gitItemResult{Index: i, Data: itemResult, Line: fmt.Sprintf("%s | %s | 失败: %s", itemName, itemPath, itemResult[`error`])}
				return
			}

			uniqueKey := p_common.TBaseClient.GetCombineKey(
				sshId,
				`group_branch_`+gitGroupId+`_`+cast.ToString(item[`id`])+`_`+cast.ToString(time.Now().UnixNano()),
			)
			sshClient, clientErr := component.ShellClient.GetClient(sshConfig, uniqueKey, sse, func(s string) []string {
				return []string{p_common.TBaseClient.FilterTerminalChars(s)}
			}, nil, nil)
			if clientErr != nil {
				itemResult[`error`] = clientErr.Error()
				resultChan <- gitItemResult{Index: i, Data: itemResult, Line: fmt.Sprintf("%s | %s | 失败: %s", itemName, itemPath, itemResult[`error`])}
				return
			}

			branchInfo, queryErr := queryCurrentBranchInfo(sshClient, itemPath, 5*time.Second)
			if queryErr != nil {
				itemResult[`error`] = queryErr.Error()
				resultChan <- gitItemResult{Index: i, Data: itemResult, Line: fmt.Sprintf("%s | %s | 失败: %s", itemName, itemPath, itemResult[`error`])}
				return
			}

			itemResult[`local_branch`] = branchInfo.LocalBranch
			itemResult[`remote_branch`] = branchInfo.RemoteBranch
			itemResult[`ok`] = true
			resultChan <- gitItemResult{
				Index: i,
				Data:  itemResult,
				Line:  fmt.Sprintf("%s | %s | 当前:%s | 远程:%s", itemName, itemPath, branchInfo.LocalBranch, branchInfo.RemoteBranch),
			}
		}(idx, gitConfig)
	}

	wg.Wait()
	close(resultChan)

	collected := make([]gitItemResult, 0, len(gitList))
	for item := range resultChan {
		collected = append(collected, item)
	}
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].Index < collected[j].Index
	})

	resultList := make([]map[string]any, 0, len(collected))
	summaryLines := make([]string, 0, len(collected)+2)
	summaryLines = append(summaryLines, fmt.Sprintf("Git分组 [%s] 分支总览（每行一个项目）", cast.ToString(groupInfo[`name`])))
	summaryLines = append(summaryLines, "名称 | 路径 | 当前分支 | 远程分支/错误")
	for _, item := range collected {
		resultList = append(resultList, item.Data)
		summaryLines = append(summaryLines, item.Line)
	}
	summaryText := strings.Join(summaryLines, "\n")
	writeSummary("\n" + summaryText)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`git_group_id`:  gitGroupId,
		`group_name`:    cast.ToString(groupInfo[`name`]),
		`list`:          resultList,
		`summary_lines`: summaryLines,
		`summary_text`:  summaryText,
	})
}

type GitCurrentBranchInfo struct {
	LocalBranch  string
	RemoteBranch string
	RawOutput    string
}

func queryCurrentBranchInfo(sshClient *gsssh.SshTerminal, codePath string, timeout time.Duration) (*GitCurrentBranchInfo, error) {
	command := p_shell.NewCommand()
	command.Cd(codePath)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	output, err := sshClient.RunCommandWait(command.GetCommand().ToStr(), timeout)
	if err != nil {
		return nil, err
	}
	localBranch, remoteBranch := parseBranchFromCurrentBranchOutput(output)
	if localBranch == `` {
		localBranch = `N/A`
	}
	if remoteBranch == `` {
		remoteBranch = `N/A`
	}
	return &GitCurrentBranchInfo{
		LocalBranch:  localBranch,
		RemoteBranch: remoteBranch,
		RawOutput:    output,
	}, nil
}

func parseBranchFromCurrentBranchOutput(output string) (string, string) {
	text := p_common.TBaseClient.FilterTerminalChars(output)
	lines := strings.Split(text, "\n")
	localBranch := ``
	remoteBranch := ``

	findNextValue := func(start int) string {
		for i := start; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == `` {
				continue
			}
			if strings.Contains(line, `当前分支：`) || strings.Contains(line, `远程分支：`) {
				continue
			}
			return line
		}
		return ``
	}

	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, `当前分支：`) && localBranch == `` {
			localBranch = findNextValue(idx + 1)
			continue
		}
		if strings.Contains(trimmed, `远程分支：`) && remoteBranch == `` {
			remoteBranch = findNextValue(idx + 1)
			continue
		}
	}
	return localBranch, remoteBranch
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
