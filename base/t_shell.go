package base

import (
	"dev_tool/base/define"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"

	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

type TShell struct {
	ShellClientMap map[string]*gsssh.SshConfig
	lock           sync.Mutex
	log            *gstool.GsSlog
}

func NewTShell() *TShell {
	log := gstool.NewSlog3(Component.Env.LogPath, `shell`)
	_ = log.CleanOldLogs(2)
	return &TShell{
		ShellClientMap: make(map[string]*gsssh.SshConfig),
		log:            log,
	}
}

// GetClient 正常输出
func (h *TShell) GetClient(sshConfig map[string]any, shellClientId, sseClientId string,
	formatStream func(string) []string) (*gsssh.SshConfig, error) {
	defer h.lock.Unlock()
	h.lock.Lock()
	sshId := cast.ToString(sshConfig[`id`])
	if sshId == `` {
		return nil, errors.New(`ssh配置错误，GetClient ` + cast.ToString(debug.Stack()))
	}
	if shell, ok := h.ShellClientMap[shellClientId]; ok && shell != nil {
		return shell, nil
	}
	gsShell := gsssh.NewSshAuthPassword(cast.ToString(sshConfig["host"]),
		cast.ToString(sshConfig["port"]), cast.ToString(sshConfig["username"]),
		cast.ToString(sshConfig["password"]))
	gsShell.GsSlog = Component.GsLog

	//设置关闭事件
	gsShell.SetFuncBroken(func() {
		_ = Component.TSse.SendMsg(sseClientId, define.SseContentTypeMsg, sseClientId+` 注意：连接已中断，下次动作时进行链接`+"\n", 0)
		h.RmClient(shellClientId)
	})
	gsShell.SetMaxRunSecond(40)
	gsShell.SetMaxBufferSize(2 * 1024 * 1024) //最大允许2M的输出
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
		//msg = gstool.StringFilterANSI(msg)
		h.log.Debugf(`receive：%s`, msg)
		if formatStream != nil {
			msgList := formatStream(msg)
			_ = Component.TSse.SendMsgChunkList(sseClientId, msgList, 10)
		} else {
			_ = Component.TSse.SendMsg(sseClientId, define.SseContentTypeMsg, msg, 10)
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

	gsShell := gsssh.NewSshAuthPassword(cast.ToString(sshConfig["host"]),
		cast.ToString(sshConfig["port"]), cast.ToString(sshConfig["username"]),
		cast.ToString(sshConfig["password"]))
	gsShell.GsSlog = Component.GsLog

	//TODO 有时间研究一下 为什么sftp的链接断开后没有重连
	//设置关闭事件
	gsShell.SetFuncBroken(func() {
		_ = Component.TSse.SendMsg(sseClientId, define.SseContentTypeMsg, sseClientId+` 注意：连接已中断，下次动作时进行链接`+"\n", 0)
		h.RmClient(shellClientId)
	})
	gsShell.SetMaxRunSecond(40)
	gsShell.SetMaxBufferSize(2 * 1024 * 1024) //最大允许2M的输出
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
		//msg = gstool.StringFilterANSI(msg)
		_ = Component.TSse.SendMsg(sseClientId, define.SseContentTypeMsg, msg, 50)
	})
	gsShell.SetFuncStartCommand(func() {
		_ = Component.TSse.SendMsg(sseClientId, define.SseContentTypeMsg, fmt.Sprintf("```%s\n#%s", `bash`, `bash`)+"\n", 0)
	})
	gsShell.SetFuncEndCommand(func() {
		_ = Component.TSse.SendMsg(sseClientId, define.SseContentTypeMsg, "```\n", 0)
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
	if ssh, ok := h.ShellClientMap[uniqueKey]; ok {
		ssh.CloseTerminal()
	}
	delete(h.ShellClientMap, uniqueKey)
}

func (h *TShell) WalkShellList(businessFunc func(uniqueKey string, gsShell *gsssh.SshConfig)) {
	defer h.lock.Unlock()
	h.lock.Lock()
	for uniqueKey, gsShell := range h.ShellClientMap {
		businessFunc(uniqueKey, gsShell)
	}
}
