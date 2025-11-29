package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"dev_tool/internal/pkg/p_api"
	"errors"
	"fmt"
	"reflect"
	"time"

	"gitee.com/Sxiaobai/gs/gstool"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func ApiCreateCollection(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_api_collection`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_api_collection`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api_collection`, `*`, map[string]any{
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_api_collection`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_api`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_api_dir`,
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
		newId, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_api_env`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_api_env`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api_env`, `*`, map[string]any{
		`id`: id,
	}).One()
	gsgin.GinResponseSuccess(c, ``, info)
}

func ApiCollections(c *gin.Context) {
	list, _ := base.Component.TSqlite.Client.QueryBySql(`select * from tbl_api_collection order by id asc`).All()
	for _, item := range list {
		item[`type`] = `collection`
		//child
		dirs, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api_dir`, `*`, map[string]any{
			`collection_id`: item[`id`],
		}).Order(`id asc`).All()
		for _, child := range dirs {
			child[`type`] = `folder`
			child[`uniqueid`] = fmt.Sprintf(`folder%d`, child[`id`])
			//查找接口
			apis, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
				`folder_id`: child[`id`],
			}).Order(`id asc`).All()
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

func ApiCollectionEnvs(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	collectionId := dataMap[`collection_id`]
	if cast.ToInt(collectionId) == 0 {
		gsgin.GinResponseError(c, `请选择集合`, nil)
		return
	}
	list, _ := base.Component.TSqlite.Client.QueryBySql(`select * from tbl_api_env where collection_id = ? order by id asc`, collectionId).All()
	//查找每一个的环境变量
	for _, item := range list {
		item[`variables`] = []map[string]any{}
		envItems, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
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
	list, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
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
		newId, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_api_env_item`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_api_env_item`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
		`id`: id,
	}).One()
	gsgin.GinResponseSuccess(c, ``, info)
}

func Apis(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	collectionId := dataMap[`collection_id`]
	dirId := dataMap[`dir_id`]
	sql := `select * from tbl_api where collection_id = ? and folder_id = ? order by id asc`
	list, _ := base.Component.TSqlite.Client.QueryBySql(sql, collectionId, dirId).All()
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
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `collection_id`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_api_dir`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_api_dir`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api_dir`, `*`, map[string]any{
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
		parsed := p_api.NewParseCurl(curlData)
		err := parsed.Parse()
		if err != nil {
			gsgin.GinResponseError(c, `Curl解析失败 `+err.Error(), nil)
			return
		}
		dataMap[`name`] = `从Curl导入`
		dataMap[`method`] = parsed.CurlStruct.Method
		dataMap[`query_params`] = parsed.CurlStruct.QueryParams
		dataMap[`protocol`] = parsed.CurlStruct.Protocol
		dataMap[`url`] = parsed.CurlStruct.Url
		dataMap[`headers`] = parsed.CurlStruct.Headers
		dataMap[`content_type`] = parsed.CurlStruct.ContentType
		dataMap[`body_form`] = parsed.CurlStruct.BodyForm
		dataMap[`body_json`] = parsed.CurlStruct.Body

	}
	updateData = gstool.MapTakeKeys(&dataMap, []string{`folder_id`, `collection_id`, `name`, `method`, `url`,
		`protocol`, `desc`, `headers`, `query_params`, `content_type`, `body_form`, `body_json`, `env_id`, `response_take`})
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
		newId, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_api`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_api`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
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
	apiInfo, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
		`id`: id,
	}).One()
	if len(apiInfo) == 0 {
		gsgin.GinResponseError(c, `api不存在`, nil)
		return
	}
	gstool.FmtPrintlnLogTime(`%s`, gstool.JsonFormat(apiInfo))
	apiCli := p_api.NewApi(apiInfo)
	err := apiCli.Run()
	if err != nil {
		gsgin.GinResponseError(c, `运行失败 `+err.Error(), nil)
		return
	}
	apiCli.ResponseTake()
	_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_api`, map[string]any{
		`id`: id,
	}, map[string]any{
		`last_result`: gstool.JsonEncode(apiCli.Result),
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, apiCli.Result)
}
