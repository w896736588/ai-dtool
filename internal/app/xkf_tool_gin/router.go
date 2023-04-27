package xkf_tool_gin

import (
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"xkf_tool/internal/app/xkf_tool"
)

func InitRouter(host, port string) error {
	defer func() {
		if err := recover(); err != nil {
			gstool.FmtPrintlnLog(`拦截失败 %#v`, err)
			xkf_tool.Logger.Errorf(`拦截失败 %#v`, err)
		}
	}()
	gsGin := gsgin.GSGin{
		Host: host,
		Port: port,
	}
	gsGin.CreateRouter()
	gsGin.SetAllow()

	//REDIS接口
	//展示所有的
	gsGin.GinH.POST(`/api/redis/list`, RedisList)
	//查询某个key
	gsGin.GinH.POST(`/api/search`, Search)
	//模糊搜索key
	gsGin.GinH.POST(`/api/keys`, Keys)
	//批量获取缓存类型
	gsGin.GinH.POST(`/api/keys/type`, KeysType)
	//获取key类型
	gsGin.GinH.POST(`/api/key/type`, GetKeyType)
	//序列化和反序列化
	gsGin.GinH.POST(`/api/serialize`, PhpSerialize)
	gsGin.GinH.POST(`/api/unserialize`, PhpUnSerialize)
	//保存string
	gsGin.GinH.POST(`/api/save/string`, SaveString)
	//删除key
	gsGin.GinH.POST(`/api/del/key`, DelKey)
	//删除sub key
	gsGin.GinH.POST(`/api/del/sub`, DelSub)
	//更改ttl
	gsGin.GinH.POST(`/api/edit/ttl`, EditTtl)
	//删除所有缓存
	gsGin.GinH.POST(`/api/delete/all`, DelAllKey)
	//创建缓存
	gsGin.GinH.POST(`/api/create`, CreateCache)
	//编辑二级缓存
	gsGin.GinH.POST(`/api/edit/sub`, EditSub)
	//找到所有启用的消费者服务
	//router.POST(`/api/supervisor/status`, SupervisorStatus)

	//shell exec
	gsGin.GinH.POST(`/api/shell/exec`, ShellExec)

	//前端页面
	gsGin.GinH.Static(`/static`, `./views/dist/static`)
	gsGin.GinH.LoadHTMLFiles(`views/dist/index.html`)
	gsGin.GinH.GET("/", func(context *gin.Context) {
		context.HTML(200, "index.html", nil)
	})
	return gsGin.GinH.Run(gsGin.Host + `:` + gsGin.Port)
}
