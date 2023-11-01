package base_module

import (
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
)

var globalMap *gstool.GsConsMap
var Logger *gstool.GsLogger

func init() {
	globalMap = gstool.GsConsMapNew(10)
}

func CreateGlobal(unikey string) {
	globalMap.D(unikey)
	global := &Global{}
	global.Init()
	globalMap.S(unikey, global)
}

//GetGlobal 获取全局配置
func GetGlobal(unikey string) *Global {
	cacheValue := globalMap.G(unikey)
	if cacheValue == nil {
		return nil
	}
	return cacheValue.Value().(*Global)
}

func (h *Global) GetEncrypt() *gstool.Encrypt {
	return h.encrypt
}

type Global struct {
	redisClientMap *gstool.GsConsMap //全局的redis客户端连接
	redisConfigMap *gstool.GsConsMap //全局的redis配置
	mysqlConfigMap *gstool.GsConsMap //全局的mysql配置
	mysqlClientMap *gstool.GsConsMap //全局的mysql客户端连接
	shellClientMap *gstool.GsConsMap //全局的shell客户端连接
	shellConfigMap *gstool.GsConsMap //全局的shell配置
	encrypt        *gstool.Encrypt   //全局的加密配置
	logger         *gstool.GsLogger  //如果设置了 那么将输出日志
	gin            *gsgin.GSGin      //API接口
}

func (h *Global) SetLogger(logger *gstool.GsLogger) {
	h.logger = logger
}

func (h *Global) SetEncrypt(encryptKey, encryptIv string) {
	h.encrypt = &gstool.Encrypt{
		Key: encryptKey,
		Iv:  encryptIv,
	}
}

func (h *Global) Init() {
	h.redisConfigMap = gstool.GsConsMapNew(10)
	h.redisClientMap = gstool.GsConsMapNew(10)
	h.mysqlConfigMap = gstool.GsConsMapNew(10) //全局的mysql配置
	h.mysqlClientMap = gstool.GsConsMapNew(10) //全局的mysql客户端连接
	h.shellClientMap = gstool.GsConsMapNew(10) //全局的shell客户端连接
	h.shellConfigMap = gstool.GsConsMapNew(10) //全局的shell配置
	h.SetLogger(Logger)
}

func (h *Global) Debug(msg string, args ...interface{}) {
	if h.logger != nil {
		h.logger.Debugf(msg, args...)
	}
}

func (h *Global) Warn(msg string, args ...interface{}) {
	if h.logger != nil {
		h.logger.Warningf(msg, args...)
	}
}

func (h *Global) Error(msg string, args ...interface{}) {
	if h.logger != nil {
		h.logger.Errorf(msg, args...)
	}
}
