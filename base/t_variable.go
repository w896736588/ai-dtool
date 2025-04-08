package base

import (
	"context"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gshttp"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"strings"
	"time"
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
	for index, replace := range *replaceList {
		for mapKey, _ := range replace {
			if mapKey == key {
				boolFind = true
				(*replaceList)[index] = map[string]string{
					key: value,
				}
			}
		}
	}
	if !boolFind {
		*replaceList = append(*replaceList, map[string]string{
			key: value,
		})
	}
}

// 是否存在待替换的变量
func (h *VariableRun) isExistReplaceParam(data string) bool {
	return gstool.RegexMatchString(data, `{[a-zA-Z0-9_]+}`)
}

func (h *VariableRun) isExistReplaceList(resultKey string, replaceList []map[string]string) bool {
	for _, replaceMap := range replaceList {
		if _, ok := replaceMap[resultKey]; ok {
			return true
		}
	}
	return false
}

// 单选替换
func (h *VariableRun) radioChooseReplace(variableForm *_struct.VariableForm, replaceList *[]map[string]string, chooseValue string) error {
	for _, option := range variableForm.Select.OptionList {
		//组装替换符
		if variableForm.ResultKey != `` && chooseValue != `` && chooseValue == option.Value {
			h.sendStreamMsgMarkdownEnter(variableForm.Name + `[` + option.Label + `]`)
			//额外属性
			sourceOptionList := make(map[string]any, 0)
			_ = gstool.JsonDecode(option.Source, &sourceOptionList)
			for optionKey, optionValue := range sourceOptionList {
				h.addReplace(replaceList, variableForm.ResultKey+`.`+optionKey, cast.ToString(optionValue))
			}
			//替换整体
			h.addReplace(replaceList, variableForm.ResultKey, option.Source)
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
		`name`:     form.Name,
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
	for _, cmd := range cmdList {
		h.sendStreamMsgMarkdownEnter(cast.ToString(cmd[`name`]) + `->执行`)
		resultKey := cast.ToString(cmd[`result_key`])
		isPre := cast.ToInt(cmd[`is_pre`])
		if isPre == 1 { //提前运行的不管
			continue
		}
		var result string
		var resultErr error
		switch cast.ToInt(cmd[`type`]) {
		case define.VariableCmdMysql:
			result, resultErr = h.runMysqlSql(cmd)
		case define.VariableCmdBash:
			result, resultErr = h.runBash(cmd)
		case define.VariableCmdRedis:
			result, resultErr = h.runRedis(cmd)
		case define.VariableCmdPlaywright:
			result, resultErr = h.runPlaywright(cmd)
		case define.VariableCmdCombine:
			result, resultErr = h.runCombine(cmd)
		case define.VariableCmdCurl:
			result, resultErr = h.runCurl(cmd)
		default:
			continue
		}
		if resultErr != nil {
			return resultErr
		}
		if resultKey != `` {
			switch cast.ToInt(cmd[`type`]) {
			case define.VariableCmdBash, define.VariableCmdCombine, define.VariableCmdCurl:
				h.addReplace(&h.ReplaceList, resultKey, result)
			default:
			}
		}
	}

	h.end()
	return nil
}

func (h *VariableRun) sendStreamMsg(msg string) error {
	return Component.TSse.SendMsg(define.SseVariable, msg)
}

// 输出换行文本
func (h *VariableRun) sendStreamMsgMarkdownEnter(msg string) {
	_ = h.sendStreamMsg(Component.TMarkDown.Enter(msg))
}

func (h *VariableRun) runMysqlSql(cmd map[string]any) (string, error) {
	mysqlId := ``
	sql := ``
	name := cast.ToString(cmd[`name`])
	cmd[`sql`] = h.replace(cast.ToString(cmd[`sql`]), h.ReplaceList)
	if cast.ToInt(cmd[`mysql_id`]) == 0 { //当没有传递mysql_id时，那么从sql里面找
		mysqlId, sql = h.ParseIdContent(cast.ToString(cmd[`sql`]))
	} else {
		mysqlId = cast.ToString(cmd[`mysql_id`])
		sql = cast.ToString(cmd[`sql`])
	}
	if cast.ToInt(mysqlId) == 0 {
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
	if len(gstool.RegexSearchString(sql, "(?i)select")) > 0 {
		h.sendStreamMsgMarkdownEnter(Component.TMarkDown.BlockQuote(name))
		h.sendStreamMsgMarkdownEnter(Component.TMarkDown.Code(sql, `sql`))
		all, allErr := mysqlClient.QueryBySql(sql).All()
		if allErr != nil {
			return ``, allErr
		}
		h.sendStreamMsgMarkdownEnter(Component.TMarkDown.Json(all))
		return gstool.JsonEncode(all), nil
	} else if len(gstool.RegexSearchString(sql, "(?i)update")) > 0 {
		h.sendStreamMsgMarkdownEnter(name + `->` + sql)
		affectRows, execErr := mysqlClient.ExecBySql(sql).Exec()
		h.sendStreamMsgMarkdownEnter(name + `->更新数` + cast.ToString(affectRows))
		if execErr != nil {
			return ``, execErr
		}
	}
	return ``, nil
}

func (h *VariableRun) runBash(cmd map[string]any) (string, error) {
	sshId := ``
	bash := ``
	cmd[`bash`] = h.replace(cast.ToString(cmd[`bash`]), h.ReplaceList)
	if cast.ToInt(cmd[`ssh_id`]) == 0 {
		sshId, bash = h.ParseIdContent(cast.ToString(cmd[`bash`]))
	} else {
		sshId = cast.ToString(cmd[`ssh_id`])
		bash = cast.ToString(cmd[`bash`])
	}
	cmdId := cast.ToString(cmd[`id`])
	if bash == `` {
		return ``, errors.New(`脚本不能为空`)
	}
	if cast.ToInt(sshId) == 0 {
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
	sshClient, sshClientErr = Component.TShell.GetClientMarkdown(sshConfig, sshUniqueKey, define.SseVariable)
	if sshClientErr != nil {
		return ``, sshClientErr
	}
	//sftp
	sftpClient, sftpClientErr := Component.TShell.GetClientMarkdown(sshConfig, sftpUniqueKey, define.SseVariable)
	if sftpClientErr != nil {
		return ``, sftpClientErr
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

func (h *VariableRun) runCurl(cmd map[string]any) (string, error) {
	url := h.replace(cast.ToString(cmd[`bash`]), h.ReplaceList)
	if url == `` {
		return ``, errors.New(`url不能为空`)
	}
	h.sendStreamMsgMarkdownEnter(Component.TMarkDown.Code(url, `shell`))
	isStream := cast.ToInt(gstool.UrlGetParam(url, `is_stream`))
	var result []byte
	var err error
	if isStream == 1 {
		result, err = gshttp.Get(url).OpenStreamBytesEnd([]byte("\n\n"), func(msg string, err error) {
			if err != nil {
				gstool.FmtPrintlnLogTime(`收到失败 %s`, err.Error())
			}
			gstool.FmtPrintlnLogTime(`curl 请求收到 ---%s---`, msg)
			sendMsg := Component.TAi.ParseStream(msg)
			gstool.FmtPrintlnLogTime(`解析后---%s---`, sendMsg)
			_ = h.sendStreamMsg(cast.ToString(sendMsg))
		}, func(bytes []byte) []byte {
			return bytes
		}).Request(200).Result()
	} else {
		result, err = gshttp.Get(url).Request(200).Result()
	}
	return cast.ToString(result), err
}

func (h *VariableRun) runPlaywright(cmd map[string]any) (string, error) {
	id := cast.ToInt(cmd[`smart_link_id`])
	label := cast.ToString(cmd[`smart_link_label`])
	if id == 0 {
		return ``, errors.New(`链接ID不能为空`)
	}
	if label == `` {
		return ``, errors.New(`链接label不能为空`)
	}
	runParams, runParamsErr := Component.TSmartLink.GetRunParams(id, label, ``, ``, 0, h.ReplaceList)
	if runParamsErr != nil {
		return ``, errors.New(runParamsErr.Error())
	}
	for {
		if Component.TSmartLink.IsRun {
			h.sendStreamMsgMarkdownEnter(`等待其他自动化链接任务完成..`)
			time.Sleep(time.Second * 1)
			continue
		} else {
			break
		}
	}
	//注册需要监听的接口
	//需要注册的uri
	listenUriList := cast.ToString(cmd[`options`])
	if listenUriList != `` {
		listenM := make([]map[string]string, 0)
		_ = gstool.JsonDecode(listenUriList, &listenM)
		for _, v := range listenM {
			uri := cast.ToString(v[`uri`])
			if uri == `` {
				continue
			}
			h.sendStreamMsgMarkdownEnter(`注册监听` + uri)
			Component.TSmartLink.ListenUrlList[uri] = &_struct.ListenUrl{
				IsSse: true,
				Callback: func(msg string, err error) {
					gstool.FmtPrintlnLogTime(`收到---%s---`, msg)
					sendMsg := Component.TAi.ParseStream(msg)
					_ = h.sendStreamMsg(cast.ToString(sendMsg))
				},
				StartCallBack: func() {
					h.sendStreamMsgMarkdownEnter(Component.TMarkDown.BlockQuote(`开始请求大模型`))
					h.sendStreamMsgMarkdownEnter(Component.TMarkDown.BlockQuote(`  `))
				},
				EndCallBack: func(msg string) {
					h.sendStreamMsgMarkdownEnter(Component.TMarkDown.BlockQuote(msg))
				},
			}
		}
	}

	Component.TSmartLink.IsRun = true
	for i := 0; i < runParams.OpenNum; i++ {
		openErr := Component.TSmartLink.OpenBrowserPlaywright(runParams)
		if openErr != nil {
			gstool.FmtPrintlnLogTime(`错误 %s`, openErr.Error())
		}
	}
	Component.TSmartLink.IsRun = false
	return ``, nil
}

func (h *VariableRun) runCombine(cmd map[string]any) (string, error) {
	return h.replace(cast.ToString(cmd[`options`]), h.ReplaceList), nil
}

func (h *VariableRun) runRedis(cmd map[string]any) (string, error) {
	redisId := ``
	redisBash := ``
	name := cast.ToString(cmd[`name`])
	cmd[`bash`] = h.replace(cast.ToString(cmd[`bash`]), h.ReplaceList)
	if cast.ToInt(cmd[`redis_id`]) == 0 {
		redisId, redisBash = h.ParseIdContent(cast.ToString(cmd[`bash`]))
	} else {
		redisId = cast.ToString(cmd[`redis_id`])
		redisBash = cast.ToString(cmd[`bash`])
	}
	if cast.ToInt(redisId) == 0 {
		return ``, errors.New(`redis不能为空`)
	}
	if redisBash == `` {
		return ``, errors.New(`redis需要删除的key不能为空`)
	}
	redisConfig, redisConfigErr := Component.TSqlite.GetRedisConfig(redisId)
	if redisConfigErr != nil {
		return ``, redisConfigErr
	}
	client, clientErr := Component.TRedis.GetClient(redisConfig)
	if clientErr != nil {
		return "", clientErr
	}
	//解析命令格式：
	//字符串删除string,delete,key
	redisBashParamList := strings.Split(redisBash, `,`)
	if len(redisBashParamList) >= 3 {
		switch redisBashParamList[0] {
		case `string`:
			switch redisBashParamList[1] {
			case `delete`:
				h.sendStreamMsgMarkdownEnter(name + `->清除redis，string key：` + redisBashParamList[2])
				client.Client.Del(context.Background(), redisBashParamList[2])
			default:
				h.sendStreamMsgMarkdownEnter(name + `->暂不支持的操作` + redisBash)
			}
		case `hash`:
			switch redisBashParamList[1] {
			case `delete`:
				h.sendStreamMsgMarkdownEnter(name + `->清除redis，hash key：` + redisBashParamList[2] + ` field：` + redisBashParamList[3])
				client.Client.HDel(context.Background(), redisBashParamList[2], redisBashParamList[3])
			default:
				h.sendStreamMsgMarkdownEnter(name + `->暂不支持的操作` + redisBash)
			}
		default:
			h.sendStreamMsgMarkdownEnter(name + `->暂不支持的操作` + redisBash)
		}
	} else {
		h.sendStreamMsgMarkdownEnter(name + `->格式错误` + redisBash)
	}
	return `操作`, nil
}

func (h *VariableRun) end() {
	h.sendStreamMsgMarkdownEnter(`执行结束`)
}

func (h *VariableRun) getVariableCmdList(variableId any) ([]map[string]any, error) {
	return Component.TSqlite.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`variable_id`: variableId,
		`status`:      1,
	}).Order(`weight asc`).All()
}
