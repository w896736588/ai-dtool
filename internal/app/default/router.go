package _default

import (
	"dev_tool/base"
	"dev_tool/internal/app/default/controller"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"net/url"
)

func InitRouter() {
	baseRouter()
	redisRouter()
	phpRouter()
	supervisorRouter()
	gitRouter()
	gitLabTokenRouter()
	globalSetRouter()
	loginRouter()
	codeRouter()
	initSocket()
	setRouter()
	setStar()
	setMarkdown()
	variable()
	smartLink()
	docker()
	ai()
	api()
}

// 基础接口
func baseRouter() {
	base.Component.TGin.GinPost(`/api/BaseLogin`, controller.BaseLogin)                       //登录
	base.Component.TGin.GinPost(`/api/BaseRegisterService`, controller.BaseRegisterService)   //注册各类服务 CheckUnikeyExist
	base.Component.TGin.GinPost(`/api/BaseCheckUnikeyExist`, controller.BaseCheckUnikeyExist) //检查unikey是否已经登录注册
	base.Component.TGin.GinPost(`/api/BaseSshList`, controller.BaseSshList)                   //ssh列表
	base.Component.TGin.GinPost(`/api/Ip`, controller.Ip)                                     //登录
}

// redis相关
func redisRouter() {
	base.Component.TGin.GinPost(`/api/RedisAvailableList`, controller.RedisAvailableList) //可用的redis列表
	base.Component.TGin.GinPost(`/api/RedisSearch`, controller.RedisSearch)               //查询某个key
	base.Component.TGin.GinPost(`/api/RedisKeys`, controller.RedisKeys)                   //模糊搜索key
	base.Component.TGin.GinPost(`/api/RedisKeysType`, controller.RedisKeysType)           //批量获取key缓存类型
	base.Component.TGin.GinPost(`/api/RedisKeyType`, controller.RedisKeyType)             //获取key类型
	base.Component.TGin.GinPost(`/api/RedisSaveString`, controller.RedisSaveString)       //保存string
	base.Component.TGin.GinPost(`/api/RedisDelKey`, controller.RedisDelKey)               //删除key
	base.Component.TGin.GinPost(`/api/RedisDelSub`, controller.RedisDelSub)               //删除sub key
	base.Component.TGin.GinPost(`/api/RedisEditTtl`, controller.RedisEditTtl)             //更改ttl
	base.Component.TGin.GinPost(`/api/RedisDeleteAll`, controller.RedisDelAllKey)         //删除所有缓存
	base.Component.TGin.GinPost(`/api/RedisCreateCache`, controller.RedisCreateCache)     //创建缓存
	base.Component.TGin.GinPost(`/api/RedisEditSub`, controller.RedisEditSub)             //编辑二级缓存
}

// php相关
func phpRouter() {
	base.Component.TGin.GinPost(`/api/PhpUnserialize`, controller.PhpPhpUnSerialize)   //PHP反序列化
	base.Component.TGin.GinPost(`/api/PhpUnserialize2`, controller.PhpPhpUnSerialize2) //PHP反序列化
}

// 消费者相关
func supervisorRouter() {
	base.Component.TGin.GinPost(`/api/SupervisorRestartAll`, controller.SupervisorRestartAll) //重启所有消费者
	base.Component.TGin.GinPost(`/api/SupervisorStopAll`, controller.SupervisorStopAll)       //重启所有消费者
	base.Component.TGin.GinPost(`/api/SupervisorStatusList`, controller.SupervisorStatusList) //查看消费者状态
	base.Component.TGin.GinPost(`/api/SupervisorConfigShow`, controller.SupervisorConfigShow) //查看消费者配置
	base.Component.TGin.GinPost(`/api/SupervisorRestart`, controller.SupervisorRestart)       //重启单个消费者
	base.Component.TGin.GinPost(`/api/SupervisorStop`, controller.SupervisorStop)             //重启单个消费者
	base.Component.TGin.GinPost(`/api/SupervisorConfList`, controller.SupervisorConfList)     //查看所有的配置
	base.Component.TGin.GinPost(`/api/SupervisorConfigList`, controller.SupervisorConfigList) //配置的supervisor
}

// git相关
func gitRouter() {
	base.Component.TGin.GinPost(`/api/GitQueryCurrentBranch`, controller.GitCurrentBranch)  //查询当前分支
	base.Component.TGin.GinPost(`/api/GitChangeBranch`, controller.GitChangeBranch)         //切换分支
	base.Component.TGin.GinPost(`/api/GitPullBranchOrigin`, controller.GitPullBranchOrigin) //拉取最新分支
	base.Component.TGin.GinPost(`/api/GitQueryStatus`, controller.QueryStatus)              //查询分支本地状态
	base.Component.TGin.GinPost(`/api/GitCommitLog`, controller.GitCommitLog)               //查询提交日志
	base.Component.TGin.GinPost(`/api/GitConfigList`, controller.GitConfigList)             //git配置
	base.Component.TGin.GinPost(`/api/CreateMerge`, controller.CreateMerge)                 //创建合并请求
}

// gitlab token相关
func gitLabTokenRouter() {
	base.Component.TGin.GinPost(`/api/Set/GitLabTokenCreate`, controller.SetGitlabTokenAdd)    //创建
	base.Component.TGin.GinPost(`/api/Set/GitLabTokenDelete`, controller.SetGitlabTokenDelete) //删除
	base.Component.TGin.GinPost(`/api/Set/GitLabTokenList`, controller.SetGitlabTokenList)     //列表
}

func globalSetRouter() {
	base.Component.TGin.GinPost(`/api/Set/GlobalCreate`, controller.SetGlobalAdd)    //创建
	base.Component.TGin.GinPost(`/api/Set/GlobalDelete`, controller.SetGlobalDelete) //删除
	base.Component.TGin.GinPost(`/api/Set/GlobalList`, controller.SetGlobalList)     //列表
}

// login相关
func loginRouter() {
	base.Component.TGin.GinPost(`/api/LoginLink`, controller.LoginLink) //拿到登录链接
}

// 代码生成相关
func codeRouter() {
	//base.Component.TGin.GinAll(`/api/CodeGenerate`, controller.GenerateCode) //生成代码
}

// 设置相关
func setRouter() {
	base.Component.TGin.GinPost(`/api/Set/SshList`, controller.SetSshList)
	base.Component.TGin.GinPost(`/api/Set/SshAdd`, controller.SetSshAdd)
	base.Component.TGin.GinPost(`/api/Set/SshDelete`, controller.SetSshDelete)
	base.Component.TGin.GinPost(`/api/Set/GitList`, controller.SetGitList)
	base.Component.TGin.GinPost(`/api/Set/GitAdd`, controller.SetGitAdd)
	base.Component.TGin.GinPost(`/api/Set/GitDelete`, controller.SetGitDelete)
	base.Component.TGin.GinPost(`/api/Set/GitGroupList`, controller.SetGitGroupList)
	base.Component.TGin.GinPost(`/api/Set/GitGroupAdd`, controller.SetGitGroupAdd)
	base.Component.TGin.GinPost(`/api/Set/GitGroupDelete`, controller.SetGitGroupDelete)
	base.Component.TGin.GinPost(`/api/Set/GitQuickList`, controller.SetGitQuickList)
	base.Component.TGin.GinPost(`/api/Set/SupervisorList`, controller.SetSupervisorctlList)
	base.Component.TGin.GinPost(`/api/Set/SupervisorAdd`, controller.SetSupervisorAdd)
	base.Component.TGin.GinPost(`/api/Set/SupervisorDelete`, controller.SetSupervisorDelete)
	base.Component.TGin.GinPost(`/api/Set/RedisList`, controller.SetRedisList)
	base.Component.TGin.GinPost(`/api/Set/RedisAdd`, controller.SetRedisAdd)
	base.Component.TGin.GinPost(`/api/Set/RedisDelete`, controller.SetRedisDelete)
	base.Component.TGin.GinPost(`/api/Set/MysqlList`, controller.SetMysqlList)
	base.Component.TGin.GinPost(`/api/Set/MysqlAdd`, controller.SetMysqlAdd)
	base.Component.TGin.GinPost(`/api/Set/MysqlDelete`, controller.SetMysqlDelete)
	base.Component.TGin.GinPost(`/api/Set/VariableGroupList`, controller.SetVariableGroupList)
	base.Component.TGin.GinPost(`/api/Set/VariableGroupAdd`, controller.SetVariableGroupAdd)
	base.Component.TGin.GinPost(`/api/Set/VariableGroupDelete`, controller.SetVariableGroupDelete)
	base.Component.TGin.GinPost(`/api/Set/CmdGroupList`, controller.SetCmdGroupList)
	base.Component.TGin.GinPost(`/api/Set/CmdGroupAdd`, controller.SetCmdGroupAdd)
	base.Component.TGin.GinPost(`/api/Set/CmdGroupDelete`, controller.SetCmdGroupDelete)
	base.Component.TGin.GinPost(`/api/Set/SmartLinkGroupList`, controller.SetSmartLinkGroupList)
	base.Component.TGin.GinPost(`/api/Set/SmartLinkGroupAdd`, controller.SetSmartLinkGroupAdd)
	base.Component.TGin.GinPost(`/api/Set/SmartLinkGroupDelete`, controller.SetSmartLinkGroupDelete)
	base.Component.TGin.GinPost(`/api/Set/DockerComposeList`, controller.SetDockerComposeList)
	base.Component.TGin.GinPost(`/api/Set/DockerComposeAdd`, controller.SetDockerComposeAdd)
	base.Component.TGin.GinPost(`/api/Set/DockerComposeDelete`, controller.SetDockerComposeDelete)
	base.Component.TGin.GinPost(`/api/Set/AccountList`, controller.SetAccountList)
	base.Component.TGin.GinPost(`/api/Set/AccountAdd`, controller.SetAccountAdd)
	base.Component.TGin.GinPost(`/api/Set/AccountDelete`, controller.SetAccountDelete)
	base.Component.TGin.GinPost(`/api/Set/AccountGroupList`, controller.SetAccountGroupList)
	base.Component.TGin.GinPost(`/api/Set/AccountGroupAdd`, controller.SetAccountGroupAdd)
	base.Component.TGin.GinPost(`/api/Set/AccountGroupDelete`, controller.SetAccountGroupDelete)
}

func setStar() {
	base.Component.TGin.GinPost(`/api/StarList`, controller.StarList)
	base.Component.TGin.GinPost(`/api/StarAdd`, controller.StarAdd)
	base.Component.TGin.GinPost(`/api/StarDel`, controller.StarDel)
}

func setMarkdown() {
	base.Component.TGin.GinPost(`/api/MarkdownHistoryList`, controller.MarkdownHistoryList)
	base.Component.TGin.GinPost(`/api/MarkdownList`, controller.MarkdownList)
	base.Component.TGin.GinPost(`/api/MarkdownAdd`, controller.MarkdownAdd)
	base.Component.TGin.GinPost(`/api/MarkdownDel`, controller.MarkdownDel)
	base.Component.TGin.GinPost(`/api/MarkdownHistoryDel`, controller.MarkdownHistoryDel)
	base.Component.TGin.GinPost(`/api/MarkdownSort`, controller.MarkdownSort)
}

func variable() {
	base.Component.TGin.GinPost(`/api/VariableList`, controller.VariableList)
	base.Component.TGin.GinPost(`/api/VariableAdd`, controller.VariableAdd)
	base.Component.TGin.GinPost(`/api/VariableDel`, controller.VariableDelete)
	base.Component.TGin.GinPost(`/api/VariableInfo`, controller.VariableInfo)
	base.Component.TGin.GinPost(`/api/VariableCmdAdd`, controller.VariableCmdAdd)
	base.Component.TGin.GinPost(`/api/VariableCmdDel`, controller.VariableCmdDelete)
	base.Component.TGin.GinPost(`/api/VariableRun`, controller.VariableCmdRun)        //执行
	base.Component.TGin.GinPost(`/api/VariableSet`, controller.VariableCmdSet)        //设置项
	base.Component.TGin.GinPost(`/api/VariableSetLogin`, controller.VariableSetLogin) //设置登录的账号密码
}

func smartLink() {
	base.Component.TGin.GinPost(`/api/SmartLinkList`, controller.SmartLinkList)
	base.Component.TGin.GinPost(`/api/SmartLinkAdd`, controller.SmartLinkAdd)
	base.Component.TGin.GinPost(`/api/SmartLinkDel`, controller.SmartLinkDelete)
	base.Component.TGin.GinPost(`/api/SmartLinkInfo`, controller.SmartLinkInfo)
	base.Component.TGin.GinPost(`/api/SmartLinkRun`, controller.SmartLinkRunPlaywright)
	base.Component.TGin.GinPost(`/api/SmartLinkRunList`, controller.SmartLinkRunPlaywrightList)
	//base.Component.TGin.GinPost(`/api/SmartLinkForward`, controller.SmartLinkPlaywrightForward)
	base.Component.TGin.GinPost(`/api/SmartLinkChromeVersion`, controller.SmartLinkPlaywrightVersion)
	base.Component.TGin.GinPost(`/api/SmartLinkChromeDownload`, controller.SmartLinkUpWebkit)
	base.Component.TGin.GinPost(`/api/SmartLinkRecycle`, controller.SmartLinkRecycle)
	base.Component.TGin.GinPost(`/api/SmartLinkDownloadPath`, controller.SmartLinkDownloadPath)
	//执行逻辑
	base.Component.TGin.GinPost(`/api/SmartProcessList`, controller.SmartProcessList)
	base.Component.TGin.GinPost(`/api/SmartProcessAdd`, controller.SmartProcessAdd)
	base.Component.TGin.GinPost(`/api/SmartProcessDelete`, controller.SmartProcessDelete)
	base.Component.TGin.GinPost(`/api/SmartProcessItemList`, controller.SmartProcessItemList)
	base.Component.TGin.GinPost(`/api/SmartProcessItemAdd`, controller.SmartProcessItemAdd)
	base.Component.TGin.GinPost(`/api/SmartProcessItemDelete`, controller.SmartProcessItemDelete)
	base.Component.TGin.GinPost(`/api/SmartProcessItemSort`, controller.SmartProcessItemSort)
}

func docker() {
	base.Component.TGin.GinPost(`/api/DockerComposeList`, controller.DockerComposeList)
	base.Component.TGin.GinPost(`/api/DockerComposeRestart`, controller.DockerComposeRestart)
	base.Component.TGin.GinPost(`/api/DockerComposeStop`, controller.DockerComposeStop)
	base.Component.TGin.GinPost(`/api/DockerComposeConfigShow`, controller.DockerComposeConfigShow)
	base.Component.TGin.GinPost(`/api/DockerComposeStart`, controller.DockerComposeStart)
}

func ai() {
	base.Component.TGin.GinPost(`/api/AiRun`, controller.AiRun)
}

func api() {
	//api git logs
	base.Component.TGin.SseRoute(`/api/GitLab`, func(urlValues url.Values, stopC chan int, c *gin.Context) *gsgin.Sse {
		clientId := base.Component.TBase.GetUnique(`api_gitlab_`)
		sse := base.Component.TSse.Sse.Register(clientId, stopC, c)
		go func() {
			controller.GitLogs(gsgin.GinGetParams(c), func(s string) {
				err := base.Component.TSse.SendMsg(clientId, s+"\n", 0)
				if err != nil {
					gstool.FmtPrintlnLogTime(`错误 %s`, err.Error())
					return
				}
			})
			close(stopC)
		}()
		return sse
	}, func(sse *gsgin.Sse) {
		err := base.Component.TSse.SendMsg(sse.ClientId, "[DONE]", 0)
		if err != nil {
			gstool.FmtPrintlnLogTime(`错误 %s`, err.Error())
			return
		}
		base.Component.TSse.Sse.UnRegister(sse.ClientId)
	})
	//sse 替换 websocket
	base.Component.TGin.SseRoute(`/sse`, func(urlValues url.Values, stopC chan int, c *gin.Context) *gsgin.Sse {
		clientId := urlValues.Get(`client_id`)
		return base.Component.TSse.Sse.Register(clientId, stopC, c)
	}, func(sse *gsgin.Sse) {
		base.Component.TSse.Sse.Pause(sse)
	})
}
