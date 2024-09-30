package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

// RunPre 执行前收集一些选择或者输入项
func (h *VariableRun) RunPre(variableId any) ([]_struct.VariableForm, []map[string]string, int, error) {
	cmdList, cmdListErr := h.getVariableCmdList(variableId)
	if cmdListErr != nil {
		h.sendSocketMsg(variableId, cmdListErr.Error())
		return nil, nil, 0, cmdListErr
	}
	replaceList := make([]map[string]string, 0)
	variableFormList := make([]_struct.VariableForm, 0)
	for _, cmd := range cmdList {
		if cast.ToInt(cmd[`is_pre`]) == 0 {
			if cast.ToInt(cmd[`type`]) == define.VariableTypeBash { //预先连接ssh
				h.sendSocketMsg(variableId, `开始检查：`+cast.ToString(cmd[`name`])+`,预先连接ssh`)
				preConnErr := h.preConnSsh(cmd)
				if preConnErr != nil {
					return nil, nil, 0, preConnErr
				}
			}
			continue
		}
		//初始化
		resultKey := cast.ToString(cmd[`result_key`])
		variableForm := _struct.VariableForm{
			VariableId:   cast.ToString(variableId),
			VariableType: cast.ToString(cmd[`type`]), //类型
			ResultKey:    resultKey,                  //输出的替换key
			IsShowOk:     0,                          //1准备好在页面上展示 0 未准备好　不决定是否能执行
			IsRunOk:      0,                          //1已经准备好执行 全部为1的时候就可以执行了
		}
		switch cast.ToInt(cmd[`type`]) {
		case define.VariableTypeInput: //输入框肯定需要进行输入
			variableForm.Input = _struct.VariableFormInput{
				Label: cast.ToString(cmd[`name`]),
			}
			h.PreShowSet(cast.ToString(variableId), cast.ToString(cmd[`name`]), &variableForm)
			break
		case define.VariableTypeRadio: //单项选择 初始的时候不存在替换值 只有选了以后才有
			variableForm.Select = _struct.VariableFormSelect{
				Label:      cast.ToString(cmd[`name`]),
				Value:      ``,
				OptionList: make([]_struct.VariableFormOption, 0),
				Options:    cast.ToString(cmd[`options`]), //原本的字符串选项集
			}
			if !h.isPreShowForm(variableForm.Select.Options) {
				break
			}
			radioErr := h.PreRadioOptionList(&variableForm)
			if radioErr != nil {
				return nil, nil, 0, radioErr
			}
			h.PreShowSet(cast.ToString(variableId), cast.ToString(cmd[`name`]), &variableForm)
			break
		case define.VariableTypeRedisChoose: //redis选择 所有配置的redis
			variableForm.Select = _struct.VariableFormSelect{
				Label:      cast.ToString(cmd[`name`]),
				Value:      ``,
				OptionList: make([]_struct.VariableFormOption, 0),
				Options:    cast.ToString(cmd[`options`]), //原本的字符串选项集
			}
			configList, configListErr := Component.TSqlite.GetAllRedisConfig()
			if configListErr != nil {
				return nil, nil, 0, configListErr
			}
			variableForm.Select.OptionList = make([]_struct.VariableFormOption, 0)
			for _, redisConfig := range configList {
				variableForm.Select.OptionList = append(variableForm.Select.OptionList, _struct.VariableFormOption{
					Label:  cast.ToString(redisConfig[`name`]),
					Value:  cast.ToString(redisConfig[`id`]),
					Source: gstool.JsonEncode(redisConfig),
				})
			}
			h.PreShowSet(cast.ToString(variableId), cast.ToString(cmd[`name`]), &variableForm)
			break
		case define.VariableTypeMysql: //执行sql 初始化
			variableForm.Sql = _struct.VariableFormSql{
				Sql:     cast.ToString(cmd[`sql`]),
				MysqlId: cast.ToString(cmd[`mysql_id`]),
			}
			if h.isPreShowForm(variableForm.Sql.Sql) {
				sqlRet := h.sqlProcessRun(&variableForm, &replaceList)
				if sqlRet != nil {
					return nil, nil, 0, sqlRet
				}
			}
			h.PreShowSet(cast.ToString(variableId), cast.ToString(cmd[`name`]), &variableForm)
			break
		default:
			//这里不管预执行
			break
		}
		variableFormList = append(variableFormList, variableForm)
	}
	return variableFormList, replaceList, 0, nil
}

// PreRadioOptionList 组装单选
func (h *VariableRun) PreRadioOptionList(variableForm *_struct.VariableForm) error {
	if len(variableForm.Select.OptionList) > 0 {
		return nil
	}
	h.sendSocketMsg(variableForm.VariableId, variableForm.Select.Label+`,准备处理单选`)
	if !h.isPreShowForm(variableForm.Select.Options) {
		h.sendSocketMsg(variableForm.VariableId, variableForm.Select.Label+`,内容尚未替换，等待选择其他选项`)
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
	h.sendSocketMsg(variableForm.VariableId, `处理单选项完成`)
	return nil
}

// preConnSsh 初始化ssh连接
func (h *VariableRun) preConnSsh(cmd map[string]any) error {
	sshId := cast.ToString(cmd[`ssh_id`])
	if sshId == `` {
		return errors.New(`ssh_id不能为空`)
	}
	sshUniqueKey := Component.TBase.GetCombineKey(`variable`, sshId, `run`)
	sftpUniqueKey := Component.TBase.GetCombineKey(`variable`, sshId, `sftp`)
	if Component.TShell.Exist(sshUniqueKey) && Component.TShell.Exist(sftpUniqueKey) {
		return nil
	}
	h.sendSocketMsg(h.VariableId, `初始化ssh连接(`+cast.ToString(cmd[`ssh_id`])+`)开始`)
	//初始化连接
	sshConfig, sshConfigErr := Component.TSqlite.GetSshConfig(sshId)
	if sshConfigErr != nil {
		return sshConfigErr
	}
	//ssh
	sshClient, sshClientErr := Component.TShell.GetClient(sshConfig, sshUniqueKey)
	if sshClientErr != nil {
		return sshClientErr
	}
	sshClient.SetSocket(h.getSocket(h.VariableId))
	//sftp
	sftpClient, sftpClientErr := Component.TShell.GetClient(sshConfig, sftpUniqueKey)
	if sftpClientErr != nil {
		return sftpClientErr
	}
	sftpClient.SetSocket(h.getSocket(h.VariableId))
	h.sendSocketMsg(h.VariableId, `初始化ssh连接(`+cast.ToString(cmd[`ssh_id`])+`)成功`)
	return nil
}

// PreShowSet 准备完成的处理
func (h *VariableRun) PreShowSet(variableId, variableCmdName string, variableForm *_struct.VariableForm) {
	variableForm.IsShowOk = 1 //默认显示是
	switch cast.ToInt(variableForm.VariableType) {
	case define.VariableTypeRadio: //单选
		if !h.isPreShowForm(variableForm.Select.Options) {
			variableForm.IsShowOk = 0 //不显示
			h.sendSocketMsg(variableId, `开始检查：`+variableCmdName+`,等待补充选项后展示`)
			return
		}
		h.sendSocketMsg(variableId, `开始检查：`+variableCmdName+`,可以展示`)
	case define.VariableTypeMysql: //执行sql
		h.sendSocketMsg(variableId, `开始检查：`+variableCmdName+`,初始化完成`)
		variableForm.IsShowOk = 0
		return
	default:
		h.sendSocketMsg(variableId, `开始检查：`+variableCmdName+`,可以展示`)
		break
	}
}
