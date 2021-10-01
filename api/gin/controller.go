package gin

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/techoner/gophp/serialize"
	"io/ioutil"
	"net/http"
	"redis_manager/base"
	"redis_manager/define"
	"redis_manager/helper"
	"time"
)

func RedisList(c *gin.Context) {
	response(c, define.ErrorCodeSuccess, `获取成功`, base.RedisList)
}

func Keys(c *gin.Context) {
	var err error
	reqBody := &define.SearchBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	var resultMap []string
	resultMap, err = redisCli.Keys(reqBody.Search).Result()
	if err == redis.Nil {
		resultMap = make([]string, 0)
	}
	response(c, define.ErrorCodeSuccess, `获取成功`, resultMap)
}

func Search(c *gin.Context) {
	var err error
	reqBody := &define.RequestBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	//找到key是什么类型
	keyType, err := redisCli.Type(reqBody.CacheKey).Result()
	if err != nil || keyType == `` {
		response(c, define.ErrorCodeRunError, `获取元素类型失败`, ``)
		return
	}
	if keyType == `string` {
		var result string
		result, err = redisCli.Get(reqBody.CacheKey).Result()
		if err == redis.Nil {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, define.ErrorCodeRunError, err.Error(), ``)
			return
		}
		response(c, define.ErrorCodeSuccess, `获取成功`, result)
	} else if keyType == `hash` {
		var resultMap map[string]string
		resultMap, err = redisCli.HGetAll(reqBody.CacheKey).Result()
		if err == redis.Nil {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, define.ErrorCodeRunError, err.Error(), ``)
			return
		}
		response(c, define.ErrorCodeSuccess, `获取成功`, resultMap)
	}
}

func GetKeyType(c *gin.Context) {
	var err error
	reqBody := &define.RequestBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	//找到key是什么类型
	keyType, err := redisCli.Type(reqBody.CacheKey).Result()
	if err != nil {
		response(c, define.ErrorCodeRunError, err.Error(), ``)
		return
	} else if keyType == `` {
		response(c, define.ErrorCodeRunError, `获取元素类型失败`, ``)
		return
	}
	//找到过期时间
	var ttl time.Duration
	ttl, err = redisCli.TTL(reqBody.CacheKey).Result()
	if err != nil {
		response(c, define.ErrorCodeRunError, err.Error(), ``)
		return
	}
	response(c, define.ErrorCodeSuccess, `获取类型和过期时间成功`, &define.TypeResponse{
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
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	ttlTime, err := redisCli.TTL(reqBody.CacheKey).Result()
	//永久
	err = redisCli.Set(reqBody.CacheKey, reqBody.Value, ttlTime).Err()
	if err != nil {
		response(c, define.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, define.ErrorCodeSuccess, `保存成功`, ``)
	}
}

func PhpUnSerialize(c *gin.Context) {
	var err error
	reqBody := &define.SerializeBody{}
	requestData(c, &reqBody)

	out, err := serialize.UnMarshal([]byte(reqBody.SerializeStr))
	if err != nil {
		response(c, define.ErrorCodeRunError, err.Error(), reqBody.SerializeStr)
		return
	}
	var returnStr string
	switch out.(type) {
	case string:
		returnStr = cast.ToString(out)
		break
	default:
		jsonStr, _ := json.Marshal(out)
		returnStr = cast.ToString(jsonStr)
		break
	}
	response(c, define.ErrorCodeSuccess, `成功`, returnStr)
}

func DelKey(c *gin.Context) {
	var err error
	reqBody := &define.DelKey{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	//永久
	err = redisCli.Del(reqBody.CacheKey).Err()
	if err != nil {
		response(c, define.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, define.ErrorCodeSuccess, `删除成功`, ``)
	}
}

func EditTtl(c *gin.Context) {
	var err error
	reqBody := &define.EditTTL{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	dru := time.Duration(reqBody.TTL) * time.Second
	err = redisCli.Expire(reqBody.CacheKey, dru).Err()
	if err != nil {
		response(c, define.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, define.ErrorCodeSuccess, `设置成功`, ``)
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
	returnJson := helper.JsonEncode(&define.Response{
		Errcode: errcode,
		Errmsg:  errmsg,
		Data:    body,
	})
	c.String(http.StatusOK, returnJson)
}

func getRedisClient(c *gin.Context, UniKey string) *redis.Client {
	if ok := base.RedisRunList[UniKey]; ok == nil {
		response(c, define.ErrorCodeErrorUniKey, `不存在的UniKey`, ``)
		return nil
	}

	return base.RedisRunList[UniKey]
}

func checkKeyExist(c *gin.Context, redisCli *redis.Client, key string) bool {
	//判断是否存在
	if existInt := redisCli.Exists(key).Val(); existInt <= 0 {
		response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 不存在`, key), ``)
		return false
	}
	return true
}
