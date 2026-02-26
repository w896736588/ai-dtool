<template>
  <div id="mainCard" v-loading="load.redisList" style="text-align: center;">
    <el-input v-model="keys" placeholder="请输入key" style="width: 600px" @keyup.enter="keysSearch"></el-input>
    <el-select v-model="redisChooseId" placeholder="选择" style="margin-left:5px;width: 200px" @change="redisDbChange">
      <el-option v-for="(value,key) in redisList" :key="value.name" :label="value.name" :value="value.id">
      </el-option>
    </el-select>
    &nbsp;
    <el-button v-loading="load.keysSearch" type="primary" @click="keysSearch">搜索</el-button>
    <el-button v-if="keys !== ''" type="primary" @click="setCacheHistory({ cacheKey : keys})">收藏</el-button>
    <el-button key="primary" type="primary" @click="$refs.redisStarRecord.showStarList();">
      收藏列表
    </el-button>
  </div>
  <div v-if="searchHistory.length > 0" class="search-history-container">
      <div class="search-history-list">
        <div v-for="(item, index) in searchHistory" :key="index" class="search-history-item">
          <span class="search-history-text" @click="handleHistorySearch(item.key)">{{ item.key }}</span>
          <span class="search-history-delete" @click="removeSearchHistory(index)">
            <el-icon><Close /></el-icon>
          </span>
        </div>
      </div>
    </div>
  <br/>
  <el-row :gutter="24">
    <el-col :span="8">
      <div :style="{ height: (scrollHeight - 15) + 'px' }" class="box-card">
        <el-button v-if="keysResult && keysResult.length > 0" size="small" type="danger" @click="delAll">
          删除所有({{ searchNum }})
        </el-button>
        <el-button v-if="keysResult && keysResult.length > 0" size="small" type="primary" @click="boolSimpleShow = !boolSimpleShow;changeSimpleShow(boolSimpleShow);">
          <template v-if="boolSimpleShow">
            取消优化
          </template>
          <template v-if="!boolSimpleShow">
            优化显示
          </template>
        </el-button>
        <p></p>
        <div :style="{ height: (scrollHeight - 65) + 'px' }" class="grid-content bg-purple">
          <el-input v-if="keysResult.length > 0" v-model="filterValue" placeholder="输入搜索过滤,空格多个条件" size="default" style="width: 99%" type="text" @input="filterList">
          </el-input>
          <el-scrollbar ref="scrollbarRef" @keydown="keyUpKeys" tabindex="0">
            <div v-if="keysResultCursor !== 0" style="text-align:center;position: relative;margin:2px;bottom:0px;background-color: #409eff;padding:5px;color:white;font-size:11px;width:100%;" @click="keysSearch(true)">
              加载更多
            </div>
            <template v-for="(value, key) in filterKeysResult" :key="key" >
              <p v-if="selectRedisKey === value.CacheKey" class="scrollbar-demo-item scrollbar-p-active" @click="callRefresh(value.CacheKey)">
                {{ value.showName }}</p>
              <p v-else class="scrollbar-demo-item scrollbar-p-default" @click="callRefresh(value.CacheKey)">
                {{ value.showName }}</p>
            </template>
            <p v-if="!keysResult || keysResult.length === 0" class="scrollbar-demo-item" style="color:#ccc;text-align: center;">
              暂无数据</p>
          </el-scrollbar>

        </div>
      </div>
    </el-col>
    <el-col :span="16">
      <div class="box-card" style="text-align: left;height:500px;">
        <el-form ref="form" v-loading="load.callRefresh">
          <redisHashList ref="redisHashList" :callMoreList="callMoreList" :callRefresh="callRefresh" :star="setCacheHistory"></redisHashList>
        </el-form>
      </div>
    </el-col>
  </el-row>
  <!--  收藏列表-->
  <redisStarRecord ref="redisStarRecord" :callStarListSearch="callStarListSearch"></redisStarRecord>
</template>
<style>
.box-card .el-tag-he {
  margin-left: 5px;
  font-size: 13px;
}

.search-history-container {
  margin-top: 10px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.search-history-title {
  font-size: 14px;
  font-weight: bold;
  color: #606266;
  margin-bottom: 8px;
}

.search-history-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.search-history-item {
  display: flex;
  align-items: center;
  padding: 5px 10px;
  background-color: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 13px;
  color: #606266;
  transition: all 0.3s;
}

.search-history-item:hover {
  border-color: #409eff;
  color: #409eff;
}

.search-history-text {
  cursor: pointer;
  margin-right: 8px;
}

.search-history-delete {
  cursor: pointer;
  color: #909399;
  font-size: 12px;
  padding: 2px;
  border-radius: 50%;
  transition: all 0.3s;
}

.search-history-delete:hover {
  color: #f56c6c;
  background-color: #fef0f0;
}

.scrollbar-demo-item {
  display: flex;
  margin: 10px;
  cursor: default;
  border-radius: 4px;
  font-size: 14px;
  line-height: 0.7;
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif !important;
  padding: 5px;
}

.scrollbar-p-default {
  color: #409eff;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

.scrollbar-p-active {
  color: red;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

.cache-table {
  width: 100%;
  font-size: 14px;
  margin-top: 10px
}
</style>

<script>
import JsonViewer from 'vue3-json-viewer'
import Clipboard from 'clipboard'
import Textarea from './base/textarea'
import redis from '../utils/base/redis.js'
import list from '../utils/base/list.js'
import base from '../utils/base.js'

import redisStarRecord from './redis/starRecord.vue'
import redisHashList from './redis/tableHash.vue'
import shell from "@/utils/base/shell";
import {onMounted, onUnmounted, ref} from 'vue';
import arr from "@/utils/base/array";
import KeyDebounceDetector from "@/utils/base/keyup";
import { Close } from '@element-plus/icons-vue';

export default {
  name: 'cacheIndex',
  components: {
    Textarea,
    JsonViewer,
    Clipboard,
    redisStarRecord,
    redisHashList,
    Close,
  },
  data() {
    return {
      cacheType: {
        STRING: 'string',
        HASH: 'hash',
        LIST: 'list',
        SET: 'set',
        ZSET: 'zset',
      },
      //加载状态
      load: {
        redisList: true, //获取redis列表
        keysSearch: false, //大搜索
        callRefresh: false, //左侧搜索
      },
      //数据库
      redisChooseId: '',
      redisChooseConfig: {},
      redisList: [],
      //key
      cache: {
        cacheKey: '', //缓存key
        cacheType: '',
      },
      historyCheck: '',
      //keys
      keys: '',
      keysResult: [],
      keysResultCursor: 0,
      filterKeysResult: [],
      searchNum: 0,
      //搜索历史
      searchHistory: [],
      searchHistoryKey: 'redis_search_history',
      //select key
      selectRedisKey: '',
      //简版显示
      boolSimpleShow: false,
      loadingStatus: {},
      filterValue: '',
      scrollHeight: 0,
    }
  },
  inject: ["showTerminal", "resizeTerminal"],
  props: {
    shellShowResult: {
      type: String
    },
  },
  filters: {},
  activated: function () {
    this.resizeTerminal()
  },
  provide() {
    return {
      callStarListSearch: this.callStarListSearch, //收藏列表点击搜索
      callRefresh: this.callRefresh, //刷新key
      callStar: this.setCacheHistory, //收藏
      callMoreList: this.callMoreList, //加载更多 hash list zset
    }
  },
  unmounted: function () {
    let _that = this
  },
  mounted: function () {
    let _that = this
    _that.loadingStatus = _that.$helperLoad.getExecTypeStatus()
    _that.boolSimpleShow = _that.getStore('boolSimpleShow') === 'true'
    _that.initSearchHistory()
    _that.getRedisList()
    _that.windowChange()
    window.addEventListener('resize', function () {
      setTimeout(function () {
        _that.windowChange()
        //_that.$refs.redisStarRecord.GetStarList()
      }, 500)
    });
  },
  methods: {
    initSearchHistory: function () {
      let _that = this
      try {
        let historyData = _that.getStore(_that.searchHistoryKey)
        if (historyData) {
          _that.searchHistory = JSON.parse(historyData)
        } else {
          _that.searchHistory = []
        }
      } catch (e) {
        _that.searchHistory = []
      }
    },
    addSearchHistory: function (searchKey) {
      let _that = this
      if (!searchKey || searchKey.trim() === '') {
        return
      }
      searchKey = searchKey.trim()
      let newHistory = {
        key: searchKey,
        timestamp: Date.now()
      }
      let existingIndex = _that.searchHistory.findIndex(item => item.key === searchKey)
      if (existingIndex !== -1) {
        _that.searchHistory.splice(existingIndex, 1)
      }
      _that.searchHistory.unshift(newHistory)
      if (_that.searchHistory.length > 10) {
        _that.searchHistory = _that.searchHistory.slice(0, 10)
      }
      _that.saveSearchHistory()
    },
    removeSearchHistory: function (index) {
      let _that = this
      _that.searchHistory.splice(index, 1)
      _that.saveSearchHistory()
    },
    saveSearchHistory: function () {
      let _that = this
      try {
        _that.setStore(_that.searchHistoryKey, JSON.stringify(_that.searchHistory))
      } catch (e) {
        console.error('保存搜索历史失败:', e)
      }
    },
    handleHistorySearch: function (historyKey) {
      let _that = this
      _that.keys = historyKey
      _that.keysSearch()
    },
    keyUpKeys: function (event) {
      let _that = this
      if(event.key === 'ArrowDown'){
        for (let i in _that.filterKeysResult){
          if (_that.selectRedisKey === _that.filterKeysResult[i].CacheKey){
            if(i < _that.filterKeysResult.length - 1){
              console.log(_that.filterKeysResult[i] , parseInt(i)+1)
              _that.callRefresh(_that.filterKeysResult[parseInt(i)+1].CacheKey)
              break
            }
          }
        }
        event.preventDefault()
        event.stopPropagation()  // 新增：阻止事件冒泡
        event.stopImmediatePropagation()  // 可选：立即停止所有事件处理
        return false;
      }else if(event.key === 'ArrowUp'){
        for (let i in _that.filterKeysResult){
          if (_that.selectRedisKey === _that.filterKeysResult[i].CacheKey){
            if(i > 0){
              _that.callRefresh(_that.filterKeysResult[i-1].CacheKey)
              break
            }
          }
        }
        event.preventDefault()
        event.stopPropagation()  // 新增：阻止事件冒泡
        event.stopImmediatePropagation()  // 可选：立即停止所有事件处理
        return false;
      }
    },
    windowChange: function () {
      let _that = this
      let _height = base.GetDivHeight()
      _that.scrollHeight = parseInt(_height)
      if (_that.$refs && _that.$refs.redisHashList) {
        _that.$refs.redisHashList.WindowChange(_that.scrollHeight)
      }
    },
    //收藏列表 点击搜索
    callStarListSearch: function (value) {
      this.keys = value.key
      this.keysSearch()
    },
    //搜索左侧列表
    filterList: function () {
      let _that = this
      let searchRet = list.QuickSearch(this.filterValue, [...this.keysResult], ['CacheKey'])
      this.searchNum = searchRet.searchNum
      this.filterKeysResult = searchRet.list
      //搜索第一个的信息
      if (_that.filterKeysResult.length >= 1) {
        _that.callRefresh(_that.filterKeysResult[0].CacheKey)
      }else{
        //清空右侧
        _that.$refs.redisHashList.ShowList(_that.redisChooseId, '', {}, '', 0)
      }
    },
    //可用redis列表
    getRedisList: function () {
      let _that = this
      _that.load.redisList = true
      redis.RedisAvailableList(function (response) {
        if (response.ErrCode === 1) {
          return
        }
        _that.redisList = response.Data
        arr.SortByKey(_that.redisList , 'name' , 'asc')
        _that.getRedisDbSelect()
        _that.load.redisList = false
        _that.keysSearch()
      })
    },
    getRedisDbSelect: function () {
      let _that = this
      _that.redisChooseId = this.getStore('redisCheckId')
      for (let i in _that.redisList) {
        if (parseInt(_that.redisChooseId) === parseInt(_that.redisList[i].id)) {
          _that.redisChooseConfig = _that.redisList[i]
        }
      }
      if (_that.redisList.length === 0) {
        _that.redisChooseId = 0
        _that.redisChooseConfig = {}
        return
      }
      if (!_that.redisChooseConfig || !_that.redisChooseConfig.id) {
        _that.redisChooseConfig = _that.redisList[0]
        _that.redisChooseId = _that.redisChooseConfig.id
      }
    },
    redisDbChange: function (value) {
      let _that = this
      _that.cacheInit()
      _that.keysResult = []
      _that.setStore('redisCheckId', this.redisChooseId)
      //切换配置
      for (let key in this.redisList) {
        if (parseInt(this.redisList[key].id) === parseInt(this.redisChooseId)) {
          _that.redisChooseConfig = this.redisList[key]
          _that.keysSearch()
          break
        }
      }
    },
    initRedisList: function () {
      for (let i in this.keysResult) {
        this.keysResult[i].showName = this.keysResult[i].CacheKey
      }
      this.filterList()
    },
    //变更简版显示
    changeSimpleShow: function (boolSimpleShow) {
      this.boolSimpleShow = boolSimpleShow
      this.setStore('boolSimpleShow', this.boolSimpleShow)
      this.sortRedisList()
    },
    sortRedisList: function () {
      //优化显示
      for (let i in this.keysResult) {
        if (this.boolSimpleShow) {
          if (this.keys !== '') {
            let indexKey = this.keysResult[i].showName.indexOf(this.keys)
            if (indexKey !== false) {
              //只支持从头开始的匹配
              let length = this.keysResult[i].showName.length
              let sub_length = indexKey + this.keys.length
              this.keysResult[i].showName =
                  '[...]' +
                  this.keysResult[i].showName.substr(
                      sub_length,
                      length - sub_length
                  )
            }
          }
        } else {
          if (this.keysResult[i].showName.substr(0, 5) === '[...]') {
            this.keysResult[i].showName = this.keysResult[i].CacheKey
          }
        }
      }
    },
    //查询单个信息
    callRefresh: function (key) {
      this.selectRedisKey = key
      let _that = this
      //临时变量
      let cache = {
        cacheKey: this.cache.cacheKey, //缓存key
        cacheType: this.cache.cacheType,
      }
      let hashResult = []
      cache.UniKey = this.redisChooseId
      cache.cacheKey = key
      cache.ExecType = 'redis_search'
      //拿到key类型
      _that.load.callRefresh = true
      redis.RedisSearch(_that.redisChooseConfig, key, 0, '', function (responseSearch) {
        setTimeout(function () {
          _that.load.callRefresh = false;
        }, 200)
        if (responseSearch.ErrCode === 1) {
          _that.$helperNotify.error('key已不存在')
          _that.keysSearch()
          return
        }
        let data = responseSearch.Data.Result
        cache.cacheType = responseSearch.Data.keyType
        if (cache.cacheType === _that.cacheType.SET) {
          for (let index in data) {
            hashResult.push({key: index, value: data[index]})
          }
        } else if (cache.cacheType === _that.cacheType.LIST) {
          for (let index in data) {
            hashResult.push({index: index, value: data[index]})
          }
        } else if (cache.cacheType === _that.cacheType.HASH) {
          for (let index in data) {
            hashResult.push({field: index, value: data[index]})
          }
        } else if (cache.cacheType === _that.cacheType.ZSET) {
          for (let index in data) {
            hashResult.push({
              member: data[index].Member,
              score: data[index].Score,
            })
          }
        }
        if (cache.cacheType === 'string') {
          _that.$refs.redisHashList.ShowList(_that.redisChooseId, cache.cacheType, {
            value: _that.transResponseData(data),
          }, cache.cacheKey, responseSearch.Data.KeyTtl)
        } else {
          _that.$refs.redisHashList.ShowList(_that.redisChooseId, cache.cacheType, hashResult, cache.cacheKey, responseSearch.Data.KeyTtl, responseSearch.Data.Length, responseSearch.Data.Cursor, responseSearch.Data.IsMore)
        }
        //临时变量赋值 防止变动太频繁
        _that.cache = cache
      })
    },
    //子项中翻页 例如hash list
    callMoreList: function (hashResult, cursor, search) {
      let _that = this
      let cache = {
        cacheKey: _that.cache.cacheKey, //缓存key
        cacheType: _that.cache.cacheType,
      }
      cache.UniKey = _that.cache.UniKey
      cache.cacheKey = _that.cache.cacheKey
      cache.ExecType = 'redis_search'
      //拿到key类型
      _that.load.callRefresh = true
      redis.RedisSearch(_that.redisChooseConfig, cache.cacheKey, cursor, search, function (responseSearch) {
        setTimeout(function () {
          _that.load.callRefresh = false;
        }, 100)
        if (responseSearch.ErrCode === 1) {
          _that.$helperNotify.error('key已不存在')
          _that.keysSearch()
          return
        }
        let data = responseSearch.Data.Result
        cache.cacheType = responseSearch.Data.keyType
        if (cache.cacheType === _that.cacheType.SET) {
          for (let index in data) {
            hashResult.push({key: index, value: data[index]})
          }
        } else if (cache.cacheType === _that.cacheType.LIST) {
          for (let index in data) {
            hashResult.push({index: index, value: data[index]})
          }
        } else if (cache.cacheType === _that.cacheType.HASH) {
          for (let index in data) {
            hashResult.push({field: index, value: data[index]})
          }
        } else if (cache.cacheType === _that.cacheType.ZSET) {
          for (let index in data) {
            hashResult.push({
              member: data[index].Member,
              score: data[index].Score,
            })
          }
        }
        _that.$refs.redisHashList.ShowList(_that.redisChooseId, cache.cacheType, hashResult, cache.cacheKey, responseSearch.Data.KeyTtl, responseSearch.Data.Length, responseSearch.Data.Cursor, responseSearch.Data.IsMore)
        //临时变量赋值 防止变动太频繁
        _that.cache = cache
      })
    },
    setCacheHistory: function (paramsT) {
      this.$refs.redisStarRecord.star(paramsT) //展示收藏弹窗
    },
    //搜索缓存 这里是模糊查询 会返回多个
    keysSearch: function (getMore) {
      let _that = this
      if (parseInt(this.redisChooseId) === 0) {
        return false
      } else {
        this.historyCheck = this.keys
      }
      if (getMore !== true) {
        _that.keysResultCursor = 0
        _that.addSearchHistory(this.keys)
      }
      _that.load.keysSearch = true
      redis.RedisKeys(this.redisChooseConfig, _that.keysResultCursor, '*' + this.keys + '*', function (response) {
            if (response.ErrCode === 1) {
              _that.load.keysSearch = false
              return
            }
            if (getMore === true) {
              for (let i in response.Data.list) {
                _that.keysResult.push(response.Data.list[i])
              }
            } else {
              _that.keysResult = response.Data.list
            }
            _that.keysResultCursor = response.Data.cursor
            _that.initRedisList()
            _that.sortRedisList()

            //清空
            if (_that.keysResult.length === 0) {
              _that.cacheInit()
            }
            //查找类型
            _that.filterList()

            setTimeout(function () {
              _that.load.keysSearch = false;
            }, 200)
          }
      )
    },
    transResponseData: function (data) {
      let returnDataType = Object.prototype.toString.call(data)
      if (
          returnDataType === '[object Array]' ||
          returnDataType === '[object Object]'
      ) {
        return JSON.stringify(data)
      } else {
        return data
      }
    },
    //清空右侧的缓存显示内容
    cacheInit: function () {
      this.$refs.redisHashList.ShowList(this.redisChooseId, '', [], '', 0)
    },
    delAll: function () {
      let _that = this
      let params = {UniKey: this.redisChooseId, Keys: _that.filterKeysResult}
      params.ExecType = 'redis_delete_batch'
      this.$confirm('确定删除' + _that.filterKeysResult.length + '个缓存吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      })
          .then(() => {
            _that.setLoading(params)
            let waitDeleteKeyList = []
            for (let index in _that.filterKeysResult) {
              waitDeleteKeyList.push(_that.filterKeysResult[index].CacheKey)
            }
            redis.RedisDelAllKey(_that.redisChooseConfig, waitDeleteKeyList, function (response) {
                  _that.keysSearch()
                  _that.cacheInit()
                  _that.cancelLoading(params)
                }
            )
          })
          .catch(() => {
          })
    },
    success: function (msg) {
      // Message.success(msg);
      this.$notify({
        title: '提示',
        message: msg,
        type: 'success',
        duration: 1000,
      })
    },
    error: function (msg) {
      // Message.error(msg);
      this.$notify({
        title: '提示',
        message: msg,
        type: 'error',
        duration: 1000,
      })
    },
    setStore: function (key, value) {
      localStorage.setItem(key, value)
    },
    getStore: function (key) {
      return localStorage.getItem(key)
    },
    setLoading: function (params) {
      this.loadingStatus[params.ExecType] = true
      let that = this
      setTimeout(function () {
        that.loadingStatus[params.ExecType] = false
      }, 25000)
    },
    cancelLoading: function (params) {
      let that = this
      setTimeout(function () {
        that.loadingStatus[params.ExecType] = false
      }, 1000)
    },
  },
}
</script>
