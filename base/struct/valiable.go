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

type VariableForm struct {
	Name       string
	VariableId string
	Id         string
	CmdType    string
	Input      VariableFormInput  `json:"Input,omitempty"`
	Select     VariableFormSelect `json:"Select,omitempty"`
	Sql        VariableFormSql    `json:"Sql,omitempty"`
	Bash       VariableFormBash   `json:"Bash,omitempty"`
	Link       VariableFormLink   `json:"Link,omitempty"`
	ResultKey  string
	IsShowOk   int  //1准备好 0未准备好  准备好了以后就会在页面上显示选项等
	IsRunOk    int  //1准备好执行（需要选择） 全部准备好以后就是可以执行了
	IsFinish   bool //是否已经结束
}

type VariableFormInput struct {
	Label       string
	Value       string
	HideSureBtn int
}

type VariableFormSql struct {
	Sql     string
	MysqlId string
}

type VariableFormBash struct {
	Bash  string
	SshId string
}

type VariableFormLink struct {
	Link string
	Desc string
}

type VariableFormSelect struct {
	Label      string
	Value      string
	OptionList []VariableFormOption
	Options    string //原本的
}

func (h *VariableFormSelect) GetSelectOption(value string) VariableFormOption {
	for _, option := range h.OptionList {
		if option.Value == value {
			return option
		}
	}
	return VariableFormOption{}
}

type VariableFormOption struct {
	Label  string
	Value  string
	Source string //原本的值
}

type VariableCmdResult struct {
	Form        VariableForm        //显示的表单
	RunStatus   int                 //0不可以执行 1可以执行 2执行结束
	ReplaceList []map[string]string //替换数据
	RunUniqueId string              //执行唯一ID
}
