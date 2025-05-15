package p_playwright

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"gitee.com/Sxiaobai/gs/gstask"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"strings"
	"time"
)

type Locator struct {
	Locators  string
	Page      *playwright.Page
	ElementOp *_struct.ElementOp
	log       *gstool.GsSlog
}

func NewLocator(locators string, page *playwright.Page, elementOp *_struct.ElementOp, log *gstool.GsSlog) *Locator {
	return &Locator{
		Locators:  locators,
		Page:      page,
		ElementOp: elementOp,
		log:       log,
	}
}

// DoBoolResult 根据json配置的元素来解析，根据设置的值进行返回
// 示例：[{"locator":".qrcode[style*='opacity: 0']","return":true},{"locator":".qrcode[style*='opacity: 0']","return":false},{"locator":".to-backstage-btn","return":false},{"locator":".new-to-backstage-btn","return":false}]
func (h *Locator) DoBoolResult(second int) (bool, error) {
	boolList := make([]_struct.ProcessBoolResult, 0)
	decodeErr := gstool.JsonDecode(h.Locators, &boolList)
	if decodeErr != nil {
		return false, errors.New(`不支持的bool_result表达式`)
	}
	task := gstask.NewTask()
	waitSecond := 3
	if second > 0 {
		waitSecond = second
	}
	for _, result := range boolList {
		call := func() *gstask.Result {
			findElemRet := h.FindLocator(result.Locator, waitSecond)
			if findElemRet.Err != nil { //查找这个元素失败 那么直接返回失败
				return &gstask.Result{
					Result: nil, //没找到 那么返回反向值
					Err:    errors.New(`没有找到元素`),
				}
			} else {
				return &gstask.Result{
					Result: result.ExistReturn, //找到了 那么返回值
					Err:    nil,
				}
			}
		}
		task.Add(gstask.CallbackFunc{
			Func:    call,
			Timeout: 5 * time.Second, //超时时间 不是于查找元素的超时时间
		})
	}
	result := task.RunOne()
	h.log.Debugf(`DoBoolResult 查找结果 %#v`, result)
	if result.Err != nil {
		h.log.Debugf(`处理：%s失败：%s`, h.Locators, result.Err.Error())
		return false, result.Err
	}
	return result.Result.(bool), result.Err
}

func (h *Locator) Do(second int) (playwright.Locator, error) {
	lists := strings.Split(h.Locators, `&&`) //多个用&&分割
	hList := strings.Split(h.Locators, `||`) //多个用||分割
	task := gstask.NewTask()
	waitSecond := 3
	if second > 0 {
		waitSecond = second
	}
	//&&处理
	if len(lists) > 1 {
		h.log.Debugf(`走&& %s`, h.Locators)
		for _, locators := range lists {
			task.Add(gstask.CallbackFunc{
				Func: func() *gstask.Result {
					return h.FindLocator(locators, waitSecond)
				},
				Timeout: 5 * time.Second, //超时时间 不是于查找元素的超时时间
			})
		}
	}
	//||处理
	if len(hList) > 1 {
		h.log.Debugf(`走|| %s`, h.Locators)
		for _, locators := range hList {
			task.Add(gstask.CallbackFunc{
				Func: func() *gstask.Result {
					return h.FindLocator(locators, waitSecond)
				},
				Timeout: 5 * time.Second,
			})
		}
	}

	//默认
	if !gstool.SContains(h.Locators, []string{`&&`, `||`}) {
		h.log.Debugf(`走默认 %s`, h.Locators)
		task.Add(gstask.CallbackFunc{
			Func: func() *gstask.Result {
				return h.FindLocator(h.Locators, waitSecond)
			},
			Timeout: 5 * time.Second,
		})
	}

	result := task.RunOne()
	h.log.Debugf(`查找结果 %#v`, result)
	if result.Err != nil {
		h.log.Debugf(`处理：%s失败：%s`, h.Locators, result.Err.Error())
		return nil, result.Err
	}
	element := result.Result.(playwright.Locator)
	switch h.ElementOp.Type {
	case define.ElementInput:
		fillErr := element.Fill(h.ElementOp.FillValue)
		return element, fillErr
	case define.ElementExist:
		return element, nil
	case define.ElementClick:
		clickErr := element.Click()
		return element, clickErr
	case define.ElementTextContent:
		content, textContentErr := element.TextContent()
		h.ElementOp.TextContent = strings.TrimSpace(content)
		return element, textContentErr
	case define.ElementCount:
		count, numErr := element.Count()
		h.ElementOp.Count = count
		return element, numErr
	default:
		return nil, errors.New(`不支持的操作`)
	}
}

func (h *Locator) parseLocator(Locator string) *_struct.Locator {
	sList := strings.Split(Locator, `|`)
	locator := _struct.Locator{
		Locator: sList[0],
		First:   false,
	}
	if gstool.ArrayExistValue(&sList, `first`) {
		locator.First = true
	}
	if strings.HasPrefix(locator.Locator, `!`) {
		locator.ExistSetNot = true
		locator.Locator = strings.TrimLeft(locator.Locator, `!`)
	}
	return &locator
}

func (h *Locator) FindLocator(locators string, waitSecond int) *gstask.Result {
	timeout := playwright.Float(cast.ToFloat64(waitSecond * 1000))
	h.log.Debugf(`查找元素%s 超时毫秒时间 %v`, locators, *timeout)
	locator := h.parseLocator(locators)
	selectorLoader := (*h.Page).Locator(locator.Locator)
	if locator.First { //首个
		selectorLoader = selectorLoader.First()
	}
	selectorLoaderWaitErr := selectorLoader.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: timeout,
	})
	//如果是反找Locator 不存在时返回正常
	if locator.ExistSetNot {
		if selectorLoaderWaitErr != nil {
			h.log.Debugf(`反查找 %s 失败`, locator.Locator)
			return &gstask.Result{
				Result: selectorLoader,
				Err:    nil,
			}
		} else {
			h.log.Debugf(`反查找 %s 成功`, locator.Locator)
			return &gstask.Result{
				Result: selectorLoader,
				Err:    errors.New(`找到了反找元素，返回失败`),
			}
		}
	} else {
		if selectorLoaderWaitErr != nil {
			h.log.Debugf(`查找 %s 失败 %s`, locator.Locator, selectorLoaderWaitErr.Error())
			return &gstask.Result{
				Result: nil,
				Err:    errors.New(`没有找到元素 ` + locator.Locator),
			}
		} else {
			h.log.Debugf(`查找 %s 成功`, locator.Locator)
			return &gstask.Result{
				Result: selectorLoader,
				Err:    nil,
			}
		}
	}
}
