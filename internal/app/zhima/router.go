package zhima

import (
	"dev_tool/base"
	_default "dev_tool/internal/app/default"
	"dev_tool/internal/app/zhima/controller"
	"gitee.com/Sxiaobai/gs/gstool"
)

func initRouter() {
	_default.InitRouter()
	wechatKefuRouter()
	vipRouter()
}

// 微信客服相关
func wechatKefuRouter() {
	gstool.FmtPrintlnLogTime(`初始化微信客服接口 `)
	base.Component.TGin.GinPost(`/api/WechatKefuStatus`, controller.WechatKefuStatus)                 //查询当前应用启动情况
	base.Component.TGin.GinPost(`/api/WechatKefuChange`, controller.WechatKefuChange)                 //切换微信客服环境
	base.Component.TGin.GinPost(`/api/WechatKefuQueryAppList`, controller.WechatKefuQueryAppList)     //查询微信客服列表
	base.Component.TGin.GinPost(`/api/WechatKefuQueryQrCdeList`, controller.WechatKefuQueryQrCdeList) //查询渠道客服二维码列表
}

// vip相关
func vipRouter() {
	base.Component.TGin.GinPost(`/api/VipChange`, controller.VipChange) //切换vip
	base.Component.TGin.GinPost(`/api/VipQuery`, controller.VipQuery)   //查询vip
}
