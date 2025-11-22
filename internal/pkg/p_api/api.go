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
	BodyForms   []map[string]any  `json:"body_forms"`  //提交的Form
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
			BodyForm:     cast.ToString(apiInfo[`body_form`]),
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
			h.Result.Url = h.BaseInfo.Url
			cli = gshttp.PostForm(h.BaseInfo.Url)
			err := h.FormatBodyData(cli, h.BaseInfo.BodyForm)
			if err != nil {
				return err
			}
		} else if h.BaseInfo.ContentType == `multipart/form-data` {
			h.Result.Url = h.BaseInfo.Url
			cli = gshttp.PostMultiForm(h.BaseInfo.Url)
			err := h.FormatBodyData(cli, h.BaseInfo.BodyForm)
			if err != nil {
				return err
			}
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
	h.Result.Headers = cli.GetHeaders()
	return nil
}

type KeyValue struct {
	Description string `json:"description"`
	Field       string `json:"field"`
	Type        string `json:"type"`
	Value       string `json:"value"`
}

func (h *Api) FormatBodyData(cli *gshttp.Client, bodyForm string) error {
	resultBodyForms := make([]map[string]any, 0)
	keyValueList := make([]KeyValue, 0)
	err := gstool.JsonDecode(bodyForm, &keyValueList)
	if err != nil {
		return gstool.Error(`解析bodyForm(%s)失败，%s`, bodyForm, err.Error())
	}
	//塞入的数据 所有的数据以数组的形式存入 如果是一个那么自然是单个，如果是多个就自动是数组传递
	bodyMaps := make(map[string][]any, 0)
	for _, keyValue := range keyValueList {
		//如果字段存在
		if _, ok := bodyMaps[keyValue.Field]; !ok {
			bodyMaps[keyValue.Field] = []any{}
		}
		if keyValue.Type == `string` {
			bodyMaps[keyValue.Field] = append(bodyMaps[keyValue.Field], cast.ToString(keyValue.Value))
			if h.BaseInfo.ContentType == `application/x-www-form-urlencoded` {
				resultBodyForms = append(resultBodyForms, map[string]any{
					`field`: keyValue.Field,
					`type`:  keyValue.Type,
					`value`: gstool.UrlEncode(keyValue.Value),
				})
			} else {
				resultBodyForms = append(resultBodyForms, map[string]any{
					`field`: keyValue.Field,
					`type`:  keyValue.Type,
					`value`: keyValue.Value,
				})
			}
		} else if keyValue.Type == `file` {
			cli.BodyFile(keyValue.Field, keyValue.Value, gstool.FileGetNameByPath(keyValue.Value))
		} else if keyValue.Type == `integer` {
			bodyMaps[keyValue.Field] = append(bodyMaps[keyValue.Field], cast.ToInt(keyValue.Value))
			if h.BaseInfo.ContentType == `application/x-www-form-urlencoded` {
				resultBodyForms = append(resultBodyForms, map[string]any{
					`field`: keyValue.Field,
					`type`:  keyValue.Type,
					`value`: gstool.UrlEncode(cast.ToString(cast.ToInt(keyValue.Value))),
				})
			} else {
				resultBodyForms = append(resultBodyForms, map[string]any{
					`field`: keyValue.Field,
					`type`:  keyValue.Type,
					`value`: cast.ToInt(keyValue.Value),
				})
			}
		} else if keyValue.Type == `float` {
			bodyMaps[keyValue.Field] = append(bodyMaps[keyValue.Field], cast.ToFloat64(keyValue.Value))
			if h.BaseInfo.ContentType == `application/x-www-form-urlencoded` {
				resultBodyForms = append(resultBodyForms, map[string]any{
					`field`: keyValue.Field,
					`type`:  keyValue.Type,
					`value`: gstool.UrlEncode(cast.ToString(cast.ToFloat64(keyValue.Value))),
				})
			} else {
				resultBodyForms = append(resultBodyForms, map[string]any{
					`field`: keyValue.Field,
					`type`:  keyValue.Type,
					`value`: cast.ToFloat64(keyValue.Value),
				})
			}
		} else if keyValue.Type == `boolean` {
			setValue := false
			if keyValue.Value == `true` {
				setValue = true
			} else if keyValue.Value == `false` {
				setValue = false
			}
			bodyMaps[keyValue.Field] = append(bodyMaps[keyValue.Field], setValue)
			if h.BaseInfo.ContentType == `application/x-www-form-urlencoded` {
				resultBodyForms = append(resultBodyForms, map[string]any{
					`field`: keyValue.Field,
					`type`:  keyValue.Type,
					`value`: gstool.UrlEncode(cast.ToString(keyValue.Value)),
				})
			} else {
				resultBodyForms = append(resultBodyForms, map[string]any{
					`field`: keyValue.Field,
					`type`:  keyValue.Type,
					`value`: cast.ToString(keyValue.Value),
				})
			}
		} else {
			return errors.New(`不支持的参数类型(` + keyValue.Type + `)`)
		}
	}
	//最终再次转换
	bodyMap := make(map[string]any)
	for k, v := range bodyMaps {
		if len(v) == 1 {
			bodyMap[k] = v[0]
		} else {
			bodyMap[k] = v // 保持数组格式
		}
	}
	cli.BodyMap(bodyMap)
	gstool.FmtPrintlnLogTime(`请求的bodyMap %v`, bodyMap)
	h.Result.BodyForms = resultBodyForms
	return nil
}
