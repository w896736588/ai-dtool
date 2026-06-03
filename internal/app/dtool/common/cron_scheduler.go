package common

import (
	"fmt"
	"sync"
	"time"

	"github.com/w896736588/go-tool/gstool"
)

// CronScheduler 管理每天固定时间触发的定时任务。
// 复用 stoppableTimer + afterFunc 模式，但省略防抖和 git 同步逻辑。
type CronScheduler struct {
	mu          sync.Mutex
	enabled     bool
	triggerTime string // "HH:MM"
	timer       stoppableTimer
	afterFunc   func(time.Duration, func()) stoppableTimer
	taskFunc    func()
}

// NewCronScheduler 创建定时调度器实例。
func NewCronScheduler() *CronScheduler {
	return &CronScheduler{
		afterFunc: func(d time.Duration, f func()) stoppableTimer {
			return &timeTimer{timer: time.AfterFunc(d, f)}
		},
	}
}

// SetTaskFunc 注入定时触发时执行的函数。
func (h *CronScheduler) SetTaskFunc(fn func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.taskFunc = fn
}

// Configure 更新配置并重启定时器。enabled 为 false 时仅停止定时器。
func (h *CronScheduler) Configure(enabled bool, triggerTime string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stopTimer()
	h.enabled = enabled
	h.triggerTime = triggerTime
	if h.enabled && h.triggerTime != `` && h.taskFunc != nil {
		h.scheduleNext()
	}
	if h.enabled {
		gstool.FmtPrintlnLogTime(`定时任务已配置 触发时间=%s`, h.triggerTime)
	} else {
		gstool.FmtPrintlnLogTime(`定时任务已关闭`)
	}
}

// Stop 停止定时器。
func (h *CronScheduler) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stopTimer()
}

// scheduleNext 计算到今天/明天触发点的 duration 并注册 afterFunc。
// 调用前必须已持有 h.mu 锁。
func (h *CronScheduler) scheduleNext() {
	now := time.Now()
	targetTime, err := parseCronTime(h.triggerTime)
	if err != nil {
		gstool.FmtPrintlnLogTime(`定时任务时间解析失败 %s`, err.Error())
		return
	}
	todayTarget := time.Date(now.Year(), now.Month(), now.Day(), targetTime.Hour(), targetTime.Minute(), 0, 0, now.Location())
	var nextTrigger time.Time
	if now.Before(todayTarget) {
		nextTrigger = todayTarget
	} else {
		nextTrigger = todayTarget.Add(24 * time.Hour)
	}
	duration := time.Until(nextTrigger)
	gstool.FmtPrintlnLogTime(`定时任务下次触发 %s（%s 后）`, nextTrigger.Format(`2006-01-02 15:04:05`), duration.Round(time.Second))
	h.timer = h.afterFunc(duration, func() {
		h.mu.Lock()
		taskFunc := h.taskFunc
		h.mu.Unlock()
		if taskFunc != nil {
			gstool.FmtPrintlnLogTime(`定时任务触发 %s`, time.Now().Format(`2006-01-02 15:04:05`))
			taskFunc()
		}
		h.mu.Lock()
		if h.enabled {
			h.scheduleNext()
		}
		h.mu.Unlock()
	})
}

// stopTimer 停止当前定时器。调用前必须已持有 h.mu 锁。
func (h *CronScheduler) stopTimer() {
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
}

// parseCronTime 将 "HH:MM" 格式字符串解析为 time.Time（仅保留时分）。
func parseCronTime(text string) (time.Time, error) {
	parsed, err := time.Parse(`15:04`, text)
	if err != nil {
		return time.Time{}, fmt.Errorf(`时间格式无效 %q，期望 HH:MM`, text)
	}
	return parsed, nil
}
