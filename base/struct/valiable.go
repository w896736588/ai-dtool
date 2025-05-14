package _struct

type VariableMysql struct {
	MysqlId any `json:"mysql_id"` //数据库id
	Sql     any `json:"sql"`      //执行的sql
}

type VariableCmd struct {
	CmdId any `json:"cmd_id"`
}
type VariableCurl struct {
	Headers any `json:"headers"`
	Method  any `json:"method"`
}

type VForm struct {
	Name       string
	VariableId string
	Id         string
	CmdType    string
	Input      VFormInput  `json:"Input,omitempty"`
	Select     VFormSelect `json:"Select,omitempty"`
	Sql        VFormSql    `json:"Sql,omitempty"`
	Bash       VFormBash   `json:"Bash,omitempty"`
	ResultKey  string
	IsShowOk   int  //1准备好 0未准备好  准备好了以后就会在页面上显示选项等
	IsRunOk    int  //1准备好执行（需要选择） 全部准备好以后就是可以执行了
	IsFinish   bool //是否已经结束
}

type VFormInput struct {
	Label       string
	Value       string
	HideSureBtn int
}

type VFormSql struct {
	Sql     string
	MysqlId string
}

type VFormBash struct {
	Bash  string
	SshId string
}

type VFormSelect struct {
	Label      string
	Value      string
	OptionList []VFormOption
	Options    string //原本的
}

func (h *VFormSelect) GetSelectOption(value string) VFormOption {
	for _, option := range h.OptionList {
		if option.Value == value {
			return option
		}
	}
	return VFormOption{}
}

type VFormOption struct {
	Label  string
	Value  string
	Source string //原本的值
}

type VCmdResult struct {
	Form        VForm               //显示的表单
	RunStatus   int                 //0不可以执行 1可以执行 2执行结束
	ReplaceList []map[string]string //替换数据
	RunUniqueId string              //执行唯一ID
	VariableId  int                 //ID
}
