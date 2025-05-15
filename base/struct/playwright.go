package _struct

import (
	"dev_tool/base/define"
	"github.com/playwright-community/playwright-go"
	"time"
)

type ContextPage struct {
	Context            playwright.BrowserContext
	SmartLinkUniqueKey string          //选项唯一值  链接配置ID_label  记录是哪个类型的context 用于计数
	UserDataIndex      int             //数据目录索引
	UserDataPath       string          //数据目录
	ContextUnique      string          //唯一标记 context 记录是哪个目录的context
	OpenType           define.OpenType //打开类型
}

type PageActiveTime struct {
	ActiveTime      time.Time
	AutoCloseSecond int
	Page            *playwright.Page
}

type ProcessResult struct {
	Locator string `json:"locator"`
	Return  bool   `json:"return"`
}
