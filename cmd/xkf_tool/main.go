package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"xkf_tool/internal/app/xkf_tool"
	"xkf_tool/internal/app/xkf_tool_gin"
)

var sw sync.RWMutex

func main() {
	xkf_tool.InitConfig()
	//host port
	port := xkf_tool.ConfigViper.GetString(`run.port`)
	host := `localhost`
	err := xkf_tool_gin.InitRouter(host, port)
	if err != nil {
		xkf_tool.Logger.Errorf(`%s`, err.Error())
		return
	}
	xkf_tool.Logger.Errorf(`监听结束信号 `)
	sc := make(chan os.Signal)
	sl := []os.Signal{
		syscall.SIGHUP,  //热升级
		syscall.SIGINT,  //Ctrl+C
		syscall.SIGTERM, //结束程序
	}
	signal.Notify(sc, sl...)
	sig := <-sc
	xkf_tool.Logger.Infof(`Sign:` + sig.String())
}
