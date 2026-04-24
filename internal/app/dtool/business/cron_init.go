package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// StartCronScheduler 从数据库读取定时任务配置并启动调度器。
func StartCronScheduler() {
	if component.CronScheduler == nil {
		return
	}
	enabled, triggerTime, err := readCronConfig()
	if err != nil {
		gstool.FmtPrintlnLogTime(`定时任务配置读取失败 %s`, err.Error())
		return
	}
	component.CronScheduler.Configure(enabled, triggerTime)
}

// StopCronScheduler 停止定时调度器。
func StopCronScheduler() {
	if component.CronScheduler != nil {
		component.CronScheduler.Stop()
	}
}

// readCronConfig 从 tbl_cron_task 读取定时任务配置。
func readCronConfig() (bool, string, error) {
	one, err := common.DbMain.CronTaskByType(define.CronTaskTypeDailyReport)
	if err != nil && !common.DbRowMissing(err) {
		return false, "", err
	}
	enabled := cast.ToInt(one[`enabled`]) == 1
	triggerTime := strings.TrimSpace(cast.ToString(one[`trigger_time`]))
	return enabled, triggerTime, nil
}
