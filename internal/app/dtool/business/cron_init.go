package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
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
	for taskType, def := range define.CronTaskRegistry {
		one, _ := common.DbMain.CronTaskByType(taskType)
		if cast.ToInt(one[`id`]) > 0 {
			continue
		}
		_ = common.DbMain.CronTaskSave(taskType, def.Name, 0, ``)
	}
}
