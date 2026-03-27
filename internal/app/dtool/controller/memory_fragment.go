package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// MemoryFragmentStatus 返回记忆库配置状态。
func MemoryFragmentStatus(c *gin.Context) {
	config := common.MemoryRuntime.Config()
	lastPushTime := common.MemoryRuntime.LastPushTime()
	lastPushError := common.MemoryRuntime.LastPushError()
	lastPushTimeDesc := `-`
	if lastPushTime > 0 {
		lastPushTimeDesc = gstool.TimeUnixToString(time.Unix(lastPushTime, 0), `Y-m-d H:i:s`)
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`configured`:          common.MemoryRuntime.IsConfigured(),
		`memory_dir`:          config.Dir,
		`memory_db_name`:      config.DBName,
		`is_git_repo`:         config.IsGitRepo,
		`last_push_time`:      lastPushTime,
		`last_push_time_desc`: lastPushTimeDesc,
		`last_push_error`:     lastPushError,
	})
}

// MemoryFragmentList 查询知识片段列表。
func MemoryFragmentList(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	list, err := memoryDB.MemoryFragmentList(cast.ToInt(dataMap[`limit`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentInfo 查询单个知识片段详情。
func MemoryFragmentInfo(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	info, err := memoryDB.MemoryFragmentInfo(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// MemoryFragmentSave 保存知识片段。
func MemoryFragmentSave(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	info, err := memoryDB.MemoryFragmentSave(
		cast.ToInt(dataMap[`id`]),
		cast.ToString(dataMap[`title`]),
		cast.ToString(dataMap[`content`]),
		memoryFragmentParseTags(dataMap[`tags`]),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	common.MemoryRuntime.ScheduleSync()
	gsgin.GinResponseSuccess(c, ``, info)
}

// MemoryFragmentDelete 软删除知识片段。
func MemoryFragmentDelete(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	_, err := memoryDB.MemoryFragmentSoftDelete(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	common.MemoryRuntime.ScheduleSync()
	gsgin.GinResponseSuccess(c, ``, nil)
}

// MemoryFragmentTrashList 查询回收站中的知识片段。
func MemoryFragmentTrashList(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	list, err := memoryDB.MemoryFragmentTrashList(cast.ToInt(dataMap[`limit`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentRestore 从回收站恢复知识片段。
func MemoryFragmentRestore(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	_, err := memoryDB.MemoryFragmentRestore(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	common.MemoryRuntime.ScheduleSync()
	gsgin.GinResponseSuccess(c, ``, nil)
}

// MemoryFragmentHardDelete 彻底删除回收站中的知识片段。
func MemoryFragmentHardDelete(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	if err := memoryDB.MemoryFragmentHardDelete(cast.ToInt(dataMap[`id`])); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	common.MemoryRuntime.ScheduleSync()
	gsgin.GinResponseSuccess(c, ``, nil)
}

// MemoryFragmentHistoryList 查询知识片段历史记录。
func MemoryFragmentHistoryList(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	list, err := memoryDB.MemoryFragmentHistoryList(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentTagList 查询知识片段标签列表。
func MemoryFragmentTagList(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	list, err := memoryDB.MemoryFragmentTagList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentSearch 搜索知识片段。
func MemoryFragmentSearch(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	list, err := memoryDB.MemoryFragmentSearch(
		cast.ToString(dataMap[`mode`]),
		cast.ToString(dataMap[`query`]),
		memoryFragmentParseTags(dataMap[`selected_tags`]),
		cast.ToInt(dataMap[`limit`]),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentOrganize 调用 AI 对当前片段内容进行整理。
func MemoryFragmentOrganize(c *gin.Context) {
	_, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	content := cast.ToString(dataMap[`content`])
	if strings.TrimSpace(content) == `` {
		gsgin.GinResponseError(c, `片段内容不能为空`, nil)
		return
	}
	modelID, prompt, err := memoryArrangeConfig()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	title := strings.TrimSpace(cast.ToString(dataMap[`title`]))
	userPrompt := buildMemoryArrangeUserPrompt(prompt, title, content)
	result, modelInfo, err := common.DbMain.InfoCrawlChatByModel(modelID, memoryArrangeSystemPrompt(), userPrompt)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`content`:  stripMarkdownCodeFence(result),
		`prompt`:   prompt,
		`model`:    cast.ToString(modelInfo[`model`]),
		`model_id`: modelID,
	})
}

func memoryDBOrResponse(c *gin.Context) (*common.CSqlite, bool) {
	if err := common.MemoryRuntime.EnsureConfigured(); err != nil {
		gsgin.GinResponseError(c, err.Error(), map[string]any{
			`configured`: false,
		})
		return nil, false
	}
	return common.MemoryRuntime.DB(), true
}

// memoryFragmentParseTags 解析请求中的标签数组。
func memoryFragmentParseTags(raw any) []string {
	switch value := raw.(type) {
	case []string:
		return value
	case []any:
		result := make([]string, 0, len(value))
		for _, item := range value {
			result = append(result, cast.ToString(item))
		}
		return result
	case string:
		if value == `` {
			return []string{}
		}
		return []string{value}
	default:
		return []string{}
	}
}

func defaultMemoryArrangePrompt() string {
	return `帮我把当前markdown进行整理格式，让它看起来更顺畅清晰，注意禁止修改内容`
}

func memoryArrangeSystemPrompt() string {
	return "你是一个 Markdown 整理助手。\n" +
		"你的任务仅限于整理格式、结构、段落、标题层级、列表、标点和可读性。\n" +
		"禁止新增事实、删除事实、改写原意、补充未提供的信息。\n" +
		"输出必须只包含整理后的 Markdown 正文，不要解释，不要加额外说明。"
}

func memoryArrangeConfig() (int, string, error) {
	modelIDText, err := common.DbMain.GlobalValue(define.GlobalMemoryArrangeModelID)
	if err != nil && !memoryConfigValueMissing(err) {
		return 0, ``, err
	}
	modelID := cast.ToInt(modelIDText)
	if modelID <= 0 {
		return 0, ``, gstool.Error(`请先在记忆设置中配置 AI 整理模型`)
	}
	modelInfo, err := common.DbMain.InfoCrawlAiModelInfo(modelID)
	if err != nil {
		return 0, ``, gstool.Error(`当前记忆整理模型不可用`)
	}
	if strings.ToLower(cast.ToString(modelInfo[`model_type`])) != `llm` {
		return 0, ``, gstool.Error(`记忆整理仅支持 LLM 模型`)
	}
	prompt, err := common.DbMain.GlobalValue(define.GlobalMemoryArrangePrompt)
	if err != nil && !memoryConfigValueMissing(err) {
		return 0, ``, err
	}
	prompt = strings.TrimSpace(prompt)
	if prompt == `` {
		prompt = defaultMemoryArrangePrompt()
	}
	return modelID, prompt, nil
}

func buildMemoryArrangeUserPrompt(prompt, title, content string) string {
	builder := strings.Builder{}
	builder.WriteString("整理要求：\n")
	builder.WriteString(strings.TrimSpace(prompt))
	if strings.TrimSpace(title) != `` {
		builder.WriteString("\n\n片段标题：\n")
		builder.WriteString(strings.TrimSpace(title))
	}
	builder.WriteString("\n\n当前 Markdown 内容如下，请直接输出整理后的完整 Markdown：\n```markdown\n")
	builder.WriteString(content)
	builder.WriteString("\n```")
	return builder.String()
}

func stripMarkdownCodeFence(content string) string {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "```") {
		return content
	}
	lineList := strings.Split(content, "\n")
	if len(lineList) < 2 {
		return content
	}
	if strings.HasPrefix(strings.TrimSpace(lineList[0]), "```") && strings.TrimSpace(lineList[len(lineList)-1]) == "```" {
		return strings.TrimSpace(strings.Join(lineList[1:len(lineList)-1], "\n"))
	}
	return content
}
