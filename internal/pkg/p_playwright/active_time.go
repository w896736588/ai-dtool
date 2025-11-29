package p_playwright

import (
	"dev_tool/base"
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/playwright-community/playwright-go"
	"sync"
	"time"
)

var pageActives map[string]_struct.PageActiveTime
var pageActivesLock sync.RWMutex

type PageActiveTime struct {
}

func InitPageActiveTime() {
	pageActivesLock.Lock()
	pageActives = make(map[string]_struct.PageActiveTime)
	pageActivesLock.Unlock()
	go func() {
		for {
			time.Sleep(time.Second)
			pageActivesLock.Lock()
			newMap := make(map[string]_struct.PageActiveTime)
			for pageUrl, activeTime := range pageActives {
				if activeTime.ActiveTime.Add(time.Second * time.Duration(activeTime.AutoCloseSecond)).Before(time.Now()) {
					base.Component.TPlaywright.Log.Infof(`自动关闭页面 %s 设置的活跃时间 %d 上次活跃时间 %s 当前时间 %s`,
						pageUrl, activeTime.AutoCloseSecond,
						gstool.TimeUnixToString(activeTime.ActiveTime, `Y-md H:i:s`),
						gstool.TimeNowUnixToString(`Y-m-d H:i:s`))
					go func() {
						_ = (*activeTime.Page).Close()
					}()
				} else {
					newMap[pageUrl] = activeTime
				}
			}
			pageActives = newMap
			pageActivesLock.Unlock()
		}
	}()

}

func NewPageActiveTime() *PageActiveTime {
	return &PageActiveTime{}
}

func (h *PageActiveTime) Add(page *playwright.Page, autoCloseSecond int) {
	go func() {
		pageActivesLock.Lock()
		defer pageActivesLock.Unlock()
		pageActives[(*page).URL()] = _struct.PageActiveTime{
			ActiveTime:      time.Now(),
			AutoCloseSecond: autoCloseSecond,
			Page:            page,
		}
	}()

}
