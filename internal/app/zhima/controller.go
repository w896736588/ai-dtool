package zhima

import (
	"context"
	"dev_tool/internal/app/zhima/define"
	"encoding/json"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsdefine"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io/ioutil"
	"net/http"
	"time"
)

var RedisHandleList []gsdb.RedisConfig

func RedisList(c *gin.Context) {
	reqBody := &define.SshExec{}
	requestData(c, &reqBody)

	for _, value := range reqBody.RedisConfigList {
		if RedisRunList[value.Name] == nil {
			//初始化链接
			gsRedisConfig := &gsdb.RedisConfig{
				Name:        value.Name,
				Host:        value.Host,
				Password:    value.Password,
				PoolSize:    value.PoolSize,
				Default:     0,
				DialTimeout: 1,
				MaxLifetime: 3600,
				IdleTimeout: 300,
			}
			gsRedis := gsdb.GsRedis{
				RedisConfig: gsRedisConfig,
			}
			err := gsRedis.CreateConn()
			if err != nil {
				continue
			}
			RedisRunList[value.Name] = &gsRedis
		}
	}
	RedisHandleList = make([]gsdb.RedisConfig, 0)
	for _, value := range reqBody.RedisConfigList {
		if RedisRunList[value.Name] != nil {
			RedisHandleList = append(RedisHandleList, value)
		}

	}
	response(c, gsdefine.Success, `获取成功`, RedisHandleList)
}

func Keys(c *gin.Context) {
	var err error
	reqBody := &define.SearchBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	var resultMap []string
	resultMap, err = redisCli.Keys(context.Background(), reqBody.Search).Result()
	if err == redis.Nil {
		resultMap = make([]string, 0)
	}
	//拿到key类型
	returnList := make([]define.KeysList, 0)
	for _, cacheKey := range resultMap {
		returnList = append(returnList, define.KeysList{
			CacheKey: cacheKey,
			Type:     ` `,
			Loading:  true,
		})
	}
	response(c, gsdefine.Success, `获取成功`, returnList)
}

func KeysType(c *gin.Context) {
	reqBody := &define.SearchKeysTypeBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	//拿到key类型
	returnList := make([]define.KeysList, 0)
	for _, cacheKey := range reqBody.KeysList {
		keyType, err := redisCli.Type(context.Background(), cacheKey).Result()
		if err == nil && keyType != `` {
			returnList = append(returnList, define.KeysList{
				CacheKey: cacheKey,
				Type:     keyType,
			})
		}
	}
	response(c, gsdefine.Success, `获取成功`, returnList)
}

func Search(c *gin.Context) {
	var err error
	reqBody := &define.RequestBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	//找到key是什么类型
	keyType, err := redisCli.Type(context.Background(), reqBody.CacheKey).Result()
	if err != nil || keyType == `` {
		response(c, gsdefine.Error, `获取元素类型失败`, ``)
		return
	}
	if keyType == define.CacheString {
		var result string
		result, err = redisCli.Get(context.Background(), reqBody.CacheKey).Result()
		if err == redis.Nil {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, gsdefine.Error, err.Error(), ``)
			return
		}
		response(c, gsdefine.Success, `获取成功`, result)
	} else if keyType == define.CacheHash {
		var resultMap map[string]string
		resultMap, err = redisCli.HGetAll(context.Background(), reqBody.CacheKey).Result()
		if err == redis.Nil {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, gsdefine.Error, err.Error(), ``)
			return
		}
		response(c, gsdefine.Success, `获取成功`, resultMap)
	} else if keyType == define.CacheList {
		var resultArray []string
		resultArray, err = redisCli.LRange(context.Background(), reqBody.CacheKey, 0, 100000).Result()
		if err == redis.Nil {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, gsdefine.Error, err.Error(), ``)
			return
		}
		response(c, gsdefine.Success, `获取成功`, resultArray)
	} else if keyType == define.CacheSet {
		var resultArray []string
		resultArray, err = redisCli.SMembers(context.Background(), reqBody.CacheKey).Result()
		if err == redis.Nil {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, gsdefine.Error, err.Error(), ``)
			return
		}
		response(c, gsdefine.Success, `获取成功`, resultArray)
	} else if keyType == define.CacheZSet {
		var resultArray []redis.Z
		resultArray, err = redisCli.ZRangeWithScores(context.Background(), reqBody.CacheKey, 0, 100000).Result()
		if err == redis.Nil {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, gsdefine.Error, err.Error(), ``)
			return
		}
		response(c, gsdefine.Success, `获取成功`, resultArray)
	}
}

func GetKeyType(c *gin.Context) {
	var err error
	reqBody := &define.RequestBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	//找到key是什么类型
	keyType, err := redisCli.Type(context.Background(), reqBody.CacheKey).Result()
	if err != nil {
		response(c, gsdefine.Error, err.Error(), ``)
		return
	} else if keyType == `` {
		response(c, gsdefine.Error, `获取元素类型失败`, ``)
		return
	}
	//找到过期时间
	var ttl time.Duration
	ttl, err = redisCli.TTL(context.Background(), reqBody.CacheKey).Result()
	if err != nil {
		response(c, gsdefine.Error, err.Error(), ``)
		return
	}
	response(c, gsdefine.Success, `获取类型和过期时间成功`, &define.TypeResponse{
		Type: keyType,
		TTL:  cast.ToInt(ttl.Seconds()),
	})
}

func PhpSerialize(c *gin.Context) {

}

func SaveString(c *gin.Context) {
	var err error
	reqBody := &define.SaveString{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	ttlTime, err := redisCli.TTL(context.Background(), reqBody.CacheKey).Result()
	//永久
	err = redisCli.Set(context.Background(), reqBody.CacheKey, reqBody.Value, ttlTime).Err()
	if err != nil {
		response(c, gsdefine.Error, err.Error(), ``)
	} else {
		response(c, gsdefine.Success, `保存成功`, ``)
	}
}

func PhpUnSerialize(c *gin.Context) {
	var err error
	reqBody := &define.SerializeBody{}
	requestData(c, &reqBody)
	var out interface{}
	out, err = gstool.PhpUnSerialize(reqBody.SerializeStr)
	if err != nil {
		response(c, gsdefine.Error, err.Error(), reqBody.SerializeStr)
		return
	}
	response(c, gsdefine.Success, `成功`, out)
}

func DelKey(c *gin.Context) {
	var err error
	reqBody := &define.DelKey{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	//永久
	err = redisCli.Del(context.Background(), reqBody.CacheKey).Err()
	if err != nil {
		response(c, gsdefine.Error, err.Error(), ``)
	} else {
		response(c, gsdefine.Success, `删除成功`, ``)
	}
}

func DelSub(c *gin.Context) {
	var err error
	reqBody := &define.DelSub{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	if reqBody.CacheType == define.CacheString {
		response(c, gsdefine.Error, `不支持字符串`, ``)
	} else if reqBody.CacheType == define.CacheHash {
		err = redisCli.HDel(context.Background(), reqBody.CacheKey, reqBody.Sub).Err()
	} else if reqBody.CacheType == define.CacheList {
		err = redisCli.LRem(context.Background(), reqBody.CacheKey, 0, reqBody.Sub).Err()
	} else if reqBody.CacheType == define.CacheSet {
		err = redisCli.SRem(context.Background(), reqBody.CacheKey, reqBody.Sub).Err()
	} else if reqBody.CacheType == define.CacheZSet {
		err = redisCli.ZRem(context.Background(), reqBody.CacheKey, reqBody.Sub).Err()
	}
	if err != nil {
		response(c, gsdefine.Error, err.Error(), ``)
	} else {
		response(c, gsdefine.Success, `删除成功`, ``)
	}
}

func EditTtl(c *gin.Context) {
	var err error
	reqBody := &define.EditTTL{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	dru := time.Duration(reqBody.TTL) * time.Second
	err = redisCli.Expire(context.Background(), reqBody.CacheKey, dru).Err()
	if err != nil {
		response(c, gsdefine.Error, err.Error(), ``)
	} else {
		response(c, gsdefine.Success, `设置成功`, ``)
	}
}

func DelAllKey(c *gin.Context) {
	var err error
	reqBody := &define.DelAllKey{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	err = redisCli.Del(context.Background(), reqBody.CacheKeys...).Err()
	if err != nil {
		response(c, gsdefine.Error, err.Error(), ``)
	} else {
		response(c, gsdefine.Success, `删除成功`, ``)
	}
}

func CreateCache(c *gin.Context) {
	var err error
	reqBody := &define.CreateCache{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey).Client; redisCli == nil {
		return
	}

	//判断是否存在
	if reqBody.BoolCreate == 1 {
		if existInt := redisCli.Exists(context.Background(), reqBody.CacheKey).Val(); existInt > 0 {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经存在`, reqBody.CacheKey), ``)
			return
		}
	} else {
		if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
			return
		}
	}

	if reqBody.CacheType == define.CacheString {
		err = redisCli.Set(context.Background(), reqBody.CacheKey, reqBody.CacheValue, time.Duration(reqBody.TTL)*time.Second).Err()
	} else if reqBody.CacheType == define.CacheHash {
		err = redisCli.HSet(context.Background(), reqBody.CacheKey, reqBody.CacheField, reqBody.CacheValue).Err()
	} else if reqBody.CacheType == define.CacheList {
		if reqBody.LPushValue != `` {
			err = redisCli.LPush(context.Background(), reqBody.CacheKey, reqBody.LPushValue).Err()
		} else if reqBody.RPushValue != `` {
			err = redisCli.RPush(context.Background(), reqBody.CacheKey, reqBody.RPushValue).Err()
		} else {
			err = redisCli.RPush(context.Background(), reqBody.CacheKey, reqBody.CacheValue).Err()
		}
	} else if reqBody.CacheType == define.CacheSet {
		err = redisCli.SAdd(context.Background(), reqBody.CacheKey, reqBody.CacheMember).Err()
	} else if reqBody.CacheType == define.CacheZSet {
		err = redisCli.ZAdd(context.Background(), reqBody.CacheKey, redis.Z{
			Score:  reqBody.CacheScore,
			Member: reqBody.CacheMember,
		}).Err()
	}
	if err != nil {
		response(c, gsdefine.Error, err.Error(), ``)
	}
	//处理过期时间
	if reqBody.BoolCreate == 1 && reqBody.TTL != 0 {
		err = redisCli.Expire(context.Background(), reqBody.CacheKey, time.Duration(reqBody.TTL)*time.Second).Err()
	}

	if err != nil {
		response(c, gsdefine.Error, err.Error(), ``)
	} else {
		response(c, gsdefine.Success, `创建成功`, ``)
	}
}

func EditSub(c *gin.Context) {
	var err error
	reqBody := &define.EditSub{}
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
	if reqBody.CacheType == define.CacheHash {
		err = redisCli.HSet(context.Background(), reqBody.CacheKey, reqBody.CacheField, reqBody.CacheValue).Err()
	} else if reqBody.CacheType == define.CacheList {
		err = redisCli.LSet(context.Background(), reqBody.CacheKey, reqBody.CacheIndex, reqBody.CacheValue).Err()
	} else if reqBody.CacheType == define.CacheZSet {
		err = redisCli.ZAdd(context.Background(), reqBody.CacheKey, redis.Z{
			Score:  reqBody.CacheScore,
			Member: reqBody.CacheMember,
		}).Err()
	}

	if err != nil {
		response(c, gsdefine.Error, err.Error(), ``)
	} else {
		response(c, gsdefine.Success, `编辑成功`, ``)
	}
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
	returnJson := gstool.JsonEncode(&define.Response{
		Errcode: errcode,
		Errmsg:  errmsg,
		Data:    body,
	})
	c.String(http.StatusOK, returnJson)
}

func getRedisClient(c *gin.Context, UniKey string) *gsdb.GsRedis {
	if ok := RedisRunList[UniKey]; ok == nil {
		response(c, define.ErrorCodeErrorUniKey, `不存在的UniKey`, ``)
		return nil
	}

	return RedisRunList[UniKey]
}

func checkKeyExist(c *gin.Context, redisCli *redis.Client, key string) bool {
	//判断是否存在
	if existInt := redisCli.Exists(context.Background(), key).Val(); existInt <= 0 {
		response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 不存在`, key), ``)
		return false
	}
	return true
}
