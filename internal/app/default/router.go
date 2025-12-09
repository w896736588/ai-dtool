package _default

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"dev_tool/internal/app/default/controller"
	"errors"
	"mime/multipart"
	"net/http"
	"net/url"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
)

func InitRouter(tGin *base.Gin) {
	baseRouter(tGin)
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
	shellOut(tGin)
	variable(tGin)
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

// 基础接口
func baseRouter(tGin *base.Gin) {
	tGin.GinPost(`/api/BaseLogin`, controller.BaseLogin)                       //登录
	tGin.GinPost(`/api/BaseRegisterService`, controller.BaseRegisterService)   //注册各类服务 CheckUnikeyExist
	tGin.GinPost(`/api/BaseCheckUnikeyExist`, controller.BaseCheckUnikeyExist) //检查unikey是否已经登录注册
	tGin.GinPost(`/api/BaseSshList`, controller.BaseSshList)                   //ssh列表
	tGin.GinPost(`/api/Ip`, controller.Ip)                                     //登录
	tGin.GinPost(`/api/ports`, controller.Ports)                               //获取支持的端口
	tGin.GinPost(`/api/Upload`, controller.Upload)                             //上传文件
}

// redis相关
func redisRouter(tGin *base.Gin) {
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
func phpRouter(tGin *base.Gin) {
	tGin.GinPost(`/api/PhpUnserialize`, controller.PhpPhpUnSerialize)   //PHP反序列化
	tGin.GinPost(`/api/PhpUnserialize2`, controller.PhpPhpUnSerialize2) //PHP反序列化
}

// 消费者相关
func supervisorRouter(tGin *base.Gin) {
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
func gitRouter(tGin *base.Gin) {
	tGin.GinPost(`/api/GitQueryCurrentBranch`, controller.GitCurrentBranch)      //查询当前分支
	tGin.GinPost(`/api/GitChangeBranch`, controller.GitChangeBranch)             //切换分支
	tGin.GinPost(`/api/GitChangeBranchRemote`, controller.GitChangeBranchRemote) //切换远程分支
	tGin.GinPost(`/api/GitPullBranchOrigin`, controller.GitPullBranchOrigin)     //拉取最新分支
	tGin.GinPost(`/api/GitQueryStatus`, controller.QueryStatus)                  //查询分支本地状态
	tGin.GinPost(`/api/GitCommitLog`, controller.GitCommitLog)                   //查询提交日志
	tGin.GinPost(`/api/GitConfigList`, controller.GitConfigList)                 //git配置
	tGin.GinPost(`/api/CreateMerge`, controller.CreateMerge)                     //创建合并请求
}

// gitlab token相关
func gitLabTokenRouter(tGin *base.Gin) {
	tGin.GinPost(`/api/Set/GitLabTokenCreate`, controller.SetGitlabTokenAdd)    //创建
	tGin.GinPost(`/api/Set/GitLabTokenDelete`, controller.SetGitlabTokenDelete) //删除
	tGin.GinPost(`/api/Set/GitLabTokenList`, controller.SetGitlabTokenList)     //列表
}

func globalSetRouter(tGin *base.Gin) {
	tGin.GinPost(`/api/Set/GlobalCreate`, controller.SetGlobalAdd)    //创建
	tGin.GinPost(`/api/Set/GlobalDelete`, controller.SetGlobalDelete) //删除
	tGin.GinPost(`/api/Set/GlobalList`, controller.SetGlobalList)     //列表
}

// 代码生成相关
func codeRouter(tGin *base.Gin) {
	//tGin.GinAll(`/api/CodeGenerate`, controller.GenerateCode) //生成代码
}

func setGroupRouter(tGin *base.Gin) {
	tGin.GinPost(`/api/Set/GroupList`, controller.GroupList)
	tGin.GinPost(`/api/Set/GroupAdd`, controller.GroupAdd)
	tGin.GinPost(`/api/Set/GroupDelete`, controller.GroupDelete)
}

// 设置相关
func setRouter(tGin *base.Gin) {
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
}

func setStar(tGin *base.Gin) {
	tGin.GinPost(`/api/StarList`, controller.StarList)
	tGin.GinPost(`/api/StarAdd`, controller.StarAdd)
	tGin.GinPost(`/api/StarDel`, controller.StarDel)
}

func setMarkdown(tGin *base.Gin) {
	tGin.GinPost(`/api/MarkdownHistoryList`, controller.MarkdownHistoryList)
	tGin.GinPost(`/api/MarkdownList`, controller.MarkdownList)
	tGin.GinPost(`/api/MarkdownAdd`, controller.MarkdownAdd)
	tGin.GinPost(`/api/MarkdownDel`, controller.MarkdownDel)
	tGin.GinPost(`/api/MarkdownHistoryDel`, controller.MarkdownHistoryDel)
	tGin.GinPost(`/api/MarkdownSort`, controller.MarkdownSort)
}

func shellOut(tGin *base.Gin) {
	tGin.GinPost(`/api/shellOut`, controller.ShellOut)
	tGin.GinPost(`/api/shellOutSetSeeId`, controller.ShellOutSetSeeId)
	tGin.GinPost(`/api/shellOutCleanErrors`, controller.ShellOutCleanErrors)
	tGin.GinPost(`/api/shellOuts`, controller.GetShellOuts)
	tGin.GinPost(`/api/shellOutDelete`, controller.ShellOutDelete)
	tGin.GinPost(`/api/shellOutStop`, controller.ShellOutStop)
	tGin.GinPost(`/api/shellOutEdit`, controller.ShellOutEdit)
	tGin.GinPost(`/api/shellOutErrorContext`, controller.ShellOutErrorContext)
	tGin.GinPost(`/api/shellOutSearchContent`, controller.ShellOutSearchContent)
	tGin.GinPost(`/api/shellOutCleanLog`, controller.ShellOutCleanLog)
	tGin.GinPost(`/api/shellOutRestart`, controller.ShellOutRestart)
}

func variable(tGin *base.Gin) {
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

func smartLink(tGin *base.Gin) {
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

func docker(tGin *base.Gin) {
	tGin.GinPost(`/api/DockerComposeList`, controller.DockerComposeList)
	tGin.GinPost(`/api/DockerComposeRestart`, controller.DockerComposeRestart)
	tGin.GinPost(`/api/DockerComposeStatus`, controller.DockerComposeStatus)
	tGin.GinPost(`/api/DockerComposeServices`, controller.DockerComposeServices)
	tGin.GinPost(`/api/DockerComposeStop`, controller.DockerComposeStop)
	tGin.GinPost(`/api/DockerComposeConfigShow`, controller.DockerComposeConfigShow)
	tGin.GinPost(`/api/DockerComposeStart`, controller.DockerComposeStart)
}

func api(tGin *base.Gin) {
	tGin.GinPost(`/api/CreateCollection`, controller.ApiCreateCollection)
	tGin.GinPost(`/api/DeleteCollection`, controller.ApiDeleteCollection)
	tGin.GinPost(`/api/Collections`, controller.ApiCollections)
	tGin.GinPost(`/api/CollectionEnvs`, controller.ApiCollectionEnvs)
	tGin.GinPost(`/api/CreateCollectionEnv`, controller.ApiCreateCollectionEnv)
	tGin.GinPost(`/api/CollectionEnvItems`, controller.ApiCollectionEnvItems)
	tGin.GinPost(`/api/CreateCollectionEnvItem`, controller.ApiCreateCollectionEnvItem)
	tGin.GinPost(`/api/CreateDir`, controller.ApiCreateDir)
	tGin.GinPost(`/api/CreateApi`, controller.ApiCreateApi)
	tGin.GinPost(`/api/DeleteApi`, controller.ApiDeleteApi)
	tGin.GinPost(`/api/DeleteDir`, controller.ApiDeleteDir)
	tGin.GinPost(`/api/Apis`, controller.Apis)
	tGin.GinPost(`/api/ApiRun`, controller.ApiRun)
	tGin.GinPost(`/api/ApiCode`, controller.ApiCode)
	tGin.GinPost(`/api/ApiWeightDown`, controller.ApiWeightDown)
	tGin.GinPost(`/api/ApiResultTake`, controller.ApiResultTake)
}

func apiUse(tGin *base.Gin) {
	//api git logs
	tGin.SseRoute(`/api/GitLab`, func(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
		clientId := define.SseGitLab
		sse := base.Component.TSse.Sse.Register(clientId, stopC, c)
		gstool.FmtPrintlnLogTime(`收到请求 /api/GitLab`)
		go func() {
			controller.GitLogs(gsgin.GinGetParams(c), func(s string) {
				err := base.Component.TSse.SendMsg(clientId, define.SseContentTypeMsg, s+"\n", 0, []string{define.SseGitLab}...)
				if err != nil {
					gstool.FmtPrintlnLogTime(`错误 %s`, err.Error())
					return
				}
			})
			close(stopC)
		}()
		return sse, nil
	}, func(sse *gsgin.Sse) {
		err := base.Component.TSse.SendMsg(sse.ClientId, define.SseContentTypeMsg, "[DONE]", 0, []string{define.SseGitLab}...)
		if err != nil {
			gstool.FmtPrintlnLogTime(`错误 %s`, err.Error())
			return
		}
		base.Component.TSse.Sse.UnRegister(sse.ClientId)
	})
	//sse 替换 websocket
	openFunc := func(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
		clientId := urlValues.Get(`client_id`)
		sseC := base.Component.TSse.Sse.GetSseByClientId(clientId)
		if sseC != nil {
			return nil, errors.New(`已存在链接`)
		}
		sse := base.Component.TSse.Sse.Register(clientId, stopC, c)
		//发送一个事件 前端才会建立连接
		_ = base.Component.TSse.Sse.Send(clientId, define.SseConnect, 1)
		define.RegisterDistributeSseId(clientId)
		return sse, nil
	}
	closeFunc := func(sse *gsgin.Sse) {
		base.Component.TSse.Sse.UnRegister(sse.ClientId)
		define.UnRegisterDistributeSseId(sse.ClientId)
	}
	tGin.SseRoute(`/sse`, openFunc, closeFunc)
}
