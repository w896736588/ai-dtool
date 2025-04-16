package base

import (
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"runtime/debug"
	"sync"
)

type TShell struct {
	ShellClientMap map[string]*gsssh.SshConfig
	lock           sync.Mutex
}

// GetClient 正常输出
func (h *TShell) GetClient(sshConfig map[string]any, shellClientId, sseClientId string, formatStream func(string) []string) (*gsssh.SshConfig, error) {
	defer h.lock.Unlock()
	h.lock.Lock()
	sshId := cast.ToString(sshConfig[`id`])
	if sshId == `` {
		return nil, errors.New(`ssh配置错误，GetClient ` + cast.ToString(debug.Stack()))
	}
	if shell, ok := h.ShellClientMap[shellClientId]; ok && shell != nil {
		return shell, nil
	}
	gsShell := &gsssh.SshConfig{
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
		GsSlog:   Component.GsLog,
	}

	//TODO 有时间研究一下 为什么sftp的链接断开后没有重连
	//设置关闭事件
	gsShell.SetFuncBroken(func() {
		_ = Component.TSse.SendMsg(sseClientId, sseClientId+` 注意：连接已中断，下次动作时进行链接`+"\n", 0)
		h.RmClient(shellClientId)
		//已经加了自动重连
		//h.ReConn(shellClientId , sshConfig)
	})
	gsShell.SetMaxRunSecond(40)
	createErr := gsShell.ConnectAuthPassword()
	if createErr != nil {
		return nil, createErr
	}
	//先执行一次确保连接正常
	_, err := gsShell.RunCommandWait(`pwd`)
	if err != nil {
		return nil, err
	}
	//回调准备输出的内容 放到这里 就不需要链接linux出现的一大段文字
	gsShell.SetFuncStreamReceive(func(msg string) {
		Component.GsLog.Errof(`%s`, msg)
		if formatStream != nil {
			Component.GsLog.Errof(`解析前的 %s`, msg)
			msgList := formatStream(msg)
			Component.GsLog.Errof(`解析后的 %s`, gstool.JsonEncode(msgList))
			_ = Component.TSse.SendMsgChunkList(sseClientId, msgList, 10)
		} else {
			_ = Component.TSse.SendMsgChunk(sseClientId, msg, Chunk{
				Type: ChunkNum,
				Num:  50,
			}, 10)
		}
	})
	//设置执行命令前处理
	gsShell.SetFuncBefore(func(command string) string {
		return command
	})
	//设置对收到的结果是否进行合并后处理 建议1-2
	gsShell.SetCombineNum(1)
	//是否显示执行命令后linux返回的执行的命令 如果设置了SetFuncBefore处理，那么就关闭
	gsShell.CloseFirstReceiveMsg()

	h.ShellClientMap[shellClientId] = gsShell
	return gsShell, nil
}

// GetClientMarkdown 输出markdown格式
func (h *TShell) GetClientMarkdown(sshConfig map[string]any, shellClientId, sseClientId string) (*gsssh.SshConfig, error) {
	defer h.lock.Unlock()
	h.lock.Lock()
	sshId := cast.ToString(sshConfig[`id`])
	if sshId == `` {
		return nil, errors.New(`ssh配置错误，GetClientMarkdown ` + cast.ToString(debug.Stack()))
	}
	if shell, ok := h.ShellClientMap[shellClientId]; ok && shell != nil {
		return shell, nil
	}
	gsShell := &gsssh.SshConfig{
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
		GsSlog:   Component.GsLog,
	}

	//TODO 有时间研究一下 为什么sftp的链接断开后没有重连
	//设置关闭事件
	gsShell.SetFuncBroken(func() {
		_ = Component.TSse.SendMsg(sseClientId, sseClientId+` 注意：连接已中断，下次动作时进行链接`+"\n", 0)
		h.RmClient(shellClientId)
		//已经加了自动重连
		//h.ReConn(shellClientId , sshConfig)
	})
	gsShell.SetMaxRunSecond(40)
	createErr := gsShell.ConnectAuthPassword()
	if createErr != nil {
		return nil, createErr
	}
	//先执行一次确保连接正常
	_, err := gsShell.RunCommandWait(`pwd`)
	if err != nil {
		return nil, err
	}
	//猪油：下面3个注册回调，放到这里的话就不会输出pwd以及连接相关信息
	//回调准备输出的内容
	gsShell.SetFuncStreamReceive(func(msg string) {
		_ = Component.TSse.SendMsgChunk(sseClientId, msg+"  \n", Chunk{
			Type: ChunkEnter,
		}, 50)
	})
	gsShell.SetFuncStartCommand(func() {
		_ = Component.TSse.SendMsg(sseClientId, fmt.Sprintf("```%s\n#%s", `bash`, `bash`)+"\n", 0)
	})
	gsShell.SetFuncEndCommand(func() {
		_ = Component.TSse.SendMsg(sseClientId, "```\n", 0)
	})
	//设置执行命令前处理
	gsShell.SetFuncBefore(func(command string) string {
		return command
		//return Component.TMarkDown.Code(command, `shell`)
	})
	//设置对收到的结果是否进行合并后处理 建议1-2
	gsShell.SetCombineNum(1)
	//是否显示执行命令后linux返回的执行的命令 如果设置了SetFuncBefore处理，那么就关闭
	gsShell.CloseFirstReceiveMsg()

	h.ShellClientMap[shellClientId] = gsShell
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
