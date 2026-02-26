<template>
  <template v-if="state.mainForm.cacheKey !== ''">
    <el-tag class="el-tag-he">{{ state.mainForm.cacheType }}</el-tag>
    <el-tag v-if="state.mainForm.startEditTTL === false" class="el-tag-he" style="cursor: pointer" @click="editTTL">
      ttl:{{ state.mainForm.ttl }}
    </el-tag>
    <el-tag v-if="state.mainForm.startEditTTL === true" class="el-tag-he">
      ttl：
      <input v-model="state.mainForm.ttl" style="width: 100px; border: 0" type="text"/>
      <el-button style="padding: 3px" type="primary" size="small" @click="saveTTL">保存</el-button>
      <el-button style="padding: 3px" type="info" size="small" @click="editTTL">取消</el-button>
    </el-tag>
    <el-tag :data-clipboard-text="state.mainForm.cacheKey" class="copyCacheKey el-tag-he" style="cursor: copy" @click="copyResult(state.mainForm.cacheKey)">
      <span v-if="state.mainForm.cacheKey.length > 75">{{ state.mainForm.cacheKey.substr(0, 75) }}...</span>
      <span v-else>{{ state.mainForm.cacheKey }}</span>
    </el-tag>
    &nbsp;
    <p>
      <el-button type="primary" size="small" @click="CallRefresh">刷新</el-button>
      <el-button v-if="state.mainForm.cacheType !== 'string'" size="small"  type="primary" @click="createSubCache">添加子项
      </el-button>
      <el-button  type="primary" size="small"  @click="Star">收藏</el-button>
      <el-button  type="danger" size="small"  @click="delCache">删除</el-button>
      <el-button v-if="state.mainForm.cacheType === 'string'"  type="primary" size="small"  @click="state.editForm.strHasSerialize = !state.editForm.strHasSerialize;editSubUnserialize();">
        序列化
      </el-button>
      <el-button v-if="state.mainForm.cacheType === 'string'"  type="primary" size="small"  @click="state.editForm.strHasJson = !state.editForm.strHasJson;editSubJson();">
        Json
      </el-button>
      <el-button v-if="state.mainForm.cacheType === 'string'" type="primary" size="small" @click="deepParse();">
        深度解析
      </el-button>
      &nbsp;
      <el-input type="text" size="small" placeholder="输入进行搜索" style="width:200px;" v-if="ArrayExist(state.mainForm.cacheType , ['hash' , 'set'])" v-model="state.search"/>
      <el-button style="margin: 5px;" type="primary" size="small" v-if="ArrayExist(state.mainForm.cacheType , ['hash' , 'set'])"  @click="CallSearchList">搜索</el-button>
      <el-button style="margin: 5px;float:right;" type="primary" size="small" v-if="state.mainForm.cacheType ==='string'"  @click="SaveString">保 存</el-button>
      <el-button v-if="state.isMore === 1" style="margin: 5px;float:right;" type="primary" size="small"  @click="CallMoreList">加载更多</el-button>
      <span style="font-size: 13px;" v-if="ArrayExist(state.mainForm.cacheType , ['hash' , 'list' , 'set' , 'zset'])">&nbsp;共{{state.length}}条，已加载{{state.hashList.length}}条</span>
    </p>


    <el-table v-if="state.mainForm.cacheType !== 'string'" :data="state.hashList" class="cache-table" :style="{ height: (state.scrollHeight - 5) + 'px' }">
      <el-table-column v-if="state.mainForm.cacheType === 'hash'" label="field" prop="value">
        <template #default="scope">
<!--          <p v-if="scope.row.field.length > 50" aria-placeholder="scope.row.field">-->
<!--            {{ scope.row.field.substr(0, 50) }}...-->
<!--          </p>-->
<!--          <p v-if="scope.row.field.length <= 50">-->
            {{ scope.row.field }}
<!--          </p>-->

        </template>
      </el-table-column>
      <el-table-column v-if="state.mainForm.cacheType === 'zset'" label="member" prop="member"></el-table-column>
      <el-table-column v-if="state.mainForm.cacheType === 'zset'" label="score" prop="score"></el-table-column>
      <el-table-column v-if="state.mainForm.cacheType === 'list'" label="index" prop="index" width="80"></el-table-column>
      <el-table-column v-if="ArrayExist(state.mainForm.cacheType , ['hash' , 'list' , 'set'])" label="value" prop="value" style="cursor: pointer;">
        <template #default="scope">
          <p v-if="scope.row.value.length > 80 && state.mainForm.cacheType === 'list'" style="cursor:pointer;color:#409eff;" @click="editSub(scope.row)">
            {{ scope.row.value.substr(0, 80) }}...
          </p>
          <p v-if="scope.row.value.length > 60 && state.mainForm.cacheType !== 'list'" style="cursor:pointer;color:#409eff;" @click="editSub(scope.row)">
            {{ scope.row.value.substr(0, 60) }}...
          </p>
          <p v-if="scope.row.value.length <= 60" style="cursor:pointer;color:#409eff;" @click="editSub(scope.row)">
            {{ scope.row.value }}
          </p>

        </template>
      </el-table-column>
      <el-table-column label="操作" width="80">
        <template #default="scope">
          <el-button v-if="state.mainForm.cacheType === 'hash'" link type="primary" @click="delSub(scope.row.field)">
            删除
          </el-button>
          <el-button v-if="state.mainForm.cacheType === 'zset'" link type="primary" @click="delSub(scope.row.member)">
            删除
          </el-button>
          <el-button v-if="state.mainForm.cacheType === 'list' || state.mainForm.cacheType === 'set'" link type="primary" @click="delSub(scope.row.value)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </template>
  <el-form v-if="state.mainForm.cacheType === 'string'" style="margin-top:10px;" >
    <el-input v-if="state.editForm.strShowType === 1" v-model="state.editForm.value" rows="20" type="textarea" :style="{ height: state.scrollHeight + 'px' }"></el-input>
    <el-input v-if="state.editForm.strShowType === 2" v-model="state.editForm.searchResult" readonly rows="20" style="background: #eeee" type="textarea"></el-input>
    <div class="pretty-json-p" v-if="state.editForm.strShowType === 3">
      <button class="copy-btn" @click="CopyJson(state.editForm.searchResult)">复制</button>
      <pre class="pretty-json" ref="jsonPre">{{ state.editForm.searchResult }}</pre>
    </div>

  </el-form>

  <el-dialog v-model="state.dialogShow" :append-to-body="true" title="编辑缓存">
    <el-form>
      <el-form-item label="操作">
        <el-button link type="primary" @click="state.editForm.strHasSerialize = !state.editForm.strHasSerialize;editSubUnserialize();">
          序列化
        </el-button>
        <el-button link type="primary" @click="state.editForm.strHasJson = !state.editForm.strHasJson;editSubJson();">
          Json
        </el-button>
        <el-button link type="primary" @click="deepParse();">
          深度解析
        </el-button>
      </el-form-item>
      <el-form-item label="field">
        <el-input v-model="state.editForm.field" autocomplete="off" readonly></el-input>
      </el-form-item>

      <el-form-item style="margin-top: 10px">
        <el-input v-if="state.editForm.strShowType === 1" v-model="state.editForm.value" rows="20" type="textarea" ></el-input>
        <el-input v-if="state.editForm.strShowType === 2" v-model="state.editForm.searchResult" readonly rows="20" type="textarea" ></el-input>
<!--        <pre class="pretty-json" v-if="state.editForm.strShowType === 3">{{ state.editForm.searchResult }}</pre>-->
        <div class="pretty-json-p" v-if="state.editForm.strShowType === 3">
          <button class="copy-btn" @click="CopyJson(state.editForm.searchResult)">复制</button>
          <pre class="pretty-json" ref="jsonPre">{{ state.editForm.searchResult }}</pre>
        </div>
      </el-form-item>

    </el-form>
    <template #footer>
      <el-button @click="state.dialogShow = false">取 消</el-button>
      <el-button type="primary" @click="funcEditSubCache">确 定</el-button>
    </template>
  </el-dialog>

  <!--新增弹窗-->
  <el-dialog v-model="state.addCacheClass" :append-to-body="true" title="新增缓存" width="70%;">
    <el-form>
      <el-form-item :label-width="100" label="类型">
        <el-select v-model="state.addSubCache.cacheType" placeholder="选择缓存类型">
          <el-option label="字符串" value="string"></el-option>
          <el-option label="哈希" value="hash"></el-option>
          <el-option label="列表" value="list"></el-option>
          <el-option label="集合" value="set"></el-option>
          <el-option label="有序集合" value="zset"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item :label-width="100" label="key">
        <el-input v-model="state.addSubCache.cacheKey" autocomplete="off"></el-input>
      </el-form-item>

      <el-form-item v-if="state.addSubCache.cacheType === 'hash'" :label-width="100" label="field">
        <el-input v-model="state.addSubCache.cacheField" autocomplete="off"></el-input>
      </el-form-item>
      <el-form-item v-if="state.addSubCache.cacheType === 'hash' || state.addSubCache.cacheType === 'string' || (state.addSubCache.cacheType === 'list' && state.addSubCache.boolCreate === 1)" :label-width="100" label="value">
        <el-input v-model="state.addSubCache.cacheValue" autocomplete="off"></el-input>
      </el-form-item>

      <el-form-item v-if="state.addSubCache.cacheType === 'list' && state.addSubCache.boolCreate === 2" :label-width="100" label="lPush">
        <el-input v-model="state.addSubCache.lPushValue" autocomplete="off"></el-input>
      </el-form-item>

      <el-form-item v-if="state.addSubCache.cacheType === 'list' && state.addSubCache.boolCreate === 2" :label-width="100" label="rPush">
        <el-input v-model="state.addSubCache.rPushValue" autocomplete="off"></el-input>
      </el-form-item>

      <el-form-item v-if="state.addSubCache.cacheType === 'set' || state.addSubCache.cacheType === 'zset'" :label-width="100" label="member">
        <el-input v-model="state.addSubCache.cacheMember" autocomplete="off"></el-input>
      </el-form-item>
      <el-form-item v-if="state.addSubCache.cacheType === 'zset'" :label-width="100" label="score">
        <el-input v-model="state.addSubCache.cacheScore" autocomplete="off"></el-input>
      </el-form-item>
      <el-form-item v-if="state.addSubCache.boolCreate === 1" :label-width="100" label="ttl/秒">
        <el-input v-model="state.addSubCache.ttl" autocomplete="off"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="state.addCacheClass = false">取 消</el-button>
      <el-button type="primary" @click="createCache">确 定</el-button>
    </template>
  </el-dialog>


  <!--深度解析-->
  <el-dialog v-model="state.isDeepParse" :append-to-body="true" title="深度解析" width="80%">
    <Decode v-if="state.isDeepParse" :source="state.editForm.value"></Decode>
  </el-dialog>

</template>
<script>
import {defineExpose, defineComponent, inject, defineEmits, getCurrentInstance, reactive} from 'vue';
import php from "@/utils/base/php";
import redis from "@/utils/base/redis";
import copy from "@/utils/base/copy";
import array from "@/utils/base/array";
import Decode from "@/components/tools/Decode.vue";

export default defineComponent({
  components: {Decode},
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

</style>