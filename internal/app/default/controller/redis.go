package controller

import (
	"context"
	"dev_tool/base"
	"dev_tool/internal/pkg/define"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"time"
)

// RedisAvailableList 拿到可用的redis列表
func RedisAvailableList(c *gin.Context) {
	runList := make([]map[string]any, 0)
	redisConfigList, redisConfigListErr := base.Component.TSqlite.GetAllRedisConfig()
	if redisConfigListErr != nil {
		gsgin.GinResponseError(c, redisConfigListErr.Error(), nil)
		return
	}
	for _, redisConfig := range redisConfigList {
		//_, clientErr := base.Component.TRedis.GetClient(redisConfig)
		//if clientErr != nil {
		//	base.Component.GsLog.Errof(`获取redis连接失败 %s`, clientErr.Error())
		//	continue
		//}
		runList = append(runList, map[string]any{
			`name`: redisConfig[`name`],
			`id`:   cast.ToString(redisConfig[`id`]),
		})
	}
	gsgin.GinResponseSuccess(c, ``, runList)
}

// RedisKeys 搜索key
func RedisKeys(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	search := reqMap[`Search`]
	if search == nil {
		gsgin.GinResponseError(c, `缺少搜索内容参数`, reqMap)
		return
	}
	cursor := cast.ToUint64(reqMap[`Cursor`])
	var resultMap []string
	resultMap, resultCursor, err := redisCli.Client.Scan(context.Background(), cursor, cast.ToString(search), 1000).Result()
	if errors.Is(err, redis.Nil) {
		resultMap = make([]string, 0)
	} else if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	//拿到key类型
	returnList := make([]map[string]any, 0)
	for _, cacheKey := range resultMap {
		returnList = append(returnList, map[string]any{
			`CacheKey`: cacheKey,
			`Type`:     ` `,
			`Loading`:  true,
		})
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`:   returnList,
		`cursor`: resultCursor,
	})
}

// RedisSearch 获取一个key的明细
func RedisSearch(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	cursor := cast.ToUint64(reqMap[`Cursor`])
	cacheKey := cast.ToString(reqMap[`CacheKey`])
	search := cast.ToString(reqMap[`Search`])
	if cacheKey == `` {
		gsgin.GinResponseError(c, `缺少搜索的key`, nil)
		return
	}
	//找到key是什么类型
	keyType, err := redisCli.Client.Type(context.Background(), cacheKey).Result()
	keyTtl, _ := redisCli.Client.TTL(context.Background(), cacheKey).Result()
	if err != nil {
		gsgin.GinResponseSuccess(c, err.Error(), ``)
		return
	}
	if keyType == `` {
		gsgin.GinResponseError(c, `缓存已不存在`, ``)
		return
	}
	var resultCursor uint64
	var length int64
	var isMore = 0 //1还有更多 0没有更多了
	var maxQuery = 200
	//通用的返回结果
	var gsCons *gstool.GsCons
	if keyType == gsdb.RedisKeyString {
		var result string
		result, err = redisCli.Client.Get(context.Background(), cacheKey).Result()
		gsCons = gstool.ConsNew(result)
	} else if keyType == gsdb.RedisKeyHash {
		var resultList []string
		resultMap := make(map[string]any)
		resultList, resultCursor, err = redisCli.Client.HScan(context.Background(), cacheKey, cursor, `*`+search+`*`, int64(maxQuery)).Result()
		for key := 0; key < len(resultList); key++ {
			resultMap[resultList[key]] = resultList[key+1]
			key++
		}
		//没有更多
		if len(resultMap) >= maxQuery {
			isMore = 1
		}
		length = redisCli.Client.HLen(context.Background(), cacheKey).Val()
		gsCons = gstool.ConsNew(resultMap)
	} else if keyType == gsdb.RedisKeyList {
		var resultArray []string
		resultArray, err = redisCli.Client.LRange(context.Background(), cacheKey, cast.ToInt64(cursor), int64(cast.ToInt(cursor)+maxQuery-1)).Result()
		if len(resultArray) >= maxQuery {
			isMore = 1
		}
		resultCursor = cursor + cast.ToUint64(maxQuery)
		length = redisCli.Client.LLen(context.Background(), cacheKey).Val()
		gsCons = gstool.ConsNew(resultArray)
	} else if keyType == gsdb.RedisKeySet {
		var resultArray []string
		resultArray, resultCursor, err = redisCli.Client.SScan(context.Background(), cacheKey, cursor, `*`+search+`*`, int64(maxQuery)).Result()
		//没有更多
		if len(resultArray) >= maxQuery {
			isMore = 1
		}
		length = redisCli.Client.SCard(context.Background(), cacheKey).Val()
		gsCons = gstool.ConsNew(resultArray)
	} else if keyType == gsdb.RedisKeyZSet {
		var resultArray []redis.Z
		resultArray, err = redisCli.Client.ZRangeWithScores(context.Background(), cacheKey, cast.ToInt64(cursor), int64(cast.ToInt(cursor)+maxQuery-1)).Result()
		if len(resultArray) >= maxQuery {
			isMore = 1
		}
		resultCursor = cursor + cast.ToUint64(maxQuery)
		length = redisCli.Client.ZCard(context.Background(), cacheKey).Val()
		gsCons = gstool.ConsNew(resultArray)
	} else {
		gsgin.GinResponseError(c, `暂不支持的缓存类型 `+keyType, ``)
		return
	}
	if errors.Is(err, redis.Nil) {
		gsgin.GinResponseError(c, fmt.Sprintf(`%s 已经不存在`, cacheKey), ``)
		return
	} else if err != nil {
		gsgin.GinResponseError(c, err.Error(), ``)
		return
	}
	if gsCons == nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`不支持的类型 %s`, keyType), ``)
	} else {
		gsgin.GinResponseSuccess(c, `获取成功`, map[string]interface{}{
			`keyType`: keyType,
			`KeyTtl`:  keyTtl.Seconds(),
			`Result`:  gsCons.Value(),
			`Cursor`:  resultCursor,
			`Length`:  length,
			`IsMore`:  isMore,
		})
	}

}

// RedisKeysType 获取redis的key类型
func RedisKeysType(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	keyList := gstool.ArrayGetFromAny(reqMap[`KeyList`])
	if len(*keyList) == 0 {
		gsgin.GinResponseError(c, `缺少keyList参数`, nil)
		return
	}
	//拿到key类型
	returnList := make([]map[string]interface{}, 0)
	for _, cacheKey := range *keyList {
		keyType, keyTypeErr := redisCli.Client.Type(context.Background(), cacheKey.ToStr()).Result()
		if keyTypeErr == nil && keyType != `` {
			returnList = append(returnList, map[string]interface{}{
				`CacheKey`: cacheKey.ToStr(),
				`Type`:     keyType,
			})
		}
	}
	gsgin.GinResponseSuccess(c, `获取成功`, returnList)
}

// RedisKeyType 获取单个key类型
func RedisKeyType(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	existErr := BaseRedisCheckKeyExist(redisCli.Client, cast.ToString(reqMap[`CacheKey`]))
	if existErr != nil {
		BaseResponseByError(c, err)
		return
	}
	cacheKey := cast.ToString(reqMap[`CacheKey`])
	//找到key是什么类型
	keyType, keyTypeErr := redisCli.Client.Type(context.Background(), cacheKey).Result()
	if keyTypeErr != nil {
		gsgin.GinResponseError(c, keyTypeErr.Error(), nil)
		return
	} else if keyType == `` {
		gsgin.GinResponseError(c, `获取元素类型失败`, nil)
		return
	}
	//找到过期时间
	ttl, ttlErr := redisCli.Client.TTL(context.Background(), cacheKey).Result()
	if ttlErr != nil {
		gsgin.GinResponseError(c, ttlErr.Error(), ``)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]interface{}{
		`Type`: keyType,
		`TTL`:  cast.ToInt(ttl.Seconds()),
	})
}

// RedisSaveString 保存字符串值
func RedisSaveString(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
	if err != nil {
		BaseResponseByError(c, err)
		return
	}

	ttlTime, err := redisCli.Client.TTL(context.Background(), cacheKey).Result()
	//永久
	err = redisCli.Client.Set(context.Background(), cacheKey, reqMap[`Value`], ttlTime).Err()
	BaseResponseByError(c, err)
}

// RedisDelKey 删除key
func RedisDelKey(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	//永久
	err = redisCli.Client.Del(context.Background(), cacheKey).Err()
	BaseResponseByError(c, err)
}

// RedisDelSub 删除子元素
func RedisDelSub(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	cacheType := cast.ToString(reqMap[`CacheType`])
	if cacheType == `` {
		BaseResponseByError(c, errors.New(`cacheType类型不能为空`))
		return
	}
	subKey := cast.ToString(reqMap[`Sub`])
	if subKey == `` {
		BaseResponseByError(c, errors.New(`Sub 类型不能为空`))
		return
	}
	if cacheType == define.CacheString {
		gsgin.GinResponseError(c, `不支持字符串`, ``)
		return
	} else if cacheType == define.CacheHash {
		err = redisCli.Client.HDel(context.Background(), cacheKey, subKey).Err()
	} else if cacheType == define.CacheList {
		err = redisCli.Client.LRem(context.Background(), cacheKey, 0, subKey).Err()
	} else if cacheType == define.CacheSet {
		err = redisCli.Client.SRem(context.Background(), cacheKey, subKey).Err()
	} else if cacheType == define.CacheZSet {
		err = redisCli.Client.ZRem(context.Background(), cacheKey, subKey).Err()
	}
	BaseResponseByError(c, err)
}

// RedisEditTtl 修改过期时间
func RedisEditTtl(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	ttl := cast.ToInt(reqMap[`TTL`])
	dru := time.Duration(ttl) * time.Second
	err = redisCli.Client.Expire(context.Background(), cacheKey, dru).Err()
	BaseResponseByError(c, err)
}

// RedisDelAllKey 批量删除key
func RedisDelAllKey(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	cacheKeyList := reqMap[`CacheKeys`].([]interface{})
	cacheKeyStrList := make([]string, 0)
	for _, v := range cacheKeyList {
		cacheKeyStrList = append(cacheKeyStrList, cast.ToString(v))
	}
	err = redisCli.Client.Del(context.Background(), cacheKeyStrList...).Err()
	BaseResponseByError(c, err)
}

// RedisCreateCache 新增key
func RedisCreateCache(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	//判断是否存在
	if cast.ToInt(reqMap[`BoolCreate`]) == 1 {
		if existInt := redisCli.Client.Exists(context.Background(), cacheKey).Val(); existInt > 0 {
			gsgin.GinResponseError(c, fmt.Sprintf(`%s 已经存在`, cacheKey), ``)
			return
		}
	} else {
		err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
		if err != nil {
			gsgin.GinResponseError(c, fmt.Sprintf(`%s key不存在`, cacheKey), ``)
			return
		}
	}

	cacheType := cast.ToString(reqMap[`CacheType`])
	if cacheType == gsdb.RedisKeyString {
		err = redisCli.Client.Set(context.Background(), cacheKey, reqMap[`CacheValue`], time.Duration(cast.ToInt64(reqMap[`TTL`]))*time.Second).Err()
	} else if cacheType == gsdb.RedisKeyHash {
		err = redisCli.Client.HSet(context.Background(), cacheKey, cast.ToString(reqMap[`CacheField`]), cast.ToString(reqMap[`CacheValue`])).Err()
	} else if cacheType == gsdb.RedisKeyList {
		if cast.ToString(reqMap[`LPushValue`]) != `` {
			err = redisCli.Client.LPush(context.Background(), cacheKey, cast.ToString(reqMap[`LPushValue`])).Err()
		} else if cast.ToString(reqMap[`RPushValue`]) != `` {
			err = redisCli.Client.RPush(context.Background(), cacheKey, cast.ToString(reqMap[`RPushValue`])).Err()
		} else {
			err = redisCli.Client.RPush(context.Background(), cacheKey, cast.ToString(reqMap[`CacheValue`])).Err()
		}
	} else if cacheType == gsdb.RedisKeySet {
		err = redisCli.Client.SAdd(context.Background(), cacheKey, cast.ToString(reqMap[`CacheMember`])).Err()
	} else if cacheType == gsdb.RedisKeyZSet {
		err = redisCli.Client.ZAdd(context.Background(), cacheKey, redis.Z{
			Score:  cast.ToFloat64(reqMap[`CacheScore`]),
			Member: cast.ToString(reqMap[`CacheMember`]),
		}).Err()
	}
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), ``)
		return
	}
	//处理过期时间
	if cast.ToInt(reqMap[`BoolCreate`]) == 1 && cast.ToInt(reqMap[`TTL`]) != 0 {
		err = redisCli.Client.Expire(context.Background(), cacheKey, time.Duration(cast.ToInt(reqMap[`TTL`]))*time.Second).Err()
	}
	BaseResponseByError(c, err)
}

// RedisEditSub 编辑子元素
func RedisEditSub(c *gin.Context) {
	reqMap, redisCli, err := getRedisComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	cacheType := cast.ToString(reqMap[`CacheType`])
	cacheField := cast.ToString(reqMap[`CacheField`])
	cacheValue := cast.ToString(reqMap[`CacheValue`])
	cacheIndex := cast.ToInt64(reqMap[`CacheIndex`])
	if cacheType == gsdb.RedisKeyHash {
		err = redisCli.Client.HSet(context.Background(), cacheKey, cacheField, cacheValue).Err()
	} else if cacheType == gsdb.RedisKeyList {
		err = redisCli.Client.LSet(context.Background(), cacheKey, cacheIndex, cacheValue).Err()
	} else if cacheType == gsdb.RedisKeyZSet {
		err = redisCli.Client.ZAdd(context.Background(), cacheKey, redis.Z{
			Score:  cast.ToFloat64(reqMap[`CacheScore`]),
			Member: cast.ToString(reqMap[`CacheMember`]),
		}).Err()
	}
	BaseResponseByError(c, err)
}

func getRedisComponent(c *gin.Context) (map[string]interface{}, *gsdb.GsRedis, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, nil, err
	}
	redisId := reqMap[`id`]
	if cast.ToString(redisId) == `` {
		return reqMap, nil, errors.New(`缺少id参数`)
	}
	redisConfig, redisConfigErr := base.Component.TSqlite.GetRedisConfig(redisId)
	if redisConfigErr != nil {
		return reqMap, nil, redisConfigErr
	}
	redisClient, redisClientErr := base.Component.TRedis.GetClient(redisConfig)
	if redisClientErr != nil {
		return reqMap, nil, redisClientErr
	}
	return reqMap, redisClient, nil
}
