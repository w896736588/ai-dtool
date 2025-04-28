package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"strings"
	"sync"
	"time"
)

type TVariable struct {
	TaskList map[string]string
	lock     sync.RWMutex
	Log      *gstool.GsSlog
}

func NewVariable() *TVariable {
	return &TVariable{
		TaskList: make(map[string]string),
		Log:      gstool.NewSlogDefault(Component.Env.LogPath, `variable`),
	}
}

func (h *TVariable) StopAll() {
	h.lock.Lock()
	defer h.lock.Unlock()
	for k, _ := range h.TaskList {
		h.TaskList[k] = "stop"
	}
	time.Sleep(1) //等待1秒 把其他任务的输出断开玩
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

func (h *TVariable) CmdList(variableId any) ([]map[string]any, error) {
	return Component.TSqlite.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`variable_id`: variableId,
		`status`:      1,
	}).Order(`weight asc`).All()
}

func (h *TVariable) CmdInfo(cmdId any) (map[string]any, error) {
	return Component.TSqlite.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`id`:     cmdId,
		`status`: 1,
	}).One()
}

func (h *TVariable) Variable(variableId any) map[string]any {
	variableInfo, _ := Component.TSqlite.Client.QuickQuery(`tbl_variable`, `*`, map[string]interface{}{
		`id`: variableId,
	}).One()
	return variableInfo
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

// ParseConfig 自带的配置查询解析
func (h *TVariable) ParseConfig(config string) (string, error) {
	if config == `{config:ssh}` {
		sshList, sshErr := Component.TSqlite.GetAllSshConfig()
		if sshErr == nil {
			for k, sshMap := range sshList {
				sshMap[`value`] = sshMap[`id`]
				sshMap[`label`] = sshMap[`name`]
				sshList[k] = sshMap
			}
			return gstool.JsonEncode(sshList), nil
		}
	}
	return ``, errors.New(`不支持的解析项：` + config)
}

// Replace 替换变量
func (h *TVariable) Replace(data string, replaceList *[]map[string]string) string {
	for _, replace := range *replaceList {
		//处理特殊情况
		for replaceKey, replaceVal := range replace {
			//取模
			matchSubList := gstool.RegexMatchSubString(data, replaceKey+`%(\d+)`)
			if len(matchSubList) >= 2 {
				data = gstool.SReplaces(data, map[string]string{
					matchSubList[0]: cast.ToString(cast.ToInt64(replaceVal) % cast.ToInt64(matchSubList[1])),
				})
			}
		}
		data = gstool.SReplaces(data, replace)
	}
	return data
}

// ParseIdContent 解析sql或者bash脚本第一行定义的id，格式：[RunUniqueId=1]
func (h *TVariable) ParseIdContent(str string) (int, string, error) {
	sqlParamList := strings.Split(str, "\n")
	content := gstool.SReplaces(str, map[string]string{
		sqlParamList[0] + "\n": ``,
	})
	baseId := sqlParamList[0]
	id := gstool.SReplaces(baseId, map[string]string{
		`[id=`: ``,
		`]`:    ``,
	})
	cId := cast.ToInt(id)
	if cId == 0 {
		return cId, content, errors.New(`id不能为空`)
	}
	return cId, content, nil
}

// AddReplace 增加替换变量
func (h *TVariable) AddReplace(replaceList *[]map[string]string, key, value string) {
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

// ChecksCanDo 检查是否需要执行 true 可以执行 false不可以执行
func (h *TVariable) ChecksCanDo(cmd map[string]any) bool {
	checks := cast.ToString(cmd[`checks`])
	if checks == `` {
		return true
	}
	//等于
	if strings.Contains(checks, `=`) {
		checkList := strings.Split(checks, `=`)
		gstool.FmtPrintlnLogTime(`判断checks %s`, gstool.JsonEncode(checkList))
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
	} else if strings.Contains(checks, `!=`) {
		checkList := strings.Split(checks, `!=`)
		gstool.FmtPrintlnLogTime(`判断checks %s`, gstool.JsonEncode(checkList))
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
	}
	return false
}

// PreConnSsh 初始化ssh连接
func (h *TVariable) PreConnSsh(sshId int) error {
	if sshId == 0 {
		return errors.New(`ssh_id不能为空`)
	}
	sshUniqueKey := Component.TBase.GetCombineKey(`variable`, sshId, `run`)
	sftpUniqueKey := Component.TBase.GetCombineKey(`variable`, sshId, `sftp`)
	if Component.TShell.Exist(sshUniqueKey) && Component.TShell.Exist(sftpUniqueKey) {
		return nil
	}
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
	return nil
}

// SelectChooseReplace 单选选中后替换
func (h *TVariable) SelectChooseReplace(variableForm *_struct.VariableForm,
	replaceList *[]map[string]string, chooseValue string) {
	for _, option := range variableForm.Select.OptionList {
		if variableForm.ResultKey != `` && chooseValue != `` && chooseValue == option.Value {
			sourceOptionList := make(map[string]any, 0)
			_ = gstool.JsonDecode(option.Source, &sourceOptionList)
			for optionKey, optionValue := range sourceOptionList {
				h.AddReplace(replaceList, variableForm.ResultKey+`.`+optionKey, cast.ToString(optionValue))
			}
			//替换整体
			h.AddReplace(replaceList, variableForm.ResultKey, option.Source)
		}
	}
}
