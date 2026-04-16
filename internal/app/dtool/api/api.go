package api

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_curl"
	"errors"
	"net/http"
	url2 "net/url"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gshttp"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
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
	Url             string            `json:"url"`              //请求的url 如果是get那么就是完整的链接
	StatusCode      int               `json:"status_code"`      //http状态码
	Errmsg          string            `json:"errmsg"`           //请求错误描述
	Result          string            `json:"result"`           //请求返回
	Status          string            `json:"status"`           //status
	Millisecond     int64             `json:"millisecond"`      //花费的时间
	Headers         map[string]string `json:"headers"`          //兼容旧字段，保留请求头
	RequestHeaders  map[string]string `json:"request_headers"`  //请求头
	ResponseHeaders map[string]string `json:"response_headers"` //返回头
	BodyForms       []map[string]any  `json:"body_forms"`       //提交的Form
	BodyRaw         string            `json:"body_raw"`         //请求体（application/json / text/plain / raw）
	ResponseTake    []ResponseTake    `json:"response_take"`    //返回参数的提取
	RequestTime     string            `json:"request_time"`     //发起请求时间
}

type Api struct {
	BaseInfo   *BaseInfo
	CurlStruct p_curl.CurlStruct
	Result
}

// mergeHeadersWithFolderDefaults 合并文件夹默认请求头与接口请求头，接口同名键优先。
func mergeHeadersWithFolderDefaults(folderHeaders, apiHeaders map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range folderHeaders {
		result[key] = value
	}
	for key, value := range apiHeaders {
		result[key] = value
	}
	return result
}

func NewApi(apiInfo map[string]any) *Api {
	headers := make(map[string]string)
	_ = gstool.JsonDecode(cast.ToString(apiInfo[`headers`]), &headers)
	folderHeaders := make(map[string]string)
	_ = gstool.JsonDecode(cast.ToString(apiInfo[`folder_headers`]), &folderHeaders)
	headers = mergeHeadersWithFolderDefaults(folderHeaders, headers)
	urlParams := make([]map[string]any, 0)
	_ = gstool.JsonDecode(cast.ToString(apiInfo[`query_params`]), &urlParams)
	urlValues := url2.Values{}
	for _, urlParam := range urlParams {
		urlValues.Add(cast.ToString(urlParam[`field`]), cast.ToString(urlParam[`value`]))
	}
	url, _ := gstool.UrlDecode(gstool.UrlAppendVals(cast.ToString(apiInfo[`url`]), urlValues))
	envItems := make(map[string]string)
	if cast.ToInt(apiInfo[`env_id`]) > 0 {
		envItemList, _ := component.SqliteClient.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
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
	bodyFormData := make([]p_curl.KeyValue, 0)
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
		CurlStruct: p_curl.CurlStruct{
			Method:      cast.ToString(apiInfo[`method`]),
			Url:         url,
			Protocol:    cast.ToString(apiInfo[`protocol`]),
			ContentType: cast.ToString(apiInfo[`content_type`]),
			Headers:     headers,
			BodyForm:    bodyFormData,
			BodyJson:    cast.ToString(apiInfo[`body_json`]),
			BodyRaw:     cast.ToString(apiInfo[`body_raw`]),
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
	h.CurlStruct.BodyJson = gstool.SReplaces(h.CurlStruct.BodyJson, h.BaseInfo.EnvItems)
	//body raw替换 / body raw replacement
	h.CurlStruct.BodyRaw = gstool.SReplaces(h.CurlStruct.BodyRaw, h.BaseInfo.EnvItems)
}

// collectResponseHeaders 扁平化响应头，便于前端直接展示和历史记录持久化。
func collectResponseHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) == 0 {
			result[key] = ``
			continue
		}
		result[key] = strings.Join(values, `; `)
	}
	return result
}

func (h *Api) Run() error {
	var cli *gshttp.Client
	h.ReplaceEnv()
	if h.CurlStruct.Method == http.MethodPost {
		if h.CurlStruct.ContentType == `application/json` {
			h.Result.Url = h.CurlStruct.Url
			h.Result.BodyRaw = h.CurlStruct.BodyJson
			cli = gshttp.PostJson(h.CurlStruct.Url).BodyStr(h.CurlStruct.BodyJson)
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
		} else if h.CurlStruct.ContentType == define.ContentTypeText || h.CurlStruct.ContentType == define.ContentTypeRaw {
			h.Result.Url = h.CurlStruct.Url
			h.Result.BodyRaw = h.CurlStruct.BodyRaw
			contentType := h.CurlStruct.ContentType
			if contentType == define.ContentTypeRaw {
				contentType = `application/octet-stream`
			}
			h.CurlStruct.Headers[`Content-Type`] = contentType
			cli = gshttp.PostJson(h.CurlStruct.Url).BodyStr(h.CurlStruct.BodyRaw)
		} else {
			return errors.New(`不支持的请求类型`)
		}
	} else {
		h.Result.Url = h.CurlStruct.Url
		cli = gshttp.Get(h.Result.Url)
	}
	//填充header
	cli.Headers(h.CurlStruct.Headers)
	h.Result.RequestHeaders = cli.GetHeaders()
	h.Result.Headers = h.Result.RequestHeaders
	h.Result.ResponseHeaders = map[string]string{}
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
	if response != nil {
		h.Result.ResponseHeaders = collectResponseHeaders(response.Header)
	}
	h.Result.StatusCode = response.StatusCode
	h.Result.Status = response.Status
	h.Result.Millisecond = time.Now().UnixMilli() - startMill
	return nil
}

func (h *Api) FormatBodyData(cli *gshttp.Client, bodyForm []p_curl.KeyValue) error {
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
			resultBodyForms = append(resultBodyForms, map[string]any{
				`field`: keyValue.Field,
				`type`:  keyValue.Type,
				`value`: keyValue.Value,
			})
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
		for _, take := range h.BaseInfo.ResponseTake {
			value := strings.TrimLeft(take.Value, `res.`)
			ret := gjson.Get(h.Result.Result, value)
			take.TakeValue = ret.String()
			h.Result.ResponseTake = append(h.Result.ResponseTake, take)
			//反写到环境变量
			_, _ = common.DbMain.Client.QuickUpdate(`tbl_api_env_item`, map[string]any{
				`env_id`: h.BaseInfo.EnvId,
				`key`:    take.ItemKey,
			}, map[string]any{
				`value`:       take.TakeValue,
				`update_time`: time.Now().Unix(),
			}).Exec()
		}
	}
}
