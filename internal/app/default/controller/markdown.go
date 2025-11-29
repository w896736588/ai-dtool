package controller

import (
	"dev_tool/base"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"strings"
	"time"
)

func MarkdownAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToString(dataMap[`name`]) == `` || cast.ToString(dataMap[`markdown_type`]) == `` {
		gsgin.GinResponseError(c, `名称，类型不能为空`, nil)
		return
	}
	id, err := base.Component.TSqlite.MarkdownAdd(dataMap[`id`], dataMap[`name`], dataMap[`markdown_type`], dataMap[`content`])
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

// MarkdownSort 排序
func MarkdownSort(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	markdownIds := cast.ToString(dataMap[`markdown_ids`])
	if markdownIds == `` {
		gsgin.GinResponseError(c, `markdown_ids不能为空`, nil)
		return
	}
	markdownIdsArr := strings.Split(markdownIds, `,`)
	for index, item := range markdownIdsArr {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_markdown`, map[string]any{
			`id`: cast.ToInt(item),
		}, map[string]interface{}{
			`weight`: index + 1,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func MarkdownHistoryDel(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	_, err := base.Component.TSqlite.MarkdownHistoryDel(dataMap[`id`])
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func MarkdownList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToString(dataMap[`markdown_type`]) == `` {
		gsgin.GinResponseError(c, `名称，类型不能为空`, nil)
		return
	}
	starList, err := base.Component.TSqlite.MarkdownList(cast.ToString(dataMap[`markdown_type`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, starList)
}

func MarkdownHistoryList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `文档id不能为空`, nil)
		return
	}
	starList, err := base.Component.TSqlite.MarkdownHistoryList(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	for i := range starList {
		starList[i][`create_time_desc`] = gstool.TimeUnixToString(time.Unix(cast.ToInt64(starList[i][`create_time`]), 0), `Y-m-d H:i:s`)
	}
	gsgin.GinResponseSuccess(c, ``, starList)
}
