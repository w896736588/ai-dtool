package zhima

import (
	"dev_tool/base"
	_default "dev_tool/internal/app/default"
)

func initRouter(tGin *base.Gin) {
	_default.InitRouter(tGin)
}
