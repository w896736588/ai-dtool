package controller

import (
	"strings"
	"testing"
	"time"
)

func TestMergeHomeTaskDailyReportTaskList_IncludesArchivedTasks(t *testing.T) {
	activeTaskList := []map[string]any{
		{
			`id`:   1,
			`name`: `进行中的任务`,
		},
	}
	archivedTaskList := []map[string]any{
		{
			`id`:   2,
			`name`: `已归档任务`,
		},
	}

	got := mergeHomeTaskDailyReportTaskList(activeTaskList, archivedTaskList)
	if len(got) != 2 {
		t.Fatalf("mergeHomeTaskDailyReportTaskList() len = %d, want 2", len(got))
	}
	if got[0][`name`] != `进行中的任务` {
		t.Fatalf("mergeHomeTaskDailyReportTaskList() first task = %v, want active task first", got[0][`name`])
	}
	if got[1][`name`] != `已归档任务` {
		t.Fatalf("mergeHomeTaskDailyReportTaskList() second task = %v, want archived task appended", got[1][`name`])
	}
}

func TestDefaultHomeTaskDailyReportPrompt_NotEmpty(t *testing.T) {
	if strings.TrimSpace(defaultHomeTaskDailyReportPrompt()) == `` {
		t.Fatalf("defaultHomeTaskDailyReportPrompt() returned empty text")
	}
}

func TestBuildHomeTaskDailyReportTitle(t *testing.T) {
	reportTime := time.Date(2026, time.March, 24, 10, 30, 0, 0, time.Local)
	got := buildHomeTaskDailyReportTitle(reportTime)
	want := `工作日报 2026-03-24`
	if got != want {
		t.Fatalf("buildHomeTaskDailyReportTitle() = %q, want %q", got, want)
	}
}

func TestBuildHomeTaskDailyReportTasksSnapshot(t *testing.T) {
	taskList := []map[string]any{
		{
			`name`:                  `首页日报`,
			`task_status`:           `开发中`,
			`start_time_desc`:       `2026-03-24`,
			`last_operated_at_desc`: `2026-03-24 14:20:00`,
			`remark`:                `补充日报入口`,
		},
	}
	got, err := buildHomeTaskDailyReportTasksSnapshot(taskList)
	if err != nil {
		t.Fatalf("buildHomeTaskDailyReportTasksSnapshot() unexpected err = %v", err)
	}
	wantContains := []string{
		`任务总数：1`,
		`名称：首页日报`,
		`状态：开发中`,
		`开始时间：2026-03-24`,
		`最后操作时间：2026-03-24 14:20:00`,
		`备注：补充日报入口`,
	}
	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Fatalf("buildHomeTaskDailyReportTasksSnapshot() missing %q\nfull snapshot:\n%s", want, got)
		}
	}
}

func TestBuildHomeTaskDailyReportTasksSnapshot_EmptyList(t *testing.T) {
	_, err := buildHomeTaskDailyReportTasksSnapshot(nil)
	if err == nil {
		t.Fatalf("buildHomeTaskDailyReportTasksSnapshot() expected err for empty list")
	}
	if !strings.Contains(err.Error(), `暂无可生成日报的活跃任务`) {
		t.Fatalf("buildHomeTaskDailyReportTasksSnapshot() err = %v", err)
	}
}

func TestBuildHomeTaskDailyReportUserPrompt(t *testing.T) {
	taskList := []map[string]any{
		{
			`name`:                  `任务A`,
			`task_status`:           `待处理`,
			`start_time_desc`:       `-`,
			`last_operated_at_desc`: `2026-03-24 09:00:00`,
			`remark`:                ``,
		},
	}
	got, err := buildHomeTaskDailyReportUserPrompt(`请按进展总结并输出 Markdown`, taskList, time.Date(2026, time.March, 24, 9, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("buildHomeTaskDailyReportUserPrompt() unexpected err = %v", err)
	}
	wantContains := []string{
		`请按进展总结并输出 Markdown`,
		`日期：2026-03-24`,
		`任务总数：1`,
		`名称：任务A`,
	}
	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Fatalf("buildHomeTaskDailyReportUserPrompt() missing %q\nfull prompt:\n%s", want, got)
		}
	}
}
