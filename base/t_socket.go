package base

import (
	"github.com/gorilla/websocket"
	"sync"
)

type TSocket struct {
	SocketList map[string]*websocket.Conn
	lock       sync.Mutex
}

func (h *TSocket) BindSsh(clientId string, conn *websocket.Conn) {
	defer h.lock.Unlock()
	h.lock.Lock()
	h.SocketList[clientId] = conn
}

func (h *TSocket) UnBindSsh(clientId string) {
	defer h.lock.Unlock()
	h.lock.Lock()
	delete(h.SocketList, clientId)
}

func (h *TSocket) GetSocket(clientId string) *websocket.Conn {
	defer h.lock.Unlock()
	h.lock.Lock()
	return h.SocketList[clientId]
}

func (h *TSocket) SendMsg(clientId, msg string) {
	defer func() {
		if err := recover(); err != nil {

		}
	}()
	socket := h.GetSocket(clientId)
	if socket == nil {
		return
	}
	_ = socket.WriteMessage(websocket.TextMessage, []byte(msg+"\n"))
}

func (h *TSocket) SendMsgReal(clientId, msg string) {
	defer func() {
		if err := recover(); err != nil {

		}
	}()
	socket := h.GetSocket(clientId)
	if socket == nil {
		return
	}
	_ = socket.WriteMessage(websocket.TextMessage, []byte(msg))
}
