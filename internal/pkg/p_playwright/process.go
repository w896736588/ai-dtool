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
	Name           string                                           //名称
	DomainLimit    string                                           //限制域名执行
	ProcessType    define.ProcessType                               //类型
	Locators       string                                           //元素选择
	Tip            string                                           //输出提示
	Checks         string                                           //检查判断 是否执行
	OutKey         string                                           //输出的判断
	Value          string                                           //值
	Domain         string                                           //域名
	WaitMills      float64                                          //等待时长
	Locator        *Locator                                         //元素解析
	ElementOp      *_struct.ElementOp                               //操作结构
	Page           *playwright.Page                                 //页面
	TakeContentMap map[string]string                                //提取
	BoolResultMap  map[string]bool                                  //结果判断
	Check          *Check                                           //判断
	IsAsync        int                                              //是否异步执行1异步
	RunCallFunc    func(define.ProcessType, string, string, string) //注册输出回调
	log            *gstool.GsSlog
	runParams      *_struct.PlaywrightRunParams
}

func NewProcess(process map[string]any, page *playwright.Page, runParams *_struct.PlaywrightRunParams,
	boolResultMap map[string]bool, takeContentMap map[string]string, log *gstool.GsSlog) *Process {
	p := &Process{
		Name:           cast.ToString(process[`name`]),
		DomainLimit:    cast.ToString(process[`domain_limit`]),
		ProcessType:    define.ProcessType(cast.ToString(process[`type`])),
		Locators:       cast.ToString(process[`locator`]),
		WaitMills:      cast.ToFloat64(process[`wait_mills`]),
		Tip:            cast.ToString(process[`tip`]),
		Checks:         base.Component.TPlaywright.ValueFormat(cast.ToString(process[`name`]), cast.ToString(process[`check_key`]), runParams),
		OutKey:         cast.ToString(process[`out_key`]),
		Value:          base.Component.TPlaywright.ValueFormat(cast.ToString(process[`name`]), cast.ToString(process[`value`]), runParams),
		RunCallFunc:    runParams.RunCallFunc,
		IsAsync:        cast.ToInt(process[`is_async`]),
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
	if err != nil || code == define.ProcessBreak || code == define.ProcessContinue {
		h.runParams.StreamFunc(h.Name, `不满足check_key条件`)
		return code, reason, err
	}
	switch h.ProcessType {
	case define.TextContent: //提取内容
		return h.PTextContent()
	case define.BoolResult:
		return h.PBoolResult()
	case define.BoolExist:
		return h.PBoolExist()
	case define.LoginUsernamePassword: //是否需要弹窗登录
		return h.PLoginUsernamePassword()
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
	element, elementErr := h.Locator.Do(h.WaitMills)
	if elementErr != nil {
		h.runParams.StreamFunc(h.Name, `提取扫码登录的二维码失败 `+elementErr.Error())
		h.callRun(elementErr.Error(), h.Locators)
	} else {
		base64Data, err := element.Evaluate(`canvas => {
		  return canvas.toDataURL('image/png'); // 导出为 PNG 格式的 Base64 字符串
		}`, nil)
		if err != nil {
			h.runParams.StreamFunc(h.Name, `提取扫码登录的二维码失败 `+err.Error())
			h.log.Debugf("提取canvas为图片失败 %v", err)
		} else {
			h.runParams.StreamFunc(h.Name, `提取二维码内容成功，请扫码登录`)
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
		h.runParams.StreamFunc(h.Name, `exist_wait类型value格式错误 `+h.Value)
		return define.ProcessBreak, ``, gstool.Error(`exist_wait类型value格式错误`)
	}
	for i := 0; i < cast.ToInt(paramList[1]); i++ {
		element, elementErr := h.Locator.Do(cast.ToFloat64(cast.ToInt(paramList[0]) * 1000))
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
		element, elementErr := h.Locator.Do(cast.ToFloat64(cast.ToInt(paramList[0]) * 1000))
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
	_, elementErr := h.Locator.Do(h.WaitMills)
	if elementErr != nil {
		h.callRun(elementErr.Error(), h.Locators)
		h.TakeContentMap[h.OutKey] = ``
		h.runParams.StreamFunc(h.Name, `未提取到内容`)
	} else {
		h.TakeContentMap[h.OutKey] = h.ElementOp.TextContent
		h.callRun(``, h.ElementOp.TextContent)
		h.runParams.StreamFunc(h.Name, `提取到内容:`+h.OutKey+`,`+h.ElementOp.TextContent)
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) PBoolResult() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	if h.Locators != `` {
		h.ElementOp.Type = define.ElementCount
		boolRet, boolErr := h.Locator.DoBoolResult(h.WaitMills)
		if boolErr != nil {
			h.runParams.StreamFunc(h.Name, `根据多个locators判断是否存在失败`)
			return define.ProcessBreak, `没有找到任意的元素` + h.Locators, errors.New(`没有找到任意的元素` + h.Locators)
		} else {
			h.BoolResultMap[h.OutKey] = boolRet
			h.runParams.StreamFunc(h.Name, `根据多个locators判断是否存在成功,`+h.OutKey+`,`+fmt.Sprintf(`%t`, boolRet))
		}
	} else {
		//根据上面的执行来判断
		h.Check.OutKeyBoolResult()
		h.runParams.StreamFunc(h.Name, `根据check_key判断是否存在成功,`+h.OutKey+`,`+fmt.Sprintf(`%t`, h.BoolResultMap[h.OutKey]))
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) PBoolExist() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	if h.Locators == `` {
		h.runParams.StreamFunc(h.Name, `locator为空，配置错误`)
		return define.ProcessBreak, `locators为空`, gstool.Error(`locators为空`)
	}
	h.ElementOp.Type = define.ElementCount
	result := h.Locator.FindLocator(h.Locators, h.WaitMills)
	if result.Err != nil || result.Result == nil {
		h.runParams.StreamFunc(h.Name, `未找到locator：`+h.Locators+` `+h.OutKey+`,设置为:false`)
		h.BoolResultMap[h.OutKey] = false
	} else {
		h.runParams.StreamFunc(h.Name, `找到locator：`+h.Locators+` `+h.OutKey+`,设置为:true`)
		h.BoolResultMap[h.OutKey] = true
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) PLoginUsernamePassword() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	h.runParams.StreamFunc(h.Name, `等待输入账号密码登录 30s后超时`)
	//根据上面的执行来判断
	h.callRun(``, ``)
	return define.ProcessOk, ``, nil
}

func (h *Process) PClick() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	h.ElementOp.Type = define.ElementClick
	_, elementErr := h.Locator.Do(h.WaitMills)
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
	h.Value = base.Component.TVariable.Replace(h.Value, h.runParams.ReplaceList)
	h.ElementOp.FillValue = h.Value
	_, elementErr := h.Locator.Do(h.WaitMills)
	if elementErr != nil {
		h.callRun(elementErr.Error(), h.Locators)
		h.runParams.StreamFunc(h.Name, `输入内容`+h.Value+`，失败,`+elementErr.Error())
		return define.ProcessBreak, `获取需要输入的元素失败`, gstool.Error(`获取元素%s失败`, h.Locators)
	}
	h.callRun(``, h.Value)
	h.runParams.StreamFunc(h.Name, `输入内容`+h.Value+`，成功`)
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
	h.runParams.StreamFunc(h.Name, fmt.Sprintf(`等待接口%s执行完，最多等待%ds`, responseUrl, waitResponse.WaitSecond))
	for i := 0; i < waitResponse.WaitSecond; i++ {
		for _, v := range h.runParams.ResponseUrls {
			if v.Url == responseUrl {
				h.runParams.StreamFunc(h.Name, fmt.Sprintf(`等待接口%s执行完，完成`, responseUrl))
				h.log.Debugf(`等待返回 %s 成功`, responseUrl)
				return define.ProcessOk, responseUrl, nil
			} else {
				h.runParams.StreamFunc(h.Name, fmt.Sprintf(`等待接口%s执行完，等待中..`, responseUrl))
			}
		}
		time.Sleep(time.Second)
	}
	h.runParams.StreamFunc(h.Name, fmt.Sprintf(`等待返回 %s 超时`, responseUrl))
	return define.ProcessBreak, fmt.Sprintf(`等待%s超时`, waitResponse.ResponseUrl), gstool.Error(`等待%s超时`, waitResponse.ResponseUrl)
}

func (h *Process) PRedirect() (define.ProcessCode, string, error) {
	//尝试解析
	processRedirect := _struct.ProcessRedirect{}
	_ = gstool.JsonDecode(h.Value, &processRedirect)
	//走多条件
	if processRedirect.Url != `` {
		for _, v := range processRedirect.RegisterResponseUrl {
			parseU, _ := url.Parse((*h.Page).URL())
			responseUrl := gstool.SReplaces(v.Url, map[string]string{
				`{domain}`: parseU.Host,
				`{scheme}`: parseU.Scheme,
			})
			h.runParams.StreamFunc(h.Name, `跳转地址后，注册需要等待执行完的url `+responseUrl)
			go func() {
				_, _ = (*h.Page).ExpectResponse(responseUrl, func() error {
					h.runParams.ResponseUrls = append(h.runParams.ResponseUrls, &_struct.ProcessResponseUrl{
						Url: responseUrl,
					})
					h.runParams.StreamFunc(h.Name, responseUrl+`执行完，准备往下执行`)
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
		h.runParams.StreamFunc(h.Name, `跳转地址 `+targetUrl)
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
		targetUrl := ``
		if strings.HasPrefix(redirectUri, `http://`) || strings.HasPrefix(redirectUri, `https://`) {
			targetUrl = redirectUri
		} else {
			targetUrl = domain + redirectUri
		}
		time.Sleep(time.Second)
		h.runParams.StreamFunc(h.Name, `跳转地址 `+targetUrl)
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
		h.runParams.StreamFunc(h.Name, `等待`+h.Value+`ms后结束page（后台执行）`)
		time.Sleep(time.Duration(cast.ToInt(h.Value)) * time.Second)
		_ = (*h.Page).Close()
	}()
	return define.ProcessOk, ``, nil
}

func (h *Process) PDomain() (define.ProcessCode, string, error) {
	if h.DomainLimit != `` && !strings.Contains(h.Domain, h.DomainLimit) {
		h.runParams.StreamFunc(h.Name, `域名`+h.Domain+`不允许执行`)
		return define.ProcessContinue, `域名过滤`, nil
	}
	return define.ProcessOk, ``, nil
}

func (h *Process) PChecks() (define.ProcessCode, string, error) {
	//不需要执行
	ignoreTypeList := []define.ProcessType{
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
	h.runParams.StreamFunc(h.Name, `关闭page`)
	return define.ProcessBreak, `页面关闭，结束`, nil
}

func (h *Process) PWait() (define.ProcessCode, string, error) {
	base.Component.TPlaywright.AddTipMsg(h.Page, h.Tip)
	time.Sleep(time.Duration(cast.ToInt(h.Value)) * time.Millisecond)
	h.runParams.StreamFunc(h.Name, `等待`+cast.ToString(cast.ToInt(h.Value)*1000)+`ms`)
	return define.ProcessOk, ``, nil
}

func (h *Process) callRun(errmsg, content string) {
	if h.RunCallFunc != nil {
		h.RunCallFunc(h.ProcessType, errmsg, h.Tip, content)
	}
}
