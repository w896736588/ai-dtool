package base

import (
	"gitee.com/Sxiaobai/gs/gsssh"
	"github.com/gorilla/websocket"
	"strings"
)

type TSocket struct {
}

func (h *TSocket) BindSsh(clientId string, conn *websocket.Conn) {
	//当socket建立的时候 会通过client_id解析关联到ssh
	//关联逻辑为通过#分割后前两位重组且匹配上的就设置
	connParams := Component.TBase.ExplainCombineKey(clientId)
	Component.TShell.WalkShellList(func(uniqueKey string, gsShell *gsssh.SshConfig) {
		tempUniqueKey := Component.TBase.GetCombineKey(connParams[0], connParams[1])
		if strings.Index(uniqueKey, tempUniqueKey) == 0 { //以tempUniqueKey开头
			gsShell.SetSocket(conn)
		}
	})
}

func (h *TSocket) UnBindSsh(clientId string) {

}
