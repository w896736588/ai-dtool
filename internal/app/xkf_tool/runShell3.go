package xkf_tool

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"sync"
	"time"
)

var RunShell3Map map[string]*gstool.GsShell
var RunShell3MapLock sync.RWMutex
var RunShell3TerminalMap map[string]*gstool.GsShell

//GetRunShell3CliTer 获取runShell3
func GetRunShell3CliTer(sshConfig *SshConfig) *gstool.GsShell {
	RunShell3MapLock.Lock()
	defer RunShell3MapLock.Unlock()
	if sshConfig.Host == `` {
		return nil
	}

	uniKey := GetSshUnikey(sshConfig)
	if RunShell3Map[uniKey] == nil {
		gsShellTerConfig := gstool.GsShellConfig{
			Host:          sshConfig.Host,
			Port:          cast.ToInt64(sshConfig.Port),
			Username:      sshConfig.Username,
			Password:      sshConfig.Password,
			TimeoutSecond: 100,
		}
		cliTerConf := gstool.GsShell{
			Config:              &gsShellTerConfig,
			IsOpenLog:           true,
			Logger:              Logger,
			TerminalRefreshTime: time.Second,
			TerminalMaxTime:     15 * time.Second,
		}
		cliTerConfErr := cliTerConf.CreateClient()
		if cliTerConfErr != nil {
			panic(`创建交互式链接失败 ` + cliTerConfErr.Error())
		} else {
			RunShell3TerminalMap[uniKey] = &cliTerConf
		}
	}
	return RunShell3TerminalMap[uniKey]
}

func InitRunShell3(reqBody *SshExec) (*gstool.GsShell, *gstool.GsShell) {
	xkfRunShell3 := GetRunShell3CliTer(&reqBody.SshConfig)
	xkfRunShell4 := GetRunShell3CliTer(&reqBody.WkSshConfig)
	return xkfRunShell3, xkfRunShell4
}
