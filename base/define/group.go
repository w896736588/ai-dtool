package define

// 分组类型
const (
	GroupTypeGit       = iota + 1 //git
	GroupTypeCmd                  //命令集
	GroupTypeVariable             //变量
	GroupTypeSmartLink            //自动化链接
	GroupTypeAccount              //账号分组 用于自动化链接
	GroupTypeShellOut             //终端输出
)

func GetGroupTypeList() []int {
	return []int{
		GroupTypeGit,
		GroupTypeCmd,
		GroupTypeVariable,
		GroupTypeSmartLink,
		GroupTypeAccount,
		GroupTypeShellOut,
	}
}
