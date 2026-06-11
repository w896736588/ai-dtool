package p_db

import (
	"dev_tool/internal/pkg/p_common"
	"errors"
	"sync"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsdb"
	"github.com/w896736588/go-tool/gstool"
)

var PgsqlClient *TPgsql

type TPgsql struct {
	PgsqlClientMap map[string]*gsdb.GsPgsql
	lock           sync.Mutex
}

func InitPgsql() {
	PgsqlClient = &TPgsql{
		PgsqlClientMap: make(map[string]*gsdb.GsPgsql),
	}
}

func (h *TPgsql) GetClient(dbConfig map[string]any, call *p_common.Call) (*gsdb.GsPgsql, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	dbId := cast.ToString(dbConfig[`id`])
	if dbId == `` {
		return nil, errors.New(`pgsql配置错误`)
	}
	if pgsqlCli, ok := h.PgsqlClientMap[dbId]; ok {
		_, err := pgsqlCli.QueryBySql(`select 1`).All()
		if err == nil {
			return pgsqlCli, nil
		}
	}
	gsPgsql := &gsdb.GsPgsql{
		PgsqlConfig: &gsdb.PgsqlConfig{
			Name:     cast.ToString(dbConfig[`name`]),
			Host:     cast.ToString(dbConfig[`host`]),
			Port:     cast.ToInt64(dbConfig[`port`]),
			Password: cast.ToString(dbConfig[`password`]),
			Username: cast.ToString(dbConfig[`username`]),
			Dbname:   cast.ToString(dbConfig[`dbname`]),
		},
	}
	gsPgsql.RegisterDebugHook(func(sql string, err error) {
		if err != nil {
			gstool.FmtPrintlnLogTime(`error sql：pgsql %s %s`, sql, err.Error())
		} else {
			gstool.FmtPrintlnLogTime(`success sql：pgsql %s`, sql)
		}
	})
	if cast.ToInt(dbConfig[`ssh_id`]) != 0 {
		sshConfig, sshConfigErr := call.GetSshConfig(dbConfig[`ssh_id`])
		if sshConfigErr != nil {
			return nil, gstool.Error(`获取ssh配置失败 %s`, sshConfigErr.Error())
		}
		gstool.FmtPrintlnLogTime(`[p_db.pgsql] create ssh bridge db_id=%s ssh_id=%s target=%s:%d`,
			dbId, cast.ToString(dbConfig[`ssh_id`]), cast.ToString(dbConfig[`host`]), cast.ToInt64(dbConfig[`port`]))
		gsPgsql.SshBridge = NewConfiguredSshBridge(sshConfig)
	}
	gstool.FmtPrintlnLogTime(`[p_db.pgsql] CreateConn begin db_id=%s db=%s host=%s port=%d use_ssh=%v`,
		dbId, cast.ToString(dbConfig[`dbname`]), cast.ToString(dbConfig[`host`]), cast.ToInt64(dbConfig[`port`]), cast.ToInt(dbConfig[`ssh_id`]) != 0)
	connErr := gsPgsql.CreateConn()
	if connErr != nil {
		gstool.FmtPrintlnLogTime(`[p_db.pgsql] CreateConn failed db_id=%s err=%s`, dbId, connErr.Error())
		return nil, connErr
	}
	gstool.FmtPrintlnLogTime(`[p_db.pgsql] CreateConn success db_id=%s`, dbId)
	h.PgsqlClientMap[dbId] = gsPgsql
	return gsPgsql, nil
}

func (h *TPgsql) RmClient(dbConfig map[string]any) {
	defer h.lock.Unlock()
	h.lock.Lock()
	dbId := cast.ToString(dbConfig[`id`])
	delete(h.PgsqlClientMap, dbId)
}
