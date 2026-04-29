package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
)

const (
	// homeTaskDailyReportDefaultPrompt 定义首页工作日报的默认提示词。
	homeTaskDailyReportDefaultPrompt = "请基于任务清单生成中文工作日报，按完成进展、进行中事项、阻塞风险进行总结，输出 Markdown。禁止编造未提供的信息。"
	// homeTaskDailyReportTitlePrefix 定义写入记忆时的标题前缀。
	homeTaskDailyReportTitlePrefix = "工作日报"
	// homeTaskDailyReportDateLayout 定义日报标题和正文中的日期格式。
	homeTaskDailyReportDateLayout = "2006-01-02"
	// homeTaskDailyReportEmptyText 统一空值展示文案，避免 prompt 中出现空串歧义。
	homeTaskDailyReportEmptyText = "-"
	// homeTaskDailyReportNoTaskError 表示当前没有可用于生成日报的活跃任务。
	homeTaskDailyReportNoTaskError = "暂无可生成日报的活跃任务"
	// homeTaskDailyReportSystemPromptText 约束模型只基于任务清单输出日报正文。
	homeTaskDailyReportSystemPromptText = "你是一个工作日报整理助手。\n你的任务是基于给定的任务清单生成中文 Markdown 工作日报。\n禁止编造未提供的事项、进度、风险或结论。\n输出必须为最终日报正文，不要附加解释。"
	// homeTaskDailyReportMemoryTag 定义写入记忆时的固定标签。
	homeTaskDailyReportMemoryTag = "工作日报"
	// homeTaskDailyReportModelRequiredError 表示日报模型尚未配置。
	homeTaskDailyReportModelRequiredError = "请先在记忆设置中配置工作日报模型"
	// homeTaskDailyReportModelUnavailableError 表示已配置的日报模型不可用。
	homeTaskDailyReportModelUnavailableError = "当前工作日报模型不可用"
	// homeTaskDailyReportModelTypeError 表示日报模型类型不符合预期。
	homeTaskDailyReportModelTypeError = "工作日报仅支持 LLM 模型"
)

// defaultHomeTaskDailyReportPrompt 返回首页工作日报默认提示词。
func defaultHomeTaskDailyReportPrompt() string {
	return homeTaskDailyReportDefaultPrompt
}

// buildHomeTaskDailyReportTitle 生成写入记忆的日报标题。
func buildHomeTaskDailyReportTitle(reportTime time.Time) string {
	return fmt.Sprintf("%s %s", homeTaskDailyReportTitlePrefix, reportTime.Format(homeTaskDailyReportDateLayout))
}

// buildHomeTaskDailyReportTasksSnapshot 将活跃任务列表格式化为可直接提交给模型的结构化文本。
func buildHomeTaskDailyReportTasksSnapshot(taskList []map[string]any) (string, error) {
	if len(taskList) == 0 {
		return "", errors.New(homeTaskDailyReportNoTaskError)
	}
	lineList := []string{
		fmt.Sprintf("任务总数：%d", len(taskList)),
	}
	for index, task := range taskList {
		lineList = append(lineList, "")
		lineList = append(lineList, fmt.Sprintf("任务 %d", index+1))
		lineList = append(lineList, fmt.Sprintf("- 名称：%s", homeTaskDailyReportValue(task["name"])))
		lineList = append(lineList, fmt.Sprintf("- 状态：%s", homeTaskDailyReportValue(task["task_status"])))
		lineList = append(lineList, fmt.Sprintf("- 开始时间：%s", homeTaskDailyReportValue(task["start_time_desc"])))
		lineList = append(lineList, fmt.Sprintf("- 最后操作时间：%s", homeTaskDailyReportValue(task["last_operated_at_desc"])))
		lineList = append(lineList, fmt.Sprintf("- 备注：%s", homeTaskDailyReportValue(task["remark"])))
	}
	return strings.Join(lineList, "\n"), nil
}

// buildHomeTaskDailyReportUserPrompt 组合用户配置提示词与任务快照，供大模型直接生成日报。
func buildHomeTaskDailyReportUserPrompt(prompt string, taskList []map[string]any, reportTime time.Time) (string, error) {
	taskSnapshot, err := buildHomeTaskDailyReportTasksSnapshot(taskList)
	if err != nil {
		return "", err
	}
	lineList := []string{
		strings.TrimSpace(prompt),
		"",
		fmt.Sprintf("日期：%s", reportTime.Format(homeTaskDailyReportDateLayout)),
		taskSnapshot,
	}
	return strings.TrimSpace(strings.Join(lineList, "\n")), nil
}

// mergeHomeTaskDailyReportTaskList 合并日报使用的任务列表，默认把已归档任务追加到未归档任务之后。
func mergeHomeTaskDailyReportTaskList(activeTaskList, archivedTaskList []map[string]any) []map[string]any {
	merged := make([]map[string]any, 0, len(activeTaskList)+len(archivedTaskList))
	merged = append(merged, activeTaskList...)
	merged = append(merged, archivedTaskList...)
	return merged
}

// homeTaskDailyReportSystemPrompt 返回首页工作日报固定 system prompt。
func homeTaskDailyReportSystemPrompt() string {
	return homeTaskDailyReportSystemPromptText
}

// homeTaskDailyReportConfig 读取并校验首页工作日报配置。
func homeTaskDailyReportConfig() (int, string, error) {
	modelIDText, err := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigDailyReportModelID)
	if err != nil && !common.DbRowMissing(err) {
		return 0, "", err
	}
	modelID := cast.ToInt(modelIDText)
	if modelID <= 0 {
		return 0, "", errors.New(homeTaskDailyReportModelRequiredError)
	}
	modelInfo, err := common.DbMain.AiModelInfo(modelID)
	if err != nil {
		return 0, "", errors.New(homeTaskDailyReportModelUnavailableError)
	}
	if strings.ToLower(cast.ToString(modelInfo["model_type"])) != "llm" {
		return 0, "", errors.New(homeTaskDailyReportModelTypeError)
	}
	prompt, err := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigDailyReportPrompt)
	if err != nil && !common.DbRowMissing(err) {
		return 0, "", err
	}
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		prompt = defaultHomeTaskDailyReportPrompt()
	}
	return modelID, prompt, nil
}

// homeTaskDailyReportValue 统一处理 prompt 中的空值展示。
func homeTaskDailyReportValue(raw any) string {
	value := strings.TrimSpace(cast.ToString(raw))
	if value == "" {
		return homeTaskDailyReportEmptyText
	}
	return value
}
