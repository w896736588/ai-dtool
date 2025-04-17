package p_variable

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"strings"
)

// RunPre 执行前收集一些选择或者输入项
func (h *VariableRun) RunPre(variableId any) ([]_struct.VariableForm, []map[string]string, int, error) {
	base.Component.TVariable.StopAll()          //停止其他任务
	base.Component.TVariable.Add(h.RunUniqueId) //注册本次任务
	variableInfo := h.Variable(variableId)
	if len(variableInfo) == 0 {
		return nil, nil, 0, errors.New(`脚本不存在`)
	}
	cmdList, cmdListErr := h.CmdList(variableId)
	if cmdListErr != nil {
		h.StreamMsg(cmdListErr.Error(), true)
		return nil, nil, 0, cmdListErr
	}
	replaceList := make([]map[string]string, 0)
	variableFormList := make([]_struct.VariableForm, 0) //需要展示在页面上的和form表单有关联的 只限于is_pre=1的
	isCanRun := 1
	for _, cmd := range cmdList {
		if base.Component.TVariable.Get(h.RunUniqueId) == `stop` {
			return nil, nil, 0, errors.New(`任务停止`)
		}
		if cast.ToInt(cmd[`is_pre`]) == 0 {
			continue
		}

		//初始化
		resultKey := cast.ToString(cmd[`result_key`])
		variableForm := _struct.VariableForm{
			VariableId:   cast.ToString(variableId),
			VariableType: cast.ToString(cmd[`type`]), //类型
			Name:         cast.ToString(cmd[`name`]), //名称
			Id:           cast.ToString(cmd[`id`]),
			ResultKey:    resultKey, //输出的替换key
			IsShowOk:     0,         //1准备好在页面上展示 0 未准备好　不决定是否能执行
			IsRunOk:      0,         //1已经准备好执行 全部为1的时候就可以执行了
		}
		switch cast.ToInt(cmd[`type`]) {
		case define.VariableCmdInput, define.VariableCmdTextarea: //输入框肯定需要进行输入
			variableForm.Input = _struct.VariableFormInput{
				Label: cast.ToString(cmd[`name`]),
				Value: cast.ToString(cmd[`default`]),
			}
			h.PreShowSet(cast.ToString(variableId), cast.ToString(cmd[`name`]), &variableForm)
			if cast.ToString(cmd[`default`]) == `` {
				isCanRun = 0
			} else {
				h.addReplace(&replaceList, variableForm.ResultKey, variableForm.Input.Value)
			}
			break
		case define.VariableCmdRadio: //单项选择 初始的时候不存在替换值 只有选了以后才有
			variableForm.Select = _struct.VariableFormSelect{
				Label:      cast.ToString(cmd[`name`]),
				Value:      ``,
				OptionList: make([]_struct.VariableFormOption, 0),
				Options:    cast.ToString(cmd[`options`]), //原本的字符串选项集
			}
			if h.isExistReplaceParam(variableForm.Select.Options) {
				isCanRun = 0
				break
			}
			if h.isExistConfigParam(variableForm.Select.Options) {
				variableForm.Select.Options = h.ParseConfig(variableForm.Select.Options)
			}
			radioErr := h.PreRadioOptionList(&variableForm)
			if radioErr != nil {
				return nil, nil, 0, radioErr
			}
			h.PreShowSet(cast.ToString(variableId), cast.ToString(cmd[`name`]), &variableForm)
			break
		case define.VariableCmdMysql: //执行sql 初始化
			id, sql := h.ParseIdContent(cast.ToString(cmd[`sql`]))
			variableForm.Sql = _struct.VariableFormSql{
				Sql:     sql,
				MysqlId: cast.ToString(id),
			}
			if h.isExistReplaceParam(variableForm.Sql.Sql) {
				isCanRun = 0
				break
			}
			isCanRun = 0
			sqlRet := h.sqlProcessRun(&variableForm, &replaceList)
			if sqlRet != nil {
				return nil, nil, 0, sqlRet
			}
			h.PreShowSet(cast.ToString(variableId), cast.ToString(cmd[`name`]), &variableForm)
			break
		default:
			//这里不管预执行
			break
		}
		variableFormList = append(variableFormList, variableForm)
	}
	h.StreamMsg(base.Component.TMarkDown.BlockQuote(cast.ToString(variableInfo[`name`])+`就绪`), true)
	return variableFormList, replaceList, isCanRun, nil
}

// ParseConfig 自带的配置查询解析
func (h *VariableRun) ParseConfig(config string) string {
	if gstool.SContains(config, []string{`{config:ssh}`}) {
		sshList, sshErr := base.Component.TSqlite.GetAllSshConfig()
		if sshErr == nil {
			for k, sshMap := range sshList {
				sshMap[`value`] = sshMap[`id`]
				sshMap[`label`] = sshMap[`name`]
				sshList[k] = sshMap
			}
			return gstool.JsonEncode(sshList)
		}
	}
	return `[]`
}

// ParseIdContent 解析sql或者bash脚本第一行定义的id，格式：[RunUniqueId=1]
func (h *VariableRun) ParseIdContent(str string) (int, string) {
	sqlParamList := strings.Split(str, "\n")
	bashContent := gstool.SReplaces(str, map[string]string{
		sqlParamList[0] + "\n": ``,
	})
	baseId := sqlParamList[0]
	id := gstool.SReplaces(baseId, map[string]string{
		`[id=`: ``,
		`]`:    ``,
	})
	sshId := cast.ToInt(id)
	return sshId, bashContent
}

// PreRadioOptionList 组装单选
func (h *VariableRun) PreRadioOptionList(variableForm *_struct.VariableForm) error {
	if len(variableForm.Select.OptionList) > 0 {
		return nil
	}
	if h.isExistReplaceParam(variableForm.Select.Options) {
		return nil
	}

	//组装选项
	optionSourceList := make([]map[string]any, 0)
	//原本的选项值
	decodeErr := gstool.JsonDecode(variableForm.Select.Options, &optionSourceList)
	if decodeErr != nil {
		return decodeErr
	}
	optionList := make([]_struct.VariableFormOption, 0)
	for _, optionMap := range optionSourceList {
		option := _struct.VariableFormOption{
			Label:  cast.ToString(optionMap[`label`]),
			Value:  cast.ToString(optionMap[`value`]),
			Source: gstool.JsonEncode(optionMap),
		}
		optionList = append(optionList, option)
	}
	variableForm.Select.OptionList = optionList
	return nil
}

// preConnSsh 初始化ssh连接
func (h *VariableRun) preConnSsh(sshId int) error {
	if sshId == 0 {
		return errors.New(`ssh_id不能为空`)
	}
	sshUniqueKey := base.Component.TBase.GetCombineKey(`variable`, sshId, `run`)
	sftpUniqueKey := base.Component.TBase.GetCombineKey(`variable`, sshId, `sftp`)
	if base.Component.TShell.Exist(sshUniqueKey) && base.Component.TShell.Exist(sftpUniqueKey) {
		return nil
	}
	//初始化连接
	sshConfig, sshConfigErr := base.Component.TSqlite.GetSshConfig(sshId)
	if sshConfigErr != nil {
		return sshConfigErr
	}
	//ssh
	_, sshClientErr := base.Component.TShell.GetClientMarkdown(sshConfig, sshUniqueKey, define.SseVariable)
	if sshClientErr != nil {
		return sshClientErr
	}
	//sftp
	_, sftpClientErr := base.Component.TShell.GetClientMarkdown(sshConfig, sftpUniqueKey, define.SseVariable)
	if sftpClientErr != nil {
		return sftpClientErr
	}
	return nil
}

// PreShowSet 准备完成的处理
func (h *VariableRun) PreShowSet(variableId, cmdName string, variableForm *_struct.VariableForm) {
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
