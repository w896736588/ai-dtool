package base

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

type TRedis struct {
	RedisClientMap map[string]*gsdb.GsRedis
	lock           sync.Mutex
}

func (h *TRedis) GetClient(redisConfig map[string]any) (*gsdb.GsRedis, error) {
	defer h.lock.Unlock()
	h.lock.Lock()
	redisId := cast.ToString(redisConfig[`id`])
	if redisId == `` {
		return nil, errors.New(`redis配置错误`)
	}
	if redisCli, ok := h.RedisClientMap[redisId]; ok {
		if redisCli.Client.Ping(context.Background()).Err() == nil {
			return redisCli, nil
		}
	}
	gsRedis := &gsdb.GsRedis{
		RedisConfig: &gsdb.RedisConfig{
			Name:              cast.ToString(redisConfig[`name`]),
			Host:              cast.ToString(redisConfig[`host`]),
			Port:              cast.ToInt64(redisConfig[`port`]),
			Password:          cast.ToString(redisConfig[`password`]),
			MaxOpenConns:      1,
			MaxIdleConns:      0,
			Default:           0,
			Username:          "",
			MaxLifetimeSecond: 3600 * 1000,
		},
	}
	if cast.ToInt(redisConfig[`ssh_id`]) != 0 {
		sshConfig, sshConfigErr := Component.TSqlite.GetSshConfig(redisConfig[`ssh_id`])
		if sshConfigErr != nil {
			return nil, gstool.Error(`获取ssh配置失败 %s`, sshConfigErr.Error())
		}
		gsRedis.Ssh = &gsssh.SshConfig{
			Name:     cast.ToString(sshConfig[`name`]),
			Host:     cast.ToString(sshConfig[`host`]),
			Port:     cast.ToString(sshConfig[`port`]),
			UserName: cast.ToString(sshConfig[`username`]),
			Password: cast.ToString(sshConfig[`password`]),
			GsSlog:   Component.GsLog,
		}
	}
	connErr := gsRedis.CreateConn()
	if connErr != nil {
		return nil, connErr
	}
	h.RedisClientMap[redisId] = gsRedis
	return gsRedis, nil
}

func (h *TRedis) PingAll() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				for redisId, client := range h.RedisClientMap {
					ret := client.Client.Ping(context.Background()).String()
					if strings.Contains(ret, `No connection could be made because the target machine actively refused it`) {
						redisConfig, redisConfigErr := Component.TSqlite.GetRedisConfig(redisId)
						if redisConfigErr != nil {
							gstool.FmtPrintlnLogTime(`获取redis配置异常 %v`, redisId)
						} else {
							gstool.FmtPrintlnLogTime(`检测到redis连接中断，开始重连 %s`, redisId)
							h.RmClient(redisConfig)
							_, getErr := h.GetClient(redisConfig)
							if getErr != nil {
								gstool.FmtPrintlnLogTime(`重新连接redis失败 %s`, getErr.Error())
							}
						}
					}
				}
			}
		}
	}()
}

func (h *TRedis) RmClient(redisConfig map[string]any) {
	defer h.lock.Unlock()
	h.lock.Lock()
	sshId := cast.ToString(redisConfig[`id`])
	delete(h.RedisClientMap, sshId)
}
