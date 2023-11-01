<template>
  <el-card>
    <el-card>
      <el-input style="width:600px;" v-model="keys" @keyup.enter.native="keysSearch"
                placeholder="请输入key"></el-input>
      <el-select v-model="redisCheck" @change="redisDbChange">
        <el-option
          v-for="(value,key) in redisList"
          :key="value.name"
          :label="value.name"
          :value="value.name">
        </el-option>
      </el-select>
      <el-button type="primary" :loading="loadingStatus['redis_search']" @click="keysSearch">查询</el-button>
      <el-button icon="el-icon-refresh-left" @click="refresh" circle></el-button>
      <el-button type="primary" icon="el-icon-plus" @click="showAddCache" circle></el-button>

      <!--        历史查找-->
      <el-popover
        placement="bottom"
        width="1000"
        v-model="historySearchVisible"
        trigger="hover"
      >
        <el-row :gutter="10" style="margin-top: 10px;">
          <el-tag effect="dark" type="" closable @close="deleteHistory(value)"
                  style="margin-left: 5px;margin-top:5px;" v-for="(value,key) in historyList" :key="key">
            <!--              <el-radio style="word-wrap:break-word;" v-model="historyCheck" @change="searchHistory(value)" :label="value.Search">{{ value.Search }}</el-radio>-->
            <span v-if="historyCheck === value.Search "
                  style="font-size:13px;color:blue;word-wrap:break-word;cursor:default;"
                  @click="searchHistory(value)">{{ value.Search }}</span>
            <span v-else style="font-size:13px;word-wrap:break-word;cursor:default;"
                  @click="historySearchVisible=false;searchHistory(value);">{{ value.Search }}</span>
          </el-tag>
        </el-row>
        <el-button slot="reference" @click="historySearchVisible = !historySearchVisible">历史记录</el-button>
      </el-popover>
    </el-card>
    <br/>
    <el-row :gutter="20">
      <el-col :span="8">
        <el-card class="box-card" style="">
          <div class="grid-content bg-purple" style="height:480px;overflow:auto;">
            <el-button :loading="loadingStatus['redis_delete_batch']" type="danger" style="margin-bottom: 7px;"
                       @click="delAll" v-if="keysResult.length > 0 "
                       size="mini">删除以下缓存
            </el-button>
            <el-checkbox v-if="keysResult.length > 0" style="margin-top:3px;margin-left:3px;" @change="changeSimpleShow"
                         v-model="boolSimpleShow">优化显示/排序
            </el-checkbox>
            <br/>
            <el-input type="text" v-if="keysResult.length > 0" v-model="filterValue" size="small" style="width:91%;"
                      placeholder="输入搜索过滤" @input="filterList"></el-input>
            <span style="font-size: 11px;" v-if="keysResult.length > 0">({{ searchNum }})</span>
            <div v-for="(value,key) in keysResult">
              <div>
                <el-tag size="medium" effect="light" style="margin-top:3px;" v-if="value.boolShow">
                  <el-link style="padding:3px;font-size: 13px;" v-if="selectRedisKey === value.CacheKey"
                           @click="search(value.CacheKey)"> {{ value.showName }}
                  </el-link>
                  <el-link style="padding:3px;font-size: 11px;color:dodgerblue;" v-else
                           @click="search(value.CacheKey)"> {{ value.showName }}
                  </el-link>
                </el-tag>
              </div>

            </div>
          </div>
        </el-card>

      </el-col>
      <el-col :span="16">
        <div class="grid-content bg-purple">
          <el-card class="box-card"
                   style="height:520px;overflow:auto;">
            <div slot="header" class="clearfix">
              <template v-if="cache.cacheKey !== ''">
                <el-tag size="medium">{{ cache.cacheType }}</el-tag>
                <el-tag size="medium" style="cursor: copy;" class="copyCacheKey"
                        :data-clipboard-text="cache.cacheKey"
                        @click="copyKey">{{ cache.cacheKey }}
                </el-tag>
              </template>
              <template v-if="cache.cacheKey !==  '' && cache.cacheType === 'string'">
                <el-tag size="medium">
                  <el-checkbox class="string-option" v-model="cache.strHasSerialize" @change="unserialize()">
                    serialize
                  </el-checkbox>
                </el-tag>
                <el-tag size="medium">
                  <el-checkbox class="string-option" v-model="cache.strHasJson" @change="json()">json
                  </el-checkbox>
                </el-tag>
              </template>
              <template v-if="cache.cacheKey !== ''">
                <el-tag size="medium" v-if="cache.startEditTTL === true">
                  ttl：
                  <input style="width:100px;border:0;" v-model="cache.ttl" type="text"/>
                  <el-button size="mini" type="primary" @click="saveTTL" style="padding:3px">保存</el-button>
                  <el-button size="mini" type="default" @click="cancelEditTTL" style="padding:3px;">取消</el-button>
                </el-tag>
                <el-tag size="medium" @click="editTTL" style="cursor:pointer" v-if="cache.startEditTTL === false">
                  ttl：
                  {{ cache.ttl }}
                </el-tag>
                <el-button icon="el-icon-refresh-left " size="medium" @click="search(cache.cacheKey)"
                           circle></el-button>
              </template>
              <template style="float:right;" v-if="cache.cacheKey !== ''">
                <el-button type="danger" size="mini" icon="el-icon-delete"
                           style="margin-left: 10px;float:right;" @click="delCache()">删除
                </el-button>
                <el-button type="primary" size="mini" icon="el-icon-plus" style="margin-left: 10px;float:right;"
                           @click="createSubCache" v-if="cache.cacheType !== cacheType.STRING">
                </el-button>

                <el-button type="primary" size="mini" icon="el-icon-check" style="margin-left: 10px;float:right;"
                           @click="saveString()"
                           v-if="cache.strShowType === 1 && cache.cacheType === cacheType.STRING">保存
                </el-button>
              </template>
            </div>

            <el-form ref="form">
              <el-form-item v-if="cache.cacheType === cacheType.STRING" style="margin-top:10px;">
                <el-input type="textarea" v-if="cache.strShowType === 1 " v-model="searchResult"
                          rows="20"></el-input>
                <el-input type="textarea" class="readonlyTextarea" v-if="cache.strShowType === 2 " readonly
                          v-model="searchResult" style="background: #eeee;" rows="20"></el-input>

                <json-viewer :value="searchResult" v-if="cache.strShowType === 3 " :expand-depth="10" copyable
                             sort></json-viewer>
              </el-form-item>

              <template
                v-if="cache.cacheType === cacheType.HASH || cache.cacheType === cacheType.LIST || cache.cacheType === cacheType.SET || cache.cacheType === cacheType.ZSET "
                style="margin-top:10px;">
                <el-table
                  :data="hashResult"
                  style="width: 100%;font-size:13px;">

                  <el-table-column v-if=" cache.cacheType === cacheType.LIST"
                                   prop="index"
                                   :key="Math.random()"
                                   label="index"
                                   width="180" sortable>
                  </el-table-column>

                  <el-table-column v-if="cache.cacheType === cacheType.HASH"
                                   prop="field"
                                   :key="Math.random()"
                                   label="field"
                                   width="180" sortable>
                  </el-table-column>
                  <el-table-column v-if="cache.cacheType === cacheType.ZSET"
                                   prop="member"
                                   :key="Math.random()"
                                   label="member"
                                   width="500" sortable>
                  </el-table-column>
                  <el-table-column v-if="cache.cacheType === cacheType.ZSET"
                                   prop="score"
                                   :key="Math.random()"
                                   label="score"
                                   width="200" sortable>
                  </el-table-column>


                  <el-table-column v-if="cache.cacheType === cacheType.HASH || cache.cacheType === cacheType.LIST"
                                   prop="value"
                                   label="value"
                                   width="550" sortable>
                    <template slot-scope="scope">
                      {{ scope.row.value | strLenFilter }}
                    </template>
                  </el-table-column>
                  <el-table-column v-if="cache.cacheType === cacheType.SET"
                                   prop="value"
                                   label="value"
                                   width="600" sortable>
                    <template slot-scope="scope">
                      {{ scope.row.value | strLenFilter }}
                    </template>
                  </el-table-column>
                  <el-table-column
                    label="操作">
                    <template slot-scope="scope">
                      <el-button type="primary" v-if="cache.cacheType !== cacheType.SET" icon="el-icon-edit"
                                 size="mini" @click="editSub(scope.row)"></el-button>
                      &nbsp;
                      <el-button type="danger" v-if="cache.cacheType === cacheType.HASH" icon="el-icon-delete"
                                 size="mini" @click="delSub(scope.row.field)"></el-button>
                      <el-button type="danger"
                                 v-if="cache.cacheType === cacheType.SET || cache.cacheType === cacheType.LIST "
                                 size="mini" icon="el-icon-delete" @click="delSub(scope.row.value)"></el-button>
                      <el-button type="danger" v-if="cache.cacheType === cacheType.ZSET" size="mini"
                                 icon="el-icon-delete" @click="delSub(scope.row.member)"></el-button>

                    </template>

                  </el-table-column>
                </el-table>
              </template>
            </el-form>

          </el-card>
        </div>
      </el-col>
    </el-row>

    <!--新增弹窗-->
    <el-dialog title="新增缓存" :visible.sync="addCacheClass" :append-to-body="true" width="70%;">
      <el-form>
        <el-form-item label="类型" :label-width="addCacheInputWidth">
          <el-select v-model="addSubCache.cacheType" placeholder="选择缓存类型">
            <el-option label="字符串" value="string"></el-option>
            <el-option label="哈希" value="hash"></el-option>
            <el-option label="列表" value="list"></el-option>
            <el-option label="集合" value="set"></el-option>
            <el-option label="有序集合" value="zset"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="key" :label-width="addCacheInputWidth">
          <el-input v-model="addSubCache.cacheKey" autocomplete="off"></el-input>
        </el-form-item>

        <el-form-item v-if="addSubCache.cacheType === cacheType.HASH" label="field"
                      :label-width="addCacheInputWidth">
          <el-input v-model="addSubCache.cacheField" autocomplete="off"></el-input>
        </el-form-item>
        <el-form-item label="value"
                      v-if="addSubCache.cacheType === cacheType.HASH || addSubCache.cacheType === cacheType.STRING || (addSubCache.cacheType === cacheType.LIST && addSubCache.boolCreate === 1) "
                      :label-width="addCacheInputWidth">
          <el-input v-model="addSubCache.cacheValue" autocomplete="off"></el-input>
        </el-form-item>

        <el-form-item label="lPush"
                      v-if="addSubCache.cacheType === cacheType.LIST && addSubCache.boolCreate === 2"
                      :label-width="addCacheInputWidth">
          <el-input v-model="addSubCache.lPushValue" autocomplete="off"></el-input>
        </el-form-item>

        <el-form-item label="rPush"
                      v-if="addSubCache.cacheType === cacheType.LIST && addSubCache.boolCreate === 2"
                      :label-width="addCacheInputWidth">
          <el-input v-model="addSubCache.rPushValue" autocomplete="off"></el-input>
        </el-form-item>


        <el-form-item v-if="addSubCache.cacheType === cacheType.SET || addSubCache.cacheType === cacheType.ZSET  "
                      label="member" :label-width="addCacheInputWidth">
          <el-input v-model="addSubCache.cacheMember" autocomplete="off"></el-input>
        </el-form-item>
        <el-form-item v-if="addSubCache.cacheType === cacheType.ZSET  " label="score"
                      :label-width="addCacheInputWidth">
          <el-input v-model="addSubCache.cacheScore" autocomplete="off"></el-input>
        </el-form-item>
        <el-form-item label="ttl/秒" :label-width="addCacheInputWidth" v-if="addSubCache.boolCreate === 1">
          <el-input v-model="addSubCache.ttl" autocomplete="off"></el-input>
        </el-form-item>


      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="addCacheClass = false">取 消</el-button>
        <el-button type="primary" @click="createCache">确 定</el-button>
      </div>
    </el-dialog>


    <!--编辑弹窗-->
    <el-dialog title="编辑缓存" :visible.sync="editCacheClass" :append-to-body="true" width="90%;">
      <template v-if="editSubCache.cacheType === cacheType.LIST || editSubCache.cacheType === cacheType.HASH ">
        <el-tag size="medium">
          <el-checkbox class="string-option" v-model="editSubCache.strHasSerialize" @change="editSubUnserialize()">
            serialize
          </el-checkbox>
        </el-tag>
        <el-tag size="medium">
          <el-checkbox class="string-option" v-model="editSubCache.strHasJson" @change="editSubJson()">json
          </el-checkbox>
        </el-tag>
      </template>
      <el-form style="margin-top: 5px;">
        <el-form-item style="margin-left:0;" v-if="editSubCache.cacheType === cacheType.HASH">
          <el-input v-model="editSubCache.field" autocomplete="off"></el-input>
        </el-form-item>

        <!--        编辑二级缓存-->
        <el-form-item style="margin-top:10px;"
                      v-if="editSubCache.cacheType === cacheType.LIST || editSubCache.cacheType === cacheType.HASH ">
          <el-input type="textarea" v-if="editSubCache.strShowType === 1" v-model="editSubCache.value"
                    rows="20"></el-input>

          <el-input type="textarea" class="readonlyTextarea" v-if="editSubCache.strShowType === 2 " readonly
                    v-model="editSubCache.searchResult" style="background: #eeee;" rows="20"></el-input>

          <json-viewer :value="editSubCache.searchResult" v-if="editSubCache.strShowType === 3 " :expand-depth="10"
                       copyable
                       sort></json-viewer>
        </el-form-item>

        <el-form-item v-if="editSubCache.cacheType === cacheType.ZSET  "
                      label="member" :label-width="addCacheInputWidth">
          <el-input v-model="editSubCache.member" autocomplete="off" disabled></el-input>
        </el-form-item>
        <el-form-item v-if="editSubCache.cacheType === cacheType.ZSET  " label="score"
                      :label-width="addCacheInputWidth">
          <el-input v-model="editSubCache.score" autocomplete="off"></el-input>
        </el-form-item>

      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="editCacheClass = false">取 消</el-button>
        <el-button type="primary" @click="funcEditSubCache">确 定</el-button>
      </div>
    </el-dialog>
  </el-card>


</template>


<script>
import Vue from "vue";
import JsonViewer from 'vue-json-viewer'
import Clipboard from 'clipboard'
import Textarea from "./textarea";
import {Message} from "element-ui";
import mod from "../utils/api/module.js"
import redis from "../utils/api/redis.js"
import php from "../utils/api/php.js"
import base from "../utils/api/base.js"

export default {
  name: "cacheIndex",
  components: {
    Textarea,
    JsonViewer,
    Clipboard,
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
      //是否显示历史搜索记录
      historySearchVisible: false,
      loading: false,
      //数据库
      redisCheck: 'common3',
      redisList: [],
      //key
      cache: {
        cacheKey: '',             //缓存key
        cacheType: '',
        strShowType: 1,          //string 才有： 1 textarea （原值） , 2 反序列化 , 3 json展示
        strHasSerialize: false,  //是否序列化
        strHasJson: false,   //是否json展示
        ttl: 0,                  //过期时间 0 永久
        startEditTTL: false,     //是否开始编辑ttl
      },
      historyCheck: '',
      //keys
      keys: '',
      keysResult: [],
      //string result
      searchResult: "",
      searchSourceResult: "",
      searchNum: 0,
      //hash result
      hashResult: [],
      //history
      historyList: [],
      //select key
      selectRedisKey: '',
      //新增二级缓存
      addCacheClass: false,
      addCacheInputWidth: "150px",
      addSubCache: {
        boolCreate: 1,     //1：外部新增一个list   2：list中增加一个值   3 ：编辑list中的一个值
        cacheType: '',     //string hash
        cacheKey: '',
        cacheField: '',
        cacheValue: '',
        ttl: 0,           //默认永久
        cacheMember: '', //集合的值
        cacheScore: '',   //有序集合分值
        lPushValue: '',
        rPushValue: '',
      },
      //编辑二级缓存
      editCacheClass: false,
      editSubCache: {
        cacheKey: '',
        cacheType: '',
        key: 0,  //list
        value: '',
        field: '', //哈希
        strShowType: 1,          //编辑用的 string和list 才有： 1 textarea （原值） , 2 反序列化 , 3 json展示
        strHasSerialize: false,  //是否序列化
        strHasJson: false,       //是否json展示
        searchResult: '',       //json  和 序列化后的值
        member: '',
        score: 0,
      },
      //简版显示
      boolSimpleShow: false,
      loadingStatus: {},
      filterValue: '',
    }
  },
  filters: {
    strLenFilter: function (value) {
      let length = 70
      if (value.length > length) {
        return value.substr(0, length) + '...'
      } else {
        return value
      }
    }
  },
  mounted: function () {
    this.getRedisList();
    this.addSubCache.cacheType = this.cacheType.STRING;
    this.loadingStatus = this.$helperLoad.getExecTypeStatus()
    this.boolSimpleShow = this.getStore('boolSimpleShow') === 'true'
  },
  methods: {
    filterList: function (event) {
      this.searchNum = 0
      //优化显示
      for (let i in this.keysResult) {
        if (this.filterValue !== '') {
          let indexKey = this.keysResult[i].showName.indexOf(this.filterValue)
          this.keysResult[i].boolShow = indexKey >= 0;
        } else {
          this.keysResult[i].boolShow = true
        }
        if (this.keysResult[i].boolShow) {
          this.searchNum++
        }
      }
    },
    getRedisList: function () {
      let _that = this
      redis.RedisAvailableList(function (response) {
        let redisNameList = response.Data;
        _that.redisList = redis.GetRedisConfigListByNameList(redisNameList);
        console.log(_that.redisList)
        _that.redisCheck = _that.redisList[0].Name;
        _that.getRedisDbSelect();
        _that.getCacheHistory();
      })
    },
    getRedisDbSelect: function () {
      this.redisCheck = this.getStore('redisCheck')
      console.log(this.redisCheck)
    },
    redisDbChange: function (value) {
      this.historyList = [];
      this.getCacheHistory();
      this.cacheInit();
      this.keysResult = [];
      this.setStore('redisCheck', this.redisCheck)
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
      this.sortRedisList();
    },
    sortRedisList: function () {
      //优化显示
      for (let i in this.keysResult) {
        if (this.boolSimpleShow) {
          if (this.keys !== '') {
            let indexKey = this.keysResult[i].showName.indexOf(this.keys)
            if (indexKey !== false) { //只支持从头开始的匹配
              let length = this.keysResult[i].showName.length
              let sub_length = indexKey + this.keys.length
              this.keysResult[i].showName = "[...]" + this.keysResult[i].showName.substr(sub_length, length - sub_length)
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
    search: function (key) {
      this.selectRedisKey = key;
      let _that = this
      this.cache.strShowType = 1;
      this.cache.strHasSerialize = false;
      this.cache.strHasJson = false;
      this.cache.UniKey = this.redisCheck
      this.cache.cacheKey = key
      this.cache.ExecType = 'redis_search'
      //拿到key类型
      redis.RedisKeyType(this.redisCheck, key, function (response) {
        if (response.Data !== '') {
          _that.cache.cacheType = response.Data.Type;
          _that.cache.ttl = response.Data.TTL;
          //拿到结果
          redis.RedisSearch(_that.redisCheck, key, function (responseSearch) {
            let data = responseSearch.Data;
            if (_that.cache.cacheType === _that.cacheType.SET) {
              _that.hashResult = [];
              for (let index in data) {
                _that.hashResult.push({"key": index, "value": data[index]});
              }
              _that.cache.strShowType = 0;
            } else if (_that.cache.cacheType === _that.cacheType.LIST) {
              _that.hashResult = [];
              for (let index in data) {
                _that.hashResult.push({"index": index, "value": data[index]});
              }
              _that.cache.strShowType = 0;
            } else if (_that.cache.cacheType === _that.cacheType.HASH) {
              _that.hashResult = [];
              for (let index in data) {
                _that.hashResult.push({"field": index, "value": data[index]});
              }
              _that.cache.strShowType = 0;
            } else if (_that.cache.cacheType === _that.cacheType.ZSET) {
              _that.hashResult = [];
              for (let index in data) {
                _that.hashResult.push({"member": data[index].Member, "score": data[index].Score});
              }
              _that.cache.strShowType = 0;
            } else if (_that.cache.cacheType === _that.cacheType.STRING) {
              _that.searchResult = _that.transResponseData(data);
              _that.searchSourceResult = _that.searchResult;
              _that.cache.strShowType = 1;
            }
            _that.addCacheInit();
          })
        } else {
          //清空右侧数据
          _that.cacheInit();
          _that.error('获取缓存类型失败，缓存可能已不存在');
        }
      })

    },
    getCacheHistory: function () {
      let historyListTemp = localStorage.getItem(this.redisCheck + 'historyList');
      this.historyList = JSON.parse(historyListTemp);
    },
    deleteHistory: function (params) {
      let listTemp = this.historyList;
      if (!listTemp) {
        listTemp = [];
      }
      let saveList = [];
      for (let key in listTemp) {
        if (listTemp[key].Search !== params.Search) {
          saveList.push(listTemp[key]);
        }
      }
      this.historyList = saveList;
      localStorage.setItem(this.redisCheck + 'historyList', JSON.stringify(saveList));
    },
    setCacheHistory: function (params) {
      if (params.Search === '') {
        return
      }
      let listTemp = this.historyList;
      if (!listTemp) {
        listTemp = [];
      }
      for (let key in listTemp) {
        if (listTemp[key].Search === params.Search) {
          return;
        }
      }
      listTemp.push(params);
      this.historyList = listTemp;
      localStorage.setItem(this.redisCheck + 'historyList', JSON.stringify(listTemp));
    },
    //搜索缓存 这里是模糊查询 会返回多个
    keysSearch: function () {
      let _that = this;
      if (this.keys === '') {
        return false;
      } else {
        this.historyCheck = this.keys;
      }
      this.loading = true;
      redis.RedisKeys(this.redisCheck, '*' + this.keys + '*', function (response) {
        _that.keysResult = response.Data;
        _that.initRedisList()
        _that.sortRedisList()
        if (_that.keysResult.length === 1) {
          _that.search(_that.keysResult[0].CacheKey);
        }
        //清空
        if (_that.keysResult.length === 0) {
          _that.cacheInit();
        }
        //记录查询key
        if (_that.keys !== '') {
          _that.setCacheHistory({Search: _that.keys});
        }
        //查找类型
        _that.loading = false;
      })
    },
    searchHistory: function (params) {
      this.keys = params.Search;
      this.UniKey = params.UniKey;
      this.historyCheck = params.Search;
      this.keysSearch();

    },
    serialize: function () {

    },
    unserialize: function () {
      let _that = this;
      if (this.cache.strHasSerialize === true) {
        php.PhpUnserialize(this.searchSourceResult, function (response) {
          if (response.ErrCode !== 0) {
            _that.cache.strHasSerialize = false;
            _that.cache.strShowType = 1;
          } else {
            _that.searchResult = _that.transResponseData(response.Data);
            _that.cache.strShowType = 2;
          }
        })
      } else {
        this.searchResult = this.searchSourceResult;
        _that.cache.strShowType = 1;
        if (this.cache.strHasJson === true) {
          this.json();
        }
      }
    },
    editSubUnserialize: function () {
      let _that = this;
      if (this.editSubCache.strHasSerialize === true) {
        php.PhpUnserialize(this.editSubCache.value, function (response) {
          if (response.ErrCode !== 0) {
            _that.editSubCache.strHasSerialize = false;
            _that.editSubCache.strShowType = 1;
          } else {
            _that.editSubCache.searchResult = _that.transResponseData(response.Data);
            _that.editSubCache.strShowType = 2;
          }

        });
      } else {
        _that.editSubCache.searchResult = _that.editSubCache.value;
        _that.editSubCache.strShowType = 1;
        if (_that.editSubCache.strHasJson === true) {
          this.editSubJson();
        }
      }
    },
    editSubJson: function () {
      if (this.editSubCache.strHasJson === true) {
        this.editSubCache.searchResult = JSON.parse(this.editSubCache.searchResult);
        this.editSubCache.strShowType = 3;
      } else {
        this.editSubCache.searchResult = this.editSubCache.value;
        this.editSubCache.strShowType = 1;
        if (this.editSubCache.strHasSerialize === true) {
          this.editSubUnserialize();
        }
      }
    },
    json: function () {
      if (this.cache.strHasJson === true) {
        this.searchResult = JSON.parse(this.searchResult);
        this.cache.strShowType = 3;
      } else {
        this.searchResult = this.searchSourceResult;
        this.cache.strShowType = 1;
        if (this.cache.strHasSerialize === true) {
          this.unserialize();
        }
      }
    },
    saveString: function () {
      let _that = this;
      redis.RedisSaveString(this.redisCheck, this.cache.cacheKey, this.searchResult, function (response) {
        _that.success('保存成功');
      })
    },
    delCache: function () {
      let _that = this;
      redis.RedisDelKey(this.redisCheck, this.cache.cacheKey, function (response) {
        _that.success('删除成功');
        let newKeysList = [];
        for (let k in _that.keysResult) {
          if (_that.keysResult[k].CacheKey !== _that.cache.cacheKey) {
            newKeysList.push(_that.keysResult[k]);
          }
        }
        _that.keysResult = newKeysList;
        _that.cacheInit();
      })
    },
    funcEditSubCache: function () {
      if (this.editSubCache.strShowType !== 1) {
        this.error('请取消格式化或序列化');
        return false;
      }
      let _that = this;
      redis.RedisEditSub(this.redisCheck, this.editSubCache.cacheKey, this.editSubCache.cacheType,
        this.editSubCache.field, this.editSubCache.value, this.editSubCache.key, this.editSubCache.score,
        this.editSubCache.member, function (response) {
          _that.success('修改成功');
          _that.search(_that.cache.cacheKey);
          _that.editCacheClass = false;
        })
    },
    transResponseData: function (data) {
      let returnDataType = Object.prototype.toString.call(data);
      if (returnDataType === '[object Array]' || returnDataType === '[object Object]') {
        return JSON.stringify(data);
      } else {
        return data;
      }
    },
    copyKey: function () {
      let clipboard = new Clipboard('.copyCacheKey')
      let _that = this
      clipboard.on('success', e => {
        _that.$helperNotify.success('复制成功');
      })
      clipboard.on('error', e => {
        // 不支持复制
        this.error('复制失败，尝试使用chrome');
      })
    },
    refresh: function () {
      this.keysResult = [];
      let _that = this;
      this.cacheInit();
      setTimeout(function () {
        _that.keysSearch()
      }, 500)

    },
    editTTL: function () {
      this.cache.startEditTTL = true;
    },
    cancelEditTTL: function (e) {
      if (this.checkNumber(this.cache.ttl) === false) {
        return
      }
      this.cache.startEditTTL = false;
      try {
        e.stopPropagation();//非IE浏览器
      } catch (e) {
        window.event.cancelBubble = true;//IE浏览器
      }
    },
    saveTTL: function () {
      let _that = this;
      if (this.checkNumber(this.cache.ttl) === false) {
        return
      }
      redis.RedisEditTtl(this.redisCheck, this.cache.cacheKey, parseInt(this.cache.ttl), function (response) {
        _that.success('修改成功');
        _that.cache.startEditTTL = false;
      })
    },
    checkNumber: function (num) {
      let result = /^[-]?[1-9][0-9]*$/.test(num);
      if (!result) {
        this.error('过期时间必须为整数');
        return false;
      }
      return true;
    },
    //清空右侧的缓存显示内容
    cacheInit: function () {
      this.cache.strShowType = 0;
      this.cache.strHasSerialize = false;
      this.cache.strHasJson = false;
      this.cache.cacheKey = '';
      this.hashResult = [];
      this.cache.cacheType = '';
    },
    delAll: function () {
      let deleteKeysList = [];
      for (var i in this.keysResult) {
        if (this.keysResult[i].boolShow) {
          deleteKeysList.push(this.keysResult[i].CacheKey);
        }
      }
      let _that = this;
      let params = {UniKey: this.redisCheck, Keys: deleteKeysList};
      params.ExecType = 'redis_delete_batch'
      this.$confirm('确定删除' + deleteKeysList.length + '个缓存吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        _that.setLoading(params)
        redis.RedisDelAllKey(this.redisCheck, deleteKeysList, function (response) {
          _that.success('删除成功');
          _that.cacheInit();
          _that.keysSearch();
          _that.cancelLoading(params)
        })
      }).catch(() => {

      });
    },
    showAddCache: function () {
      this.addCacheClass = true;
    },
    createCache: function () {
      let _that = this;
      let params = this.addSubCache;
      params.UniKey = this.redisCheck;
      params.cacheScore = parseFloat(params.cacheScore);
      redis.RedisCreateCache(
        this.redisCheck, params.cacheKey, params.boolCreate, params.cacheType, params.cacheField, params.cacheValue,
        params.lPushValue, params.rPushValue, params.cacheMember, params.cacheScore, function (response) {
          _that.success('创建成功');
          _that.addCacheClass = false;
          if (_that.addSubCache.boolCreate === 1) {
            _that.keysSearch();
          } else {
            _that.search(_that.cache.cacheKey);
          }
        })
    },
    delSub: function (sub) {
      let params = {
        UniKey: this.redisCheck,
        cacheType: this.cache.cacheType,
        cacheKey: this.cache.cacheKey,
        sub: sub + '',
      };
      let _that = this;
      if (this.cache.cacheType === this.cacheType.LIST) {
        this.$confirm('确定删除list中所有值为[' + sub + ']的缓存吗?', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          redis.RedisDelSub(this.redisCheck, params.cacheKey, params.cacheType, params.sub, function (response) {
            _that.success('删除成功');
            _that.search(_that.cache.cacheKey);
          })
        }).catch(() => {
          return false;
        });
      } else {
        redis.RedisDelSub(this.redisCheck, params.cacheKey, params.cacheType, params.sub, function (response) {
          _that.success('删除成功');
          _that.search(_that.cache.cacheKey);
        })
      }
    },
    editSub: function (row) {
      //直接把cache赋值给addCache竟然是引用传值？？？
      this.editSubCache.cacheType = this.cache.cacheType;
      this.editSubCache.cacheKey = this.cache.cacheKey;
      this.editSubCache.strHasSerialize = false;
      this.editSubCache.strHasJson = false;
      this.editSubCache.strShowType = 1;
      this.editSubCache.key = row.key;
      this.editSubCache.index = row.index;
      this.editSubCache.value = row.value;
      this.editSubCache.searchResult = row.value;
      this.editSubCache.member = row.member;
      this.editSubCache.value = row.value;
      this.editSubCache.score = parseFloat(row.score);
      this.editSubCache.field = row.field;//hash的
      this.editCacheClass = true;
    },
    addCacheInit: function () {
      this.addSubCache.boolCreate = 1;
      this.addSubCache.cacheType = '';
      this.addSubCache.cacheKey = '';
      this.addSubCache.cacheField = '';
      this.addSubCache.cacheValue = '';
      this.addSubCache.ttl = 0;
      this.addSubCache.cacheMember = '';
      this.addSubCache.cacheScore = '';
      this.addSubCache.lPushValue = '';
      this.addSubCache.rPushValue = ''
    },
    createSubCache: function () {
      //直接把cache赋值给addCache竟然是引用传值？？？
      this.addSubCache.cacheType = this.cache.cacheType;
      this.addSubCache.cacheKey = this.cache.cacheKey;
      this.addSubCache.boolCreate = 2;
      this.addCacheClass = true;
    },
    success: function (msg) {
      // Message.success(msg);
      this.$notify({title: '提示', message: msg, type: 'success', duration: 1000});
    },
    warning: function (msg) {
      // Message.warning(msg);
      this.$notify({title: '提示', message: msg, type: 'warning', duration: 1000});
    },
    info: function (msg) {
      // Message.info(msg);
      //this.$notify({title: '提示', message: msg});
      this.$notify({title: '提示', message: msg, type: 'info', duration: 1000});
    },
    error: function (msg) {
      // Message.error(msg);
      this.$notify({title: '提示', message: msg, type: 'error', duration: 1000});
    },
    chunk: function (arr, size) {
      let objArr = [];
      let index = 0;
      let objArrLen = arr.length / size;
      for (let i = 0; i < objArrLen; i++) {
        let arrTemp = [];
        for (let j = 0; j < size; j++) {
          arrTemp[j] = arr[index++];
          if (index === arr.length) {
            break;
          }
        }
        objArr[i] = arrTemp;
      }
      return objArr;
    },
    setStore: function (key, value) {
      localStorage.setItem(key, value);
    },
    getStore: function (key) {
      return localStorage.getItem(key);
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
  }
}

</script>

