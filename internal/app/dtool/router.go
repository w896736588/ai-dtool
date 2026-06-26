package dtool

import (
	"dev_tool/internal/app/dtool/controller"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/middleware"
	"dev_tool/internal/pkg/p_define"
	"dev_tool/internal/pkg/p_gin"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
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
	mysqlRouter(tGin)
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
	taskStatus(tGin)
	taskWorkflow(tGin)
	workflowTemplate(tGin)
	shellOut(tGin)
	variableRouter(tGin)
	smartLink(tGin)
	docker(tGin)
	screenshotRouter(tGin)
	api(tGin)
	apiUse(tGin)
	mcp(tGin)
	agentCli(tGin)
	webhookConfig(tGin)
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
	tGin.GinPost(`/api/BaseLogin`, controller.BaseLogin)                             //Safe 登录
	tGin.GinPost(`/api/BaseLoginStatus`, controller.BaseLoginStatus)                 //Safe 登录状态检查
	tGin.GinPost(`/api/BaseRegisterService`, controller.BaseRegisterService)         //注册各类服务 CheckUnikeyExist
	tGin.GinPost(`/api/BaseCheckUnikeyExist`, controller.BaseCheckUnikeyExist)       //检查unikey是否已经登录注册
	tGin.GinPost(`/api/BaseSshList`, controller.BaseSshList)                         //ssh列表
	tGin.GinPost(`/api/Ip`, controller.Ip)                                           //外网IP
	tGin.GinPost(`/api/GetLocalIP`, controller.GetLocalIP)                           //局域网IP
	tGin.GinPost(`/api/Upload`, controller.Upload)                                   //上传文件
	tGin.GinPost(`/api/MemoryFragmentShareInfo`, controller.MemoryFragmentShareInfo) //知识片段分享只读详情
	tGin.GinGet(`/share/:id/:token`, controller.MemoryFragmentSharePage)             //知识片段分享纯HTML页面
	tGin.GinGet(`/api/download/:name`, controller.DownloadWebFile)                   //下载 web/download 目录文件
	tGin.GinGet(`/web/download/:name`, controller.DownloadWebFile)                   //兼容 web/download 直链下载
	tGin.GinGet(`/memory/images/:name`, controller.MemoryFragmentImageServe)         //记忆库图片静态服务
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
	tGin.GinPost(`/api/GitUploadFile`, controller.GitUploadFile)                 //上传文件到Git项目
	tGin.GinPost(`/api/GitCurrentBranch`, controller.GitCurrentBranchById)       //通过git_id查询当前分支
	tGin.GinPost(`/api/GitPull`, controller.GitPull)                             //通过git_id拉取当前分支最新代码
	tGin.GinPost(`/api/GitChangeBranchById`, controller.GitChangeBranchById)     //通过git_id切换分支
	tGin.SseRoute(`/api/GitCleanupAndSwitchBranchById`, controller.GitCleanupAndSwitchBranchByIdStream, controller.GitCleanupAndSwitchBranchByIdStreamClose)
}

// MySQL查询相关
func mysqlRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/MysqlTables`, controller.MysqlTables)                 //查询MySQL所有表
	tGin.GinPost(`/api/MysqlTableStructure`, controller.MysqlTableStructure) //查询MySQL表结构
	tGin.GinPost(`/api/MysqlQuery`, controller.MysqlQuery)                   //执行MySQL查询
	tGin.GinPost(`/api/MysqlExec`, controller.MysqlExec)                     //执行MySQL写入
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
	tGin.GinPost(`/api/Set/SshStatus`, controller.SetSshStatus)
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
	tGin.GinPost(`/api/Set/MainDBStorageAnalysis`, controller.SetMainDBStorageAnalysis)
	tGin.GinPost(`/api/Set/MainDBStorageVacuum`, controller.SetMainDBStorageVacuum)
	tGin.GinPost(`/api/Set/RuntimeConfigSave`, controller.SetRuntimeConfigSave)
	tGin.GinPost(`/api/Set/RuntimeConfigItemSave`, controller.SetRuntimeConfigItemSave)
	tGin.GinPost(`/api/Set/CronConfigTypes`, controller.SetCronConfigTypes)
	tGin.GinPost(`/api/Set/CronConfigGet`, controller.SetCronConfigGet)
	tGin.GinPost(`/api/Set/CronConfigSave`, controller.SetCronConfigSave)
	tGin.GinPost(`/api/Set/HomeTaskConfigGet`, controller.SetHomeTaskConfigGet)
	tGin.GinPost(`/api/Set/HomeTaskConfigSave`, controller.SetHomeTaskConfigSave)
	tGin.GinPost(`/api/Set/PromptChangeLogList`, controller.SetPromptChangeLogList)
	tGin.GinPost(`/api/Set/LocalDirList`, controller.SetLocalDirList)
	tGin.GinPost(`/api/Set/LocalDirBatchCheck`, controller.SetLocalDirBatchCheck)
	tGin.GinPost(`/api/Set/LocalBranchBatchCheck`, controller.SetLocalBranchBatchCheck)
	tGin.GinPost(`/api/Set/LocalBranchMismatchDetail`, controller.SetLocalBranchMismatchDetail)
	tGin.GinPost(`/api/Set/RemoteBranchCheck`, controller.SetRemoteBranchCheck)
	tGin.GinPost(`/api/Set/RemoteBranchPush`, controller.SetRemoteBranchPush)
	tGin.GinPost(`/api/Set/OpenLocalDir`, controller.SetOpenLocalDir)
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
	tGin.GinPost(`/api/GitPendingCommitPush`, controller.GitPendingCommitPush)
	tGin.GinPost(`/api/MemoryFragmentStatus`, controller.MemoryFragmentStatus)
	tGin.GinPost(`/api/MemoryFragmentFolderList`, controller.MemoryFragmentFolderList)
	tGin.GinPost(`/api/MemoryFragmentFolderCreate`, controller.MemoryFragmentFolderCreate)
	tGin.GinPost(`/api/MemoryFragmentFolderUpdate`, controller.MemoryFragmentFolderUpdate)
	tGin.GinPost(`/api/MemoryFragmentFolderChange`, controller.MemoryFragmentFolderChange)
	tGin.GinPost(`/api/MemoryFragmentList`, controller.MemoryFragmentList)
	tGin.GinPost(`/api/MemoryFragmentInfo`, controller.MemoryFragmentInfo)
	tGin.GinPost(`/api/MemoryFragmentSave`, controller.MemoryFragmentSave)
	tGin.GinPost(`/api/MemoryFragmentCreate`, controller.MemoryFragmentCreate)
	tGin.GinPost(`/api/MemoryFragmentSaveById`, controller.MemoryFragmentSaveById)
	tGin.GinPost(`/api/MemoryFragmentDelete`, controller.MemoryFragmentDelete)
	tGin.GinPost(`/api/MemoryFragmentTrashList`, controller.MemoryFragmentTrashList)
	tGin.GinPost(`/api/MemoryFragmentRestore`, controller.MemoryFragmentRestore)
	tGin.GinPost(`/api/MemoryFragmentHardDelete`, controller.MemoryFragmentHardDelete)
	tGin.GinPost(`/api/MemoryFragmentHistoryList`, controller.MemoryFragmentHistoryList)
	tGin.GinPost(`/api/MemoryFragmentTagList`, controller.MemoryFragmentTagList)
	tGin.GinPost(`/api/MemoryFragmentSearch`, controller.MemoryFragmentSearch)
	tGin.GinPost(`/api/MemoryFragmentOrganize`, controller.MemoryFragmentOrganize)
	tGin.GinPost(`/api/MemoryFragmentShareCreate`, controller.MemoryFragmentShareCreate)
	tGin.GinPost(`/api/MemoryFragmentImageUpload`, controller.MemoryFragmentImageUpload)
	tGin.GinPost(`/api/MemoryFragmentUploadZip`, controller.MemoryFragmentUploadZip)
	tGin.GinPost(`/api/MemoryFragmentUpdateZip`, controller.MemoryFragmentUpdateZip)
	tGin.GinGet(`/api/MemoryFragmentDownloadZip`, controller.MemoryFragmentDownloadZip)
	tGin.GinPost(`/api/MemoryFragmentBatchInfoByPaths`, controller.MemoryFragmentBatchInfoByPaths)
	tGin.GinPost(`/api/MemoryFragmentReferences`, controller.MemoryFragmentReferences)
	tGin.GinPost(`/api/AsyncTaskList`, controller.AsyncTaskList)
	tGin.GinPost(`/api/AsyncTaskInfo`, controller.AsyncTaskInfo)
	tGin.GinPost(`/api/AsyncTaskAction`, controller.AsyncTaskAction)
	tGin.GinPost(`/api/AsyncTaskDelete`, controller.AsyncTaskDelete)
	tGin.GinPost(`/api/AsyncTaskRetry`, controller.AsyncTaskRetry)
}

func homeTask(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/HomeTaskList`, controller.HomeTaskList)
	tGin.GinPost(`/api/HomeTaskCount`, controller.HomeTaskCount)
	tGin.GinPost(`/api/HomeTaskInfo`, controller.HomeTaskInfo)
	tGin.GinPost(`/api/HomeTaskSave`, controller.HomeTaskSave)
	tGin.GinPost(`/api/HomeTaskArchiveToggle`, controller.HomeTaskArchiveToggle)
	tGin.GinPost(`/api/HomeTaskStatusQuickUpdate`, controller.HomeTaskStatusQuickUpdate)
	tGin.GinPost(`/api/HomeTaskDelete`, controller.HomeTaskDelete)
	tGin.GinPost(`/api/HomeTaskDailyReportGenerate`, controller.HomeTaskDailyReportGenerate)
	tGin.GinPost(`/api/HomeTaskLastDevConfigByGitId`, controller.HomeTaskLastDevConfigByGitId)
	tGin.GinPost(`/api/HomeTaskBranchNameGenerate`, controller.HomeTaskBranchNameGenerate)
	tGin.GinPost(`/api/HomeTaskZcodeSessionIdAppend`, controller.HomeTaskZcodeSessionIdAppend)
	tGin.GinPost(`/api/HomeTaskUnusedLocalDirs`, controller.HomeTaskUnusedLocalDirs)
	// SSE 聚合推送：页面附加数据
	tGin.GinPost(`/api/HomeTaskPageDataLoad`, controller.HomeTaskPageDataLoad)
	tGin.GinPost(`/api/HomeTaskPageDataDirCheck`, controller.CheckAndPushLocalDirs)
	tGin.GinPost(`/api/HomeTaskPageDataBranchCheck`, controller.CheckAndPushBranchStatus)
}

// taskStatus 任务状态管理路由
func taskStatus(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/TaskStatusList`, controller.TaskStatusList)
	tGin.GinPost(`/api/TaskStatusSave`, controller.TaskStatusSave)
	tGin.GinPost(`/api/TaskStatusDelete`, controller.TaskStatusDelete)
	tGin.GinPost(`/api/TaskStatusSort`, controller.TaskStatusSort)
}

func taskWorkflow(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/task/workflow/create_or_get`, controller.TaskWorkflowCreateOrGet)
	tGin.GinPost(`/api/task/workflow/info`, controller.TaskWorkflowInfo)
	tGin.GinPost(`/api/task/workflow/requirement/fetch`, controller.TaskWorkflowRequirementFetch)
	tGin.GinPost(`/api/task/workflow/dev-plan/init`, controller.TaskWorkflowDevPlanInit)
	tGin.GinPost(`/api/task/workflow/dev-plan/info`, controller.TaskWorkflowDevPlanInfo)
	tGin.GinPost(`/api/task/workflow/dev-plan/save`, controller.TaskWorkflowDevPlanSave)
	tGin.GinPost(`/api/task/workflow/ui-assist/generate`, controller.TaskWorkflowUIAssistGenerate)
	tGin.GinPost(`/api/task/workflow/ui-assist/info`, controller.TaskWorkflowUIAssistInfo)
	tGin.GinPost(`/api/task/workflow/coverage/generate`, controller.TaskWorkflowCoverageGenerate)
	tGin.GinPost(`/api/task/workflow/coverage/info`, controller.TaskWorkflowCoverageInfo)
	tGin.GinPost(`/api/task/workflow/test-plan/generate`, controller.TaskWorkflowTestPlanGenerate)
	tGin.GinPost(`/api/task/workflow/test-plan/info`, controller.TaskWorkflowTestPlanInfo)
	tGin.GinPost(`/api/task/workflow/test-run/execute`, controller.TaskWorkflowTestRunExecute)
	tGin.GinPost(`/api/task/workflow/test-run/list`, controller.TaskWorkflowTestRunList)
	tGin.GinPost(`/api/task/workflow/prompts/save`, controller.TaskWorkflowPromptsSave)
	tGin.GinPost(`/api/task/workflow/prompts/restore`, controller.TaskWorkflowPromptsRestore)
	tGin.GinPost(`/api/task/workflow/api-doc/reset`, controller.TaskWorkflowApiDocReset)
	tGin.GinPost(`/api/task/workflow/node-status/update`, controller.TaskWorkflowNodeStatusUpdate)
	tGin.GinPost(`/api/task/workflow/batch-node-status`, controller.TaskWorkflowBatchNodeStatus)
	tGin.GinPost(`/api/task/workflow/issue-fix/resolve`, controller.TaskWorkflowIssueFixResolve)
	tGin.GinPost(`/api/task/workflow/chat/send`, controller.TaskWorkflowChatSend)
	tGin.GinPost(`/api/task/workflow/chat/continue`, controller.TaskWorkflowChatContinue)
	tGin.GinPost(`/api/task/workflow/chat/stop`, controller.TaskWorkflowChatStop)
	tGin.GinPost(`/api/task/workflow/chat/list`, controller.TaskWorkflowChatList)
	tGin.GinPost(`/api/task/workflow/chat/detail`, controller.TaskWorkflowChatDetail)
	tGin.GinPost(`/api/task/workflow/chat/dirs`, controller.TaskWorkflowChatDirs)
	tGin.GinPost(`/api/task/workflow/chat/list-by-prompt-type`, controller.TaskWorkflowChatListByPromptType)
	tGin.GinPost(`/api/task/workflow/chat/list-by-agent-cli`, controller.TaskWorkflowChatListByAgentCli)
	tGin.GinPost(`/api/agent/chat/send`, controller.AgentChatSend)
	tGin.GinPost(`/api/agent/chat/list-by-agent-cli`, controller.AgentChatListByAgentCli)
	tGin.GinPost(`/api/agent/chat/mark-read`, controller.AgentChatMarkRead)
	tGin.GinPost(`/api/task/workflow/zcode/save`, controller.TaskWorkflowZcodeSave)
	tGin.GinPost(`/api/task/workflow/zcode/get`, controller.TaskWorkflowZcodeGet)
	tGin.GinPost(`/api/task/workflow/zcode/delete`, controller.TaskWorkflowZcodeDelete)
	tGin.GinPost(`/api/task/workflow/file-changes/summary`, controller.TaskWorkflowFileChangesSummary)
	tGin.GinPost(`/api/task/workflow/file-changes/detail`, controller.TaskWorkflowFileChangesDetail)
	tGin.GinPost(`/api/task/workflow/file-changes/file-diff`, controller.TaskWorkflowFileChangesFileDiff)
	tGin.GinPost(`/api/task/workflow/open-in-editor`, controller.TaskWorkflowOpenInEditor)
}

func workflowTemplate(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/workflow/template/list`, controller.WorkflowTemplateList)
	tGin.GinPost(`/api/workflow/template/save`, controller.WorkflowTemplateSave)
	tGin.GinPost(`/api/workflow/template/delete`, controller.WorkflowTemplateDelete)
	tGin.GinPost(`/api/workflow/template/set-default`, controller.WorkflowTemplateSetDefault)
	tGin.GinPost(`/api/workflow/template/import`, controller.WorkflowTemplateImport)
	tGin.GinPost(`/api/workflow/template/step/save`, controller.WorkflowTemplateStepSave)
	tGin.GinPost(`/api/workflow/template/step/delete`, controller.WorkflowTemplateStepDelete)
	tGin.GinPost(`/api/workflow/template/step/sort`, controller.WorkflowTemplateStepSort)
	// 简化接口：仅返回 id+name，供下拉选择
	tGin.GinPost(`/api/workflow/template/list-basic`, controller.WorkflowTemplateListBasic)
	// 动态读取 skills 目录列表
	tGin.GinPost(`/api/workflow/skill/list`, controller.WorkflowSkillList)
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
	tGin.GinPost(`/api/ai/browser/session/open`, controller.AIBrowserSessionOpen)
	tGin.GinPost(`/api/ai/browser/session/capture-headers`, controller.AIBrowserSessionCaptureHeaders)
	// 新表 smart_link 接口
	tGin.GinPost(`/api/SmartLinkItemList`, controller.SmartLinkItemList)
	tGin.GinPost(`/api/SmartLinkItemAdd`, controller.SmartLinkItemAdd)
	tGin.GinPost(`/api/SmartLinkItemDelete`, controller.SmartLinkItemDelete)
	tGin.GinPost(`/api/SmartLinkItemInfo`, controller.SmartLinkItemInfo)
	tGin.GinPost(`/api/SmartLinkMigrateOldData`, controller.SmartLinkMigrateOldData)
	tGin.GinPost(`/api/SmartLinkRun`, controller.SmartLinkRunPlaywright)
	tGin.GinPost(`/api/SmartLinkRunList`, controller.SmartLinkRunPlaywrightList)
	//tGin.GinPost(`/api/SmartLinkForward`, controller.SmartLinkPlaywrightForward)
	tGin.GinPost(`/api/SmartLinkChromeVersion`, controller.SmartLinkPlaywrightVersion)
	tGin.GinPost(`/api/SmartLinkChromeDownload`, controller.SmartLinkUpWebkit)
	tGin.GinPost(`/api/SmartLinkRecycle`, controller.SmartLinkRecycle)
	tGin.GinPost(`/api/SmartLinkDownloadPath`, controller.SmartLinkDownloadPath)
	tGin.GinPost(`/api/SmartLinkOpenDataDir`, controller.SmartLinkOpenDataDir)
	tGin.GinPost(`/api/SmartLinkLocatorAutoExtract`, controller.SmartLinkLocatorAutoExtract)
	tGin.GinPost(`/api/smart-link/scrape-to-markdown`, controller.SmartLinkScrapeToMarkdown)
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
	tGin.GinPost(`/api/DockerServiceRestart`, controller.DockerServiceRestart)
	tGin.GinPost(`/api/DockerServiceLogs`, controller.DockerServiceLogs)
}

func api(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/CreateCollection`, controller.ApiCreateCollection)
	tGin.GinPost(`/api/DeleteCollection`, controller.ApiDeleteCollection)
	tGin.GinPost(`/api/Collections`, controller.ApiCollections)
	tGin.GinPost(`/api/CollectionListBasic`, controller.ApiCollectionListBasic)
	tGin.GinPost(`/api/CollectionFoldersBasic`, controller.ApiCollectionFoldersBasic)
	tGin.GinPost(`/api/CollectionFoldersBatchBasic`, controller.ApiCollectionFoldersBatchBasic)
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
	tGin.GinPost(`/api/FolderApisBatchBasic`, controller.ApiFolderApisBatchBasic)
	tGin.GinPost(`/api/ApisDetailByIds`, controller.ApiApisDetailByIds)
	tGin.GinPost(`/api/ApiRun`, controller.ApiRun)
	tGin.GinPost(`/api/ApiCode`, controller.ApiCode)
	tGin.GinPost(`/api/ApiWeightDown`, controller.ApiWeightDown)
	tGin.GinPost(`/api/ApiTakeJsonResult`, controller.ApiTakeJsonResult)
	tGin.GinPost(`/api/ApiBatchImport`, controller.ApiBatchImport)
	tGin.GinPost(`/api/FolderDetail`, controller.ApiFolderDetail)
	tGin.GinPost(`/api/ApiMove`, controller.ApiMoveApi)
	tGin.GinPost(`/api/ArchiveFolderList`, controller.ApiArchiveFolderList)
	tGin.GinPost(`/api/CleanupCandidateFolders`, controller.ApiCleanupCandidateFolders)
	tGin.GinPost(`/api/CleanupArchiveFolders`, controller.ApiCleanupArchiveFolders)
	tGin.GinPost(`/api/RestoreFolder`, controller.ApiRestoreFolder)
	tGin.GinPost(`/api/PermanentDeleteDir`, controller.ApiPermanentDeleteDir)
	tGin.GinPost(`/api/FolderApisMarkdown`, controller.ApiFolderApisMarkdown)
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
	// AI 智能搜索 SSE 端点
	tGin.SseRoute(`/api/MemoryFragmentAiSearch`, func(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
		return controller.MemoryFragmentAiSearch(urlValues, stopC, c)
	}, func(sse *gsgin.Sse) {
		_ = sse.SendToChan(gstool.JsonEncode(p_define.SseData{
			Data: "[DONE]",
			Type: p_define.SseContentTypeMsg,
		}))
		sse.UnRegister()
	})
	// SSE 可用端口查询接口（所有 gin 实例均可访问）
	tGin.GinPost(`/api/SseAvailablePort`, controller.SseAvailablePort)
	// SSE 所有活跃连接详情接口
	tGin.GinPost(`/api/SseConnectionDetails`, controller.SseConnectionDetails)
	// 判断当前 gin 实例是否是 SSE 端口，仅 SSE 端口才注册 /sse 路由
	if controller.IsSsePort(tGin.Port) {
		openFunc := controller.BuildSseOpenFunc(tGin.Port)
		closeFunc := controller.BuildSseCloseFunc()
		tGin.SseRoute(`/sse`, openFunc, closeFunc)
		// AgentCli 业务独立 SSE
		tGin.SseRoute(`/sse/agent_cli`, controller.AgentCliChatSseOpen, controller.AgentCliChatSseClose)
		// TaskWorkflow 业务独立 SSE
		tGin.SseRoute(`/sse/task_workflow`, controller.TaskWorkflowChatSseOpen, controller.TaskWorkflowChatSseClose)
		// Fullpage 独立 SSE（专用于 shell_out 日志流，不与通用 /sse 混合）
		tGin.SseRoute(`/sse/fullpage`, controller.FullpageSseOpen, controller.FullpageSseClose)
	}
}

func screenshotRouter(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/Screenshot`, controller.Screenshot)
}

// mcp 路由
func mcp(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/McpTypeList`, controller.McpTypeList)
	tGin.GinPost(`/api/McpBindingList`, controller.McpBindingList)
	tGin.GinPost(`/api/McpBindingAdd`, controller.McpBindingAdd)
	tGin.GinPost(`/api/McpBindingRemove`, controller.McpBindingRemove)
	tGin.GinPost(`/api/McpBindingInstruction`, controller.McpBindingInstruction)
	tGin.GinPost(`/api/McpAgentTargetList`, controller.McpAgentTargetList)
	tGin.GinPost(`/api/McpAgentTargetSave`, controller.McpAgentTargetSave)
	tGin.GinPost(`/api/McpAgentTargetDelete`, controller.McpAgentTargetDelete)
	tGin.GinPost(`/api/McpConfigPreview`, controller.McpConfigPreview)
	tGin.GinPost(`/api/McpChromeDevtoolsConfigList`, controller.McpChromeDevtoolsConfigList)
	tGin.GinPost(`/api/McpChromeDevtoolsConfigSave`, controller.McpChromeDevtoolsConfigSave)
	tGin.GinPost(`/api/McpChromeDevtoolsConfigDelete`, controller.McpChromeDevtoolsConfigDelete)
}

// agentCli 路由
func agentCli(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/AgentCliList`, controller.AgentCliList)
	tGin.GinPost(`/api/AgentCliSave`, controller.AgentCliSave)
	tGin.GinPost(`/api/AgentCliDelete`, controller.AgentCliDelete)
	tGin.GinPost(`/api/AgentCliReadSettings`, controller.AgentCliReadSettings)
	tGin.GinPost(`/api/AgentCliWriteMcpServers`, controller.AgentCliWriteMcpServers)
	tGin.GinPost(`/api/AgentCliWriteDeepSeek`, controller.AgentCliWriteDeepSeek)
	tGin.GinPost(`/api/AgentCliToggleEnabled`, controller.AgentCliToggleEnabled)
	// AgentCli 分组管理
	tGin.GinPost(`/api/AgentCliGroupList`, controller.AgentCliGroupList)
	tGin.GinPost(`/api/AgentCliGroupSave`, controller.AgentCliGroupSave)
	tGin.GinPost(`/api/AgentCliGroupDelete`, controller.AgentCliGroupDelete)
	tGin.GinPost(`/api/AgentCliGroupRelSave`, controller.AgentCliGroupRelSave)
	tGin.GinPost(`/api/AgentCliPromptTemplateList`, controller.AgentCliPromptTemplateList)
	tGin.GinPost(`/api/AgentCliPromptTemplateSave`, controller.AgentCliPromptTemplateSave)
	tGin.GinPost(`/api/AgentCliPromptTemplateDelete`, controller.AgentCliPromptTemplateDelete)
}

// webhookConfig 路由
func webhookConfig(tGin *p_gin.Gin) {
	tGin.GinPost(`/api/WebhookConfigList`, controller.WebhookConfigList)
	tGin.GinPost(`/api/WebhookConfigSave`, controller.WebhookConfigSave)
	tGin.GinPost(`/api/WebhookConfigDelete`, controller.WebhookConfigDelete)
	tGin.GinPost(`/api/WebhookConfigTest`, controller.WebhookConfigTest)
}
