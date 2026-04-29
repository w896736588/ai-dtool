package controller

import (
	"dev_tool/internal/app/dtool/common"
	"encoding/json"
	"errors"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// smartLinkLocatorAutoExtractResult 描述 AI 自动提取后的基础定位表单。
// smartLinkLocatorAutoExtractResult stores the AI-generated base locator form payload.
type smartLinkLocatorAutoExtractResult struct {
	LocatorEditorMode     string                               `json:"locator_editor_mode"`
	LocatorStructuredForm smartLinkLocatorStructuredFormResult `json:"locator_structured_form"`
	Reason                string                               `json:"reason"`
}

// smartLinkLocatorStructuredFormResult 描述简单模式下的基础定位表单字段。
// smartLinkLocatorStructuredFormResult describes the simple-mode locator form fields.
type smartLinkLocatorStructuredFormResult struct {
	Kind         string `json:"kind"`
	Value        string `json:"value"`
	TargetText   string `json:"target_text"`
	Exact        bool   `json:"exact"`
	Negate       bool   `json:"negate"`
	PickMode     string `json:"pick_mode"`
	Nth          int    `json:"nth"`
	TimeoutMills int    `json:"timeout_mills"`
}

// SmartLinkLocatorAutoExtract 调用 AI 自动生成基础定位配置。
// SmartLinkLocatorAutoExtract calls AI to generate base locator config automatically.
func SmartLinkLocatorAutoExtract(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	modelID := cast.ToInt(dataMap[`model_id`])
	htmlSource := strings.TrimSpace(cast.ToString(dataMap[`html_source`]))
	targetDesc := strings.TrimSpace(cast.ToString(dataMap[`target_desc`]))
	if modelID <= 0 {
		gsgin.GinResponseError(c, `请选择AI模型`, nil)
		return
	}
	if htmlSource == `` {
		gsgin.GinResponseError(c, `网页源码不能为空`, nil)
		return
	}
	if targetDesc == `` {
		gsgin.GinResponseError(c, `目标元素描述不能为空`, nil)
		return
	}
	rawContent, modelInfo, err := common.DbMain.AIChatByModel(
		modelID,
		smartLinkLocatorAutoExtractSystemPrompt(),
		buildSmartLinkLocatorAutoExtractUserPrompt(htmlSource, targetDesc),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), map[string]any{
			`system_prompt`: smartLinkLocatorAutoExtractSystemPrompt(),
		})
		return
	}
	result, parseErr := parseSmartLinkLocatorAutoExtractResult(rawContent)
	if parseErr != nil {
		gsgin.GinResponseSuccess(c, ``, map[string]any{
			`is_valid`:      false,
			`parse_error`:   parseErr.Error(),
			`system_prompt`: smartLinkLocatorAutoExtractSystemPrompt(),
			`raw_content`:   rawContent,
		})
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`is_valid`:        true,
		`system_prompt`:   smartLinkLocatorAutoExtractSystemPrompt(),
		`raw_content`:     rawContent,
		`locator_payload`: result,
		`model`:           cast.ToString(modelInfo[`model`]),
		`provider_name`:   cast.ToString(modelInfo[`provider_name`]),
		`model_id`:        modelID,
	})
}

func smartLinkLocatorAutoExtractSystemPrompt() string {
	return strings.TrimSpace(`你是一个 Playwright 网页元素定位配置生成器。

你需要根据“网页源码”和“目标元素描述”，生成一个可直接回填到基础 locator 表单的 JSON 对象。

这是一个 Playwright 定位系统，支持的定位类型只有：
- css
- button_text
- text
- label
- placeholder
- alt_text
- title
- test_id

这是一个 Playwright 结果提取系统，支持的 pick_mode 只有：
- none
- first
- last
- nth

它们的含义分别是：
- css: Playwright locator(selector)
- button_text: 用按钮文案定位，适合 button 或可点击按钮
- text: 用页面文本定位
- label: 用表单标签文本定位输入框
- placeholder: 用 placeholder 文本定位输入框
- alt_text: 用 alt 文本定位图片
- title: 用 title 属性定位元素
- test_id: 用测试标识定位元素
- none: 不额外指定 first/last/nth
- first: 只取第一个匹配元素
- last: 只取最后一个匹配元素
- nth: 取第 N 个匹配元素，N 从 0 开始

请遵守这些规则：
1. locator_editor_mode 只能返回 simple
2. kind 只能从支持的定位类型里选一种
3. pick_mode 只能从 none / first / last / nth 里选一种
4. 如果 pick_mode 不是 nth，则 nth 固定返回 0
5. 优先选择最稳定、最简洁、最不容易误匹配的方式
6. 如果元素是按钮，优先考虑 button_text
7. 如果是输入框，优先考虑 label 或 placeholder
8. 非必要不要使用 css
9. 不要输出 XPath
10. 不要输出 Markdown，不要输出解释，只能输出 JSON
11. 如果无法可靠定位，也必须输出 JSON，但 reason 要说明原因
12. timeout_mills 固定返回 3000

返回格式必须严格等于：
{
  "locator_editor_mode": "simple",
  "locator_structured_form": {
    "kind": "css|button_text|text|label|placeholder|alt_text|title|test_id",
    "value": "",
    "target_text": "",
    "exact": false,
    "negate": false,
    "pick_mode": "none|first|last|nth",
    "nth": 0,
    "timeout_mills": 3000
  },
  "reason": ""
}`)
}

func buildSmartLinkLocatorAutoExtractUserPrompt(htmlSource, targetDesc string) string {
	builder := strings.Builder{}
	builder.WriteString("目标元素描述：\n")
	builder.WriteString(strings.TrimSpace(targetDesc))
	builder.WriteString("\n\n网页源码：\n```html\n")
	builder.WriteString(htmlSource)
	builder.WriteString("\n```")
	return builder.String()
}

func parseSmartLinkLocatorAutoExtractResult(rawContent string) (*smartLinkLocatorAutoExtractResult, error) {
	content := stripMarkdownCodeFence(rawContent)
	result := &smartLinkLocatorAutoExtractResult{}
	if err := json.Unmarshal([]byte(content), result); err != nil {
		return nil, err
	}
	if err := validateSmartLinkLocatorAutoExtractResult(result); err != nil {
		return nil, err
	}
	return result, nil
}

func validateSmartLinkLocatorAutoExtractResult(result *smartLinkLocatorAutoExtractResult) error {
	if result == nil {
		return errors.New(`结果不能为空`)
	}
	if strings.TrimSpace(result.LocatorEditorMode) != `simple` {
		return errors.New(`locator_editor_mode 只能为 simple`)
	}
	form := result.LocatorStructuredForm
	kind := strings.TrimSpace(form.Kind)
	if !containsText(`|css|button_text|text|label|placeholder|alt_text|title|test_id|`, `|`+kind+`|`) {
		return errors.New(`kind 不在支持范围内`)
	}
	pickMode := strings.TrimSpace(form.PickMode)
	if pickMode == `` {
		pickMode = `none`
		result.LocatorStructuredForm.PickMode = pickMode
	}
	if !containsText(`|none|first|last|nth|`, `|`+pickMode+`|`) {
		return errors.New(`pick_mode 不在支持范围内`)
	}
	if kind == `button_text` {
		if strings.TrimSpace(form.Value) == `` && strings.TrimSpace(form.TargetText) == `` {
			return errors.New(`button_text 需要 value 或 target_text`)
		}
	} else if strings.TrimSpace(form.Value) == `` {
		return errors.New(`value 不能为空`)
	}
	if pickMode != `nth` {
		result.LocatorStructuredForm.Nth = 0
	}
	if result.LocatorStructuredForm.TimeoutMills <= 0 {
		result.LocatorStructuredForm.TimeoutMills = 3000
	}
	return nil
}

func containsText(source, target string) bool {
	return strings.Contains(source, target)
}
