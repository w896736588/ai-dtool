package p_api

import (
	"errors"
	"net/http"
	"time"

	"gitee.com/Sxiaobai/gs/gshttp"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

type BaseInfo struct {
	Id            string            `json:"id"`
	Name          string            `json:"name"`
	FolderId      string            `json:"folder_id"`
	CollectionId  string            `json:"collection_id"`
	Method        string            `json:"method"`
	Url           string            `json:"url"`
	Protocol      string            `json:"protocol"`
	Desc          string            `json:"desc"`
	ContentType   string            `json:"content_type"`
	Headers       map[string]string `json:"headers"`
	QueryParams   string            `json:"query_params"`
	BodyForm      string            `json:"body_form"`       // application/x-www-form-urlencoded
	BodyJson      string            `json:"body_json"`       // application/json
	BodyMultiForm string            `json:"body_multi_form"` // multipart/form-data
}

type Result struct {
	Url         string            `json:"url"`         //请求的url 如果是get那么就是完整的链接
	StatusCode  int               `json:"status_code"` //http状态码
	Errmsg      string            `json:"errmsg"`      //请求错误描述
	Result      string            `json:"result"`      //请求返回
	Status      string            `json:"status"`      //status
	Millisecond int64             `json:"millisecond"` //花费的时间
	Headers     map[string]string `json:"headers"`     //header
}

type Api struct {
	BaseInfo *BaseInfo
	Result
}

func NewApi(apiInfo map[string]any) *Api {
	headers := make(map[string]string)
	_ = gstool.JsonDecode(cast.ToString(apiInfo[`headers`]), &headers)
	urlParams := make(map[string]any)
	_ = gstool.JsonDecode(cast.ToString(apiInfo[`query_params`]), &urlParams)
	requestUrl := cast.ToString(apiInfo[`protocol`]) + `://` + gstool.UrlAppendParams(cast.ToString(apiInfo[`url`]), urlParams)
	return &Api{
		BaseInfo: &BaseInfo{
			Id:           cast.ToString(apiInfo[`id`]),
			Name:         cast.ToString(apiInfo[`name`]),
			FolderId:     cast.ToString(apiInfo[`folder_id`]),
			CollectionId: cast.ToString(apiInfo[`collection_id`]),
			Method:       cast.ToString(apiInfo[`method`]),
			Url:          requestUrl,
			Protocol:     cast.ToString(apiInfo[`protocol`]),
			Desc:         cast.ToString(apiInfo[`desc`]),
			ContentType:  cast.ToString(apiInfo[`content_type`]),
			Headers:      headers,
		},
		Result: Result{},
	}
}

func (h *Api) Run() error {
	var cli *gshttp.Client
	if h.BaseInfo.Method == http.MethodPost {
		if h.BaseInfo.ContentType == `application/json` {
			h.Result.Url = h.BaseInfo.Url
			cli = gshttp.PostJson(h.BaseInfo.Url).BodyStr(h.BaseInfo.BodyJson)
		} else if h.BaseInfo.ContentType == `application/x-www-form-urlencoded` {
			urlParams := make(map[string]any)
			_ = gstool.JsonDecode(h.BaseInfo.BodyForm, &urlParams)
			h.Result.Url = h.BaseInfo.Url
			cli = gshttp.PostForm(h.BaseInfo.Url).BodyMap(urlParams)
		} else if h.BaseInfo.ContentType == `multipart/form-data` {
			urlParams := make(map[string]any)
			_ = gstool.JsonDecode(h.BaseInfo.BodyForm, &urlParams)
			h.Result.Url = h.BaseInfo.Url
			cli = gshttp.PostMultiForm(h.BaseInfo.Url).BodyMap(urlParams)
		} else {
			return errors.New(`不支持的请求类型`)
		}
	} else {
		cli = gshttp.Get(h.Result.Url)
	}
	//填充header
	cli.Headers(h.BaseInfo.Headers)
	startMill := time.Now().UnixMilli()
	cli.Request(20)
	if cli.ErrInfo() != nil {
		return cli.ErrInfo()
	}
	var err error
	h.Result.Result, err = cli.ResultStr()
	if err != nil {
		h.Result.Errmsg = err.Error()
	}
	response := cli.Response()
	h.Result.StatusCode = response.StatusCode
	h.Result.Status = response.Status
	h.Result.Millisecond = time.Now().UnixMilli() - startMill
	h.Result.Headers = h.BaseInfo.Headers
	return nil
}
