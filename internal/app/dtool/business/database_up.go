package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"fmt"
	"os"
	"path/filepath"

	"gitee.com/Sxiaobai/gs/v2/gsdefine"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

type TDataBaseUp struct {
	db             *common.CSqlite
	databaseUpPath string
	tableName      string
}

var DataBaseUp *TDataBaseUp

func NewTDataBaseUp() *TDataBaseUp {
	return newDatabaseUp(common.DbMain, component.EnvClient.DatabaseUpPath, `tbl_database_up`)
}

func NewMemoryDataBaseUp(db *common.CSqlite, databaseUpPath string) *TDataBaseUp {
	return newDatabaseUp(db, databaseUpPath, `tbl_memory_database_up`)
}

func newDatabaseUp(db *common.CSqlite, databaseUpPath, tableName string) *TDataBaseUp {
	return &TDataBaseUp{
		db:             db,
		databaseUpPath: databaseUpPath,
		tableName:      tableName,
	}
}

func (h *TDataBaseUp) Run() {
	h.CheckDataBaseUp()
	h.Up()
}

func (h *TDataBaseUp) CheckDataBaseUp() {
	name, err := h.db.Client.QuickQuery(`sqlite_master`, `name`, map[string]any{
		`name`: h.tableName,
	}).Value(`name`)
	if err != nil {
		panic(fmt.Sprintf(`数据库升级表查询失败 %s`, err.Error()))
		return
	}
	if cast.ToString(name) == `` {
		_, err := h.db.Client.ExecBySql(fmt.Sprintf(`
			CREATE TABLE "%s" (
			  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			  "filename" TEXT NOT NULL DEFAULT ''
			);
		`, h.tableName)).Exec()
		if err != nil {
			panic(fmt.Sprintf(`数据库升级表创建失败 %s`, err.Error()))
			return
		}
	}
}

func (h *TDataBaseUp) Up() {
	allAlreadyUpFiles, err := h.db.Client.QuickQuery(h.tableName, `filename`, nil).All()
	if err != nil {
		gstool.FmtPrintlnLogTime(`数据库升级表查询失败 %s`, err.Error())
		return
	}
	upFileNames := make([]string, 0)
	for _, alreadyUpFile := range allAlreadyUpFiles {
		upFileNames = append(upFileNames, cast.ToString(alreadyUpFile[`filename`]))
	}
	files := make([]string, 0)
	//循环处理
	checkDirs := make([]string, 0)
	for startYear := 2025; startYear <= cast.ToInt(gstool.TimeNowUnixToString(`Y`)); startYear++ {
		for month := 1; month <= 12; month++ {
			month := fmt.Sprintf(`%02d`, month)
			dir := filepath.Join(h.databaseUpPath, cast.ToString(startYear), month)
			exist, _ := gstool.DirPathExists(dir)
			if exist {
				checkDirs = append(checkDirs, dir)
			}
		}
	}
	for _, dir := range checkDirs {
		walkErr := gstool.DirWalk(dir, func(path string, info os.FileInfo, err error) {
			if info.IsDir() {
				return
			}
			if !gstool.ArrayExistValue(&upFileNames, info.Name()) {
				files = append(files, filepath.Join(dir, info.Name()))
			}
		})
		if walkErr != nil {
			gstool.FmtPrintlnLogTime(`数据库升级文件扫描失败 %s`, walkErr.Error())
			return
		}
	}

	gstool.ArraySort(files, gsdefine.SortAsc)
	for _, file := range files {
		gstool.FmtPrintlnLogTime(`开始处理升级文件 %s`, file)
		sql, err := gstool.FileGetContent(file)
		if err != nil {
			gstool.FmtPrintlnLogTime(`读取文件内容失败%s %s`, file, err.Error())
			return
		}
		_, err = h.db.Client.ExecBySql(sql).Exec()
		if err != nil {
			gstool.FmtPrintlnLogTime(`数据库升级文件执行失败 %s %s`, file, err.Error())
			continue
		}
		_, err = h.db.Client.QuickCreate(h.tableName, map[string]any{
			`filename`: gstool.FileGetNameByPath(file),
		}).Exec()
		if err != nil {
			gstool.FmtPrintlnLogTime(`数据库升级表插入失败 %s`, err.Error())
			return
		}
	}
}
