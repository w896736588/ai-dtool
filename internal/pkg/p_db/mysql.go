package p_db

import (
	"dev_tool/internal/pkg/p_common"
	"errors"
	"sync"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsdb"
	"github.com/w896736588/go-tool/gstool"
)

var MysqlClient *TMysql

type TMysql struct {
	MysqlClientMap map[string]*gsdb.GsMysql
	lock           sync.Mutex
}

func InitMysql() {
	MysqlClient = &TMysql{
		MysqlClientMap: make(map[string]*gsdb.GsMysql),
	}
}

func (h *TMysql) GetClient(mysqlConfig map[string]any, call *p_common.Call) (*gsdb.GsMysql, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	mysqlId := cast.ToString(mysqlConfig[`id`])
	if mysqlId == `` {
		return nil, errors.New(`mysql配置错误`)
	}
	if mysqlCli, ok := h.MysqlClientMap[mysqlId]; ok {
		_, err := mysqlCli.QueryBySql(`select 1`).All()
		if err == nil {
			return mysqlCli, nil
		}
	}
	gsMysql := &gsdb.GsMysql{
		MysqlConfig: &gsdb.MysqlConfig{
			Name:     cast.ToString(mysqlConfig[`name`]),
			Host:     cast.ToString(mysqlConfig[`host`]),
			Port:     cast.ToInt64(mysqlConfig[`port`]),
			Password: cast.ToString(mysqlConfig[`password`]),
			Username: cast.ToString(mysqlConfig[`username`]),
			Dbname:   cast.ToString(mysqlConfig[`dbname`]),
		},
	}
	gsMysql.RegisterDebugHook(func(sql string, err error) {
		if err != nil {
			gstool.FmtPrintlnLogTime(`error sql：mysql %s %s`, sql, err.Error())
		} else {
			gstool.FmtPrintlnLogTime(`success sql：mysql %s`, sql)
		}
	})
	if cast.ToInt(mysqlConfig[`ssh_id`]) != 0 {
		sshConfig, sshConfigErr := call.GetSshConfig(mysqlConfig[`ssh_id`])
		if sshConfigErr != nil {
			return nil, gstool.Error(`获取ssh配置失败 %s`, sshConfigErr.Error())
		}
		gstool.FmtPrintlnLogTime(`[p_db.mysql] create ssh bridge mysql_id=%s ssh_id=%s target=%s:%d`,
			mysqlId, cast.ToString(mysqlConfig[`ssh_id`]), cast.ToString(mysqlConfig[`host`]), cast.ToInt64(mysqlConfig[`port`]))
		gsMysql.SshBridge = NewConfiguredSshBridge(sshConfig)
	}
	gstool.FmtPrintlnLogTime(`[p_db.mysql] CreateConn begin mysql_id=%s db=%s host=%s port=%d use_ssh=%v`,
		mysqlId, cast.ToString(mysqlConfig[`dbname`]), cast.ToString(mysqlConfig[`host`]), cast.ToInt64(mysqlConfig[`port`]), cast.ToInt(mysqlConfig[`ssh_id`]) != 0)
	connErr := gsMysql.CreateConn()
	if connErr != nil {
		gstool.FmtPrintlnLogTime(`[p_db.mysql] CreateConn failed mysql_id=%s err=%s`, mysqlId, connErr.Error())
		return nil, connErr
	}
	gstool.FmtPrintlnLogTime(`[p_db.mysql] CreateConn success mysql_id=%s`, mysqlId)
	h.MysqlClientMap[mysqlId] = gsMysql
	return gsMysql, nil
}

func (h *TMysql) PingAll() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				//TODO 连接检查
			}
		}
	}()
}

func (h *TMysql) RmClient(mysqlConfig map[string]any) {
	defer h.lock.Unlock()
	h.lock.Lock()
	mysqlId := cast.ToString(mysqlConfig[`id`])
	delete(h.MysqlClientMap, mysqlId)
}
