package p_api

import (
	"gitee.com/Sxiaobai/gs/gstool"
)

const (
	FieldTypeString = "string"
	FieldTypeInt    = "int"
	FieldTypeFloat  = "float"
	FieldTypeBool   = "bool"
	FieldTypeFile   = "file"
)

type ApiDefine struct {
	Method      string            `json:"method"`
	Url         string            `json:"url"`
	Protocol    string            `json:"protocol"`
	Desc        string            `json:"desc"`
	ContentType string            `json:"content_type"`
	Headers     map[string]string `json:"headers"`
	QueryParams string            `json:"query_params"`
	BodyForm    []KeyValue        `json:"body_form"` // application/x-www-form-urlencoded
	BodyJson    string            `json:"body_json"` // application/json
}

type KeyValue struct {
	Description string `json:"description"`
	Field       string `json:"field"`
	Type        string `json:"type"`
	Value       string `json:"value"`
}

// UrlParseParams 从URL字符串中解析参数
func UrlParseParams(urlStr string) []KeyValue {
	params, err := gstool.UrlParseParams(urlStr)
	if err != nil {
		return []KeyValue{}
	}
	var result []KeyValue
	for _, paramValue := range params {
		for key, value := range paramValue {
			result = append(result, KeyValue{
				Description: "",
				Field:       key,
				Type:        FieldTypeString,
				Value:       value,
			})
		}
	}
	return result
}
