package base

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"redis_manager/define"
)

var RedisList []define.RedisConfig
var ConfigViper *viper.Viper
var ConfigRunViper *viper.Viper

//  RedisWebSocket
var RedisWebSocket define.WebSocketStruct

func InitConfig() {
	initLog()

	//设置redisWebSocket配置
	configViper := viper.New()
	configViper.AddConfigPath(`config`)
	configViper.SetConfigName(`config`)
	configViper.SetConfigType(`ini`)
	if err := configViper.ReadInConfig(); err != nil {
		panic(`读取配置失败 config/config.ini`)
	}
	RedisWebSocket = define.WebSocketStruct{
		Host: configViper.GetString(`redisWebSocket.host`),
		Port: configViper.GetString(`redisWebSocket.port`),
	}

	initRedis()

}

// initRedis 初始化redis
// @author frog
// @date 2022-04-11 16:11:11
func initRedis() {
	RedisList = make([]define.RedisConfig, 0)
	ConfigViper = viper.New()
	ConfigViper.AddConfigPath(`config`)
	ConfigViper.SetConfigName(`config`)
	ConfigViper.SetConfigType(`ini`)
	if err := ConfigViper.ReadInConfig(); err != nil {
		panic(`读取配置失败 config/config.ini`)
	}

}

func initLog() {
	l, _ := log.ParseLevel(log.DebugLevel.String())
	log.SetLevel(l)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}
