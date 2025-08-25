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

// ProcessBoolResult 用于bool_result类型判断
type ProcessBoolResult struct {
	Locator     string `json:"locator"` //查找的元素
	ExistReturn bool   `json:"return"`  //如果有out_key 这个元素存在时返回什么
}

type ProcessRedirect struct {
	Url                 string `json:"url"`
	RegisterResponseUrl []struct {
		Url        string `json:"url"`
		WaitSecond int    `json:"wait_second"`
	} `json:"register_response_url"`
}

type ProcessResponseUrl struct {
	Url string `json:"url"`
}

type ProcessWaitUrl struct {
	ResponseUrl string `json:"response_url"`
	WaitSecond  int    `json:"wait_second"`
}

type StreamJson struct {
	Op          string `json:"op"`
	Mask        string `json:"mask"`
	EventOffset int    `json:"eventOffset"`
	Block       struct {
		Id   string `json:"id"`
		Text struct {
			Content string `json:"content"`
		} `json:"text"`
	} `json:"block"`
}
