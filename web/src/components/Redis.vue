<template>
  <div class="redis-page-container">
    <!-- 顶部搜索区域 -->
    <div class="redis-header-card" v-loading="load.redisList">
      <div class="header-title">
        <svg class="header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <span>Redis 管理器</span>
      </div>
      <div class="search-row">
        <el-input 
          v-model="keys" 
          placeholder="请输入key进行搜索..." 
          class="search-input"
          @keyup.enter="keysSearch"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="redisChooseId" placeholder="选择Redis实例" class="redis-select" @change="redisDbChange">
          <el-option v-for="(value,key) in redisList" :key="value.name" :label="value.name" :value="value.id">
          </el-option>
        </el-select>
        <el-button v-loading="load.keysSearch" type="primary" class="search-btn" @click="keysSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button v-if="keys !== ''" class="action-btn star-btn" @click="setCacheHistory({ cacheKey : keys})">
          <el-icon><Star /></el-icon>
          收藏
        </el-button>
        <el-button class="action-btn list-btn" @click="$refs.redisStarRecord.showStarList();">
          <el-icon><Collection /></el-icon>
          收藏列表
        </el-button>
      </div>
      <!-- 搜索历史 -->
      <div v-if="searchHistory.length > 0" class="search-history-container">
        <span class="history-label">搜索历史:</span>
        <div class="search-history-list">
          <div v-for="(item, index) in searchHistory" :key="index" class="search-history-item">
            <span class="search-history-text" @click="handleHistorySearch(item.key)">{{ item.key }}</span>
            <span class="search-history-delete" @click="removeSearchHistory(index)">
              <el-icon><Close /></el-icon>
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="main-content-wrapper">
      <div class="main-content">
        <!-- 左侧Key列表 -->
        <div class="key-list-wrapper">
          <div class="key-list-card">
          <div class="key-list-header">
            <div class="header-left">
              <span class="key-count" v-if="keysResult && keysResult.length > 0">
                共 <strong>{{ searchNum }}</strong> 个Key
              </span>
            </div>
            <div class="header-right">
              <el-button v-if="keysResult && keysResult.length > 0" size="small" type="danger" plain @click="delAll">
                <el-icon><Delete /></el-icon>
                删除所有
              </el-button>
              <el-button v-if="keysResult && keysResult.length > 0" size="small" type="primary" plain @click="boolSimpleShow = !boolSimpleShow;changeSimpleShow(boolSimpleShow);">
                <el-icon v-if="boolSimpleShow"><View /></el-icon>
                <el-icon v-else><Hide /></el-icon>
                {{ boolSimpleShow ? '取消优化' : '优化显示' }}
              </el-button>
            </div>
          </div>
          <div class="key-list-content">
            <el-input v-if="keysResult.length > 0" v-model="filterValue" placeholder="过滤key,空格多个条件" size="small" class="filter-input" type="text" @input="filterList" clearable>
              <template #prefix>
                <el-icon><Filter /></el-icon>
              </template>
            </el-input>
            <el-scrollbar ref="scrollbarRef" @keydown="keyUpKeys" tabindex="0" class="key-scrollbar" :style="{ height: '100%' }">
              <div v-if="keysResultCursor !== 0" class="load-more-btn" @click="keysSearch(true)">
                <el-icon><Download /></el-icon>
                加载更多
              </div>
              <template v-for="(value, key) in filterKeysResult" :key="key" >
                <div 
                  :class="['key-item', selectRedisKey === value.CacheKey ? 'key-item-active' : '']" 
                  @click="callRefresh(value.CacheKey)"
                >
                  <el-icon class="key-icon"><Key /></el-icon>
                  <span class="key-text">{{ value.showName }}</span>
                </div>
              </template>
              <div v-if="!keysResult || keysResult.length === 0" class="empty-state">
                <el-icon class="empty-icon"><FolderOpened /></el-icon>
                <span>暂无数据，请搜索</span>
              </div>
            </el-scrollbar>
          </div>
          </div>
        </div>
        <!-- 右侧详情区域 -->
        <div class="detail-wrapper">
          <div class="detail-card">
            <el-form ref="form" v-loading="load.callRefresh">
              <redisHashList ref="redisHashList" :callMoreList="callMoreList" :callRefresh="callRefresh" :star="setCacheHistory"></redisHashList>
            </el-form>
          </div>
        </div>
      </div>
    </div>
    <!--  收藏列表-->
    <redisStarRecord ref="redisStarRecord" :callStarListSearch="callStarListSearch"></redisStarRecord>
  </div>
</template>
<style>
/* 页面容器 */
.redis-page-container {
  padding: 0;
  width: 100%;
}

/* 顶部卡片样式 */
.redis-header-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  padding: 20px 24px;
  margin-bottom: 20px;
  box-shadow: 0 4px 20px rgba(102, 126, 234, 0.25);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #fff;
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 16px;
}

.header-icon {
  width: 28px;
  height: 28px;
  color: #fff;
}

.search-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.search-input {
  flex: 1;
  max-width: 500px;
}

.search-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.redis-select {
  width: 200px;
}

.redis-select :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.search-btn {
  border-radius: 10px;
  padding: 0 20px;
  height: 40px;
  background: #fff;
  color: #667eea;
  border: none;
  font-weight: 500;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.search-btn:hover {
  background: #f5f7fa;
}

.action-btn {
  border-radius: 10px;
  padding: 0 16px;
  height: 40px;
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.3);
  font-weight: 500;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  color: #fff;
  border-color: rgba(255, 255, 255, 0.5);
}

.star-btn {
  background: rgba(255, 193, 7, 0.3);
  border-color: rgba(255, 193, 7, 0.5);
}

.list-btn {
  background: rgba(103, 58, 183, 0.3);
  border-color: rgba(103, 58, 183, 0.5);
}

/* 搜索历史 */
.search-history-container {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.history-label {
  color: rgba(255, 255, 255, 0.8);
  font-size: 13px;
}

.search-history-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.search-history-item {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  font-size: 13px;
  color: #fff;
  transition: all 0.3s ease;
}

.search-history-item:hover {
  background: rgba(255, 255, 255, 0.25);
  border-color: rgba(255, 255, 255, 0.4);
  transform: translateY(-1px);
}

.search-history-text {
  cursor: pointer;
  margin-right: 8px;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-history-delete {
  cursor: pointer;
  color: rgba(255, 255, 255, 0.7);
  font-size: 12px;
  padding: 2px;
  border-radius: 50%;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.search-history-delete:hover {
  color: #fff;
  background: rgba(255, 87, 87, 0.5);
}

/* 主内容区域 */
.main-content-wrapper {
  width: 100%;
  overflow: hidden;
}

.main-content {
  display: flex;
  gap: 20px;
  height: calc(100vh - 220px);
  min-height: 400px;
}

/* 左侧Key列表 */
.key-list-wrapper {
  flex: 0 0 33.333%;
  max-width: 33.333%;
}

/* 右侧详情 */
.detail-wrapper {
  flex: 0 0 66.666%;
  max-width: 66.666%;
}

/* 左侧Key列表卡片 */
.key-list-card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.key-list-header {
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fafbfc;
}

.header-left {
  display: flex;
  align-items: center;
}

.key-count {
  font-size: 14px;
  color: #606266;
}

.key-count strong {
  color: #667eea;
  font-size: 16px;
}

.header-right {
  display: flex;
  gap: 8px;
}

.key-list-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 0;
  min-height: 0;
}

.filter-input {
  margin: 12px;
  width: calc(100% - 24px);
  flex-shrink: 0;
}

.filter-input :deep(.el-input__wrapper) {
  border-radius: 8px;
}

.key-scrollbar {
  flex: 1;
  padding: 0 12px 12px;
  overflow: hidden;
}

.key-scrollbar :deep(.el-scrollbar__view) {
  min-height: 100%;
}

.load-more-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px;
  margin-bottom: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  border-radius: 10px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.load-more-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.key-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  margin-bottom: 6px;
  border-radius: 10px;
  cursor: pointer;
  font-size: 13px;
  font-family: 'Consolas', 'Monaco', monospace;
  background: #f8f9fc;
  border: 1px solid transparent;
  transition: all 0.2s ease;
}

.key-item:hover {
  background: #eef1fc;
  border-color: #667eea;
}

.key-item-active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  border-color: transparent;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.key-item-active .key-icon {
  color: #fff;
}

.key-icon {
  color: #667eea;
  font-size: 14px;
  flex-shrink: 0;
}

.key-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #909399;
  min-height: 200px;
}

.empty-icon {
  font-size: 56px;
  margin-bottom: 16px;
  color: #c0c4cc;
}

/* 右侧详情卡片 */
.detail-card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  overflow: hidden;
  padding: 20px;
  height: 100%;
}

.box-card .el-tag-he {
  margin-left: 5px;
  font-size: 13px;
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

/* 响应式 */
@media (max-width: 1200px) {
  .search-input {
    max-width: 350px;
  }
  
  .redis-select {
    width: 160px;
  }
  
  .key-list-wrapper {
    flex: 0 0 40%;
    max-width: 40%;
  }
  
  .detail-wrapper {
    flex: 0 0 60%;
    max-width: 60%;
  }
}

@media (max-width: 768px) {
  .search-row {
    flex-wrap: wrap;
  }
  
  .search-input {
    max-width: 100%;
  }
  
  .redis-select {
    width: 100%;
  }
  
  .main-content {
    flex-direction: column;
    height: auto;
  }
  
  .key-list-wrapper,
  .detail-wrapper {
    flex: none;
    max-width: 100%;
    width: 100%;
  }
  
  .key-list-card,
  .detail-card {
    height: auto;
    min-height: 300px;
  }
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
import { Close, Search, Star, Collection, Delete, View, Hide, Filter, Download, Key, FolderOpened } from '@element-plus/icons-vue';

export default {
  name: 'cacheIndex',
  components: {
    Textarea,
    JsonViewer,
    Clipboard,
    redisStarRecord,
    redisHashList,
    Close,
    Search,
    Star,
    Collection,
    Delete,
    View,
    Hide,
    Filter,
    Download,
    Key,
    FolderOpened,
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
