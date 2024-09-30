package define

// 变量类型
const (
	VariableTypeMysql        = iota + 1 //mysql执行
	VariableTypeCmd                     //从命令集中取值
	VariableTypeInput                   //等待客户输入
	VariableTypeRandStr                 //不重复字符串
	VariableTypeCurl                    //从curl中取值
	VariableTypePython                  //python脚本
	VariableTypeDockerChoose            //docker列表
	VariableTypeBash                    //bash脚本
	VariableTypeRadio                   //单项选择
	VariableTypeLink                    //地址跳转
	VariableTypeRedisDelete             //删除redis缓存
	VariableTypeRedisChoose             //选择redis
	VariableTypeMysqlChoose             //选择mysql
)

const (
	VariableStatusNormal = iota + 1
	VariableStatusDelete
)
