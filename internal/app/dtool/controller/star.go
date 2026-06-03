package controller

import (
	"dev_tool/internal/app/dtool/common"
	"github.com/gin-gonic/gin"
	"github.com/w896736588/go-tool/gsgin"
)

func StarAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	_, err := common.DbMain.StarAdd(dataMap[`id`], dataMap[`name`], dataMap[`key`], dataMap[`value`], dataMap[`type`])
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func StarDel(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	_, err := common.DbMain.StarDel(dataMap[`id`])
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func StarList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	starList, err := common.DbMain.StarList(dataMap[`type`])
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, starList)
}
