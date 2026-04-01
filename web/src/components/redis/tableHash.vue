<template>
  <template v-if="state.mainForm.cacheKey !== ''">
    <!-- Key信息头部 -->
    <div class="key-info-header">
      <div class="key-tags">
        <el-tag class="type-tag" :type="getTypeTagType(state.mainForm.cacheType)">
          {{ state.mainForm.cacheType.toUpperCase() }}
        </el-tag>
        <el-tag v-if="state.mainForm.startEditTTL === false" class="ttl-tag" @click="editTTL">
          <el-icon><Timer /></el-icon>
          TTL: {{ state.mainForm.ttl }}s
        </el-tag>
        <el-tag v-if="state.mainForm.startEditTTL === true" class="ttl-edit-tag">
          TTL:
          <input v-model="state.mainForm.ttl" class="ttl-input" type="text"/>
          <pl-button size="small" type="primary" @click="saveTTL">保存</pl-button>
          <pl-button size="small" @click="editTTL">取消</pl-button>
        </el-tag>
        <el-tag class="key-name-tag" @click="copyResult(state.mainForm.cacheKey)">
          <el-icon><DocumentCopy /></el-icon>
          <span v-if="state.mainForm.cacheKey.length > 60">{{ state.mainForm.cacheKey.substr(0, 60) }}...</span>
          <span v-else>{{ state.mainForm.cacheKey }}</span>
        </el-tag>
      </div>
    </div>

    <!-- 操作按钮栏 -->
    <div class="action-toolbar">
      <div class="action-left">
        <pl-button type="primary" size="small" plain @click="CallRefresh">
          <el-icon><Refresh /></el-icon>刷新
        </pl-button>
        <pl-button v-if="state.mainForm.cacheType !== 'string'" size="small" type="success" plain @click="createSubCache">
          <el-icon><Plus /></el-icon>添加子项
        </pl-button>
        <pl-button size="small" type="warning" plain @click="Star">
          <el-icon><Star /></el-icon>收藏
        </pl-button>
        <pl-button size="small" type="danger" plain @click="delCache">
          <el-icon><Delete /></el-icon>删除
        </pl-button>
        <el-divider direction="vertical" v-if="state.mainForm.cacheType === 'string'" />
        <pl-button v-if="state.mainForm.cacheType === 'string'" size="small" plain @click="state.editForm.strHasSerialize = !state.editForm.strHasSerialize;editSubUnserialize();">
          <el-icon><Connection /></el-icon>序列化
        </pl-button>
        <pl-button v-if="state.mainForm.cacheType === 'string'" size="small" plain @click="state.editForm.strHasJson = !state.editForm.strHasJson;editSubJson();">
          <el-icon><Document /></el-icon>Json
        </pl-button>
        <pl-button v-if="state.mainForm.cacheType === 'string'" size="small" plain @click="deepParse();">
          <el-icon><DataAnalysis /></el-icon>深度解析
        </pl-button>
      </div>
      <div class="action-right">
        <el-input v-if="ArrayExist(state.mainForm.cacheType , ['hash' , 'set'])" v-model="state.search" size="small" placeholder="搜索..." class="search-input" clearable>
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <pl-button v-if="ArrayExist(state.mainForm.cacheType , ['hash' , 'set'])" size="small" type="primary" @click="CallSearchList">搜索</pl-button>
        <pl-button v-if="state.isMore === 1" size="small" type="primary" plain @click="CallMoreList">
          <el-icon><Download /></el-icon>加载更多
        </pl-button>
        <pl-button v-if="state.mainForm.cacheType ==='string'" size="small" type="primary" @click="SaveString">
          <el-icon><Check /></el-icon>保存
        </pl-button>
      </div>
    </div>

    <!-- 数据统计 -->
    <div class="data-stats" v-if="ArrayExist(state.mainForm.cacheType , ['hash' , 'list' , 'set' , 'zset'])">
      <el-icon><DataLine /></el-icon>
      共 <strong>{{ state.length }}</strong> 条，已加载 <strong>{{ state.hashList.length }}</strong> 条
    </div>

    <!-- 数据表格 -->
    <el-table v-if="state.mainForm.cacheType !== 'string'" :data="state.hashList" class="data-table" :style="{ height: (state.scrollHeight - 5) + 'px' }" stripe>
      <el-table-column v-if="state.mainForm.cacheType === 'hash'" label="field" prop="value">
        <template #default="scope">
          <span class="field-value">{{ scope.row.field }}</span>
        </template>
      </el-table-column>
      <el-table-column v-if="state.mainForm.cacheType === 'zset'" label="member" prop="member">
        <template #default="scope">
          <span class="member-value">{{ scope.row.member }}</span>
        </template>
      </el-table-column>
      <el-table-column v-if="state.mainForm.cacheType === 'zset'" label="score" prop="score" width="120">
        <template #default="scope">
          <el-tag size="small" type="info">{{ scope.row.score }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column v-if="state.mainForm.cacheType === 'list'" label="index" prop="index" width="80">
        <template #default="scope">
          <el-tag size="small" type="info">{{ scope.row.index }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column v-if="ArrayExist(state.mainForm.cacheType , ['hash' , 'list' , 'set'])" label="value" prop="value">
        <template #default="scope">
          <div class="value-cell" @click="editSub(scope.row)">
            <span v-if="scope.row.value.length > 80 && state.mainForm.cacheType === 'list'" class="value-text">
              {{ scope.row.value.substr(0, 80) }}...
            </span>
            <span v-else-if="scope.row.value.length > 60 && state.mainForm.cacheType !== 'list'" class="value-text">
              {{ scope.row.value.substr(0, 60) }}...
            </span>
            <span v-else class="value-text">{{ scope.row.value }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="scope">
          <pl-button v-if="state.mainForm.cacheType === 'hash'" link type="danger" @click="delSub(scope.row.field)">
            删除
          </pl-button>
          <pl-button v-if="state.mainForm.cacheType === 'zset'" link type="danger" @click="delSub(scope.row.member)">
            删除
          </pl-button>
          <pl-button v-if="state.mainForm.cacheType === 'list' || state.mainForm.cacheType === 'set'" link type="danger" @click="delSub(scope.row.value)">
            删除
          </pl-button>
        </template>
      </el-table-column>
    </el-table>
  </template>

  <!-- String类型编辑区 -->
  <el-form v-if="state.mainForm.cacheType === 'string'" class="string-editor">
    <el-input v-if="state.editForm.strShowType === 1" v-model="state.editForm.value" rows="20" type="textarea" :style="{ height: state.scrollHeight + 'px' }" class="string-textarea"></el-input>
    <el-input v-if="state.editForm.strShowType === 2" v-model="state.editForm.searchResult" readonly rows="20" type="textarea" :style="{ height: state.scrollHeight + 'px' }" class="string-textarea readonly"></el-input>
    <div class="json-viewer" v-if="state.editForm.strShowType === 3">
      <pl-button class="copy-btn" link @click="CopyJson(state.editForm.searchResult)">
        <el-icon><DocumentCopy /></el-icon> 复制
      </pl-button>
      <pre class="json-content" ref="jsonPre">{{ state.editForm.searchResult }}</pre>
    </div>
  </el-form>

  <!-- 编辑弹窗 -->
  <el-dialog v-model="state.dialogShow" :append-to-body="true" title="编辑缓存" width="600px" class="edit-dialog">
    <el-form label-width="80px">
      <el-form-item label="操作">
        <pl-button link type="primary" @click="state.editForm.strHasSerialize = !state.editForm.strHasSerialize;editSubUnserialize();">
          序列化
        </pl-button>
        <pl-button link type="primary" @click="state.editForm.strHasJson = !state.editForm.strHasJson;editSubJson();">
          Json
        </pl-button>
        <pl-button link type="primary" @click="deepParse();">
          深度解析
        </pl-button>
      </el-form-item>
      <el-form-item label="field">
        <el-input v-model="state.editForm.field" autocomplete="off" readonly></el-input>
      </el-form-item>
      <el-form-item style="margin-top: 10px">
        <el-input v-if="state.editForm.strShowType === 1" v-model="state.editForm.value" rows="20" type="textarea"></el-input>
        <el-input v-if="state.editForm.strShowType === 2" v-model="state.editForm.searchResult" readonly rows="20" type="textarea"></el-input>
        <div class="json-viewer" v-if="state.editForm.strShowType === 3">
          <pl-button class="copy-btn" link @click="CopyJson(state.editForm.searchResult)">
            <el-icon><DocumentCopy /></el-icon> 复制
          </pl-button>
          <pre class="json-content" ref="jsonPre">{{ state.editForm.searchResult }}</pre>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <pl-button @click="state.dialogShow = false">取 消</pl-button>
      <pl-button type="primary" @click="funcEditSubCache">确 定</pl-button>
    </template>
  </el-dialog>

  <!-- 新增弹窗 -->
  <el-dialog v-model="state.addCacheClass" :append-to-body="true" title="新增缓存" width="500px" class="add-dialog">
    <el-form label-width="80px">
      <el-form-item label="类型">
        <el-select v-model="state.addSubCache.cacheType" placeholder="选择缓存类型" style="width: 100%">
          <el-option label="字符串 (String)" value="string"></el-option>
          <el-option label="哈希 (Hash)" value="hash"></el-option>
          <el-option label="列表 (List)" value="list"></el-option>
          <el-option label="集合 (Set)" value="set"></el-option>
          <el-option label="有序集合 (ZSet)" value="zset"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="key">
        <el-input v-model="state.addSubCache.cacheKey" autocomplete="off" placeholder="输入key名称"></el-input>
      </el-form-item>
      <el-form-item v-if="state.addSubCache.cacheType === 'hash'" label="field">
        <el-input v-model="state.addSubCache.cacheField" autocomplete="off" placeholder="输入field"></el-input>
      </el-form-item>
      <el-form-item v-if="state.addSubCache.cacheType === 'hash' || state.addSubCache.cacheType === 'string' || (state.addSubCache.cacheType === 'list' && state.addSubCache.boolCreate === 1)" label="value">
        <el-input v-model="state.addSubCache.cacheValue" autocomplete="off" placeholder="输入value"></el-input>
      </el-form-item>
      <el-form-item v-if="state.addSubCache.cacheType === 'list' && state.addSubCache.boolCreate === 2" label="lPush">
        <el-input v-model="state.addSubCache.lPushValue" autocomplete="off" placeholder="左侧插入"></el-input>
      </el-form-item>
      <el-form-item v-if="state.addSubCache.cacheType === 'list' && state.addSubCache.boolCreate === 2" label="rPush">
        <el-input v-model="state.addSubCache.rPushValue" autocomplete="off" placeholder="右侧插入"></el-input>
      </el-form-item>
      <el-form-item v-if="state.addSubCache.cacheType === 'set' || state.addSubCache.cacheType === 'zset'" label="member">
        <el-input v-model="state.addSubCache.cacheMember" autocomplete="off" placeholder="输入member"></el-input>
      </el-form-item>
      <el-form-item v-if="state.addSubCache.cacheType === 'zset'" label="score">
        <el-input v-model="state.addSubCache.cacheScore" autocomplete="off" placeholder="输入score"></el-input>
      </el-form-item>
      <el-form-item v-if="state.addSubCache.boolCreate === 1" label="TTL(秒)">
        <el-input v-model="state.addSubCache.ttl" autocomplete="off" placeholder="-1表示永久"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <pl-button @click="state.addCacheClass = false">取 消</pl-button>
      <pl-button type="primary" @click="createCache">确 定</pl-button>
    </template>
  </el-dialog>

  <!-- 深度解析 -->
  <el-dialog v-model="state.isDeepParse" :append-to-body="true" title="深度解析" width="80%" class="deep-parse-dialog">
    <Decode v-if="state.isDeepParse" :source="state.editForm.value"></Decode>
  </el-dialog>
</template>
<script>
import {defineExpose, defineComponent, inject, defineEmits, getCurrentInstance, reactive} from 'vue';
import { Timer, DocumentCopy, Refresh, Plus, Star, Delete, Connection, Document, DataAnalysis, Download, Check, DataLine, Search } from '@element-plus/icons-vue';
import php from "@/utils/base/php";
import redis from "@/utils/base/redis";
import copy from "@/utils/base/copy";
import array from "@/utils/base/array";
import Decode from "@/components/tools/Decode.vue";

export default defineComponent({
  components: {Decode, Timer, DocumentCopy, Refresh, Plus, Star, Delete, Connection, Document, DataAnalysis, Download, Check, DataLine, Search},
  props: {
  },
  data() {
    return {}
  },
  setup() {
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const _callRefresh = inject('callRefresh');
    const _star = inject('callStar')
    const _callMoreList = inject('callMoreList')
    //展示列表
    const ShowList = function (redisChooseId, cacheType, hashList, cacheKey, ttl , length , cursor , isMore) {
      //缓存本身属性
      state.mainForm.cacheKey = cacheKey
      state.mainForm.cacheType = cacheType
      state.mainForm.startEditTTL = false
      state.length = parseInt(length)
      state.isMore = parseInt(isMore)
      state.cursor = parseInt(cursor)
      state.mainForm.ttl = ttl
      //额外信息
      state.redisChooseId = redisChooseId
      //特殊处理
      if (cacheType === 'string') {
        editSub(hashList)
      } else {
        state.hashList = hashList
      }
    };
    const ArrayExist = function (key, arrayList) {
      return array.Exist(key, arrayList)
    };
    // 获取类型标签样式
    const getTypeTagType = function (cacheType) {
      const typeMap = {
        'string': 'success',
        'hash': 'primary',
        'list': 'warning',
        'set': 'info',
        'zset': 'danger'
      }
      return typeMap[cacheType] || 'info'
    };
    //加载更多
    const CallMoreList = function (){
      if(state.search !== ''){
        _callMoreList([] , 0 , state.search)
      }else{
        _callMoreList(state.hashList , state.cursor)
      }
    }
    //搜索
    const CallSearchList = function (){
      if(state.search !== ''){
        _callMoreList([] , 0 , state.search)
      }else{
        _callMoreList([] , state.cursor)
      }
    }
    //刷新
    const CallRefresh = function () {
      _callRefresh(state.mainForm.cacheKey)
    };
    //star
    const Star = function () {
      _star({cacheKey: state.mainForm.cacheKey})
    };
    //json
    const editSubJson = function () {
      if (state.editForm.strHasJson === true) {
        state.editForm.searchResult = JSON.parse(
            state.editForm.searchResult
        )
        state.editForm.strShowType = 3
      } else {
        state.editForm.searchResult = state.editForm.value
        state.editForm.strShowType = 1
        if (state.editForm.strHasSerialize === true) {
          editSubUnserialize()
        }
      }
    };
    //深度解析
    const deepParse = function(){
      state.isDeepParse = true
      console.log(state.editForm.value)
    }
    const createCache = function () {
      let params = state.addSubCache
      params.UniKey = state.redisChooseId
      params.cacheScore = parseFloat(params.cacheScore)
      redis.RedisCreateCache(
          {id:state.redisChooseId},
          params.cacheKey,
          params.boolCreate,
          params.cacheType,
          params.cacheField,
          params.cacheValue,
          params.lPushValue,
          params.rPushValue,
          params.cacheMember,
          params.cacheScore,
          function (response) {
            instance.$helperNotify.success('创建成功')
            state.addCacheClass = false
            CallRefresh()
          }
      )
    };
    const saveTTL = function () {
      let result = /^[-]?[1-9][0-9]*$/.test(state.mainForm.ttl)
      if (!result) {
        instance.$helperNotify.error('过期时间必须为整数')
        return
      }
      redis.RedisEditTtl(
          {id:state.redisChooseId},
          state.mainForm.cacheKey,
          parseInt(state.mainForm.ttl),
          function (response) {
            instance.$helperNotify.success('修改成功')
            state.mainForm.startEditTTL = false
          }
      )
    };
    const SaveString = function () {
      if (state.editForm.strShowType !== 1) {
        instance.$helperNotify.error('请取消格式化或序列化')
        return false
      }
      redis.RedisSaveString({id:state.redisChooseId}, state.mainForm.cacheKey, state.editForm.value, function (response) {
            instance.$helperNotify.success('保存成功')
          }
      )
    };
    //copy
    const copyResult = function (copyString) {
      let index = copy.SetCopyContent(copyString)
      copy.handleCopy(index)
      instance.$helperNotify.success('复制成功')
    };
    //反序列化
    const editSubUnserialize = function () {
      if (state.editForm.strHasSerialize === true) {
        php.PhpUnserialize(state.editForm.value, function (response) {
          if (response.ErrCode !== 0) {
            state.editForm.strHasSerialize = false
            state.editForm.strShowType = 1
          } else {
            state.editForm.searchResult = transResponseData(
                response.Data
            )
            state.editForm.strShowType = 2
          }
        })
      } else {
        state.editForm.searchResult = state.editForm.value
        state.editForm.strShowType = 1
        if (state.editForm.strHasJson === true) {
          this.editSubJson()
        }
      }
    };

    //编辑ttl
    const editTTL = function () {
      state.mainForm.startEditTTL = !state.mainForm.startEditTTL
    };
    //转换属性
    const transResponseData = function (data) {
      let returnDataType = Object.prototype.toString.call(data)
      if (returnDataType === '[object Array]' || returnDataType === '[object Object]') {
        return JSON.stringify(data)
      } else {
        return data
      }
    };
    //删除主元素
    const delCache = function () {
      redis.RedisDelKey(
          {id:state.redisChooseId},
          state.mainForm.cacheKey,
          function (response) {
            instance.$helperNotify.success('删除成功')
            CallRefresh()
          }
      )
    };
    //删除子元素 hash set zset list
    const delSub = function (sub) {
      let params = {
        UniKey: state.redisChooseId,
        cacheType: state.mainForm.cacheType,
        cacheKey: state.mainForm.cacheKey,
        sub: sub + '',
      }
      if (state.mainForm.cacheType === 'list') {
        proxy.$confirm('确定删除list中所有值为[' + sub + ']的缓存吗?', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        })
            .then(() => {
              redis.RedisDelSub({id:state.redisChooseId}, params.cacheKey, params.cacheType, params.sub, function (response) {
                    instance.$helperNotify.success('删除成功')
                    CallRefresh()
                  }
              )
            })
            .catch(() => {
              return false
            })
      } else {
        redis.RedisDelSub({id:state.redisChooseId}, params.cacheKey, params.cacheType, params.sub, function (response) {
              instance.$helperNotify.success('删除成功')
              CallRefresh()
            }
        )
      }
    };
    //编辑保存
    const funcEditSubCache = function () {
      if (state.editForm.strShowType !== 1) {
        instance.$helperNotify.error('请取消格式化或序列化')
        return false
      }
      redis.RedisEditSub(
          {id:state.redisChooseId},
          state.mainForm.cacheKey,
          state.mainForm.cacheType,
          state.editForm.field,
          state.editForm.value,
          state.editForm.key,
          state.editForm.score,
          state.editForm.member,
          function (response) {
            instance.$helperNotify.success('修改成功')
            CallRefresh()
            state.dialogShow = false
          }
      )
    };
    const WindowChange = function (height){
      state.scrollHeight = height - 50
    }
    //编辑缓存
    const editSub = function (value) {
      state.editForm.cacheType = state.mainForm.cacheType
      state.editForm.cacheKey = state.mainForm.cacheKey
      state.editForm.strHasSerialize = false
      state.editForm.strHasJson = false
      state.editForm.strShowType = 1 //1原始输入框 可以编辑保存  2 反序列化 3 json解码  2和3都不能编辑
      state.editForm.key = value.key
      state.editForm.index = value.index
      state.editForm.value = value.value
      state.editForm.searchResult = value.value
      state.editForm.member = value.member
      state.editForm.value = value.value
      state.editForm.score = parseFloat(value.score)
      state.editForm.field = value.field //hash的
      if (state.mainForm.cacheType !== 'string') {
        state.dialogShow = true
      }
    };
    const createSubCache = function () {
      state.addSubCache.cacheType = state.mainForm.cacheType
      state.addSubCache.cacheKey = state.mainForm.cacheKey
      state.addSubCache.boolCreate = 2
      state.addCacheClass = true
    };
    const CopyJson = function(copyContent){
      let index = copy.SetCopyContent(copyContent)
      copy.handleCopy(index)
    }
    //固有属性
    const state = reactive({
      hashList: [], //hash列表
      cursor : 0, //hash 或者list或者zset的游标
      length :0 ,//hash 或者list或者zset的长度
      isMore : 0, //hash或者list或者zset是否还有更多
      search : '', //hash的列表搜索内容
      dialogShow: false,//展示编辑弹窗
      redisChooseId: '', //当前操作的redis唯一值
      addCacheClass: false, //添加子元素弹窗开关
      isDeepParse : false, //是否深度解析
      scrollHeight : 0,
      mainForm: { //主属性
        startEditTTL: false, //开始编辑TTL开关
        ttl: 0,//当前剩余描述
        cacheKey: '', //缓存的主key
        cacheType: '', //缓存类型
      },
      addSubCache: { //新增子元素form
        boolCreate: 1, //1：外部新增一个list   2：list中增加一个值   3 ：编辑list中的一个值
        cacheType: '', //string hash
        cacheKey: '',
        cacheField: '',
        cacheValue: '',
        ttl: 0, //默认永久
        cacheMember: '', //集合的值
        cacheScore: '', //有序集合分值
        lPushValue: '',
        rPushValue: '',
      },//添加子元素
      editForm: { //编辑表单
        cacheKey: '',
        cacheType: '',
        key: 0, //list
        value: '',
        field: '', //哈希
        strShowType: 1, //编辑用的 string和list 才有： 1 textarea （原值） , 2 反序列化 , 3 json展示
        strHasSerialize: false, //是否序列化
        strHasJson: false, //是否json展示
        searchResult: '', //json  和 序列化后的值
        member: '',
        score: 0,
      },
    })
    return {
      state,
      ShowList,
      editSub,
      editSubUnserialize,
      editSubJson,
      deepParse,
      funcEditSubCache,
      delSub,
      SaveString,
      Star,
      CallRefresh,
      editTTL,
      saveTTL,
      copyResult,
      delCache,
      createSubCache,
      createCache,
      ArrayExist,
      WindowChange,
      CallMoreList,
      CallSearchList,
      CopyJson,
      getTypeTagType,
    }
  },
  mounted() {

  },
  methods: {
    confirmClick() {
      this.$emit('confirmClick')
    },
  },
})
</script>

<style scoped>
.key-info-header {
  padding: 14px;
  background: #f7f7f2;
  border: 1px solid #e8e8e0;
  border-radius: 10px;
  margin-bottom: 12px;
}

.key-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.type-tag {
  font-weight: 600;
  border-radius: 8px;
}

.ttl-tag {
  cursor: pointer;
  border-radius: 8px;
  background: #fff;
  border: 1px solid #dde3d8;
  display: flex;
  align-items: center;
  gap: 6px;
}

.ttl-tag:hover {
  border-color: #93b793;
  color: #3f6f3f;
}

.ttl-edit-tag {
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.ttl-input {
  width: 80px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 4px 8px;
  font-size: 13px;
}

.key-name-tag {
  cursor: copy;
  border-radius: 8px;
  background: #fff;
  border: 1px solid #dde3d8;
  max-width: 520px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
}

.key-name-tag:hover {
  border-color: #93b793;
  background: #f2f7ee;
}

.action-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: #fff;
  border-radius: 10px;
  margin-bottom: 10px;
  border: 1px solid #e8e8e0;
  flex-wrap: wrap;
  gap: 8px;
}

.action-left, .action-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.search-input {
  width: 200px;
}

.data-stats {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: #f2f7ee;
  border: 1px solid #e0eadb;
  border-radius: 8px;
  margin-bottom: 10px;
  font-size: 13px;
  color: #606050;
}

.data-stats strong {
  color: #3f6f3f;
}

.data-table {
  border-radius: 10px;
  overflow: hidden;
}

.data-table :deep(.el-table__header-wrapper) {
  background: #f7f7f2;
}

.data-table :deep(.el-table__header th) {
  background: #f7f7f2;
  color: #606050;
  font-weight: 600;
}

.data-table :deep(.el-table__row:hover > td) {
  background-color: #f3f7ef !important;
}

.field-value, .member-value {
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
}

.value-cell {
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
}

.value-cell:hover {
  background: #eef4ea;
}

.value-text {
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
  color: #4f804f;
}

.string-editor {
  margin-top: 10px;
}

.string-textarea {
  border-radius: 10px;
}

.string-textarea :deep(.el-textarea__inner) {
  border-radius: 10px;
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
  line-height: 1.6;
  background: #fafbf8;
  border: 1px solid #dde3d8;
}

.string-textarea.readonly :deep(.el-textarea__inner) {
  background: #f4f7f2;
}

.json-viewer {
  position: relative;
  background: #1f221d;
  border-radius: 10px;
  padding: 14px;
  min-height: 320px;
}

.copy-btn {
  position: absolute;
  top: 10px;
  right: 10px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.18);
  color: #fff;
  padding: 5px 10px;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
}

.copy-btn:hover {
  background: rgba(255, 255, 255, 0.18);
}

.json-content {
  color: #d7dfd1;
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
  line-height: 1.55;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  max-height: 600px;
  overflow: auto;
}

</style>

