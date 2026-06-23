package index

import (
	"dev_tool/internal/app/dtool/common"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/w896736588/go-tool/gstool"
)

// RetrieveResult 检索结果。
type RetrieveResult struct {
	Found      bool   // 是否命中已有脚本
	ScriptName string // 命中的脚本名称
	SkillName  string // 命中的 skill 名称
	Summary    string // 匹配摘要
}

// Retrieve 在索引中检索匹配用户任务的脚本。
// 使用 AI 判断 scripts.md 中是否有可复用的脚本。
// modelId 为 FC 模型 ID，userMessage 为用户任务描述。
func Retrieve(db *common.CSqlite, modelId int, indexPath, userMessage string) *RetrieveResult {
	if modelId <= 0 {
		return &RetrieveResult{Found: false}
	}
	// 读取 scripts.md
	scriptsContent := ReadIndexFile(indexPath, ScriptsFileName)
	if scriptsContent == `` {
		gstool.FmtPrintlnLogTime(`[butler-retrieve] scripts.md 为空，跳过检索`)
		return &RetrieveResult{Found: false}
	}
	// 使用 AI 判断是否有匹配的脚本
	prompt := buildRetrievePrompt(userMessage, scriptsContent)
	messages := []map[string]any{
		{`role`: `system`, `content`: retrieveSystemPrompt},
		{`role`: `user`, `content`: prompt},
	}
	content, _, _, _, err := db.AIChatByModelWithTools(modelId, messages, nil)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-retrieve] AI 检索失败 %s`, err.Error())
		return &RetrieveResult{Found: false}
	}
	return parseRetrieveResult(content)
}

// retrieveSystemPrompt 检索匹配的系统提示词。
const retrieveSystemPrompt = `你是一个脚本检索器。根据用户任务描述，判断现有脚本工具索引中是否有可以直接复用的脚本。

**优先级规则**：
- dtool-butler 节下的自进化脚本通常比模块通用脚本更具体、更贴合实际任务，应优先匹配
- 如果 dtool-butler 节和模块通用节都有匹配项，优先选择 dtool-butler 下的脚本
- 同一个任务可以有多个匹配，选择最具体、功能描述最接近的那个

如果找到匹配的脚本，请输出 JSON 格式：
{"found": true, "skill_name": "skill名称", "script_name": "脚本名称", "summary": "匹配原因简述"}

如果没有找到匹配的脚本，请输出：
{"found": false}

只输出 JSON，不要输出其他内容。`

// buildRetrievePrompt 构建检索的用户提示词。
func buildRetrievePrompt(userMessage, scriptsContent string) string {
	// 截断索引内容避免过长
	truncatedContent := scriptsContent
	if len(truncatedContent) > 3000 {
		truncatedContent = truncatedContent[:3000] + `\n...(内容已截断)`
	}
	return fmt.Sprintf(`用户任务：%s

现有脚本工具索引：
%s

请判断是否有可复用的脚本。`, userMessage, truncatedContent)
}

// parseRetrieveResult 解析 AI 返回的检索结果，容错提取。
func parseRetrieveResult(content string) *RetrieveResult {
	result := &RetrieveResult{Found: false}
	text := strings.TrimSpace(content)
	// 尝试提取 JSON
	jsonStart := strings.Index(text, `{`)
	jsonEnd := strings.LastIndex(text, `}`)
	if jsonStart < 0 || jsonEnd < 0 || jsonEnd <= jsonStart {
		return result
	}
	jsonStr := text[jsonStart : jsonEnd+1]
	var data map[string]any
	if err := jsonUnmarshal([]byte(jsonStr), &data); err != nil {
		gstool.FmtPrintlnLogTime(`[butler-retrieve] JSON 解析失败 %s`, err.Error())
		return result
	}
	if found, ok := data[`found`].(bool); ok && found {
		result.Found = true
		result.SkillName = toString(data[`skill_name`])
		result.ScriptName = toString(data[`script_name`])
		result.Summary = toString(data[`summary`])
	}
	return result
}

// jsonUnmarshal JSON 解析包装。
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// toString 安全地将 any 转为 string。
func toString(v any) string {
	if v == nil {
		return ``
	}
	return fmt.Sprintf(`%v`, v)
}
