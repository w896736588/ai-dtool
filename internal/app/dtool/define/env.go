package define

type Base struct {
	DbFileName   string
	DbPath       string
	LogDbPath    string
	MemoryDBPath string
	MemoryDBName string
	WebPath      string
}

type DbConfig struct {
	DbName string
	DbPath string
}

type WebConfig struct {
	WebPath string
}

type SmartLinkConfig struct {
	RunMode       SmartLinkRunMode
	ClientVersion string
	SourcePath    string
}

type Env struct {
	RootPath           string
	PkgPath            string
	AppName            string
	ConfigFile         string
	ConfigPath         string
	DatabaseUpPath     string
	LogDatabaseUpPath  string
	LogPath            string
	NodePath           string
	WebkitDriverPath   string
	WebkitDownloadPath string
	WebkitDataPath     string
	PythonCommand      string
	Ports              []string
	ApiPorts           []string
	SsePorts           []string
	ConfigBase         *Base
	DbConfig           *DbConfig
	LogDbConfig        *DbConfig
	WebConfig          *WebConfig
	SmartLinkConfig    *SmartLinkConfig
}
