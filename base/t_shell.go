package base

import (
	"errors"
	"gitee.com/Sxiaobai/gs/gsssh"
	"github.com/spf13/cast"
	"sync"
)

type TShell struct {
	ShellClientMap map[string]*gsssh.SshConfig
	lock           sync.Mutex
}

func (h *TShell) GetClient(sshConfig map[string]any, uniqueKey string) (*gsssh.SshConfig, error) {
	defer h.lock.Unlock()
	h.lock.Lock()
	sshId := cast.ToString(sshConfig[`id`])
	if sshId == `` {
		return nil, errors.New(`ssh配置错误`)
	}
	if shell, ok := h.ShellClientMap[uniqueKey]; ok {
		return shell, nil
	}
	gsShell := &gsssh.SshConfig{
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
		GsSlog:   Component.GsLog,
	}
	createErr := gsShell.ConnectAuthPassword()
	if createErr != nil {
		return nil, createErr
	}
	_, err := gsShell.RunCommandWait(`pwd`)
	if err != nil {
		return nil, err
	}
	//设置回调
	gsShell.SetFuncBefore(func(command string) string {
		return `■■ ` + command
	})
	gsShell.SetCombineNum(1)
	gsShell.CloseFirstReceiveMsg()
	h.ShellClientMap[uniqueKey] = gsShell
	return gsShell, nil
}

func (h *TShell) Exist(uniqueKey string) bool {
	defer h.lock.Unlock()
	h.lock.Lock()
	if _, ok := h.ShellClientMap[uniqueKey]; ok {
		return true
	}
	return false
}

// RmClient 移除连接
func (h *TShell) RmClient(uniqueKey string) {
	defer h.lock.Unlock()
	h.lock.Lock()
	delete(h.ShellClientMap, uniqueKey)
}

func (h *TShell) WalkShellList(businessFunc func(uniqueKey string, gsShell *gsssh.SshConfig)) {
	defer h.lock.Unlock()
	h.lock.Lock()
	for uniqueKey, gsShell := range h.ShellClientMap {
		businessFunc(uniqueKey, gsShell)
	}
}
