package base_module

import (
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

func (h *Global) MysqlGetClient(name string) (*gsdb.GsMysql, error) {
	clientValue := h.mysqlClientMap.G(name)
	if clientValue != nil {
		return clientValue.Value().(*gsdb.GsMysql), nil
	}
	config, err := h.MysqlGetConfig(name)
	if err != nil {
		return nil, err
	}
	gsDb := gsdb.NewMysql(config, true)
	gsDb.OpenDebug()
	err = gsDb.CreateConn()
	if err != nil {
		return nil, err
	}
	h.mysqlClientMap.S(name, gsDb)
	return gsDb, nil
}

func (h *Global) MysqlGetConfig(name string) (*gsdb.MysqlConfig, error) {
	returnConfig := &gsdb.MysqlConfig{}
	valueConfig := h.mysqlConfigMap.G(name)
	if valueConfig == nil {
		return nil, errors.New(`未注册的mysql服务`)
	}
	err := gstool.JsonDecode(valueConfig.ToStr(), returnConfig)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(`解析失败 %#v %s`, valueConfig, err.Error()))
	}
	return returnConfig, nil
}

func (h *Global) MysqlSetConfig(config *gsdb.MysqlConfig) {
	if config.Name == `` {
		h.Warn(`未设置name，可能存在问题，每个配置需要设置不同的name %v`, config)
	}
	h.mysqlConfigMap.S(cast.ToString(config.Name), gstool.JsonEncode(config))
}
