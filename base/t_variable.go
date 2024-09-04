package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

type VariableRun struct {
	VariableId      string
	ReplaceList     []map[string]string
	sshClientList   map[string]*gsssh.SshConfig
	sftpSessionList map[string]*gsssh.SshConfig
}

func NewVariable() VariableRun {
	return VariableRun{
		sshClientList:   make(map[string]*gsssh.SshConfig),
		sftpSessionList: make(map[string]*gsssh.SshConfig),
	}
}

// RunPre 执行前收集一些选择或者输入项
func (h *VariableRun) RunPre(variableId any) ([]_struct.VariableForm, []map[string]string, int, error) {
	cmdList, cmdListErr := h.getVariableCmdList(variableId)
	if cmdListErr != nil {
		return nil, nil, 0, cmdListErr
	}
	//输出的表单
	waitPreNum := 0
	replaceList := make([]map[string]string, 0)
	variableFormList := make([]_struct.VariableForm, 0)
	for _, cmd := range cmdList {
		if cast.ToInt(cmd[`is_pre`]) == 0 {
			continue
		}
		//初始化
		resultKey := cast.ToString(cmd[`result_key`])
		variableForm := _struct.VariableForm{
			VariableType: cast.ToString(cmd[`type`]), //类型
			ResultKey:    resultKey,                  //输出的替换key
			IsPreOk:      0,                          //未准备好
		}
		switch cast.ToInt(cmd[`type`]) {
		case define.VariableTypeInput: //输入框肯定需要进行输入
			variableForm.Input = _struct.VariableFormInput{
				Label: cast.ToString(cmd[`name`]),
			}
			waitPreNum++
			break
		case define.VariableTypeRadio: //单项选择 初始的时候不存在替换值 只有选了以后才有
			variableForm.Select = _struct.VariableFormSelect{
				Label:      cast.ToString(cmd[`name`]),
				Value:      ``,
				OptionList: make([]_struct.VariableFormOption, 0),
				Options:    cast.ToString(cmd[`options`]), //原本的字符串选项集
			}
			if !h.isPreShowForm(variableForm.Select.Options) {
				waitPreNum++
				break
			}
			radioErr := h.radioPreProcess(&variableForm, &replaceList, variableForm.Select.Value)
			if radioErr != nil {
				return nil, nil, 0, radioErr
			}
			break
		case define.VariableTypeMysql: //执行sql
			variableForm.Sql = _struct.VariableFormSql{
				Sql:     cast.ToString(cmd[`sql`]),
				MysqlId: cast.ToString(cmd[`mysql_id`]),
			}
			if h.isPreShowForm(variableForm.Sql.Sql) {
				sqlRet := h.sqlPreProcess(&variableForm, &replaceList)
				if sqlRet != nil {
					return nil, nil, 0, sqlRet
				}
				variableForm.IsPreOk = 1
			} else {
				waitPreNum++
			}
			break
		default:
			break
		}
		variableFormList = append(variableFormList, variableForm)
	}
	isCanRun := 1
	if waitPreNum > 0 {
		isCanRun = 0
	}
	return variableFormList, replaceList, isCanRun, nil
}

// RunProcess 执行前收集一些选择或者输入项 可以多次调用 有些待输入的还有替换符 可以多次执行
func (h *VariableRun) RunProcess(variableFormList []_struct.VariableForm, replaceList []map[string]string) ([]_struct.VariableForm, []map[string]string, int, error) {
	waitPreNum := 0
	needInputNum := 0
	inputNum := 0
	for key, variableForm := range variableFormList {
		switch cast.ToInt(variableForm.VariableType) {
		case define.VariableTypeInput: //输入框 不存在替换
			if variableForm.Input.Value != `` {
				variableForm.IsPreOk = 1
				if variableForm.ResultKey != `` {
					h.addReplace(&replaceList, variableForm.ResultKey, variableForm.Input.Value)
				}
			} else {
				waitPreNum++
			}
			needInputNum++
			if variableForm.Input.Value != `` {
				inputNum++
			}
			break
		case define.VariableTypeRadio: //单项选择
			variableForm.Select.Options = h.replace(variableForm.Select.Options, replaceList)
			needInputNum++
			if variableForm.Select.Value != `` {
				inputNum++
			}
			//没有选择 或者没有被替换 那么还是等参数
			if !h.isPreShowForm(variableForm.Select.Options) {
				waitPreNum++
				break
			}
			radioErr := h.radioPreProcess(&variableForm, &replaceList, variableForm.Select.Value)
			if radioErr != nil {
				return nil, nil, 0, radioErr
			}
			variableForm.IsPreOk = 1
			gstool.FmtPrintlnLogTime(`%#v`, variableForm.Select.OptionList)
			break
		case define.VariableTypeMysql: //执行sql
			variableForm.Sql.Sql = h.replace(variableForm.Sql.Sql, replaceList)
			gstool.FmtPrintlnLogTime(`sql %s`, variableForm.Sql.Sql)
			if !h.isPreShowForm(variableForm.Sql.Sql) {
				waitPreNum++
				break
			}
			sqlRet := h.sqlPreProcess(&variableForm, &replaceList)
			if sqlRet != nil {
				return nil, nil, 0, sqlRet
			}
			variableForm.IsPreOk = 1
			break
		default:
			break
		}
		variableFormList[key] = variableForm
	}
	//是否能够运行
	isCanRun := 1
	if waitPreNum > 0 || inputNum < needInputNum {
		isCanRun = 0
	}
	return variableFormList, replaceList, isCanRun, nil
}

func (h *VariableRun) replace(data string, replaceList []map[string]string) string {
	for _, replace := range replaceList {
		data = gstool.StringReplaces(data, replace)
	}
	return data
}

// 是否已经可以显示在页面上
func (h *VariableRun) addReplace(replaceList *[]map[string]string, key, value string) {
	boolFind := false
	for _, replace := range *replaceList {
		for mapKey, _ := range replace {
			if mapKey == key {
				boolFind = true
			}
		}
	}
	if !boolFind {
		*replaceList = append(*replaceList, map[string]string{
			key: value,
		})
	}
}

// 是否已经可以显示在页面上
func (h *VariableRun) isPreShowForm(data string) bool {
	return !gstool.RegexMatchString(data, `{[a-zA-Z0-9_]+}`)
}

// 组装单选属性
func (h *VariableRun) radioPreProcess(variableForm *_struct.VariableForm, replaceList *[]map[string]string, chooseValue string) error {
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
		//组装替换符
		if variableForm.ResultKey != `` && chooseValue != `` && chooseValue == option.Value {
			for optionKey, optionValue := range optionMap {
				h.addReplace(replaceList, variableForm.ResultKey+`.`+optionKey, cast.ToString(optionValue))
			}
			h.addReplace(replaceList, variableForm.ResultKey, gstool.JsonEncode(optionMap))
		}
	}
	variableForm.Select.OptionList = optionList
	return nil
}

// 组装sql查询属性
func (h *VariableRun) sqlPreProcess(form *_struct.VariableForm, replaceList *[]map[string]string) error {
	//如果带有替换符 那么忽略
	sql := cast.ToString(form.Sql.Sql)
	mysqlId := cast.ToInt(form.Sql.MysqlId)
	mysqlRet, mysqlRetErr := h.runMysqlSql(map[string]any{
		`sql`:      sql,
		`mysql_id`: mysqlId,
	})
	if mysqlRetErr != nil {
		return mysqlRetErr
	}
	if mysqlRet == `[]` {
		return errors.New(`未查找到数据`)
	}
	if form.ResultKey != `` {
		//TODO 这里需要支持[0].xxx 替换等 后面在搞
		h.addReplace(replaceList, form.ResultKey, mysqlRet)
	}
	return nil
}

// RunDone 最终执行
func (h *VariableRun) RunDone(variableId any, replaceList []map[string]string) error {
	h.VariableId = cast.ToString(variableId)
	h.ReplaceList = replaceList
	cmdList, cmdListErr := h.getVariableCmdList(variableId)
	if cmdListErr != nil {
		return cmdListErr
	}
	for _, cmd := range cmdList {
		resultKey := cast.ToString(cmd[`result_key`])
		isPre := cast.ToInt(cmd[`is_pre`])
		if isPre == 1 {
			continue
		}
		var result string
		var resultErr error
		switch cast.ToInt(cmd[`type`]) {
		case define.VariableTypeMysql:
			result, resultErr = h.runMysqlSql(cmd)
			break
		case define.VariableTypeBash:
			result, resultErr = h.runBash(cmd)
			break
		default:
			continue
		}
		if resultErr != nil {
			return resultErr
		}
		if resultKey != `` {
			//TODO 需要增加替换json或者数组
			//h.addReplace(&h.ReplaceList, resultKey, result)
		}
		gstool.FmtPrintlnLogTime(`%s`, result)
	}

	h.end()
	return nil
}

func (h *VariableRun) runMysqlSql(cmd map[string]any) (string, error) {
	sql := cast.ToString(cmd[`sql`])
	mysqlId := cast.ToInt(cmd[`mysql_id`])
	if mysqlId == 0 {
		return ``, errors.New(`mysql_id不能为空`)
	}
	mysqlConfig, mysqlConfigErr := Component.TSqlite.GetMysqlConfig(mysqlId)
	if mysqlConfigErr != nil {
		return "", mysqlConfigErr
	}
	mysqlClient, mysqlClientErr := Component.TMysql.GetClient(mysqlConfig)
	if mysqlClientErr != nil {
		return ``, mysqlClientErr
	}
	sql = h.replace(sql, h.ReplaceList)
	all, allErr := mysqlClient.QueryBySql(sql).All()
	if allErr != nil {
		return ``, allErr
	}
	return gstool.JsonEncode(all), nil
}

func (h *VariableRun) runBash(cmd map[string]any) (string, error) {
	sshId := cast.ToInt(cmd[`ssh_id`])
	bash := cast.ToString(cmd[`bash`])
	cmdId := cast.ToString(cmd[`id`])
	if bash == `` {
		return ``, errors.New(`脚本不能为空`)
	}
	bash = h.replace(bash, h.ReplaceList)
	if sshId == 0 {
		return ``, errors.New(`ssh不能为空`)
	}
	sshConfig, sshConfigErr := Component.TSqlite.GetSshConfig(sshId)
	if sshConfigErr != nil {
		return ``, sshConfigErr
	}
	var sshClientErr error
	var sshClient *gsssh.SshConfig
	//ssh
	sshUniqueKey := Component.TBase.GetCombineKey(cmd[`variable_id`], `variable`, sshConfig[`id`], `run`)
	sshClient, sshClientErr = Component.TShell.GetClient(sshConfig, sshUniqueKey)
	if sshClientErr != nil {
		return ``, sshClientErr
	}
	if _, ok := h.sshClientList[sshUniqueKey]; !ok {
		h.sshClientList[sshUniqueKey] = sshClient
	}
	//sftp
	sftpUniqueKey := Component.TBase.GetCombineKey(cmd[`variable_id`], `variable`, sshConfig[`id`], `sftp`)
	sftpClient, sftpClientErr := Component.TShell.GetClient(sshConfig, sftpUniqueKey)
	if sftpClientErr != nil {
		return ``, sftpClientErr
	}
	if _, ok := h.sftpSessionList[sftpUniqueKey]; !ok {
		h.sftpSessionList[sftpUniqueKey] = sftpClient
	}
	var err error
	//创建目录
	_, err = sshClient.RunCommandWait(`sudo mkdir -p /var/www/variable`)
	if err != nil {
		return ``, err
	}
	//写入脚本 用replace后不知道为什么打印日志没有问题，一执行echo就会重复写入几次 但是不执行h.replace又没有问题
	//_, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo echo '%s' > /var/www/variable/variable_%s.sh`, bash, cmdId))
	//if err != nil {
	//	return ``, err
	//}
	err = sftpClient.UploadFile(fmt.Sprintf(`/var/www/variable/variable_%s.sh`, cmdId), bash)
	if err != nil {
		return "", err
	}
	_, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo chmod +x /var/www/variable/variable_%s.sh`, cmdId))
	if err != nil {
		return ``, err
	}
	//var result string
	var result string
	result, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo /var/www/variable/variable_%s.sh`, cmdId))
	if err != nil {
		return ``, err
	}
	return result, nil
}

func (h *VariableRun) end() {
	for uniqueKey, sshClient := range h.sshClientList {
		sshClient.Close()
		Component.TShell.RmClient(uniqueKey)
	}
	for uniqueKey, sftpClient := range h.sftpSessionList {
		sftpClient.Close()
		Component.TShell.RmClient(uniqueKey)
	}
	h.sshClientList = make(map[string]*gsssh.SshConfig)
	h.sftpSessionList = make(map[string]*gsssh.SshConfig)
}

func (h *VariableRun) getVariableCmdList(variableId any) ([]map[string]any, error) {
	return Component.TSqlite.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`variable_id`: variableId,
	}).Order(`weight asc`).All()
}
