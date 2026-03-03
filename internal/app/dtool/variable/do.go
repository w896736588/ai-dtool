package variable

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"

	"github.com/spf13/cast"
)

type Variable struct {
	RunCmdId    int               //当前运行的cmd
	VariableId  int               //脚本ID
	ReplaceList map[string]string //替换列表
	TaskId      string            //最终执行时才会有的taskId
	Sse         *p_sse.SseShell   //流式输出方法
	Call        *p_common.Call
}

func NewVariable(sse *p_sse.SseShell, variableId, runCmdId int, taskId string, replaceList map[string]string, call *p_common.Call) *Variable {
	variable := &Variable{
		VariableId:  variableId,
		RunCmdId:    runCmdId,
		ReplaceList: replaceList,
		TaskId:      taskId,
		Sse:         sse,
		Call:        call,
	}
	if variable.RunCmdId == 0 {
		variable.InitRunUniqueId()
	}

	return variable
}

func (h *Variable) InitRunUniqueId() {
	//清除服务端所有的消息
	h.Sse.CleanMsg()
	//清除前端所有的消息
	h.Sse.Send(define.SseEventClean)
}

func (h *Variable) Run() (_struct.VCmdResult, error) {
	//注入全局替换
	VariableClient.RegisterAllGlobal(h.ReplaceList, &p_sse.SseShell{
		Sse:             h.Sse.Sse,
		SseDistributeId: h.Sse.SseDistributeId,
	}, h.Call)
	//初始化结果
	cmdResult := _struct.VCmdResult{
		VariableId: h.VariableId,
	}
	cmdList, cmdErr := common.DbMain.CmdList(h.VariableId)
	if cmdErr != nil {
		return cmdResult, cmdErr
	}
	//当前执行的cmd
	cmdInfo, _ := common.DbMain.CmdInfo(h.RunCmdId)
	runWeight := cast.ToInt(cmdInfo[`weight`])
	havePlaywright := false //如果有自定义链接 那么不输出end
	for _, cmd := range cmdList {
		if VariableClient.IsStop(h.TaskId) {
			return cmdResult, errors.New(`任务被取消`)
		}
		name := cast.ToString(cmd[`name`])
		cmdId := cast.ToString(cmd[`id`])
		weight := cast.ToInt(cmd[`weight`])
		//最终执行时 需要从当前cmd开始执行
		//非最终执行时，从下一个开始执行
		if h.TaskId != `` {
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
		if !VariableClient.ChecksCanDo(cmd) {
			h.Sse.Send(fmt.Sprintf(`%s %s %s %s`, p_common.TMarkDownClient.Bold(`check`), name, p_common.TMarkDownClient.Bold(`not run：`), cmd[`checks`]) + "\n")
			continue
		}
		cmdType := cast.ToInt(cmd[`type`])
		runType := cast.ToString(cmd[`run_type`])
		//非最终执行并且等待客户点击运行
		if h.TaskId == `` && runType == define.RunTypeRun {
			h.Sse.Send(fmt.Sprintf(`%s %s`, p_common.TMarkDownClient.Bold(`wait run 请点击执行`), name) + "\n")
			cmdResult.ReplaceList = h.ReplaceList
			cmdResult.Form = _struct.VForm{Id: cmdId}
			cmdResult.RunStatus = define.RunStatusCanRun
			return cmdResult, nil
		}
		if cmdType == define.VariableCmdPlaywright {
			havePlaywright = true
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
			define.VariableCmdCombine, define.VariableCmdWindowCommand, define.VariableCmdUpload,
			define.VariableCmdLlm:
			runErr := h.RunCmd(cmd)
			cmdResult.ReplaceList = h.ReplaceList
			if runErr != nil {
				return cmdResult, runErr
			}
		default:
			return cmdResult, errors.New(fmt.Sprintf(`不支持的类型%s（%s）`, cast.ToString(cmd[`type`]), name))
		}
		//开始执行 那么一直到底
		if h.TaskId != `` {
			continue
		}
		//输出表单
		if runType == define.RunTypeForm {
			cmdResult.RunStatus = define.RunStatusWaitRun
			h.Sse.Send(p_common.TMarkDownClient.Bold(`form`) + " " + name + "\n")
			return cmdResult, nil
		}
		//中间层
		if runType == define.RunTypeMiddle {
			continue
		}
	}
	//执行结束
	if !havePlaywright {
		h.Sse.Send(p_common.TMarkDownClient.Bold(`end.`) + "\n")
	}
	cmdResult.RunStatus = define.RunStatusFinish
	return cmdResult, nil
}

func (h *Variable) RunCmd(cmd map[string]any) error {
	//执行
	rCmd := NewRCmd(cmd, h.ReplaceList, h.Sse, h.Call)
	var err error
	switch cast.ToInt(cmd[`type`]) {
	case define.VariableCmdMysql:
		err = rCmd.RunMysql()
	case define.VariableCmdUpload:
		_, err = rCmd.RunUpload()
	case define.VariableCmdBash:
		_, err = rCmd.RunBash()
	case define.VariableCmdWindowCommand:
		_, err = rCmd.RunWindowsCmd()
	case define.VariableCmdCommand:
		_, err = rCmd.RunCommand()
	case define.VariableCmdRedis:
		_, err = rCmd.RunRedis()
	case define.VariableCmdCurl:
		_, err = rCmd.RunCurl()
	case define.VariableCmdCombine:
		_, err = rCmd.RunCombine()
	case define.VariableCmdPlaywright:
		_, err = rCmd.RunPlaywright(func() bool {
			return VariableClient.IsStop(h.TaskId)
		})
	case define.VariableCmdLlm:
		_, err = rCmd.RunLlm()
	default:
		return errors.New(`不支持的类型` + cast.ToString(cmd[`type`]))
	}
	return err
}

// BuildCmd 构建cmd表单
func (h *Variable) BuildCmd(cmd map[string]any) (_struct.VForm, error) {
	form := _struct.VForm{
		VariableId: cast.ToString(h.VariableId),      //脚本ID
		Name:       cast.ToString(cmd[`name`]),       //名称
		Id:         cast.ToString(cmd[`id`]),         //执行的cmd ID
		ResultKey:  cast.ToString(cmd[`result_key`]), //输出的替换key
		CmdType:    cast.ToString(cmd[`type`]),       //cmd 类型
	}
	//执行
	vCmd := NewPCmd(h.Sse, cmd, h.ReplaceList, h.Call)
	var err error
	switch cast.ToInt(cmd[`type`]) {
	case define.VariableCmdInput, define.VariableCmdTextarea:
		err = vCmd.ParseInput(&form)
	case define.VariableCmdRadio:
		err = vCmd.ParseSelect(&form)
	default:
		err = errors.New(`没有处理的类型`)
	}
	if err != nil {
		return form, err
	}
	return form, nil
}

func (h *Variable) Replace(cmd map[string]any) {
	//curl类型的不进行替换，因为会出现一些\n 如果json_decode会报错
	if cast.ToInt(cmd[`type`]) != define.VariableCmdCurl {
		cmd[`options`] = p_common.Replace(cast.ToString(cmd[`options`]), h.ReplaceList)
	}
	cmd[`checks`] = p_common.Replace(cast.ToString(cmd[`checks`]), h.ReplaceList)
}

func (h *Variable) input(cmd map[string]any, variableForm *_struct.VForm) {
	variableForm.Input = _struct.VFormInput{
		Label: cast.ToString(cmd[`name`]),
		Value: cast.ToString(cmd[`default`]),
	}
}
