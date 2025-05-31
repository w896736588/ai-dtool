package define

// 变量类型
const (
	VariableCmdMysql        = iota + 1 //mysql执行
	VariableCmdCmd                     //从命令集中取值
	VariableCmdInput                   //等待客户输入
	VariableCmdRandStr                 //不重复字符串
	VariableCmdCurl                    //从curl中取值
	VariableCmdPython                  //python脚本
	VariableCmdDockerChoose            //docker列表
	VariableCmdBash                    //bash脚本
	VariableCmdRadio                   //单项选择
	VariableCmdLink                    //地址跳转
	VariableCmdRedis                   //redis缓存操作
	VariableCmdRedisChoose             //选择redis 【废弃】
	VariableCmdMysqlChoose             //选择mysql 【废弃】
	VariableCmdSshChoose               //选择ssh 【废弃】
	VariableCmdPlaywright              //选择playwright
	VariableCmdCombine                 //内容收集
	VariableCmdTextarea                //输入框 textarea
	VariableCmdCommand                 //直接执行命令 不同于复杂的bash（复杂的bash可能在磁盘空间不足时无法执行）
	VariableCmdMakefile                //makefile
	VariableCmdUpload                  //上传文件
)

const (
	VariableStatusNormal = iota + 1
	VariableStatusDelete
)

const (
	VariableTypeBash = iota + 1 //脚本
	VariableTypeLink            //链接
)

type RunResult string

const RunStatus RunResult = "to_next"           //继续下一个
const RunResultShowForm RunResult = "show_form" //让前端新增加一个form
const RunResultWaitRun RunResult = "wait_run"   //可以执行了，前端点击执行按钮
const RunResultBroken RunResult = "broken"      //报错中断

const RunStatusWaitRun = 0 //不可以执行
const RunStatusCanRun = 1  //不可以执行
const RunStatusFinish = 2

const RunTypeForm = `form`     //输出表单给前端
const RunTypeMiddle = `middle` //输出中间结果
const RunTypeRun = `run`       //最终执行，等待用户确认（一旦确认最终执行，那么往下的所有cmd都会直接执行，不会再考虑runType）
