package define

type Base struct {
	DbFileName string //db名称
	DbPath     string //数据库文件所在目录
	WebPath    string //前端dist所在目录
}

type DbConfig struct {
	DbName string //数据库文件名
	DbPath string //数据库文件所在目录
}

type WebConfig struct {
	WebPath string //dist目录
}

type Env struct {
	RootPath             string     //项目根目录
	PkgPath              string     //pkg目录
	AppName              string     //项目名称
	ConfigFile           string     //配置文件名
	ConfigPath           string     //配置文件目录
	DatabaseUpPath       string     //数据库升级目录
	MemoryDatabaseUpPath string     //记忆数据库升级目录
	LogPath              string     //日志目录
	NodePath             string     //node js可执行程序目录
	WebkitDriverPath     string     //浏览器核心目录
	WebkitDownloadPath   string     //浏览器核心下载临时文件数据目录
	WebkitDataPath       string     //浏览器核心用户数据目录
	PythonCommand        string     //python可执行命令
	Crawl4AIHost         string     //crawl4ai服务host
	Crawl4AIPort         string     //crawl4ai服务端口
	Crawl4AIBaseURL      string     //crawl4ai服务地址
	Crawl4AIDataPath     string     //crawl4ai数据目录
	Crawl4AIScriptPath   string     //crawl4ai服务脚本路径
	Ports                []string   //gin支持的端口
	ConfigBase           *Base      //基础配置
	DbConfig             *DbConfig  //数据库配置
	WebConfig            *WebConfig //web配置
}
