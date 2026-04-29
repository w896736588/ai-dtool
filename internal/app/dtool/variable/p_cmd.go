package variable

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_sse"
	"errors"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

type PCmd struct {
	cmd         map[string]any
	replaceList map[string]string
	Sse         *p_sse.SseShell
	Call        *p_common.Call
}

func NewPCmd(sse *p_sse.SseShell, cmd map[string]any, replace map[string]string, call *p_common.Call) *PCmd {
	return &PCmd{cmd: cmd, replaceList: replace, Sse: sse, Call: call}
}

func (h *PCmd) ParseInput(form *_struct.VForm) error {
	form.Input = _struct.VFormInput{
		Label: cast.ToString(h.cmd[`name`]),
		Value: cast.ToString(h.cmd[`default`]),
	}
	return nil
}

func (h *PCmd) ParseSelect(form *_struct.VForm) error {
	form.Select = _struct.VFormSelect{
		Label:      cast.ToString(h.cmd[`name`]),
		Value:      ``,
		OptionList: make([]_struct.VFormOption, 0),
		Options:    cast.ToString(h.cmd[`options`]), //原本的字符串选项集 json字符串
	}
	//整体替换选项
	form.Select.Options = p_common.Replace(form.Select.Options, h.replaceList)
	//解析配置
	var parseErr error
	form.Select.Options, parseErr = component.VariableClient.ParseConfig(form.Select.Options, h.Call)
	if parseErr != nil {
		return parseErr
	}
	//开始判断
	if component.VariableClient.ExistReplaceParamFull(form.Select.Options) {
		return errors.New(`存在未进行替换的内容：` + form.Select.Options)
	}
	//组装选项
	optionSourceList := make([]map[string]any, 0)
	decodeErr := gstool.JsonDecode(form.Select.Options, &optionSourceList)
	if decodeErr != nil {
		return errors.New(`解析` + form.Select.Options + ` 失败：` + decodeErr.Error())
	}
	for _, optionMap := range optionSourceList {
		form.Select.OptionList = append(form.Select.OptionList, _struct.VFormOption{
			Label:  cast.ToString(optionMap[`label`]),
			Value:  cast.ToString(optionMap[`value`]),
			Source: gstool.JsonEncode(optionMap),
		})
	}
	return nil
}

// SelectChooseReplace 单选选中后替换
func (h *PCmd) SelectChooseReplace(variableForm *_struct.VForm, replaceList map[string]string, chooseValue string) error {
	for _, option := range variableForm.Select.OptionList {
		if variableForm.ResultKey != `` && chooseValue != `` && chooseValue == option.Value {
			sourceOptionList := make(map[string]any, 0)
			_ = gstool.JsonDecode(option.Source, &sourceOptionList)
			for optionKey, optionValue := range sourceOptionList {
				component.VariableClient.AddReplace(replaceList, variableForm.ResultKey+`.`+optionKey, cast.ToString(optionValue))
			}
			//替换整体
			component.VariableClient.AddReplace(replaceList, variableForm.ResultKey, option.Source)
		}
	}
	return nil
}
