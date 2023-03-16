package base

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"redis_manager/define"
	"strings"
)

var RedisList []define.RedisConfig
var ConfigViper *viper.Viper
var EncryptMain *Encrypt //加密
var RedisConfigViper *viper.Viper

func InitConfig() {
	initLog()
	ConfigViper = viper.New()
	ConfigViper.AddConfigPath(`config`)
	ConfigViper.SetConfigName(`config`)
	ConfigViper.SetConfigType(`ini`)
	if err := ConfigViper.ReadInConfig(); err != nil {
		panic(`读取配置失败 config/config.ini`)
	}
	initEncrypt()
}

// initEncrypt 初始化
// @auth frog
// @date 2023-03-14 15:29:29
func initEncrypt() {
	EncryptMain = &Encrypt{
		Key: ConfigViper.GetString(`encrypt.key`),
		Iv:  ConfigViper.GetString(`encrypt.iv`),
	}
}

// initRedis 初始化redis
// @auth frog
// @date 2023-03-15 09:53:47
func initRedis() {
	RedisConfigViper = viper.New()
	RedisConfigViper.AddConfigPath(`config`)
	RedisConfigViper.SetConfigName(`redis`)
	RedisConfigViper.SetConfigType(`json`)
	if err := ConfigViper.ReadInConfig(); err != nil {
		panic(`读取配置失败 config/redis.json`)
	}
	redisGroupNames := ConfigViper.GetString(`redis.groupNames`)
	redisGroupNameList := strings.Split(redisGroupNames, `,`)
	for _, redisGroupName := range redisGroupNameList {
		log.Debugf(redisGroupName)

		redisConfigList := RedisConfigViper.Get(redisGroupName)
		log.Debugf(`redisConfig %#v`, redisConfigList)
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
