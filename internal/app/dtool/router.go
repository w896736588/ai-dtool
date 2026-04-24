package dtool

import (
	"dev_tool/internal/app/dtool/controller"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/middleware"
	"dev_tool/internal/pkg/p_define"
	"dev_tool/internal/pkg/p_gin"
	"errors"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
)

func InitRouter(tGin *p_gin.Gin) {
	// 注册 SafeAuth 中间件（需要在基础路由之后，其他受保护路由之前）
	// 但白名单接口需要在中间件之前注册，所以这里采用另一种方式：
	// 1. 先注册白名单接口
	baseRouter(tGin)

	// 2. 注册 SafeAuth 中间件到所有后续路由
	tGin.UseMiddleware(middleware.SafeAuthMiddleware())

	toolRouter(tGin)
	redisRouter(tGin)
	phpRouter(tGin)
	supervisorRouter(tGin)
	gitRouter(tGin)
	gitLabTokenRouter(tGin)
	globalSetRouter(tGin)
	codeRouter(tGin)
	//initSocket()
	setRouter(tGin)
	setGroupRouter(tGin)
	setStar(tGin)
	setMarkdown(tGin)
	setMemoryFragment(tGin)
	homeTask(tGin)
	shellOut(tGin)
	variableRouter(tGin)
	smartLink(tGin)
	docker(tGin)
	api(tGin)
	apiUse(tGin)
	tGin.GinPost(`/test/multiformdata`, func(c *gin.Context) {
		// 解析 multipart/form-data
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Failed to parse form data",
				"details": err.Error(),
			})
			return
		}

		// 获取所有普通字段
		allValues := make(map[string][]string)
		for key, values := range form.Value {
			allValues[key] = values
		}

		// 获取所有文件
		allFiles := make(map[string][]*multipart.FileHeader)
		for key, files := range form.File {
			allFiles[key] = files
		}

		// 统计信息
		fileInfos := []gin.H{}
		for fieldName, files := range allFiles {
			for _, file := range files {
				fileInfos = append(fileInfos, gin.H{
					"field_name": fieldName,
					"filename":   file.Filename,
					"size":       file.Size,
					"header":     file.Header,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"all_fields": allValues,
			"all_files":  fileInfos,
			"summary": gin.H{
				"field_count": len(allValues),
				"file_count":  len(fileInfos),
			},
		})
		return
	})
}

func toolRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/ToolPortProcessList`, controller.ToolPortProcessList)
	tGin.GinPost(`/api/ToolPortProcessKill`, controller.ToolPortProcessKill)
	tGin.GinPost(`/api/ToolManagedProcessStatus`, controller.ToolManagedProcessStatus)
	tGin.GinPost(`/api/ToolManagedProcessEnsureRunning`, controller.ToolManagedProcessEnsureRunning)
	tGin.GinPost(`/api/ToolManagedProcessStart`, controller.ToolManagedProcessStart)
	tGin.GinPost(`/api/ToolManagedProcessStop`, controller.ToolManagedProcessStop)
	tGin.GinPost(`/api/ToolManagedProcessRestart`, controller.ToolManagedProcessRestart)
	tGin.GinPost(`/api/ToolManagedProcessLogTail`, controller.ToolManagedProcessLogTail)
}

// 基础接口
func baseRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/BaseLogin`, controller.BaseLogin)                       //Safe 登录
	tGin.GinPost(`/api/BaseLoginStatus`, controller.BaseLoginStatus)           //Safe 登录状态检查
	tGin.GinPost(`/api/BaseRegisterService`, controller.BaseRegisterService)   //注册各类服务 CheckUnikeyExist
	tGin.GinPost(`/api/BaseCheckUnikeyExist`, controller.BaseCheckUnikeyExist) //检查unikey是否已经登录注册
	tGin.GinPost(`/api/BaseSshList`, controller.BaseSshList)                   //ssh列表
	tGin.GinPost(`/api/Ip`, controller.Ip)                                     //外网IP
	tGin.GinPost(`/api/GetLocalIP`, controller.GetLocalIP)                     //局域网IP
	tGin.GinPost(`/api/Upload`, controller.Upload)                             //上传文件
	tGin.GinGet(`/api/download/:name`, controller.DownloadWebFile)             //下载 web/download 目录文件
	tGin.GinGet(`/web/download/:name`, controller.DownloadWebFile)             //兼容 web/download 直链下载
}

// redis相关
func redisRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/RedisAvailableList`, controller.RedisAvailableList) //可用的redis列表
	tGin.GinPost(`/api/RedisSearch`, controller.RedisSearch)               //查询某个key
	tGin.GinPost(`/api/RedisKeys`, controller.RedisKeys)                   //模糊搜索key
	tGin.GinPost(`/api/RedisKeysType`, controller.RedisKeysType)           //批量获取key缓存类型
	tGin.GinPost(`/api/RedisKeyType`, controller.RedisKeyType)             //获取key类型
	tGin.GinPost(`/api/RedisSaveString`, controller.RedisSaveString)       //保存string
	tGin.GinPost(`/api/RedisDelKey`, controller.RedisDelKey)               //删除key
	tGin.GinPost(`/api/RedisDelSub`, controller.RedisDelSub)               //删除sub key
	tGin.GinPost(`/api/RedisEditTtl`, controller.RedisEditTtl)             //更改ttl
	tGin.GinPost(`/api/RedisDeleteAll`, controller.RedisDelAllKey)         //删除所有缓存
	tGin.GinPost(`/api/RedisCreateCache`, controller.RedisCreateCache)     //创建缓存
	tGin.GinPost(`/api/RedisEditSub`, controller.RedisEditSub)             //编辑二级缓存
}

// php相关
func phpRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/PhpUnserialize`, controller.PhpPhpUnSerialize)   //PHP反序列化
	tGin.GinPost(`/api/PhpUnserialize2`, controller.PhpPhpUnSerialize2) //PHP反序列化
}

// 消费者相关
func supervisorRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/SupervisorRestartAll`, controller.SupervisorRestartAll) //重启所有消费者
	tGin.GinPost(`/api/SupervisorStopAll`, controller.SupervisorStopAll)       //重启所有消费者
	tGin.GinPost(`/api/SupervisorStatusList`, controller.SupervisorStatusList) //查看消费者状态
	tGin.GinPost(`/api/SupervisorConfigShow`, controller.SupervisorConfigShow) //查看消费者配置
	tGin.GinPost(`/api/SupervisorRestart`, controller.SupervisorRestart)       //重启单个消费者
	tGin.GinPost(`/api/SupervisorStop`, controller.SupervisorStop)             //重启单个消费者
	tGin.GinPost(`/api/SupervisorConfList`, controller.SupervisorConfList)     //查看所有的配置
	tGin.GinPost(`/api/SupervisorConfigList`, controller.SupervisorConfigList) //配置的supervisor
}

// git相关
func gitRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/GitQueryCurrentBranch`, controller.GitCurrentBranch)      //查询当前分支
	tGin.GinPost(`/api/GitChangeBranch`, controller.GitChangeBranch)             //切换分支
	tGin.GinPost(`/api/GitChangeBranchRemote`, controller.GitChangeBranchRemote) //切换远程分支
	tGin.GinPost(`/api/GitPullBranchOrigin`, controller.GitPullBranchOrigin)     //拉取最新分支
	tGin.GinPost(`/api/GitRemoteBranchList`, controller.GitRemoteBranchList)     //查询远程分支列表
	tGin.GinPost(`/api/GitQuickCreateBranch`, controller.GitQuickCreateBranch)   //快捷创建分支
	tGin.GinPost(`/api/GitQueryStatus`, controller.QueryStatus)                  //查询分支本地状态
	tGin.GinPost(`/api/GitCommitLog`, controller.GitCommitLog)                   //查询提交日志
	tGin.GinPost(`/api/GitConfigList`, controller.GitConfigList)                 //git配置
	tGin.GinPost(`/api/GitGroupBranchList`, controller.GitGroupBranchList)       //查询某个git组下所有项目分支
	tGin.GinPost(`/api/CreateMerge`, controller.CreateMerge)                     //创建合并请求
	tGin.GinPost(`/api/GitSetSafeLog`, controller.GitSetSafeLog)                 //设置项目安全
	tGin.GinPost(`/api/GitSaveCredentials`, controller.GitSaveCredentials)       //保存git记住密码账号
}

// gitlab token相关
func gitLabTokenRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/Set/GitLabTokenCreate`, controller.SetGitlabTokenAdd)    //创建
	tGin.GinPost(`/api/Set/GitLabTokenDelete`, controller.SetGitlabTokenDelete) //删除
	tGin.GinPost(`/api/Set/GitLabTokenList`, controller.SetGitlabTokenList)     //列表
}

func globalSetRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/Set/GlobalCreate`, controller.SetGlobalAdd)    //创建
	tGin.GinPost(`/api/Set/GlobalDelete`, controller.SetGlobalDelete) //删除
	tGin.GinPost(`/api/Set/GlobalList`, controller.SetGlobalList)     //列表
}

// 代码生成相关
func codeRouter(tGin *p_gin.Gin) {
	//tGin.GinAll(`/api/CodeGenerate`, controller.GenerateCode) //生成代码
}

func setGroupRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/Set/GroupList`, controller.GroupList)
	tGin.GinPost(`/api/Set/GroupAdd`, controller.GroupAdd)
	tGin.GinPost(`/api/Set/GroupDelete`, controller.GroupDelete)
}

// 设置相关
func setRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/Set/SshList`, controller.SetSshList)
	tGin.GinPost(`/api/Set/SshAdd`, controller.SetSshAdd)
	tGin.GinPost(`/api/Set/SshDelete`, controller.SetSshDelete)
	tGin.GinPost(`/api/Set/GitList`, controller.SetGitList)
	tGin.GinPost(`/api/Set/GitAdd`, controller.SetGitAdd)
	tGin.GinPost(`/api/Set/GitDelete`, controller.SetGitDelete)
	tGin.GinPost(`/api/Set/GitGroupList`, controller.SetGitGroupList)
	tGin.GinPost(`/api/Set/GitGroupAdd`, controller.SetGitGroupAdd)
	tGin.GinPost(`/api/Set/GitGroupDelete`, controller.SetGitGroupDelete)
	tGin.GinPost(`/api/Set/GitQuickList`, controller.SetGitQuickList)
	tGin.GinPost(`/api/Set/SupervisorList`, controller.SetSupervisorctlList)
	tGin.GinPost(`/api/Set/SupervisorAdd`, controller.SetSupervisorAdd)
	tGin.GinPost(`/api/Set/SupervisorDelete`, controller.SetSupervisorDelete)
	tGin.GinPost(`/api/Set/RedisList`, controller.SetRedisList)
	tGin.GinPost(`/api/Set/RedisAdd`, controller.SetRedisAdd)
	tGin.GinPost(`/api/Set/RedisDelete`, controller.SetRedisDelete)
	tGin.GinPost(`/api/Set/MysqlList`, controller.SetMysqlList)
	tGin.GinPost(`/api/Set/MysqlAdd`, controller.SetMysqlAdd)
	tGin.GinPost(`/api/Set/MysqlDelete`, controller.SetMysqlDelete)
	tGin.GinPost(`/api/Set/VariableGroupList`, controller.SetVariableGroupList)
	tGin.GinPost(`/api/Set/VariableGroupAdd`, controller.SetVariableGroupAdd)
	tGin.GinPost(`/api/Set/VariableGroupDelete`, controller.SetVariableGroupDelete)
	tGin.GinPost(`/api/Set/CmdGroupList`, controller.SetCmdGroupList)
	tGin.GinPost(`/api/Set/CmdGroupAdd`, controller.SetCmdGroupAdd)
	tGin.GinPost(`/api/Set/CmdGroupDelete`, controller.SetCmdGroupDelete)
	tGin.GinPost(`/api/Set/SmartLinkGroupList`, controller.SetSmartLinkGroupList)
	tGin.GinPost(`/api/Set/SmartLinkGroupAdd`, controller.SetSmartLinkGroupAdd)
	tGin.GinPost(`/api/Set/SmartLinkGroupDelete`, controller.SetSmartLinkGroupDelete)
	tGin.GinPost(`/api/Set/DockerComposeList`, controller.SetDockerComposeList)
	tGin.GinPost(`/api/Set/DockerComposeAdd`, controller.SetDockerComposeAdd)
	tGin.GinPost(`/api/Set/DockerComposeDelete`, controller.SetDockerComposeDelete)
	tGin.GinPost(`/api/Set/AccountList`, controller.SetAccountList)
	tGin.GinPost(`/api/Set/AccountAdd`, controller.SetAccountAdd)
	tGin.GinPost(`/api/Set/AccountDelete`, controller.SetAccountDelete)
	tGin.GinPost(`/api/Set/AccountGroupList`, controller.SetAccountGroupList)
	tGin.GinPost(`/api/Set/AccountGroupAdd`, controller.SetAccountGroupAdd)
	tGin.GinPost(`/api/Set/AccountGroupDelete`, controller.SetAccountGroupDelete)
	tGin.GinPost(`/api/Set/AiProviderList`, controller.SetAiProviderList)
	tGin.GinPost(`/api/Set/AiProviderAdd`, controller.SetAiProviderAdd)
	tGin.GinPost(`/api/Set/AiProviderDelete`, controller.SetAiProviderDelete)
	tGin.GinPost(`/api/Set/AiModelList`, controller.SetAiModelList)
	tGin.GinPost(`/api/Set/AiModelAdd`, controller.SetAiModelAdd)
	tGin.GinPost(`/api/Set/AiModelDelete`, controller.SetAiModelDelete)
	tGin.GinPost(`/api/Set/AiModelTest`, controller.SetAiModelTest)
	tGin.GinPost(`/api/Set/AiRequestLogList`, controller.SetAiRequestLogList)
	tGin.GinPost(`/api/Set/MemoryConfigGet`, controller.SetMemoryConfigGet)
	tGin.GinPost(`/api/Set/MemoryConfigSave`, controller.SetMemoryConfigSave)
	tGin.GinPost(`/api/Set/RuntimeConfigSave`, controller.SetRuntimeConfigSave)
	tGin.GinPost(`/api/Set/RuntimeDatabaseGitSync`, controller.SetRuntimeDatabaseGitSync)
	tGin.GinPost(`/api/Set/RuntimeConfigItemSave`, controller.SetRuntimeConfigItemSave)
	tGin.GinPost(`/api/Set/CronConfigGet`, controller.SetCronConfigGet)
	tGin.GinPost(`/api/Set/CronConfigSave`, controller.SetCronConfigSave)
}

func setStar(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/StarList`, controller.StarList)
	tGin.GinPost(`/api/StarAdd`, controller.StarAdd)
	tGin.GinPost(`/api/StarDel`, controller.StarDel)
}

func setMarkdown(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/MarkdownHistoryList`, controller.MarkdownHistoryList)
	tGin.GinPost(`/api/MarkdownList`, controller.MarkdownList)
	tGin.GinPost(`/api/MarkdownAdd`, controller.MarkdownAdd)
	tGin.GinPost(`/api/MarkdownDel`, controller.MarkdownDel)
	tGin.GinPost(`/api/MarkdownHistoryDel`, controller.MarkdownHistoryDel)
	tGin.GinPost(`/api/MarkdownSort`, controller.MarkdownSort)
}

func setMemoryFragment(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/GitPendingStatus`, controller.GitPendingStatus)
	tGin.GinPost(`/api/MemoryFragmentStatus`, controller.MemoryFragmentStatus)
	tGin.GinPost(`/api/MemoryFragmentList`, controller.MemoryFragmentList)
	tGin.GinPost(`/api/MemoryFragmentInfo`, controller.MemoryFragmentInfo)
	tGin.GinPost(`/api/MemoryFragmentSave`, controller.MemoryFragmentSave)
	tGin.GinPost(`/api/MemoryFragmentDelete`, controller.MemoryFragmentDelete)
	tGin.GinPost(`/api/MemoryFragmentTrashList`, controller.MemoryFragmentTrashList)
	tGin.GinPost(`/api/MemoryFragmentRestore`, controller.MemoryFragmentRestore)
	tGin.GinPost(`/api/MemoryFragmentHardDelete`, controller.MemoryFragmentHardDelete)
	tGin.GinPost(`/api/MemoryFragmentHistoryList`, controller.MemoryFragmentHistoryList)
	tGin.GinPost(`/api/MemoryFragmentTagList`, controller.MemoryFragmentTagList)
	tGin.GinPost(`/api/MemoryFragmentSearch`, controller.MemoryFragmentSearch)
	tGin.GinPost(`/api/MemoryFragmentOrganize`, controller.MemoryFragmentOrganize)
	tGin.GinPost(`/api/AsyncTaskList`, controller.AsyncTaskList)
	tGin.GinPost(`/api/AsyncTaskInfo`, controller.AsyncTaskInfo)
	tGin.GinPost(`/api/AsyncTaskAction`, controller.AsyncTaskAction)
	tGin.GinPost(`/api/AsyncTaskDelete`, controller.AsyncTaskDelete)
}

func homeTask(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/HomeTaskList`, controller.HomeTaskList)
	tGin.GinPost(`/api/HomeTaskSave`, controller.HomeTaskSave)
	tGin.GinPost(`/api/HomeTaskArchiveToggle`, controller.HomeTaskArchiveToggle)
	tGin.GinPost(`/api/HomeTaskStatusQuickUpdate`, controller.HomeTaskStatusQuickUpdate)
	tGin.GinPost(`/api/HomeTaskDelete`, controller.HomeTaskDelete)
	tGin.GinPost(`/api/HomeTaskDailyReportGenerate`, controller.HomeTaskDailyReportGenerate)
}

func shellOut(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/shellOut`, controller.ShellOut)
	tGin.GinPost(`/api/shellOutSetSeeId`, controller.ShellOutSetSeeId)
	tGin.GinPost(`/api/shellOutCleanErrors`, controller.ShellOutCleanErrors)
	tGin.GinPost(`/api/shellOuts`, controller.GetShellOuts)
	tGin.GinPost(`/api/ShellOutRuleSetList`, controller.ShellOutRuleSetList)
	tGin.GinPost(`/api/ShellOutRuleSetInfo`, controller.ShellOutRuleSetInfo)
	tGin.GinPost(`/api/ShellOutRuleSetSave`, controller.ShellOutRuleSetSave)
	tGin.GinPost(`/api/ShellOutRuleSetDelete`, controller.ShellOutRuleSetDelete)
	tGin.GinPost(`/api/ShellOutRuleImportLegacy`, controller.ShellOutRuleImportLegacy)
	tGin.GinPost(`/api/shellOutDelete`, controller.ShellOutDelete)
	tGin.GinPost(`/api/shellOutStop`, controller.ShellOutStop)
	tGin.GinPost(`/api/shellOutEdit`, controller.ShellOutEdit)
	tGin.GinPost(`/api/shellOutErrorContext`, controller.ShellOutErrorContext)
	tGin.GinPost(`/api/shellOutSearchContent`, controller.ShellOutSearchContent)
	tGin.GinPost(`/api/shellOutCleanLog`, controller.ShellOutCleanLog)
	tGin.GinPost(`/api/shellOutReconnect`, controller.ShellOutReconnect)
}

func variableRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/VariableList`, controller.VariableList)
	tGin.GinPost(`/api/VariableAdd`, controller.VariableAdd)
	tGin.GinPost(`/api/VariableDel`, controller.VariableDelete)
	tGin.GinPost(`/api/VariableInfo`, controller.VariableInfo)
	tGin.GinPost(`/api/VariableCmdAdd`, controller.VariableCmdAdd)
	tGin.GinPost(`/api/VariableCmdDel`, controller.VariableCmdDelete)
	tGin.GinPost(`/api/VariableRun`, controller.VariableCmdRun)        //执行
	tGin.GinPost(`/api/VariableSet`, controller.VariableCmdSet)        //设置项
	tGin.GinPost(`/api/VariableSetLogin`, controller.VariableSetLogin) //设置登录的账号密码
}

func smartLink(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/SmartLinkList`, controller.SmartLinkList)
	tGin.GinPost(`/api/SmartLinkAdd`, controller.SmartLinkAdd)
	tGin.GinPost(`/api/SmartLinkDel`, controller.SmartLinkDelete)
	tGin.GinPost(`/api/SmartLinkInfo`, controller.SmartLinkInfo)
	tGin.GinPost(`/api/SmartLinkRun`, controller.SmartLinkRunPlaywright)
	tGin.GinPost(`/api/SmartLinkRunList`, controller.SmartLinkRunPlaywrightList)
	//tGin.GinPost(`/api/SmartLinkForward`, controller.SmartLinkPlaywrightForward)
	tGin.GinPost(`/api/SmartLinkChromeVersion`, controller.SmartLinkPlaywrightVersion)
	tGin.GinPost(`/api/SmartLinkChromeDownload`, controller.SmartLinkUpWebkit)
	tGin.GinPost(`/api/SmartLinkRecycle`, controller.SmartLinkRecycle)
	tGin.GinPost(`/api/SmartLinkDownloadPath`, controller.SmartLinkDownloadPath)
	tGin.GinPost(`/api/SmartLinkLocatorAutoExtract`, controller.SmartLinkLocatorAutoExtract)
	// 本地客户端相关接口
	tGin.GinGet(`/api/smart-link/runtime-config`, controller.SmartLinkRuntimeConfig)
	tGin.GinGet(`/api/smart-link/client-status`, controller.SmartLinkClientStatus)
	tGin.GinPost(`/api/smart-link/client-build/start`, controller.SmartLinkClientBuildStart)
	tGin.GinGet(`/api/smart-link/client-build/status`, controller.SmartLinkClientBuildStatus)
	tGin.GinGet(`/api/smart-link/client-build/download/:job_id`, controller.SmartLinkClientBuildDownload)
	tGin.GinPost(`/api/smart-link/task/create`, controller.SmartLinkTaskCreate)
	tGin.GinPost(`/api/smart-link/scrape-to-markdown`, controller.SmartLinkScrapeToMarkdown)
	tGin.GinPost(`/api/smart-link/task/result-file`, controller.SmartLinkTaskResultFileUpload)
	tGin.GinGet(`/api/agent/ws`, controller.AgentWs)
	//执行逻辑
	tGin.GinPost(`/api/SmartProcessList`, controller.SmartProcessList)
	tGin.GinPost(`/api/SmartProcessAdd`, controller.SmartProcessAdd)
	tGin.GinPost(`/api/SmartProcessDelete`, controller.SmartProcessDelete)
	tGin.GinPost(`/api/SmartProcessItemList`, controller.SmartProcessItemList)
	tGin.GinPost(`/api/SmartProcessItemAdd`, controller.SmartProcessItemAdd)
	tGin.GinPost(`/api/SmartProcessItemDelete`, controller.SmartProcessItemDelete)
	tGin.GinPost(`/api/SmartProcessItemSort`, controller.SmartProcessItemSort)
	tGin.GinPost(`/api/SmartProcessSetPosition`, controller.SmartProcessSetPosition)
	tGin.GinPost(`/api/SmartProcessSetRelation`, controller.SmartProcessSetRelation)
	tGin.GinPost(`/api/SmartProcessCancelRelation`, controller.SmartProcessCancelRelation)
}

func docker(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/DockerComposeList`, controller.DockerComposeList)
	tGin.GinPost(`/api/DockerComposeRestart`, controller.DockerComposeRestart)
	tGin.GinPost(`/api/DockerComposeStatus`, controller.DockerComposeStatus)
	tGin.GinPost(`/api/DockerComposeServices`, controller.DockerComposeServices)
	tGin.GinPost(`/api/DockerComposeStop`, controller.DockerComposeStop)
	tGin.GinPost(`/api/DockerComposeConfigShow`, controller.DockerComposeConfigShow)
	tGin.GinPost(`/api/DockerComposeStart`, controller.DockerComposeStart)
	tGin.GinPost(`/api/DockerImageList`, controller.DockerImageList)
	tGin.GinPost(`/api/DockerImageContainers`, controller.DockerImageContainers)
	tGin.GinPost(`/api/DockerImageRemove`, controller.DockerImageRemove)
	tGin.GinPost(`/api/DockerContainerStop`, controller.DockerContainerStop)
	tGin.GinPost(`/api/DockerContainerRemove`, controller.DockerContainerRemove)
	tGin.GinPost(`/api/DockerContainerLogTruncate`, controller.DockerContainerLogTruncate)
}

func api(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/CreateCollection`, controller.ApiCreateCollection)
	tGin.GinPost(`/api/DeleteCollection`, controller.ApiDeleteCollection)
	tGin.GinPost(`/api/Collections`, controller.ApiCollections)
	tGin.GinPost(`/api/CollectionListBasic`, controller.ApiCollectionListBasic)
	tGin.GinPost(`/api/CollectionFoldersBasic`, controller.ApiCollectionFoldersBasic)
	tGin.GinPost(`/api/CollectionEnvs`, controller.ApiCollectionEnvs)
	tGin.GinPost(`/api/CreateCollectionEnv`, controller.ApiCreateCollectionEnv)
	tGin.GinPost(`/api/CollectionEnvItems`, controller.ApiCollectionEnvItems)
	tGin.GinPost(`/api/CreateCollectionEnvItem`, controller.ApiCreateCollectionEnvItem)
	tGin.GinPost(`/api/CreateDir`, controller.ApiCreateDir)
	tGin.GinPost(`/api/CreateApi`, controller.ApiCreateApi)
	tGin.GinPost(`/api/DeleteApi`, controller.ApiDeleteApi)
	tGin.GinPost(`/api/DeleteDir`, controller.ApiDeleteDir)
	tGin.GinPost(`/api/Apis`, controller.Apis)
	tGin.GinPost(`/api/FolderApisBasic`, controller.ApiFolderApisBasic)
	tGin.GinPost(`/api/ApisDetailByIds`, controller.ApiApisDetailByIds)
	tGin.GinPost(`/api/ApiRun`, controller.ApiRun)
	tGin.GinPost(`/api/ApiCode`, controller.ApiCode)
	tGin.GinPost(`/api/ApiWeightDown`, controller.ApiWeightDown)
	tGin.GinPost(`/api/ApiTakeJsonResult`, controller.ApiTakeJsonResult)
	tGin.GinPost(`/api/ApiBatchImport`, controller.ApiBatchImport)
	tGin.GinPost(`/api/FolderDetail`, controller.ApiFolderDetail)
	tGin.GinPost(`/api/ApiMove`, controller.ApiMoveApi)
}

func apiUse(tGin *p_gin.Gin) {
	//api git logs
	tGin.SseRoute(`/api/GitLab`, func(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
		clientId := define.SseGitLab
		sse := gsgin.SseRegister(clientId, stopC, c)
		go func() {
			controller.GitLogs(gsgin.GinGetParams(c), func(s string) {
				if strings.Contains(s, `commit 共`) {
					return
				}
				err := sse.SendToChan(s + "\n\n")
				if err != nil {
					gstool.FmtPrintlnLogTime(`错误 %s`, err.Error())
					return
				}
			})
			close(stopC)
		}()
		return sse, nil
	}, func(sse *gsgin.Sse) {
		err := sse.SendToChan(gstool.JsonEncode(p_define.SseData{
			SseDistributeId: "",
			Data:            "[DONE]",
			Type:            p_define.SseContentTypeMsg,
		}))
		if err != nil {
			gstool.FmtPrintlnLogTime(`错误 %s`, err.Error())
			return
		}
		sse.UnRegister()
	})
	openFunc := func(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
		clientId := urlValues.Get(`client_id`)
		sseC := gsgin.SseGetByClientId(clientId)
		if sseC != nil {
			return nil, errors.New(`已存在链接`)
		}
		sse := gsgin.SseRegister(clientId, stopC, c)
		//发送一个事件 前端才会建立连接
		_ = sse.SendToChan(define.SseConnect)
		// 中文注释：Shell 连接状态复用普通 SSE 通道推送，无需单独订阅 shell_connections client_id。
		// English comment: Shell connection status now rides on the normal SSE channel for this client.
		controller.BindShellConnectionsSSE(sse, stopC, 5*time.Second)
		// 中文注释：异步任务状态复用普通 SSE 通道推送，页面加载时初始化一次，后续后端主动推送。
		// English comment: Async task status now rides on the normal SSE channel for this client.
		controller.BindAsyncTasksSSE(sse, stopC, 5*time.Second)
		// 中文注释：记忆库状态复用普通 SSE 通道推送，替代原来的轮询方式。
		// English comment: Memory fragment status now rides on the normal SSE channel for this client.
		controller.BindMemoryFragmentStatusSSE(sse, stopC, 10*time.Second)
		// 中文注释：本地客户端状态复用普通 SSE 通道推送，替代前端 5s 轮询。
		// English comment: Smart-link client status now rides on the normal SSE channel for this client.
		controller.BindSmartLinkClientStatusSSE(sse, stopC, 5*time.Second)
		// 中文注释：Git 待提交状态及倒计时复用普通 SSE 通道推送，替代前端 10s 轮询。
		// English comment: Git pending status and countdown now ride on the normal SSE channel for this client.
		controller.BindGitPendingStatusSSE(sse, stopC, 5*time.Second)
		return sse, nil
	}
	closeFunc := func(sse *gsgin.Sse) {
		sse.UnRegister()
	}
	tGin.SseRoute(`/sse`, openFunc, closeFunc)
}
