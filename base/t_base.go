package base

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

type TBase struct {
	StartMillUnix int64
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

func (h *TBase) GetUnique(prefix string) string {
	h.StartMillUnix += 1
	return fmt.Sprintf(`%s%d`, prefix, h.StartMillUnix)
}
