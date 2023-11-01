import base from './base'
import mod from './module'

//当前连接正常的redis
function RedisAvailableList(callBack) {
  base.BasePost('/api/RedisAvailableList', {}, function (response) {
    callBack(response)
  })
}

//保存字符串
function RedisSaveString(redisName , cacheKey, cacheValue, callBack) {
  base.BasePost('/api/RedisSaveString', {
    RedisName: redisName,
    CacheKey: cacheKey,
    Value: cacheValue,
  }, function (response) {
    callBack(response)
  })
}

//删除key
function RedisDelKey(redisName , cacheKey, callBack) {
  base.BasePost('/api/RedisDelKey', {
    RedisName: redisName,
    CacheKey: cacheKey,
  }, function (response) {
    callBack(response)
  })
}

//删除子元素
function RedisDelSub(redisName , cacheKey, cacheType, sub, callBack) {
  base.BasePost('/api/RedisDelSub', {
    RedisName: redisName,
    CacheKey: cacheKey,
    CacheType: cacheType,
    Sub: sub,
  }, function (response) {
    callBack(response)
  })
}

//修改ttl
function RedisEditTtl(redisName , cacheKey, ttl, callBack) {
  base.BasePost('/api/RedisEditTtl', {
    RedisName: redisName,
    CacheKey: cacheKey,
    TTL: ttl,
  }, function (response) {
    callBack(response)
  })
}

//批量删除key
function RedisDelAllKey(redisName , cacheKeys, callBack) {
  base.BasePost('/api/RedisDeleteAll', {
    RedisName: redisName,
    CacheKeys: cacheKeys,
  }, function (response) {
    callBack(response)
  })
}

//创建key
function RedisCreateCache(redisName , CacheKey, BoolCreate, CacheType, CacheField,
                          CacheValue, LPushValue, RPushValue, CacheMember, CacheScore, callBack) {
  base.BasePost('/api/RedisCreateCache', {
    RedisName: redisName,
    CacheKey: CacheKey,
    BoolCreate: BoolCreate,
    CacheType: CacheType,
    CacheField: CacheField,
    CacheValue: CacheValue,
    LPushValue: LPushValue,
    RPushValue: RPushValue,
    CacheMember: CacheMember,
    CacheScore: CacheScore,
  }, function (response) {
    callBack(response)
  })
}

//编辑子元素
function RedisEditSub(redisName , CacheKey, CacheType, CacheField, CacheValue, CacheIndex, CacheScore, CacheMember, callBack) {
  base.BasePost('/api/RedisEditSub', {
    RedisName: redisName,
    CacheKey: CacheKey,
    CacheType: CacheType,
    CacheField: CacheField,
    CacheValue: CacheValue,
    CacheIndex: CacheIndex,
    CacheMember: CacheMember,
    CacheScore: CacheScore,
  }, function (response) {
    callBack(response)
  })
}

//根据redis名字列表拿到配置
function GetRedisConfigListByNameList(nameList) {
  let redisConfigList = mod.GetRedisConfigList()
  let returnConfigList = []
  for (let i in nameList) {
    for (let j in redisConfigList) {
      if (redisConfigList[j].name === nameList[i]) {
        returnConfigList.push(redisConfigList[j])
      }
    }
  }
  return returnConfigList;
}

//搜索某个具体的key
function RedisSearch(redisName, searchKey, callBack) {
  base.BasePost('/api/RedisSearch', {
    RedisName: redisName,
    CacheKey: searchKey,
  }, function (response) {
    callBack(response)
  })
}

//搜索
function RedisKeys(redisName, search, callBack) {
  base.BasePost('/api/RedisKeys', {
    RedisName: redisName,
    Search: search,
  }, function (response) {
    callBack(response)
  })
}

//获取key的类型
function RedisKeyType(redisName, search, callBack) {
  base.BasePost('/api/RedisKeyType', {
    RedisName: redisName,
    CacheKey: search,
  }, function (response) {
    callBack(response)
  })
}

export default {
  RedisAvailableList,
  GetRedisConfigListByNameList,
  RedisSearch,
  RedisKeys,
  RedisKeyType,
  RedisSaveString,
  RedisDelKey,
  RedisDelSub,
  RedisEditTtl,
  RedisDelAllKey,
  RedisCreateCache,
  RedisEditSub,
}
