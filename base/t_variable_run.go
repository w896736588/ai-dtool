package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"github.com/spf13/cast"
)

// RunProcess 执行前收集一些选择或者输入项 可以多次调用 有些待输入的还有替换符 可以多次执行 这里不管是否显示参数
func (h *VariableRun) RunProcess(variableFormList []_struct.VariableForm, replaceList []map[string]string) ([]_struct.VariableForm, []map[string]string, int, error) {
	needInputNum := len(variableFormList)
	inputNum := 0
	for key, variableForm := range variableFormList {
		variableForm.IsRunOk = 1 //预设该项已经执行过
		switch cast.ToInt(variableForm.VariableType) {
		case define.VariableCmdInput, define.VariableCmdTextarea: //输入框 不存在替换
			if variableForm.Input.Value != `` {
				h.addReplace(&replaceList, variableForm.ResultKey, variableForm.Input.Value)
				if h.CmdId == variableForm.Id {
					h.StreamMsg(Component.TMarkDown.BlockQuote(variableForm.Input.Label+","+variableForm.Input.Value), true)
				}
			} else {
				variableForm.IsRunOk = 0
			}
			break
		case define.VariableCmdRadio: //单项选择
			variableForm.Select.Options = h.replace(variableForm.Select.Options, replaceList)
			radioErr := h.PreRadioOptionList(&variableForm)
			if radioErr != nil {
				return nil, nil, 0, radioErr
			}
			if h.isExistReplaceParam(variableForm.Select.Options) {
				variableForm.IsRunOk = 0
				break
			}
			if variableForm.Select.Value == `` {
				variableForm.IsRunOk = 0
			} else {
				replaceChooseErr := h.radioChooseReplace(&variableForm, &replaceList, variableForm.Select.Value)
				if replaceChooseErr != nil {
					return nil, nil, 0, replaceChooseErr
				}
			}
			break
		case define.VariableCmdMysql: //执行sql
			variableForm.Sql.Sql = h.replace(variableForm.Sql.Sql, replaceList)
			if h.isExistReplaceParam(variableForm.Sql.Sql) {
				variableForm.IsRunOk = 0
				break
			}
			if h.isExistReplaceList(variableForm.ResultKey, replaceList) {
				break
			}
			//执行sql
			sqlRet := h.sqlProcessRun(&variableForm, &replaceList)
			if sqlRet != nil {
				return nil, nil, 0, sqlRet
			} else {
				if h.CmdId == variableForm.Id {
					h.StreamMsg(Component.TMarkDown.Code(variableForm.Sql.Sql, `sql`), true)
				}
			}
			break
		default:
			variableForm.IsRunOk = 1
			break
		}
		inputNum += variableForm.IsRunOk
		variableFormList[key] = variableForm
	}
	//是否能够运行
	isCanRun := 1
	if inputNum < needInputNum {
		isCanRun = 0
	}
	return variableFormList, replaceList, isCanRun, nil
}

// ProcessSet 变更中进行检测
func (h *VariableRun) ProcessSet(variableId, cmdName string, variableForm *_struct.VariableForm) {
	variableForm.IsShowOk = 1 //默认显示是
	switch cast.ToInt(variableForm.VariableType) {
	case define.VariableCmdRadio: //单选
		if h.isExistReplaceParam(variableForm.Select.Options) {
			variableForm.IsShowOk = 0 //不显示
			return
		}
	case define.VariableCmdMysql: //执行sql
		variableForm.IsShowOk = 0
		return
	default:
		break
	}
}
