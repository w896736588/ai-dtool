package p_variable

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"strings"
)

type VariableSet struct {
	VariableId  int
	RunCmdId    int
	EditValue   string
	RunUniqueId string
	ReplaceList map[string]string
	StreamMsg   func(string, bool)
}

func NewVariableSet(variableId, runCmdId int, editValue, runUniqueId string, replaceList map[string]string) *VariableSet {
	return &VariableSet{
		RunCmdId:    runCmdId,
		RunUniqueId: runUniqueId,
		EditValue:   editValue,
		VariableId:  variableId,
		ReplaceList: replaceList,
		StreamMsg:   base.Component.TVariable.StreamMsgFuncBySseId(``, runUniqueId),
	}
}

func (h *VariableSet) Set() (_struct.VCmdResult, error) {
	cmd, _ := base.Component.TVariable.CmdInfo(h.RunCmdId)
	form := _struct.VForm{
		VariableId: cast.ToString(h.VariableId),      //脚本ID
		Name:       cast.ToString(cmd[`name`]),       //名称
		Id:         cast.ToString(cmd[`id`]),         //执行的cmd ID
		ResultKey:  cast.ToString(cmd[`result_key`]), //输出的替换key
		CmdType:    cast.ToString(cmd[`type`]),       //cmd 类型

	}
	cmdResult := _struct.VCmdResult{
		VariableId: h.VariableId,
	}
	cmdResult.RunUniqueId = h.RunUniqueId
	vCmd := NewPCmd(``, cmd, h.ReplaceList, h.RunUniqueId)
	switch cast.ToInt(form.CmdType) {
	case define.VariableCmdRadio: //单选
		err := vCmd.ParseSelect(&form)
		if err != nil {
			return cmdResult, errors.New(`解析select失败 ` + err.Error())
		}
		vCmd.StreamMsg(fmt.Sprintf(`%s %s %s %s`,
			base.Component.TMarkDown.Bold(`set`),
			form.Name,
			base.Component.TMarkDown.Bold(`choose：`),
			form.Select.GetSelectOption(h.EditValue).Label), true)
		base.Component.TVariable.SelectChooseReplace(&form, h.ReplaceList, h.EditValue)
	case define.VariableCmdInput, define.VariableCmdTextarea:
		if gstool.SContains(strings.ToLower(form.Name), []string{`php`}) {
			vCmd.StreamMsg(fmt.Sprintf(`%s %s %s`,
				base.Component.TMarkDown.Bold(`set`),
				form.Name,
				base.Component.TMarkDown.Bold(`input：`)), true)
			vCmd.StreamMsg(base.Component.TMarkDown.Code(h.EditValue, `php`), true)
		} else if gstool.SContains(strings.ToLower(form.Name), []string{`sql`}) {
			vCmd.StreamMsg(fmt.Sprintf(`%s %s %s`,
				base.Component.TMarkDown.Bold(`set`),
				form.Name, base.Component.TMarkDown.Bold(`input：`)), true)
			vCmd.StreamMsg(base.Component.TMarkDown.Code(h.EditValue, `sql`), true)
		} else {
			vCmd.StreamMsg(fmt.Sprintf(`%s %s %s %s`,
				base.Component.TMarkDown.Bold(`set`),
				form.Name, base.Component.TMarkDown.Bold(`input：`),
				h.EditValue), true)
		}
		err := vCmd.ParseInput(&form)
		if err != nil {
			return cmdResult, errors.New(`解析input失败 ` + err.Error())
		}
		base.Component.TVariable.AddReplace(h.ReplaceList, form.ResultKey, h.EditValue)
	default:
		cmdResult.RunStatus = define.RunStatusFinish
		return cmdResult, errors.New(`不支持的操作` + form.CmdType)
	}
	cmdResult.ReplaceList = h.ReplaceList
	cmdResult.Form = form
	return cmdResult, nil
}
