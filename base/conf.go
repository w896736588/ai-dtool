package base

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"redis_manager/define"
	"redis_manager/helper"
	"time"
)

var RedisList *[]define.RedisConfig
var ConfigViper *viper.Viper
var ConfigRunViper *viper.Viper

func InitConfig() {
	initLog()
	tempRedisList := make([]define.RedisConfig, 0)
	RedisList = &tempRedisList
	ConfigViper = viper.New()
	ConfigViper.AddConfigPath(`config`)
	ConfigViper.SetConfigName(`config`)
	ConfigViper.SetConfigType(`ini`)
	if err := ConfigViper.ReadInConfig(); err != nil {
		panic(`读取配置失败 config/config.ini`)
	}
	log.Debugf(`run_fike:%s`, ConfigViper.GetString(`run.file`))
	ConfigRunViper = viper.New()
	ConfigRunViper.AddConfigPath(`config`)
	ConfigRunViper.SetConfigName(ConfigViper.GetString(`run.file`))
	ConfigRunViper.SetConfigType(`ini`)
	if err := ConfigRunViper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf(`读取配置失败 config/%s.ini`, ConfigViper.GetString(`run.file`)))
	}
	allSettings := ConfigRunViper.AllSettings()
	log.Debugf(`allSettings:%#v`, allSettings)
	cTime := cast.ToInt(time.Now().Unix())
	for _, value := range allSettings {
		cTime++
		mapTemp := value.(map[string]interface{})
		UniKey := cast.ToString(mapTemp[`host`]) + cast.ToString(mapTemp[`password`]) + cast.ToString(mapTemp[`sshhost`]) + cast.ToString(mapTemp[`sshport`]) + cast.ToString(mapTemp[`sshuser`]) + cast.ToString(mapTemp[`sshpassword`]) + cast.ToString(mapTemp[`name`])
		*RedisList = append(*RedisList, define.RedisConfig{
			Name:        cast.ToString(mapTemp[`name`]),
			Host:        cast.ToString(mapTemp[`host`]),
			Password:    cast.ToString(mapTemp[`password`]),
			PoolSize:    cast.ToString(mapTemp[`poolsize`]),
			SshHost:     cast.ToString(mapTemp[`sshhost`]),
			SshPort:     cast.ToString(mapTemp[`sshport`]),
			SshUser:     cast.ToString(mapTemp[`sshuser`]),
			SshPassword: cast.ToString(mapTemp[`sshpassword`]),
			Default:     cast.ToInt(mapTemp[`default`]),
			UniKey:      helper.Md5(UniKey),
		})
	}
	log.Debugf(`redisList %#v`, RedisList)
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
