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
	"regexp"
	"sort"
	"strconv"
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
	// 所有通过 SSH 的 Git 操作前，默认先执行“目录安全 + 保存账号密码”。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}
	command := p_shell.NewCommand()
	//command.Sudo()
	command.Cd(codePath)
	command.Echo(`当前分支：`)
	command.GitShowBranch()
	command.Echo(`远程分支：`)
	command.GitShowOriginBranch()
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), time.Second*4)
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
	// 所有通过 SSH 的 Git 操作前，默认先执行“目录安全 + 保存账号密码”。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}
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
	// 所有通过 SSH 的 Git 操作前，默认先执行“目录安全 + 保存账号密码”。
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
	// 所有通过 SSH 的 Git 操作前，默认先执行“目录安全 + 保存账号密码”。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}
	command1 := p_shell.NewCommand()
	command1.Init()
	//command.Sudo() 不要用sudo否则服务器会提示输入密码，导致执行被卡死
	command1.Cd(codePath)
	command1.GitShowBranch()
	currentBranch, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr(), getGitOperationTimeout(gitOperationPull))
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
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), getGitOperationTimeout(gitOperationPull))
	gsgin.GinResponseSuccess(c, ``, result)
}

// GitRemoteBranchList 查询指定仓库的全部远程分支
func GitRemoteBranchList(c *gin.Context) {
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

	// 所有通过 SSH 的 Git 操作前，默认先执行“目录安全 + 保存账号密码”。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
		return
	}

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
	reqMap, sshClient, _, err := getGitComponent(c)
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
	// 所有通过 SSH 的 Git 操作前，默认先执行“目录安全 + 保存账号密码”。
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
		gsgin.GinResponseError(c, runErr.Error(), nil)
		return
	}
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
	// 所有通过 SSH 的 Git 操作前，默认先执行“目录安全 + 保存账号密码”。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
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
	// 所有通过 SSH 的 Git 操作前，默认先执行“目录安全 + 保存账号密码”。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
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

	// 统一发送汇总信息（串行流程下无需并发锁）
	writeSummary := func(msg string) {
		if sse != nil && sse.Sse != nil {
			sse.Send(msg)
		}
	}

	totalCount := len(gitList)
	resultList := make([]map[string]any, 0, totalCount)
	summaryLines := make([]string, 0, totalCount+4)
	summaryLines = append(summaryLines, fmt.Sprintf("### Git分组 `%s` 分支总览", cast.ToString(groupInfo[`name`])))
	summaryLines = append(summaryLines, "")
	summaryHeaderIndex := len(summaryLines)
	summaryLines = append(summaryLines, "| 名称 | 路径 | 当前分支 | 远程分支/错误 |")
	summaryLines = append(summaryLines, "| --- | --- | --- | --- |")
	// 先输出表头，后续只输出表格行，不输出其他日志文本
	summaryLines[summaryHeaderIndex] = "| 鍚嶇О | 璺緞 | 褰撳墠鍒嗘敮 | 杩滅▼鍒嗘敮/閿欒 | 鏄惁鏈変汉浣跨敤 |"
	summaryLines[summaryHeaderIndex+1] = "| --- | --- | --- | --- | --- |"
	summaryLines[summaryHeaderIndex] = "| 名称 | 路径 | 当前分支 | 远程分支/错误 | 是否有人使用 |"
	writeSummary("\n" + strings.Join(summaryLines, "\n") + "\n")

	for _, item := range gitList {
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
			`usage_status`:  `N/A`,
			`usage_owners`:  []string{},
			`ok`:            false,
			`error`:         ``,
		}
		tableLine := ``
		if itemPath == `` || sshId == `` {
			itemResult[`error`] = `缺少 code_path 或 ssh_id 配置`
			tableLine = fmt.Sprintf(
				"| %s | %s | %s | %s |",
				escapeMarkdownTableCell(itemName),
				escapeMarkdownTableCell(itemPath),
				escapeMarkdownTableCell(`N/A`),
				escapeMarkdownTableCell(`失败: `+cast.ToString(itemResult[`error`])),
			)
		} else {
			sshConfig, sshErr := common.DbMain.GetSshConfig(sshId)
			if sshErr != nil || len(sshConfig) == 0 {
				itemResult[`error`] = `SSH配置不存在`
				tableLine = fmt.Sprintf(
					"| %s | %s | %s | %s |",
					escapeMarkdownTableCell(itemName),
					escapeMarkdownTableCell(itemPath),
					escapeMarkdownTableCell(`N/A`),
					escapeMarkdownTableCell(`失败: `+cast.ToString(itemResult[`error`])),
				)
			} else {
				uniqueKey := p_common.TBaseClient.GetCombineKey(
					sshId,
					`group_branch_`+gitGroupId+`_`+cast.ToString(item[`id`])+`_`+cast.ToString(time.Now().UnixNano()),
				)
				// 组内批量查询只保留汇总结果，避免每个项目的原始 shell 输出刷屏
				silentSse := &p_sse.SseShell{}
				sshClient, clientErr := component.ShellClient.GetClient(sshConfig, uniqueKey, silentSse, nil, nil, nil)
				if clientErr != nil {
					itemResult[`error`] = clientErr.Error()
					tableLine = fmt.Sprintf(
						"| %s | %s | %s | %s |",
						escapeMarkdownTableCell(itemName),
						escapeMarkdownTableCell(itemPath),
						escapeMarkdownTableCell(`N/A`),
						escapeMarkdownTableCell(`失败: `+cast.ToString(itemResult[`error`])),
					)
				} else if prepareErr := prepareGitOperationEnv(sshClient, itemPath); prepareErr != nil {
					// 复用统一预处理，避免分组查询单独实现认证方案。
					itemResult[`error`] = prepareErr.Error()
					tableLine = fmt.Sprintf(
						"| %s | %s | %s | %s |",
						escapeMarkdownTableCell(itemName),
						escapeMarkdownTableCell(itemPath),
						escapeMarkdownTableCell(`N/A`),
						escapeMarkdownTableCell(`失败: `+cast.ToString(itemResult[`error`])),
					)
				} else {
					branchInfo, queryErr := queryCurrentBranchInfo(sshClient, itemPath, 5*time.Second)
					if queryErr != nil {
						itemResult[`error`] = queryErr.Error()
						tableLine = fmt.Sprintf(
							"| %s | %s | %s | %s |",
							escapeMarkdownTableCell(itemName),
							escapeMarkdownTableCell(itemPath),
							escapeMarkdownTableCell(`N/A`),
							escapeMarkdownTableCell(`失败: `+cast.ToString(itemResult[`error`])),
						)
					} else {
						usageInfo := queryBranchUsageInfo(sshClient, itemPath, branchInfo, 8*time.Second)
						itemResult[`local_branch`] = branchInfo.LocalBranch
						itemResult[`remote_branch`] = branchInfo.RemoteBranch
						itemResult[`usage_status`] = usageInfo.UsageDisplay
						itemResult[`usage_owners`] = usageInfo.Owners
						itemResult[`ok`] = true
						tableLine = fmt.Sprintf(
							"| %s | %s | %s | %s |",
							escapeMarkdownTableCell(itemName),
							escapeMarkdownTableCell(itemPath),
							escapeMarkdownTableCell(branchInfo.LocalBranch),
							escapeMarkdownTableCell(branchInfo.RemoteBranch),
						)
					}
				}
			}
		}

		tableLine = appendMarkdownTableUsageCell(tableLine, cast.ToString(itemResult[`usage_status`]))
		writeSummary(tableLine + "\n")

		resultList = append(resultList, itemResult)
		summaryLines = append(summaryLines, tableLine)
	}
	summaryText := strings.Join(summaryLines, "\n")
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

type GitBranchUsageInfo struct {
	UsageDisplay string
	Owners       []string
}

const (
	gitDefaultCommandTimeout    = 40 * time.Second
	gitBranchChangeTimeout      = 5 * time.Minute
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

// queryCurrentBranchInfo 合并本地+远程分支查询为单次SSH命令，减少网络往返
func queryCurrentBranchInfo(sshClient *gsssh.SshTerminal, codePath string, timeout time.Duration) (*GitCurrentBranchInfo, error) {
	const branchSep = `__DT_BRANCH_SEP__`
	combinedCmd := p_shell.NewCommand()
	combinedCmd.Cd(codePath)
	combinedCmd.Echo(`__DT_LOCAL_BRANCH_BEGIN__`)
	combinedCmd.GitShowBranch()
	combinedCmd.Echo(`__DT_LOCAL_BRANCH_END__`)
	combinedCmd.Echo(`__DT_REMOTE_BRANCH_BEGIN__`)
	combinedCmd.GitShowOriginBranch()
	combinedCmd.Echo(`__DT_REMOTE_BRANCH_END__`)

	combinedOutput, err := sshClient.RunCommandWait(combinedCmd.GetCommand().ToStr(), timeout)
	if err != nil {
		return nil, err
	}
	return parseCurrentBranchInfoFromCombinedOutput(combinedOutput), nil

	lines := strings.Split(combinedOutput, "\n")
	gstool.FmtPrintlnLogTime(`合并查询结果：%s`, gstool.JsonEncode(lines))

	// 定位分隔符行（跳过首行命令回显）
	sepIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(p_common.TBaseClient.FilterTerminalChars(lines[i])) == branchSep {
			sepIdx = i
			break
		}
	}

	// 解析本地分支：命令回显之后、分隔符之前的第一行
	localBranch := ``
	if sepIdx > 1 {
		candidate := strings.TrimSpace(p_common.TBaseClient.FilterTerminalChars(lines[1]))
		// 过滤掉终端提示符等噪声，避免把控制串当成本地分支
		if !isBranchParseNoise(candidate) {
			localBranch = candidate
		}
	} else if len(lines) > 1 {
		candidate := strings.TrimSpace(p_common.TBaseClient.FilterTerminalChars(lines[1]))
		// 过滤掉终端提示符等噪声，避免把控制串当成本地分支
		if !isBranchParseNoise(candidate) {
			localBranch = candidate
		}
	}

	// 解析远程分支：分隔符之后第一个非空行
	remoteBranch := ``
	if sepIdx >= 0 {
		for i := sepIdx + 1; i < len(lines); i++ {
			cleaned := strings.TrimSpace(p_common.TBaseClient.FilterTerminalChars(lines[i]))
			// 仅接受合法的远程分支行（如 "<sha>\trefs/heads/xxx"），
			// 终端提示符/控制串等噪声直接跳过
			if cleaned == `` || isBranchParseNoise(cleaned) || !isLikelyRemoteBranch(cleaned) {
				continue
			}
			if cleaned != `` {
				remoteBranch = cleaned
				break
			}
		}
	}

	return &GitCurrentBranchInfo{
		LocalBranch:  localBranch,
		RemoteBranch: remoteBranch,
		RawOutput:    buildCurrentBranchDisplayOutput(localBranch, remoteBranch),
	}, nil
}

// parseCurrentBranchInfoFromCombinedOutput 解析组内批量查询使用的合并命令输出。
// 这里统一复用带 begin/end 标记的分支解析逻辑，避免依赖脆弱的固定行号。
func parseCurrentBranchInfoFromCombinedOutput(output string) *GitCurrentBranchInfo {
	localBranch, remoteBranch := parseBranchFromCurrentBranchOutput(output)
	return &GitCurrentBranchInfo{
		LocalBranch:  localBranch,
		RemoteBranch: remoteBranch,
		RawOutput:    buildCurrentBranchDisplayOutput(localBranch, remoteBranch),
	}
}

func runRemoteBranchQuery(sshClient *gsssh.SshTerminal, codePath string, timeout time.Duration) (string, error) {
	remoteCommand := p_shell.NewCommand()
	remoteCommand.Cd(codePath)
	remoteCommand.Echo(`远程分支：`)
	remoteCommand.GitShowOriginBranch()
	return sshClient.RunCommandWait(remoteCommand.GetCommand().ToStr(), timeout)
}

func parseBranchFromCurrentBranchOutput(output string) (string, string) {
	text := p_common.TBaseClient.FilterTerminalChars(output)
	lines := strings.Split(text, "\n")
	localBranch := ``
	remoteBranch := ``
	localBegin := false
	localEnd := false
	remoteBegin := false
	remoteEnd := false

	findNextValue := func(start int) string {
		for i := start; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == `` {
				continue
			}
			if isBranchParseNoise(line) {
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
		if strings.Contains(trimmed, `__DT_LOCAL_BRANCH_BEGIN__`) {
			localBegin = true
			continue
		}
		if strings.Contains(trimmed, `__DT_LOCAL_BRANCH_END__`) {
			localEnd = true
			continue
		}
		if strings.Contains(trimmed, `__DT_REMOTE_BRANCH_BEGIN__`) {
			remoteBegin = true
			continue
		}
		if strings.Contains(trimmed, `__DT_REMOTE_BRANCH_END__`) {
			remoteEnd = true
			continue
		}

		if localBegin && !localEnd && localBranch == `` {
			if isLikelyLocalBranch(trimmed) {
				localBranch = trimmed
				continue
			}
		}
		if remoteBegin && !remoteEnd && remoteBranch == `` {
			if isLikelyRemoteBranch(trimmed) {
				remoteBranch = trimmed
				continue
			}
		}

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

func isLikelyLocalBranch(line string) bool {
	if isBranchParseNoise(line) {
		return false
	}
	if strings.Contains(line, " ") || strings.Contains(line, "\t") {
		return false
	}
	return true
}

func isLikelyRemoteBranch(line string) bool {
	if isBranchParseNoise(line) {
		return false
	}
	if strings.Contains(line, `refs/heads/`) {
		return true
	}
	return false
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

func normalizeRemoteBranchDisplay(value string) string {
	v := strings.TrimSpace(value)
	if v == "" || v == "N/A" {
		return "N/A"
	}
	if strings.HasPrefix(v, "refs/heads/") {
		return strings.TrimPrefix(v, "refs/heads/")
	}
	fields := strings.Fields(v)
	if len(fields) >= 2 {
		last := fields[len(fields)-1]
		if strings.HasPrefix(last, "refs/heads/") {
			return strings.TrimPrefix(last, "refs/heads/")
		}
		return last
	}
	return v
}

// queryBranchUsageInfo 先检查远程分支，再回退到最近 2 小时工作区文件属主。
// queryBranchUsageInfo checks remote branch first, then falls back to recent workspace owners within 2 hours.
func queryBranchUsageInfo(sshClient *gsssh.SshTerminal, codePath string, branchInfo *GitCurrentBranchInfo, timeout time.Duration) GitBranchUsageInfo {
	// 关键判断 / Key decision: 本地分支不存在时，直接视为无人使用并返回统一占位符。
	if strings.TrimSpace(branchInfo.LocalBranch) == "" {
		return GitBranchUsageInfo{
			UsageDisplay: gitBranchUsageNoneDisplay,
			Owners:       []string{},
		}
	}

	remoteBranch := normalizeRemoteBranchDisplay(branchInfo.RemoteBranch)
	// 关键判断 / Key decision: 只要存在远程分支，就直接标记为“有人使用”。
	if remoteBranch != "" && remoteBranch != "N/A" {
		return GitBranchUsageInfo{
			UsageDisplay: gitBranchUsageUsedDisplay,
			Owners:       []string{},
		}
	}

	owners, err := queryRecentWorkspaceOwners(sshClient, codePath, timeout)
	if err != nil {
		return GitBranchUsageInfo{
			UsageDisplay: gitBranchUsageNoneDisplay,
			Owners:       []string{},
		}
	}
	return GitBranchUsageInfo{
		UsageDisplay: buildBranchUsageDisplay(branchInfo.LocalBranch, remoteBranch, owners...),
		Owners:       owners,
	}
}

// queryRecentWorkspaceOwners 查询 git status 变更文件中最近 2 小时有改动的文件属主。
// queryRecentWorkspaceOwners returns owners of changed files touched within the last 2 hours.
func queryRecentWorkspaceOwners(sshClient *gsssh.SshTerminal, codePath string, timeout time.Duration) ([]string, error) {
	statusCommand := p_shell.NewCommand()
	statusCommand.Cd(codePath)
	statusCommand.SetCommand(`git status --porcelain --untracked-files=all`)
	statusOutput, err := sshClient.RunCommandWait(statusCommand.GetCommand().ToStr(), timeout)
	if err != nil {
		return nil, err
	}

	fileList := parseGitStatusEntries(statusOutput)
	if len(fileList) == 0 {
		return []string{}, nil
	}

	statCommand := p_shell.NewCommand()
	statCommand.Cd(codePath)
	statCommand.SetCommand(buildStatOwnersCommand(fileList))
	statOutput, err := sshClient.RunCommandWait(statCommand.GetCommand().ToStr(), timeout)
	if err != nil {
		return nil, err
	}
	return parseRecentUsageOwners(statOutput, time.Now(), 2*time.Hour), nil
}

// parseGitStatusEntries 解析 git status --porcelain 的变更文件路径。
// parseGitStatusEntries extracts changed file paths from git status --porcelain output.
func parseGitStatusEntries(output string) []string {
	lines := strings.Split(p_common.TBaseClient.FilterTerminalChars(output), "\n")
	seen := make(map[string]struct{})
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || len(line) < 4 {
			continue
		}
		pathPart := strings.TrimSpace(line[3:])
		if pathPart == "" {
			continue
		}
		// 重命名场景 / Rename case: 使用新路径判断最近活跃状态。
		if strings.Contains(pathPart, " -> ") {
			parts := strings.Split(pathPart, " -> ")
			pathPart = strings.TrimSpace(parts[len(parts)-1])
		}
		pathPart = strings.Trim(pathPart, `"`)
		if pathPart == "" {
			continue
		}
		if _, exist := seen[pathPart]; exist {
			continue
		}
		seen[pathPart] = struct{}{}
		result = append(result, pathPart)
	}
	return result
}

// buildStatOwnersCommand 构造批量查询属主和 mtime 的 shell 命令。
// buildStatOwnersCommand builds a shell command for querying owner and mtime in batch.
func buildStatOwnersCommand(fileList []string) string {
	commandList := make([]string, 0, len(fileList))
	for _, filePath := range fileList {
		quotedPath := quoteShellArg(filePath)
		commandList = append(commandList, fmt.Sprintf(`[ -e %s ] && stat -c '%%U|%%Y|%%n' -- %s`, quotedPath, quotedPath))
	}
	return strings.Join(commandList, " ; ")
}

// parseRecentUsageOwners 过滤最近时间窗口内修改文件的属主。
// parseRecentUsageOwners filters owners whose files were modified within the recent time window.
func parseRecentUsageOwners(output string, now time.Time, recentWindow time.Duration) []string {
	lines := strings.Split(p_common.TBaseClient.FilterTerminalChars(output), "\n")
	ownerSet := make(map[string]struct{})
	ownerList := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, "|", 3)
		if len(parts) < 3 {
			continue
		}
		owner := strings.TrimSpace(parts[0])
		unixTS, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil || owner == "" {
			continue
		}
		modifiedAt := time.Unix(unixTS, 0)
		// 关键判断 / Key decision: 仅保留最近 2 小时内的文件属主。
		if now.Sub(modifiedAt) > recentWindow {
			continue
		}
		if _, exist := ownerSet[owner]; exist {
			continue
		}
		ownerSet[owner] = struct{}{}
		ownerList = append(ownerList, owner)
	}
	sort.Strings(ownerList)
	return ownerList
}

// buildBranchUsageDisplay 生成“是否有人使用”列的显示内容。
// buildBranchUsageDisplay builds the display text for the usage column.
func buildBranchUsageDisplay(localBranch, remoteBranch string, owners ...string) string {
	// 关键判断 / Key decision: 没有本地分支时，统一展示为 "-"。
	if strings.TrimSpace(localBranch) == "" {
		return gitBranchUsageNoneDisplay
	}
	if normalized := normalizeRemoteBranchDisplay(remoteBranch); normalized != "" && normalized != "N/A" {
		return gitBranchUsageUsedDisplay
	}
	if len(owners) > 0 {
		return strings.Join(owners, ", ")
	}
	return gitBranchUsageNoneDisplay
}

// appendMarkdownTableUsageCell 为现有 Markdown 表格行追加“是否有人使用”列。
// appendMarkdownTableUsageCell appends the usage column to an existing Markdown table row.
func appendMarkdownTableUsageCell(line, usage string) string {
	trimmedLine := strings.TrimRight(line, " ")
	if strings.HasSuffix(trimmedLine, "|") {
		trimmedLine = strings.TrimSuffix(trimmedLine, "|")
	}
	return trimmedLine + " | " + escapeMarkdownTableCell(usage) + " |"
}

// quoteShellArg 使用单引号安全转义 shell 参数。
// quoteShellArg safely escapes a shell argument using single quotes.
func quoteShellArg(value string) string {
	return `'` + strings.ReplaceAll(value, `'`, `'"'"'`) + `'`
}

// buildCurrentBranchDisplayOutput 组装“当前分支”接口的展示输出。
// buildCurrentBranchDisplayOutput builds display output for the current branch API.
func buildCurrentBranchDisplayOutput(localBranch, remoteBranch string) string {
	local := strings.TrimSpace(localBranch)
	remote := strings.TrimSpace(remoteBranch)
	if local == `` {
		local = `N/A`
	}
	if remote == `` {
		remote = `N/A`
	}
	return strings.Join([]string{
		`当前分支：`,
		local,
		`远程分支：`,
		remote,
	}, "\n")
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
	// 所有通过 SSH 的 Git 操作前，默认先执行“目录安全 + 保存账号密码”。
	if prepareErr := prepareGitOperationEnv(sshClient, codePath); prepareErr != nil {
		gsgin.GinResponseError(c, prepareErr.Error(), nil)
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
