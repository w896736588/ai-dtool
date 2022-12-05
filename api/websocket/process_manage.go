package websocket

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)
import "golang.org/x/net/websocket"

// init 初始化
// @author frog
// @date 2022-03-29 16:15:47
func Init() {
	http.Handle(`/super/config`, websocket.Handler(superConfigList))
	log.Debugf(`尝试建立socket 8888`)
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Errorf(`连接socket失败 %#v`, err.Error())
	} else {
		log.Debugf(`建立socket成功 8888`)
	}
}

// superConfigList
// @author frog
// @date 2022-03-29 16:23:42
func superConfigList(ws *websocket.Conn) {
	msg := make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		log.Errorf(`错误 %s`, err.Error())
	}
	log.Debugf(`出来了 %s`, msg[:n])
}
