package controller

import (
	"archive/zip"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_define"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// buildMemoryFragmentStatusPayload 构造记忆库状态数据。
func buildMemoryFragmentStatusPayload() map[string]any {
	config := component.MemoryRuntime.Config()
	nextPushTime := component.MemoryRuntime.NextPushTime()
	lastPushTime := component.MemoryRuntime.LastPushTime()
	lastPushError := component.MemoryRuntime.LastPushError()
	indexReady := false
	fragmentCount := 0
	trashCount := 0
	if runtimeStore, ok := component.MemoryRuntime.DB().(interface {
		IndexReady() bool
		FragmentCount() int
		TrashCount() int
	}); ok {
		indexReady = runtimeStore.IndexReady()
		fragmentCount = runtimeStore.FragmentCount()
		trashCount = runtimeStore.TrashCount()
	}
	nextPushTimeDesc := `-`
	lastPushTimeDesc := `-`
	if nextPushTime > 0 {
		nextPushTimeDesc = gstool.TimeUnixToString(time.Unix(nextPushTime, 0), `Y-m-d H:i:s`)
	}
	if lastPushTime > 0 {
		lastPushTimeDesc = gstool.TimeUnixToString(time.Unix(lastPushTime, 0), `Y-m-d H:i:s`)
	}
	return map[string]any{
		`configured`:              component.MemoryRuntime.IsConfigured(),
		`memory_dir`:              config.Dir,
		`git_repo_enabled`:        config.GitRepoEnabled,
		`is_git_repo`:             config.IsGitRepo,
		`auto_push_delay_minutes`: config.AutoPushDelayMinutes,
		`index_ready`:             indexReady,
		`fragment_count`:          fragmentCount,
		`trash_count`:             trashCount,
		`next_push_time`:          nextPushTime,
		`next_push_time_desc`:     nextPushTimeDesc,
		`last_push_time`:          lastPushTime,
		`last_push_time_desc`:     lastPushTimeDesc,
		`last_push_error`:         lastPushError,
	}
}

// MemoryFragmentStatus 返回记忆库配置状态。
func MemoryFragmentStatus(c *gin.Context) {
	gsgin.GinResponseSuccess(c, ``, buildMemoryFragmentStatusPayload())
}

// sendMemoryFragmentStatusSnapshot 向指定 SSE 连接发送一次记忆库状态快照。
func sendMemoryFragmentStatusSnapshot(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	data := buildMemoryFragmentStatusPayload()
	err := sse.SendToChan(gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseMemoryFragmentStatus,
		Data:            data,
		Type:            p_define.SseContentTypeMsg,
	}))
	if err != nil {
		gstool.FmtPrintlnLogTime(`MemoryFragmentStatus广播错误 %s`, err.Error())
	}
}

// BindMemoryFragmentStatusSSE 为普通 SSE client 绑定记忆库状态推送。
func BindMemoryFragmentStatusSSE(sse *gsgin.Sse, stopC chan int, interval time.Duration) {
	if sse == nil {
		return
	}
	if interval <= 0 {
		interval = 10 * time.Second
	}
	// 建连后立即推一次，避免前端初次打开时要等下一个周期。
	sendMemoryFragmentStatusSnapshot(sse)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sendMemoryFragmentStatusSnapshot(sse)
			case <-stopC:
				return
			}
		}
	}()
}

// MemoryFragmentBatchInfoByPaths 批量按文件路径查询片段摘要。
func MemoryFragmentBatchInfoByPaths(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	pathsRaw, _ := dataMap[`paths`].([]any)
	paths := make([]string, 0, len(pathsRaw))
	for _, p := range pathsRaw {
		if s, ok := p.(string); ok {
			paths = append(paths, s)
		}
	}
	results := memoryDB.MemoryFragmentBatchInfoByPaths(paths)
	gsgin.GinResponseSuccess(c, ``, results)
}

// MemoryFragmentList 查询知识片段列表。
func MemoryFragmentList(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	limit := cast.ToInt(dataMap[`limit`])
	offset := cast.ToInt(dataMap[`offset`])
	if limit <= 0 {
		limit = 10
	}
	// 多查一条用于判断是否还有更多数据
	list, err := memoryDB.MemoryFragmentList(limit+1, offset)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	hasMore := len(list) > limit
	if hasMore {
		list = list[:limit]
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`:     list,
		`has_more`: hasMore,
	})
}

// MemoryFragmentInfo 查询单个知识片段详情。
func MemoryFragmentInfo(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	fragmentID := strings.TrimSpace(cast.ToString(dataMap[`id`]))
	if fragmentID == `` || fragmentID == `0` {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	info, err := memoryDB.MemoryFragmentInfo(fragmentID)
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
		cast.ToString(dataMap[`id`]),
		cast.ToString(dataMap[`title`]),
		cast.ToString(dataMap[`content`]),
		memoryFragmentParseTags(dataMap[`tags`]),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	broadcastMemoryFragmentUpsert(info)
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
	fragmentID := strings.TrimSpace(cast.ToString(dataMap[`id`]))
	if fragmentID == `` || fragmentID == `0` {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	_, err := memoryDB.MemoryFragmentSoftDelete(fragmentID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	broadcastMemoryFragmentDelete(fragmentID)
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
	fragmentID := strings.TrimSpace(cast.ToString(dataMap[`id`]))
	if fragmentID == `` || fragmentID == `0` {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	_, err := memoryDB.MemoryFragmentRestore(fragmentID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	if info, infoErr := memoryDB.MemoryFragmentInfo(fragmentID); infoErr == nil {
		broadcastMemoryFragmentUpsert(info)
	}
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
	fragmentID := strings.TrimSpace(cast.ToString(dataMap[`id`]))
	if fragmentID == `` || fragmentID == `0` {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	if err := memoryDB.MemoryFragmentHardDelete(fragmentID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	broadcastMemoryFragmentDelete(fragmentID)
	gsgin.GinResponseSuccess(c, ``, nil)
}

const (
	// memoryFragmentSseActionUpsert 表示知识片段新增或更新。 // memoryFragmentSseActionUpsert marks a fragment upsert event.
	memoryFragmentSseActionUpsert = `upsert`
	// memoryFragmentSseActionDelete 表示知识片段删除或移出当前列表。 // memoryFragmentSseActionDelete marks a fragment delete event.
	memoryFragmentSseActionDelete = `delete`
	// memoryFragmentSseStatusPrefix 是 gsgin.SseStatus 返回值里的 client_id 前缀。 // memoryFragmentSseStatusPrefix matches the client-id prefix returned by gsgin.SseStatus.
	memoryFragmentSseStatusPrefix = `ClientId:`
)

// broadcastMemoryFragmentUpsert 广播知识片段新增或更新事件。 // broadcastMemoryFragmentUpsert broadcasts a fragment upsert event.
func broadcastMemoryFragmentUpsert(fragment map[string]any) {
	fragmentID := strings.TrimSpace(cast.ToString(fragment[`id`]))
	if fragmentID == `` {
		fragmentID = strings.TrimSpace(cast.ToString(fragment[`file_id`]))
	}
	if fragmentID == `` {
		return
	}
	broadcastMemoryFragmentEvent(memoryFragmentSseActionUpsert, fragmentID, fragment)
}

// broadcastMemoryFragmentDelete 广播知识片段删除事件。 // broadcastMemoryFragmentDelete broadcasts a fragment delete event.
func broadcastMemoryFragmentDelete(fragmentID string) {
	normalizedID := strings.TrimSpace(fragmentID)
	if normalizedID == `` || normalizedID == `0` {
		return
	}
	broadcastMemoryFragmentEvent(memoryFragmentSseActionDelete, normalizedID, nil)
}

// broadcastMemoryFragmentEvent 把知识片段变更广播到所有普通 SSE 客户端。 // broadcastMemoryFragmentEvent pushes fragment changes to all normal SSE clients.
func broadcastMemoryFragmentEvent(action, fragmentID string, fragment map[string]any) {
	payload := map[string]any{
		`action`:      strings.TrimSpace(action),
		`fragment_id`: strings.TrimSpace(fragmentID),
	}
	if fragment != nil {
		payload[`fragment`] = fragment
	}
	msg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseMemoryFragmentUpdates,
		Data:            payload,
		Type:            p_define.SseContentTypeMsg,
	})
	// 中文注释：这里只复用全局普通 SSE 通道，避免为知识片段同步再单独维护一套长连接。
	// English comment: Reuse the shared SSE channel so fragment sync does not require a second long-lived connection.
	for _, item := range gsgin.SseStatus() {
		clientID := strings.TrimSpace(strings.TrimPrefix(item, memoryFragmentSseStatusPrefix))
		if clientID == `` || clientID == item {
			continue
		}
		sse := gsgin.SseGetByClientId(clientID)
		if sse == nil {
			continue
		}
		_ = sse.SendToChan(msg)
	}
}

// MemoryFragmentHistoryList 查询知识片段历史记录。
func MemoryFragmentHistoryList(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	fragmentID := strings.TrimSpace(cast.ToString(dataMap[`id`]))
	if fragmentID == `` || fragmentID == `0` {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	config := component.MemoryRuntime.Config()
	result := map[string]any{
		`list`:             []map[string]any{},
		`git_repo_enabled`: config.GitRepoEnabled,
		`is_git_repo`:      config.IsGitRepo,
		`history_source`:   `none`,
		`setting_hint`:     `请到“设置” -> “记忆设置”中开启 Git 管理（memoryDbIsGitRepo）后，再查看知识片段历史记录。`,
	}
	if !config.GitRepoEnabled || !config.IsGitRepo {
		gsgin.GinResponseSuccess(c, ``, result)
		return
	}
	list, err := memoryDB.MemoryFragmentHistoryList(fragmentID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	result[`list`] = list
	result[`history_source`] = `git`
	gsgin.GinResponseSuccess(c, ``, result)
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

// MemoryFragmentOrganize 创建知识片段整理异步任务。 // MemoryFragmentOrganize creates an async memory fragment arrange task.
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
	fragmentID := strings.TrimSpace(cast.ToString(dataMap[`id`]))
	title := strings.TrimSpace(cast.ToString(dataMap[`title`]))
	taskInfo, err := createAsyncTask(
		asyncTaskTypeMemoryFragmentArrange,
		`整理知识片段 `+title,
		fragmentID,
		map[string]any{
			`fragment_id`: fragmentID,
			`title`:       title,
			`content`:     content,
		},
		func(taskID int) {
			runAsyncTaskAndPersistResult(taskID, func() (map[string]any, error) {
				resultMap, buildErr := buildAsyncMemoryArrangeResult(title, content)
				if buildErr != nil {
					return nil, buildErr
				}
				resultMap[`fragment_id`] = fragmentID
				return resultMap, nil
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

func memoryDBOrResponse(c *gin.Context) (common.MemoryFragmentStore, bool) {
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		gsgin.GinResponseError(c, err.Error(), map[string]any{
			`configured`: false,
		})
		return nil, false
	}
	return component.MemoryRuntime.DB(), true
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
	modelIDText, err := common.DbMain.MemoryConfigValue(define.MemoryConfigArrangeModelID)
	if err != nil && !common.DbRowMissing(err) {
		return 0, ``, err
	}
	modelID := cast.ToInt(modelIDText)
	if modelID <= 0 {
		return 0, ``, gstool.Error(`请先在记忆设置中配置 AI 整理模型`)
	}
	modelInfo, err := common.DbMain.AiModelInfo(modelID)
	if err != nil {
		return 0, ``, gstool.Error(`当前记忆整理模型不可用`)
	}
	if strings.ToLower(cast.ToString(modelInfo[`model_type`])) != `llm` {
		return 0, ``, gstool.Error(`记忆整理仅支持 LLM 模型`)
	}
	prompt, err := common.DbMain.MemoryConfigValue(define.MemoryConfigArrangePrompt)
	if err != nil && !common.DbRowMissing(err) {
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

// allowedImageExts 记忆库图片上传允许的文件扩展名。
var allowedImageExts = map[string]bool{
	`.png`: true, `.jpg`: true, `.jpeg`: true, `.gif`: true, `.webp`: true, `.bmp`: true, `.svg`: true,
}

// MemoryFragmentImageUpload 上传图片到记忆库 images 目录。
func MemoryFragmentImageUpload(c *gin.Context) {
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	file, err := c.FormFile(`file`)
	if err != nil {
		gsgin.GinResponseError(c, `上传失败:`+err.Error(), nil)
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		gsgin.GinResponseError(c, `不支持的图片格式: `+ext, nil)
		return
	}
	imageDir := filepath.Join(component.MemoryRuntime.Config().Dir, `images`)
	_ = gstool.DirCreatePath(imageDir)
	newName := fmt.Sprintf(`%d%s`, time.Now().UnixMicro(), ext)
	dst := filepath.Join(imageDir, newName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		gsgin.GinResponseError(c, `保存图片失败:`+err.Error(), nil)
		return
	}
	urlPath := `/memory/images/` + newName
	gsgin.GinResponseSuccess(c, ``, map[string]string{
		`url`: urlPath,
	})
}

// MemoryFragmentImageServe 提供记忆库图片的静态文件服务。
func MemoryFragmentImageServe(c *gin.Context) {
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		c.Status(404)
		return
	}
	imageName := strings.TrimSpace(c.Param(`name`))
	if imageName == `` || strings.ContainsAny(imageName, `/\`) {
		c.Status(404)
		return
	}
	imagePath := filepath.Join(component.MemoryRuntime.Config().Dir, `images`, imageName)
	if _, err := os.Stat(imagePath); err != nil {
		c.Status(404)
		return
	}
	c.File(imagePath)
}

// MemoryFragmentUploadZip 上传 ZIP 文件，解析 content.md + images/，创建知识片段。
func MemoryFragmentUploadZip(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	file, err := c.FormFile(`file`)
	if err != nil {
		gsgin.GinResponseError(c, `上传失败:`+err.Error(), nil)
		return
	}
	if !strings.HasSuffix(strings.ToLower(file.Filename), `.zip`) {
		gsgin.GinResponseError(c, `仅支持 .zip 文件`, nil)
		return
	}

	// 保存到临时文件
	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, fmt.Sprintf(`fragment_upload_%d.zip`, time.Now().UnixMicro()))
	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		gsgin.GinResponseError(c, `保存临时文件失败:`+err.Error(), nil)
		return
	}
	defer os.Remove(tmpPath)

	// 解压 ZIP
	reader, err := zip.OpenReader(tmpPath)
	if err != nil {
		gsgin.GinResponseError(c, `打开 ZIP 文件失败:`+err.Error(), nil)
		return
	}
	defer reader.Close()

	// 读取 content.md
	var markdownContent string
	for _, f := range reader.File {
		if f.Name == `content.md` {
			rc, openErr := f.Open()
			if openErr != nil {
				gsgin.GinResponseError(c, `打开 content.md 失败:`+openErr.Error(), nil)
				return
			}
			content, readErr := io.ReadAll(rc)
			_ = rc.Close()
			if readErr != nil {
				gsgin.GinResponseError(c, `读取 content.md 失败:`+readErr.Error(), nil)
				return
			}
			markdownContent = string(content)
			break
		}
	}
	if markdownContent == `` {
		gsgin.GinResponseError(c, `ZIP 中未找到 content.md`, nil)
		return
	}

	// 从 content.md 提取标题（第一个 # 行）
	title := `导入的知识片段`
	lines := strings.Split(markdownContent, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, `# `) {
			title = strings.TrimPrefix(trimmed, `# `)
			break
		}
	}

	// 保存图片并重写路径
	memoryDir := component.MemoryRuntime.Config().Dir
	pathMapping, imgErr := saveScrapeImagesToMemoryDir(&reader.Reader, memoryDir)
	if imgErr != nil {
		gsgin.GinResponseError(c, `保存图片失败:`+imgErr.Error(), nil)
		return
	}
	markdownContent = rewriteScrapeImagePaths(markdownContent, pathMapping)
	markdownContent = prefixMemoryImagePaths(markdownContent)

	// 创建知识片段
	info, saveErr := memoryDB.MemoryFragmentSave(``, title, markdownContent, nil)
	if saveErr != nil {
		gsgin.GinResponseError(c, `创建片段失败:`+saveErr.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	broadcastMemoryFragmentUpsert(info)
	gsgin.GinResponseSuccess(c, ``, info)
}
