package define

type OpenType int

type CmdType string

const TextContent CmdType = `text_content` //提取内容
const BoolResult CmdType = `bool_result`
const Close CmdType = `close`
const Wait CmdType = `wait`
const WaitClose CmdType = `wait_close`
const Click CmdType = `click`
const Input CmdType = `input`
const RedirectUri CmdType = `redirect_uri`
const Exit CmdType = `exit` //结束

const ElementClick = `click`              //点击
const ElementTextContent = `text_content` //提取
const ElementInput = `input`              //输入
const ElementExist = `exist`              //元素存在
const ElementCount = `count`              //元素个数

var (
	OpenTypeDirect        OpenType = 1 //直接打开链接 通过js，现有浏览器打开
	OpenTypeWebkitSilence OpenType = 2 //静默打开(内置核心打开)
	OpenTypeWebkitChrome  OpenType = 3 //浏览器打开(内置核心打开)
)

const (
	SmartLinkStatusNormal = iota + 1
	SmartLinkStatusDelete
)
