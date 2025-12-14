package business

import (
	"dev_tool/internal/app/dtool/common"
	"fmt"
	"gitee.com/Sxiaobai/gs/v2/gsdefine"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
	"os"
	"path/filepath"
)

type TDataBaseUp struct {
}

var DataBaseUp *TDataBaseUp

func NewTDataBaseUp() *TDataBaseUp {
	return &TDataBaseUp{}
}

func (h *TDataBaseUp) Run() {
	h.CheckDataBaseUp()
	h.Up()
}

func (h *TDataBaseUp) CheckDataBaseUp() {
	gstool.FmtPrintlnLogTime(`开始检查数据库升级表`)
	name, err := common.DbMain.Client.QuickQuery(`sqlite_master`, `name`, map[string]any{
		`name`: `tbl_database_up`,
	}).Select()
	if err != nil {
		panic(fmt.Sprintf(`数据库升级表查询失败 %s`, err.Error()))
		return
	}
	if cast.ToString(name) == `` {
		_, err := common.DbMain.Client.ExecBySql(`
			CREATE TABLE "tbl_database_up" (
			  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			  "filename" TEXT NOT NULL DEFAULT ''
			);
		`).Exec()
		if err != nil {
			panic(fmt.Sprintf(`数据库升级表创建失败 %s`, err.Error()))
			return
		}
	}
}

func (h *TDataBaseUp) Up() {
	allAlreadyUpFiles, err := common.DbMain.Client.QuickQuery(`tbl_database_up`, `filename`, nil).All()
	if err != nil {
		gstool.FmtPrintlnLogTime(`数据库升级表查询失败 %s`, err.Error())
		return
	}
	upFileNames := make([]string, 0)
	for _, alreadyUpFile := range allAlreadyUpFiles {
		upFileNames = append(upFileNames, cast.ToString(alreadyUpFile[`filename`]))
	}
	gstool.FmtPrintlnLogTime(`当前已执行sql文件 %d`, len(allAlreadyUpFiles))
	files := make([]string, 0)
	//循环处理
	checkDirs := make([]string, 0)
	for startYear := 2025; startYear <= cast.ToInt(gstool.TimeNowUnixToString(`Y`)); startYear++ {
		for month := 1; month <= 12; month++ {
			month := fmt.Sprintf(`%02d`, month)
			dir := filepath.Join(common.EnvClient.DatabaseUpPath, cast.ToString(startYear), month)
			exist, _ := gstool.DirPathExists(dir)
			if exist {
				checkDirs = append(checkDirs, dir)
			}
		}
	}
	gstool.FmtPrintlnLogTime(`检查的文件夹 %s`, gstool.JsonFormat(checkDirs))
	for _, dir := range checkDirs {
		gstool.FmtPrintlnLogTime(`开始扫描升级目录 %s`, dir)
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
		_, err = common.DbMain.Client.ExecBySql(sql).Exec()
		if err != nil {
			gstool.FmtPrintlnLogTime(`数据库升级文件执行失败 %s`, err.Error())
			continue
		}
		_, err = common.DbMain.Client.QuickCreate(`tbl_database_up`, map[string]any{
			`filename`: file,
		}).Exec()
		if err != nil {
			gstool.FmtPrintlnLogTime(`数据库升级表插入失败 %s`, err.Error())
			return
		}
	}
}
