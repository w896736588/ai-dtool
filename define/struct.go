package define

type RedisConfig struct {
	Name        string
	Host        string
	Password    string
	PoolSize    string
	SshHost     string
	SshPort     string
	SshUser     string
	SshPassword string
	SshName     string
	SshPrefix   string
	Default     int
	UniKey      string
	Connection  bool
}
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
	ParamOne  string    `json:"ParamOne"`  //根据ExecType定义具体代表啥内容
	ParamTwo  string    `json:"ParamTwo"`  //根据ExecType定义具体代表啥内容
	EnvName   string    `json:"EnvName"`   //环境名
	ExecType  string    `json:"ExecType"`  //执行类型
	SshConfig SshConfig `json:"SshConfig"` //ssh连接配置
}
