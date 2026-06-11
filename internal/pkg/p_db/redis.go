package p_db

import (
	"context"
	"dev_tool/internal/pkg/p_common"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsdb"
	"github.com/w896736588/go-tool/gstool"
)

type TRedis struct {
	RedisClientMap map[string]*gsdb.GsRedis
	lock           sync.Mutex
}

func (h *TRedis) GetClient(redisConfig map[string]any, call *p_common.Call) (*gsdb.GsRedis, error) {
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
		sshConfig, sshConfigErr := call.GetSshConfig(redisConfig[`ssh_id`])
		if sshConfigErr != nil {
			return nil, gstool.Error(`获取ssh配置失败 %s`, sshConfigErr.Error())
		}
		gstool.FmtPrintlnLogTime(`[p_db.redis] create ssh bridge redis_id=%s ssh_id=%s target=%s:%d`,
			redisId, cast.ToString(redisConfig[`ssh_id`]), cast.ToString(redisConfig[`host`]), cast.ToInt64(redisConfig[`port`]))
		gsRedis.SshBridge = NewConfiguredSshBridge(sshConfig)
	}
	gstool.FmtPrintlnLogTime(`[p_db.redis] CreateConn begin redis_id=%s host=%s port=%d use_ssh=%v`,
		redisId, cast.ToString(redisConfig[`host`]), cast.ToInt64(redisConfig[`port`]), cast.ToInt(redisConfig[`ssh_id`]) != 0)
	connErr := gsRedis.CreateConn()
	if connErr != nil {
		gstool.FmtPrintlnLogTime(`[p_db.redis] CreateConn failed redis_id=%s err=%s`, redisId, connErr.Error())
		return nil, connErr
	}
	gstool.FmtPrintlnLogTime(`[p_db.redis] CreateConn success redis_id=%s`, redisId)
	h.RedisClientMap[redisId] = gsRedis
	return gsRedis, nil
}

func (h *TRedis) PingAll(call *p_common.Call) {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				for redisId, client := range h.RedisClientMap {
					ret := client.Client.Ping(context.Background()).String()
					if strings.Contains(ret, `No connection could be made because the target machine actively refused it`) {
						redisConfig, redisConfigErr := call.GetRedisConfig(redisId)
						if redisConfigErr != nil {
							gstool.FmtPrintlnLogTime(`获取redis配置异常 %v`, redisId)
						} else {
							gstool.FmtPrintlnLogTime(`检测到redis连接中断，开始重连 %s`, redisId)
							h.RmClient(redisConfig)
							_, getErr := h.GetClient(redisConfig, call)
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
