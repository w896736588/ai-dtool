package xkf_tool

import (
	"flag"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsnsq"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
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
var RunShellMap map[string]*gstool.GsShell
var RunShellMapLock sync.RWMutex
var RunShellTerminalMap map[string]*gstool.GsShell

func InitConfig() {
	defer func() {
		if err := recover(); err != nil {
			gstool.FmtPrintlnLog(`初始化失败 %#v`, err)
		}
	}()
	flag.StringVar(&Env, "env", "prod", "pro则为线上环境，dev则未开发环境，默认pro线上环境")
	flag.Parse()
	if Env == `dev` {
		_, RootPath, _, _ = runtime.Caller(0)
		RootPath = gstool.DirUpNum(RootPath, 4)
	} else {
		var err error
		sysType := runtime.GOOS
		RootPath, err = os.Getwd()
		if sysType == `windows` {
			RootPath = strings.ReplaceAll(RootPath, `\`, `/`)
		}

		gstool.FmtPrintlnLog(`当前的目录为 %s`, RootPath)
		if err != nil {
			gstool.FmtPrintlnLog(`getWd失败 %s`, err.Error())
		}
	}
	Logger = gstool.CreateLogger(RootPath+`/logs`, `xkf_tool`)
	gstool.FmtPrintlnLog(`日志路径 %s`, RootPath+`/logs/xkf_tool`)
	ConfigViper = viper.New()
	ConfigViper.AddConfigPath(RootPath + `/config`)
	ConfigViper.SetConfigName(`config`)
	ConfigViper.SetConfigType(`ini`)
	RedisRunList = make(map[string]*gsdb.GsRedis)
	ProducerMap = make(map[string]*gsnsq.NsqStruct)
	RunShellMap = make(map[string]*gstool.GsShell)
	RunShellTerminalMap = make(map[string]*gstool.GsShell)
	if err := ConfigViper.ReadInConfig(); err != nil {
		panic(`读取配置失败 config/config.ini`)
	}
	EncryptMain = &gstool.Encrypt{
		Key: ConfigViper.GetString(`encrypt.key`),
		Iv:  ConfigViper.GetString(`encrypt.iv`),
	}
}

//GetProducer 拿到生产者
func GetProducer(host, port, topic string) *gsnsq.NsqStruct {
	checkKey := host + port + topic
	if producer, ok := ProducerMap[checkKey]; ok {
		return producer
	}
	producer := gsnsq.NsqInit(topic)
	err := producer.CreateProducer(gsnsq.NsqConfig{
		Host: host,
		Port: port,
	})
	if err != nil {
		Logger.Errorf(`GetProducer ` + err.Error())
		return nil
	}
	ProducerMap[checkKey] = producer
	return producer
}

//GetRunShellCliTer
func GetRunShellCliTer(sshConfig *SshConfig) *gstool.GsShell {
	RunShellMapLock.Lock()
	defer RunShellMapLock.Unlock()
	if sshConfig.Host == `` {
		return nil
	}

	uniKey := fmt.Sprintf(`%s%s%s%s`, sshConfig.Host, sshConfig.Port, sshConfig.Username, sshConfig.Port)
	if RunShellMap[uniKey] == nil {
		gsShellTerConfig := gstool.GsShellConfig{
			Host:          sshConfig.Host,
			Port:          cast.ToInt64(sshConfig.Port),
			Username:      sshConfig.Username,
			Password:      sshConfig.Password,
			TimeoutSecond: 100,
		}
		cliTerConf := gstool.GsShell{
			Config:              &gsShellTerConfig,
			IsOpenLog:           true,
			Logger:              Logger,
			TerminalRefreshTime: 100 * time.Millisecond,
			TerminalMaxTime:     10 * time.Second,
		}
		cliTerConfErr := cliTerConf.CreateClient()
		if cliTerConfErr != nil {
			panic(`创建交互式链接失败 ` + cliTerConfErr.Error())
		} else {
			RunShellTerminalMap[uniKey] = &cliTerConf
		}
	}
	return RunShellTerminalMap[uniKey]
}

//GetDevMysql x
func GetDevMysql(reqBody *SshExec) {

	if reqBody.XkfDevDbConfig.Host != `` && reqBody.XkfDevDbConfig.Host != nil && XkfDevMysql == nil {
		gsMysqlConfig := gsdb.MysqlConfig{
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
		}
		XkfDevMysql = &gsMysql

		//初始化第二个
		gsMysqlConfig.Dbname = `appurl_test`
		gsAppMysql := gsdb.GsMysql{MysqlConfig: gsMysqlConfig}
		err = gsAppMysql.CreateConn()
		if err != nil {
			Logger.Errorf(`初始化mysql错误 %#v`, err)
		}
		AppurlDevMysql = &gsAppMysql
	}
}
