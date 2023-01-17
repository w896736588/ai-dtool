package websocket

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

// DistributeMsg  分发消息
// @author frog
// @date 2022-04-20 10:47:27
func RedisDistributeMsg(reqMsgByte []byte) []byte {
	requestBody := define.WebSocketEventStruct{}
	err := json.Unmarshal(reqMsgByte, &requestBody)
	if err != nil {
		log.Errorf(`解析json失败 %s %s`, err.Error(), cast.ToString(reqMsgByte))
		return []byte(``)
	}
	//如果是ping 那么回复pang
	if requestBody.Event == `ping` {
		return []byte(`pang`)
	} else {
		switch requestBody.Event {
		case `RedisList`:
			return []byte(RedisList())
		}
	}
	return []byte(``)
}

func RedisList() string {
	for key, value := range base.RedisList {
		if _, ok := base.RedisRunList[value.UniKey]; !ok {
			value.Connection = false
		}
		base.RedisList[key] = value
	}
	return helper.JsonEncode(&define.Response{
		Errcode: define.ErrorCodeSuccess,
		Errmsg:  `成功`,
		Data:    base.RedisList,
	})
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
	//拿到key类型
	returnList := make([]define.KeysList, 0)
	for _, cacheKey := range resultMap {
		returnList = append(returnList, define.KeysList{
			CacheKey: cacheKey,
			Type:     ` `,
			Loading:  true,
		})
	}
	response(c, define.ErrorCodeSuccess, `获取成功`, returnList)
}

func KeysType(c *gin.Context) {
	reqBody := &define.SearchKeysTypeBody{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	//拿到key类型
	returnList := make([]define.KeysList, 0)
	for _, cacheKey := range reqBody.KeysList {
		keyType, err := redisCli.Type(cacheKey).Result()
		if err == nil && keyType != `` {
			returnList = append(returnList, define.KeysList{
				CacheKey: cacheKey,
				Type:     keyType,
			})
		}
	}
	response(c, define.ErrorCodeSuccess, `获取成功`, returnList)
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
	if keyType == define.CacheString {
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
	} else if keyType == define.CacheHash {
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
	} else if keyType == define.CacheList {
		var resultArray []string
		resultArray, err = redisCli.LRange(reqBody.CacheKey, 0, 100000).Result()
		if err == redis.Nil {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, define.ErrorCodeRunError, err.Error(), ``)
			return
		}
		response(c, define.ErrorCodeSuccess, `获取成功`, resultArray)
	} else if keyType == define.CacheSet {
		var resultArray []string
		resultArray, err = redisCli.SMembers(reqBody.CacheKey).Result()
		if err == redis.Nil {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, define.ErrorCodeRunError, err.Error(), ``)
			return
		}
		response(c, define.ErrorCodeSuccess, `获取成功`, resultArray)
	} else if keyType == define.CacheZSet {
		var resultArray []redis.Z
		resultArray, err = redisCli.ZRangeWithScores(reqBody.CacheKey, 0, 100000).Result()
		if err == redis.Nil {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经不存在`, reqBody.CacheKey), ``)
			return
		} else if err != nil {
			response(c, define.ErrorCodeRunError, err.Error(), ``)
			return
		}
		response(c, define.ErrorCodeSuccess, `获取成功`, resultArray)
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

func DelSub(c *gin.Context) {
	var err error
	reqBody := &define.DelSub{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		return
	}

	if reqBody.CacheType == define.CacheString {
		response(c, define.ErrorCodeRunError, `不支持字符串`, ``)
	} else if reqBody.CacheType == define.CacheHash {
		err = redisCli.HDel(reqBody.CacheKey, reqBody.Sub).Err()
	} else if reqBody.CacheType == define.CacheList {
		err = redisCli.LRem(reqBody.CacheKey, 0, reqBody.Sub).Err()
	} else if reqBody.CacheType == define.CacheSet {
		err = redisCli.SRem(reqBody.CacheKey, reqBody.Sub).Err()
	} else if reqBody.CacheType == define.CacheZSet {
		err = redisCli.ZRem(reqBody.CacheKey, reqBody.Sub).Err()
	}
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

func DelAllKey(c *gin.Context) {
	var err error
	reqBody := &define.DelAllKey{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	err = redisCli.Del(reqBody.CacheKeys...).Err()
	if err != nil {
		response(c, define.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, define.ErrorCodeSuccess, `删除成功`, ``)
	}
}

func CreateCache(c *gin.Context) {
	var err error
	reqBody := &define.CreateCache{}
	requestData(c, &reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	//判断是否存在
	if reqBody.BoolCreate == 1 {
		if existInt := redisCli.Exists(reqBody.CacheKey).Val(); existInt > 0 {
			response(c, define.ErrorKeyNotExist, fmt.Sprintf(`%s 已经存在`, reqBody.CacheKey), ``)
			return
		}
	} else {
		if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
			return
		}
	}

	if reqBody.CacheType == define.CacheString {
		err = redisCli.Set(reqBody.CacheKey, reqBody.CacheValue, time.Duration(reqBody.TTL)*time.Second).Err()
	} else if reqBody.CacheType == define.CacheHash {
		err = redisCli.HSet(reqBody.CacheKey, reqBody.CacheField, reqBody.CacheValue).Err()
	} else if reqBody.CacheType == define.CacheList {
		if reqBody.LPushValue != `` {
			err = redisCli.LPush(reqBody.CacheKey, reqBody.LPushValue).Err()
		} else if reqBody.RPushValue != `` {
			err = redisCli.RPush(reqBody.CacheKey, reqBody.RPushValue).Err()
		} else {
			err = redisCli.RPush(reqBody.CacheKey, reqBody.CacheValue).Err()
		}
	} else if reqBody.CacheType == define.CacheSet {
		err = redisCli.SAdd(reqBody.CacheKey, reqBody.CacheMember).Err()
	} else if reqBody.CacheType == define.CacheZSet {
		err = redisCli.ZAdd(reqBody.CacheKey, redis.Z{
			Score:  reqBody.CacheScore,
			Member: reqBody.CacheMember,
		}).Err()
	}
	if err != nil {
		response(c, define.ErrorCodeRunError, err.Error(), ``)
	}
	//处理过期时间
	if reqBody.BoolCreate == 1 && reqBody.TTL != 0 {
		err = redisCli.Expire(reqBody.CacheKey, time.Duration(reqBody.TTL)*time.Second).Err()
	}

	if err != nil {
		response(c, define.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, define.ErrorCodeSuccess, `创建成功`, ``)
	}
}

func EditSub(c *gin.Context) {
	var err error
	reqBody := &define.EditSub{}
	requestData(c, &reqBody)
	log.Errorf(`editSub %#v`, reqBody)

	var redisCli *redis.Client
	if redisCli = getRedisClient(c, reqBody.UniKey); redisCli == nil {
		return
	}

	if exist := checkKeyExist(c, redisCli, reqBody.CacheKey); exist == false {
		log.Errorf(`exist %v`, exist)
		return
	}
	if reqBody.CacheType == define.CacheHash {
		err = redisCli.HSet(reqBody.CacheKey, reqBody.CacheField, reqBody.CacheValue).Err()
	} else if reqBody.CacheType == define.CacheList {
		err = redisCli.LSet(reqBody.CacheKey, reqBody.CacheIndex, reqBody.CacheValue).Err()
	} else if reqBody.CacheType == define.CacheZSet {
		err = redisCli.ZAdd(reqBody.CacheKey, redis.Z{
			Score:  reqBody.CacheScore,
			Member: reqBody.CacheMember,
		}).Err()
	}

	if err != nil {
		response(c, define.ErrorCodeRunError, err.Error(), ``)
	} else {
		response(c, define.ErrorCodeSuccess, `编辑成功`, ``)
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
