package p_playwright

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"gitee.com/Sxiaobai/gs/gstask"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
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

func (h *Locator) Do() (playwright.Locator, error) {
	lists := strings.Split(h.Locators, `&&`) //多个用&&分割
	hList := strings.Split(h.Locators, `||`) //多个用||分割
	task := gstask.NewTask()
	waitSecond := playwright.Float(3000)
	//&&处理
	if len(lists) > 1 {
		h.log.Debugf(`走&& %s`, h.Locators)
		for _, locators := range lists {
			task.Add(gstask.CallbackFunc{
				Func: func() *gstask.Result {
					return h.FindLocator(locators, waitSecond)
				},
				Timeout: 5 * time.Second,
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

func (h *Locator) FindLocator(locators string, waitSecond *float64) *gstask.Result {
	locator := h.parseLocator(locators)
	selectorLoader := (*h.Page).Locator(locator.Locator)
	if locator.First { //首个
		selectorLoader = selectorLoader.First()
	}
	selectorLoaderWaitErr := selectorLoader.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: waitSecond,
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
			h.log.Debugf(`查找 %s 失败`, locator.Locator)
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
