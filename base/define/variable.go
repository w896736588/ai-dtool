package define

// 变量类型
const (
	VariableTypeMysql   = iota + 1 //从mysql中取值
	VariableTypeCmd                //从命令集中取值
	VariableTypeInput              //等待客户输入
	VariableTypeRandStr            //不重复字符串
	VariableTypeCurl               //从curl中取值
	VariableTypePython             //python脚本
	VariableTypeDocker             //docker列表
	VariableTypeBash               //bash脚本
	VariableTypeRadio              //单项选择
)
