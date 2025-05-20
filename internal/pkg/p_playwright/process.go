package p_playwright

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"net/url"
	"strings"
	"time"
)

type Process struct {
	DomainLimit    string                                           //限制域名执行
	ProcessType    define.ProcessType                               //类型
	Locators       string                                           //元素选择
	Tip            string                                           //输出提示
	Checks         string                                           //检查判断 是否执行
	OutKey         string                                           //输出的判断
	Value          string                                           //值
	Domain         string                                           //域名
	Locator        *Locator                                         //元素解析
	ElementOp      *_struct.ElementOp                               //操作结构
	Page           *playwright.Page                                 //页面
	TakeContentMap map[string]string                                //提取
	BoolResultMap  map[string]bool                                  //结果判断
	Check          *Check                                           //判断
	RunCallFunc    func(define.ProcessType, string, string, string) //注册输出回调
	log            *gstool.GsSlog
	runParams      *_struct.PlaywrightRunParams
}

func NewProcess(process map[string]any, page *playwright.Page, runParams *_struct.PlaywrightRunParams,
	boolResultMap map[string]bool, takeContentMap map[string]string, log *gstool.GsSlog) *Process {
	p := &Process{
		DomainLimit:    cast.ToString(process[`domain_limit`]),
		ProcessType:    define.ProcessType(cast.ToString(process[`type`])),
		Locators:       cast.ToString(process[`Locator`]),
		Tip:            cast.ToString(process[`tip`]),
		Checks:         base.Component.TPlaywright.ValueFormat(cast.ToString(process[`check_key`]), runParams),
		OutKey:         cast.ToString(process[`out_key`]),
		Value:          base.Component.TPlaywright.ValueFormat(cast.ToString(process[`value`]), runParams),
		RunCallFunc:    runParams.RunCallFunc,
		Domain:         runParams.Domain,
		ElementOp:      &_struct.ElementOp{},
		BoolResultMap:  boolResultMap,
		TakeContentMap: takeContentMap,
		runParams:      runParams,
		Page:           page,
		log:            log,
	}
	p.Check = NewCheck(p.OutKey, p.Checks, p.BoolResultMap, p.log)
	p.Locator = NewLocator(p.Locators, page, p.ElementOp, p.log) //元素解析
	return p
}

func (h *Process) Do() (define.ProcessCode, string, error) {
	code, reason, err := h.PDomain()
	if err != nil || code == define.ProcessBreak || code == define.ProcessContinue {
		return code, reason, err
	}
	code, reason, err = h.PChecks()
	h.log.Debugf(`判断 tip %s checks %s %s %s %v`, h.Tip, h.Checks, code, reason, err)
	if err != nil || code == define.ProcessBreak || code == define.ProcessContinue {
		return code, reason, err
	}
	h.log.Debugf(`tip %s checks %s 允许执行`, h.Tip, h.Checks)
	switch h.ProcessType {
	case define.TextContent: //提取内容
		return h.PTextContent()
	case define.BoolResult: //bool结果判断
		return h.PBoolResult()
	case define.Close:
		return h.PClose()
	case define.Wait:
		return h.PWait()
	case define.WaitClose:
		return h.PWaitClose()
	case define.Click: //点击
		return h.PClick()
	case define.Input: //输入
		return h.PInput()
	case define.RedirectUri: //跳转 保持当前域名
		return h.PRedirect()
	case define.WaitUrl:
		return h.PWaitUrl()
	case define.CanvasImage:
		return h.CanvasImage()
	case define.ExistWait:
		return h.ExistWait()
	case define.NoExistWait:
		return h.NoExistWait()
	default:
		return define.ProcessBreak, `不支持的类型`, gstool.Error(`不支持的类型%s`, h.ProcessType)
	}
}

func (h *Process) CanvasImage() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	h.ElementOp.Type = define.ElementExist
	element, elementErr := h.Locator.Do(0)
	if elementErr != nil {
		h.callRun(elementErr.Error(), h.Locators)
	} else {
		base64Data, err := element.Evaluate(`canvas => {
		  return canvas.toDataURL('image/png'); // 导出为 PNG 格式的 Base64 字符串
		}`, nil)
		if err != nil {
			h.log.Debugf("提取canvas为图片失败 %v", err)
		} else {
			// 提取 Base64 部分（去掉前缀 "data:image/png;base64,"）
			base64Str := strings.Split(base64Data.(string), ",")[1]
			h.callRun(`获取二维码成功`, fmt.Sprintf(`<img src='data:image/png;base64,%s' />`, base64Str))
		}
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) ExistWait() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	h.ElementOp.Type = define.ElementExist
	paramList := strings.Split(h.Value, `|`)
	if len(paramList) != 2 {
		return define.ProcessBreak, ``, gstool.Error(`exist_wait类型value格式错误`)
	}
	for i := 0; i < cast.ToInt(paramList[1]); i++ {
		element, elementErr := h.Locator.Do(cast.ToInt(paramList[0]))
		if elementErr != nil || element == nil {
			h.callRun(fmt.Sprintf(`等待中(%d/%d)..`, i+1, cast.ToInt(paramList[1])), h.Locators)
		} else {
			if h.OutKey != `` {
				h.BoolResultMap[h.OutKey] = true
			}
			return define.ProcessOk, ``, nil
		}
	}
	if h.OutKey != `` {
		h.BoolResultMap[h.OutKey] = false //不存在
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) NoExistWait() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	h.ElementOp.Type = define.ElementExist
	paramList := strings.Split(h.Value, `|`)
	if len(paramList) != 2 {
		return define.ProcessBreak, ``, gstool.Error(`exist_wait类型value格式错误`)
	}
	for i := 0; i < cast.ToInt(paramList[1]); i++ {
		element, elementErr := h.Locator.Do(cast.ToInt(paramList[0]))
		if elementErr != nil || element == nil {
			if h.OutKey != `` {
				h.BoolResultMap[h.OutKey] = false
			}
			return define.ProcessOk, ``, nil
		} else {
			time.Sleep(time.Second * time.Duration(cast.ToInt(paramList[0])))
			h.callRun(fmt.Sprintf(`等待中(%d/%d)..`, i+1, cast.ToInt(paramList[1])), h.Locators)
		}
	}
	if h.OutKey != `` {
		h.BoolResultMap[h.OutKey] = true //最终都没有消失，说明没有达到目的
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) PTextContent() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	h.ElementOp.Type = define.ElementTextContent
	_, elementErr := h.Locator.Do(0)
	if elementErr != nil {
		h.callRun(elementErr.Error(), h.Locators)
		h.TakeContentMap[h.OutKey] = ``
	} else {
		h.TakeContentMap[h.OutKey] = h.ElementOp.TextContent
		h.callRun(``, h.ElementOp.TextContent)
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) PBoolResult() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	if h.Locators != `` {
		h.ElementOp.Type = define.ElementCount
		boolRet, boolErr := h.Locator.DoBoolResult(0)
		if boolErr != nil {
			return define.ProcessBreak, `没有找到任意的元素` + h.Locators, errors.New(`没有找到任意的元素` + h.Locators)
		} else {
			h.BoolResultMap[h.OutKey] = boolRet
		}
		h.log.Debugf(`判断 %s`, gstool.JsonEncode(h.BoolResultMap))
	} else {
		//根据上面的执行来判断
		h.Check.OutKeyBoolResult()
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) PClick() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	h.log.Debugf(`点击 %s 允许`, h.Tip)
	h.ElementOp.Type = define.ElementClick
	_, elementErr := h.Locator.Do(0)
	if elementErr != nil {
		h.callRun(elementErr.Error(), h.Locators)
		return define.ProcessBreak, `获取需要点击的元素失败`, gstool.Error(`获取元素%s失败`, h.Locators)
	} else {
		h.callRun(``, h.Locators)
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) PInput() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	h.ElementOp.Type = define.ElementInput
	h.ElementOp.FillValue = h.Value
	_, elementErr := h.Locator.Do(0)
	if elementErr != nil {
		h.callRun(elementErr.Error(), h.Locators)
		return define.ProcessBreak, `获取需要输入的元素失败`, gstool.Error(`获取元素%s失败`, h.Locators)
	}
	h.callRun(``, h.Value)
	return define.ProcessOk, ``, nil
}

func (h *Process) PWaitUrl() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	waitResponse := _struct.ProcessWaitUrl{}
	_ = gstool.JsonDecode(h.Value, &waitResponse)
	parseU, _ := url.Parse((*h.Page).URL())
	responseUrl := gstool.SReplaces(waitResponse.ResponseUrl, map[string]string{
		`{domain}`: parseU.Host,
		`{scheme}`: parseU.Scheme,
	})
	h.log.Debugf(`准备的 %s`, responseUrl)
	h.log.Debugf(`全部的 %s`, gstool.JsonEncode(h.runParams.ResponseUrls))
	for i := 0; i < waitResponse.WaitSecond; i++ {
		for _, v := range h.runParams.ResponseUrls {
			if v.Url == responseUrl {
				h.log.Debugf(`等待返回 %s 成功`, responseUrl)
				return define.ProcessOk, responseUrl, nil
			}
		}
		time.Sleep(time.Second)
	}
	return define.ProcessBreak, fmt.Sprintf(`等待%s超时`, waitResponse.ResponseUrl), gstool.Error(`等待%s超时`, waitResponse.ResponseUrl)
}

func (h *Process) PRedirect() (define.ProcessCode, string, error) {
	//尝试解析
	processRedirect := _struct.ProcessRedirect{}
	_ = gstool.JsonDecode(h.Value, &processRedirect)
	//走多条件
	if processRedirect.Url != `` {
		for _, v := range processRedirect.RegisterResponseUrl {
			go func() {
				parseU, _ := url.Parse((*h.Page).URL())
				responseUrl := gstool.SReplaces(v.Url, map[string]string{
					`{domain}`: parseU.Host,
					`{scheme}`: parseU.Scheme,
				})
				h.log.Debugf(`注册等待返回 %s 超时 %d`, responseUrl, v.WaitSecond)
				_, _ = (*h.Page).ExpectResponse(responseUrl, func() error {
					h.log.Debugf(`请求%s完成 `, responseUrl)
					h.runParams.ResponseUrls = append(h.runParams.ResponseUrls, &_struct.ProcessResponseUrl{
						Url: responseUrl,
					})
					return nil
				}, playwright.PageExpectResponseOptions{Timeout: playwright.Float(cast.ToFloat64(v.WaitSecond) * 1000)})
			}()

		}
		//链接
		redirectUri := processRedirect.Url
		base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
		currentURL := (*h.Page).URL()
		parsedURL, err := url.Parse(currentURL)
		if err != nil {
			h.callRun(fmt.Sprintf(`解析url%s失败%s`, currentURL, err.Error()), currentURL)
			return define.ProcessBreak, `解析域名失败`, gstool.Error(`解析url%s失败%s`, currentURL, err.Error())
		}
		domain := parsedURL.Scheme + `://` + parsedURL.Host
		targetUrl := domain + redirectUri
		time.Sleep(time.Second)
		if _, goErr := (*h.Page).Goto(targetUrl); goErr != nil {
			h.callRun(goErr.Error(), targetUrl)
			return define.ProcessBreak, `跳转失败`, goErr
		} else {
			h.callRun(``, targetUrl)
		}
		return define.ProcessOk, ``, nil
	} else {
		//链接
		redirectUri := h.Value
		base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
		currentURL := (*h.Page).URL()
		parsedURL, err := url.Parse(currentURL)
		if err != nil {
			h.callRun(fmt.Sprintf(`解析url%s失败%s`, currentURL, err.Error()), currentURL)
			return define.ProcessBreak, `解析域名失败`, gstool.Error(`解析url%s失败%s`, currentURL, err.Error())
		}
		domain := parsedURL.Scheme + `://` + parsedURL.Host
		targetUrl := domain + redirectUri
		time.Sleep(time.Second)
		if _, goErr := (*h.Page).Goto(targetUrl); goErr != nil {
			h.callRun(goErr.Error(), targetUrl)
			return define.ProcessBreak, `跳转失败`, goErr
		} else {
			h.callRun(``, targetUrl)
		}
		return define.ProcessOk, ``, nil
	}

}

func (h *Process) PWaitClose() (define.ProcessCode, string, error) {
	go func() {
		base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
		time.Sleep(time.Duration(cast.ToInt(h.Value)) * time.Second)
		_ = (*h.Page).Close()
	}()
	return define.ProcessOk, ``, nil
}

func (h *Process) PDomain() (define.ProcessCode, string, error) {
	if h.DomainLimit != `` && !strings.Contains(h.Domain, h.DomainLimit) {
		return define.ProcessContinue, `域名过滤`, nil
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) PChecks() (define.ProcessCode, string, error) {
	//不需要执行
	ignoreTypeList := []define.ProcessType{
		define.TextContent,
		define.BoolResult,
	}
	if !gstool.ArrayExistValue(&ignoreTypeList, h.ProcessType) && !h.Check.AllowCheckKey() {
		return define.ProcessContinue, `checks检查未通过`, nil
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) PClose() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	_ = (*h.Page).Close()
	return define.ProcessBreak, `页面关闭，结束`, nil
}

func (h *Process) PWait() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	time.Sleep(time.Duration(cast.ToInt(h.Value)) * time.Millisecond)
	return define.ProcessOk, ``, nil
}

func (h *Process) callRun(errmsg, content string) {
	if h.RunCallFunc != nil {
		h.RunCallFunc(h.ProcessType, errmsg, h.Tip, content)
	}
}
