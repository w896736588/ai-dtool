package zhima

import (
	"dev_tool/base_module"
	"dev_tool/internal/app/zhima/controller"
	"fmt"
	"gitee.com/Sxiaobai/gs/gssocket"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"net/http"
	"strings"
)

var AppName = `zhima`
var Logger *gstool.GsLogger
var RootPath string
var ConfigViper *viper.Viper
var Encrypt *gstool.Encrypt //全局加密
var Global *base_module.Global
var GlobalGin *base_module.Global //专为gin处理的
var SocketServer gssocket.Server

func InitBase() {
	var err error
	RootPath, err = gstool.GetRootPath()
	if err != nil {
		panic(err.Error())
	}
	getConfigViper()
	getEncrypt()
	getLogger()
	getGin()
	getGlobal()
}

func Stop() {
	err := GlobalGin.GinStop(10)
	if err != nil {
		Global.Error(err.Error())
	}
}
func getLogger() {
	Logger = gstool.CreateLogger(RootPath+`/logs/`+AppName, ``)
	base_module.Logger = Logger
}
func getGlobal() {
	Global = &base_module.Global{}
	Global.SetLogger(Logger)
}
func getConfigViper() {
	ConfigViper = viper.New()
	ConfigViper.AddConfigPath(RootPath + `/config/` + AppName)
	ConfigViper.SetConfigName(`config`)
	ConfigViper.SetConfigType(`ini`)
	if err := ConfigViper.ReadInConfig(); err != nil {
		panic(err.Error())
	}
}
func getEncrypt() {
	Encrypt = &gstool.Encrypt{
		Key: ConfigViper.GetString(`encrypt.key`),
		Iv:  ConfigViper.GetString(`encrypt.iv`),
	}
}
func getGin() {
	host := ConfigViper.GetString(`run.host`)
	port := ConfigViper.GetString(`run.port`)
	GlobalGin = &base_module.Global{}
	GlobalGin.SetLogger(Logger)
	GlobalGin.GinInit(host, port)
	GlobalGin.GinSetAllowCrossDomain()
	GlobalGin.GinStatic(`/static`, `./views/dist/static`)
	GlobalGin.GinLoadHTMLFiles(`views/dist/index.html`)
	GlobalGin.GinGet(`/`, func(context *gin.Context) {
		context.HTML(200, `index.html`, nil)
	})
	baseRouter()
	redisRouter()
	phpRouter()
	consumerRouter()
	gitRouter()
	wechatKefuRouter()
	vipRouter()
	dockerRouter()
	loginRouter()
	initSocket()
	GlobalGin.GinRun()
}

func GetClientId(unikey, shellName string) string {
	return unikey + `#` + shellName
}

func initSocket() {
	SocketServer = gssocket.Server{
		Host:        fmt.Sprintf(`0.0.0.0:%s`, ConfigViper.GetString(`run.wsPort`)),
		Uri:         `/socket`,
		AllowOrigin: true,
	}
	SocketServer.GetClientFunc = func(r *http.Request) string {
		return GetClientId(r.FormValue(`Unikey`), r.FormValue(`ShellName`))
	}
	SocketServer.ReceMsgFunc = func(s string) string {
		return ``
	}
	SocketServer.ConnectFunc = func(clientId string, conn *websocket.Conn) {
		connParams := strings.Split(clientId, `#`)
		shell, err := base_module.GetGlobal(connParams[0]).ShellPushGetClient(connParams[1])
		if err != nil {
			gstool.FmtPrintlnLog(`获取client失败 %s`, err.Error())
		} else {
			shell.WsConn = conn
		}
	}
	go SocketServer.Start()
}

//基础接口
func baseRouter() {
	GlobalGin.GinPost(`/api/BaseLogin`, controller.BaseLogin)                       //登录
	GlobalGin.GinPost(`/api/BaseRegisterService`, controller.BaseRegisterService)   //注册各类服务 CheckUnikeyExist
	GlobalGin.GinPost(`/api/BaseCheckUnikeyExist`, controller.BaseCheckUnikeyExist) //检查unikey是否已经登录注册
}

//redis相关
func redisRouter() {
	GlobalGin.GinPost(`/api/RedisAvailableList`, controller.RedisAvailableList) //可用的redis列表
	GlobalGin.GinPost(`/api/RedisSearch`, controller.RedisSearch)               //查询某个key
	GlobalGin.GinPost(`/api/RedisKeys`, controller.RedisKeys)                   //模糊搜索key
	GlobalGin.GinPost(`/api/RedisKeysType`, controller.RedisKeysType)           //批量获取key缓存类型
	GlobalGin.GinPost(`/api/RedisKeyType`, controller.RedisKeyType)             //获取key类型
	GlobalGin.GinPost(`/api/RedisSaveString`, controller.RedisSaveString)       //保存string
	GlobalGin.GinPost(`/api/RedisDelKey`, controller.RedisDelKey)               //删除key
	GlobalGin.GinPost(`/api/RedisDelSub`, controller.RedisDelSub)               //删除sub key
	GlobalGin.GinPost(`/api/RedisEditTtl`, controller.RedisEditTtl)             //更改ttl
	GlobalGin.GinPost(`/api/RedisDeleteAll`, controller.RedisDelAllKey)         //删除所有缓存
	GlobalGin.GinPost(`/api/RedisCreateCache`, controller.RedisCreateCache)     //创建缓存
	GlobalGin.GinPost(`/api/RedisEditSub`, controller.RedisEditSub)             //编辑二级缓存
}

//php相关
func phpRouter() {
	GlobalGin.GinPost(`/api/PhpUnserialize`, controller.PhpPhpUnSerialize) //PHP反序列化
}

//消费者相关
func consumerRouter() {
	GlobalGin.GinPost(`/api/ConsumerRestartAll`, controller.ConsumerRestartAll) //重启所有消费者
	GlobalGin.GinPost(`/api/ConsumerStopAll`, controller.ConsumerStopAll)       //重启所有消费者
	GlobalGin.GinPost(`/api/ConsumerStatusList`, controller.ConsumerStatusList) //查看消费者状态
	GlobalGin.GinPost(`/api/ConsumerConfigShow`, controller.ConsumerConfigShow) //查看消费者配置
	GlobalGin.GinPost(`/api/ConsumerRestart`, controller.ConsumerRestart)       //重启单个消费者
	GlobalGin.GinPost(`/api/ConsumerStop`, controller.ConsumerStop)             //重启单个消费者
	GlobalGin.GinPost(`/api/ConsumerConfigList`, controller.ConsumerConfigList) //查看所有的配置
}

//git相关
func gitRouter() {
	GlobalGin.GinPost(`/api/GitQueryCurrentBranch`, controller.GitCurrentBranch)  //查询当前分支
	GlobalGin.GinPost(`/api/GitChangeBranch`, controller.GitChangeBranch)         //切换分支
	GlobalGin.GinPost(`/api/GitPullBranchOrigin`, controller.GitPullBranchOrigin) //拉取最新分支
	GlobalGin.GinPost(`/api/GitQueryStatus`, controller.QueryStatus)              //查询分支本地状态
}

//微信客服相关
func wechatKefuRouter() {
	GlobalGin.GinPost(`/api/WechatKefuStatus`, controller.WechatKefuStatus)                 //查询当前应用启动情况
	GlobalGin.GinPost(`/api/WechatKefuChange`, controller.WechatKefuChange)                 //切换微信客服环境
	GlobalGin.GinPost(`/api/WechatKefuQueryAppList`, controller.WechatKefuQueryAppList)     //查询微信客服列表
	GlobalGin.GinPost(`/api/WechatKefuQueryQrCdeList`, controller.WechatKefuQueryQrCdeList) //查询渠道客服二维码列表
}

//vip相关
func vipRouter() {
	GlobalGin.GinPost(`/api/VipChange`, controller.VipChange) //切换vip
	GlobalGin.GinPost(`/api/VipQuery`, controller.VipQuery)   //查询vip
}

//login相关
func loginRouter() {
	GlobalGin.GinPost(`/api/LoginLink`, controller.LoginLink) //拿到登录链接
}

//Docker相关
func dockerRouter() {
	GlobalGin.GinPost(`/api/DockerRestart`, controller.DockerRestart)         //重启Docker DockerShowCompose
	GlobalGin.GinPost(`/api/DockerShowCompose`, controller.DockerShowCompose) //查看配置文件
	GlobalGin.GinPost(`/api/DockerExec`, controller.DockerExec)               //执行命令
	GlobalGin.GinPost(`/api/DockerPs`, controller.DockerPs)                   //ps
}
