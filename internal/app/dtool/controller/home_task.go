package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// HomeTaskList 查询首页任务列表。
func HomeTaskList(c *gin.Context) {
	request := _struct.HomeTaskListRequest{}
	_ = gsgin.GinPostBody(c, &request)
	list, err := common.DbMain.HomeTaskList(request.IsArchived)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	enrichHomeTaskListWithMemoryFragment(list)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`task_list`: list,
	})
}

// HomeTaskSave 保存首页任务。
func HomeTaskSave(c *gin.Context) {
	request := _struct.HomeTaskSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	memoryFragmentID, err := ensureHomeTaskMemoryFragment(
		request.ID, request.Name, normalizeHomeTaskMemoryFragmentID(request.MemoryFragmentID),
		request.TapdUrl, request.ApiHost, request.ApiToken,
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	info, err := common.DbMain.HomeTaskSave(request.ID, request.Name, request.TaskStatus, request.StartTime, memoryFragmentID, request.TapdUrl)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	enrichHomeTaskListWithMemoryFragment([]map[string]any{info})

	// 新建任务且指定了 tapd_url 时，自动创建 TAPD 网页抓取异步任务。
	if strings.TrimSpace(request.TapdUrl) != `` && request.ID <= 0 && strings.TrimSpace(memoryFragmentID) != `` {
		taskName := strings.TrimSpace(request.Name)
		_, asyncErr := createAsyncTask(
			asyncTaskTypeHomeTaskTapdScrape,
			"抓取TAPD网页: "+taskName,
			cast.ToString(info[`id`]),
			map[string]any{
				`tapd_url`:           request.TapdUrl,
				`memory_fragment_id`: memoryFragmentID,
				`task_name`:          taskName,
			},
			func(asyncTaskID int) {
				runAsyncTaskAndPersistResult(asyncTaskID, func() (map[string]any, error) {
					return buildAsyncHomeTaskTapdScrapeResultWithLog(request.TapdUrl, memoryFragmentID, func(step, message string) {
						appendAsyncTaskRunLog(asyncTaskID, step, message)
						BroadcastAsyncTasksUpdate()
					})
				})
			},
		)
		if asyncErr != nil {
			gstool.FmtPrintlnLogTime(`[HomeTaskSave] 创建 TAPD 抓取异步任务失败 err=%s`, asyncErr.Error())
		}
	}

	gsgin.GinResponseSuccess(c, ``, info)
}

// HomeTaskArchiveToggle 切换首页任务归档状态。
func HomeTaskArchiveToggle(c *gin.Context) {
	request := _struct.HomeTaskArchiveToggleRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.HomeTaskArchiveToggle(request.ID, request.IsArchived)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	enrichHomeTaskListWithMemoryFragment([]map[string]any{info})
	gsgin.GinResponseSuccess(c, ``, info)
}

// HomeTaskStatusQuickUpdate 快捷切换首页任务状态。
func HomeTaskStatusQuickUpdate(c *gin.Context) {
	request := _struct.HomeTaskStatusQuickUpdateRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.HomeTaskStatusQuickUpdate(request.ID, request.TaskStatus)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	enrichHomeTaskListWithMemoryFragment([]map[string]any{info})
	gsgin.GinResponseSuccess(c, ``, info)
}

// HomeTaskDelete 删除首页任务。
func HomeTaskDelete(c *gin.Context) {
	request := _struct.HomeTaskDeleteRequest{}
	_ = gsgin.GinPostBody(c, &request)
	err := common.DbMain.HomeTaskDelete(request.ID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// HomeTaskDailyReportGenerate 创建首页工作日报异步任务。 // HomeTaskDailyReportGenerate creates an async home-task daily report task.
func HomeTaskDailyReportGenerate(c *gin.Context) {
	if _, ok := memoryDBOrResponse(c); !ok {
		return
	}
	taskList, err := common.DbMain.HomeTaskListTodayUpdated()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	reportTime := time.Now().Unix()
	if _, err = buildHomeTaskDailyReportTasksSnapshot(taskList); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskInfo, err := createAsyncTask(
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
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`task_id`:     taskInfo[`id`],
		`task_status`: taskInfo[`task_status`],
		`task_type`:   taskInfo[`task_type`],
	})
}

func ensureHomeTaskMemoryFragment(taskID int, taskName string, memoryFragmentID string, tapdUrl string, apiHost string, apiToken string) (string, error) {
	taskName = strings.TrimSpace(taskName)
	if taskName == `` {
		return ``, gstool.Error(`任务名称不能为空`)
	}
	if component.MemoryRuntime == nil {
		return ``, common.ErrMemoryNotConfigured
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``, err
	}
	memoryDB := component.MemoryRuntime.DB()
	if memoryFragmentID != `` {
		if _, infoErr := memoryDB.MemoryFragmentInfo(memoryFragmentID); infoErr != nil {
			return ``, infoErr
		}
		return memoryFragmentID, nil
	}
	if !shouldAutoCreateHomeTaskMemoryFragment(taskID, memoryFragmentID) {
		return ``, nil
	}
	fragmentContent := buildHomeTaskFragmentContent(taskName, tapdUrl, apiHost, apiToken)
	fragmentInfo, saveErr := memoryDB.MemoryFragmentSave(0, taskName, fragmentContent, []string{`需求`})
	if saveErr != nil {
		return ``, saveErr
	}
	component.MemoryRuntime.ScheduleSync()
	fragmentID := strings.TrimSpace(cast.ToString(fragmentInfo[`file_id`]))
	if fragmentID == `` || fragmentID == `0` {
		return ``, gstool.Error(`自动创建知识片段失败`)
	}
	return fragmentID, nil
}

func shouldAutoCreateHomeTaskMemoryFragment(taskID int, memoryFragmentID string) bool {
	return taskID <= 0 && strings.TrimSpace(memoryFragmentID) == ``
}

func buildHomeTaskFragmentContent(taskName string, tapdUrl string, apiHost string, apiToken string) string {
	content := "# " + taskName + "\n\n"
	promptTemplate, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigFragmentPrompt)
	promptTemplate = strings.TrimSpace(promptTemplate)
	if promptTemplate == `` {
		return content
	}
	resolved := promptTemplate
	resolved = strings.ReplaceAll(resolved, "{tapd_url}", tapdUrl)
	resolved = strings.ReplaceAll(resolved, "{api_host}", apiHost)
	resolved = strings.ReplaceAll(resolved, "{api_token}", apiToken)
	content += resolved + "\n"
	return content
}

func normalizeHomeTaskMemoryFragmentID(raw any) string {
	idText := strings.TrimSpace(cast.ToString(raw))
	if idText == `` || idText == `0` {
		return ``
	}
	return idText
}

func enrichHomeTaskListWithMemoryFragment(list []map[string]any) {
	if component.MemoryRuntime == nil || component.MemoryRuntime.EnsureConfigured() != nil {
		return
	}
	memoryDB := component.MemoryRuntime.DB()
	for _, item := range list {
		memoryFragmentID := normalizeHomeTaskMemoryFragmentID(item[`memory_fragment_id`])
		item[`memory_fragment_id`] = memoryFragmentID
		if memoryFragmentID == `` {
			item[`memory_fragment`] = map[string]any{}
			continue
		}
		info, infoErr := memoryDB.MemoryFragmentInfo(memoryFragmentID)
		if infoErr != nil {
			item[`memory_fragment`] = map[string]any{
				`id`:      memoryFragmentID,
				`file_id`: memoryFragmentID,
				`title`:   `关联片段不存在`,
				`tags`:    []string{},
				`content`: ``,
				`missing`: true,
			}
			continue
		}
		item[`memory_fragment`] = map[string]any{
			`id`:      info[`id`],
			`file_id`: cast.ToString(info[`file_id`]),
			`title`:   cast.ToString(info[`title`]),
			`tags`:    cast.ToStringSlice(info[`tags`]),
			`content`: cast.ToString(info[`content`]),
			`missing`: false,
		}
	}
}
