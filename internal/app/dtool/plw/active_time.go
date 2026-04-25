package plw

import (
	"dev_tool/internal/app/dtool/struct"
	"github.com/playwright-community/playwright-go"
	"sync"
	"time"
)

var pageActives = make(map[string]_struct.PageActiveTime)
var pageActivesLock sync.RWMutex
var pageActivesLoopOnce sync.Once

type PageActiveTime struct {
}

func InitPageActiveTime() {
	// 先确保活跃页容器存在，再启动一次性清理协程。
	ensurePageActivesInitialized()
	pageActivesLoopOnce.Do(func() {
		go func() {
			for {
				time.Sleep(time.Second)
				pageActivesLock.Lock()
				newMap := make(map[string]_struct.PageActiveTime)
				for pageUrl, activeTime := range pageActives {
					if activeTime.ActiveTime.Add(time.Second * time.Duration(activeTime.AutoCloseSecond)).Before(time.Now()) {
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
	})

}

func NewPageActiveTime() *PageActiveTime {
	// 创建实例前确保全局 map 已初始化，避免遗漏初始化时写入 nil map。
	ensurePageActivesInitialized()
	return &PageActiveTime{}
}

func (h *PageActiveTime) Add(page *playwright.Page, autoCloseSecond int) {
	go func() {
		pageActivesLock.Lock()
		defer pageActivesLock.Unlock()
		if pageActives == nil {
			pageActives = make(map[string]_struct.PageActiveTime)
		}
		pageActives[(*page).URL()] = _struct.PageActiveTime{
			ActiveTime:      time.Now(),
			AutoCloseSecond: autoCloseSecond,
			Page:            page,
		}
	}()
}

// ensurePageActivesInitialized 确保活跃页 map 已初始化，避免写入 nil map。
func ensurePageActivesInitialized() {
	pageActivesLock.Lock()
	defer pageActivesLock.Unlock()
	if pageActives != nil {
		return
	}
	pageActives = make(map[string]_struct.PageActiveTime)
}
