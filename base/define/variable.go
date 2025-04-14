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
)

const (
	VariableStatusNormal = iota + 1
	VariableStatusDelete
)

const (
	VariableTypeBash = iota + 1 //脚本
	VariableTypeLink            //链接
)
