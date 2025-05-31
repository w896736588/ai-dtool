package p_variable

import (
	"bytes"
	"context"
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"dev_tool/internal/pkg/p_playwright"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gshttp"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"os/exec"
	"strings"
	"sync"
)

type RCmd struct {
	cmd            map[string]any
	replaceList    *[]map[string]string
	StreamMsg      func(string, bool)
	PlaywrightLock sync.RWMutex
	RunUniqueId    string
}

func NewRCmd(cmd map[string]any, replace *[]map[string]string, RunUniqueId string, streamMsg func(string, bool)) *RCmd {
	return &RCmd{
		cmd:         cmd,
		replaceList: replace,
		StreamMsg:   streamMsg,
		RunUniqueId: RunUniqueId,
	}
}

func (h *RCmd) RunMysql() error {
	cmdSql := cast.ToString(h.cmd[`sql`])
	resultKey := cast.ToString(h.cmd[`result_key`])
	//替换
	cmdSql = base.Component.TVariable.Replace(cmdSql, h.replaceList)
	//解析Id
	mysqlId, sql, err := base.Component.TVariable.ParseIdContent(cmdSql)
	if err != nil {
		return err
	}
	//检查是否还有未替换的
	if base.Component.TVariable.ExistReplaceParam(sql) {
		return errors.New(`还存在未替换的参数：` + sql)
	}
	//执行
	mysqlConfig, mysqlConfigErr := base.Component.TSqlite.GetMysqlConfig(mysqlId)
	if mysqlConfigErr != nil {
		return errors.New(`找不到mysql配置 ` + mysqlConfigErr.Error())
	}
	mysqlClient, mysqlClientErr := base.Component.TMysql.GetClient(mysqlConfig)
	if mysqlClientErr != nil {
		return errors.New(`获取mysql client 失败 ` + mysqlClientErr.Error())
	}
	result := ``
	h.StreamMsg(fmt.Sprintf(`%s %s 执行sql:`,
		base.Component.TMarkDown.Bold(`run`),
		cast.ToString(h.cmd[`name`])), true)
	if len(gstool.RegexSearchString(sql, "(?i)select")) > 0 {
		result = base.Component.TMarkDown.Code(sql, `sql`)
		h.StreamMsg(result, true)
		all, allErr := mysqlClient.QueryBySql(sql).All()
		if allErr != nil {
			return allErr
		}
		//增加替换变量
		if resultKey != `` && len(all) > 0 {
			base.Component.TVariable.AddReplace(h.replaceList, resultKey, gstool.JsonEncode(all))
		}
		return nil
	} else if len(gstool.RegexSearchString(sql, "(?i)update")) > 0 {
		affectRows, execErr := mysqlClient.ExecBySql(sql).Exec()
		result = base.Component.TMarkDown.Code("-- 更新数"+cast.ToString(affectRows)+"\n"+sql, `sql`)
		h.StreamMsg(result, true)
		if execErr != nil {
			return execErr
		}
	}
	return nil
}

func (h *RCmd) RunBat() (string, error) {
	bat := cast.ToString(h.cmd[`bash`])
	//cmdId := cast.ToString(h.cmd[`id`])
	base.Component.TVariable.Log.Debugf(`run bash \n 替换列表 %s`, gstool.JsonEncode(h.replaceList))
	bat = base.Component.TVariable.Replace(bat, h.replaceList)
	// 构建命令
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command(`cmd.exe`, `/C`, bat)
	gstool.FmtPrintlnLogTime(`bat %s`, bat)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	h.StreamMsg(fmt.Sprintf("执行：%s ", bat), true)
	err := cmd.Run()
	if err != nil {
		return ``, gstool.Error("make 执行失败: %v\n", err)
	}
	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()
	h.StreamMsg(stdoutStr, true)
	h.StreamMsg(stderrStr, true)
	h.StreamMsg(`构建完成`, true)
	return ``, nil
}

func (h *RCmd) RunBash() (string, error) {
	cmdBash := cast.ToString(h.cmd[`bash`])
	cmdId := cast.ToString(h.cmd[`id`])
	base.Component.TVariable.Log.Debugf(`run bash \n 替换列表 %s`, gstool.JsonEncode(h.replaceList))
	cmdBash = base.Component.TVariable.Replace(cmdBash, h.replaceList)
	sshId, bash, parseIdErr := base.Component.TVariable.ParseIdContent(cmdBash)
	if parseIdErr != nil {
		return ``, parseIdErr
	}
	//如果脚本还有未替换的
	if base.Component.TVariable.ExistReplaceParam(bash) {
		return ``, gstool.Error("执行的脚本还存在需要替换的内容")
	}
	//注册ssh
	sshUniqueKey := base.Component.TBase.GetCombineKey(`variable`, sshId, `run`)
	sftpUniqueKey := base.Component.TBase.GetCombineKey(`variable`, sshId, `sftp`)
	//链接ssh
	preConnErr := base.Component.TVariable.PreConnSsh(sshId, sshUniqueKey, sftpUniqueKey)
	if preConnErr != nil {
		return ``, gstool.Error(`链接失败 %s`, preConnErr.Error())
	}
	if !base.Component.TShell.Exist(sshUniqueKey) || !base.Component.TShell.Exist(sftpUniqueKey) {
		return ``, errors.New(`ssh连接未初始化`)
	}
	//全局注册 用于停止
	base.Component.TVariable.AddSshClient(h.RunUniqueId, sshUniqueKey)
	base.Component.TVariable.AddSshClient(h.RunUniqueId, sftpUniqueKey)
	//初始化
	sshConfig, sshConfigErr := base.Component.TSqlite.GetSshConfig(sshId)
	if sshConfigErr != nil {
		return ``, sshConfigErr
	}
	var sshClientErr error
	var sshClient *gsssh.SshConfig
	//家目录
	home := `/var/www`
	if cast.ToString(sshConfig[`home`]) != `` {
		home = cast.ToString(sshConfig[`home`])
	}
	variableDir := home + `/variable`
	//ssh
	sshClient, sshClientErr = base.Component.TShell.GetClientMarkdown(sshConfig, sshUniqueKey, define.SseVariable)
	if sshClientErr != nil {
		return ``, sshClientErr
	}
	//sftp
	sftpClient, sftpClientErr := base.Component.TShell.GetClientMarkdown(sshConfig, sftpUniqueKey, define.SseVariable)
	if sftpClientErr != nil {
		return ``, sftpClientErr
	}
	var err error
	//创建目录
	_, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo mkdir -p %s`, variableDir))
	if err != nil {
		return ``, err
	}
	//写入脚本 用replace后不知道为什么打印日志没有问题，一执行echo就会重复写入几次 但是不执行h.replace又没有问题
	//_, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo echo '%s' > /var/www/variable/variable_%s.sh`, bash, cmdId))
	//if err != nil {
	//	return ``, err
	//}
	base.Component.TVariable.Log.Debugf(`%s \n %s `, fmt.Sprintf(variableDir+`/variable_%s.sh`, cmdId), bash)
	err = sftpClient.UploadFile(fmt.Sprintf(variableDir+`/variable_%s.sh`, cmdId), bash)
	if err != nil {
		return "", err
	}
	_, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo chmod +x %s/variable_%s.sh`, variableDir, cmdId))
	if err != nil {
		return ``, err
	}
	//var result string
	var result string
	result, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo %s/variable_%s.sh`, variableDir, cmdId))
	if err != nil {
		return ``, err
	}
	return result, nil
}

func (h *RCmd) RunUpload() (string, error) {
	cmdBash := cast.ToString(h.cmd[`bash`])
	//cmdId := cast.ToString(h.cmd[`id`])
	cmdBash = base.Component.TVariable.Replace(cmdBash, h.replaceList)
	sshId, bash, parseIdErr := base.Component.TVariable.ParseIdContent(cmdBash)
	if parseIdErr != nil {
		return ``, parseIdErr
	}
	//如果脚本还有未替换的
	if base.Component.TVariable.ExistReplaceParam(bash) {
		return ``, gstool.Error("上传的脚本还存在需要替换的内容")
	}
	//解析配置
	bash = gstool.SReplaces(bash, map[string]string{
		`\`: `\\`,
	})
	uploadConfig := make(map[string]any)
	deErr := gstool.JsonDecode(bash, &uploadConfig)
	if deErr != nil {
		gstool.FmtPrintlnLogTime(`--%s-- %s`, bash, deErr.Error())
		h.StreamMsg(fmt.Sprintf(`解析上传配置失败 %s`, bash), true)
		return ``, deErr
	}
	targetDir := cast.ToString(uploadConfig[`target_dir`])
	sourceFile := cast.ToString(uploadConfig[`source_file`])
	if targetDir == `` {
		h.StreamMsg(`目标目录不能为空`, true)
		return ``, gstool.Error(`目标目录不能为空`)
	}
	if sourceFile == `` {
		h.StreamMsg(`源文件不能为空`, true)
		return ``, gstool.Error(`源文件不能为空`)
	}
	if !gstool.FileIsExisted(sourceFile) {
		h.StreamMsg(`源文件不存在`, true)
		return ``, gstool.Error(`源文件不存在`)
	}
	//注册ssh
	sftpUniqueKey := base.Component.TBase.GetCombineKey(`variable`, sshId, `sftp`)
	base.Component.TVariable.AddSshClient(h.RunUniqueId, sftpUniqueKey)
	//初始化
	sshConfig, sshConfigErr := base.Component.TSqlite.GetSshConfig(sshId)
	if sshConfigErr != nil {
		return ``, sshConfigErr
	}
	//sftp
	sftpClient, sftpClientErr := base.Component.TShell.GetClientMarkdown(sshConfig, sftpUniqueKey, define.SseVariable)
	if sftpClientErr != nil {
		return ``, sftpClientErr
	}
	var err error
	fileName := gstool.FileGetNameByPath(sourceFile)
	targetFile := targetDir + `/` + fileName
	h.StreamMsg(fmt.Sprintf(`准备上传文件 %s 到目标文件 %s`, sourceFile, targetFile), true)
	err = sftpClient.UploadFile(targetFile, sourceFile)
	if err != nil {
		h.StreamMsg(fmt.Sprintf(`上传文件失败 %s`, err.Error()), true)
		return "", err
	}
	h.StreamMsg(`上传完成`, true)
	return ``, nil
}

func (h *RCmd) RunCommand() (string, error) {
	cmdBash := cast.ToString(h.cmd[`bash`])
	cmdBash = base.Component.TVariable.Replace(cmdBash, h.replaceList)
	sshId, bash, parseIdErr := base.Component.TVariable.ParseIdContent(cmdBash)
	if parseIdErr != nil {
		return ``, parseIdErr
	}
	//注册client
	sshUniqueKey := base.Component.TBase.GetCombineKey(`variable`, sshId, `run`)
	sftpUniqueKey := base.Component.TBase.GetCombineKey(`variable`, sshId, `sftp`)
	preConnErr := base.Component.TVariable.PreConnSsh(sshId, sshUniqueKey, sftpUniqueKey)
	if preConnErr != nil {
		return ``, gstool.Error(`链接失败 %s`, preConnErr.Error())
	}
	//全局注册 用于停止
	base.Component.TVariable.AddSshClient(h.RunUniqueId, sshUniqueKey)
	base.Component.TVariable.AddSshClient(h.RunUniqueId, sftpUniqueKey)
	//分离出来多行命令
	commandList := strings.Split(bash, "\n")
	for _, command := range commandList {
		if command == "" {
			continue
		}
		sshUniqueKey := base.Component.TBase.GetCombineKey(`variable`, sshId, `run`)
		if !base.Component.TShell.Exist(sshUniqueKey) {
			return ``, errors.New(`ssh连接未初始化`)
		}
		sshConfig, sshConfigErr := base.Component.TSqlite.GetSshConfig(sshId)
		if sshConfigErr != nil {
			return ``, sshConfigErr
		}
		var sshClientErr error
		var sshClient *gsssh.SshConfig
		//ssh
		sshClient, sshClientErr = base.Component.TShell.GetClientMarkdown(sshConfig, sshUniqueKey, define.SseVariable)
		if sshClientErr != nil {
			return ``, sshClientErr
		}
		var err error
		runCmd := base.Command{}
		runCmd.SetCommand(command)
		runCmd.Sudo()
		h.StreamMsg(base.Component.TMarkDown.Code(runCmd.GetCommand().ToStr(), `bash`), true)
		_, err = sshClient.RunCommandWait(runCmd.GetCommand().ToStr())
		if err != nil {
			return ``, err
		}
	}
	return ``, nil
}

func (h *RCmd) RunCurl() (string, error) {
	resultKey := cast.ToString(h.cmd[`result_key`])
	url := base.Component.TVariable.Replace(cast.ToString(h.cmd[`bash`]), h.replaceList)
	if url == `` {
		return ``, errors.New(`url不能为空`)
	}
	url, _ = base.Component.TVariable.ParseConfig(url)
	h.StreamMsg(base.Component.TMarkDown.BlockQuote(`请求url,`+url), true)
	isStream := cast.ToInt(gstool.UrlGetParam(url, `is_stream`))
	var result []byte
	var err error
	if isStream == 1 {
		result, err = gshttp.Get(url).OpenStreamBytesEnd([]byte("\n\n"), func(msg string, err error) {
			if err != nil {
				return
			}
			sendMsg := base.Component.TAi.ParseStream(url, msg)
			h.StreamMsg(base.Component.TMarkDown.BlockQuote(cast.ToString(sendMsg)), false)
		}, func(bytes []byte) []byte {
			sendMsg := base.Component.TAi.ParseStream(url, cast.ToString(bytes))
			if gstool.SContains(cast.ToString(sendMsg), []string{`commit 共：`, `获取完项目列表 共：`}) { //这种内容不要汇集到最终结果中
				return []byte{}
			} else {
				return bytes
			}
		}).Request(200).Result()
	} else {
		result, err = gshttp.Get(url).Request(200).Result()
	}
	//增加替换变量
	if resultKey != `` {
		base.Component.TVariable.AddReplace(h.replaceList, resultKey, cast.ToString(result))
	}
	return cast.ToString(result), err
}

func (h *RCmd) RunPlaywright() (string, error) {
	id := cast.ToInt(h.cmd[`smart_link_id`])
	label := cast.ToString(h.cmd[`smart_link_label`])
	if id == 0 {
		return ``, errors.New(`链接ID不能为空`)
	}
	if label == `` {
		return ``, errors.New(`链接label不能为空`)
	}
	runParams, runParamsErr := base.Component.TPlaywright.GetRunParams(id, label, ``, ``, 0, h.replaceList)
	if runParamsErr != nil {
		return ``, errors.New(runParamsErr.Error())
	}
	//注册链接执行时需要输出的文本类型
	runParams.RunCallFunc = func(cmdType define.ProcessType, errmsg, tip, content string) {
		switch cmdType {
		case define.Input:
			h.StreamMsg(base.Component.TMarkDown.Bold(tip)+`,`+content+` `+errmsg, true)
		case define.CanvasImage:
			h.StreamMsg(base.Component.TMarkDown.Bold(tip)+`,`+errmsg, true)
			h.StreamMsg(content, true)
		case define.ExistWait, define.NoExistWait:
			h.StreamMsg(base.Component.TMarkDown.Bold(tip)+`,`+errmsg, true)
		}
	}
	//注册需要监听的接口
	//需要注册的uri
	listenUriList := cast.ToString(h.cmd[`options`])
	ListenUrlList := make(map[string]*_struct.ListenUrl)
	if listenUriList != `` {
		listenM := make([]map[string]string, 0)
		_ = gstool.JsonDecode(listenUriList, &listenM)
		for _, v := range listenM {
			uri := cast.ToString(v[`uri`])
			if uri == `` {
				continue
			}
			ListenUrlList[uri] = &_struct.ListenUrl{
				IsSse: true,
				Callback: func(url, msg string, err error) {
					base.Component.TVariable.Log.Debugf(`收到消息---%s---`, msg)
					sendMsg := base.Component.TAi.ParseStream(url, msg)
					h.StreamMsg(cast.ToString(sendMsg), false)
				},
				StartCallBack: func(url string) {
					base.Component.TVariable.Log.Debugf(`监听到%s`, url)
					h.StreamMsg(base.Component.TMarkDown.BlockQuote("开始回答...")+"\n\n", true)
				},
				EndCallBack: func(msg string) {
				},
			}
		}
	}
	runParams.ListenUrlList = ListenUrlList
	for i := 0; i < runParams.OpenNum; i++ {
		h.StreamMsg("\n"+base.Component.TMarkDown.Bold(label)+`,启动`, true)
		p := p_playwright.NewPlaywright(runParams, base.Component.TVariable.Log)
		openErr := p.Open()
		if openErr != nil {
			h.StreamMsg(base.Component.TMarkDown.BlockQuote(cast.ToString(h.cmd[`name`])+`,启动失败，`+openErr.Error()), true)
		}
	}
	return ``, nil
}

func (h *RCmd) RunCombine() (string, error) {
	resultKey := cast.ToString(h.cmd[`result_key`])
	combine := base.Component.TVariable.Replace(cast.ToString(h.cmd[`options`]), h.replaceList)
	//增加替换变量
	if resultKey != `` {
		base.Component.TVariable.AddReplace(h.replaceList, resultKey, combine)
	}
	return ``, nil
}

func (h *RCmd) RunRedis() (string, error) {
	name := cast.ToString(h.cmd[`name`])
	cmdBash := base.Component.TVariable.Replace(cast.ToString(h.cmd[`bash`]), h.replaceList)
	redisId, redisBash, parseErr := base.Component.TVariable.ParseIdContent(cmdBash)
	if parseErr != nil {
		return ``, errors.New(`redis解析失败` + parseErr.Error())
	}
	if redisBash == `` {
		return ``, errors.New(`redis需要删除的key不能为空`)
	}
	redisConfig, redisConfigErr := base.Component.TSqlite.GetRedisConfig(redisId)
	if redisConfigErr != nil {
		return ``, redisConfigErr
	}
	client, clientErr := base.Component.TRedis.GetClient(redisConfig)
	if clientErr != nil {
		return "", clientErr
	}
	h.StreamMsg(name+`,`+redisBash, true)
	//解析命令格式：
	//字符串删除string,delete,key
	redisBashParamList := strings.Split(redisBash, `,`)
	if len(redisBashParamList) >= 3 {
		switch redisBashParamList[0] {
		case `string`:
			switch redisBashParamList[1] {
			case `delete`:
				client.Client.Del(context.Background(), redisBashParamList[2])
			default:
				return ``, errors.New(`暂不支持的操作，` + redisBash)
			}
		case `hash`:
			switch redisBashParamList[1] {
			case `delete`:
				client.Client.HDel(context.Background(), redisBashParamList[2], redisBashParamList[3])
			default:
				return ``, errors.New(`暂不支持的操作，` + redisBash)
			}
		default:
			return ``, errors.New(`暂不支持的操作，` + redisBash)
		}
	} else {
		return ``, errors.New(`格式错误，` + redisBash)
	}
	return `操作`, nil
}
