package api

import (
	"dev_tool/internal/pkg/p_curl"
)

type ApiDefine struct {
	Method      string            `json:"method"`
	Url         string            `json:"url"`
	Protocol    string            `json:"protocol"`
	Desc        string            `json:"desc"`
	ContentType string            `json:"content_type"`
	Headers     map[string]string `json:"headers"`
	QueryParams string            `json:"query_params"`
	BodyForm    []p_curl.KeyValue `json:"body_form"` // application/x-www-form-urlencoded
	BodyJson    string            `json:"body_json"` // application/json
	BodyRaw     string            `json:"body_raw"`  // text/plain 或 raw
}
