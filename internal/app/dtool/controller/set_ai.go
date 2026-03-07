package controller

import (
	"dev_tool/internal/app/dtool/common"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

const (
	aiModelTypeLlm       = `llm`
	aiModelTypeEmbedding = `embedding`
)

// normalizeAiModelType 规范化模型类型，空值默认按 llm 处理。
func normalizeAiModelType(modelType string) string {
	modelType = strings.TrimSpace(strings.ToLower(modelType))
	switch modelType {
	case ``:
		return aiModelTypeLlm
	case aiModelTypeLlm, aiModelTypeEmbedding:
		return modelType
	default:
		return ``
	}
}

// SetAiProviderList 查询 AI 服务商配置列表
func SetAiProviderList(c *gin.Context) {
	sql := `select id,name,provider_type as request_format,provider_type,base_url,api_key,status,create_time,update_time 
from tbl_ai_provider 
where status = 1 
order by id desc`
	list, err := common.DbMain.Client.QueryBySql(sql).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// SetAiProviderAdd 新增或更新 AI 服务商配置
func SetAiProviderAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `base_url`, `api_key`})
	if cast.ToString(updateData[`name`]) == `` {
		gsgin.GinResponseError(c, `服务商名称不能为空`, nil)
		return
	}
	requestFormat := cast.ToString(dataMap[`request_format`])
	if requestFormat == `` {
		requestFormat = cast.ToString(dataMap[`provider_type`])
	}
	if requestFormat == `` {
		requestFormat = `openai`
	}
	if requestFormat != `openai` {
		gsgin.GinResponseError(c, `请求格式仅支持openai`, nil)
		return
	}
	updateData[`provider_type`] = requestFormat
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`status`] = 1
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, err := common.DbMain.Client.QuickCreate(`tbl_ai_provider`, updateData).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, err := common.DbMain.Client.QuickUpdate(`tbl_ai_provider`, map[string]any{
			`id`: dataMap[`id`],
		}, updateData).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetAiProviderDelete 删除 AI 服务商配置（软删除）
func SetAiProviderDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	_, err := common.DbMain.Client.QuickUpdate(`tbl_ai_provider`, map[string]any{
		`id`: dataMap[`id`],
	}, map[string]any{
		`status`:      0,
		`update_time`: time.Now().Unix(),
	}).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	// 级联软删除模型，避免前端还读到孤儿数据
	_, _ = common.DbMain.Client.QuickUpdate(`tbl_ai_model`, map[string]any{
		`provider_id`: dataMap[`id`],
		`status`:      1,
	}, map[string]any{
		`status`:      0,
		`update_time`: time.Now().Unix(),
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetAiModelList 查询 AI 模型配置列表（含服务商信息）
func SetAiModelList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	sql := `select m.id,m.provider_id,m.name,m.model,ifnull(m.model_type,'llm') as model_type,m.status,m.create_time,m.update_time,
p.name as provider_name,p.provider_type as request_format,p.provider_type,p.base_url,p.api_key 
from tbl_ai_model m 
left join tbl_ai_provider p on p.id = m.provider_id 
where m.status = 1 and p.status = 1`
	paramList := make([]any, 0)
	providerId := cast.ToInt(dataMap[`provider_id`])
	if providerId > 0 {
		sql += ` and m.provider_id = ?`
		paramList = append(paramList, providerId)
	}
	modelType := strings.TrimSpace(strings.ToLower(cast.ToString(dataMap[`model_type`])))
	if modelType != `` {
		if normalizeAiModelType(modelType) == `` {
			gsgin.GinResponseError(c, `模型类型仅支持 llm 或 embedding`, nil)
			return
		}
		sql += ` and ifnull(m.model_type,'llm') = ?`
		paramList = append(paramList, modelType)
	}
	sql += ` order by m.id desc`
	list, err := common.DbMain.Client.QueryBySql(sql, paramList...).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// SetAiModelAdd 新增或更新 AI 模型配置
func SetAiModelAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`provider_id`, `name`, `model`})
	if cast.ToInt(updateData[`provider_id`]) == 0 {
		gsgin.GinResponseError(c, `请选择服务商`, nil)
		return
	}
	if cast.ToString(updateData[`model`]) == `` {
		gsgin.GinResponseError(c, `模型标识不能为空`, nil)
		return
	}
	if cast.ToString(updateData[`name`]) == `` {
		updateData[`name`] = cast.ToString(updateData[`model`])
	}
	modelType := normalizeAiModelType(cast.ToString(dataMap[`model_type`]))
	if modelType == `` {
		gsgin.GinResponseError(c, `模型类型仅支持 llm 或 embedding`, nil)
		return
	}
	updateData[`model_type`] = modelType
	providerInfo, err := common.DbMain.Client.QuickQuery(`tbl_ai_provider`, `id`, map[string]any{
		`id`:     updateData[`provider_id`],
		`status`: 1,
	}).One()
	if err != nil || len(providerInfo) == 0 {
		gsgin.GinResponseError(c, `服务商不存在或已删除`, nil)
		return
	}
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`status`] = 1
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, err := common.DbMain.Client.QuickCreate(`tbl_ai_model`, updateData).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, err := common.DbMain.Client.QuickUpdate(`tbl_ai_model`, map[string]any{
			`id`: dataMap[`id`],
		}, updateData).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetAiModelDelete 删除 AI 模型配置（软删除）
func SetAiModelDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	_, err := common.DbMain.Client.QuickUpdate(`tbl_ai_model`, map[string]any{
		`id`: dataMap[`id`],
	}, map[string]any{
		`status`:      0,
		`update_time`: time.Now().Unix(),
	}).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}
