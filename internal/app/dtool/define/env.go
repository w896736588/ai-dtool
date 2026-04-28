package define

type Base struct {
	DbFileName                   string // 主库 db file name / main database file name
	DbPath                       string // 主库目录 db directory / main database directory
	DbIsGitRepo                  bool   // 主库是否按 git 仓库处理 main db git repo flag / whether main db should be treated as a git repository
	DbAutoPushDelayMinutes       int    // 主库自动同步延迟分钟数 main db auto push delay minutes / auto-push interval for main database
	LogDbPath                    string // 日志库目录 log db directory / log database directory
	MemoryDBPath                 string // 记忆库目录 memory db directory / memory database directory
	MemoryDBName                 string // 记忆库文件名 memory db file name / memory database file name
	MemoryDBIsGitRepo            bool   // 记忆库是否按 git 仓库处理 memory git repo flag / whether memory db should be treated as a git repository
	MemoryDBAutoPushDelayMinutes int    // 知识库变更后延迟多少分钟自动 push / delayed auto-push minutes for memory database changes
	WebPath                      string // 前端 dist 目录 web dist path / frontend dist directory
}

type DbConfig struct {
	DbName      string // 数据库文件名 db file name / database file name
	DbPath      string // 数据库目录 db directory / database directory
	DbIsGitRepo bool   // 数据库是否按 git 仓库处理 db git repo flag / whether database should be treated as a git repository
}

type WebConfig struct {
	WebPath string // dist 目录 dist path / dist directory
}

type SmartLinkConfig struct {
	RunMode       SmartLinkRunMode // 运行模式 run mode / execution mode
	ClientVersion string           // 客户端版本要求 client version / required client version
	SourcePath    string           // 源代码目录 source path / Go source code root for cross-compilation
}

type Env struct {
	RootPath           string           // 项目根目录 root path / project root directory
	PkgPath            string           // pkg 目录 pkg path / package directory
	AppName            string           // 项目名称 app name / project name
	ConfigFile         string           // 配置文件名 config file name / config file name
	ConfigPath         string           // 配置文件目录 config path / config directory
	DatabaseUpPath     string           // 主库升级目录 database upgrade path / main database migration directory
	LogDatabaseUpPath  string           // 日志库升级目录 log database upgrade path / log database migration directory
	LogPath            string           // 日志目录 log path / log directory
	NodePath           string           // Node 路径 node path / Node executable path
	WebkitDriverPath   string           // 浏览器驱动目录 webkit driver path / browser driver directory
	WebkitDownloadPath string           // 浏览器下载目录 webkit download path / browser download directory
	WebkitDataPath     string           // 浏览器数据目录 webkit data path / browser data directory
	PythonCommand      string           // Python 命令 python command / Python executable command
	Ports              []string         // gin 端口 ports / gin ports
	ApiPorts           []string         // API 端口 api ports / API ports
	SsePort            string           // SSE 端口 sse port / SSE port
	ConfigBase         *Base            // 基础配置 base config / base configuration
	DbConfig           *DbConfig        // 主库配置 main db config / main database configuration
	LogDbConfig        *DbConfig        // 日志库配置 log db config / log database configuration
	WebConfig          *WebConfig       // Web 配置 web config / web configuration
	SmartLinkConfig    *SmartLinkConfig // 自定义网页配置 smart link config / smart link configuration
}
