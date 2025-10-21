package base

import (
	"net/url"

	"gitee.com/Sxiaobai/gs/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type Gin struct {
	gin   *gsgin.GSGin //API接口
	IsRun bool
}

func (h *Gin) GinInit(host, port string) *gsgin.GSGin {
	if h.gin != nil {
		return h.gin
	}
	gsGin := &gsgin.GSGin{
		Host: host,
		Port: cast.ToInt(port),
	}
	gsGin.CreateRouter()
	gsGin.GinH.Use(gin.Logger())
	gsGin.GinH.UseH2C = true
	h.gin = gsGin
	return h.gin
}

func (h *Gin) GinSetAllowCrossDomain() {
	h.gin.SetAllow()
}

func (h *Gin) SetMode(mode string) {
	gin.SetMode(mode)
}

func (h *Gin) GinPost(route string, call ...gin.HandlerFunc) {
	h.gin.GinH.POST(route, call...)
}

func (h *Gin) GinGet(route string, call ...gin.HandlerFunc) {
	h.gin.GinH.GET(route, call...)
}

func (h *Gin) GinAll(route string, call ...gin.HandlerFunc) {
	h.gin.GinH.GET(route, call...)
	h.gin.GinH.POST(route, call...)
}

func (h *Gin) GinStatic(route, root string) {
	h.gin.GinH.Static(route, root)
}

func (h *Gin) GinStaticFile(relativePath, filepath string) {
	h.gin.GinH.StaticFile(relativePath, filepath)
}

func (h *Gin) GinLoadHTMLFiles(file ...string) {
	h.gin.GinH.LoadHTMLFiles(file...)
}

func (h *Gin) GinRun() {
	h.gin.Run()
}

func (h *Gin) GinStop(waitSecond int64) error {
	return h.gin.Stop(waitSecond)
}

func (h *Gin) SseRoute(route string,
	openFunc func(urlValues url.Values, stopC chan int, c *gin.Context) *gsgin.Sse, closeFunc func(sse *gsgin.Sse)) {
	h.gin.SseRoute(route, true, openFunc, closeFunc)
}
