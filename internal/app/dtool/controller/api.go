package controller

import (
	"dev_tool/internal/app/dtool/api"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_curl"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
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
		`env_id`:        cast.ToInt(item[`env_id`]),
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

// buildFolderCleanupInfo 构建清理候选文件夹的返回结构。
func buildFolderCleanupInfo(item map[string]any) map[string]any {
	return map[string]any{
		`id`:                   item[`id`],
		`collection_id`:        item[`collection_id`],
		`collection_name`:      item[`collection_name`],
		`name`:                 item[`name`],
		`child_count`:          cast.ToInt(item[`child_count`]),
		`create_time`:          cast.ToInt64(item[`create_time`]),
		`update_time`:          cast.ToInt64(item[`update_time`]),
		`last_api_update_time`: cast.ToInt64(item[`last_api_update_time`]),
		`last_run_time`:        cast.ToInt64(item[`last_run_time`]),
		`latest_active_time`:   cast.ToInt64(item[`latest_active_time`]),
		`type`:                 define.ApiTypeFolder,
		`uniqueid`:             fmt.Sprintf(`folder%d`, cast.ToInt(item[`id`])),
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
	changeType := `collection_updated`
	if cast.ToInt(dataMap[`id`]) == 0 {
		changeType = `collection_created`
	}
	go BroadcastApiChange(c.GetHeader("SseClientId"), changeType, map[string]any{
		`collection_id`: cast.ToInt(id),
	})
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
	go BroadcastApiChange(c.GetHeader("SseClientId"), `collection_deleted`, map[string]any{
		`collection_id`: cast.ToInt(dataMap[`id`]),
	})
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
	go BroadcastApiChange(c.GetHeader("SseClientId"), `api_deleted`, map[string]any{
		`collection_id`: cast.ToInt(dataMap[`collection_id`]),
		`folder_id`:     cast.ToInt(dataMap[`folder_id`]),
		`api_id`:        cast.ToInt(dataMap[`id`]),
	})
	gsgin.GinResponseSuccess(c, ``, nil)
}

func ApiDeleteDir(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	folderId := cast.ToInt(dataMap[`id`])
	collectionId := cast.ToInt(dataMap[`collection_id`])
	if folderId == 0 {
		gsgin.GinResponseError(c, `请选择文件夹`, nil)
		return
	}
	// 软删除：标记为归档，记录原始集合ID
	_, err := common.DbMain.Client.ExecBySql(
		`UPDATE tbl_api_dir SET archived = 1, original_collection_id = ?, update_time = ? WHERE id = ?`,
		collectionId, time.Now().Unix(), folderId,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, `删除失败 `+err.Error(), nil)
		return
	}
	go BroadcastApiChange(c.GetHeader("SseClientId"), `folder_archived`, map[string]any{
		`collection_id`: collectionId,
		`folder_id`:     folderId,
	})
	gsgin.GinResponseSuccess(c, ``, nil)
}

// ApiCleanupCandidateFolders 查询超过指定天数未修改且未执行的文件夹。
func ApiCleanupCandidateFolders(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	days := cast.ToInt(dataMap[`days`])
	if days <= 0 {
		days = 7
	}
	cutoffTime := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()
	list, err := common.DbMain.Client.QueryBySql(`
select d.id,
       d.collection_id,
       c.name as collection_name,
       d.name,
       d.create_time,
       d.update_time,
       count(a.id) as child_count,
       coalesce(max(a.update_time), 0) as last_api_update_time,
       coalesce(max(a.last_run_time), 0) as last_run_time,
       max(
         d.update_time,
         coalesce(max(a.update_time), 0),
         coalesce(max(a.last_run_time), 0)
       ) as latest_active_time
from tbl_api_dir d
left join tbl_api_collection c on c.id = d.collection_id
left join tbl_api a on a.folder_id = d.id
where d.archived = 0
group by d.id, d.collection_id, c.name, d.name, d.create_time, d.update_time
having count(a.id) > 0
   and max(
         d.update_time,
         coalesce(max(a.update_time), 0),
         coalesce(max(a.last_run_time), 0)
       ) < ?
order by latest_active_time asc, d.id asc`, cutoffTime).All()
	if err != nil {
		gsgin.GinResponseError(c, `查询失败 `+err.Error(), nil)
		return
	}
	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		result = append(result, buildFolderCleanupInfo(item))
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`days`: days,
		`list`: result,
	})
}

// ApiCleanupArchiveFolders 批量归档清理候选文件夹。
func ApiCleanupArchiveFolders(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	folderIDs := parseApiIDs(dataMap[`folder_ids`])
	if len(folderIDs) == 0 {
		gsgin.GinResponseError(c, `请选择文件夹`, nil)
		return
	}
	archivedIDs := make([]int, 0, len(folderIDs))
	for _, folderID := range folderIDs {
		folderInfo, _ := common.DbMain.Client.QueryBySql(
			`SELECT id, collection_id, archived FROM tbl_api_dir WHERE id = ?`, folderID,
		).One()
		if len(folderInfo) == 0 || cast.ToInt(folderInfo[`archived`]) == 1 {
			continue
		}
		collectionID := cast.ToInt(folderInfo[`collection_id`])
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_api_dir SET archived = 1, original_collection_id = ?, update_time = ? WHERE id = ?`,
			collectionID, time.Now().Unix(), folderID,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, `归档失败 `+err.Error(), nil)
			return
		}
		archivedIDs = append(archivedIDs, folderID)
		go BroadcastApiChange(c.GetHeader("SseClientId"), `folder_archived`, map[string]any{
			`collection_id`: collectionID,
			`folder_id`:     folderID,
		})
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`folder_ids`: archivedIDs,
		`count`:      len(archivedIDs),
	})
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
left join tbl_api_dir d on d.collection_id = c.id and d.archived = 0
group by c.id, c.name, c.create_time, c.update_time
order by c.id asc`).All()
	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		result = append(result, buildCollectionBasicInfo(item))
	}
	// query archive folder count
	archiveRows, _ := common.DbMain.Client.QueryBySql(`select count(*) as cnt from tbl_api_dir where archived = 1`).All()
	archiveCount := 0
	if len(archiveRows) > 0 {
		archiveCount = cast.ToInt(archiveRows[0][`cnt`])
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`:          result,
		`archive_count`: archiveCount,
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
	       d.env_id,
       d.create_time,
       d.update_time,
       count(a.id) as child_count
from tbl_api_dir d
left join tbl_api a on a.folder_id = d.id
where d.collection_id = ? and d.archived = 0
group by d.id, d.collection_id, d.name, d.headers, d.env_id, d.create_time, d.update_time
order by d.id asc`, collectionId).All()
	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		result = append(result, buildFolderBasicInfo(item))
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: result,
	})
}

// ApiCollectionFoldersBatchBasic 批量按集合查询文件夹基础信息，一次请求返回多个集合的文件夹。
func ApiCollectionFoldersBatchBasic(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	collectionIds := parseApiIDs(dataMap[`collection_ids`])
	if len(collectionIds) == 0 {
		gsgin.GinResponseError(c, `请选择集合`, nil)
		return
	}
	placeholders := make([]string, 0, len(collectionIds))
	args := make([]any, 0, len(collectionIds))
	for _, id := range collectionIds {
		placeholders = append(placeholders, `?`)
		args = append(args, id)
	}
	sql := `
select d.id,
       d.collection_id,
       d.name,
       d.headers,
       d.env_id,
       d.create_time,
       d.update_time,
       count(a.id) as child_count
from tbl_api_dir d
left join tbl_api a on a.folder_id = d.id
where d.collection_id in (` + strings.Join(placeholders, `,`) + `) and d.archived = 0
group by d.id, d.collection_id, d.name, d.headers, d.env_id, d.create_time, d.update_time
order by d.id asc`
	list, _ := common.DbMain.Client.QueryBySql(sql, args...).All()
	// 按 collection_id 分组
	grouped := make(map[string][]map[string]any)
	for _, item := range list {
		collectionId := cast.ToInt(item[`collection_id`])
		key := cast.ToString(collectionId)
		grouped[key] = append(grouped[key], buildFolderBasicInfo(item))
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`data`: grouped,
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

// ApiFolderApisBatchBasic 批量按文件夹查询接口基础信息，一次请求返回多个文件夹的接口。
func ApiFolderApisBatchBasic(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	folderIds := parseApiIDs(dataMap[`folder_ids`])
	if len(folderIds) == 0 {
		gsgin.GinResponseError(c, `请选择文件夹`, nil)
		return
	}
	placeholders := make([]string, 0, len(folderIds))
	args := make([]any, 0, len(folderIds))
	for _, id := range folderIds {
		placeholders = append(placeholders, `?`)
		args = append(args, id)
	}
	sql := `select id,folder_id,collection_id,name,method,url,desc,env_id,weight,create_time,update_time from tbl_api where folder_id in (` + strings.Join(placeholders, `,`) + `) order by weight,id asc`
	list, _ := common.DbMain.Client.QueryBySql(sql, args...).All()
	// 按 folder_id 分组
	grouped := make(map[string][]map[string]any)
	for _, item := range list {
		folderId := cast.ToInt(item[`folder_id`])
		key := cast.ToString(folderId)
		grouped[key] = append(grouped[key], buildApiBasicInfo(item))
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`data`: grouped,
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

func ApiDeleteCollectionEnv(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	envId := cast.ToInt(dataMap[`id`])
	if envId == 0 {
		gsgin.GinResponseError(c, `请选择环境`, nil)
		return
	}
	// 删除环境关联的所有变量项
	_, _ = common.DbMain.Client.QuickDelete(`tbl_api_env_item`, map[string]any{
		`env_id`: envId,
	}).Exec()
	// 删除环境
	_, _ = common.DbMain.Client.QuickDelete(`tbl_api_env`, map[string]any{
		`id`: envId,
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, nil)
}

func ApiDeleteCollectionEnvItem(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	itemId := cast.ToInt(dataMap[`id`])
	if itemId == 0 {
		gsgin.GinResponseError(c, `请选择变量`, nil)
		return
	}
	_, _ = common.DbMain.Client.QuickDelete(`tbl_api_env_item`, map[string]any{
		`id`: itemId,
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, nil)
}

func ApiCreateCollectionEnvItem(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var id any
	collectionId := cast.ToInt(dataMap[`collection_id`])
	envId := cast.ToInt(dataMap[`env_id`])
	key := cast.ToString(dataMap[`key`])
	itemId := cast.ToInt(dataMap[`id`])
	if envId == 0 || collectionId == 0 {
		gsgin.GinResponseError(c, `请选择集合和环境`, nil)
		return
	}
	if key == `` {
		gsgin.GinResponseError(c, `变量名不能为空`, nil)
		return
	}
	// 检测同环境下变量名是否重复
	exist, _ := common.DbMain.Client.QueryBySql(
		`select id from tbl_api_env_item where collection_id = ? and env_id = ? and key = ? and id != ?`,
		collectionId, envId, key, itemId,
	).All()
	if len(exist) > 0 {
		gsgin.GinResponseError(c, `变量名 "`+key+`" 已存在`, nil)
		return
	}
	updateData := gstool.MapTakeKeys(&dataMap, []string{`collection_id`, `env_id`, `desc`, `key`, `value`})
	if itemId == 0 {
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
				`id`: itemId,
			}, updateData).Exec()
		id = itemId
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
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `collection_id`, `headers`, `env_id`})
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
	changeType := `folder_updated`
	if cast.ToInt(dataMap[`id`]) == 0 {
		changeType = `folder_created`
	}
	go BroadcastApiChange(c.GetHeader("SseClientId"), changeType, map[string]any{
		`collection_id`: cast.ToInt(info[`collection_id`]),
		`folder_id`:     cast.ToInt(id),
	})
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
		dataMap[`body_raw`] = parsed.CurlStruct.BodyRaw

	}
	updateData = gstool.MapTakeKeys(&dataMap, []string{`folder_id`, `collection_id`, `name`, `method`, `url`,
		`protocol`, `desc`, `headers`, `query_params`, `content_type`, `body_form`, `body_json`, `body_raw`,
		`env_id`, `response_take`, `take_result`, `take_result_desc`})
	for key, value := range updateData {
		if gstool.ArrayExistValue(&[]string{reflect.Array.String(), reflect.Map.String(), reflect.Slice.String()}, gstool.ReflectGetType(value).String()) {
			updateData[key] = gstool.JsonEncode(value)
		}
	}
	gstool.FmtPrintlnLogTime(`保存后 %s`, gstool.JsonEncode(updateData))
	ensureCreateApiOptionalFieldsDefaults(updateData)
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
		gstool.FmtPrintlnLogTime(`最终更新的数据 %s`, gstool.JsonEncode(updateData))
		updateData[`update_time`] = time.Now().Unix()
		_, err = common.DbMain.Client.QuickUpdate(`tbl_api`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		if err != nil {
			gsgin.GinResponseError(c, `更新失败 `+err.Error(), nil)
			return
		}
		id = dataMap[`id`]
	}
	info, _ := common.DbMain.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
		`id`: id,
	}).One()
	if len(info) > 0 {
		info[`type`] = define.ApiTypeApi
		info[`uniqueid`] = fmt.Sprintf(`api%d`, info[`id`])
	}
	changeType := `api_updated`
	if cast.ToInt(dataMap[`id`]) == 0 {
		changeType = `api_created`
	}
	go BroadcastApiChange(c.GetHeader("SseClientId"), changeType, map[string]any{
		`collection_id`: cast.ToInt(info[`collection_id`]),
		`folder_id`:     cast.ToInt(info[`folder_id`]),
		`api_id`:        cast.ToInt(id),
	})
	gsgin.GinResponseSuccess(c, ``, info)
}

const normalizedIntegerType = `integer`

// validateArrayItemTypes 中文：校验数组参数项中的类型字段，禁止继续写入旧的 int 类型名。 English: Validate array item types and reject the legacy int type name.
func validateArrayItemTypes(items []map[string]any, errmsg string) error {
	for _, item := range items {
		// 中文：明确禁止旧 int 类型，确保接口定义只保留当前规范。 English: Explicitly reject legacy int to keep definitions aligned with the current schema.
		if cast.ToString(item[`type`]) == `int` {
			return errors.New(errmsg + `, type 仅支持 ` + normalizedIntegerType + `，不支持 int`)
		}
	}
	return nil
}

func ensureCreateApiOptionalFieldsDefaults(updateData map[string]any) {
	if queryParams, exists := updateData[`query_params`]; !exists || queryParams == nil || cast.ToString(queryParams) == `` {
		updateData[`query_params`] = `[]`
	}
	if headers, exists := updateData[`headers`]; !exists || headers == nil || cast.ToString(headers) == `` {
		updateData[`headers`] = `{}`
	}
	if bodyForm, exists := updateData[`body_form`]; !exists || bodyForm == nil || cast.ToString(bodyForm) == `` {
		updateData[`body_form`] = `[]`
	}
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
	if err := validateArrayItemTypes(queryParamsDataNew, errmsg); err != nil {
		return ``, err
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
	folderInfo, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `headers,env_id`, map[string]any{
		`id`: apiInfo[`folder_id`],
	}).One()
	if len(folderInfo) > 0 {
		apiInfo[`folder_headers`] = folderInfo[`headers`]
		// 接口自身无环境时继承文件夹的环境
		if cast.ToInt(apiInfo[`env_id`]) == 0 {
			apiInfo[`env_id`] = folderInfo[`env_id`]
		}
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
		`last_result`:   gstool.JsonEncode(apiCli.Result),
		`last_run_time`: time.Now().Unix(),
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
	// 接口自身无环境时继承文件夹的环境
	if cast.ToInt(apiInfo[`env_id`]) == 0 {
		folderInfo, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `env_id`, map[string]any{
			`id`: apiInfo[`folder_id`],
		}).One()
		if len(folderInfo) > 0 {
			apiInfo[`env_id`] = folderInfo[`env_id`]
		}
	}
	apiCli := api.NewApi(apiInfo)
	code := apiCli.GenerateCode(cast.ToString(codeType))
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
	go BroadcastApiChange(c.GetHeader("SseClientId"), `api_weight_changed`, map[string]any{
		`collection_id`: cast.ToInt(apiInfo[`collection_id`]),
		`folder_id`:     cast.ToInt(apiInfo[`folder_id`]),
		`api_id`:        cast.ToInt(id),
	})
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

	go BroadcastApiChange(c.GetHeader("SseClientId"), `batch_imported`, map[string]any{
		`collection_id`: collectionId,
	})
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
			`protocol`, `desc`, `headers`, `query_params`, `content_type`, `body_form`, `body_json`, `body_raw`,
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

	folderCollectionId := cast.ToInt(folderInfo[`collection_id`])

	// Update api folder_id and collection_id to match the target folder.
	_, err := common.DbMain.Client.QuickUpdate(`tbl_api`, map[string]any{
		`id`: apiId,
	}, map[string]any{
		`folder_id`:     newFolderId,
		`collection_id`: folderCollectionId,
		`update_time`:   time.Now().Unix(),
	}).Exec()

	if err != nil {
		gsgin.GinResponseError(c, `移动失败: `+err.Error(), nil)
		return
	}

	go BroadcastApiChange(c.GetHeader("SseClientId"), `api_moved`, map[string]any{
		`collection_id`: folderCollectionId,
		`folder_id`:     newFolderId,
		`api_id`:        apiId,
		`old_folder_id`: cast.ToInt(apiInfo[`folder_id`]),
	})

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
		// 接口自身无环境时继承文件夹的环境
		envId := cast.ToInt(api[`env_id`])
		if envId == 0 {
			envId = cast.ToInt(dir[`env_id`])
		}
		if envId > 0 {
			envItems, _ := common.DbMain.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
				`env_id`: envId,
			}).All()
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

// ApiArchiveFolderList 查询所有已归档的文件夹列表。
func ApiArchiveFolderList(c *gin.Context) {
	list, _ := common.DbMain.Client.QueryBySql(`
select d.id,
       d.name,
       d.headers,
       d.env_id,
       d.original_collection_id,
       d.create_time,
       d.update_time,
       count(a.id) as child_count
from tbl_api_dir d
left join tbl_api a on a.folder_id = d.id
where d.archived = 1
group by d.id, d.name, d.headers, d.env_id, d.original_collection_id, d.create_time, d.update_time
order by d.update_time desc`).All()
	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		result = append(result, buildFolderBasicInfo(item))
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: result,
	})
}

// ApiRestoreFolder 从归档中恢复文件夹到原集合。
func ApiRestoreFolder(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	folderId := cast.ToInt(dataMap[`id`])
	if folderId == 0 {
		gsgin.GinResponseError(c, `请选择文件夹`, nil)
		return
	}
	folderInfo, _ := common.DbMain.Client.QueryBySql(
		`SELECT original_collection_id FROM tbl_api_dir WHERE id = ?`, folderId,
	).One()
	if folderInfo == nil {
		gsgin.GinResponseError(c, `文件夹不存在`, nil)
		return
	}
	originalCollectionId := cast.ToInt(folderInfo[`original_collection_id`])
	if originalCollectionId == 0 {
		gsgin.GinResponseError(c, `无法确定原集合，请手动迁移`, nil)
		return
	}
	_, _ = common.DbMain.Client.QueryBySql(
		`UPDATE tbl_api_dir SET archived = 0, collection_id = ?, update_time = ? WHERE id = ?`,
		originalCollectionId, time.Now().Unix(), folderId,
	).Exec()
	go BroadcastApiChange(c.GetHeader("SseClientId"), `folder_restored`, map[string]any{
		`collection_id`: originalCollectionId,
		`folder_id`:     folderId,
	})
	gsgin.GinResponseSuccess(c, ``, nil)
}

// ApiPermanentDeleteDir 永久删除归档文件夹及其下所有接口。
func ApiPermanentDeleteDir(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	folderId := cast.ToInt(dataMap[`id`])
	if folderId == 0 {
		gsgin.GinResponseError(c, `请选择文件夹`, nil)
		return
	}
	// 删除该文件夹下的所有接口
	_, _ = common.DbMain.Client.QueryBySql(
		`DELETE FROM tbl_api WHERE folder_id = ?`, folderId,
	).Exec()
	// 删除文件夹
	_, _ = common.DbMain.Client.QueryBySql(
		`DELETE FROM tbl_api_dir WHERE id = ? AND archived = 1`, folderId,
	).Exec()
	go BroadcastApiChange(c.GetHeader("SseClientId"), `folder_permanent_deleted`, map[string]any{
		`folder_id`: folderId,
	})
	gsgin.GinResponseSuccess(c, ``, nil)
}

// parseJsonArrayMap 解析 JSON 数组字符串为 map 切片。
func parseJsonArrayMap(s string) []map[string]any {
	result := make([]map[string]any, 0)
	if s == `` {
		return result
	}
	_ = gstool.JsonDecode(s, &result)
	return result
}

// formatMarkdownTimestamp 格式化 Unix 时间戳为可读字符串。
func formatMarkdownTimestamp(unixTimestamp int64) string {
	if unixTimestamp == 0 {
		return `未知时间`
	}
	return time.Unix(unixTimestamp, 0).Format(`2006-01-02 15:04:05`)
}

// formatMarkdownBlockquote 将内容格式化为 Markdown 引用块。
func formatMarkdownBlockquote(content string) string {
	if content == `` {
		return ``
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if line == `` {
			lines[i] = `>`
		} else {
			lines[i] = `> ` + line
		}
	}
	return strings.Join(lines, "\n")
}

// formatEnterToMarkdown 将换行符替换为 <br> 并合并多余空格。
func formatEnterToMarkdown(data string) string {
	if data == `` {
		return ``
	}
	replacer := strings.NewReplacer("\r\n", "<br>", "\n", "<br>", "\r", "<br>")
	result := replacer.Replace(data)
	var sb strings.Builder
	prevSpace := false
	for _, r := range result {
		if r == ' ' {
			if prevSpace {
				continue
			}
			prevSpace = true
		} else {
			prevSpace = false
		}
		sb.WriteRune(r)
	}
	return strings.TrimSpace(sb.String())
}

// formatJsonBodyStr 格式化 JSON 字符串为缩进格式。
func formatJsonBodyStr(s string) string {
	var obj any
	if gstool.JsonDecode(s, &obj) == nil {
		b, err := json.MarshalIndent(obj, "", "  ")
		if err == nil {
			return string(b)
		}
	}
	return s
}

// formatJsonValue 格式化任意值为缩进 JSON 字符串。
func formatJsonValue(val any) string {
	b, err := json.MarshalIndent(val, "", "  ")
	if err == nil {
		return string(b)
	}
	return fmt.Sprintf("%v", val)
}

// hasApiBodyContent 判断接口是否有请求体内容。
func hasApiBodyContent(apiItem map[string]any) bool {
	contentType := cast.ToString(apiItem[`content_type`])
	if contentType == `application/json` {
		bodyJson := cast.ToString(apiItem[`body_json`])
		if bodyJson != `` {
			var parsed map[string]any
			if gstool.JsonDecode(bodyJson, &parsed) == nil && len(parsed) > 0 {
				return true
			}
		}
	} else if contentType == `application/x-www-form-urlencoded` || contentType == `multipart/form-data` {
		bodyForm := cast.ToString(apiItem[`body_form`])
		if bodyForm != `` {
			parsed := make([]map[string]any, 0)
			if gstool.JsonDecode(bodyForm, &parsed) == nil && len(parsed) > 0 {
				return true
			}
		}
	}
	if bodyRaw := cast.ToString(apiItem[`body_raw`]); bodyRaw != `` {
		return true
	}
	return false
}

// replaceApiEnvVars 替换 URL 中的环境变量占位符。
func replaceApiEnvVars(url string, envItems []map[string]any) string {
	for _, item := range envItems {
		key := cast.ToString(item[`key`])
		value := cast.ToString(item[`value`])
		if key != `` && value != `` {
			url = strings.ReplaceAll(url, `$`+key+`$`, value)
		}
	}
	return url
}

// buildFolderApisMarkdown 根据文件夹 ID 生成该文件夹下所有接口的 Markdown 文档。
func buildFolderApisMarkdown(folderId int) string {
	if folderId <= 0 {
		return ``
	}
	dir, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `*`, map[string]any{
		`id`: folderId,
	}).One()
	if len(dir) == 0 {
		return ``
	}
	folderName := cast.ToString(dir[`name`])
	apis, _ := common.DbMain.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
		`folder_id`: folderId,
	}).Order(`weight,id asc`).All()
	dirEnvId := cast.ToInt(dir[`env_id`])
	dirEnvItems := make([]map[string]any, 0)
	if dirEnvId > 0 {
		dirEnvItems, _ = common.DbMain.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
			`env_id`: dirEnvId,
		}).All()
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`## %s 接口文档`, folderName))
	sb.WriteString("\n\n")
	for i, apiItem := range apis {
		sb.WriteString(fmt.Sprintf(`### %d. %s`, i+1, cast.ToString(apiItem[`name`])))
		sb.WriteString("\n\n")
		sb.WriteString("| 项目 | 详情 |\n")
		sb.WriteString("| --- | --- |\n")
		apiMethod := cast.ToString(apiItem[`method`])
		apiUrl := cast.ToString(apiItem[`url`])
		contentType := cast.ToString(apiItem[`content_type`])
		apiEnvId := cast.ToInt(apiItem[`env_id`])
		if apiEnvId > 0 && apiEnvId != dirEnvId {
			apiEnvItems, _ := common.DbMain.Client.QuickQuery(`tbl_api_env_item`, `*`, map[string]any{
				`env_id`: apiEnvId,
			}).All()
			apiUrl = replaceApiEnvVars(apiUrl, apiEnvItems)
		} else {
			apiUrl = replaceApiEnvVars(apiUrl, dirEnvItems)
		}
		sb.WriteString(fmt.Sprintf("| 请求URL | `%s` `%s` |\n", apiMethod, apiUrl))
		hasBody := hasApiBodyContent(apiItem)
		if hasBody {
			sb.WriteString(fmt.Sprintf("| 请求类型 | `%s` |\n", contentType))
		}
		sb.WriteString(fmt.Sprintf("| 创建时间 | %s |\n", formatMarkdownTimestamp(cast.ToInt64(apiItem[`create_time`]))))
		sb.WriteString("\n")
		desc := cast.ToString(apiItem[`desc`])
		if desc != `` {
			sb.WriteString("备注\n\n")
			sb.WriteString(formatMarkdownBlockquote(desc))
			sb.WriteString("\n\n")
		}
		queryParams := parseJsonArrayMap(cast.ToString(apiItem[`query_params`]))
		if len(queryParams) > 0 {
			sb.WriteString("请求参数\n\n")
			sb.WriteString("| 字段名 | 类型 | 值 | 描述 |\n")
			sb.WriteString("| --- | --- | --- | --- |\n")
			for _, param := range queryParams {
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
					cast.ToString(param[`field`]),
					cast.ToString(param[`type`]),
					cast.ToString(param[`value`]),
					formatEnterToMarkdown(cast.ToString(param[`description`]))))
			}
			sb.WriteString("\n")
		}
		if hasBody {
			sb.WriteString("请求体\n\n")
			if contentType == `application/json` {
				bodyJson := cast.ToString(apiItem[`body_json`])
				var parsed map[string]any
				if gstool.JsonDecode(bodyJson, &parsed) == nil && len(parsed) > 0 {
					sb.WriteString("json\n")
					sb.WriteString(formatJsonBodyStr(bodyJson))
					sb.WriteString("\n\n")
				}
			} else if contentType == `application/x-www-form-urlencoded` || contentType == `multipart/form-data` {
				bodyForm := parseJsonArrayMap(cast.ToString(apiItem[`body_form`]))
				if len(bodyForm) > 0 {
					sb.WriteString("| 字段名 | 类型 | 值 | 描述 |\n")
					sb.WriteString("| --- | --- | --- | --- |\n")
					for _, param := range bodyForm {
						sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
							cast.ToString(param[`field`]),
							cast.ToString(param[`type`]),
							cast.ToString(param[`value`]),
							formatEnterToMarkdown(cast.ToString(param[`description`]))))
					}
					sb.WriteString("\n")
				}
			} else if bodyRaw := cast.ToString(apiItem[`body_raw`]); bodyRaw != `` {
				sb.WriteString("\n" + bodyRaw + "\n```\n\n")
			}
		}
		takeResultData := parseJsonArrayMap(cast.ToString(apiItem[`take_result`]))
		if len(takeResultData) > 0 {
			sb.WriteString("返回结果说明\n\n")
			sb.WriteString("| 字段名 | 类型 | 说明 |\n")
			sb.WriteString("| --- | --- | --- |\n")
			for _, item := range takeResultData {
				sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
					cast.ToString(item[`key`]),
					cast.ToString(item[`type`]),
					cast.ToString(item[`desc`])))
			}
			sb.WriteString("\n")
		}
		lastResult := cast.ToString(apiItem[`last_result`])
		if lastResult != `` && lastResult != `{}` && lastResult != `null` {
			sb.WriteString("返回结果示例\n\n")
			var lastResultData map[string]any
			if gstool.JsonDecode(lastResult, &lastResultData) == nil {
				if resultVal, exists := lastResultData[`result`]; exists {
					sb.WriteString(formatMarkdownBlockquote(formatJsonValue(resultVal)))
				} else {
					sb.WriteString(formatMarkdownBlockquote(formatJsonValue(lastResultData)))
				}
			} else {
				sb.WriteString(formatMarkdownBlockquote(lastResult))
			}
			sb.WriteString("\n\n")
		}
		if i < len(apis)-1 {
			sb.WriteString("---\n\n")
		}
	}
	return sb.String()
}

// ApiFolderApisMarkdown 通过文件夹 ID 获取接口文档（Markdown 格式），格式与前端"复制所有接口(Markdown)"按钮一致。
func ApiFolderApisMarkdown(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	folderId := cast.ToInt(dataMap[`folder_id`])
	if folderId <= 0 {
		gsgin.GinResponseError(c, `请选择文件夹`, nil)
		return
	}
	markdown := buildFolderApisMarkdown(folderId)
	if markdown == `` {
		gsgin.GinResponseError(c, `文件夹不存在或无接口`, nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`markdown`: markdown,
	})
}
