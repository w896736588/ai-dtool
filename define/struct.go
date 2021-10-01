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
	Default     int
	UniKey      string
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

type Response struct {
	Errcode int         `json:"ErrCode"`
	Errmsg  string      `json:"ErrMsg"`
	Data    interface{} `json:"Data"`
}
