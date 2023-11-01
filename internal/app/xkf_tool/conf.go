package xkf_tool

import (
	"flag"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsnsq"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"sync"
)

var ConfigViper *viper.Viper
var EncryptMain *gstool.Encrypt //加密
var RedisRunList map[string]*gsdb.GsRedis
var XkfDevMysql *gsdb.GsMysql
var AppurlDevMysql *gsdb.GsMysql
var Logger *gstool.GsLogger
var ProducerMap map[string]*gsnsq.NsqStruct
var RootPath string
var Env string
var DbInitLock sync.RWMutex

func InitConfig() {
	defer func() {
		if err := recover(); err != nil {
			gstool.FmtPrintlnLog(`初始化失败 %#v`, err)
		}
	}()
	flag.StringVar(&Env, "env", "prod", "pro则为线上环境，dev则未开发环境，默认pro线上环境")
	flag.Parse()
	var err error
	RootPath, err = gstool.GetRootPath()
	if err != nil {
		panic(err.Error())
	}
	Logger = gstool.CreateLogger(RootPath+`/logs`, `xkf_tool`)
	gstool.FmtPrintlnLog(`日志路径 %s`, RootPath+`/logs/xkf_tool`)
	ConfigViper = viper.New()
	ConfigViper.AddConfigPath(RootPath + `/config`)
	ConfigViper.SetConfigName(`config`)
	ConfigViper.SetConfigType(`ini`)
	RedisRunList = make(map[string]*gsdb.GsRedis)
	ProducerMap = make(map[string]*gsnsq.NsqStruct)
	RunShell3TerminalMap = make(map[string]*gstool.GsShell)
	RunShellCli4Map = gstool.HighMapCreate(100)
	XkfDevMysql = nil
	AppurlDevMysql = nil
	if err := ConfigViper.ReadInConfig(); err != nil {
		panic(`读取配置失败 config/config.ini`)
	}
	EncryptMain = &gstool.Encrypt{
		Key: ConfigViper.GetString(`encrypt.key`),
		Iv:  ConfigViper.GetString(`encrypt.iv`),
	}
	go InitXkfSocket()
	go InitWkSocket()
}

//GetProducer 拿到生产者
//func GetProducer(host, port, topic string) *gsnsq.NsqStruct {
//	checkKey := host + port + topic
//	if producer, ok := ProducerMap[checkKey]; ok {
//		return producer
//	}
//	producer := gsnsq.NsqInit(topic)
//	err := producer.CreateProducer(gsnsq.NsqConfig{
//		Host: host,
//		Port: port,
//	})
//	if err != nil {
//		Logger.Errorf(`GetProducer ` + err.Error())
//		return nil
//	}
//	ProducerMap[checkKey] = producer
//	return producer
//}

//GetDevMysql x
func GetDevMysql(reqBody *SshExec) {
	DbInitLock.Lock()
	defer DbInitLock.Unlock()
	if cast.ToInt(reqBody.XkfDevDbConfig.Port) != 0 && XkfDevMysql == nil {
		gsMysqlConfig := &gsdb.MysqlConfig{
			Host:              reqBody.XkfDevDbConfig.Host,
			Port:              reqBody.XkfDevDbConfig.Port,
			Username:          reqBody.XkfDevDbConfig.Username,
			Password:          reqBody.XkfDevDbConfig.Password,
			Dbname:            reqBody.XkfDevDbConfig.Dbname,
			PoolSize:          1,
			MaxOpenConns:      1,
			MaxIdleConns:      1,
			MaxLifetimeSecond: 60,
		}
		gsMysql := gsdb.GsMysql{MysqlConfig: gsMysqlConfig}
		var err error
		err = gsMysql.CreateConn()
		if err != nil {
			Logger.Errorf(`初始化mysql错误 %#v`, err)
			return
		}
		XkfDevMysql = &gsMysql

		//初始化第二个
		gsMysqlConfig.Dbname = `appurl_test`
		gsAppMysql := gsdb.GsMysql{MysqlConfig: gsMysqlConfig}
		err = gsAppMysql.CreateConn()
		if err != nil {
			Logger.Errorf(`初始化mysql错误 %#v`, err)
			return
		}
		AppurlDevMysql = &gsAppMysql
	}
}

//GetSshUnikey 拿到ssh 唯一值
func GetSshUnikey(sshConfig *SshConfig) string {
	return gstool.Md5(fmt.Sprintf(`%s%s%s%s`, sshConfig.Host, sshConfig.Port, sshConfig.Username, sshConfig.Port))
}
