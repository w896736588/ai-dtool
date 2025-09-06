package _struct

import (
	"dev_tool/base/define"
)

type PlaywrightRunParams struct {
	Id                  int                                              //链接ID
	Link                string                                           //打开的链接
	LinkIdLabel         string                                           //在链接下面的唯一值   索引值_label 例如第一个链接 id_label
	OpenNum             int                                              //打开次数 0会被默认为1次
	Cookie              string                                           //打开链接时需要设置的cookie
	Headers             map[string]string                                //设置的headers
	OpenType            define.OpenType                                  //打开类型 1通过js打开 2 静默打开(内置核心打开)  3 浏览器打开(内置核心打开)
	CombineType         int                                              //查找context方案
	ProcessList         []map[string]any                                 //执行流程
	ReplaceList         map[string]string                                //替换内容
	BrowserAuthUsername string                                           //浏览器自带验证用户名
	BrowserAuthPassword string                                           //浏览器自带验证密码
	Domain              string                                           //域名
	Scheme              string                                           //协议
	LocatorTimeout      float64                                          //获取元素超时时间秒
	GetPageTimeout      float64                                          //开启page超时时间
	LastIndexLabel      string                                           //用于查找最后一次使用的index 优先赋值前端传过来的userName,其次赋值label
	LinkId              string                                           //context唯一ID link_id_xx
	DownloadFinds       []string                                         //哪些url请求会被定义为下载
	AutoCloseSecond     int                                              //多少秒内没有操作 就进行关闭page 0表示不处理
	Channel             string                                           //浏览器类型
	RunCallFunc         func(define.ProcessType, string, string, string) //注册输出回调
	StreamFunc          func(string, string)                             //执行输出
	ListenUrlList       map[string]*ListenUrl                            //监听
	ResponseUrls        []*ProcessResponseUrl                            //注册等待请求完成
	ShowCookies         []ShowCookie                                     //信息提取
}

type ListenUrl struct {
	ParseType     string
	IsSse         bool
	Callback      func(string, string, error)
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
