package p_api

import (
	"dev_tool/base"
	"errors"
	"net/http"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/gstool"
	"gitee.com/Sxiaobai/gs/v2/gshttp"
	"github.com/spf13/cast"
)

type BaseInfo struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	FolderId     string            `json:"folder_id"`
	CollectionId string            `json:"collection_id"`
	ResponseTake []ResponseTake    `json:"response_take"` // 提取
	EnvItems     map[string]string `json:"env_items"`     //环境变量
	EnvId        int               `json:"env_id"`        //所属环境变量
}

type ResponseTake struct {
	Description string `json:"description"`
	ItemKey     string `json:"item_key"`
	Value       string `json:"value"`
	TakeValue   string `json:"take_value"`
}

type Result struct {
	Url          string            `json:"url"`           //请求的url 如果是get那么就是完整的链接
	StatusCode   int               `json:"status_code"`   //http状态码
	Errmsg       string            `json:"errmsg"`        //请求错误描述
	Result       string            `json:"result"`        //请求返回
	Status       string            `json:"status"`        //status
	Millisecond  int64             `json:"millisecond"`   //花费的时间
	Headers      map[string]string `json:"headers"`       //header
	BodyForms    []map[string]any  `json:"body_forms"`    //提交的Form
	ResponseTake []ResponseTake    `json:"response_take"` //返回参数的提取
	RequestTime  string            `json:"request_time"`  //发起请求时间
}

type Api struct {
	BaseInfo   *BaseInfo
	CurlStruct CurlStruct
	Result
}

func NewApi(apiInfo map[string]any) *Api {
	headers := make(map[string]string)
	_ = gstool.JsonDecode(cast.ToString(apiInfo[`headers`]), &headers)
	urlParams := make(map[string]any)
	_ = gstool.JsonDecode(cast.ToString(apiInfo[`query_params`]), &urlParams)
	url, _ := gstool.UrlDecode(gstool.UrlAppendParams(cast.ToString(apiInfo[`url`]), urlParams))
	requestUrl := cast.ToString(apiInfo[`protocol`]) + `://` + url
	envItems := make(map[string]string)
	if cast.ToInt(apiInfo[`env_id`]) > 0 {
		envItemList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
			`env_id`: apiInfo[`env_id`],
		}).All()
		for _, envItem := range envItemList {
			envItems[`$`+cast.ToString(envItem[`key`])+`$`] = cast.ToString(envItem[`value`])
		}
	}
	//response take
	responseTake := make([]ResponseTake, 0)
	_ = gstool.JsonDecode(cast.ToString(apiInfo[`response_take`]), &responseTake)
	//body form
	bodyFormData := make([]KeyValue, 0)
	err := gstool.JsonDecode(cast.ToString(apiInfo[`body_form`]), &bodyFormData)
	if err != nil {
		gstool.FmtPrintlnLogTime(`解析bodyForm(%s)失败，%s`, cast.ToString(apiInfo[`body_form`]), err.Error())
	}
	return &Api{
		BaseInfo: &BaseInfo{
			Id:           cast.ToString(apiInfo[`id`]),
			Name:         cast.ToString(apiInfo[`name`]),
			FolderId:     cast.ToString(apiInfo[`folder_id`]),
			CollectionId: cast.ToString(apiInfo[`collection_id`]),
			ResponseTake: responseTake,
			EnvItems:     envItems,
			EnvId:        cast.ToInt(apiInfo[`env_id`]),
		},
		CurlStruct: CurlStruct{
			Method:      cast.ToString(apiInfo[`method`]),
			Url:         requestUrl,
			Protocol:    cast.ToString(apiInfo[`protocol`]),
			ContentType: cast.ToString(apiInfo[`content_type`]),
			Headers:     headers,
			BodyForm:    bodyFormData,
		},
		Result: Result{
			ResponseTake: responseTake,
			RequestTime:  gstool.TimeNowUnixToString(`Y-m-d H:i:s`),
		},
	}
}

func (h *Api) ReplaceEnv() {
	//url替换
	h.CurlStruct.Url = gstool.SReplaces(h.CurlStruct.Url, h.BaseInfo.EnvItems)
	//headers替换
	for k, v := range h.CurlStruct.Headers {
		h.CurlStruct.Headers[k] = gstool.SReplaces(v, h.BaseInfo.EnvItems)
	}
	//body form替换
	for k, v := range h.CurlStruct.BodyForm {
		h.CurlStruct.BodyForm[k].Value = gstool.SReplaces(v.Value, h.BaseInfo.EnvItems)
	}
	//body json替换
	h.CurlStruct.Body = gstool.SReplaces(h.CurlStruct.Body, h.BaseInfo.EnvItems)
}

func (h *Api) Run() error {
	var cli *gshttp.Client
	h.ReplaceEnv()
	if h.CurlStruct.Method == http.MethodPost {
		if h.CurlStruct.ContentType == `application/json` {
			h.Result.Url = h.CurlStruct.Url
			cli = gshttp.PostJson(h.CurlStruct.Url).BodyStr(h.CurlStruct.Body)
		} else if h.CurlStruct.ContentType == `application/x-www-form-urlencoded` {
			h.Result.Url = h.CurlStruct.Url
			cli = gshttp.PostForm(h.CurlStruct.Url)
			err := h.FormatBodyData(cli, h.CurlStruct.BodyForm)
			if err != nil {
				return err
			}
		} else if h.CurlStruct.ContentType == `multipart/form-data` {
			h.Result.Url = h.CurlStruct.Url
			cli = gshttp.PostMultiForm(h.CurlStruct.Url)
			err := h.FormatBodyData(cli, h.CurlStruct.BodyForm)
			if err != nil {
				return err
			}
		} else {
			return errors.New(`不支持的请求类型`)
		}
	} else {
		h.Result.Url = h.CurlStruct.Url
		cli = gshttp.Get(h.Result.Url)
	}
	//填充header
	cli.Headers(h.CurlStruct.Headers)
	h.Result.Headers = cli.GetHeaders()
	startMill := time.Now().UnixMilli()
	cli.Request(20)
	if cli.ErrInfo() != nil {
		h.Result.Millisecond = time.Now().UnixMilli() - startMill
		h.Result.Errmsg = cli.ErrInfo().Error()
		h.Result.Result = ``
		return nil
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
	return nil
}

func (h *Api) FormatBodyData(cli *gshttp.Client, bodyForm []KeyValue) error {
	resultBodyForms := make([]map[string]any, 0)
	//塞入的数据 所有的数据以数组的形式存入 如果是一个那么自然是单个，如果是多个就自动是数组传递
	bodyMaps := make(map[string][]any, 0)
	for _, keyValue := range bodyForm {
		//如果字段存在
		if _, ok := bodyMaps[keyValue.Field]; !ok {
			bodyMaps[keyValue.Field] = []any{}
		}
		if keyValue.Type == `string` {
			bodyMaps[keyValue.Field] = append(bodyMaps[keyValue.Field], cast.ToString(keyValue.Value))
			if h.CurlStruct.ContentType == `application/x-www-form-urlencoded` {
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
			if h.CurlStruct.ContentType == `application/x-www-form-urlencoded` {
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
			if h.CurlStruct.ContentType == `application/x-www-form-urlencoded` {
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
			if h.CurlStruct.ContentType == `application/x-www-form-urlencoded` {
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
	h.Result.BodyForms = resultBodyForms
	return nil
}

func (h *Api) ResponseTake() {
	h.Result.ResponseTake = make([]ResponseTake, 0)
	if h.Result.Result != `` {
		extra, err := gstool.NewJsonExtractorFromJSON(h.Result.Result)
		if err != nil {
			gstool.FmtPrintlnLogTime(`参数提取失败 %s`, err.Error())
			return
		}
		for _, take := range h.BaseInfo.ResponseTake {
			value := strings.TrimLeft(take.Value, `res.`)
			takeValue, err := extra.Extract(value)
			if err != nil {
				continue
			}
			if takeValue == nil {
				continue
			}
			take.TakeValue = cast.ToString(takeValue)
			h.Result.ResponseTake = append(h.Result.ResponseTake, take)
			//反写到环境变量
			_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_api_env_item`, map[string]any{
				`env_id`: h.BaseInfo.EnvId,
				`key`:    take.ItemKey,
			}, map[string]any{
				`value`:       take.TakeValue,
				`update_time`: time.Now().Unix(),
			}).Exec()
		}
	}
}
