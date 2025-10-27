package base

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdefine"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"os"
	"path/filepath"
)

type TDataBaseUp struct {
}

func NewTDataBaseUp() *TDataBaseUp {
	return &TDataBaseUp{}
}

func (h *TDataBaseUp) Run() {
	h.CheckDataBaseUp()
	h.Up()
}

func (h *TDataBaseUp) CheckDataBaseUp() {
	gstool.FmtPrintlnLogTime(`开始检查数据库升级表`)
	name, err := Component.TSqlite.Client.QuickQuery(`sqlite_master`, `name`, map[string]any{
		`name`: `tbl_database_up`,
	}).Select()
	if err != nil {
		Component.GsLog.Errof(`数据库升级表创建失败 %s`, err.Error())
		panic(fmt.Sprintf(`数据库升级表创建失败 %s`, err.Error()))
		return
	}
	if cast.ToString(name) == `` {
		_, err := Component.TSqlite.Client.ExecBySql(`
			CREATE TABLE "tbl_database_up" (
			  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			  "filename" TEXT NOT NULL DEFAULT ''
			);
		`).Exec()
		if err != nil {
			Component.GsLog.Errof(`数据库升级表创建失败 %s`, err.Error())
			panic(fmt.Sprintf(`数据库升级表创建失败 %s`, err.Error()))
			return
		}
	}
}

func (h *TDataBaseUp) Up() {
	gstool.FmtPrintlnLogTime(`开始扫描升级sql文件 %s`, Component.Env.DatabaseUpPath)
	maxFileName, err := Component.TSqlite.Client.QuickQuery(`tbl_database_up`, `filename`, nil).
		Order(` id desc`).Limit(1).Select()
	if err != nil {
		Component.GsLog.Errof(`数据库升级表查询失败 %s`, err.Error())
		panic(fmt.Sprintf(`数据库升级表查询失败 %s`, err.Error()))
		return
	}
	gstool.FmtPrintlnLogTime(`当前最大版本文件 %s`, maxFileName)
	files := make([]string, 0)
	walkErr := gstool.DirWalk(Component.Env.DatabaseUpPath, func(path string, info os.FileInfo, err error) {
		if info.IsDir() {
			return
		}
		if info.Name() > cast.ToString(maxFileName) {
			files = append(files, info.Name())
		}
	})
	if walkErr != nil {
		Component.GsLog.Errof(`数据库升级文件扫描失败 %s`, walkErr.Error())
		panic(fmt.Sprintf(`数据库升级文件扫描失败 %s`, walkErr.Error()))
		return
	}
	gstool.ArraySort(files, gsdefine.SortAsc)
	for _, file := range files {
		gstool.FmtPrintlnLogTime(`开始处理升级文件 %s`, file)
		sql, err := gstool.FileGetContent(filepath.Join(Component.Env.DatabaseUpPath, file))
		if err != nil {
			Component.GsLog.Errof(`数据库升级文件读取失败 %s`, err.Error())
			gstool.FmtPrintlnLogTime(`读取文件内容失败%s %s`, filepath.Join(Component.Env.DatabaseUpPath, file), err.Error())
			return
		}
		_, err = Component.TSqlite.Client.ExecBySql(sql).Exec()
		if err != nil {
			Component.GsLog.Errof(`数据库升级文件执行失败 %s`, err.Error())
			gstool.FmtPrintlnLogTime(`数据库升级文件执行失败 %s`, err.Error())
			return
		}
		_, err = Component.TSqlite.Client.QuickCreate(`tbl_database_up`, map[string]any{
			`filename`: file,
		}).Exec()
		if err != nil {
			Component.GsLog.Errof(`数据库升级表插入失败 %s`, err.Error())
			gstool.FmtPrintlnLogTime(`数据库升级表插入失败 %s`, err.Error())
			return
		}
	}
}
