package define

type ProcessCode string

const ProcessContinue ProcessCode = "continue"
const ProcessBreak ProcessCode = "break"
const ProcessOk ProcessCode = "ok"

type OpenType int

type ProcessType string

const TextContent ProcessType = `text_content` //提取内容
const BoolResult ProcessType = `bool_result`
const Close ProcessType = `close`
const Wait ProcessType = `wait`
const WaitClose ProcessType = `wait_close`
const Click ProcessType = `click`
const Input ProcessType = `input`
const RedirectUri ProcessType = `redirect_uri`
const Exit ProcessType = `exit` //结束

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
