package controller

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

const mainDBStorageAlertThresholdBytes int64 = 100 * 1024 * 1024

type mainDBStorageSummary struct {
	DBPath               string           `json:"db_path"`
	DBName               string           `json:"db_name"`
	DBDir                string           `json:"db_dir"`
	FileSizeBytes        int64            `json:"file_size_bytes"`
	FileSizeText         string           `json:"file_size_text"`
	AlertThreshold       int64            `json:"alert_threshold_bytes"`
	ExceedsLimit         bool             `json:"exceeds_limit"`
	AnalysisMode         string           `json:"analysis_mode"`
	AnalysisNote         string           `json:"analysis_note"`
	PageSizeBytes        int64            `json:"page_size_bytes"`
	PageCount            int64            `json:"page_count"`
	FreePageCount        int64            `json:"free_page_count"`
	FreeBytes            int64            `json:"free_bytes"`
	FreeSizeText         string           `json:"free_size_text"`
	UsedBytes            int64            `json:"used_bytes"`
	UsedSizeText         string           `json:"used_size_text"`
	UnattributedBytes    int64            `json:"unattributed_bytes"`
	UnattributedSizeText string           `json:"unattributed_size_text"`
	Tables               []map[string]any `json:"tables"`
}

func readMainDBStorageSummary() (mainDBStorageSummary, error) {
	config := business.ReadMainDBConfig()
	summary := mainDBStorageSummary{
		DBPath:         strings.TrimSpace(config.DBPath),
		DBName:         strings.TrimSpace(config.DBName),
		DBDir:          strings.TrimSpace(config.Dir),
		AlertThreshold: mainDBStorageAlertThresholdBytes,
		AnalysisMode:   `dbstat`,
		Tables:         make([]map[string]any, 0),
	}
	if summary.DBPath == `` {
		return summary, fmt.Errorf(`主库未配置`)
	}

	dbPath := filepath.Clean(summary.DBPath)
	info, err := os.Stat(dbPath)
	if err != nil {
		return summary, err
	}
	if info.IsDir() {
		return summary, fmt.Errorf(`主库路径不是文件`)
	}

	summary.FileSizeBytes = info.Size()
	summary.FileSizeText = formatStorageBytes(info.Size())
	summary.ExceedsLimit = info.Size() > mainDBStorageAlertThresholdBytes

	if common.DbMain == nil || common.DbMain.Client == nil {
		return summary, fmt.Errorf(`主库未初始化`)
	}

	fillMainDBPageMetrics(&summary)

	rows, err := readMainDBStorageByDBStat()
	if err != nil {
		if !isDBStatUnavailableErr(err) {
			return summary, err
		}
		summary.AnalysisMode = `estimated_payload`
		summary.AnalysisNote = `当前 SQLite 未启用 dbstat，表大小改为按行数据 payload 估算；索引、页碎片和部分元数据会计入“未归属空间”。`
		rows, err = readMainDBStorageByEstimatedPayload()
		if err != nil {
			return summary, err
		}
	} else {
		summary.AnalysisNote = `当前使用 dbstat，结果按表聚合并包含该表索引占用。`
	}

	summary.Tables = rows
	sort.SliceStable(summary.Tables, func(i, j int) bool {
		leftBytes := cast.ToInt64(summary.Tables[i][`total_bytes`])
		rightBytes := cast.ToInt64(summary.Tables[j][`total_bytes`])
		if leftBytes == rightBytes {
			return fmt.Sprint(summary.Tables[i][`name`]) < fmt.Sprint(summary.Tables[j][`name`])
		}
		return leftBytes > rightBytes
	})
	appendMainDBSyntheticRows(&summary)
	return summary, nil
}

func fillMainDBPageMetrics(summary *mainDBStorageSummary) {
	pageSize := queryMainDBInt64(`pragma page_size;`)
	pageCount := queryMainDBInt64(`pragma page_count;`)
	freePageCount := queryMainDBInt64(`pragma freelist_count;`)
	summary.PageSizeBytes = pageSize
	summary.PageCount = pageCount
	summary.FreePageCount = freePageCount
	summary.FreeBytes = pageSize * freePageCount
	summary.FreeSizeText = formatStorageBytes(summary.FreeBytes)
	summary.UsedBytes = pageSize * (pageCount - freePageCount)
	if summary.UsedBytes < 0 {
		summary.UsedBytes = 0
	}
	summary.UsedSizeText = formatStorageBytes(summary.UsedBytes)
}

func readMainDBStorageByDBStat() ([]map[string]any, error) {
	rows, err := common.DbMain.Client.QueryBySql(`
with table_objects as (
  select name as object_name, name as table_name
  from sqlite_schema
  where type = 'table'
    and name not like 'sqlite_%'
  union all
  select il.name as object_name, m.name as table_name
  from sqlite_schema m,
       pragma_index_list(m.name) il
  where m.type = 'table'
    and m.name not like 'sqlite_%'
)
select
  table_name as name,
  sum(d.pgsize) as total_bytes,
  count(*) as page_count
from table_objects o
join dbstat d on d.name = o.object_name
where d.aggregate = false
group by table_name
order by total_bytes desc, table_name asc
`).All()
	if err != nil {
		return nil, err
	}
	return normalizeMainDBStorageRows(rows, `dbstat`), nil
}

func readMainDBStorageByEstimatedPayload() ([]map[string]any, error) {
	tableRows, err := common.DbMain.Client.QueryBySql(`
select name
from sqlite_schema
where type = 'table'
  and name not like 'sqlite_%'
order by name asc
`).All()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, 0, len(tableRows))
	for _, tableRow := range tableRows {
		tableName := strings.TrimSpace(cast.ToString(tableRow[`name`]))
		if tableName == `` {
			continue
		}
		rowCount, payloadBytes, indexCount, tableErr := estimateMainDBTablePayload(tableName)
		if tableErr != nil {
			return nil, tableErr
		}
		result = append(result, map[string]any{
			`name`:            tableName,
			`total_bytes`:     payloadBytes,
			`total_size_text`: formatStorageBytes(payloadBytes),
			`page_count`:      int64(0),
			`row_count`:       rowCount,
			`index_count`:     indexCount,
			`mode`:            `estimated_payload`,
		})
	}
	return result, nil
}

func estimateMainDBTablePayload(tableName string) (int64, int64, int64, error) {
	columnRows, err := common.DbMain.Client.QueryBySql(fmt.Sprintf(`
select name
from pragma_table_info(%s)
order by cid asc
`, quoteSQLiteString(tableName))).All()
	if err != nil {
		return 0, 0, 0, err
	}

	indexRows, err := common.DbMain.Client.QueryBySql(fmt.Sprintf(`
select count(*) as cnt
from pragma_index_list(%s)
`, quoteSQLiteString(tableName))).All()
	if err != nil {
		return 0, 0, 0, err
	}
	indexCount := int64(0)
	if len(indexRows) > 0 {
		indexCount = cast.ToInt64(indexRows[0][`cnt`])
	}

	payloadExprList := make([]string, 0, len(columnRows))
	for _, columnRow := range columnRows {
		columnName := strings.TrimSpace(cast.ToString(columnRow[`name`]))
		if columnName == `` {
			continue
		}
		payloadExprList = append(payloadExprList, fmt.Sprintf(`ifnull(length(cast(%s as blob)), 0)`, quoteSQLiteIdentifier(columnName)))
	}
	payloadExpr := `0`
	if len(payloadExprList) > 0 {
		payloadExpr = strings.Join(payloadExprList, ` + `)
	}

	rows, err := common.DbMain.Client.QueryBySql(fmt.Sprintf(`
select
  count(*) as row_count,
  ifnull(sum(%s), 0) as payload_bytes
from %s
`, payloadExpr, quoteSQLiteIdentifier(tableName))).All()
	if err != nil {
		return 0, 0, 0, err
	}
	if len(rows) == 0 {
		return 0, 0, indexCount, nil
	}
	return cast.ToInt64(rows[0][`row_count`]), cast.ToInt64(rows[0][`payload_bytes`]), indexCount, nil
}

func normalizeMainDBStorageRows(rows []map[string]any, mode string) []map[string]any {
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(cast.ToString(row[`name`]))
		totalBytes := cast.ToInt64(row[`total_bytes`])
		pageCount := cast.ToInt64(row[`page_count`])
		result = append(result, map[string]any{
			`name`:            name,
			`total_bytes`:     totalBytes,
			`total_size_text`: formatStorageBytes(totalBytes),
			`page_count`:      pageCount,
			`mode`:            mode,
		})
	}
	return result
}

func appendMainDBSyntheticRows(summary *mainDBStorageSummary) {
	totalTableBytes := int64(0)
	for _, row := range summary.Tables {
		totalTableBytes += cast.ToInt64(row[`total_bytes`])
	}

	if summary.FreeBytes > 0 {
		summary.Tables = append(summary.Tables, map[string]any{
			`name`:            `[free pages]`,
			`total_bytes`:     summary.FreeBytes,
			`total_size_text`: formatStorageBytes(summary.FreeBytes),
			`page_count`:      summary.FreePageCount,
			`mode`:            `free_pages`,
		})
	}

	baseUsedBytes := summary.UsedBytes
	if baseUsedBytes <= 0 {
		baseUsedBytes = summary.FileSizeBytes - summary.FreeBytes
	}
	summary.UnattributedBytes = baseUsedBytes - totalTableBytes
	if summary.UnattributedBytes < 0 {
		summary.UnattributedBytes = 0
	}
	summary.UnattributedSizeText = formatStorageBytes(summary.UnattributedBytes)

	if summary.UnattributedBytes > 0 {
		label := `[indexes/meta/fragmentation]`
		if summary.AnalysisMode == `dbstat` {
			label = `[sqlite internal/meta]`
		}
		summary.Tables = append(summary.Tables, map[string]any{
			`name`:            label,
			`total_bytes`:     summary.UnattributedBytes,
			`total_size_text`: formatStorageBytes(summary.UnattributedBytes),
			`page_count`:      int64(0),
			`mode`:            `derived`,
		})
	}

	sort.SliceStable(summary.Tables, func(i, j int) bool {
		leftBytes := cast.ToInt64(summary.Tables[i][`total_bytes`])
		rightBytes := cast.ToInt64(summary.Tables[j][`total_bytes`])
		if leftBytes == rightBytes {
			return fmt.Sprint(summary.Tables[i][`name`]) < fmt.Sprint(summary.Tables[j][`name`])
		}
		return leftBytes > rightBytes
	})
}

func queryMainDBInt64(sql string) int64 {
	rows, err := common.DbMain.Client.QueryBySql(sql).All()
	if err != nil || len(rows) == 0 {
		return 0
	}
	for _, row := range rows[0] {
		return cast.ToInt64(row)
	}
	return 0
}

func isDBStatUnavailableErr(err error) bool {
	if err == nil {
		return false
	}
	errText := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(errText, `no such table: dbstat`) || strings.Contains(errText, `no such module: dbstat`)
}

func quoteSQLiteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func quoteSQLiteString(value string) string {
	return `'` + strings.ReplaceAll(value, `'`, `''`) + `'`
}

func formatStorageBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf(`%d B`, size)
	}
	units := []string{`KB`, `MB`, `GB`, `TB`}
	value := float64(size)
	for _, unit := range units {
		value /= 1024
		if value < 1024 || unit == units[len(units)-1] {
			return fmt.Sprintf(`%.2f %s`, value, unit)
		}
	}
	return fmt.Sprintf(`%d B`, size)
}

// SetMainDBStorageAnalysis returns the main sqlite file size and a table-level storage breakdown.
func SetMainDBStorageAnalysis(c *gin.Context) {
	summary, err := readMainDBStorageSummary()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, summary)
}

// SetMainDBStorageVacuum runs VACUUM on the main sqlite database to reclaim free pages.
func SetMainDBStorageVacuum(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	if _, err := common.DbMain.Client.ExecBySql(`VACUUM;`).Exec(); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	summary, err := readMainDBStorageSummary()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, `清理完成`, summary)
}
