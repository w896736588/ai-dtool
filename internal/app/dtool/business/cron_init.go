package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"strings"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

var cronTaskFuncRegistry = map[string]func(){}

func StartCronScheduler() {
	seedDefaultCronTasks()
	list, err := common.DbMain.CronTaskList()
	if err != nil {
		gstool.FmtPrintlnLogTime(`定时任务列表读取失败 %s`, err.Error())
		return
	}
	for _, row := range list {
		taskType := cast.ToString(row[`type`])
		enabled := cast.ToInt(row[`enabled`]) == 1
		triggerTime := strings.TrimSpace(cast.ToString(row[`trigger_time`]))
		startCronSchedulerByType(taskType, enabled, triggerTime)
	}
}

func StopCronScheduler() {
	for taskType, scheduler := range component.CronSchedulers {
		scheduler.Stop()
		delete(component.CronSchedulers, taskType)
	}
}

func startCronSchedulerByType(taskType string, enabled bool, triggerTime string) {
	if old, ok := component.CronSchedulers[taskType]; ok {
		old.Stop()
		delete(component.CronSchedulers, taskType)
	}
	if !enabled || triggerTime == `` {
		return
	}
	taskFunc, ok := component.CronTaskFuncRegistry[taskType]
	if !ok {
		taskFunc, ok = cronTaskFuncRegistry[taskType]
		if !ok {
			return
		}
	}
	scheduler := common.NewCronScheduler()
	scheduler.SetTaskFunc(taskFunc)
	scheduler.Configure(true, triggerTime)
	component.CronSchedulers[taskType] = scheduler
}

func seedDefaultCronTasks() {
	// 先彻底清理已废弃的兜底同步定时任务记录（如“同步知识片段（兜底）”“同步主库（兜底）”）。
	cleanupObsoleteCronTasks()
	for taskType, def := range define.CronTaskRegistry {
		one, _ := common.DbMain.CronTaskByType(taskType)
		if cast.ToInt(one[`id`]) > 0 {
			continue
		}
		_ = common.DbMain.CronTaskSave(taskType, def.Name, 0, ``)
	}
}

// cleanupObsoleteCronTasks 彻底删除 define.DeprecatedCronTaskTypes 中已废弃的定时任务记录，
// 避免这些无用的兜底同步任务残留在数据库与 Schedule 界面中。
func cleanupObsoleteCronTasks() {
	for _, taskType := range define.DeprecatedCronTaskTypes {
		if _, err := common.DbMain.Client.QuickDelete(`tbl_cron_task`, map[string]any{
			`type`: taskType,
		}).Exec(); err != nil {
			gstool.FmtPrintlnLogTime(`删除废弃定时任务 %s 失败 %s`, taskType, err.Error())
		}
	}
}
