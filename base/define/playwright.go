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
const CanvasImage ProcessType = `canvas_image`                      //提取canvas中的图片
const ExistWait ProcessType = `exist_wait`                          //等待元素出现
const NoExistWait ProcessType = `no_exist_wait`                     //等待元素消息
const WaitUrl ProcessType = `wait_url`                              //等待某个url执行完成
const LoginUsernamePassword ProcessType = `login_username_password` //输入账号密码
const BoolExist = `bool_exist`                                      //是否存在

const ElementClick = `click`              //点击
const ElementTextContent = `text_content` //提取
const ElementInput = `input`              //输入
const ElementExist = `exist`              //元素存在
const ElementCount = `count`              //元素个数

const MaxUserDataIndex = 500

const RegisterLinkParseTypeStream = `stream`
const RegisterLinkParseTypeJson = `application/connect+json`

var (
	OpenTypeDirect        OpenType = 1 //直接打开链接 通过js，现有浏览器打开
	OpenTypeWebkitSilence OpenType = 2 //静默打开(内置核心打开)
	OpenTypeWebkitChrome  OpenType = 3 //浏览器打开(内置核心打开)
)

const (
	SmartLinkStatusNormal = iota + 1
	SmartLinkStatusDelete
)

const CombineTypeFind = 1 //自动查找可以用的context
const CombineTypeLast = 2 //使用上一次登录的context
const CombineTypeNo = 3   //每次打开新的context
const CombineTypeFix = 4  //固定id为索引
