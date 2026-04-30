package variable

import (
	"bytes"
	"context"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	_struct "dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_curl"
	"dev_tool/internal/pkg/p_db"
	"dev_tool/internal/pkg/p_shell"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
	"net/http"

	"github.com/tidwall/gjson"

	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// _dbQueryer unified db query interface
type _dbQueryer interface {
	QueryBySql(sql string, args ...any) *gsdb.SqlQuick
	ExecBySql(sql string, args ...any) *gsdb.SqlQuick
}

// llmRunConfig 大模型执行配置
type llmRunConfig struct {
	Provider     string  `json:"provider"`      // 服务商类型，当前仅支持 openai
	Model        string  `json:"model"`         // 模型名称
	SystemPrompt string  `json:"system_prompt"` // 系统提示词
	Prompt       string  `json:"prompt"`        // 用户提示词（来自 bash 字段）
	Temperature  float64 `json:"temperature"`   // 温度参数
	BaseURL      string  `json:"base_url"`      // 可选，覆盖全局 base_url
	ApiKey       string  `json:"api_key"`       // 可选，覆盖全局 api_key
}

type RCmd struct {
	cmd            map[string]any
	replaceList    map[string]string
	PlaywrightLock sync.RWMutex
	Sse            *p_sse.SseShell
	Call           *p_common.Call
}

func NewRCmd(cmd map[string]any, replace map[string]string, sse *p_sse.SseShell, call *p_common.Call) *RCmd {
	return &RCmd{
		cmd:         cmd,
		replaceList: replace,
		Sse:         sse,
		Call:        call,
	}
}

func (h *RCmd) RunMysql() error {
	cmdSql := cast.ToString(h.cmd[`sql`])
	resultKey := cast.ToString(h.cmd[`result_key`])
	//替换
	cmdSql = p_common.Replace(cmdSql, h.replaceList)
	//解析Id
	mysqlId, sql, err := component.VariableClient.ParseIdContent(cmdSql)
	if err != nil {
		return err
	}
	//检查是否还有未替换的
	if component.VariableClient.ExistReplaceParam(sql) {
		return errors.New(`还存在未替换的参数：` + sql)
	}
	//执行
	mysqlConfig, mysqlConfigErr := h.Call.QueryMysqlConfig(mysqlId)
	if mysqlConfigErr != nil {
		return errors.New(`找不到mysql配置 ` + mysqlConfigErr.Error())
	}
	dbType := cast.ToString(mysqlConfig[`db_type`])
	if dbType == `` {
		dbType = `mysql`
	}
	var dbClient _dbQueryer
	if dbType == `pgsql` {
		pgsqlClient, pgsqlErr := p_db.PgsqlClient.GetClient(mysqlConfig, h.Call)
		if pgsqlErr != nil {
			return errors.New(`获取pgsql client 失败 ` + pgsqlErr.Error())
		}
		dbClient = pgsqlClient
	} else {
		mysqlClient, mysqlErr := p_db.MysqlClient.GetClient(mysqlConfig, h.Call)
		if mysqlErr != nil {
			return errors.New(`获取mysql client 失败 ` + mysqlErr.Error())
		}
		dbClient = mysqlClient
	}
	result := ``
	h.Sse.Send(fmt.Sprintf(`%s %s 执行sql:`,
		p_common.TMarkDownClient.Bold(`run`),
		cast.ToString(h.cmd[`name`])) + "\n")
	if len(gstool.RegexSearchString(sql, "(?i)select")) > 0 {
		result = p_common.TMarkDownClient.Code(sql, `sql`)
		h.Sse.Send(result + "\n")
		all, allErr := dbClient.QueryBySql(sql).All()
		if allErr != nil {
			return allErr
		}
		//增加替换变量
		if resultKey != `` && len(all) > 0 {
			component.VariableClient.AddReplace(h.replaceList, resultKey, gstool.JsonEncode(all))
		}
		return nil
	} else if len(gstool.RegexSearchString(sql, "(?i)update")) > 0 {
		affectRows, execErr := dbClient.ExecBySql(sql).Exec()
		result = p_common.TMarkDownClient.Code("-- 更新数"+cast.ToString(affectRows)+"\n"+sql, `sql`)
		h.Sse.Send(result + "\n")
		if execErr != nil {
			return execErr
		}
	}
	return nil
}

func (h *RCmd) RunWindowsCmd() (string, error) {
	bat := cast.ToString(h.cmd[`bash`])
	//cmdId := cast.ToString(h.cmd[`id`])
	bat = p_common.Replace(bat, h.replaceList)
	makeRet, err := h.Call.QueryGlobalConfig(map[string]any{
		`key`: define.GlobalMake,
	})
	if err != nil {
		return ``, gstool.Error(`获取全局变量失败 %s`, err.Error())
	}
	if len(makeRet) > 0 {
		bat = gstool.SReplaces(bat, map[string]string{
			cast.ToString(makeRet[`key`]): cast.ToString(makeRet[`value`]),
		})
	}
	h.Sse.Send(fmt.Sprintf("执行：%s ", bat) + "\n")
	// 构建命令
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command(`cmd.exe`, `/C`, bat)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()
	if err != nil {
		return ``, gstool.Error("make 执行失败: %v\n %s", err, stderrStr)
	}
	h.Sse.Send(stdoutStr + "\n")
	h.Sse.Send(stderrStr + "\n")
	return ``, nil
}

func (h *RCmd) RunBash() (string, error) {
	cmdBash := cast.ToString(h.cmd[`bash`])
	cmdId := cast.ToString(h.cmd[`id`])
	cmdBash = p_common.Replace(cmdBash, h.replaceList)
	sshId, bash, parseIdErr := component.VariableClient.ParseIdContent(cmdBash)
	if parseIdErr != nil {
		gstool.FmtPrintlnLogTime(`解析配置失败`)
		return ``, parseIdErr
	}
	//如果脚本还有未替换的
	if component.VariableClient.ExistReplaceParam(bash) {
		gstool.FmtPrintlnLogTime(`执行的脚本存在未替换的参数`)
		return ``, gstool.Error("执行的脚本还存在需要替换的内容")
	}
	//注册ssh
	sshUniqueKey := p_common.TBaseClient.GetCombineKey(`variable`, sshId, `run`)
	sftpUniqueKey := p_common.TBaseClient.GetCombineKey(`variable`, sshId, `sftp`)
	//链接ssh
	preConnErr := component.VariableClient.PreConnSsh(sshId, sshUniqueKey, sftpUniqueKey, &p_sse.SseShell{
		Sse:             h.Sse.Sse,
		SseDistributeId: h.Sse.SseDistributeId,
	}, h.Call)
	if preConnErr != nil {
		gstool.FmtPrintlnLogTime(`链接失败 %s`, preConnErr.Error())
		return ``, gstool.Error(`链接失败 %s`, preConnErr.Error())
	}
	if !component.ShellClient.Exist(sshUniqueKey) || !component.ShellClient.Exist(sftpUniqueKey) {
		gstool.FmtPrintlnLogTime(`ssh连接未初始化`)
		return ``, errors.New(`ssh连接未初始化`)
	}
	//初始化
	sshConfig, sshConfigErr := h.Call.GetSshConfig(sshId)
	if sshConfigErr != nil {
		gstool.FmtPrintlnLogTime(`获取ssh配置失败 %s`, sshConfigErr.Error())
		return ``, sshConfigErr
	}
	var sshClientErr error
	var sshClient *gsssh.SshTerminal
	//家目录
	home := `/var/www`
	if cast.ToString(sshConfig[`home`]) != `` {
		home = cast.ToString(sshConfig[`home`])
	}
	variableDir := home + `/variable`
	//ssh
	sshClient, sshClientErr = component.ShellClient.GetClientMarkdown(sshConfig, sshUniqueKey, &p_sse.SseShell{
		Sse:             h.Sse.Sse,
		SseDistributeId: h.Sse.SseDistributeId,
	})
	if sshClientErr != nil {
		gstool.FmtPrintlnLogTime(`获取ssh client 失败 %s`, sshClientErr.Error())
		return ``, sshClientErr
	}
	h.Sse.Send(`开始将脚本上传到%s`+"\n", fmt.Sprintf(variableDir+`/variable_%s.sh`, cmdId))
	//sftp
	sshOnce, sshOnceErr := component.ShellClient.GetSshOnce(sshConfig)
	if sshOnceErr != nil {
		gstool.FmtPrintlnLogTime(`获取ssh once 失败 %s`, sshOnceErr.Error())
		return ``, sshOnceErr
	}
	var err error
	//创建目录
	_, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo mkdir -p %s`, variableDir), 40*time.Second)
	if err != nil {
		h.Sse.Send(`创建目录%s失败 %s`, variableDir, err.Error()+"\n")
		return ``, err
	}
	//写入脚本 用replace后不知道为什么打印日志没有问题，一执行echo就会重复写入几次 但是不执行h.replace又没有问题
	//_, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo echo '%s' > /var/www/variable/variable_%s.sh`, bash, cmdId))
	//if err != nil {
	//	return ``, err
	//}
	component.VariableClient.GetLog().Debugf(`%s \n %s `, fmt.Sprintf(variableDir+`/variable_%s.sh`, cmdId), bash)
	err = sshOnce.UploadFile(fmt.Sprintf(variableDir+`/variable_%s.sh`, cmdId), bash, ``)
	if err != nil {
		h.Sse.Send(`上传失败 %s %s`, fmt.Sprintf(variableDir+`/variable_%s.sh`, cmdId), err.Error()+"\n")
		return "", gstool.Error(`上传失败 %s %s`, fmt.Sprintf(variableDir+`/variable_%s.sh`, cmdId), err.Error())
	}
	h.Sse.Send(`上传成功 %s`+"\n", fmt.Sprintf(variableDir+`/variable_%s.sh`, cmdId))
	_, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo chmod +x %s/variable_%s.sh`, variableDir, cmdId), 40*time.Second)
	if err != nil {
		gstool.FmtPrintlnLogTime(`修改权限失败 %s`, err.Error())
		return ``, err
	}
	//var result string
	var result string
	result, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo %s/variable_%s.sh`, variableDir, cmdId), 40*time.Second)
	if err != nil {
		gstool.FmtPrintlnLogTime(`执行失败 %s`, err.Error())
		return ``, err
	}
	return result, nil
}

func (h *RCmd) RunUpload() (string, error) {
	cmdBash := cast.ToString(h.cmd[`bash`])
	//cmdId := cast.ToString(h.cmd[`id`])
	cmdBash = p_common.Replace(cmdBash, h.replaceList)
	sshId, bash, parseIdErr := component.VariableClient.ParseIdContent(cmdBash)
	if parseIdErr != nil {
		return ``, parseIdErr
	}
	//如果脚本还有未替换的
	if component.VariableClient.ExistReplaceParam(bash) {
		return ``, gstool.Error("上传的脚本还存在需要替换的内容")
	}
	//解析配置
	bash = gstool.SReplaces(bash, map[string]string{
		`\`: `\\`,
	})
	uploadConfig := make(map[string]any)
	deErr := gstool.JsonDecode(bash, &uploadConfig)
	if deErr != nil {
		h.Sse.Send(fmt.Sprintf(`解析上传配置失败 %s`, bash) + "\n")
		return ``, deErr
	}
	targetDir := cast.ToString(uploadConfig[`target_dir`])
	sourceFile := cast.ToString(uploadConfig[`source_file`])
	if targetDir == `` {
		h.Sse.Send(`目标目录不能为空` + "\n")
		return ``, gstool.Error(`目标目录不能为空`)
	}
	if sourceFile == `` {
		h.Sse.Send(`源文件不能为空` + "\n")
		return ``, gstool.Error(`源文件不能为空`)
	}
	//初始化
	sshConfig, sshConfigErr := h.Call.GetSshConfig(sshId)
	if sshConfigErr != nil {
		return ``, sshConfigErr
	}
	//sftp
	sshOnce, sshOnceErr := component.ShellClient.GetSshOnce(sshConfig)
	if sshOnceErr != nil {
		return ``, sshOnceErr
	}
	//如果是上传文件
	isErr := false
	if gstool.FileIsExisted(sourceFile) {
		h.Sse.Send(`本地存在文件：` + sourceFile + `，准备上传` + "\n")
		uploadErr := h.uploadFile(sshConfig, sshId, sshOnce, sourceFile, targetDir)
		if uploadErr != nil {
			return ``, uploadErr
		}
	} else {
		h.Sse.Send(`本地不存在文件,那么将上传整个文件夹：` + sourceFile + "\n")
		_ = gstool.DirWalk(sourceFile, func(path string, info os.FileInfo, err error) {
			if err != nil {
				return
			}
			if info.IsDir() {
				return
			}
			uploadErr := h.uploadFile(sshConfig, sshId, sshOnce, path, targetDir)
			if uploadErr != nil {
				component.VariableClient.GetLog().Errof(`上传失败`)
				h.Sse.Send(fmt.Sprintf(`上传失败 %s`, uploadErr.Error()) + "\n")
				isErr = true
				return
			}
		})
	}
	if isErr {
		return ``, gstool.Error(`上传失败`)
	}
	h.Sse.Send(`上传完成` + "\n")
	return ``, nil
}

func (h *RCmd) uploadFile(sshConfig map[string]any, sshId int, sshOnce *gsssh.SshOnce, sourceFile, targetDir string) error {
	var err error
	fileName := gstool.FileGetNameByPath(sourceFile)
	targetTempFileName := fileName + p_common.TBaseClient.GetUnique(`_upload`)
	targetTempFile := targetDir + `/` + targetTempFileName
	fileSizeMb, _ := gstool.FileSize(sourceFile, `mb`)
	h.Sse.Send(fmt.Sprintf(`准备上传文件 %s  %s 到目标文件 %s`, fileSizeMb, sourceFile, targetTempFile) + "\n")
	startTime := gstool.TimeNowUnixToString(`Y-m-d H:i:s`)
	h.Sse.Send(fmt.Sprintf(`[PROCESS]%s %s`, startTime, `上传进度:\s+\d+%\s+\(\d+\/\d+\s+bytes\)`))
	var lastPrintedStep int = -1
	err = sshOnce.UploadFileProcessScp(targetTempFile, sourceFile, func(bytesWritten, totalBytes int64) {
		// 计算当前进度百分比
		currentPercent := float64(bytesWritten) / float64(totalBytes) * 100
		currentStep := int(currentPercent) / 1 // 每1%为一个step

		// 只有当进入新的5%区间或完成时才打印
		if currentStep > lastPrintedStep || bytesWritten == totalBytes {
			h.Sse.Send(fmt.Sprintf("%s 上传进度: %d%% (%d/%d bytes)",
				startTime,
				currentStep*1, // 显示5%的整数倍
				bytesWritten,
				totalBytes) + "\n")

			lastPrintedStep = currentStep

			// 上传完成时换行
			if bytesWritten == totalBytes {
				h.Sse.Send(fmt.Sprintf("%s 上传进度: 100%% (%d/%d bytes)",
					startTime,
					bytesWritten,
					totalBytes) + "\n")
			}
		}
	})
	time.Sleep(time.Millisecond * 500)
	if err != nil {
		h.Sse.Send(fmt.Sprintf(`上传文件失败 %s`, err.Error()) + "\n")
		return err
	}
	//ssh
	sshUniqueKey := p_common.TBaseClient.GetCombineKey(`variable`, sshId, `run`)
	sshClient, sshClientErr := component.ShellClient.GetClientMarkdown(sshConfig, sshUniqueKey, &p_sse.SseShell{
		Sse:             h.Sse.Sse,
		SseDistributeId: h.Sse.SseDistributeId,
	})
	if sshClientErr != nil {
		h.Sse.Send(fmt.Sprintf(`上传文件失败2 %s`, sshClientErr.Error()) + "\n")
		return gstool.Error(`上传失败 %s`, sshClientErr.Error())
	}
	h.Sse.Send(fmt.Sprintf(`迁移%s %s`, targetTempFile, targetDir+`/`+fileName) + "\n")
	_, err = sshClient.RunCommandWait(fmt.Sprintf(`sudo mv %s %s`, targetTempFile, targetDir+`/`+fileName), 40*time.Second)
	if err != nil {
		h.Sse.Send(fmt.Sprintf(`迁移失败 %s`, err.Error()) + "\n")
		return gstool.Error(`迁移失败 %s`, err.Error())
	}
	return nil
}

func (h *RCmd) RunCommand() (string, error) {
	cmdBash := cast.ToString(h.cmd[`bash`])
	cmdBash = p_common.Replace(cmdBash, h.replaceList)
	sshId, bash, parseIdErr := component.VariableClient.ParseIdContent(cmdBash)
	if parseIdErr != nil {
		return ``, parseIdErr
	}
	//注册client
	sshUniqueKey := p_common.TBaseClient.GetCombineKey(`variable`, sshId, `run`)
	sftpUniqueKey := p_common.TBaseClient.GetCombineKey(`variable`, sshId, `sftp`)
	preConnErr := component.VariableClient.PreConnSsh(sshId, sshUniqueKey, sftpUniqueKey, &p_sse.SseShell{
		Sse:             h.Sse.Sse,
		SseDistributeId: h.Sse.SseDistributeId,
	}, h.Call)
	if preConnErr != nil {
		return ``, gstool.Error(`链接失败 %s`, preConnErr.Error())
	}
	//分离出来多行命令
	commandList := strings.Split(bash, "\n")
	for _, command := range commandList {
		if command == "" {
			continue
		}
		sshUniqueKey := p_common.TBaseClient.GetCombineKey(`variable`, sshId, `run`)
		if !component.ShellClient.Exist(sshUniqueKey) {
			return ``, errors.New(`ssh连接未初始化`)
		}
		sshConfig, sshConfigErr := h.Call.GetSshConfig(sshId)
		if sshConfigErr != nil {
			return ``, sshConfigErr
		}
		var sshClientErr error
		var sshClient *gsssh.SshTerminal
		//ssh
		sshClient, sshClientErr = component.ShellClient.GetClientMarkdown(sshConfig, sshUniqueKey, &p_sse.SseShell{
			Sse:             h.Sse.Sse,
			SseDistributeId: h.Sse.SseDistributeId,
		})
		if sshClientErr != nil {
			return ``, sshClientErr
		}
		var err error
		runCmd := p_shell.Command{}
		runCmd.SetCommand(command)
		runCmd.Sudo()
		h.Sse.Send(p_common.TMarkDownClient.Code(runCmd.GetCommand().ToStr(), `bash`) + "\n")
		_, err = sshClient.RunCommandWait(runCmd.GetCommand().ToStr(), 40*time.Second)
		if err != nil {
			return ``, err
		}
	}
	return ``, nil
}

func (h *RCmd) RunCurl() (string, error) {
	resultKey := cast.ToString(h.cmd[`result_key`])
	parseConfig := _struct.CurlParseConfig{}
	//解析配置
	options := cast.ToString(h.cmd[`options`])
	err := gstool.JsonDecode(options, &parseConfig)
	if err != nil {
		h.Sse.Send(`解析失败 ` + options + "\n")
		return ``, err
	}
	//对url进行替换
	parseConfig.Url = p_common.Replace(parseConfig.Url, h.replaceList)
	//如果是openai的curl请求格式
	isOpenAi := false
	parseConfig.Body, isOpenAi, err = p_common.ReplaceOpenAiBody(parseConfig.Body, h.replaceList, h.Sse)
	if err != nil {
		return ``, err
	}
	pCurl := p_curl.CurlRun{
		ParseConfig: parseConfig,
		CurlEvents: _struct.CurlEvents{
			StreamDataCall: func(s string) {
				if isOpenAi {
					message := p_common.ExtractOpenAiMessage(s)
					if message != `` {
						h.Sse.Send(message)
					}
				} else {
					h.Sse.Send(s)
				}
			},
			NoticeCall: func(s string) {
				h.Sse.Send(fmt.Sprintf(`%s`, s) + "\n")
			},
			EndCall: func() {
				//h.Sse.Send(fmt.Sprintf(`结束请求 %s`, parseConfig.Uri) + "\n")
			},
			StartCall: func() {
				h.Sse.Send("\n***\n")
			},
		},
	}
	result, err := pCurl.Run()
	//增加替换变量
	if resultKey != `` {
		component.VariableClient.AddReplace(h.replaceList, resultKey, cast.ToString(result))
	}
	return cast.ToString(result), err
}

// parseLlmRunConfig 解析大模型命令配置
func parseLlmRunConfig(cmd map[string]any) (llmRunConfig, error) {
	cfg := llmRunConfig{
		Provider: cast.ToString(cmd[`provider`]),
		Model:    cast.ToString(cmd[`smart_link_label`]),
		Prompt:   cast.ToString(cmd[`bash`]),
	}
	options := cast.ToString(cmd[`options`])
	if options != `` {
		if err := gstool.JsonDecode(options, &cfg); err != nil {
			return cfg, errors.New(`解析大模型配置失败: ` + err.Error())
		}
	}
	if cfg.Provider == `` {
		cfg.Provider = `openai`
	}
	if strings.TrimSpace(cfg.Model) == `` {
		return cfg, errors.New(`模型不能为空`)
	}
	if strings.TrimSpace(cfg.Prompt) == `` {
		return cfg, errors.New(`提示词不能为空`)
	}
	return cfg, nil
}

// RunLlm 请求大模型并将结果写入 SSE/替换变量
func (h *RCmd) RunLlm() (string, error) {
	cfg, err := parseLlmRunConfig(h.cmd)
	if err != nil {
		return ``, err
	}
	if strings.ToLower(cfg.Provider) != `openai` {
		return ``, errors.New(`当前仅支持 openai 服务商`)
	}
	cfg.Prompt = p_common.Replace(cfg.Prompt, h.replaceList)
	cfg.SystemPrompt = p_common.Replace(cfg.SystemPrompt, h.replaceList)

	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == `` {
		baseURLConf, qErr := h.Call.QueryGlobalConfig(map[string]any{
			`key`: `{openai_base_url}`,
		})
		if qErr != nil {
			return ``, errors.New(`查询 openai_base_url 失败: ` + qErr.Error())
		}
		baseURL = strings.TrimSpace(cast.ToString(baseURLConf[`value`]))
	}
	if baseURL == `` {
		baseURL = `https://api.openai.com`
	}
	baseURL = resolveOpenAIChatCompletionsURL(baseURL)

	apiKey := strings.TrimSpace(cfg.ApiKey)
	if apiKey == `` {
		apiKeyConf, qErr := h.Call.QueryGlobalConfig(map[string]any{
			`key`: `{openai_api_key}`,
		})
		if qErr != nil {
			return ``, errors.New(`查询 openai_api_key 失败: ` + qErr.Error())
		}
		apiKey = strings.TrimSpace(cast.ToString(apiKeyConf[`value`]))
	}
	if apiKey == `` {
		return ``, errors.New(`openai_api_key 不能为空，请先在全局配置中设置`)
	}

	messages := make([]map[string]string, 0, 2)
	if strings.TrimSpace(cfg.SystemPrompt) != `` {
		messages = append(messages, map[string]string{
			`role`:    `system`,
			`content`: cfg.SystemPrompt,
		})
	}
	messages = append(messages, map[string]string{
		`role`:    `user`,
		`content`: cfg.Prompt,
	})
	bodyMap := map[string]any{
		`model`:    cfg.Model,
		`messages`: messages,
	}
	if cfg.Temperature > 0 {
		bodyMap[`temperature`] = cfg.Temperature
	}

	parseConfig := _struct.CurlParseConfig{
		Url:         baseURL,
		Method:      http.MethodPost,
		ContentType: define.ContentTypeJson,
		Body:        gstool.JsonEncode(bodyMap),
		Headers: map[string]string{
			`Authorization`: `Bearer ` + apiKey,
			`Content-Type`:  `application/json`,
		},
	}
	h.Sse.Send(fmt.Sprintf(`%s %s`, p_common.TMarkDownClient.Bold(`run llm`), cfg.Model) + "\n")
	pCurl := p_curl.CurlRun{
		ParseConfig: parseConfig,
		CurlEvents: _struct.CurlEvents{
			NoticeCall: func(s string) {
				h.Sse.Send(s + "\n")
			},
		},
	}
	result, runErr := pCurl.Run()
	if runErr != nil {
		return ``, runErr
	}
	content := p_common.ExtractOpenAiMessage(cast.ToString(result))
	if content == `` {
		content = cast.ToString(result)
	} else {
		h.Sse.Send(content + "\n")
	}
	resultKey := cast.ToString(h.cmd[`result_key`])
	if resultKey != `` {
		component.VariableClient.AddReplace(h.replaceList, resultKey, content)
	}
	return content, nil
}

func resolveOpenAIChatCompletionsURL(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == `` {
		baseURL = `https://api.openai.com`
	}
	if strings.Contains(baseURL, `://`) {
		schemeSplit := strings.SplitN(baseURL, `://`, 2)
		if len(schemeSplit) == 2 {
			pathIndex := strings.Index(schemeSplit[1], `/`)
			if pathIndex >= 0 {
				return strings.TrimRight(baseURL, `/`)
			}
		}
	}
	if strings.Contains(baseURL, `/chat/completions`) {
		return strings.TrimRight(baseURL, `/`)
	}
	return strings.TrimRight(baseURL, `/`) + `/v1/chat/completions`
}

func (h *RCmd) RunPlaywright(stopCall func() bool) (string, error) {
	id := cast.ToInt(h.cmd[`smart_link_id`])
	label := cast.ToString(h.cmd[`smart_link_label`])
	if id == 0 {
		return ``, errors.New(`链接ID不能为空`)
	}
	if label == `` {
		return ``, errors.New(`链接label不能为空`)
	}
	runParams, runParamsErr := plw.GetRunParams(id, label, ``, ``, 0, 0, h.replaceList)
	if runParamsErr != nil {
		return ``, errors.New(runParamsErr.Error())
	}
	sse := &p_sse.SseShell{
		Sse:             h.Sse.Sse,
		SseDistributeId: h.Sse.SseDistributeId,
	}
	//注册链接执行时需要输出的文本类型
	runParams.RunCallFunc = func(cmdType define.ProcessType, errmsg, tip, content string) {
		switch cmdType {
		case define.Input:
			//h.Sse.Send(p_common.TMarkDownClient.Bold(tip)+`,`+content+` `+errmsg+ "\n")
		case define.CanvasImage:
			//h.Sse.Send(p_common.TMarkDownClient.Bold(tip)+`,`+errmsg+ "\n")
			sse.Send(content + "\n")
		case define.ExistWait, define.NoExistWait:
			//h.Sse.Send(p_common.TMarkDownClient.Bold(tip)+`,`+errmsg+ "\n")
		case define.LoginUsernamePassword: //前端弹窗输入账号密码
			component.VariableClient.ClearLoginCredentials()
			sse.Send(define.SseEventLogin)
			for i := 0; i < 30; i++ {
				time.Sleep(time.Second * 2)
				if component.VariableClient.GetLoginUsername() != `` && component.VariableClient.GetLoginPassword() != `` {
					break
				}
			}
			h.replaceList[`{user_name}`] = component.VariableClient.GetLoginUsername()
			h.replaceList[`{password}`] = component.VariableClient.GetLoginPassword()
		}
	}
	//注册需要监听的接口
	listenUriList := cast.ToString(h.cmd[`options`])
	if listenUriList != `` {
		parseConfigs := make([]_struct.CurlParseConfig, 0)
		_ = gstool.JsonDecode(listenUriList, &parseConfigs)
		for _, parseConfig := range parseConfigs {
			uri := parseConfig.Uri
			if uri == `` {
				continue
			}
			runParams.ListenCurls[uri] = p_curl.NewCurlRun(parseConfig, _struct.CurlEvents{
				StreamDataCall: func(s string) {
					h.StreamDataReceive(sse, parseConfig, s)
				},
				NoticeCall: func(s string) {
					sse.Send(s + "\n")
				},
				EndCall: func() {
					sse.Send("\n" + p_common.TMarkDownClient.Bold("end ") + "\n")
				},
				StartCall: func() {
					runParams.StopEchoTips = true
					component.VariableClient.GetLog().Debugf(`监听到%s`, parseConfig.Uri)
					sse.Send(p_common.TMarkDownClient.BlockQuote("开始回答...") + "\n")
				},
			})
		}
	}

	sse.Send("\n" + p_common.TMarkDownClient.Bold(label) + `,启动` + "\n")
	streamFunc := func(name, msg string) {
		if runParams.StopEchoTips {
			return
		}
		sse.Send(p_common.TMarkDownClient.Bold(name) + `,` + msg + "\n")
	}
	runParams.StreamFunc = streamFunc
	p := plw.NewPlaywright(runParams, component.VariableClient.GetLog())
	openErr := p.Open(h.Call, stopCall)
	if openErr != nil {
		sse.Send(p_common.TMarkDownClient.BlockQuote(cast.ToString(h.cmd[`name`])+`,启动失败，`+openErr.Error()) + "\n")
		return ``, openErr
	}
	return ``, nil
}

// StreamDataReceive 流式结果解析
func (h *RCmd) StreamDataReceive(sse *p_sse.SseShell, parseConfig _struct.CurlParseConfig, msg string) {
	if parseConfig.IsStream == 1 { //流式
		if len(parseConfig.TakeJsons) > 0 { //配置了提取规则
			jsonLists := gstool.JsonParseFromStr([]byte(msg))
			for _, part := range jsonLists {
				if part == `` {
					continue
				}
				for _, takeJson := range parseConfig.TakeJsons {
					realTakeJson, _ := strings.CutPrefix(takeJson.Take, `res.`)
					ret := gjson.Get(part, realTakeJson)
					component.VariableClient.GetLog().Debugf(`提取json成功#%s#%v`, part, ret.String())
					if ret.String() != `` { //发送到sse
						component.VariableClient.GetLog().Debugf(`发送到sse#%s#`, ret.String())
						sse.Send(ret.String())
					}
				}
			}
		} else {
			sse.Send(cast.ToString(msg))
		}
	} else {
		sse.Send(cast.ToString(msg))
	}
}

func (h *RCmd) RunCombine() (string, error) {
	resultKey := cast.ToString(h.cmd[`result_key`])
	combine := p_common.Replace(cast.ToString(h.cmd[`options`]), h.replaceList)

	//增加替换变量
	if resultKey != `` {
		component.VariableClient.AddReplace(h.replaceList, resultKey, combine)
	}
	h.Sse.Send(p_common.TMarkDownClient.Bold(`合并内容`) + `,` + combine + "\n")
	return ``, nil
}

func (h *RCmd) RunRedis() (string, error) {
	name := cast.ToString(h.cmd[`name`])
	cmdBash := p_common.Replace(cast.ToString(h.cmd[`bash`]), h.replaceList)
	redisId, redisBash, parseErr := component.VariableClient.ParseIdContent(cmdBash)
	if parseErr != nil {
		return ``, errors.New(`redis解析失败` + parseErr.Error())
	}
	if redisBash == `` {
		return ``, errors.New(`redis需要删除的key不能为空`)
	}
	redisConfig, redisConfigErr := h.Call.GetRedisConfig(redisId)
	if redisConfigErr != nil {
		return ``, redisConfigErr
	}
	client, clientErr := component.RedisClient.GetClient(redisConfig, h.Call)
	if clientErr != nil {
		return "", clientErr
	}
	h.Sse.Send(name + `,` + redisBash + "\n")
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
