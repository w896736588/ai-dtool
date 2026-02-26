import base from '../base'
import mod from '../module'

//当前连接正常的redis
function RedisAvailableList(callBack) {
    base.BasePost('/api/RedisAvailableList', {}, callBack)
}

//保存字符串
function RedisSaveString(redisChooseConfig, cacheKey, cacheValue, callBack) {
    redisChooseConfig.CacheKey = cacheKey
    redisChooseConfig.Value = cacheValue
    base.BasePost('/api/RedisSaveString', redisChooseConfig, callBack)
}

//删除key
function RedisDelKey(redisChooseConfig, cacheKey, callBack) {
    redisChooseConfig.CacheKey = cacheKey
    base.BasePost('/api/RedisDelKey', redisChooseConfig, callBack)
}

//删除子元素
function RedisDelSub(redisChooseConfig, cacheKey, cacheType, sub, callBack) {
    redisChooseConfig.CacheKey = cacheKey
    redisChooseConfig.CacheType = cacheType
    redisChooseConfig.Sub = sub
    base.BasePost('/api/RedisDelSub', redisChooseConfig, callBack)
}

//修改ttl
function RedisEditTtl(redisChooseConfig, cacheKey, ttl, callBack) {
    redisChooseConfig.CacheKey = cacheKey
    redisChooseConfig.TTL = ttl
    base.BasePost('/api/RedisEditTtl', redisChooseConfig, callBack)
}

//批量删除key
function RedisDelAllKey(redisChooseConfig, cacheKeys, callBack) {
    redisChooseConfig.CacheKeys = cacheKeys
    base.BasePost('/api/RedisDeleteAll', redisChooseConfig, callBack)
}

//创建key
function RedisCreateCache(redisChooseConfig, CacheKey, BoolCreate, CacheType, CacheField, CacheValue, LPushValue, RPushValue, CacheMember, CacheScore, callBack) {
    redisChooseConfig.CacheKey = CacheKey
    redisChooseConfig.BoolCreate = BoolCreate
    redisChooseConfig.CacheType = CacheType
    redisChooseConfig.CacheField = CacheField
    redisChooseConfig.CacheValue = CacheValue
    redisChooseConfig.LPushValue = LPushValue
    redisChooseConfig.RPushValue = RPushValue
    redisChooseConfig.CacheMember = CacheMember
    redisChooseConfig.CacheScore = CacheScore
    base.BasePost('/api/RedisCreateCache', redisChooseConfig, callBack)
}

//编辑子元素
function RedisEditSub(redisChooseConfig, CacheKey, CacheType, CacheField, CacheValue, CacheIndex, CacheScore, CacheMember, callBack) {
    redisChooseConfig.CacheKey = CacheKey
    redisChooseConfig.CacheType = CacheType
    redisChooseConfig.CacheField = CacheField
    redisChooseConfig.CacheValue = CacheValue
    redisChooseConfig.CacheIndex = CacheIndex
    redisChooseConfig.CacheMember = CacheMember
    redisChooseConfig.CacheScore = CacheScore
    base.BasePost('/api/RedisEditSub', redisChooseConfig, callBack)
}

//搜索某个具体的key
function RedisSearch(redisChooseConfig, searchKey, cursor, search , callBack) {
    redisChooseConfig.CacheKey = searchKey
    redisChooseConfig.Cursor = cursor
    redisChooseConfig.Search = search
    base.BasePost('/api/RedisSearch', redisChooseConfig , callBack)
}

//搜索
function RedisKeys(redisChooseConfig, keysResultCursor, search, callBack) {
    redisChooseConfig.Search = search
    redisChooseConfig.Cursor = keysResultCursor
    base.BasePost('/api/RedisKeys', redisChooseConfig , callBack)
}

export default {
    RedisAvailableList,
    RedisSearch,
    RedisKeys,
    RedisSaveString,
    RedisDelKey,
    RedisDelSub,
    RedisEditTtl,
    RedisDelAllKey,
    RedisCreateCache,
    RedisEditSub,
}
