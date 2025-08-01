package base

import (
	"errors"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"sync"
	"time"
)

type TMysql struct {
	MysqlClientMap map[string]*gsdb.GsMysql
	lock           sync.Mutex
}

func (h *TMysql) GetClient(mysqlConfig map[string]any) (*gsdb.GsMysql, error) {
	defer h.lock.Unlock()
	h.lock.Lock()
	mysqlId := cast.ToString(mysqlConfig[`id`])
	if mysqlId == `` {
		return nil, errors.New(`mysql配置错误`)
	}
	if redisCli, ok := h.MysqlClientMap[mysqlId]; ok {
		return redisCli, nil
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
		GsLog: Component.GsLog,
	}
	gsMysql.OpenDebug()
	if cast.ToInt(mysqlConfig[`ssh_id`]) != 0 {
		sshConfig, sshConfigErr := Component.TSqlite.GetSshConfig(mysqlConfig[`ssh_id`])
		if sshConfigErr != nil {
			return nil, gstool.Error(`获取ssh配置失败 %s`, sshConfigErr.Error())
		}
		gsMysql.Ssh = &gsssh.SshConfig{
			Name:     cast.ToString(sshConfig[`name`]),
			Host:     cast.ToString(sshConfig[`host`]),
			Port:     cast.ToString(sshConfig[`port`]),
			UserName: cast.ToString(sshConfig[`username`]),
			Password: cast.ToString(sshConfig[`password`]),
			GsSlog:   Component.GsLog,
		}
	}
	connErr := gsMysql.CreateConn()
	if connErr != nil {
		return nil, connErr
	}
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
