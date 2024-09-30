package base

import (
	"gitee.com/Sxiaobai/gs/gsdefine"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gorilla/websocket"
	"runtime/debug"
	"strings"
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
	//当socket建立的时候 会通过client_id解析关联到ssh
	//关联逻辑为通过#分割后前两位重组且匹配上的就设置
	connParams := Component.TBase.ExplainCombineKey(clientId)
	if len(connParams) != 2 {
		gstool.FmtPrintlnLogTime(`错误的clientId %s %s %s`, clientId, gsdefine.Error, debug.Stack())
		return
	}
	Component.TShell.WalkShellList(func(uniqueKey string, gsShell *gsssh.SshConfig) {
		tempUniqueKey := Component.TBase.GetCombineKey(connParams[0], connParams[1])
		if strings.Index(uniqueKey, tempUniqueKey) == 0 { //以tempUniqueKey开头
			gsShell.SetSocket(conn)
		}
	})
}

func (h *TSocket) UnBindSsh(clientId string) {
	defer h.lock.Unlock()
	h.lock.Lock()
	delete(h.SocketList, clientId)
	connParams := Component.TBase.ExplainCombineKey(clientId)
	Component.TShell.WalkShellList(func(uniqueKey string, gsShell *gsssh.SshConfig) {
		tempUniqueKey := Component.TBase.GetCombineKey(connParams[0], connParams[1])
		if strings.Index(uniqueKey, tempUniqueKey) == 0 { //以tempUniqueKey开头
			gsShell.SetSocket(nil)
		}
	})
}

func (h *TSocket) GetSocket(clientId string) *websocket.Conn {
	defer h.lock.Unlock()
	h.lock.Lock()
	return h.SocketList[clientId]
}
