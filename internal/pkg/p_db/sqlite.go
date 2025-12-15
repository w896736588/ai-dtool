package p_db

import (
	"fmt"
	"path/filepath"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gstool"
)

func InitSqlite(dbPath, dbName string) (*gsdb.GsSqlite, error) {
	var err error
	sqliteClient, err := gsdb.NewSqlite(filepath.Join(dbPath, dbName), true)
	if err != nil {
		return nil, gstool.Error(fmt.Sprintf(`连接sqlite失败 %s`, err.Error()))
	}
	//sqlite.OpenDebug()
	createErr := sqliteClient.CreateConn()
	if createErr != nil {
		return nil, gstool.Error(fmt.Sprintf(`打开sqlite失败 %s`, createErr.Error()))
	}
	return sqliteClient, nil
}
