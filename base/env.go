package base

type Env struct {
	IsBuild            bool   //true 打包模式 false go run模式
	RootPath           string //项目根目录
	PkgPath            string //pkg目录
	AppName            string //项目名称
	ConfigPath         string //配置文件目录
	LogPath            string //日志目录
	WebkitPath         string //浏览器核心目录
	PlaywrightDownload string //浏览器核心下载临时文件数据目录
	PlaywrightUserData string //浏览器核心用户数据目录
}
