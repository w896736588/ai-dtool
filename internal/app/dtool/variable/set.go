package variable

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

type VariableSet struct {
	VariableId  int
	RunCmdId    int
	EditValue   string
	ReplaceList map[string]string
	Sse         *p_sse.SseShell
	Call        *p_common.Call
}

func NewVariableSet(sse *p_sse.SseShell, variableId, runCmdId int, editValue string, replaceList map[string]string, call *p_common.Call) *VariableSet {
	variableSet := &VariableSet{
		RunCmdId:    runCmdId,
		EditValue:   editValue,
		VariableId:  variableId,
		ReplaceList: replaceList,
		Sse:         sse,
		Call:        call,
	}
	return variableSet
}

func (h *VariableSet) Set() (_struct.VCmdResult, error) {
	cmd, _ := h.Call.CmdInfo(h.RunCmdId)
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
	vCmd := NewPCmd(h.Sse, cmd, h.ReplaceList, h.Call)
	switch cast.ToInt(form.CmdType) {
	case define.VariableCmdRadio: //单选
		err := vCmd.ParseSelect(&form)
		if err != nil {
			return cmdResult, errors.New(`解析select失败 ` + err.Error())
		}
		h.Sse.Send(fmt.Sprintf(`%s %s %s %s`,
			p_common.TMarkDownClient.Bold(`set`),
			form.Name,
			p_common.TMarkDownClient.Bold(`choose：`),
			form.Select.GetSelectOption(h.EditValue).Label) + "\n")
		component.VariableClient.SelectChooseReplace(&form, h.ReplaceList, h.EditValue)
	case define.VariableCmdInput, define.VariableCmdTextarea:
		if gstool.SContains(strings.ToLower(form.Name), []string{`php`}) {
			h.Sse.Send(fmt.Sprintf(`%s %s %s`,
				p_common.TMarkDownClient.Bold(`set`),
				form.Name,
				p_common.TMarkDownClient.Bold(`input：`)) + "\n")
			h.Sse.Send(p_common.TMarkDownClient.Code(h.EditValue, `php`) + "\n")
		} else if gstool.SContains(strings.ToLower(form.Name), []string{`sql`}) {
			h.Sse.Send(fmt.Sprintf(`%s %s %s`,
				p_common.TMarkDownClient.Bold(`set`),
				form.Name, p_common.TMarkDownClient.Bold(`input：`)) + "\n")
			h.Sse.Send(p_common.TMarkDownClient.Code(h.EditValue, `sql`) + "\n")
		} else {
			h.Sse.Send(fmt.Sprintf(`%s %s %s %s`,
				p_common.TMarkDownClient.Bold(`set`),
				form.Name, p_common.TMarkDownClient.Bold(`input：`),
				h.EditValue) + "\n")
		}
		err := vCmd.ParseInput(&form)
		if err != nil {
			return cmdResult, errors.New(`解析input失败 ` + err.Error())
		}
		component.VariableClient.AddReplace(h.ReplaceList, form.ResultKey, h.EditValue)
	default:
		cmdResult.RunStatus = define.RunStatusFinish
		return cmdResult, errors.New(`不支持的操作` + form.CmdType)
	}
	cmdResult.ReplaceList = h.ReplaceList
	cmdResult.Form = form
	return cmdResult, nil
}
