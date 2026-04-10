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

const (
	// sshPoolSize 连接池固定槽位数量
	sshPoolSize = 5
	// sshPoolCoolDown 连接使用后（含新建后）的冷却等待时间
	sshPoolCoolDown = time.Second
)

// poolSlot 连接池单个槽位，管理一个 SSH 连接的生命周期和冷却状态。
type poolSlot struct {
	client        *gsssh.SshTerminal // SSH 客户端实例
	shellClientId string             // 当前绑定的 client key（用于判断复用）
	sshConfig     map[string]any     // 当前绑定的 SSH 配置
	lastUsedAt    time.Time          // 最后释放/创建时间（用于冷却判断）
	inUse         bool               // 是否正在被占用
}

// sshConnectionPool 固定大小 SSH 连接池，包含冷却等待机制。
// 新建连接或复用连接释放后均需等待 sshPoolCoolDown 后才能再次分配。
type sshConnectionPool struct {
	slots [sshPoolSize]poolSlot
	mu    sync.Mutex
}

// newSSHConnectionPool 创建一个空的 SSH 连接池。
func newSSHConnectionPool() *sshConnectionPool {
	return &sshConnectionPool{}
}

// Acquire 从连接池中获取一个可用的 SSH 客户端。
// 规则：
//   - 优先复用已有且 key 匹配的空闲连接（已过冷却期）
//   - 无匹配时在空闲槽位上新建连接
//   - 所有槽位忙或未冷却时轮询等待（100ms 间隔）
//   - 新建连接的 lastUsedAt 设为 now，确保首次使用前也经过冷却期
func (p *sshConnectionPool) Acquire(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell) (*gsssh.SshTerminal, error) {
	for {
		p.mu.Lock()
		for i := range p.slots {
			slot := &p.slots[i]
			if slot.inUse {
				continue
			}
			// 检查是否已过冷却期
			if time.Since(slot.lastUsedAt) < sshPoolCoolDown {
				continue
			}
			// 命中已有连接且 key 一致 → 直接复用
			if slot.client != nil && slot.shellClientId == shellClientId {
				slot.inUse = true
				p.mu.Unlock()
				return slot.client, nil
			}
			// 需要新建或重建连接
			if slot.client != nil {
				slot.client.CloseTerminal()
			}
			client, err := component.ShellClient.GetClient(sshConfig, shellClientId, sse, nil, nil, nil)
			if err != nil {
				p.mu.Unlock()
				return nil, err
			}
			slot.client = client
			slot.shellClientId = shellClientId
			slot.sshConfig = sshConfig
			slot.lastUsedAt = time.Now() // 新建后也设冷却起点
			slot.inUse = true
			p.mu.Unlock()
			return client, nil
		}
		p.mu.Unlock()
		// 所有槽位忙或未冷却，短暂等待重试
		time.Sleep(100 * time.Millisecond)
	}
}

// Release 释放指定索引的槽位并记录冷却起始时间。
func (p *sshConnectionPool) Release(idx int) {
	if idx < 0 || idx >= sshPoolSize {
		return
	}
	p.mu.Lock()
	p.slots[idx].inUse = false
	p.slots[idx].lastUsedAt = time.Now()
	p.mu.Unlock()
}

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
		gsgin.GinResponseError(c, runErr.Error(), result)
		return
	}
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
