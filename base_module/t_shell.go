package base_module

import (
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

func (h *Global) ShellGetClient(name string) (*gstool.GsShell, error) {
	clientValue := h.shellClientMap.G(name)
	if clientValue != nil {
		return clientValue.Value().(*gstool.GsShell), nil
	}
	config, err := h.ShellGetConfig(name)
	gstool.FmtPrintlnLog(`获取shell %s 配置 %#v`, name, config)
	if err != nil {
		return nil, err
	}
	gsShell := &gstool.GsShell{
		Config: config,
	}
	err = gsShell.CreateClient()
	if err != nil {
		return nil, err
	}
	h.shellClientMap.S(name, gsShell)
	return gsShell, nil
}

func (h *Global) ShellGetConfig(name string) (*gstool.ShellConfig, error) {
	returnConfig := &gstool.ShellConfig{}
	valueConfig := h.shellConfigMap.G(name)
	if valueConfig == nil {
		return nil, errors.New(`未注册的服务`)
	}
	err := gstool.JsonDecode(valueConfig.ToStr(), returnConfig)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(`解析失败 %#v %s`, valueConfig, err.Error()))
	}
	return returnConfig, nil
}

func (h *Global) ShellSetConfig(config *gstool.ShellConfig) {
	if config.Name == `` {
		h.Warn(`未设置name，可能存在问题，每个配置需要设置不同的name %v`, config)
	}
	h.shellConfigMap.S(cast.ToString(config.Name), gstool.JsonEncode(config))
}

//流
func (h *Global) ShellPushGetClient(name string) (*gstool.GsShellPush, error) {
	clientValue := h.shellClientMap.G(name)
	if clientValue != nil {
		return clientValue.Value().(*gstool.GsShellPush), nil
	}
	config, err := h.ShellPushGetConfig(name)
	if err != nil {
		return nil, err
	}
	gsShell := &gstool.GsShellPush{
		Config: config,
		Logger: h.logger,
	}
	err = gsShell.CreateClient()
	gstool.FmtPrintlnLog(`创建shell完成 %s`, name)
	if err != nil {
		return nil, err
	}
	h.shellClientMap.S(name, gsShell)
	return gsShell, nil
}

func (h *Global) ShellPushGetConfig(name string) (*gstool.ShellPushConfig, error) {
	returnConfig := &gstool.ShellPushConfig{}
	valueConfig := h.shellConfigMap.G(name)
	if valueConfig == nil {
		return nil, errors.New(`未注册的服务`)
	}
	err := gstool.JsonDecode(valueConfig.ToStr(), returnConfig)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(`解析失败 %#v %s`, valueConfig, err.Error()))
	}
	return returnConfig, nil
}

func (h *Global) ShellPushSetConfig(config *gstool.ShellPushConfig) {
	if config.Name == `` {
		h.Warn(`未设置name，可能存在问题，每个配置需要设置不同的name %v`, config)
	}
	h.shellConfigMap.S(cast.ToString(config.Name), gstool.JsonEncode(config))
}
