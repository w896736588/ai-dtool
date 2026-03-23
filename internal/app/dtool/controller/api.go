package controller

import (
	"dev_tool/internal/app/dtool/api"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_curl"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// parseApiIDs 解析接口 ID，兼容数组和逗号分隔字符串。
func parseApiIDs(raw any) []int {
	result := make([]int, 0)
	exists := make(map[int]struct{})
	appendID := func(id int) {
		if id <= 0 {
			return
		}
		if _, ok := exists[id]; ok {
			return
		}
		exists[id] = struct{}{}
		result = append(result, id)
	}
	switch value := raw.(type) {
	case []any:
		for _, item := range value {
			appendID(cast.ToInt(item))
		}
	case []int:
		for _, item := range value {
			appendID(item)
		}
	case []string:
		for _, item := range value {
			appendID(cast.ToInt(strings.TrimSpace(item)))
		}
	case string:
		for _, item := range strings.Split(value, ",") {
			appendID(cast.ToInt(strings.TrimSpace(item)))
		}
	}
	return result
}

// buildCollectionBasicInfo 构建集合基础信息。
func buildCollectionBasicInfo(item map[string]any) map[string]any {
	return map[string]any{
		`id`:          item[`id`],
		`name`:        item[`name`],
		`child_count`: cast.ToInt(item[`child_count`]),
		`create_time`: item[`create_time`],
		`update_time`: item[`update_time`],
		`type`:        define.ApiTypeCollection,
		`uniqueid`:    fmt.Sprintf(`collection%d`, cast.ToInt(item[`id`])),
	}
}

// buildFolderBasicInfo 构建文件夹基础信息。
func buildFolderBasicInfo(item map[string]any) map[string]any {
	return map[string]any{
		`id`:            item[`id`],
		`collection_id`: item[`collection_id`],
		`name`:          item[`name`],
		`headers`:       cast.ToString(item[`headers`]),
		`child_count`:   cast.ToInt(item[`child_count`]),
		`create_time`:   item[`create_time`],
		`update_time`:   item[`update_time`],
		`type`:          define.ApiTypeFolder,
		`uniqueid`:      fmt.Sprintf(`folder%d`, cast.ToInt(item[`id`])),
	}
}

// buildApiBasicInfo 构建接口基础信息，不返回请求明细字段。
func buildApiBasicInfo(item map[string]any) map[string]any {
	return map[string]any{
		`id`:            item[`id`],
		`folder_id`:     item[`folder_id`],
		`collection_id`: item[`collection_id`],
		`name`:          item[`name`],
		`method`:        item[`method`],
		`url`:           item[`url`],
		`desc`:          item[`desc`],
		`env_id`:        item[`env_id`],
		`weight`:        item[`weight`],
		`create_time`:   item[`create_time`],
		`update_time`:   item[`update_time`],
		`type`:          define.ApiTypeApi,
		`uniqueid`:      fmt.Sprintf(`api%d`, cast.ToInt(item[`id`])),
	}
}

// sortAPIListByIDs 按传入 ID 顺序重排接口列表。
func sortAPIListByIDs(list []map[string]any, ids []int) []map[string]any {
	itemMap := make(map[int]map[string]any, len(list))
	for _, item := range list {
		itemMap[cast.ToInt(item[`id`])] = item
	}
	result := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		if item, ok := itemMap[id]; ok {
			result = append(result, item)
		}
	}
	return result
}

func ApiCreateCollection(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := common.DbMain.Client.QuickCreate(`tbl_api_collection`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_api_collection`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := common.DbMain.Client.QuickQuery(`tbl_api_collection`, `*`, map[string]any{
		`id`: id,
	}).One()
	if len(info) > 0 {
		info[`type`] = define.ApiTypeCollection
		info[`uniqueid`] = fmt.Sprintf(`collection%d`, info[`id`])
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

func ApiDeleteCollection(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `请选择集合`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_api_collection`,
			map[string]any{
				`id`: dataMap[`id`],
			}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func ApiDeleteApi(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `请选择集合`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_api`,
			map[string]any{
				`id`: dataMap[`id`],
			}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func ApiDeleteDir(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `请选择集合`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_api_dir`,
			map[string]any{
				`id`: dataMap[`id`],
			}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func ApiCreateCollectionEnv(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `collection_id`, `desc`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := common.DbMain.Client.QuickCreate(`tbl_api_env`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_api_env`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := common.DbMain.Client.QuickQuery(`tbl_api_env`, `*`, map[string]any{
		`id`: id,
	}).One()
	gsgin.GinResponseSuccess(c, ``, info)
}

func ApiCollections(c *gin.Context) {
	list, _ := common.DbMain.Client.QueryBySql(`select * from tbl_api_collection order by id asc`).All()
	for _, item := range list {
		item[`type`] = `collection`
		//child
		dirs, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `*`, map[string]any{
			`collection_id`: item[`id`],
		}).Order(`id asc`).All()
		for _, child := range dirs {
			child[`type`] = `folder`
			child[`uniqueid`] = fmt.Sprintf(`folder%d`, child[`id`])
			//查找接口
			apis, _ := common.DbMain.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
				`folder_id`: child[`id`],
			}).Order(`weight,id asc`).All()
			for _, api := range apis {
				api[`type`] = `api`
				api[`uniqueid`] = fmt.Sprintf(`api%d`, api[`id`])
			}
			child[`children`] = apis
		}
		item[`uniqueid`] = fmt.Sprintf(`collection%d`, item[`id`])
		item[`children`] = dirs
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// ApiCollectionListBasic 查询所有集合基础信息。
func ApiCollectionListBasic(c *gin.Context) {
	list, _ := common.DbMain.Client.QueryBySql(`
select c.id,
       c.name,
       c.create_time,
       c.update_time,
       count(d.id) as child_count
from tbl_api_collection c
left join tbl_api_dir d on d.collection_id = c.id
group by c.id, c.name, c.create_time, c.update_time
order by c.id asc`).All()
	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		result = append(result, buildCollectionBasicInfo(item))
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: result,
	})
}

// ApiCollectionFoldersBasic 按集合查询文件夹基础信息。
func ApiCollectionFoldersBasic(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	collectionId := cast.ToInt(dataMap[`collection_id`])
	if collectionId <= 0 {
		gsgin.GinResponseError(c, `请选择集合`, nil)
		return
	}
	list, _ := common.DbMain.Client.QueryBySql(`
select d.id,
       d.collection_id,
       d.name,
       d.headers,
       d.create_time,
       d.update_time,
       count(a.id) as child_count
from tbl_api_dir d
left join tbl_api a on a.folder_id = d.id
where d.collection_id = ?
group by d.id, d.collection_id, d.name, d.headers, d.create_time, d.update_time
order by d.id asc`, collectionId).All()
	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		result = append(result, buildFolderBasicInfo(item))
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: result,
	})
}

// ApiFolderApisBasic 按文件夹查询接口基础信息。
func ApiFolderApisBasic(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	folderId := cast.ToInt(dataMap[`folder_id`])
	if folderId <= 0 {
		gsgin.GinResponseError(c, `请选择文件夹`, nil)
		return
	}
	list, _ := common.DbMain.Client.QueryBySql(`select id,folder_id,collection_id,name,method,url,desc,env_id,weight,create_time,update_time from tbl_api where folder_id = ? order by weight,id asc`, folderId).All()
	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		result = append(result, buildApiBasicInfo(item))
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: result,
	})
}

// ApiApisDetailByIds 按若干接口 ID 查询接口明细。
func ApiApisDetailByIds(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	ids := parseApiIDs(dataMap[`ids`])
	if len(ids) == 0 {
		gsgin.GinResponseError(c, `请选择接口`, nil)
		return
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, `?`)
		args = append(args, id)
	}
	sql := `select * from tbl_api where id in (` + strings.Join(placeholders, `,`) + `)`
	list, _ := common.DbMain.Client.QueryBySql(sql, args...).All()
	for _, item := range list {
		item[`type`] = define.ApiTypeApi
		item[`uniqueid`] = fmt.Sprintf(`api%d`, cast.ToInt(item[`id`]))
	}
	list = sortAPIListByIDs(list, ids)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

func ApiCollectionEnvs(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	collectionId := dataMap[`collection_id`]
	if cast.ToInt(collectionId) == 0 {
		gsgin.GinResponseError(c, `请选择集合`, nil)
		return
	}
	list, _ := common.DbMain.Client.QueryBySql(`select * from tbl_api_env where collection_id = ? order by id asc`, collectionId).All()
	//查找每一个的环境变量
	for _, item := range list {
		item[`variables`] = []map[string]any{}
		envItems, _ := common.DbMain.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
			`env_id`: item[`id`],
		}).All()
		for _, envItem := range envItems {
			envItem[`type`] = `env_item`
		}
		item[`variables`] = envItems
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

func ApiCollectionEnvItems(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	collectionId := dataMap[`collection_id`]
	envId := dataMap[`env_id`]
	if cast.ToInt(envId) == 0 {
		gsgin.GinResponseSuccess(c, ``, map[string]any{
			`list`: []map[string]any{},
		})
		return
	}
	if cast.ToInt(collectionId) == 0 {
		gsgin.GinResponseError(c, `请选择集合`, nil)
		return
	}
	list, _ := common.DbMain.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
		`collection_id`: collectionId,
		`env_id`:        envId,
	}).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

func ApiCreateCollectionEnvItem(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var id any
	if cast.ToInt(dataMap[`env_id`]) == 0 || cast.ToInt(dataMap[`collection_id`]) == 0 {
		gsgin.GinResponseError(c, `请选择集合和环境`, nil)
		return
	}
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `collection_id`, `env_id`, `desc`, `key`, `value`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := common.DbMain.Client.QuickCreate(`tbl_api_env_item`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_api_env_item`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := common.DbMain.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
		`id`: id,
	}).One()
	gsgin.GinResponseSuccess(c, ``, info)
}

func Apis(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	collectionId := dataMap[`collection_id`]
	dirId := dataMap[`dir_id`]
	sql := `select * from tbl_api where collection_id = ? and folder_id = ? order by weight,id asc`
	list, _ := common.DbMain.Client.QueryBySql(sql, collectionId, dirId).All()
	for _, item := range list {
		item[`type`] = `api`
		item[`uniqueid`] = fmt.Sprintf(`api%d`, item[`id`])
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

func ApiCreateDir(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `collection_id`, `headers`})
	var err error
	for key, value := range updateData {
		if gstool.ArrayExistValue(&[]string{reflect.Array.String(), reflect.Map.String(), reflect.Slice.String()}, gstool.ReflectGetType(value).String()) {
			updateData[key] = gstool.JsonEncode(value)
		}
	}
	if cast.ToString(updateData[`headers`]) == `` {
		updateData[`headers`] = `{}`
	}
	// 中文注释：统一清理并校验文件夹默认请求头，避免保存非法 JSON。
	updateData[`headers`], err = filterEmptyMap(cast.ToString(updateData[`headers`]), `headers格式错误`, 500)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), ``)
		return
	}
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := common.DbMain.Client.QuickCreate(`tbl_api_dir`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_api_dir`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `*`, map[string]any{
		`id`: id,
	}).One()
	if len(info) > 0 {
		info[`type`] = define.ApiTypeFolder
		info[`uniqueid`] = fmt.Sprintf(`folder%d`, info[`id`])
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

func ApiCreateApi(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var id any
	curlData := cast.ToString(dataMap[`curlData`])
	var updateData map[string]any
	if curlData != `` {
		parsed := p_curl.NewParseCurl(curlData)
		err := parsed.Parse()
		if err != nil {
			gsgin.GinResponseError(c, `Curl解析失败 `+err.Error(), nil)
			return
		}
		dataMap[`name`] = `从Curl导入`
		dataMap[`method`] = parsed.CurlStruct.Method
		dataMap[`query_params`] = parsed.CurlStruct.QueryParams
		dataMap[`protocol`] = parsed.CurlStruct.Protocol
		if strings.ToLower(parsed.CurlStruct.Protocol) == `http` {
			dataMap[`url`] = `http://` + parsed.CurlStruct.Url
		} else {
			dataMap[`url`] = `https://` + parsed.CurlStruct.Url
		}
		dataMap[`headers`] = parsed.CurlStruct.Headers
		dataMap[`content_type`] = parsed.CurlStruct.ContentType
		dataMap[`body_form`] = parsed.CurlStruct.BodyForm
		dataMap[`body_json`] = parsed.CurlStruct.BodyJson

	}
	updateData = gstool.MapTakeKeys(&dataMap, []string{`folder_id`, `collection_id`, `name`, `method`, `url`,
		`protocol`, `desc`, `headers`, `query_params`, `content_type`, `body_form`, `body_json`,
		`env_id`, `response_take`, `take_result`, `take_result_desc`})
	for key, value := range updateData {
		if gstool.ArrayExistValue(&[]string{reflect.Array.String(), reflect.Map.String(), reflect.Slice.String()}, gstool.ReflectGetType(value).String()) {
			updateData[key] = gstool.JsonEncode(value)
		}
	}
	var err error
	//处理请求参数空值
	updateData[`query_params`], err = filterEmptyArrayMap(cast.ToString(updateData[`query_params`]), `field`, `请求参数格式错误`, 500)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), ``)
		return
	}
	//处理headers参数空值
	updateData[`headers`], err = filterEmptyMap(cast.ToString(updateData[`headers`]), `headers格式错误`, 500)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), ``)
		return
	}
	//处理请求参数空值
	updateData[`body_form`], err = filterEmptyArrayMap(cast.ToString(updateData[`body_form`]), `field`, `请求体格式错误`, 500)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), ``)
		return
	}
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := common.DbMain.Client.QuickCreate(`tbl_api`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_api`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := common.DbMain.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
		`id`: id,
	}).One()
	if len(info) > 0 {
		info[`type`] = define.ApiTypeApi
		info[`uniqueid`] = fmt.Sprintf(`api%d`, info[`id`])
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

func filterEmptyArrayMap(queryParams, fieldKey, errmsg string, max int) (string, error) {
	queryParamsData := make([]map[string]any, 0)
	queryParamsDataNew := make([]map[string]any, 0)
	dErr := gstool.JsonDecode(queryParams, &queryParamsData)
	if dErr != nil {
		return ``, errors.New(errmsg)
	}
	for _, item := range queryParamsData {
		if cast.ToString(item[fieldKey]) != `` {
			queryParamsDataNew = append(queryParamsDataNew, item)
		}
	}
	if len(queryParamsDataNew) > max {
		return ``, errors.New(errmsg + `,最多` + cast.ToString(max) + `条`)
	}
	return gstool.JsonEncode(queryParamsDataNew), nil
}

func filterEmptyMap(queryParams, errmsg string, max int) (string, error) {
	queryParamsData := make(map[string]any)
	queryParamsDataNew := make(map[string]any)
	dErr := gstool.JsonDecode(queryParams, &queryParamsData)
	if dErr != nil {
		return ``, errors.New(errmsg)
	}
	for key, item := range queryParamsData {
		if key != `` {
			queryParamsDataNew[key] = item
		}
	}
	if len(queryParamsDataNew) > max {
		return ``, errors.New(errmsg + `,最多` + cast.ToString(max) + `条`)
	}
	return gstool.JsonEncode(queryParamsDataNew), nil
}

func ApiRun(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	id := dataMap[`id`]
	apiInfo, _ := common.DbMain.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
		`id`: id,
	}).One()
	if len(apiInfo) == 0 {
		gsgin.GinResponseError(c, `api不存在`, nil)
		return
	}
	// 中文注释：运行前加载所属目录默认请求头，用于与接口请求头做覆盖合并。
	folderInfo, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `headers`, map[string]any{
		`id`: apiInfo[`folder_id`],
	}).One()
	if len(folderInfo) > 0 {
		apiInfo[`folder_headers`] = folderInfo[`headers`]
	}
	apiCli := api.NewApi(apiInfo)
	err := apiCli.Run()
	if err != nil {
		gsgin.GinResponseError(c, `运行失败 `+err.Error(), nil)
		return
	}
	apiCli.ResponseTake()
	_, _ = common.DbMain.Client.QuickUpdate(`tbl_api`, map[string]any{
		`id`: id,
	}, map[string]any{
		`last_result`: gstool.JsonEncode(apiCli.Result),
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, apiCli.Result)
}

func ApiCode(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	id := dataMap[`id`]
	apiInfo, _ := common.DbMain.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
		`id`: id,
	}).One()
	if len(apiInfo) == 0 {
		gsgin.GinResponseError(c, `api不存在`, nil)
		return
	}
	codeType := dataMap[`code_type`]
	apiCli := api.NewApi(apiInfo)
	code := ``
	if codeType == `curl bash(chrome)` {
		code = apiCli.ToChromeCurlBash()
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`code`: code,
	})
	return
}

func ApiWeightDown(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	id := dataMap[`id`]
	apiInfo, err := common.DbMain.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
		`id`: id,
	}).One()
	if err != nil {
		gsgin.GinResponseError(c, `api不存在`, nil)
		return
	}
	_, _ = common.DbMain.Client.QuickUpdate(`tbl_api`, map[string]any{
		`id`: id,
	}, map[string]any{
		`weight`: cast.ToInt(apiInfo[`weight`]) + 1,
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
	return
}

func ApiTakeJsonResult(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	json := cast.ToString(dataMap[`json`])
	id := cast.ToInt(dataMap[`id`])
	list, err := gstool.JsonFlatPaths(json)
	if err != nil {
		gsgin.GinResponseError(c, `json格式错误`, nil)
		return
	}
	apiInfo, err := common.DbMain.GetApiInfo(id)
	if err != nil {
		gsgin.GinResponseError(c, `api不存在`, nil)
		return
	}
	takeResults := make([]gstool.FlatItem, 0)
	_ = gstool.JsonDecode(cast.ToString(apiInfo[`take_result`]), &takeResults)
	gstool.FmtPrintlnLogTime(`已经提取 的参数描述 %s`, gstool.JsonFormat(takeResults))
	for _, takeResult := range takeResults {
		for key, _ := range list {
			if list[key].Key == takeResult.Key {
				list[key].Desc = takeResult.Desc
			}
		}
	}
	gsgin.GinResponseSuccess(c, ``, list)
	return
}

func ApiBatchImport(c *gin.Context) {
	// 从form-data中获取json字符串
	jsonStr := c.PostForm(`json`)
	if jsonStr == `` {
		gsgin.GinResponseError(c, `json参数必须提供`, nil)
		return
	}

	// 解析JSON字符串
	importData := make(map[string]any)
	err := gstool.JsonDecode(jsonStr, &importData)
	if err != nil {
		gsgin.GinResponseError(c, `json格式错误: `+err.Error(), nil)
		return
	}
	collectionId := cast.ToInt(c.PostForm(`collection_id`))
	// 只接受collection_id，必须已存在
	importCollectionId := cast.ToInt(importData[`collection_id`])
	if collectionId == 0 {
		collectionId = importCollectionId
	}
	if collectionId == 0 {
		gsgin.GinResponseError(c, `collection_id必须提供`, nil)
		return
	}

	// 验证集合是否存在
	collectionInfo, _ := common.DbMain.Client.QuickQuery(`tbl_api_collection`, `*`, map[string]any{
		`id`: collectionId,
	}).One()
	if len(collectionInfo) == 0 {
		gsgin.GinResponseError(c, `集合不存在`, nil)
		return
	}

	// 处理导入项
	items, ok := importData[`items`].([]any)
	if !ok {
		gsgin.GinResponseError(c, `items格式错误`, nil)
		return
	}

	importResult := map[string]any{
		`folders_created`: 0,
		`folders_updated`: 0,
		`apis_created`:    0,
		`errors`:          []string{},
	}

	for _, item := range items {
		itemMap, ok := item.(map[string]any)
		if !ok {
			importResult[`errors`] = append(importResult[`errors`].([]string), `item格式错误`)
			continue
		}
		itemType := cast.ToString(itemMap[`type`])
		// 根级别只允许folder类型
		if itemType == `folder` {
			err := processFolderItem(c, collectionId, itemMap, importResult)
			if err != nil {
				importResult[`errors`] = append(importResult[`errors`].([]string), `创建文件夹失败: `+err.Error())
			}
		} else {
			importResult[`errors`] = append(importResult[`errors`].([]string), `根级别只允许folder类型，不允许: `+itemType)
		}
	}

	gsgin.GinResponseSuccess(c, `导入完成`, map[string]any{
		`result`: importResult,
	})
}

func processFolderItem(c *gin.Context, collectionId int, folderData map[string]any, importResult map[string]any) error {
	folderName := cast.ToString(folderData[`name`])
	folderHeaders := cast.ToString(folderData[`headers`])
	if folderHeaders == `` {
		folderHeaders = `{}`
	}
	// 中文注释：导入时沿用目录默认请求头清洗规则，保证与单个保存入口行为一致。
	filteredFolderHeaders, err := filterEmptyMap(folderHeaders, `headers格式错误`, 500)
	if err != nil {
		return err
	}

	// 检查同名文件夹是否已存在
	existingFolder, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `id`, map[string]any{
		`collection_id`: collectionId,
		`name`:          folderName,
	}).One()

	var folderId int
	if len(existingFolder) > 0 {
		// 文件夹已存在，删除该文件夹下的所有接口（覆盖式更新）
		folderId = cast.ToInt(existingFolder[`id`])
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_api_dir`, map[string]any{
			`id`: folderId,
		}, map[string]any{
			`headers`:     filteredFolderHeaders,
			`update_time`: time.Now().Unix(),
		}).Exec()
		_, _ = common.DbMain.Client.QuickDelete(`tbl_api`, map[string]any{
			`folder_id`: folderId,
		}).Exec()
		importResult[`folders_updated`] = importResult[`folders_updated`].(int) + 1
	} else {
		// 文件夹不存在，创建新文件夹
		updateData := map[string]any{
			`name`:          folderName,
			`collection_id`: collectionId,
			`headers`:       filteredFolderHeaders,
			`create_time`:   time.Now().Unix(),
			`update_time`:   time.Now().Unix(),
		}
		newId, createErr := common.DbMain.Client.QuickCreate(`tbl_api_dir`, updateData).Exec()
		if createErr != nil {
			return createErr
		}
		folderId = cast.ToInt(newId)
		importResult[`folders_created`] = importResult[`folders_created`].(int) + 1
	}

	// 处理子项 - 文件夹下只能包含api
	children, ok := folderData[`children`].([]any)
	if ok {
		for _, child := range children {
			childMap, ok := child.(map[string]any)
			if !ok {
				importResult[`errors`] = append(importResult[`errors`].([]string), `子项格式错误`)
				continue
			}
			itemType := cast.ToString(childMap[`type`])
			// 文件夹下只允许api类型，不允许嵌套folder
			if itemType == `api` {
				err := processApiItem(c, collectionId, folderId, childMap, importResult)
				if err != nil {
					importResult[`errors`] = append(importResult[`errors`].([]string), `创建接口失败: `+err.Error())
				}
			} else {
				importResult[`errors`] = append(importResult[`errors`].([]string), `文件夹下只允许api类型，不允许: `+itemType)
			}
		}
	}
	return nil
}

func processApiItem(c *gin.Context, collectionId, folderId int, apiData map[string]any, importResult map[string]any) error {
	var err error
	// 支持从curl导入
	curlData := cast.ToString(apiData[`curlData`])
	var updateData map[string]any

	if curlData != `` {
		// 从curl解析
		parsed := p_curl.NewParseCurl(curlData)
		err = parsed.Parse()
		if err != nil {
			return errors.New(`Curl解析失败: ` + err.Error())
		}
		updateData = map[string]any{
			`folder_id`:     folderId,
			`collection_id`: collectionId,
			`name`:          cast.ToString(apiData[`name`]),
			`method`:        parsed.CurlStruct.Method,
			`query_params`:  parsed.CurlStruct.QueryParams,
			`protocol`:      parsed.CurlStruct.Protocol,
			`url`:           parsed.CurlStruct.Url,
			`headers`:       parsed.CurlStruct.Headers,
			`content_type`:  parsed.CurlStruct.ContentType,
			`body_form`:     parsed.CurlStruct.BodyForm,
			`body_json`:     parsed.CurlStruct.BodyJson,
			`env_id`:        cast.ToInt(apiData[`env_id`]),
			`desc`:          cast.ToString(apiData[`desc`]),
		}
		if strings.ToLower(parsed.CurlStruct.Protocol) == `http` {
			updateData[`url`] = `http://` + parsed.CurlStruct.Url
		} else {
			updateData[`url`] = `https://` + parsed.CurlStruct.Url
		}
	} else {
		// 直接从JSON数据创建
		updateData = gstool.MapTakeKeys(&apiData, []string{
			`folder_id`, `collection_id`, `name`, `method`, `url`,
			`protocol`, `desc`, `headers`, `query_params`, `content_type`, `body_form`, `body_json`,
			`env_id`, `response_take`, `take_result`, `take_result_desc`,
		})
		updateData[`folder_id`] = folderId
		updateData[`collection_id`] = collectionId
	}

	// 处理数组/对象类型的字段，转换为JSON字符串
	for key, value := range updateData {
		if gstool.ArrayExistValue(&[]string{reflect.Array.String(), reflect.Map.String(), reflect.Slice.String()}, gstool.ReflectGetType(value).String()) {
			updateData[key] = gstool.JsonEncode(value)
		}
	}

	// 处理空值过滤 - 只处理存在的字段
	if queryParams, exists := updateData[`query_params`]; exists && queryParams != nil && cast.ToString(queryParams) != `` {
		updateData[`query_params`], err = filterEmptyArrayMap(cast.ToString(queryParams), `field`, `请求参数格式错误`, 500)
		if err != nil {
			return err
		}
	} else {
		// 如果query_params不存在或为空，设置为空数组字符串
		updateData[`query_params`] = `[]`
	}

	if headers, exists := updateData[`headers`]; exists && headers != nil && cast.ToString(headers) != `` {
		updateData[`headers`], err = filterEmptyMap(cast.ToString(headers), `headers格式错误`, 500)
		if err != nil {
			return err
		}
	} else {
		// 如果headers不存在或为空，设置为空对象字符串
		updateData[`headers`] = `{}`
	}

	if bodyForm, exists := updateData[`body_form`]; exists && bodyForm != nil && cast.ToString(bodyForm) != `` {
		updateData[`body_form`], err = filterEmptyArrayMap(cast.ToString(bodyForm), `field`, `请求体格式错误`, 500)
		if err != nil {
			return err
		}
	} else {
		// 如果body_form不存在或为空，设置为空数组字符串
		updateData[`body_form`] = `[]`
	}

	// 处理env_id：检查环境是否存在且属于当前集合
	envId := cast.ToInt(updateData[`env_id`])
	if envId > 0 {
		// 检查环境是否存在且属于当前集合
		envInfo, _ := common.DbMain.Client.QuickQuery(`tbl_api_env`, `id`, map[string]any{
			`id`:            envId,
			`collection_id`: collectionId,
		}).One()
		if len(envInfo) == 0 {
			// 环境不存在或不属于当前集合，删除env_id字段
			delete(updateData, `env_id`)
		}
	} else {
		// env_id为0，删除该字段
		delete(updateData, `env_id`)
	}

	// 创建接口
	updateData[`create_time`] = time.Now().Unix()
	updateData[`update_time`] = time.Now().Unix()
	_, err = common.DbMain.Client.QuickCreate(`tbl_api`, updateData).Exec()
	if err != nil {
		return errors.New(`创建接口失败: ` + err.Error())
	}

	importResult[`apis_created`] = importResult[`apis_created`].(int) + 1
	return nil
}

func ApiMoveApi(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	apiId := cast.ToInt(dataMap[`api_id`])
	newFolderId := cast.ToInt(dataMap[`folder_id`])

	if apiId == 0 {
		gsgin.GinResponseError(c, `请选择要移动的接口`, nil)
		return
	}

	if newFolderId == 0 {
		gsgin.GinResponseError(c, `请选择目标文件夹`, nil)
		return
	}

	// Check if api exists
	apiInfo, _ := common.DbMain.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
		`id`: apiId,
	}).One()
	if len(apiInfo) == 0 {
		gsgin.GinResponseError(c, `接口不存在`, nil)
		return
	}

	// Check if target folder exists
	folderInfo, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `*`, map[string]any{
		`id`: newFolderId,
	}).One()
	if len(folderInfo) == 0 {
		gsgin.GinResponseError(c, `目标文件夹不存在`, nil)
		return
	}

	// Check if target folder belongs to the same collection
	apiCollectionId := cast.ToInt(apiInfo[`collection_id`])
	folderCollectionId := cast.ToInt(folderInfo[`collection_id`])
	if apiCollectionId != folderCollectionId {
		gsgin.GinResponseError(c, `不能移动到不同集合的文件夹`, nil)
		return
	}

	// Update api folder_id
	_, err := common.DbMain.Client.QuickUpdate(`tbl_api`, map[string]any{
		`id`: apiId,
	}, map[string]any{
		`folder_id`:   newFolderId,
		`update_time`: time.Now().Unix(),
	}).Exec()

	if err != nil {
		gsgin.GinResponseError(c, `移动失败: `+err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, `移动成功`, nil)
}

func ApiFolderDetail(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	dir, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `*`, map[string]any{
		`id`: dataMap[`dir_id`],
	}).One()
	if len(dir) == 0 {
		gsgin.GinResponseError(c, "目录不存在", nil)
		return
	}
	dir[`type`] = `folder`
	dir[`uniqueid`] = fmt.Sprintf(`folder%d`, dir[`id`])
	//查找接口
	apis, _ := common.DbMain.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
		`folder_id`: dir[`id`],
	}).Order(`weight,id asc`).All()
	for _, api := range apis {
		api[`type`] = `api`
		api[`uniqueid`] = fmt.Sprintf(`api%d`, api[`id`])
		// If env_id exists, fetch environment variables and replace URL
		envId := cast.ToInt(api[`env_id`])
		if envId > 0 {
			// Query all environment variables for this env_id
			envItems, _ := common.DbMain.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
				`env_id`: envId,
			}).All()
			// Replace variables in URL
			url := cast.ToString(api[`url`])
			for _, envItem := range envItems {
				key := cast.ToString(envItem[`key`])
				value := cast.ToString(envItem[`value`])
				if key != `` && value != `` {
					url = strings.ReplaceAll(url, `$`+key+`$`, value)
				}
			}
			api[`url`] = url
		}
	}
	dir[`children`] = apis
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`dir`: dir,
	})
}
