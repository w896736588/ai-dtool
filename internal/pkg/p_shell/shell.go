package p_shell

import (
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"io"

	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
	"golang.org/x/crypto/ssh"
)

type Shell struct {
	ShellClientMap map[string]*gsssh.SshTerminal
	lock           sync.Mutex
	LogPath        string
	log            *gstool.GsSlog
}

func NewShell(logPath string) *Shell {
	log := gstool.NewSlog3(logPath, `shell`)
	_ = log.CleanOldLogs(2)
	return &Shell{
		ShellClientMap: make(map[string]*gsssh.SshTerminal),
		log:            log,
		LogPath:        logPath,
	}
}

// GetClient 正常输出
func (h *Shell) GetClient(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell,
	formatStream func(string) []string, promptKeywords []string, promptFunc func(string, io.WriteCloser, *ssh.Session)) (*gsssh.SshTerminal, error) {
	defer h.lock.Unlock()
	h.lock.Lock()
	sshId := cast.ToString(sshConfig[`id`])
	if sshId == `` {
		return nil, errors.New(`ssh配置错误，GetClient ` + cast.ToString(debug.Stack()))
	}
	if shell, ok := h.ShellClientMap[shellClientId]; ok && shell != nil {
		return shell, nil
	}
	gsShell := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
		Name:     "",
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
	}))
	//设置关闭事件
	gsShell.SetFuncBroken(func(msg string) {
		sse.Send(` 注意：连接已中断，下次动作时进行链接,` + msg + "\n")
		h.RmClient(shellClientId)
	})
	gsShell.SetPtyConfig(gsssh.PtyConfig{
		Echo: 1,
	})
	gsShell.SetMaxBufferSize(2 * 1024 * 1024) //最大允许2M的输出
	//先执行一次确保连接正常
	maxRunSecond := time.Second * 40
	_, err := gsShell.RunCommandWait(`pwd`, maxRunSecond)
	if err != nil {
		return nil, err
	}
	//回调准备输出的内容 放到这里 就不需要链接linux出现的一大段文字
	gsShell.SetFuncReceiveMsg(func(msg string) string {
		if formatStream != nil {
			msgList := formatStream(msg)
			for _, msg := range msgList {
				sse.Send(msg)
			}
		} else {
			sse.Send(msg)
		}
		return msg
	})
	//提示验证提示
	if len(promptKeywords) == 0 {
		promptKeywords = []string{
			"Username for",
			"Password for",
			"用户名：",
			"密码：",
			"passphrase",
			"Passphrase",
		}
	}
	gsShell.SetAuthPromptKeywords(promptKeywords)
	if promptFunc != nil {
		gsShell.SetFuncAuthPrompt(promptFunc)
	}
	h.ShellClientMap[shellClientId] = gsShell
	return gsShell, nil
}

// GetClientMarkdown 输出markdown格式
func (h *Shell) GetClientMarkdown(sshConfig map[string]any, shellClientId string, sse *p_sse.SseShell) (*gsssh.SshTerminal, error) {
	defer h.lock.Unlock()
	h.lock.Lock()
	sshId := cast.ToString(sshConfig[`id`])
	if sshId == `` {
		return nil, errors.New(`ssh配置错误，GetClientMarkdown ` + cast.ToString(debug.Stack()))
	}
	if shell, ok := h.ShellClientMap[shellClientId]; ok && shell != nil {
		h.SetSse(shell, sse)
		return shell, nil
	}
	gsShell := gsssh.NewSshTerminal(gsssh.NewSsh(&gsssh.SshConfig{
		Name:     "",
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
	}))
	gsShell.SetPtyConfig(gsssh.PtyConfig{
		Echo: 1,
	})
	//TODO 有时间研究一下 为什么sftp的链接断开后没有重连
	//设置关闭事件
	gsShell.SetFuncBroken(func(msg string) {
		sse.Send(` 注意：连接已中断，下次动作时进行链接,` + msg + "\n")
		h.RmClient(shellClientId)
	})
	gsShell.SetMaxBufferSize(2 * 1024 * 1024) //最大允许2M的输出
	//先执行一次确保连接正常
	_, err := gsShell.RunCommandWait(`pwd`, time.Second*40)
	if err != nil {
		return nil, err
	}
	//猪油：下面3个注册回调，放到这里的话就不会输出pwd以及连接相关信息
	h.SetSse(gsShell, sse)
	//提示输入账号密码登处理
	gsShell.SetAuthPromptKeywords([]string{
		"Username for",
		"Password for",
		"用户名：",
		"密码：",
		"passphrase",
		"Passphrase",
	})
	gsShell.SetFuncAuthPrompt(func(prompt string, stdin io.WriteCloser, session *ssh.Session) {
		// 发送 Ctrl+C 信号（模拟终端中断）
		if session != nil {
			_ = session.Signal(ssh.SIGINT)
			//清除认证缓存
			if strings.Contains(strings.ToLower(prompt), `git`) {
				_, _ = stdin.Write([]byte("git credential-cache exit; unset GIT_ASKPASS\n"))
			}
			sse.Send("\n需要输入账号或密码，暂时不支持，请解决后再次执行\n")
			return
		}
		return
	})

	h.ShellClientMap[shellClientId] = gsShell
	return gsShell, nil
}

func (h *Shell) SetSse(gsShell *gsssh.SshTerminal, sse *p_sse.SseShell) {
	//回调准备输出的内容
	gsShell.SetFuncReceiveMsg(func(msg string) string {
		sse.Send(msg)
		return msg
	})
}

func (h *Shell) GetSshOnce(sshConfig map[string]any) (*gsssh.SshOnce, error) {
	sshId := cast.ToString(sshConfig[`id`])
	if sshId == `` {
		return nil, errors.New(`ssh配置错误，GetClientMarkdown ` + cast.ToString(debug.Stack()))
	}

	return gsssh.NewSshOnce(gsssh.NewSsh(&gsssh.SshConfig{
		Name:     "",
		Host:     cast.ToString(sshConfig["host"]),
		Port:     cast.ToString(sshConfig["port"]),
		UserName: cast.ToString(sshConfig["username"]),
		Password: cast.ToString(sshConfig["password"]),
	})), nil

}

func (h *Shell) Exist(uniqueKey string) bool {
	defer h.lock.Unlock()
	h.lock.Lock()
	if _, ok := h.ShellClientMap[uniqueKey]; ok {
		return true
	}
	return false
}

// RmClient 移除连接
func (h *Shell) RmClient(uniqueKey string) {
	defer h.lock.Unlock()
	h.lock.Lock()
	if ssh, ok := h.ShellClientMap[uniqueKey]; ok {
		ssh.CloseTerminal()
	}
	delete(h.ShellClientMap, uniqueKey)
}

func (h *Shell) WalkShellList(businessFunc func(uniqueKey string, gsShell *gsssh.SshTerminal)) {
	defer h.lock.Unlock()
	h.lock.Lock()
	for uniqueKey, gsShell := range h.ShellClientMap {
		businessFunc(uniqueKey, gsShell)
	}
}
