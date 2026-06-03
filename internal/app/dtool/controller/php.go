package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

// PhpPhpUnSerialize 反序列化（php）
func PhpPhpUnSerialize(c *gin.Context) {
	var err error
	reqMap, err := GetGlobalReqParamsM(c)
	var out interface{}
	out, err = gstool.PhpUnSerialize(cast.ToString(reqMap[`SerializeStr`]))
	if err != nil {
		gstool.FmtPrintlnLog(`反序列化失败 %s`, err.Error())
		gsgin.GinResponseError(c, err.Error(), reqMap[`SerializeStr`])
		return
	}
	gsgin.GinResponseSuccess(c, `成功`, gstool.JsonEncode(out))
}

// PhpPhpUnSerialize2 反序列化（php）
func PhpPhpUnSerialize2(c *gin.Context) {
	var err error
	reqMap, err := GetGlobalReqParamsM(c)
	var out interface{}
	out, err = gstool.PhpUnSerialize(cast.ToString(reqMap[`SerializeStr`]))
	if err != nil {
		gstool.FmtPrintlnLog(`反序列化失败 %s`, err.Error())
		gsgin.GinResponseSuccess(c, err.Error(), reqMap[`SerializeStr`])
		return
	}
	rList := make([]any, 0)
	if slice, ok := out.([]any); ok {
		for _, v := range slice {
			dData := make(map[string]any)
			dErr := gstool.JsonDecode(cast.ToString(v), &dData)
			if dErr != nil {
				rList = append(rList, v)
			} else {
				rList = append(rList, gstool.JsonFormat(dData))
			}
		}
	} else {
		rList = append(rList, out)
	}
	gsgin.GinResponseSuccess(c, `成功`, rList)
}
