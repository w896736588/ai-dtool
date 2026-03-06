package controller

import (
	"dev_tool/internal/app/dtool/common"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// MemoryFragmentList 查询记忆片段列表。
func MemoryFragmentList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	list, err := common.DbMain.MemoryFragmentList(cast.ToInt(dataMap[`limit`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentInfo 查询单个记忆片段详情。
func MemoryFragmentInfo(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	info, err := common.DbMain.MemoryFragmentInfo(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// MemoryFragmentSave 保存记忆片段。
func MemoryFragmentSave(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	info, err := common.DbMain.MemoryFragmentSave(
		cast.ToInt(dataMap[`id`]),
		cast.ToString(dataMap[`title`]),
		cast.ToString(dataMap[`content`]),
		memoryFragmentParseTags(dataMap[`tags`]),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// MemoryFragmentDelete 软删除记忆片段。
func MemoryFragmentDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	_, err := common.DbMain.MemoryFragmentSoftDelete(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// MemoryFragmentHistoryList 查询记忆片段历史记录。
func MemoryFragmentHistoryList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	list, err := common.DbMain.MemoryFragmentHistoryList(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentTagList 查询记忆片段标签列表。
func MemoryFragmentTagList(c *gin.Context) {
	list, err := common.DbMain.MemoryFragmentTagList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentSearch 搜索记忆片段。
func MemoryFragmentSearch(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	list, err := common.DbMain.MemoryFragmentSearch(
		cast.ToString(dataMap[`mode`]),
		cast.ToString(dataMap[`query`]),
		memoryFragmentParseTags(dataMap[`selected_tags`]),
		cast.ToInt(dataMap[`limit`]),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// memoryFragmentParseTags 解析请求中的标签数组。
func memoryFragmentParseTags(raw any) []string {
	switch value := raw.(type) {
	case []string:
		return value
	case []any:
		result := make([]string, 0, len(value))
		for _, item := range value {
			result = append(result, cast.ToString(item))
		}
		return result
	case string:
		if value == `` {
			return []string{}
		}
		return []string{value}
	default:
		return []string{}
	}
}
