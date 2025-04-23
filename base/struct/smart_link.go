package _struct

import (
	"dev_tool/base/define"
)

type PlaywrightRunParams struct {
	Id                  int                                          //链接ID
	Link                string                                       //打开的链接
	SmartLinkUniqueKey  string                                       //在链接下面的唯一值   索引值_label 例如第一个链接 id_label
	OpenNum             int                                          //打开次数 0会被默认为1次
	Cookie              string                                       //打开链接时需要设置的cookie
	Headers             map[string]string                            //设置的headers
	OpenType            define.OpenType                              //打开类型 1通过js打开 2 静默打开(内置核心打开)  3 浏览器打开(内置核心打开)
	IsCombine           bool                                         //是否合并到同一浏览器 true合并，false不合并
	ProcessList         []map[string]any                             //执行流程
	ReplaceList         []map[string]string                          //替换内容
	IsSaveUserData      bool                                         //是否保存用户数据 true保存，false不保存
	BrowserAuthUsername string                                       //浏览器自带验证用户名
	BrowserAuthPassword string                                       //浏览器自带验证密码
	Domain              string                                       //域名
	LocatorTimeout      float64                                      //获取元素超时时间秒
	GetPageTimeout      float64                                      //开启page超时时间
	UserName            string                                       //选择的登录账号
	Password            string                                       //登录密码
	FixDataId           int                                          //是否固定保存数据目录
	DownloadFinds       []string                                     //哪些url请求会被定义为下载
	AutoCloseSecond     int                                          //多少秒内没有操作 就进行关闭page 0表示不处理
	Channel             string                                       //浏览器类型
	RunCallFunc         func(define.CmdType, string, string, string) //注册输出回调
}

type ListenUrl struct {
	IsSse         bool
	Callback      func(string, error)
	StartCallBack func(string)
	EndCallBack   func(msg string)
}

type ShowCookie struct {
	FindType     string   `json:"find_type"`      //查找类型 cookie 直接根据cookie的key进行匹配  any 任意值中进行处理
	FormatList   []string `json:"format_list"`    //对值进行格式化类型 url_decode
	FindKey      string   `json:"find_key"`       //查找的key
	RegexFindKey string   `json:"regex_find_key"` //正则匹配的key
	Label        string   `json:"label"`
	DomainList   []string `json:"Domain_list"`
}

type Locator struct {
	Locator     string //寻找
	First       bool   //是否只寻找第一个
	ExistSetNot bool   //如果存在时那么就表示Locator不存在
}

type ElementOp struct {
	Type        string
	FillValue   string
	TextContent string
	Count       int //元素数量
}
