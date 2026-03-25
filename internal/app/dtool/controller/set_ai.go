package controller

import (
	"bytes"
	"dev_tool/internal/app/dtool/common"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

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
		gsgin.GinResponseError(c, `请求格式仅支持 openai`, nil)
		return
	}
	updateData[`base_url`] = normalizeAiProviderBaseURL(cast.ToString(updateData[`base_url`]))
	if cast.ToString(updateData[`base_url`]) == `` {
		gsgin.GinResponseError(c, `基础域名不能为空`, nil)
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
	sql := `select m.*,p.name as provider_name,p.provider_type as request_format,p.provider_type,p.base_url,p.api_key
from tbl_ai_model m
left join tbl_ai_provider p on p.id = m.provider_id
where m.status = 1 and p.status = 1`
	paramList := make([]any, 0)
	providerID := cast.ToInt(dataMap[`provider_id`])
	if providerID > 0 {
		sql += ` and m.provider_id = ?`
		paramList = append(paramList, providerID)
	}
	rawModelType := strings.TrimSpace(cast.ToString(dataMap[`model_type`]))
	if rawModelType != `` {
		sql += ` and m.model_type = ?`
		paramList = append(paramList, normalizeAiModelType(rawModelType))
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
	updateData := gstool.MapTakeKeys(&dataMap, []string{`provider_id`, `name`, `model`, `uri`, `model_type`})
	if cast.ToInt(updateData[`provider_id`]) == 0 {
		gsgin.GinResponseError(c, `请选择服务商`, nil)
		return
	}
	if cast.ToString(updateData[`model`]) == `` {
		gsgin.GinResponseError(c, `模型标识不能为空`, nil)
		return
	}
	updateData[`uri`] = normalizeAiModelURI(cast.ToString(updateData[`uri`]))
	if cast.ToString(updateData[`uri`]) == `` {
		gsgin.GinResponseError(c, `模型 URI 不能为空`, nil)
		return
	}
	updateData[`model_type`] = normalizeAiModelType(cast.ToString(updateData[`model_type`]))
	if cast.ToString(updateData[`name`]) == `` {
		updateData[`name`] = cast.ToString(updateData[`model`])
	}
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

// SetAiModelTest 测试 AI 模型连通性
func SetAiModelTest(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	modelID := cast.ToInt(dataMap[`id`])
	if modelID == 0 {
		gsgin.GinResponseError(c, `模型 id 不能为空`, nil)
		return
	}
	modelInfo, err := common.DbMain.InfoCrawlAiModelInfo(modelID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	requestURL := strings.TrimRight(strings.TrimSpace(cast.ToString(modelInfo[`base_url`])), `/`) +
		normalizeAiModelURI(cast.ToString(modelInfo[`uri`]))
	if requestURL == `` {
		gsgin.GinResponseError(c, `模型请求地址不能为空`, nil)
		return
	}
	apiKey := strings.TrimSpace(cast.ToString(modelInfo[`api_key`]))
	if apiKey == `` {
		gsgin.GinResponseError(c, `API Key 不能为空`, nil)
		return
	}
	method, bodyMap, err := buildAiModelConnectivityRequest(modelInfo)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(method, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	request.Header.Set(`Authorization`, `Bearer `+apiKey)
	request.Header.Set(`Content-Type`, `application/json`)
	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		gsgin.GinResponseError(c, `连通失败: `+err.Error(), nil)
		return
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(response.Body)
	if response.StatusCode >= 300 {
		gsgin.GinResponseError(c, `连通失败: HTTP `+cast.ToString(response.StatusCode)+` `+truncateAiTestResponse(responseBody), nil)
		return
	}
	gsgin.GinResponseSuccess(c, `连通成功`, map[string]any{
		`status_code`: response.StatusCode,
		`message`:     `连通成功`,
	})
}

func normalizeAiProviderBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == `` {
		return ``
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == `` || parsed.Host == `` {
		return strings.TrimRight(raw, `/`)
	}
	return strings.TrimRight(parsed.Scheme+`://`+parsed.Host, `/`)
}

func normalizeAiModelURI(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == `` {
		return ``
	}
	if !strings.HasPrefix(raw, `/`) {
		raw = `/` + raw
	}
	return raw
}

func normalizeAiModelType(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	switch raw {
	case `embedding`:
		return `embedding`
	case `llm`, ``:
		return `llm`
	default:
		return `llm`
	}
}

func buildAiModelConnectivityRequest(modelInfo map[string]any) (string, map[string]any, error) {
	modelType := strings.ToLower(strings.TrimSpace(cast.ToString(modelInfo[`model_type`])))
	if modelType == `` {
		modelType = `llm`
	}
	modelName := strings.TrimSpace(cast.ToString(modelInfo[`model`]))
	if modelName == `` {
		return ``, nil, errors.New(`模型标识不能为空`)
	}
	switch modelType {
	case `llm`:
		return http.MethodPost, map[string]any{
			`model`: modelName,
			`messages`: []map[string]string{
				{
					`role`:    `user`,
					`content`: `ping`,
				},
			},
		}, nil
	case `embedding`:
		return http.MethodPost, map[string]any{
			`model`: modelName,
			`input`: `ping`,
		}, nil
	default:
		return ``, nil, errors.New(`不支持的模型类型: ` + modelType)
	}
}

func truncateAiTestResponse(responseBody []byte) string {
	text := strings.TrimSpace(string(responseBody))
	if text == `` {
		return `empty response`
	}
	runes := []rune(text)
	if len(runes) > 180 {
		return string(runes[:180]) + `...`
	}
	return text
}
