package p_variable

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"github.com/spf13/cast"
)

type Variable struct {
	RunCmdId    int                 //当前运行的cmd
	VariableId  int                 //脚本ID
	RunUniqueId string              //本次执行唯一ID
	ReplaceList []map[string]string //替换列表
	IsRun       int                 //最终执行1 最终执行
	StreamMsg   func(string, bool)  //输出方法
}

func NewVariable(variableId, runCmdId int, isRun int, replaceList []map[string]string, streamMsg func(string, bool)) *Variable {
	return &Variable{
		VariableId:  variableId,
		RunCmdId:    runCmdId,
		ReplaceList: replaceList,
		IsRun:       isRun,
		RunUniqueId: base.Component.TBase.GetUnique(`variable`),
		StreamMsg:   streamMsg,
	}
}

func (h *Variable) Run() (_struct.VariableCmdResult, error) {
	//其他任务处理
	base.Component.TVariable.StopAll()
	base.Component.TVariable.Add(h.RunUniqueId)
	//开始
	cmdResult := _struct.VariableCmdResult{}
	cmdList, cmdErr := base.Component.TVariable.CmdList(h.VariableId)
	if cmdErr != nil {
		return cmdResult, cmdErr
	}
	//当前执行的cmd
	cmdInfo, _ := base.Component.TVariable.CmdInfo(h.RunCmdId)
	runWeight := cast.ToInt(cmdInfo[`weight`])
	for _, cmd := range cmdList {
		name := cast.ToString(cmd[`name`])
		cmdId := cast.ToString(cmd[`id`])
		weight := cast.ToInt(cmd[`weight`])
		//最终执行时 需要从当前cmd开始执行
		//非最终执行时，从下一个开始执行
		if h.IsRun == 1 {
			if weight < runWeight {
				continue
			}
		} else {
			if weight <= runWeight {
				continue
			}
		}
		//替换
		h.Replace(cmd)
		//是否需要执行
		if !base.Component.TVariable.ChecksCanDo(cmd) {
			h.StreamMsg(fmt.Sprintf(`不需要执行（%s）,checks:%s`, name, cmd[`checks`]), true)
			continue
		}
		cmdType := cast.ToInt(cmd[`type`])
		runType := cast.ToString(cmd[`run_type`])
		//非最终执行并且等待客户点击运行
		if h.IsRun != 1 && runType == define.RunTypeRun {
			h.StreamMsg(fmt.Sprintf(`可以执行（%s）`, name), true)
			cmdResult.ReplaceList = h.ReplaceList
			cmdResult.Form = _struct.VariableForm{Id: cmdId}
			cmdResult.RunStatus = define.RunStatusCanRun
			return cmdResult, nil
		}
		var err error
		switch cmdType {
		case define.VariableCmdInput, define.VariableCmdTextarea, define.VariableCmdRadio:
			cmdResult.Form, err = h.BuildCmd(cmd)
			cmdResult.ReplaceList = h.ReplaceList
			if err != nil {
				return cmdResult, err
			}
		case define.VariableCmdMysql, define.VariableCmdBash, define.VariableCmdCommand,
			define.VariableCmdRedis, define.VariableCmdCurl, define.VariableCmdPlaywright,
			define.VariableCmdCombine:
			runErr := h.RunCmd(cmd)
			cmdResult.ReplaceList = h.ReplaceList
			if runErr != nil {
				return cmdResult, runErr
			}
		default:
			return cmdResult, errors.New(fmt.Sprintf(`不支持的类型%s（%s）`, cast.ToString(cmd[`type`]), name))
		}
		//开始执行 那么一直到底
		if h.IsRun == 1 {
			continue
		}
		//输出表单
		if runType == define.RunTypeForm {
			cmdResult.RunStatus = define.RunStatusWaitRun
			h.StreamMsg(fmt.Sprintf(`输出表单（%s）`, name), true)
			return cmdResult, nil
		}
		//中间层
		if runType == define.RunTypeMiddle {
			continue
		}
	}
	//执行结束
	h.StreamMsg(`执行结束`, true)
	cmdResult.RunStatus = define.RunStatusFinish
	return cmdResult, nil
}

func (h *Variable) RunCmd(cmd map[string]any) error {
	//替换
	h.Replace(cmd)
	//执行
	rCmd := NewRCmd(cmd, &h.ReplaceList, h.StreamMsg)
	var err error
	switch cast.ToInt(cmd[`type`]) {
	case define.VariableCmdMysql:
		err = rCmd.RunMysql()
	case define.VariableCmdBash:
		_, err = rCmd.RunBash()
	case define.VariableCmdCommand:
		_, err = rCmd.RunCommand()
	case define.VariableCmdRedis:
		_, err = rCmd.RunRedis()
	case define.VariableCmdCurl:
		_, err = rCmd.RunCurl()
	case define.VariableCmdCombine:
		_, err = rCmd.RunCombine()
	case define.VariableCmdPlaywright:
		_, err = rCmd.RunPlaywright()
	default:
		return errors.New(`不支持的类型` + cast.ToString(cmd[`type`]))
	}
	return err
}

// BuildCmd 构建cmd表单
func (h *Variable) BuildCmd(cmd map[string]any) (_struct.VariableForm, error) {
	variableForm := _struct.VariableForm{
		VariableId: cast.ToString(h.VariableId),      //脚本ID
		Name:       cast.ToString(cmd[`name`]),       //名称
		Id:         cast.ToString(cmd[`id`]),         //执行的cmd ID
		ResultKey:  cast.ToString(cmd[`result_key`]), //输出的替换key
		CmdType:    cast.ToString(cmd[`type`]),       //cmd 类型
	}
	//执行
	vCmd := NewPCmd(cmd, &h.ReplaceList, h.StreamMsg)
	var err error
	switch cast.ToInt(cmd[`type`]) {
	case define.VariableCmdInput, define.VariableCmdTextarea:
		variableForm.Input, err = vCmd.ParseInput()
	case define.VariableCmdRadio:
		variableForm.Select, err = vCmd.ParseSelect()
	default:
		err = errors.New(`没有处理的类型`)
	}
	if err != nil {
		return variableForm, err
	}
	return variableForm, nil
}

func (h *Variable) Replace(cmd map[string]any) {
	cmd[`options`] = base.Component.TVariable.Replace(cast.ToString(cmd[`options`]), &h.ReplaceList)
	cmd[`checks`] = base.Component.TVariable.Replace(cast.ToString(cmd[`checks`]), &h.ReplaceList)
}

func (h *Variable) input(cmd map[string]any, variableForm *_struct.VariableForm) {
	variableForm.Input = _struct.VariableFormInput{
		Label: cast.ToString(cmd[`name`]),
		Value: cast.ToString(cmd[`default`]),
	}
}
