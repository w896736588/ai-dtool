package main

import (
	"fmt"
	"redis_manager/api/gin"
	"redis_manager/api/websocket"
	"redis_manager/base"
)

func main() {
	base.InitConfig()
	base.InitRedis()
	websocket.InitRedisWebSocket()
	router := gin.InitRouter()
	router.Run(fmt.Sprintf(`:%s`, base.ConfigViper.GetString(`run.port`)))
}
