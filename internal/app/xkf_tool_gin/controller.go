package xkf_tool_gin

import (
	"encoding/json"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"xkf_tool/internal/app/xkf_tool"
)

var RedisHandleList []gsdb.GsRedisConfig

func RedisList(c *gin.Context) {
	reqBody := &xkf_tool.SshExec{}
	requestData(c, &reqBody)

	for _, value := range reqBody.RedisConfigList {
		if xkf_tool.RedisRunList[value.Name] == nil {
			//初始化链接
			gsRedisConfig := &gsdb.GsRedisConfig{
				Name:     value.Name,
				Host:     value.Host,
				Password: value.Password,
				PoolSize: value.PoolSize,
				Default:  0,
			}
			gsredis := gsdb.GsRedis{
				RedisConfig: gsRedisConfig,
			}
			err := gsredis.CreateConn()
			if err != nil {
				continue
			}
			xkf_tool.RedisRunList[value.Name] = &gsredis
		}
	}
	RedisHandleList = make([]gsdb.GsRedisConfig, 0)
	for _, value := range reqBody.RedisConfigList {
		if xkf_tool.RedisRunList[value.Name] != nil {
			RedisHandleList = append(RedisHandleList, value)
		}

	}
	response(c, xkf_tool.ErrorCodeSuccess, `获取成功`, RedisHandleList)
}

func Keys(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.SearchBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	var resultMap []string
	resultMap, err = redisCli.Keys(reqBody.Search).Result()
	if err == redis.Nil {
		resultMap = make([]string, 0)
	}
	//拿到key类型
	returnList := make([]xkf_tool.KeysList, 0)
	for _, cacheKey := range resultMap {
		returnList = append(returnList, xkf_tool.KeysList{
			CacheKey: cacheKey,
			Type:     ` `,
			Loading:  true,
		})
	}
	response(c, xkf_tool.ErrorCodeSuccess, `获取成功`, returnList)
}

func KeysType(c *gin.Context) {
	reqBody := &xkf_tool.SearchKeysTypeBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	//拿到key类型
	returnList := make([]xkf_tool.KeysList, 0)
	for _, cacheKey := range reqBody.KeysList {
		keyType, err := redisCli.Type(cacheKey).Result()
		if err == nil && keyType != `` {
			returnList = append(returnList, xkf_tool.KeysList{
				CacheKey: cacheKey,
				Type:     keyType,
			})
		}
	}
	response(c, xkf_tool.ErrorCodeSuccess, `获取成功`, returnList)
}

func Search(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.RequestBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	//找到key是什么类型
	keyType, err := redisCli.Type(reqBody.CacheKey).Result()
	if err != nil || keyType == `` {
		response(c, xkf_tool.ErrorCodeRunError, `获取元素类型失败`, ``)
		return
	}
	if keyType == xkf_tool.CacheString {
		var result string
		result, err = redisCli.Get(reqBody.CacheKey).Result()
		if err == redis.Nil {
			response(c, xkf_tool.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
			return
		}
		response(c, xkf_tool.ErrorCodeSuccess, `获取成功`, result)
	} else if keyType == xkf_tool.CacheHash {
		var resultMap map[string]string
		resultMap, err = redisCli.HGetAll(reqBody.CacheKey).Result()
		if err == redis.Nil {
			response(c, xkf_tool.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
			return
		}
		response(c, xkf_tool.ErrorCodeSuccess, `获取成功`, resultMap)
	} else if keyType == xkf_tool.CacheList {
		var resultArray []string
		resultArray, err = redisCli.LRange(reqBody.CacheKey, 0, 100000).Result()
		if err == redis.Nil {
			response(c, xkf_tool.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
			return
		}
		response(c, xkf_tool.ErrorCodeSuccess, `获取成功`, resultArray)
	} else if keyType == xkf_tool.CacheSet {
		var resultArray []string
		resultArray, err = redisCli.SMembers(reqBody.CacheKey).Result()
		if err == redis.Nil {
			response(c, xkf_tool.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
			return
		}
		response(c, xkf_tool.ErrorCodeSuccess, `获取成功`, resultArray)
	} else if keyType == xkf_tool.CacheZSet {
		var resultArray []redis.Z
		resultArray, err = redisCli.ZRangeWithScores(reqBody.CacheKey, 0, 100000).Result()
		if err == redis.Nil {
			response(c, xkf_tool.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
			return
		}
		response(c, xkf_tool.ErrorCodeSuccess, `获取成功`, resultArray)
	}
}

func GetKeyType(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.RequestBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	//找到key是什么类型
	keyType, err := redisCli.Type(reqBody.CacheKey).Result()
	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
		return
	} else if keyType == `` {
		response(c, xkf_tool.ErrorCodeRunError, `获取元素类型失败`, ``)
		return
	}
	//找到过期时间
	var ttl time.Duration
	ttl, err = redisCli.TTL(reqBody.CacheKey).Result()
	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
		return
	}
	response(c, xkf_tool.ErrorCodeSuccess, `获取类型和过期时间成功`, &xkf_tool.TypeResponse{
		Type: keyType,
		TTL:  cast.ToInt(ttl.Seconds()),
	})
}

func PhpSerialize(c *gin.Context) {

}

func SaveString(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.SaveString{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	ttlTime, err := redisCli.TTL(reqBody.CacheKey).Result()
	//永久
	err = redisCli.Set(reqBody.CacheKey, reqBody.Value, ttlTime).Err()
	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, xkf_tool.ErrorCodeSuccess, `保存成功`, ``)
	}
}

func PhpUnSerialize(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.SerializeBody{}
	requestData(c, &reqBody)
	var out string
	out, err = gstool.PhpUnSerialize(reqBody.SerializeStr)
	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), reqBody.SerializeStr)
		return
	}
	response(c, xkf_tool.ErrorCodeSuccess, `成功`, out)
}

func DelKey(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.DelKey{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	//永久
	err = redisCli.Del(reqBody.CacheKey).Err()
	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, xkf_tool.ErrorCodeSuccess, `删除成功`, ``)
	}
}

func DelSub(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.DelSub{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	if reqBody.CacheType == xkf_tool.CacheString {
		response(c, xkf_tool.ErrorCodeRunError, `不支持字符串`, ``)
	} else if reqBody.CacheType == xkf_tool.CacheHash {
		err = redisCli.HDel(reqBody.CacheKey, reqBody.Sub).Err()
	} else if reqBody.CacheType == xkf_tool.CacheList {
		err = redisCli.LRem(reqBody.CacheKey, 0, reqBody.Sub).Err()
	} else if reqBody.CacheType == xkf_tool.CacheSet {
		err = redisCli.SRem(reqBody.CacheKey, reqBody.Sub).Err()
	} else if reqBody.CacheType == xkf_tool.CacheZSet {
		err = redisCli.ZRem(reqBody.CacheKey, reqBody.Sub).Err()
	}
	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, xkf_tool.ErrorCodeSuccess, `删除成功`, ``)
	}
}

func EditTtl(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.EditTTL{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	dru := time.Duration(reqBody.TTL) * time.Second
	err = redisCli.Expire(reqBody.CacheKey, dru).Err()
	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, xkf_tool.ErrorCodeSuccess, `设置成功`, ``)
	}
}

func DelAllKey(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.DelAllKey{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	err = redisCli.Del(reqBody.CacheKeys...).Err()
	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, xkf_tool.ErrorCodeSuccess, `删除成功`, ``)
	}
}

func CreateCache(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.CreateCache{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	//判断是否存在
	if reqBody.BoolCreate == 1 {
		if existInt := redisCli.Exists(reqBody.CacheKey).Val(); existInt > 0 {
			response(c, xkf_tool.ErrorKeyNotExist, fmt.Sprintf(`%s 已经存在`, reqBody.CacheKey), ``)
			return
		}
	} else {
		if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
			return
		}
	}

	if reqBody.CacheType == xkf_tool.CacheString {
		err = redisCli.Set(reqBody.CacheKey, reqBody.CacheValue, time.Duration(reqBody.TTL)*time.Second).Err()
	} else if reqBody.CacheType == xkf_tool.CacheHash {
		err = redisCli.HSet(reqBody.CacheKey, reqBody.CacheField, reqBody.CacheValue).Err()
	} else if reqBody.CacheType == xkf_tool.CacheList {
		if reqBody.LPushValue != `` {
			err = redisCli.LPush(reqBody.CacheKey, reqBody.LPushValue).Err()
		} else if reqBody.RPushValue != `` {
			err = redisCli.RPush(reqBody.CacheKey, reqBody.RPushValue).Err()
		} else {
			err = redisCli.RPush(reqBody.CacheKey, reqBody.CacheValue).Err()
		}
	} else if reqBody.CacheType == xkf_tool.CacheSet {
		err = redisCli.SAdd(reqBody.CacheKey, reqBody.CacheMember).Err()
	} else if reqBody.CacheType == xkf_tool.CacheZSet {
		err = redisCli.ZAdd(reqBody.CacheKey, redis.Z{
			Score:  reqBody.CacheScore,
			Member: reqBody.CacheMember,
		}).Err()
	}
	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
	}
	//处理过期时间
	if reqBody.BoolCreate == 1 && reqBody.TTL != 0 {
		err = redisCli.Expire(reqBody.CacheKey, time.Duration(reqBody.TTL)*time.Second).Err()
	}

	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, xkf_tool.ErrorCodeSuccess, `创建成功`, ``)
	}
}

func EditSub(c *gin.Context) {
	var err error
	reqBody := &xkf_tool.EditSub{}
	requestData(c, &reqBody)
	log.Errorf(`editSub %#v`, reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		log.Errorf(`exist %v`, exist)
		return
	}
	if reqBody.CacheType == xkf_tool.CacheHash {
		err = redisCli.HSet(reqBody.CacheKey, reqBody.CacheField, reqBody.CacheValue).Err()
	} else if reqBody.CacheType == xkf_tool.CacheList {
		err = redisCli.LSet(reqBody.CacheKey, reqBody.CacheIndex, reqBody.CacheValue).Err()
	} else if reqBody.CacheType == xkf_tool.CacheZSet {
		err = redisCli.ZAdd(reqBody.CacheKey, redis.Z{
			Score:  reqBody.CacheScore,
			Member: reqBody.CacheMember,
		}).Err()
	}

	if err != nil {
		response(c, xkf_tool.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, xkf_tool.ErrorCodeSuccess, `编辑成功`, ``)
	}
}

// SupervisorStatus supervisor 状态
// @author frog
// @date 2022-04-11 15:22:27
//func SupervisorStatus(c *gin.Context) {
//	reqBody := &define.SshDo{}
//	requestData(c, &reqBody)
//	ret := base.Exec(reqBody, `supervisorctl status`)
//	//解析ret
//	supervisorNameList := strings.Split(strings.Replace(ret, "\n", " #ENTER# ", -1), `#ENTER#`)
//	response(c, define.ErrorCodeSuccess, `成功`, supervisorNameList)
//}

// ShellExec 执行shell命令
// @auth frog
// @date 2022-12-02 15:00:23
// @param c
func ShellExec(c *gin.Context) {
	defer func() {
		err := recover()
		if err != nil {
			response(c, xkf_tool.ErrorCodeSuccess, `成功`, fmt.Sprintf(`%#v`, err))
		}
	}()

	reqBody := &xkf_tool.SshExec{}
	requestData(c, &reqBody)
	xkf_tool.Logger.Infof(reqBody.ExecType)

	//初始化shell
	xkf_tool.Logger.Infof(`获取shell`)
	cliConf, cliTerConf := xkf_tool.GetRunShellCli(reqBody)
	xkf_tool.Logger.Infof(`获取shell成功`)
	handle := &Command{}
	handle.Filter()
	//初始化mysql
	xkf_tool.GetDevMysql(reqBody)

	switch reqBody.ExecType {
	case `query_current_branch`: //查询当前代码环境分支
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.QueryCurrentBranch(reqBody, cliTerConf), ``))
		return
	case `change_branch`: //切换分支
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.ChangeBranch(reqBody, cliTerConf), ``))
		return
	case `pull_branch_origin`: //拉取当前环境最新代码
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.PullBranchOrigin(reqBody, cliTerConf), ``))
		return
	case `wechat_kefu_status`: //查询微信客服所在的环境
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.WechatKefuStatus(reqBody, cliConf), ``))
		return
	case `wechat_kefu_change`: //切换微信客服到当前代码环境
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.WechatKefuChange(reqBody, cliConf), ``))
		return
	case `query_env_wechatkefu_list`: //微信客服
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, handle.QueryEnvWechatKefuList(reqBody))
		return
	case `supervisor_restart_all`: //消费者管理
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.SupervisorRestartAll(reqBody, cliTerConf), ``))
		return
	case `supervisor_status_list`: //消费者列表
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.SupervisorStatusList(reqBody, cliTerConf), ``))
		return
	case `supervisor_config_show`: //查看supervisor配置
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.SupervisorConfigShow(reqBody, cliTerConf), ``))
		return
	case `supervisor_restart`: //重启消费者
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.SupervisorRestart(reqBody, cliTerConf), ``))
		return
	case `supervisor_stop`: //停止消费者
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.SupervisorStop(reqBody, cliConf), ``))
		return
	case `git_status`: //git status
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.QueryStatus(reqBody, cliConf), ``))
		return
	case `show_log`:
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.ShowLog(reqBody, cliConf), ``))
		return
	case `docker_exec`:
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.DockerExec(reqBody, cliConf), ``))
		return
	case `change_vip_type`:
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.ChangeVipType(reqBody), ``))
		return
	case `query_vip_info`: //查询VIP信息
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.QueryVipType(reqBody), ``))
		return
	case `login_xkf`: //登录地址获取
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.GetLoginUrl(reqBody), ``))
		return
	case `check_all_docker_status`: //检查所有docker状态
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.CheckAllDockerStatus(reqBody, cliConf), ``))
		return
	case `restart_docker`: //重启docker
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.RestartDocker(reqBody, cliConf), ``))
		return
	case `show_compose`: //查看docker compose内容
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.ShowCompose(reqBody, cliConf), ``))
		return
	case `SupervisorConfList`:
		response(c, xkf_tool.ErrorCodeSuccess, `成功`, strings.Join(handle.SupervisorConfList(reqBody, cliTerConf), ``))
		return
	}
	response(c, xkf_tool.ErrorCodeSuccess, `成功`, nil)
}

func requestData(c *gin.Context, requestBody interface{}) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorf(`error readAll %#v`, err.Error())
	}
	_ = json.Unmarshal(body, &requestBody)
	return
}

func response(c *gin.Context, errcode int, errmsg string, body interface{}) {
	returnJson := gstool.JsonEncode(&xkf_tool.Response{
		Errcode: errcode,
		Errmsg:  errmsg,
		Data:    body,
	})
	c.String(http.StatusOK, returnJson)
}

func getRedisClient(c *gin.Context, UniKey string) *gsdb.GsRedis {
	if ok := xkf_tool.RedisRunList[UniKey]; ok == nil {
		response(c, xkf_tool.ErrorCodeErrorUniKey, `不存在的UniKey`, ``)
		return nil
	}

	return xkf_tool.RedisRunList[UniKey]
}

func checkKeyExist(c *gin.Context, redisCli *redis.Client, key string) bool {
	//判断是否存在
	if existInt := redisCli.Exists(key).Val(); existInt <= 0 {
		response(c, xkf_tool.ErrorKeyNotExist, fmt.Sprintf(`%s 不存在`, key), ``)
		return false
	}
	return true
}
