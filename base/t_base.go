package base

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

type TBase struct {
}

// GetCombineKey 组装key
func (h *TBase) GetCombineKey(params ...any) string {
	strParamsList := gstool.Array2Str(&params)
	return strings.Join(strParamsList, `#`)
}

// ExplainCombineKey 分解key
func (h *TBase) ExplainCombineKey(uniqueKey string) []string {
	return strings.Split(uniqueKey, `#`)
}
