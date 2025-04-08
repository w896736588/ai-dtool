package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"strings"
)

// RunPre 执行前收集一些选择或者输入项
func (h *VariableRun) RunPre(variableId any) ([]_struct.VariableForm, []map[string]string, int, error) {
	cmdList, cmdListErr := h.getVariableCmdList(variableId)
	if cmdListErr != nil {
		h.sendStreamMsgMarkdownEnter(cmdListErr.Error())
		return nil, nil, 0, cmdListErr
	}
	replaceList := make([]map[string]string, 0)
	variableFormList := make([]_struct.VariableForm, 0) //需要展示在页面上的和form表单有关联的 只限于is_pre=1的
	isCanRun := 1
	for _, cmd := range cmdList {
		name := cast.ToString(cmd[`name`])
		if cast.ToInt(cmd[`is_pre`]) == 0 { //不需要提前执行
			if cast.ToInt(cmd[`type`]) == define.VariableCmdBash { //预先连接ssh
				id, _ := h.ParseIdContent(cast.ToString(cmd[`bash`]))
				if id == `` {
					return nil, nil, 0, errors.New(`bash脚本中id格式错误`)
				}
				cmd[`ssh_id`] = id
				h.sendStreamMsgMarkdownEnter(name + `->开始检查：` + cast.ToString(cmd[`name`]) + `,预先连接ssh`)
				preConnErr := h.preConnSsh(cmd)
				if preConnErr != nil {
					return nil, nil, 0, preConnErr
				} else {
					h.sendStreamMsgMarkdownEnter(name + `->ssh连接成功`)
				}
			}
			continue
		}

		//初始化
		resultKey := cast.ToString(cmd[`result_key`])
		variableForm := _struct.VariableForm{
			VariableId:   cast.ToString(variableId),
			VariableType: cast.ToString(cmd[`type`]), //类型
			Name:         cast.ToString(cmd[`name`]), //名称
			ResultKey:    resultKey,                  //输出的替换key
			IsShowOk:     0,                          //1准备好在页面上展示 0 未准备好　不决定是否能执行
			IsRunOk:      0,                          //1已经准备好执行 全部为1的时候就可以执行了
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
				h.sendStreamMsgMarkdownEnter(name + `->待输入内容`)
			} else {
				h.addReplace(&replaceList, variableForm.ResultKey, variableForm.Input.Value)
				h.sendStreamMsgMarkdownEnter(name + `->` + variableForm.Input.Value)
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
				h.sendStreamMsgMarkdownEnter(name + `->待其他选项执行完 `)
				isCanRun = 0
				break
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
				MysqlId: id,
			}
			if h.isExistReplaceParam(variableForm.Sql.Sql) {
				h.sendStreamMsgMarkdownEnter(name + `->待其他选项执行完`)
				isCanRun = 0
				break
			}
			h.sendStreamMsgMarkdownEnter(name + `->执行`)
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
	return variableFormList, replaceList, isCanRun, nil
}

// ParseIdContent 解析sql或者bash脚本第一行定义的id，格式：[id=1]
func (h *VariableRun) ParseIdContent(str string) (string, string) {
	sqlParamList := strings.Split(str, "\n")
	id := gstool.StringReplaces(sqlParamList[0], map[string]string{
		`[id=`: ``,
		`]`:    ``,
	})
	return id, gstool.StringReplaces(str, map[string]string{
		sqlParamList[0] + "\n": ``,
	})
}

// PreRadioOptionList 组装单选
func (h *VariableRun) PreRadioOptionList(variableForm *_struct.VariableForm) error {
	if len(variableForm.Select.OptionList) > 0 {
		return nil
	}
	if h.isExistReplaceParam(variableForm.Select.Options) {
		h.sendStreamMsgMarkdownEnter(variableForm.Select.Label + `->等待选择其他选项`)
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
func (h *VariableRun) preConnSsh(cmd map[string]any) error {
	sshId := cast.ToString(cmd[`ssh_id`])
	if sshId == `` {
		return errors.New(`ssh_id不能为空`)
	}
	name := cast.ToString(cmd[`name`])
	sshUniqueKey := Component.TBase.GetCombineKey(`variable`, sshId, `run`)
	sftpUniqueKey := Component.TBase.GetCombineKey(`variable`, sshId, `sftp`)
	if Component.TShell.Exist(sshUniqueKey) && Component.TShell.Exist(sftpUniqueKey) {
		return nil
	}
	h.sendStreamMsgMarkdownEnter(name + `[开始ssh连接]`)
	//初始化连接
	sshConfig, sshConfigErr := Component.TSqlite.GetSshConfig(sshId)
	if sshConfigErr != nil {
		return sshConfigErr
	}
	//ssh
	_, sshClientErr := Component.TShell.GetClientMarkdown(sshConfig, sshUniqueKey, define.SseVariable)
	if sshClientErr != nil {
		return sshClientErr
	}
	//sftp
	_, sftpClientErr := Component.TShell.GetClientMarkdown(sshConfig, sftpUniqueKey, define.SseVariable)
	if sftpClientErr != nil {
		return sftpClientErr
	}
	h.sendStreamMsgMarkdownEnter(name + `[连接成功ssh成功]`)
	return nil
}

// PreShowSet 准备完成的处理
func (h *VariableRun) PreShowSet(variableId, cmdName string, variableForm *_struct.VariableForm) {
	variableForm.IsShowOk = 1 //默认显示是
	switch cast.ToInt(variableForm.VariableType) {
	case define.VariableCmdRadio: //单选
		if h.isExistReplaceParam(variableForm.Select.Options) {
			variableForm.IsShowOk = 0 //不显示
			h.sendStreamMsgMarkdownEnter(cmdName + `->等待补充选项后展示`)
			return
		}
	case define.VariableCmdMysql: //执行sql
		variableForm.IsShowOk = 0
		return
	default:
		break
	}
}
