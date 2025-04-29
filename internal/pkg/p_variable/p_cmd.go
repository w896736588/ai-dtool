package p_variable

import (
	"dev_tool/base"
	_struct "dev_tool/base/struct"
	"errors"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

type PCmd struct {
	cmd         map[string]any
	replaceList *[]map[string]string
	StreamMsg   func(string, bool)
	RunUniqueId string
}

func NewPCmd(cmd map[string]any, replace *[]map[string]string, runUniqueId string) *PCmd {
	return &PCmd{cmd: cmd, replaceList: replace, RunUniqueId: runUniqueId, StreamMsg: base.Component.TVariable.StreamMsgFunc(runUniqueId)}
}

func (h *PCmd) ParseInput() (_struct.VariableFormInput, error) {
	return _struct.VariableFormInput{
		Label: cast.ToString(h.cmd[`name`]),
		Value: cast.ToString(h.cmd[`default`]),
	}, nil
}

func (h *PCmd) ParseSelect() (_struct.VariableFormSelect, error) {
	form := _struct.VariableFormSelect{
		Label:      cast.ToString(h.cmd[`name`]),
		Value:      ``,
		OptionList: make([]_struct.VariableFormOption, 0),
		Options:    cast.ToString(h.cmd[`options`]), //原本的字符串选项集 json字符串
	}
	//解析配置
	if base.Component.TVariable.ExistConfigParamFull(form.Options) {
		var parseErr error
		form.Options, parseErr = base.Component.TVariable.ParseConfig(form.Options)
		if parseErr != nil {
			return form, parseErr
		}
	}
	//整体替换选项
	form.Options = base.Component.TVariable.Replace(form.Options, h.replaceList)
	//开始判断
	if base.Component.TVariable.ExistReplaceParamFull(form.Options) {
		return form, errors.New(`存在未进行替换的内容：` + form.Options)
	}
	//组装选项
	optionSourceList := make([]map[string]any, 0)
	decodeErr := gstool.JsonDecode(form.Options, &optionSourceList)
	if decodeErr != nil {
		return form, errors.New(`解析` + form.Options + ` 失败：` + decodeErr.Error())
	}
	for _, optionMap := range optionSourceList {
		form.OptionList = append(form.OptionList, _struct.VariableFormOption{
			Label:  cast.ToString(optionMap[`label`]),
			Value:  cast.ToString(optionMap[`value`]),
			Source: gstool.JsonEncode(optionMap),
		})
	}
	return form, nil
}

// SelectChooseReplace 单选选中后替换
func (h *PCmd) SelectChooseReplace(variableForm *_struct.VariableForm, replaceList *[]map[string]string, chooseValue string) error {
	for _, option := range variableForm.Select.OptionList {
		if variableForm.ResultKey != `` && chooseValue != `` && chooseValue == option.Value {
			sourceOptionList := make(map[string]any, 0)
			_ = gstool.JsonDecode(option.Source, &sourceOptionList)
			for optionKey, optionValue := range sourceOptionList {
				base.Component.TVariable.AddReplace(replaceList, variableForm.ResultKey+`.`+optionKey, cast.ToString(optionValue))
			}
			//替换整体
			base.Component.TVariable.AddReplace(replaceList, variableForm.ResultKey, option.Source)
		}
	}
	return nil
}
