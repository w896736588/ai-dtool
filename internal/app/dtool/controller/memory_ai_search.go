package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

const (
	aiSearchClientIDPrefix = `memory_ai_search_`
	aiSearchBatchSize      = 50
	aiSearchMaxContentLen  = 3000
	aiSearchMaxTotalLen    = 30000
)

// aiSearchEvent 定义 SSE 搜索步骤事件。
type aiSearchEvent struct {
	Step         string `json:"step"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	Data         any    `json:"data,omitempty"`
	DurationMs   int64  `json:"duration_ms,omitempty"`
	InputTokens  int    `json:"input_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	Response     string `json:"response,omitempty"`
}

// MemoryFragmentAiSearch 是 AI 智能搜索的 SSE 端点处理器。
func MemoryFragmentAiSearch(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
	query := strings.TrimSpace(urlValues.Get(`query`))
	if query == `` {
		return nil, fmt.Errorf(`搜索关键词不能为空`)
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return nil, err
	}
	memoryDB := component.MemoryRuntime.DB()

	// 读取 AI 搜索模型配置
	modelIDText, err := common.DbMain.MemoryConfigValue(define.MemoryConfigAiSearchModelID)
	if err != nil && !common.DbRowMissing(err) {
		return nil, err
	}
	modelID := cast.ToInt(modelIDText)
	if modelID <= 0 {
		return nil, fmt.Errorf(`请先在知识片段设置中配置 AI 智能搜索模型`)
	}
	modelInfo, err := common.DbMain.AiModelInfo(modelID)
	if err != nil {
		return nil, fmt.Errorf(`AI 搜索模型不可用`)
	}
	if strings.ToLower(cast.ToString(modelInfo[`model_type`])) != `llm` {
		return nil, fmt.Errorf(`AI 搜索仅支持 LLM 模型`)
	}

	// 注册 SSE 客户端
	clientID := aiSearchClientIDPrefix + cast.ToString(cast.ToInt(urlValues.Get(`t`)))
	sse := gsgin.SseRegister(clientID, stopC, c)

	// 后台执行搜索流程
	go runAiSearchFlow(query, memoryDB, modelID, sse, stopC)

	return sse, nil
}

// aiSearchStopped 检查客户端是否已断开连接。
func aiSearchStopped(stopC chan int) bool {
	select {
	case <-stopC:
		return true
	default:
		return false
	}
}

// runAiSearchFlow 执行多步骤 AI 搜索流程。
func runAiSearchFlow(query string, memoryDB common.MemoryFragmentStore, modelID int, sse *gsgin.Sse, stopC chan int) {
	defer func() {
		defer func() { recover() }()
		close(stopC)
	}()

	// ---- 步骤1: AI 生成扩展关键词 ----
	sendSearchEvent(sse, `keywords`, `running`, `正在扩展搜索关键词...`, nil)
	keywordsText, _, kwUsage, kwDuration, err := common.DbMain.AIChatByModelWithUsage(modelID, aiSearchKeywordSystemPrompt, buildKeywordGenPrompt(query))
	if err != nil {
		sendSearchEvent(sse, `error`, `error`, `关键词生成失败: `+err.Error(), nil)
		return
	}
	if aiSearchStopped(stopC) {
		return
	}
	keywords := parseKeywordsResponse(keywordsText)
	if len(keywords) == 0 {
		keywords = []string{query}
	}
	sendSearchDoneEvent(sse, `keywords`, ``, map[string]any{`keywords`: keywords}, buildKeywordGenPrompt(query), keywordsText, kwDuration, kwUsage)

	// ---- 步骤2: rg 搜索片段（OR 逻辑）----
	sendSearchEvent(sse, `search`, `running`, `正在搜索知识片段...`, nil)
	searchStart := time.Now()
	results, err := memoryDB.SearchFragmentsOr(keywords, 200)
	searchDuration := time.Since(searchStart).Milliseconds()
	if err != nil {
		sendSearchEvent(sse, `error`, `error`, `搜索失败: `+err.Error(), nil)
		return
	}
	// 构建片段摘要列表供前端展示
	searchFragments := make([]map[string]any, 0, len(results))
	for _, item := range results {
		searchFragments = append(searchFragments, map[string]any{
			`id`:    item[`id`],
			`title`: item[`title`],
		})
	}
	sendSearchDoneEvent(sse, `search`, ``, map[string]any{
		`total`:     len(results),
		`fragments`: searchFragments,
	}, strings.Join(keywords, `, `), ``, searchDuration, nil)

	if len(results) == 0 {
		// 无结果时让 AI 流式回答
		sendSearchEvent(sse, `answer`, `running`, `未找到相关片段，正在基于 AI 知识回答...`, nil)
		_, _, _ = common.DbMain.AIChatStreamByModel(modelID, aiSearchSynthesizeSystemPrompt, buildNoResultPrompt(query), func(chunk string) {
			if aiSearchStopped(stopC) {
				return
			}
			if chunk != `` {
				sendSearchEvent(sse, `answer`, `streaming`, ``, chunk)
			}
		})
		sendSearchEvent(sse, `done`, `done`, ``, map[string]any{`referenced_fragments`: []any{}})
		return
	}

	// ---- 步骤3: AI 评估标题，选择需要读取的片段 ----
	sendSearchEvent(sse, `judge`, `running`, `正在评估片段相关性...`, nil)
	allSelected := make([]map[string]any, 0)
	totalPages := (len(results) + aiSearchBatchSize - 1) / aiSearchBatchSize
	var judgeTotalDuration int64
	var judgeTotalUsage *common.AiChatUsage
	judgePromptBuilder := strings.Builder{}
	judgeResponseBuilder := strings.Builder{}

	for page := 0; page*aiSearchBatchSize < len(results); page++ {
		if aiSearchStopped(stopC) {
			return
		}
		end := (page + 1) * aiSearchBatchSize
		if end > len(results) {
			end = len(results)
		}
		batch := results[page*aiSearchBatchSize : end]
		titles := make([]string, 0, len(batch))
		for _, item := range batch {
			titles = append(titles, cast.ToString(item[`title`]))
		}

		judgePrompt := buildTitleJudgePrompt(query, titles)
		judgePromptBuilder.WriteString(judgePrompt)
		judgePromptBuilder.WriteString("\n---\n")
		judgeText, _, judgeUsage, judgeDuration, err := common.DbMain.AIChatByModelWithUsage(modelID, aiSearchJudgeSystemPrompt, judgePrompt)
		judgeTotalDuration += judgeDuration
		if judgeUsage != nil {
			if judgeTotalUsage == nil {
				judgeTotalUsage = &common.AiChatUsage{}
			}
			judgeTotalUsage.InputTokens += judgeUsage.InputTokens
			judgeTotalUsage.OutputTokens += judgeUsage.OutputTokens
		}
		if err != nil {
			sendSearchEvent(sse, `error`, `error`, `标题评估失败: `+err.Error(), nil)
			return
		}
		if aiSearchStopped(stopC) {
			return
		}
		judgeResponseBuilder.WriteString(judgeText)
		judgeResponseBuilder.WriteString("\n")
		selectedIndices := parseTitleJudgeResponse(judgeText, len(batch))
		for _, idx := range selectedIndices {
			if idx >= 0 && idx < len(batch) {
				allSelected = append(allSelected, batch[idx])
			}
		}

		// 如果还有下一页，询问 AI 是否继续
		if end < len(results) {
			remaining := len(results) - end
			continuePrompt := buildContinuePrompt(query, remaining)
			continueText, _, contUsage, contDuration, contErr := common.DbMain.AIChatByModelWithUsage(modelID, aiSearchJudgeSystemPrompt, continuePrompt)
			judgeTotalDuration += contDuration
			if contUsage != nil {
				if judgeTotalUsage == nil {
					judgeTotalUsage = &common.AiChatUsage{}
				}
				judgeTotalUsage.InputTokens += contUsage.InputTokens
				judgeTotalUsage.OutputTokens += contUsage.OutputTokens
			}
			if contErr != nil || !parseContinueResponse(continueText) {
				break
			}
		}
	}
	// 选中片段摘要列表
	selectedFragments := make([]map[string]any, 0, len(allSelected))
	for _, item := range allSelected {
		selectedFragments = append(selectedFragments, map[string]any{
			`id`:    item[`id`],
			`title`: item[`title`],
		})
	}
	sendSearchDoneEvent(sse, `judge`, ``, map[string]any{
		`selected_count`:     len(allSelected),
		`total`:              len(results),
		`total_pages`:        totalPages,
		`selected_fragments`: selectedFragments,
	}, judgePromptBuilder.String(), judgeResponseBuilder.String(), judgeTotalDuration, judgeTotalUsage)

	if aiSearchStopped(stopC) {
		return
	}

	// ---- 步骤4: 读取选中片段内容 ----
	sendSearchEvent(sse, `read`, `running`, `正在读取片段内容...`, nil)
	readStart := time.Now()
	collectedContent := strings.Builder{}
	totalWritten := 0
	readFragmentList := make([]map[string]any, 0, len(allSelected))
	for i, item := range allSelected {
		if aiSearchStopped(stopC) {
			return
		}
		filePath := cast.ToString(item[`file_path`])
		title := cast.ToString(item[`title`])
		if filePath == `` {
			continue
		}
		content, err := memoryDB.ReadFragmentContent(filePath)
		if err != nil {
			continue
		}
		// 截断单片段内容
		if len(content) > aiSearchMaxContentLen {
			content = content[:aiSearchMaxContentLen] + `...`
		}
		// 检查总量
		if totalWritten+len(content) > aiSearchMaxTotalLen {
			remaining := aiSearchMaxTotalLen - totalWritten
			if remaining > 0 {
				content = content[:remaining] + `...`
			} else {
				break
			}
		}
		collectedContent.WriteString(fmt.Sprintf("## %s\n\n%s\n\n---\n\n", title, content))
		totalWritten += len(content)
		readFragmentList = append(readFragmentList, map[string]any{
			`id`:    item[`id`],
			`title`: title,
		})
		sendSearchEvent(sse, `read`, `running`, ``, map[string]any{
			`current`: i + 1,
			`total`:   len(allSelected),
			`title`:   title,
		})
	}
	readDuration := time.Since(readStart).Milliseconds()
	sendSearchDoneEvent(sse, `read`, ``, map[string]any{
		`read_fragments`: readFragmentList,
		`total_chars`:    totalWritten,
	}, fmt.Sprintf(`共读取 %d 个片段`, len(readFragmentList)), ``, readDuration, nil)

	if aiSearchStopped(stopC) {
		return
	}

	// ---- 步骤5: 流式生成综合回答 ----
	sendSearchEvent(sse, `answer`, `running`, `正在生成回答...`, nil)
	answerStart := time.Now()
	synthesizePrompt := buildSynthesizePrompt(query, collectedContent.String())
	_, _, err = common.DbMain.AIChatStreamByModel(modelID, aiSearchSynthesizeSystemPrompt, synthesizePrompt, func(chunk string) {
		if aiSearchStopped(stopC) {
			return
		}
		if chunk != `` {
			sendSearchEvent(sse, `answer`, `streaming`, ``, chunk)
		}
	})
	answerDuration := time.Since(answerStart).Milliseconds()
	if aiSearchStopped(stopC) {
		return
	}
	sendSearchDoneEvent(sse, `answer`, `回答生成完成`, nil, query, ``, answerDuration, nil)
	if err != nil {
		sendSearchEvent(sse, `error`, `error`, `回答生成失败: `+err.Error(), nil)
		return
	}

	// ---- 完成 ----
	referencedFragments := make([]map[string]any, 0, len(allSelected))
	for _, item := range allSelected {
		referencedFragments = append(referencedFragments, map[string]any{
			`id`:    item[`id`],
			`title`: item[`title`],
		})
	}
	sendSearchEvent(sse, `done`, `done`, ``, map[string]any{
		`referenced_fragments`: referencedFragments,
	})
}

// sendSearchEvent 通过 SSE 发送一个搜索步骤事件。
func sendSearchEvent(sse *gsgin.Sse, step, status, message string, data any) {
	event := aiSearchEvent{
		Step:    step,
		Status:  status,
		Message: message,
		Data:    data,
	}
	_ = sse.SendToChan(gstool.JsonEncode(event))
}

// sendSearchDoneEvent 发送步骤完成事件，附带耗时、token、完整提示词和回复。
func sendSearchDoneEvent(sse *gsgin.Sse, step, message string, data any, prompt, response string, durationMs int64, usage *common.AiChatUsage) {
	event := aiSearchEvent{
		Step:       step,
		Status:     `done`,
		Message:    message,
		Data:       data,
		Prompt:     prompt,
		Response:   response,
		DurationMs: durationMs,
	}
	if usage != nil {
		event.InputTokens = usage.InputTokens
		event.OutputTokens = usage.OutputTokens
	}
	_ = sse.SendToChan(gstool.JsonEncode(event))
}

// ---- 提示词 ----

const aiSearchKeywordSystemPrompt = `你是一个搜索关键词扩展助手。用户会给你一个问题，你需要根据这个问题联想出更多相关的搜索关键词，帮助在知识库中找到相关内容。
规则：
1. 返回关键词列表，每行一个
2. 包含原始问题中的核心词
3. 添加同义词、相关术语、可能的相关主题
4. 关键词应该简洁（1-4个字）
5. 不要返回编号，只返回关键词
6. 最多返回 15 个关键词
7. 直接返回关键词列表，不要任何解释`

const aiSearchJudgeSystemPrompt = `你是一个搜索结果评估助手。用户会给你一个问题和一组片段标题，你需要判断哪些标题与问题相关。

规则：
1. 仔细分析问题意图
2. 返回你认为与问题相关的标题索引号（从0开始）
3. 只返回索引号，用逗号分隔，例如: 0,2,5,8
4. 如果没有相关的，返回 NONE
5. 宁可多选也不要遗漏
6. 直接返回结果，不要任何解释`

const aiSearchSynthesizeSystemPrompt = `你是一个知识库助手。用户会给你一个问题和从知识库中检索到的相关片段内容。请你：
1. 基于检索到的片段内容回答用户的问题
2. 如果片段内容不足以完全回答问题，可以适当补充你的知识，但要明确说明哪些是知识库中的信息，哪些是你补充的
3. 使用 Markdown 格式组织回答，使其清晰易读
4. 引用片段时要标注来源标题
5. 回答要全面、准确、有条理
6. 使用中文回答`

func buildKeywordGenPrompt(query string) string {
	return `用户的问题: ` + query + `\n\n请生成搜索关键词:`
}

func buildTitleJudgePrompt(query string, titles []string) string {
	var sb strings.Builder
	sb.WriteString(`用户的问题: `)
	sb.WriteString(query)
	sb.WriteString(`\n\n片段标题列表:\n`)
	for i, title := range titles {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i, title))
	}
	sb.WriteString(`\n请返回与问题相关的标题索引号:`)
	return sb.String()
}

func buildContinuePrompt(query string, remainingCount int) string {
	return fmt.Sprintf(`用户的问题: %s\n\n当前已评估了一批片段标题，还有 %d 个片段标题未评估。\n\n请判断是否需要继续评估: 如果当前问题需要更全面的信息，回复 YES；如果当前信息已经足够，回复 NO。\n\n直接回复 YES 或 NO:`, query, remainingCount)
}

func buildSynthesizePrompt(query, collectedContent string) string {
	return fmt.Sprintf(`用户的问题:\n%s\n\n检索到的知识片段内容:\n\n%s\n\n请基于以上内容回答用户的问题:`, query, collectedContent)
}

func buildNoResultPrompt(query string) string {
	return fmt.Sprintf(`用户在知识库中搜索了以下问题但没有找到相关片段:\n%s\n\n请根据你的知识给出有帮助的回答，并说明知识库中暂无相关记录，建议用户可以先创建相关片段。`, query)
}

// ---- 响应解析 ----

var (
	keywordLineRegex = regexp.MustCompile(`^\d+[\.\)、]\s*(.+)$`)
	indexListRegex   = regexp.MustCompile(`\d+`)
)

// parseKeywordsResponse 解析 AI 返回的关键词列表。
func parseKeywordsResponse(text string) []string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	keywords := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == `` {
			continue
		}
		// 尝试去掉编号前缀 "1. xxx" 或 "1) xxx" 或 "1、xxx"
		if m := keywordLineRegex.FindStringSubmatch(line); len(m) > 1 {
			line = strings.TrimSpace(m[1])
		}
		// 去掉可能的引号包裹
		line = strings.Trim(line, `"'`)
		if line != `` && len(line) <= 20 {
			keywords = append(keywords, line)
		}
	}
	return keywords
}

// parseTitleJudgeResponse 解析 AI 返回的标题索引列表。
func parseTitleJudgeResponse(text string, maxIndex int) []int {
	text = strings.ToUpper(strings.TrimSpace(text))
	if text == `NONE` || text == `` {
		return nil
	}
	matches := indexListRegex.FindAllString(text, -1)
	result := make([]int, 0, len(matches))
	for _, m := range matches {
		idx := cast.ToInt(m)
		if idx >= 0 && idx < maxIndex {
			result = append(result, idx)
		}
	}
	return result
}

// parseContinueResponse 解析 AI 返回的是否继续决策。
func parseContinueResponse(text string) bool {
	text = strings.ToUpper(strings.TrimSpace(text))
	return strings.Contains(text, `YES`)
}
