package base_module

import (
	"gitee.com/Sxiaobai/gs/gsencrypt"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
)

var globalMap *gstool.GsConsMap
var Logger *gstool.GsSlog
var RootPath string //根目录

func init() {
	globalMap = gstool.GsConsMapNew(10)
}

func CreateGlobal(unikey string) {
	globalMap.D(unikey)
	global := &Global{}
	global.Init()
	globalMap.S(unikey, global)
}

// GetGlobal 获取全局配置
func GetGlobal(unikey string) *Global {
	cacheValue := globalMap.G(unikey)
	if cacheValue == nil {
		return nil
	}
	return cacheValue.Value().(*Global)
}

func (h *Global) GetEncrypt() *gsencrypt.DesCbc {
	return h.encrypt
}

type Global struct {
	redisClientMap *gstool.GsConsMap           //全局的redis客户端连接
	redisConfigMap *gstool.GsConsMap           //全局的redis配置
	mysqlConfigMap *gstool.GsConsMap           //全局的mysql配置
	mysqlClientMap *gstool.GsConsMap           //全局的mysql客户端连接
	shellMap       map[string]*gsssh.SshConfig //全局的shell配置
	encrypt        *gsencrypt.DesCbc           //全局的加密配置
	logger         *gstool.GsSlog              //如果设置了 那么将输出日志
	gin            *gsgin.GSGin                //API接口
}

func (h *Global) SetLogger(logger *gstool.GsSlog) {
	h.logger = logger
}

func (h *Global) GetLogger() *gstool.GsSlog {
	return h.logger
}

func (h *Global) SetEncrypt(encryptKey, encryptIv string) {
	h.encrypt = &gsencrypt.DesCbc{
		Key: encryptKey,
		Iv:  encryptIv,
	}
}

func (h *Global) Init() {
	h.redisConfigMap = gstool.GsConsMapNew(10)
	h.redisClientMap = gstool.GsConsMapNew(10)
	h.mysqlConfigMap = gstool.GsConsMapNew(10)     //全局的mysql配置
	h.mysqlClientMap = gstool.GsConsMapNew(10)     //全局的mysql客户端连接
	h.shellMap = make(map[string]*gsssh.SshConfig) //全局的shell配置
	h.SetLogger(Logger)
}

func (h *Global) Debug(msg string, args ...interface{}) {
	if h.logger != nil {
		h.logger.Debugf(msg, args...)
	}
}

func (h *Global) Warn(msg string, args ...interface{}) {
	if h.logger != nil {
		h.logger.Warnf(msg, args...)
	}
}

func (h *Global) Error(msg string, args ...interface{}) {
	if h.logger != nil {
		h.logger.Errof(msg, args...)
	}
}
