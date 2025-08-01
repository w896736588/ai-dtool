package controller

import (
	"dev_tool/base"
	"gitee.com/Sxiaobai/gs/gsgin"
	"github.com/gin-gonic/gin"
)

func MarkdownAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	id, err := base.Component.TSqlite.MarkdownAdd(dataMap[`id`], dataMap[`name`], dataMap[`content`])
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`id`: id,
	})
}

func MarkdownDel(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	_, err := base.Component.TSqlite.MarkdownDel(dataMap[`id`])
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func MarkdownList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	starList, err := base.Component.TSqlite.MarkdownList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, starList)
}
