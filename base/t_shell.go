package base

import (
	"errors"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"sync"
	"time"
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
	if shell, ok := h.ShellClientMap[uniqueKey]; ok && shell != nil {
		return shell, nil
	}
	gsShell := &gsssh.SshConfig{
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
		GsSlog:   Component.GsLog,
	}
	gsShell.SetMaxRunSecond(20)
	createErr := gsShell.ConnectAuthPassword()
	if createErr != nil {
		return nil, createErr
	}
	//先执行一次确保连接正常
	_, err := gsShell.RunCommandWait(`pwd`)
	if err != nil {
		return nil, err
	}
	//设置执行命令前处理
	gsShell.SetFuncBefore(func(command string) string {
		return `■■ ` + command
	})
	//设置对收到的结果是否进行合并后处理 建议1-2
	gsShell.SetCombineNum(1)
	//是否显示执行命令后linux返回的执行的命令 如果设置了SetFuncBefore处理，那么就关闭
	gsShell.CloseFirstReceiveMsg()
	//设置关闭事件 用来进行重连
	//TODO 有时间研究一下 为什么sftp的链接断开后没有重连
	gsShell.SetFuncBroken(func() {
		gstool.FmtPrintlnLogTime(uniqueKey + `连接中断，下次动作时进行链接`)
		Component.TSocket.SendMsg(uniqueKey, uniqueKey+` 注意：连接已中断，下次动作时进行链接`)
		h.RmClient(uniqueKey)
		//h.ReConn(uniqueKey, sshConfig)
	})
	gsShell.SetFuncReceiveMsg(func(msg string) string {
		Component.TSse.Send(uniqueKey, msg)
		return msg
	})
	h.ShellClientMap[uniqueKey] = gsShell
	return gsShell, nil
}

// ReConn 重连
func (h *TShell) ReConn(uniqueKey string, sshConfig map[string]any) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			Component.TSocket.SendMsg(uniqueKey, uniqueKey+` 准备进行重连`)
			shell, err := h.GetClient(sshConfig, uniqueKey)
			if err == nil { //连接成功 那么中断重连
				Component.TSocket.SendMsg(uniqueKey, uniqueKey+` 重连成功`)
				shell.SetSocket(Component.TSocket.GetSocket(uniqueKey))
				break
			} else {
				Component.TSocket.SendMsg(uniqueKey, uniqueKey+` 重连失败 `+err.Error())
			}
		}
	}()
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
