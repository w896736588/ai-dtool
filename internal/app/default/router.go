package _default

import (
	"dev_tool/base"
	"dev_tool/internal/app/default/controller"
)

func InitRouter() {
	baseRouter()
	redisRouter()
	phpRouter()
	supervisorRouter()
	gitRouter()
	loginRouter()
	codeRouter()
	initSocket()
	setRouter()
	setStar()
	variable()
	smartLink()
}

// 基础接口
func baseRouter() {
	base.Component.TGin.GinPost(`/api/BaseLogin`, controller.BaseLogin)                       //登录
	base.Component.TGin.GinPost(`/api/BaseRegisterService`, controller.BaseRegisterService)   //注册各类服务 CheckUnikeyExist
	base.Component.TGin.GinPost(`/api/BaseCheckUnikeyExist`, controller.BaseCheckUnikeyExist) //检查unikey是否已经登录注册
	base.Component.TGin.GinPost(`/api/BaseSshList`, controller.BaseSshList)                   //ssh列表
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
	base.Component.TGin.GinPost(`/api/PhpUnserialize`, controller.PhpPhpUnSerialize) //PHP反序列化
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
}

func setStar() {
	base.Component.TGin.GinPost(`/api/StarList`, controller.StarList)
	base.Component.TGin.GinPost(`/api/StarAdd`, controller.StarAdd)
	base.Component.TGin.GinPost(`/api/StarDel`, controller.StarDel)
}

func variable() {
	base.Component.TGin.GinPost(`/api/VariableList`, controller.VariableList)
	base.Component.TGin.GinPost(`/api/VariableAdd`, controller.VariableAdd)
	base.Component.TGin.GinPost(`/api/VariableDel`, controller.VariableDelete)
	base.Component.TGin.GinPost(`/api/VariableInfo`, controller.VariableInfo)
	base.Component.TGin.GinPost(`/api/VariableCmdAdd`, controller.VariableCmdAdd)
	base.Component.TGin.GinPost(`/api/VariableCmdDel`, controller.VariableCmdDelete)
	base.Component.TGin.GinPost(`/api/VariableRunPre`, controller.VariableCmdRunPre)         //执行第一步
	base.Component.TGin.GinPost(`/api/VariableRunProcess`, controller.VariableCmdRunProcess) //执行中收集信息
	base.Component.TGin.GinPost(`/api/VariableRunDone`, controller.VariableCmdRunDone)       //执行
}

func smartLink() {
	base.Component.TGin.GinPost(`/api/SmartLinkList`, controller.SmartLinkList)
	base.Component.TGin.GinPost(`/api/SmartLinkAdd`, controller.SmartLinkAdd)
	base.Component.TGin.GinPost(`/api/SmartLinkDel`, controller.SmartLinkDelete)
	base.Component.TGin.GinPost(`/api/SmartLinkInfo`, controller.SmartLinkInfo)
	base.Component.TGin.GinPost(`/api/SmartLinkProcessAdd`, controller.SmartLinkProcessAdd)
	base.Component.TGin.GinPost(`/api/SmartLinkProcessDel`, controller.SmartLinkProcessDel)
	base.Component.TGin.GinPost(`/api/SmartLinkRun`, controller.SmartLinkRunPlaywright)
}
