package controller

import (
	"dev_tool/internal/app/dtool/common"
	"errors"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// smartLinkLocatorAutoExtractResult 描述 AI 自动提取后的 locator 字符串。
// smartLinkLocatorAutoExtractResult stores the AI-generated locator string.
type smartLinkLocatorAutoExtractResult struct {
	Value string `json:"value"`
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

你需要根据“网页源码”和“目标元素描述”，生成一个可直接传给 Playwright locator(...) 的匹配字符串。

这里的返回值必须是 selector 字符串，本次只允许返回：
- CSS 选择器
- XPath 表达式

请遵守这些规则：
1. 只返回一个字符串，不要返回 JSON
2. 不要输出 Markdown，不要输出解释，不要输出代码块
3. 返回值必须能直接放进 Playwright locator(...) 里使用
4. 可以返回 CSS，也可以返回 XPath
5. 优先选择最稳定、最简洁、最不容易误匹配的表达式
6. 如果无法可靠定位，也必须只返回一个字符串，此时返回 NOT_FOUND

正确示例：
.login-form .submit-btn
//button[normalize-space()="登录"]

错误示例：
{"value": ".submit-btn"}
json code block:
".submit-btn"
end code block`)
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
	result := &smartLinkLocatorAutoExtractResult{Value: strings.TrimSpace(content)}
	if err := validateSmartLinkLocatorAutoExtractResult(result); err != nil {
		return nil, err
	}
	return result, nil
}

func validateSmartLinkLocatorAutoExtractResult(result *smartLinkLocatorAutoExtractResult) error {
	if result == nil {
		return errors.New(`结果不能为空`)
	}
	result.Value = strings.TrimSpace(result.Value)
	if result.Value == `` {
		return errors.New(`value 不能为空`)
	}
	return nil
}
