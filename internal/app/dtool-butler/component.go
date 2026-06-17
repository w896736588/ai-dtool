package dtool_butler

import (
	"dev_tool/internal/app/dtool-butler/bot"
	"dev_tool/internal/app/dtool-butler/butler"
	"dev_tool/internal/app/dtool-butler/define"
	"dev_tool/internal/app/dtool/common"
)

// 全局实例，管家进程内共享。
var (
	Env        *define.Env
	DbMain     *common.CSqlite
	BotGateway bot.Gateway
	ButlerCore *butler.Core
)
