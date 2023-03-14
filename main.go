package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"redis_manager/api/gin"
	"redis_manager/base"
)

func main() {
	base.InitConfig()
	base.InitRedis()
	router := gin.InitRouter()
	err := router.Run(fmt.Sprintf(`:%s`, base.ConfigViper.GetString(`run.port`)))
	if err != nil {
		log.Errorf(`%s`, err.Error())
		return
	}
}
