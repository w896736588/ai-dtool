package _struct

import (
	"dev_tool/base/define"
)

type SmartLinkRunParams struct {
	Id                  int                   //链接ID
	Link                string                //打开的链接
	SmartLinkUniqueKey  string                //在链接下面的唯一值   索引值_label 例如第一个链接 id_label
	OpenNum             int                   //打开次数 0会被默认为1次
	Cookie              string                //打开链接时需要设置的cookie
	Headers             map[string]string     //设置的headers
	OpenType            define.OpenType       //打开类型 1通过js打开
	IsCombine           bool                  //是否合并到同一浏览器 true合并，false不合并
	ProcessList         []map[string]any      //执行流程
	ReplaceList         []map[string]string   //替换内容
	IsSaveUserData      bool                  //是否保存用户数据 true保存，false不保存
	BrowserAuthUsername string                //浏览器自带验证用户名
	BrowserAuthPassword string                //浏览器自带验证密码
	Domain              string                //域名
	Timeout             float64               //超时时间秒
	UserName            string                //选择的登录账号
	Password            string                //登录密码
	FixDataId           int                   //是否固定保存数据目录
	DownloadFinds       []string              //哪些url请求会被定义为下载
	AutoCloseSecond     int                   //多少秒内没有操作 就进行关闭page 0表示不处理
	ListenUrl           map[string]*ListenUrl //监听url
}

type ListenUrl struct {
	IsSse         bool
	Callback      func(string, error)
	StartCallBack func()
	EndCallBack   func(msg string)
}
