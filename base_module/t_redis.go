package base_module

import (
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gstool"
)

func (h *Global) RedisGetClient(name string) (*gsdb.GsRedis, error) {
	gstool.FmtPrintlnLog(`获取redis name为%s`, name)
	clientValue := h.redisClientMap.G(name)
	if clientValue != nil {
		return clientValue.Value().(*gsdb.GsRedis), nil
	}
	config, err := h.RedisGetConfig(name)
	if err != nil {
		return nil, err
	}
	gstool.FmtPrintlnLog(`获取redis配置为%#v`, config)
	gsDb := &gsdb.GsRedis{
		RedisConfig: config,
	}
	err = gsDb.CreateConn()
	if err != nil {
		return nil, err
	}
	h.redisClientMap.S(name, gsDb)
	return gsDb, nil
}

func (h *Global) RedisGetConfig(name string) (*gsdb.RedisConfig, error) {
	returnConfig := &gsdb.RedisConfig{}
	valueConfig := h.redisConfigMap.G(name)
	gstool.FmtPrintlnLog(`获取redis配置 %s %#v`, name, valueConfig)
	if valueConfig == nil {
		return nil, errors.New(`未注册的服务`)
	}
	err := gstool.JsonDecode(valueConfig.ToStr(), returnConfig)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(`解析失败 %#v %s`, valueConfig, err.Error()))
	}
	return returnConfig, nil
}

//RedisEachConfigList 循环配置
func (h *Global) RedisEachConfigList(call func(string, *gstool.GsCons)) {
	h.redisConfigMap.Each(call)
}

func (h *Global) RedisSetConfig(config *gsdb.RedisConfig) {
	if config.Name == `` {
		h.Warn(`未设置name，可能存在问题，每个配置需要设置不同的name %v`, config)
	}
	h.redisConfigMap.S(config.Name, gstool.JsonEncode(config))
}
