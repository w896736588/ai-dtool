package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/pkg/p_db"
	"fmt"
	"regexp"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

const (
	DbTypeMysql = `mysql`
	DbTypePgsql = `pgsql`
)

// DbQueryer 统一 MySQL/PGSQL 查询接口
type DbQueryer interface {
	QueryBySql(sql string, args ...any) *gsdb.SqlQuick
	ExecBySql(sql string, args ...any) *gsdb.SqlQuick
}

// getDbClient 根据配置的 db_type 获取对应的数据库客户端
func getDbClient(dbId string) (map[string]any, DbQueryer, error) {
	dbConfig, configErr := common.DbMain.GetMysqlConfig(dbId)
	if configErr != nil || len(dbConfig) == 0 {
		return nil, nil, fmt.Errorf(`未找到id为 "%s" 的数据库配置`, dbId)
	}
	dbType := cast.ToString(dbConfig[`db_type`])
	if dbType == `` {
		dbType = DbTypeMysql
	}
	switch dbType {
	case DbTypePgsql:
		client, clientErr := component.PgsqlClient.GetClient(dbConfig, common.GetCall())
		if clientErr != nil {
			return nil, nil, fmt.Errorf(`获取PgSQL连接失败: %s`, clientErr.Error())
		}
		return dbConfig, client, nil
	default:
		client, clientErr := component.MysqlClient.GetClient(dbConfig, common.GetCall())
		if clientErr != nil {
			return nil, nil, fmt.Errorf(`获取MySQL连接失败: %s`, clientErr.Error())
		}
		return dbConfig, client, nil
	}
}

// getDbConfigType 获取指定数据库配置的类型
func getDbConfigType(dbId string) string {
	dbConfig, err := common.DbMain.GetMysqlConfig(dbId)
	if err != nil || len(dbConfig) == 0 {
		return DbTypeMysql
	}
	dbType := cast.ToString(dbConfig[`db_type`])
	if dbType == `` {
		return DbTypeMysql
	}
	return dbType
}

// MysqlTables 查询数据库配置对应数据库的所有表
// 参数：mysql_id(数据库配置ID)
func MysqlTables(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, nil)
		return
	}

	mysqlId := cast.ToString(reqMap[`mysql_id`])
	if mysqlId == `` {
		gsgin.GinResponseError(c, `mysql_id不能为空`, nil)
		return
	}

	dbConfig, dbClient, err := getDbClient(mysqlId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	dbname := cast.ToString(dbConfig[`dbname`])
	sql := fmt.Sprintf(
		`SELECT TABLE_NAME AS table_name, TABLE_COMMENT AS table_comment FROM information_schema.TABLES WHERE TABLE_SCHEMA = '%s' ORDER BY TABLE_NAME`,
		dbname,
	)
	list, queryErr := dbClient.QueryBySql(sql).All()
	if queryErr != nil {
		gsgin.GinResponseError(c, `查询表列表失败: `+queryErr.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// MysqlTableStructure 查询表结构
// 参数：mysql_id(数据库配置ID), table_name(表名)
func MysqlTableStructure(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, nil)
		return
	}

	mysqlId := cast.ToString(reqMap[`mysql_id`])
	tableName := strings.TrimSpace(cast.ToString(reqMap[`table_name`]))

	if mysqlId == `` {
		gsgin.GinResponseError(c, `mysql_id不能为空`, nil)
		return
	}
	if tableName == `` {
		gsgin.GinResponseError(c, `table_name不能为空`, nil)
		return
	}
	if !isSafeTableName(tableName) {
		gsgin.GinResponseError(c, `table_name不合法`, nil)
		return
	}

	_, dbClient, err := getDbClient(mysqlId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	dbType := getDbConfigType(mysqlId)
	var sql string
	if dbType == DbTypePgsql {
		sql = fmt.Sprintf(
			`SELECT column_name AS Field, data_type AS Type, CASE WHEN is_nullable = 'YES' THEN 'YES' ELSE 'NO' END AS Null, column_default AS Default, '' AS Extra FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position`,
			tableName,
		)
	} else {
		sql = fmt.Sprintf(`SHOW FULL COLUMNS FROM %s`, tableName)
	}

	list, queryErr := dbClient.QueryBySql(sql).All()
	if queryErr != nil {
		gsgin.GinResponseError(c, `查询表结构失败: `+queryErr.Error(), nil)
		return
	}

	// 查询索引信息
	var indexSql string
	if dbType == DbTypePgsql {
		indexSql = fmt.Sprintf(
			`SELECT indexname AS IndexName, indexdef AS IndexDef FROM pg_indexes WHERE tablename = '%s' ORDER BY indexname`,
			tableName,
		)
	} else {
		indexSql = fmt.Sprintf(`SHOW INDEX FROM %s`, tableName)
	}
	indexList, indexErr := dbClient.QueryBySql(indexSql).All()
	if indexErr != nil {
		gsgin.GinResponseError(c, `查询索引失败: `+indexErr.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`:       list,
		`index_list`: indexList,
	})
}

// MysqlQuery 执行数据库查询（仅支持SELECT）
// 参数：mysql_id(数据库配置ID), sql(查询SQL)
func MysqlQuery(c *gin.Context) {
	reqMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, nil)
		return
	}

	mysqlId := cast.ToString(reqMap[`mysql_id`])
	sql := strings.TrimSpace(cast.ToString(reqMap[`sql`]))

	if mysqlId == `` {
		gsgin.GinResponseError(c, `mysql_id不能为空`, nil)
		return
	}
	if sql == `` {
		gsgin.GinResponseError(c, `sql不能为空`, nil)
		return
	}
	if !isSelectQuery(sql) {
		gsgin.GinResponseError(c, `仅支持SELECT查询`, nil)
		return
	}

	_, dbClient, err := getDbClient(mysqlId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	list, queryErr := dbClient.QueryBySql(sql).All()
	if queryErr != nil {
		gsgin.GinResponseError(c, `查询失败: `+queryErr.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// isSafeTableName 校验表名只包含安全字符，防止SQL注入
func isSafeTableName(name string) bool {
	ok, _ := regexp.MatchString(`^[A-Za-z0-9_\.]+$`, name)
	return ok
}

// isSelectQuery 检查SQL是否为SELECT语句
func isSelectQuery(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	return strings.HasPrefix(upper, `SELECT`)
}

// GetDbQueryerById 根据配置ID获取 DbQueryer（供其他包使用）
func GetDbQueryerById(dbId string) (DbQueryer, error) {
	dbConfig, configErr := common.DbMain.GetMysqlConfig(dbId)
	if configErr != nil || len(dbConfig) == 0 {
		return nil, fmt.Errorf(`未找到id为 "%s" 的数据库配置`, dbId)
	}
	dbType := cast.ToString(dbConfig[`db_type`])
	if dbType == `` {
		dbType = DbTypeMysql
	}
	switch dbType {
	case DbTypePgsql:
		return p_db.PgsqlClient.GetClient(dbConfig, common.GetCall())
	default:
		return p_db.MysqlClient.GetClient(dbConfig, common.GetCall())
	}
}
