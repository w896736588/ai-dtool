package variable

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

type TVariable struct {
	TaskList      map[string]string
	SshClientList map[string][]string
	lock          sync.RWMutex
	Log           *gstool.GsSlog
	//临时输入的账号密码
	LoginUsername string
	LoginPassword string
}

var VariableClient *TVariable

func NewVariableClient() *TVariable {
	log := gstool.NewSlog3(component.EnvClient.LogPath, `variable`)
	_ = log.CleanOldLogs(2)
	return &TVariable{
		TaskList:      make(map[string]string),
		SshClientList: make(map[string][]string),
		Log:           log,
	}
}

func (h *TVariable) StopAll() {
	h.lock.Lock()
	defer h.lock.Unlock()
	for k, _ := range h.TaskList {
		h.TaskList[k] = "stop"
		if clientList, ok := h.SshClientList[k]; ok {
			for _, clientId := range clientList {
				component.ShellClient.RmClient(clientId)
			}
			delete(h.SshClientList, k)
		}
	}
}

func (h *TVariable) StopOther(runUniqueId string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	for k, _ := range h.TaskList {
		if k == runUniqueId {
			continue
		}
		h.TaskList[k] = "stop"
		if clientList, ok := h.SshClientList[k]; ok {
			for _, clientId := range clientList {
				component.ShellClient.RmClient(clientId)
			}
			delete(h.SshClientList, k)
		}
	}
	time.Sleep(time.Second)
}

func (h *TVariable) Add(id string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.TaskList[id] = "run"
}

func (h *TVariable) Del(id string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.TaskList, id)
}

func (h *TVariable) Get(id string) string {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.TaskList[id]
}

func (h *TVariable) AddSshClient(id, clientId string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if clientList, ok := h.SshClientList[id]; ok {
		clientList = append(clientList, clientId)
		h.SshClientList[id] = clientList
	} else {
		clientList = []string{clientId}
		h.SshClientList[id] = clientList
	}
}

func (h *TVariable) DelSshClient(id string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.SshClientList, id)
}

// WaitReplace 是否属于待替换的字符串
// 格式为：{run_aaa}或者{run_type}.value
func (h *TVariable) WaitReplace(str string) bool {
	return gstool.RegexMatchString(str, `{[a-zA-Z0-9_]+}[.]*[a-zA-Z0-9_]*`)
}

// GetReplaceVal 获取待替换的值
func (h *TVariable) GetReplaceVal(replaceKey string, replaceList *[]map[string]string) (string, bool) {
	for _, replaceMap := range *replaceList {
		if value, ok := replaceMap[replaceKey]; ok {
			return value, true
		}
	}
	return ``, false
}

// ExistReplaceParam 是否存在待替换的变量
func (h *TVariable) ExistReplaceParam(data string) bool {
	return gstool.RegexMatchString(data, `{[a-zA-Z0-9_]+}`)
}

// ExistReplaceParamFull 是否存在待替换的变量Full 匹配整行
func (h *TVariable) ExistReplaceParamFull(data string) bool {
	return gstool.RegexMatchString(data, `^{[a-zA-Z0-9_]+}$`)
}

// ExistConfigParamFull 是否存在待获取配置的变量Full 例如:{config:ssh} 获取ssh配置
func (h *TVariable) ExistConfigParamFull(data string) bool {
	return gstool.RegexMatchString(data, `^{[a-zA-Z0-9_:]+}$`)
}

// ParseIdContent 解析sql或者bash脚本第一行定义的id，格式：[RunUniqueId=1]
func (h *TVariable) ParseIdContent(str string) (int, string, error) {
	re := regexp.MustCompile(`[\r\n]`)
	sqlParamList := re.Split(str, -1)
	//过滤掉空行
	sqlParamList = gstool.ArrayFilterEmpty(&sqlParamList)
	content := strings.Join(sqlParamList[1:], "\n")
	baseId := sqlParamList[0]
	id := gstool.SReplaces(baseId, map[string]string{
		`[id=`: ``,
		`]`:    ``,
	})
	cId := cast.ToInt(id)
	if cId == 0 {
		return cId, content, errors.New(`id不能为空 ` + str + ` ` + gstool.JsonFormat(sqlParamList))
	}
	return cId, content, nil
}

// AddReplace 增加替换变量
func (h *TVariable) AddReplace(replaceList map[string]string, key, value string) {
	if key == `` {
		return
	}
	replaceList[key] = value
}

func (h *TVariable) RegisterAllGlobal(replaceList map[string]string, sse *p_sse.SseShell, call *p_common.Call) {
	allGlobalList, err := call.AllGlobal()
	if err != nil {
		gstool.FmtPrintlnLogTime(`query all global error %s`, err.Error())
		return
	}
	for _, globalMap := range allGlobalList {
		if existVal, ok := replaceList[cast.ToString(globalMap[`key`])]; ok {
			if cast.ToString(globalMap[`value`]) == existVal {
				continue
			}
		}
		h.AddReplace(replaceList, cast.ToString(globalMap[`key`]), cast.ToString(globalMap[`value`]))
		sse.Send(`注入全局常量 ` + fmt.Sprintf(`%s,%s`, cast.ToString(globalMap[`key`]), cast.ToString(globalMap[`value`])) + "\n")
	}
}

// ChecksCanDo 检查是否需要执行 true 可以执行 false不可以执行
func (h *TVariable) ChecksCanDo(cmd map[string]any) bool {
	checks := cast.ToString(cmd[`checks`])
	if checks == `` {
		return true
	}
	enquire := ` = `
	notEnquire := ` != `
	in := ` in `
	notIn := ` not in `
	//等于
	if strings.Contains(checks, enquire) {
		checkList := strings.Split(checks, enquire)
		if len(checkList) != 2 { //不是两个条件 那么就返回不显示 格式不对
			return false
		}
		realCheck0 := checkList[0]
		realCheck1 := checkList[1]
		//匹配上了 那么返回不禁用
		if realCheck0 == realCheck1 {
			return true
		}
		//禁显示
		return false
	} else if strings.Contains(checks, notEnquire) {
		checkList := strings.Split(checks, notEnquire)
		if len(checkList) != 2 { //不是两个条件 那么就返回不显示 格式不对
			return false
		}
		realCheck0 := checkList[0]
		realCheck1 := checkList[1]
		//匹配上了 那么返回不禁用
		if realCheck0 != realCheck1 {
			return true
		}
		//禁显示
		return false
	} else if strings.Contains(checks, in) {
		checkList := strings.Split(checks, in)
		if len(checkList) != 2 { //不是两个条件 那么就返回不显示 格式不对
			return false
		}
		realCheck0 := checkList[0]
		realCheckList := strings.Split(checkList[1], `,`)
		//匹配上了 那么返回不禁用
		if gstool.ArrayExistValue(&realCheckList, realCheck0) {
			return true
		}
		//禁显示
		return false
	} else if strings.Contains(checks, notIn) {
		checkList := strings.Split(checks, notIn)
		if len(checkList) != 2 { //不是两个条件 那么就返回不显示 格式不对
			return false
		}
		realCheck0 := checkList[0]
		realCheckList := strings.Split(checkList[1], `,`)
		//匹配上了 那么返回不禁用
		if !gstool.ArrayExistValue(&realCheckList, realCheck0) {
			return true
		}
		//禁显示
		return false
	}
	return false
}

// PreConnSsh 初始化ssh连接
func (h *TVariable) PreConnSsh(sshId int, sshUniqueKey, sftpUniqueKey string, sse *p_sse.SseShell, call *p_common.Call) error {
	if sshId == 0 {
		return errors.New(`ssh_id不能为空`)
	}

	if component.ShellClient.Exist(sshUniqueKey) && component.ShellClient.Exist(sftpUniqueKey) {
		return nil
	}
	//初始化连接
	sshConfig, sshConfigErr := call.GetSshConfig(sshId)
	if sshConfigErr != nil {
		return sshConfigErr
	}
	//ssh
	_, sshClientErr := component.ShellClient.GetClientMarkdown(sshConfig, sshUniqueKey, sse)
	if sshClientErr != nil {
		return sshClientErr
	}
	//sftp
	_, sftpClientErr := component.ShellClient.GetClientMarkdown(sshConfig, sftpUniqueKey, sse)
	if sftpClientErr != nil {
		return sftpClientErr
	}
	return nil
}

// SelectChooseReplace 单选选中后替换
func (h *TVariable) SelectChooseReplace(variableForm *_struct.VForm,
	replaceList map[string]string, chooseValue string) {

	//gstool.FmtPrintlnLogTime(`resultKey %s`, variableForm.ResultKey)
	//gstool.FmtPrintlnLogTime(`chooseValue %s`, chooseValue)
	//gstool.FmtPrintlnLogTime(`option list %s`, gstool.JsonEncode(variableForm.Select.OptionList))

	for _, option := range variableForm.Select.OptionList {
		if variableForm.ResultKey != `` && chooseValue != `` && chooseValue == option.Value {
			sourceOptionList := make(map[string]any, 0)
			_ = gstool.JsonDecode(option.Source, &sourceOptionList)
			for optionKey, optionValue := range sourceOptionList {
				//这里本来预计为{product}.label这种替换，但是因为map是无序的，所以循环替换的时候会导致{product}.label只会被替换掉{product}
				h.AddReplace(replaceList, variableForm.ResultKey+`.`+optionKey, cast.ToString(optionValue))
				//再增加一个替换的{product.label} 防止前面说的情况
				h.AddReplace(replaceList, strings.Replace(variableForm.ResultKey, `}`, `.`+optionKey+`}`, 1), cast.ToString(optionValue))
			}
			//替换整体
			h.AddReplace(replaceList, variableForm.ResultKey, option.Source)
		}
	}
}

// ParseConfig 自带的配置查询解析
func (h *TVariable) ParseConfig(config string, call *p_common.Call) (string, error) {
	if config == `{config:ssh}` {
		sshList, sshErr := call.GetAllSshConfig()
		if sshErr == nil {
			for k, sshMap := range sshList {
				sshMap[`value`] = sshMap[`id`]
				sshMap[`label`] = sshMap[`name`]
				sshList[k] = sshMap
			}
			return gstool.JsonEncode(sshList), nil
		}
	} else if gstool.RegexMatchString(config, `{config:account_group:\w*}`) { //账号列表
		retList := gstool.RegexMatchSubString(config, `{config:account_group:(\w*)}`)
		if len(retList) != 2 {
			return ``, gstool.Error(`获取配置失败 %s`, config)
		}
		//查询组
		accountGroup, err := call.QueryGroupInfo(map[string]any{
			`type`: define.GroupTypeAccount,
			`name`: retList[1],
		})
		if err != nil {
			return ``, gstool.Error(`获取组配置失败 %s`, err.Error())
		}
		accountList, accountListErr := call.QueryAccountList(map[string]any{
			`account_group_id`: cast.ToInt(accountGroup[`id`]),
		})
		if accountListErr == nil {
			for k, accountMap := range accountList {
				accountMap[`value`] = accountMap[`id`]
				accountMap[`label`] = accountMap[`username`]
				accountList[k] = accountMap
			}
			return gstool.JsonEncode(accountList), nil
		}
	}
	return config, nil
}
