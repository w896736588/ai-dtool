package define

const (
	OpenTypeDirect        = iota + 1 //直接打开链接 通过js，现有浏览器打开
	OpenTypeWebkitSilence            //静默打开(内置核心打开)
	OpenTypeWebkitChrome             //浏览器打开(内置核心打开)
)

const (
	SmartLinkStatusNormal = iota + 1
	SmartLinkStatusDelete
)

const (
	SmartLinkProcessClick    = iota + 1 //选中并点击
	SmartLinkProcessUsername            //选中并输入账号 链接如果配置了账号密码 那么就自动输入
	SmartLinkProcessPassword            //选中并输入密码 链接如果配置了账号密码 那么就自动输入
)
