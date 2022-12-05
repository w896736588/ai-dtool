package websocket

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"redis_manager/base"
)
import "github.com/gorilla/websocket"

//升级tcp
var upGrader = websocket.Upgrader{
	//检查域名 是否支持跨域等
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	//读取的缓冲区大小
	ReadBufferSize: 1024,
	//写入缓冲区大小
	WriteBufferSize: 1024,
}

// Init 初始化redis 的 websocket
// @author frog
// @date 2022-04-20 09:37:37
func InitRedisWebSocket() {
	go func() {
		r := gin.Default()
		r.GET("/redisWebSocket", redisWebSocket)
		log.Debugf(`websocket %s`, base.RedisWebSocket.Host+`:`+base.RedisWebSocket.Port)
		err := r.Run(base.RedisWebSocket.Host + `:` + base.RedisWebSocket.Port)
		if err != nil {
			log.Errorf(`建立redisWebSocket失败 %s`, err.Error())
			panic(err.Error())
		}
	}()

}

// redisWebSocket
// @author frog
// @date 2022-04-20 10:43:49
func redisWebSocket(c *gin.Context) {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf(`redisWebSocket error %s`, err.Error())
		return
	}
	defer ws.Close()
	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Errorf(`redisWebSocket 收取消息失败 %s`, err.Error())
			break
		}
		responseMsg := RedisDistributeMsg(message)
		err = ws.WriteMessage(mt, responseMsg)
		if err != nil {
			log.Errorf(`redisWebSocket 发送消息失败 %s`, err.Error())
			break
		}
	}
}
