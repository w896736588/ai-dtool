package plw

import (
	"dev_tool/internal/app/dtool/define"
	"errors"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/w896736588/go-tool/gstask"
	"github.com/w896736588/go-tool/gstool"
)

const (
	// boolResultDefaultWaitMills 表示 bool_result 未单独配置时的默认等待时长，单位毫秒。
	boolResultDefaultWaitMills = 3000

	// boolResultTaskTimeout 表示单条 bool_result 规则的最长执行时间。
	boolResultTaskTimeout = 5 * time.Second
)

// Locator 作为迁移期兼容层保留旧调用方式。
// 新的查询解析与动作执行统一委托给 LocatorService。
type Locator struct {
	Locators  string
	Input     *LocatorInput
	Page      *playwright.Page
	ElementOp *ElementOp
	log       *gstool.GsSlog
	service   *LocatorService
	parseErr  error
}

// NewLocator 创建 Locator 兼容层。
func NewLocator(locators string, input *LocatorInput, page *playwright.Page, elementOp *ElementOp, log *gstool.GsSlog, parseErr error) *Locator {
	return &Locator{
		Locators:  locators,
		Input:     input,
		Page:      page,
		ElementOp: elementOp,
		log:       log,
		service:   NewLocatorService(),
		parseErr:  parseErr,
	}
}

// DoBoolResult 根据 json 配置的多个元素条件返回对应布尔值。
// 当前仍兼容旧表达式格式，内部定位动作已统一走 LocatorService。
func (h *Locator) DoBoolResult(waitMills float64, streamFunc func(string)) (bool, error) {
	if h.parseErr != nil {
		h.emitBoolResultLog(streamFunc, fmt.Sprintf(`bool_result locator 解析失败: %s`, h.parseErr.Error()))
		return false, h.parseErr
	}
	boolList, err := decodeBoolResultRules(h.Locators)
	if err != nil {
		h.emitBoolResultLog(streamFunc, fmt.Sprintf(`bool_result 规则解析失败，原始 locator: %s`, h.Locators))
	}
	if err != nil {
		return false, errors.New(`不支持的bool_result表达式`)
	}
	task := gstask.NewTask()
	if waitMills == 0 {
		waitMills = boolResultDefaultWaitMills
	}
	h.emitBoolResultLog(streamFunc, fmt.Sprintf(`开始执行 bool_result，共 %d 条规则，等待 %dms`, len(boolList), int(waitMills)))
	for index, result := range boolList {
		current := result
		ruleIndex := index + 1
		task.Add(gstask.CallbackFunc{
			Func: func() *gstask.Result {
				h.emitBoolResultLog(streamFunc, fmt.Sprintf(`规则[%d] 开始检查，return=%t，locator=%s`, ruleIndex, current.ExistReturn, h.formatBoolResultLocator(current.Locator)))
				locatorInput, _, parseErr := h.service.ParseInputValue(current.Locator)
				if parseErr != nil {
					h.emitBoolResultLog(streamFunc, fmt.Sprintf(`规则[%d] locator 解析失败: %s`, ruleIndex, parseErr.Error()))
					return &gstask.Result{
						Result: nil,
						Err:    parseErr,
					}
				}
				h.emitBoolResultLog(streamFunc, fmt.Sprintf(`规则[%d] locator 解析成功，开始查找元素`, ruleIndex))
				findElemRet := h.findWithInput(locatorInput, waitMills)
				if findElemRet.Err != nil {
					h.emitBoolResultLog(streamFunc, fmt.Sprintf(`规则[%d] 未命中元素: %s`, ruleIndex, findElemRet.Err.Error()))
					return &gstask.Result{
						Result: nil,
						Err:    errors.New(`没有找到元素`),
					}
				}
				h.emitBoolResultLog(streamFunc, fmt.Sprintf(`规则[%d] 命中成功，返回 %t`, ruleIndex, current.ExistReturn))
				return &gstask.Result{
					Result: current.ExistReturn,
					Err:    nil,
				}
			},
			Timeout: boolResultTaskTimeout,
		})
	}
	result := task.RunOne()
	h.log.Debugf(`DoBoolResult 查找结果 %#v`, result)
	if result.Err != nil {
		h.log.Debugf(`处理：%s失败：%s`, h.Locators, result.Err.Error())
		h.emitBoolResultLog(streamFunc, fmt.Sprintf(`bool_result 执行失败: %s`, result.Err.Error()))
		return false, result.Err
	}
	h.emitBoolResultLog(streamFunc, fmt.Sprintf(`bool_result 执行成功，最终返回 %t`, result.Result.(bool)))
	return result.Result.(bool), result.Err
}

// decodeBoolResultRules 解析 bool_result 规则，仅支持结构化 locator。
func decodeBoolResultRules(raw string) ([]boolResultRule, error) {
	boolList := make([]boolResultRule, 0)
	if err := gstool.JsonDecode(raw, &boolList); err != nil {
		return nil, err
	}
	return boolList, nil
}

// boolResultRule 表示 bool_result 中的结构化 locator 规则。
type boolResultRule struct {
	// Locator 必须是结构化 locator 对象。
	Locator any `json:"locator"`

	// ExistReturn 表示命中当前规则时应该返回的布尔值。
	ExistReturn bool `json:"return"`
}

// Do 保留旧入口，内部统一走结构化 LocatorService。
func (h *Locator) Do(waitMills float64) (playwright.Locator, error) {
	if h.parseErr != nil {
		return nil, h.parseErr
	}
	if waitMills == 0 {
		waitMills = 3000
	}
	task := gstask.NewTask()
	task.Add(gstask.CallbackFunc{
		Func: func() *gstask.Result {
			return h.findWithInput(h.Input, waitMills)
		},
		Timeout: 5 * time.Second,
	})
	result := task.RunOne()
	h.log.Debugf(`查找结果 %#v`, result)
	if result.Err != nil {
		h.log.Debugf(`处理：%s失败：%s`, h.Locators, result.Err.Error())
		return nil, result.Err
	}
	return result.Result.(playwright.Locator), nil
}

// findWithInput 统一处理等待、动作执行和兼容回填。
func (h *Locator) findWithInput(input *LocatorInput, waitMills float64) *gstask.Result {
	if h.Page == nil || *h.Page == nil {
		return &gstask.Result{Result: nil, Err: errors.New(`page 不能为空`)}
	}
	action := h.buildAction()
	result, err := h.service.FindAndExecute(h.Page, input, action, waitMills)
	if err != nil {
		return &gstask.Result{Result: nil, Err: err}
	}
	h.fillLegacyResult(result)
	return &gstask.Result{
		Result: result.Locator,
		Err:    nil,
	}
}

// buildAction 根据旧 ElementOp 状态构建新动作定义。
func (h *Locator) buildAction() *ElementAction {
	if h.ElementOp == nil {
		return nil
	}
	switch h.ElementOp.Type {
	case define.ElementInput:
		return &ElementAction{
			Type:  define.ElementInput,
			Value: h.ElementOp.FillValue,
		}
	case define.ElementExist:
		return &ElementAction{Type: define.ElementExist}
	case define.ElementClick:
		return &ElementAction{Type: define.ElementClick}
	case define.ElementTextContent:
		return &ElementAction{Type: define.ElementTextContent}
	case define.ElementCount:
		return &ElementAction{Type: define.ElementCount}
	default:
		return nil
	}
}

// fillLegacyResult 将新执行结果回填到旧 ElementOp，保证迁移期逻辑兼容。
func (h *Locator) fillLegacyResult(result *ElementResult) {
	if h.ElementOp == nil || result == nil {
		return
	}
	h.ElementOp.TextContent = result.TextContent
	h.ElementOp.Count = result.Count
}

// emitBoolResultLog 统一处理 bool_result 的运行日志输出，避免回调为空时重复判断。
func (h *Locator) emitBoolResultLog(streamFunc func(string), message string) {
	if streamFunc == nil || message == `` {
		return
	}
	streamFunc(message)
}

// formatBoolResultLocator 将任意 locator 配置格式化成日志文本，便于排查真实执行值。
func (h *Locator) formatBoolResultLocator(raw any) string {
	if raw == nil {
		return ``
	}
	if text, ok := raw.(string); ok {
		return text
	}
	return gstool.JsonEncode(raw)
}
