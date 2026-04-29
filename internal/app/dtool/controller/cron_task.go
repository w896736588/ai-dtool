package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

// CronDailyReportGenerate 是定时任务触发的日报生成入口，不依赖 gin.Context。
// 逻辑与 HomeTaskDailyReportGenerate 一致，但跳过 HTTP 响应处理。
func CronDailyReportGenerate() {
	if component.MemoryRuntime == nil {
		return
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		gstool.FmtPrintlnLogTime(`定时日报：记忆库未配置，跳过 %s`, err.Error())
		return
	}
	taskList, err := common.DbMain.HomeTaskListTodayUpdated()
	if err != nil {
		gstool.FmtPrintlnLogTime(`定时日报：读取今日变更任务失败 %s`, err.Error())
		return
	}
	reportTime := time.Now().Unix()
	if _, err = buildHomeTaskDailyReportTasksSnapshot(taskList); err != nil {
		gstool.FmtPrintlnLogTime(`定时日报：%s`, err.Error())
		return
	}
	if _, err = createAsyncTask(
		asyncTaskTypeHomeTaskDailyReport,
		buildHomeTaskDailyReportTitle(time.Unix(reportTime, 0)),
		``,
		map[string]any{
			`report_time`: reportTime,
			`task_count`:  len(taskList),
		},
		func(taskID int) {
			runAsyncTaskAndPersistResult(taskID, func() (map[string]any, error) {
				return buildAsyncHomeTaskDailyReportResult(taskList, reportTime)
			})
		},
	); err != nil {
		gstool.FmtPrintlnLogTime(`定时日报：创建异步任务失败 %s`, err.Error())
		return
	}
	gstool.FmtPrintlnLogTime(`定时日报：异步任务已创建`)
	_ = common.DbMain.CronTaskUpdateLastTriggerTime(define.CronTaskTypeDailyReport)
}
