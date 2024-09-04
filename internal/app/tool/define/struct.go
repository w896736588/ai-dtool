package define

import "gitee.com/Sxiaobai/gs/gsdb"

type RequestBody struct {
	UniKey         string `json:"UniKey"`
	CacheKey       string `json:"cacheKey"`
	CacheSerialize string `json:"cacheSerialize"`
	CacheType      string `json:"cacheType"`
}

type SearchBody struct {
	UniKey string `json:"UniKey"`
	Search string `json:"search"`
}

type SearchKeysTypeBody struct {
	UniKey   string   `json:"UniKey"`
	KeysList []string `json:"KeysList"`
}

type SerializeBody struct {
	SerializeStr string `json:"SerializeStr"`
}

type SaveString struct {
	UniKey   string `json:"UniKey"`
	CacheKey string `json:"Key"`
	Value    string `json:"Value"`
}

type TypeResponse struct {
	Type string `json:"Type"`
	TTL  int    `json:"TTL"`
}

type DelKey struct {
	UniKey   string `json:"UniKey"`
	CacheKey string `json:"Key"`
}

type EditTTL struct {
	UniKey   string `json:"UniKey"`
	CacheKey string `json:"Key"`
	TTL      int    `json:"TTL"`
}

type DelAllKey struct {
	UniKey    string   `json:"UniKey"`
	CacheKeys []string `json:"Keys"`
}

type Response struct {
	Errcode int         `json:"ErrCode"`
	Errmsg  string      `json:"ErrMsg"`
	Data    interface{} `json:"Data"`
}

type KeysList struct {
	CacheKey string
	Type     string
	Loading  bool
}

type DelSub struct {
	UniKey    string `json:"UniKey"`
	CacheType string `json:"cacheType"`
	CacheKey  string `json:"cacheKey"`
	Sub       string `json:"sub"`
}

type CreateCache struct {
	UniKey      string  `json:"UniKey"`
	CacheType   string  `json:"cacheType"`
	CacheKey    string  `json:"cacheKey"`
	CacheField  string  `json:"cacheField"`
	CacheValue  string  `json:"cacheValue"`
	TTL         int     `json:"ttl"`
	CacheMember string  `json:"cacheMember"`
	CacheScore  float64 `json:"cacheScore"`
	BoolCreate  int     `json:"boolCreate"`
	LPushValue  string  `json:"lPushValue"`
	RPushValue  string  `json:"rPushValue"`
}

type EditSub struct {
	UniKey      string  `json:"UniKey"`
	CacheType   string  `json:"cacheType"`
	CacheKey    string  `json:"cacheKey"`
	CacheField  string  `json:"cacheField"`
	CacheValue  string  `json:"cacheValue"`
	CacheIndex  int64   `json:"index"`
	CacheMember string  `json:"cacheMember"`
	CacheScore  float64 `json:"cacheScore"`
}
type SshConfig struct {
	UniKey   string
	Name     string
	Username string
	Password string
	Host     string
	Port     string
	SshType  string
}
type SshDo struct {
	UniKey string `json:"UniKey"`
}
type CmdStruct struct {
	ConfigFile string
	Name       string
}

type WebSocketStruct struct {
	Host string
	Port string
}

type WebSocketEventStruct struct {
	Event       string `json:"Event"`
	ReqJsonBody string `json:"ReqBody"`
}

// SshExec xShell 执行命令操作
// @auth frog
// @date 2022-12-02 15:02:13
type SshExec struct {
	ParentType            string             `json:"ParentType"`      //系统类别
	CodePath              string             `json:"CodePath"`        // 代码目录 docker_apps/yii_customer_service   weike_customer_service4
	BranchName            string             `json:"BranchName"`      //分支名
	ExecType              string             `json:"ExecType"`        //执行类型
	SshConfig             SshConfig          `json:"SshConfig"`       //ssh连接配置
	WkSshConfig           SshConfig          `json:"WkSshConfig"`     //ssh连接配置
	SshName               string             `json:"SshName"`         //ssh name
	WechatKefuAppid       string             `json:"WechatKefuAppid"` //微信客服appid或者应用Id
	DockerList            []DockerConfig     `json:"DockerList"`
	DockerId              string             `json:"DockerId"`              //操作哪个docker
	DockerCodeName        string             `json:"DockerCodeName"`        //docker name  common3  common4等
	DockerCodePath        string             `json:"DockerCodePath"`        //docker内代码路径
	SupervisorRestartName string             `json:"SupervisorRestartName"` //消费者重启名
	SupervisorConfigPath  string             `json:"SupervisorConfigPath"`  // 消费者配置内容
	RedisConfigList       []gsdb.RedisConfig `json:"redisConfigList"`       //redis配置列表
	LogFile               string             `json:"LogFile"`               //日志文件名
	DockerExecCommand     string             `json:"DockerExecCommand"`     //docker内执行的命令
	XkfDevDbConfig        gsdb.MysqlConfig   `json:"xkfDevDbConfig"`        //数据库链接
	Account               string             `json:"Account"`               //管理员ID或者账号
	VipLevel              interface{}        `json:"VipLevel"`              //vip版本
	SystemType            interface{}        `json:"SystemType"`            //系统类型
	ExpiredDay            interface{}        `json:"expiredDay"`            //VIP过期时间
	LoginUrl              string             `json:"loginUrl"`              //登录后跳转的路由
	LoginHost             string             `json:"loginHost"`             //登录路由
}

type DockerConfig struct {
	Name    string `json:"Name"`
	Id      string `json:"Id"`
	SshName string `json:"SshName"`
}

type TblWechatapp struct {
	Id      string `json:"_id"`
	Appid   string `json:"app_id"`
	AppType string `json:"app_type"`
	UserId  string `json:"user_id"`
	AppName string `json:"app_name"`
}

type TblUser struct {
	Id       string `json:"_id"`
	Username string `json:"user_name"`
}

type TblVip struct {
	UserId      string `json:"user_id"`
	VipType     string `json:"vip_type"`
	ExpiredTime string `json:"expired_time"`
}
