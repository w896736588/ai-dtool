package base

import (
	"dev_tool/base_module"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

type TCode struct {
}

func (h *TCode) FindCode(sshConfig map[string]any, dirPath string) []string {
	codeDirList := make([]string, 0)
	command := base_module.Command{}
	command.Sudo()
	command.FindGitDir(dirPath, 2)
	uniqueKey := Component.TBase.GetCombineKey(sshConfig[`id`], `code`)
	client, err := Component.TShell.GetClient(sshConfig, uniqueKey)
	if err != nil {
		gstool.FmtPrintlnLogTime(`连接ssh失败 %s`, err.Error())
		return codeDirList
	}
	defer Component.TShell.RmClient(uniqueKey)
	defer client.Close()
	ret, retErr := client.RunCommandWait(command.GetCommand().ToStr())
	if retErr != nil {
		gstool.FmtPrintlnLogTime(`执行命令失败 %s`, retErr.Error())
	} else {
		codeDirList = append(codeDirList, strings.Split(ret, "\n")...)
	}
	return codeDirList
}
