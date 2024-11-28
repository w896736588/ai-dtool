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
		case define.VariableCmdInput: //输入框 不存在替换
			if variableForm.Input.Value != `` {
				h.addReplace(&replaceList, variableForm.ResultKey, variableForm.Input.Value)
				h.sendSocketMsg(variableForm.VariableId, variableForm.Input.Label+`：`+variableForm.Input.Value)
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
			if !h.isPreShowForm(variableForm.Select.Options) {
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
		case define.VariableCmdRedisChoose: //选择redis 这个不需要替换
		case define.VariableCmdSshChoose: //选择ssh 这个不需要替换 准备链接
			variableForm.Select.Options = h.replace(variableForm.Select.Options, replaceList)
			if variableForm.Select.Value == `` {
				variableForm.IsRunOk = 0
			}
			if len(variableForm.Select.OptionList) > 0 {
				for _, option := range variableForm.Select.OptionList {
					if option.Value == variableForm.Select.Value {
						h.sendSocketMsg(variableForm.VariableId, variableForm.Select.Label+`：`+option.Label)
						if cast.ToInt(variableForm.VariableType) == define.VariableCmdSshChoose {
							h.sendSocketMsg(variableForm.VariableId, `尝试建立链接：`+option.Label)
							preConnErr := h.preConnSsh(map[string]any{
								`ssh_id`: option.Value,
								`name`:   option.Label,
							})
							if preConnErr != nil {
								return nil, nil, 0, preConnErr
							}
						}
					}
				}
			}
			break
		case define.VariableCmdMysql: //执行sql
			variableForm.Sql.Sql = h.replace(variableForm.Sql.Sql, replaceList)
			if !h.isPreShowForm(variableForm.Sql.Sql) {
				variableForm.IsRunOk = 0
				break
			}
			//执行sql
			sqlRet := h.sqlProcessRun(&variableForm, &replaceList)
			if sqlRet != nil {
				return nil, nil, 0, sqlRet
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
func (h *VariableRun) ProcessSet(variableId, variableCmdName string, variableForm *_struct.VariableForm) {
	variableForm.IsShowOk = 1 //默认显示是
	switch cast.ToInt(variableForm.VariableType) {
	case define.VariableCmdRadio: //单选
		if !h.isPreShowForm(variableForm.Select.Options) {
			variableForm.IsShowOk = 0 //不显示
			h.sendSocketMsg(variableId, `开始检查：`+variableCmdName+`,等待补充选项后展示`)
			return
		}
		h.sendSocketMsg(variableId, `开始检查：`+variableCmdName+`,可以展示`)
	case define.VariableCmdMysql: //执行sql
		h.sendSocketMsg(variableId, `开始检查：`+variableCmdName+`,初始化完成`)
		variableForm.IsShowOk = 0
		return
	default:
		h.sendSocketMsg(variableId, `开始检查：`+variableCmdName+`,可以展示`)
		break
	}
}
