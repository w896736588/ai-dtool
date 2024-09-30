package base

import (
	"context"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
)

type VariableRun struct {
	VariableId  string
	ReplaceList []map[string]string
}

func NewVariable() VariableRun {
	return VariableRun{}
}

func (h *VariableRun) replace(data string, replaceList []map[string]string) string {
	for _, replace := range replaceList {
		//处理特殊情况
		for replaceKey, replaceVal := range replace {
			//取模
			matchSubList := gstool.RegexMatchSubString(data, replaceKey+`%(\d+)`)
			if len(matchSubList) >= 2 {
				data = gstool.StringReplaces(data, map[string]string{
					matchSubList[0]: cast.ToString(cast.ToInt64(replaceVal) % cast.ToInt64(matchSubList[1])),
				})
			}
		}
		data = gstool.StringReplaces(data, replace)

	}
	return data
}

// 是否已经可以显示在页面上
func (h *VariableRun) addReplace(replaceList *[]map[string]string, key, value string) {
	if key == `` {
		return
	}
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

// 单选替换
func (h *VariableRun) radioChooseReplace(variableForm *_struct.VariableForm, replaceList *[]map[string]string, chooseValue string) error {
	for _, option := range variableForm.Select.OptionList {
		//组装替换符
		if variableForm.ResultKey != `` && chooseValue != `` && chooseValue == option.Value {
			//额外属性
			sourceOptionList := make(map[string]any, 0)
			_ = gstool.JsonDecode(option.Source, &sourceOptionList)
			for optionKey, optionValue := range sourceOptionList {
				h.addReplace(replaceList, variableForm.ResultKey+`.`+optionKey, cast.ToString(optionValue))
			}
			//替换整体
			h.addReplace(replaceList, variableForm.ResultKey, gstool.JsonEncode(option))
		}
	}
	return nil
}

// 执行sql
func (h *VariableRun) sqlProcessRun(form *_struct.VariableForm, replaceList *[]map[string]string) error {
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
func (h *VariableRun) RunDone(variableId any, replaceList []map[string]string, variableFormList []_struct.VariableForm) error {
	h.VariableId = cast.ToString(variableId)
	h.ReplaceList = replaceList
	cmdList, cmdListErr := h.getVariableCmdList(variableId)
	if cmdListErr != nil {
		return cmdListErr
	}
	//拿到值
	redisId := ``
	for _, variableForm := range variableFormList {
		if cast.ToInt(variableForm.VariableType) == define.VariableTypeRedisChoose {
			redisId = variableForm.Select.Value
		}
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
		case define.VariableTypeRedisDelete:
			result, resultErr = h.runRedisDelete(cmd, cast.ToInt(redisId), cast.ToString(variableId))
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
		//暂时没啥用
		Component.GsLog.Debugf(`执行结果 %s`, result)
	}

	h.end()
	return nil
}

func (h *VariableRun) getSocket(variableId string) *websocket.Conn {
	uniqueKey := Component.TBase.GetCombineKey(variableId, `variable`)
	return Component.TSocket.GetSocket(uniqueKey)
}

func (h *VariableRun) sendSocketMsg(variableId any, msg string) {
	msg = ` ` + msg
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	socket := h.getSocket(cast.ToString(variableId))
	if socket != nil {
		err := socket.WriteMessage(websocket.TextMessage, []byte(msg+"\n"))
		if err != nil {
			return
		}
	}
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
	if len(gstool.RegexSearchString(sql, "(?i)select")) > 0 {
		h.sendSocketMsg(h.VariableId, `执行查询：`+sql)
		all, allErr := mysqlClient.QueryBySql(sql).All()
		h.sendSocketMsg(h.VariableId, `结果：`+gstool.JsonEncode(all))
		if allErr != nil {
			return ``, allErr
		}
		return gstool.JsonEncode(all), nil
	} else if len(gstool.RegexSearchString(sql, "(?i)update")) > 0 {
		h.sendSocketMsg(h.VariableId, `执行：`+sql)
		affectRows, execErr := mysqlClient.ExecBySql(sql).Exec()
		h.sendSocketMsg(h.VariableId, `更新数：`+cast.ToString(affectRows))
		if execErr != nil {
			return ``, execErr
		}
	}
	return ``, nil
}

func (h *VariableRun) runBash(cmd map[string]any) (string, error) {
	sshId := cast.ToInt(cmd[`ssh_id`])
	bash := cast.ToString(cmd[`bash`])
	cmdId := cast.ToString(cmd[`id`])
	if bash == `` {
		return ``, errors.New(`脚本不能为空`)
	}
	gstool.FmtPrintlnLogTime(`替换前的脚本 %s %#v`, bash, h.ReplaceList)
	bash = h.replace(bash, h.ReplaceList)
	gstool.FmtPrintlnLogTime(`替换后的脚本 %s`, bash)
	if sshId == 0 {
		return ``, errors.New(`ssh不能为空`)
	}
	sshUniqueKey := Component.TBase.GetCombineKey(`variable`, sshId, `run`)
	sftpUniqueKey := Component.TBase.GetCombineKey(`variable`, sshId, `sftp`)
	if !Component.TShell.Exist(sshUniqueKey) || !Component.TShell.Exist(sftpUniqueKey) {
		return ``, errors.New(`ssh连接未初始化`)
	}
	sshConfig, sshConfigErr := Component.TSqlite.GetSshConfig(sshId)
	if sshConfigErr != nil {
		return ``, sshConfigErr
	}
	var sshClientErr error
	var sshClient *gsssh.SshConfig
	//ssh
	sshClient, sshClientErr = Component.TShell.GetClient(sshConfig, sshUniqueKey)
	if sshClientErr != nil {
		return ``, sshClientErr
	}
	sshClient.SetSocket(h.getSocket(h.VariableId))
	//sftp
	sftpClient, sftpClientErr := Component.TShell.GetClient(sshConfig, sftpUniqueKey)
	if sftpClientErr != nil {
		return ``, sftpClientErr
	}
	sftpClient.SetSocket(h.getSocket(h.VariableId))
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

func (h *VariableRun) runRedisDelete(cmd map[string]any, redisId int, variableId string) (string, error) {
	config := cast.ToString(cmd[`bash`])
	if config == `` {
		return ``, errors.New(`redis需要删除的key不能为空`)
	}
	config = h.replace(config, h.ReplaceList)
	if redisId == 0 {
		return ``, errors.New(`redis不能为空`)
	}
	redisConfig, redisConfigErr := Component.TSqlite.GetRedisConfig(redisId)
	if redisConfigErr != nil {
		return ``, redisConfigErr
	}
	client, clientErr := Component.TRedis.GetClient(redisConfig)
	if clientErr != nil {
		return "", clientErr
	}
	h.sendSocketMsg(variableId, `清除redis，key：`+config)
	client.Client.Del(context.Background(), config)
	return `删除缓存成功`, nil
}

func (h *VariableRun) end() {
	h.sendSocketMsg(h.VariableId, `执行结束`)
}

func (h *VariableRun) getVariableCmdList(variableId any) ([]map[string]any, error) {
	return Component.TSqlite.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`variable_id`: variableId,
	}).Order(`weight asc`).All()
}
