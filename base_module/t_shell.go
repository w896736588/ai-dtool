package base_module

import (
	"errors"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
)

func (h *Global) ShellGet(name string) (*gsssh.SshConfig, error) {
	exist := gstool.MapKeyExist(&h.shellMap, name)
	if !exist {
		return nil, errors.New(`配置` + name + `不存在`)
	}
	gsShell := h.shellMap[name]
	return gsShell, nil
}

func (h *Global) ShellSet(shell *gsssh.SshConfig) {
	h.shellMap[shell.Name] = shell
}
