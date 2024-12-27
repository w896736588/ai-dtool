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
	VariableCmdRedisDelete             //删除redis缓存
	VariableCmdRedisChoose             //选择redis
	VariableCmdMysqlChoose             //选择mysql
	VariableCmdSshChoose               //选择ssh
)

const (
	VariableStatusNormal = iota + 1
	VariableStatusDelete
)

const (
	VariableTypeBash = iota + 1 //脚本
	VariableTypeLink            //链接
)
