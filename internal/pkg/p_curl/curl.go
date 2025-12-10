package p_curl

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/v2/gshttp"
	"gitee.com/Sxiaobai/gs/v2/gshttp/stream"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
	"net/http"
	"time"
)

type CurlRun struct {
	IsStream      int               //1流式 流式仅适用于post
	Method        string            //请求方式
	Url           string            //地址
	ContentType   string            //请求方式
	Headers       map[string]string //请求头
	Body          string            //body数据
	ReceiveSignal string            //流式接收时按字符串进行分割
	ReceiveRegex  string            //流式接收时按正则进行分割
	TakeJsons     []struct {        //json结果规则
		Take string `json:"take"` //json提取
	} `json:"take_jsons"`
	Retry          int          `json:"retry"`        //尝试次数
	RetrySecond    int          `json:"retry_second"` //每次尝试间隔几秒
	StreamDataCall func(string) //流式接收到数据后回调
	NoticeCall     func(string) //正常消息返回 不时http的返回
	EndCall        func()       //请求结束的返回
}

func (h *CurlRun) Run() (string, error) {
	retryNum := max(1, h.Retry)
	retryWaitSecond := max(2, h.RetrySecond)
	var err error
	for i := 0; i < retryNum; i++ { //重试
		base.Component.GsLog.Debugf(`----------------------`)
		cli, err := h.GetGsHttpClient()
		if err != nil {
			return ``, err
		}
		var res []byte
		if h.IsStream == 1 {
			res, err = h.streamRun(cli)
		} else {
			res, err = cli.Request(200).Result()
		}
		if err == nil {
			h.NoticeCall(fmt.Sprintf(`%s 第%d次尝试，成功`, "\n"+gstool.TimeNowUnixToString(`Y-m-d H:i:s`), i))
			return cast.ToString(res), nil
		} else {
			h.NoticeCall(fmt.Sprintf(`%s 第%d次尝试，失败 %s`, "\n"+gstool.TimeNowUnixToString(`Y-m-d H:i:s`), i, err.Error()))
			time.Sleep(time.Second * time.Duration(retryWaitSecond))
		}
	}
	return ``, err
}

func (h *CurlRun) streamRun(cli *gshttp.Client) ([]byte, error) {
	var fac gshttp.StreamInterface
	if len(h.ReceiveRegex) > 0 {
		base.Component.TVariable.Log.Debugf(`通过正则分割接收 %q`, h.ReceiveRegex)
		fac = &stream.Reges{
			Reges: h.ReceiveRegex,
			CallFunc: func(s string, err error) {
				h.StreamDataCall(s)
			},
			FormatFunc: nil,
		}
	} else if len(h.ReceiveSignal) > 0 {
		fac = &stream.Byts{
			Byts: []byte(h.ReceiveSignal),
			CallFunc: func(s string, err error) {
				h.StreamDataCall(s)
			},
			FormatFunc: nil,
		}
	}
	if fac != nil {
		return cli.SetStreamFac(fac).Request(200).Result()
	} else {
		return cli.Request(200).Result()
	}
}

func (h *CurlRun) GetGsHttpClient() (*gshttp.Client, error) {
	if h.ContentType == define.ContentTypeJson {
		return gshttp.PostJson(h.Url).
			BodyStr(h.Body).
			Headers(h.Headers), nil
	} else if h.ContentType == define.ContentTypeForm {
		return gshttp.PostForm(h.Url).
			BodyStr(h.Body).
			Headers(h.Headers), nil
	} else if h.ContentType == define.ContentTypeMultiForm {
		return gshttp.PostMultiForm(h.Url).
			BodyStr(h.Body).
			Headers(h.Headers), nil
	} else if h.Method == http.MethodGet {
		return gshttp.Get(h.Url).
			Headers(h.Headers), nil
	} else {
		return nil, errors.New(`不支持的请求配置`)
	}
}
