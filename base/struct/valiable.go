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
	VariableType string
	Input        VariableFormInput  `json:"Input,omitempty"`
	Select       VariableFormSelect `json:"Select,omitempty"`
	Sql          VariableFormSql    `json:"Sql,omitempty"`
	ResultKey    string
	IsPreOk      int //1准备好 0未准备好
}

type VariableFormInput struct {
	Label string
	Value string
}

type VariableFormSql struct {
	Sql     string
	MysqlId string
}

type VariableFormSelect struct {
	Label      string
	Value      string
	OptionList []VariableFormOption
	Options    string //原本的
}

type VariableFormOption struct {
	Label  string
	Value  string
	Source string //原本的值
}
