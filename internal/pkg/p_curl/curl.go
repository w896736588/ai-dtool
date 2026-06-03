package p_curl

import (
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/struct"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gshttp"
	"github.com/w896736588/go-tool/gshttp/stream"
	"github.com/w896736588/go-tool/gstool"
)

type CurlRun struct {
	ParseConfig _struct.CurlParseConfig //基础配置
	CurlEvents  _struct.CurlEvents
}

func NewCurlRun(parseConfig _struct.CurlParseConfig, curlEvents _struct.CurlEvents) *CurlRun {
	if curlEvents.StartCall == nil {
		curlEvents.StartCall = func() {}
	}
	if curlEvents.NoticeCall == nil {
		curlEvents.NoticeCall = func(s string) {}
	}
	if curlEvents.StreamDataCall == nil {
		curlEvents.StreamDataCall = func(s string) {}
	}
	if curlEvents.EndCall == nil {
		curlEvents.EndCall = func() {}
	}
	return &CurlRun{
		ParseConfig: parseConfig,
		CurlEvents:  curlEvents,
	}
}

func (h *CurlRun) Run() (string, error) {
	retryNum := max(1, h.ParseConfig.Retry)
	retryWaitSecond := max(2, h.ParseConfig.RetrySecond)
	var err error
	for i := 0; i < retryNum; i++ { //重试
		cli, err := h.GetGsHttpClient()
		if err != nil {
			return ``, err
		}
		var res []byte
		if h.ParseConfig.IsStream == 1 {
			res, err = h.streamRun(cli)
		} else {
			res, err = cli.Request(200).Result()
		}
		if err == nil {
			h.CurlEvents.NoticeCall(fmt.Sprintf(`%s 第%d次尝试，成功`, "\n"+gstool.TimeNowUnixToString(`Y-m-d H:i:s`), i) + "\n")
			return cast.ToString(res), nil
		} else {
			h.CurlEvents.NoticeCall(fmt.Sprintf(`%s 第%d次尝试，失败 %s`, "\n"+gstool.TimeNowUnixToString(`Y-m-d H:i:s`), i, err.Error()) + "\n")
			time.Sleep(time.Second * time.Duration(retryWaitSecond))
		}
	}
	return ``, err
}

func (h *CurlRun) streamRun(cli *gshttp.Client) ([]byte, error) {
	h.CurlEvents.StartCall()
	var fac gshttp.StreamInterface
	if len(h.ParseConfig.ReceiveRegex) > 0 {
		fac = &stream.Reges{
			Reges: h.ParseConfig.ReceiveRegex,
			CallFunc: func(s string, err error) {
				h.CurlEvents.StreamDataCall(s)
			},
			FormatFunc: nil,
		}
	} else if len(h.ParseConfig.ReceiveSignal) > 0 {
		fac = &stream.Byts{
			Byts: []byte(h.ParseConfig.ReceiveSignal),
			CallFunc: func(s string, err error) {
				h.CurlEvents.StreamDataCall(s)
			},
			FormatFunc: nil,
		}
		gstool.FmtPrintlnLogTime(`按照字符串进行分割`)
	}
	if fac != nil {
		return cli.SetStreamFac(fac).Request(200).Result()
	} else {
		return cli.Request(200).Result()
	}
}

func (h *CurlRun) GetGsHttpClient() (*gshttp.Client, error) {
	if h.ParseConfig.ContentType == define.ContentTypeJson {
		return gshttp.PostJson(h.ParseConfig.Url).
			BodyStr(h.ParseConfig.Body).
			Headers(h.ParseConfig.Headers), nil
	} else if h.ParseConfig.ContentType == define.ContentTypeForm {
		return gshttp.PostForm(h.ParseConfig.Url).
			BodyStr(h.ParseConfig.Body).
			Headers(h.ParseConfig.Headers), nil
	} else if h.ParseConfig.ContentType == define.ContentTypeMultiForm {
		return gshttp.PostMultiForm(h.ParseConfig.Url).
			BodyStr(h.ParseConfig.Body).
			Headers(h.ParseConfig.Headers), nil
	} else if h.ParseConfig.ContentType == define.ContentTypeText || h.ParseConfig.ContentType == define.ContentTypeRaw {
		contentType := h.ParseConfig.ContentType
		if contentType == define.ContentTypeRaw {
			contentType = `application/octet-stream`
		}
		h.ParseConfig.Headers[`Content-Type`] = contentType
		return gshttp.PostJson(h.ParseConfig.Url).
			BodyStr(h.ParseConfig.Body).
			Headers(h.ParseConfig.Headers), nil
	} else if h.ParseConfig.Method == http.MethodGet {
		return gshttp.Get(h.ParseConfig.Url).
			Headers(h.ParseConfig.Headers), nil
	} else {
		return nil, errors.New(`不支持的请求配置`)
	}
}
