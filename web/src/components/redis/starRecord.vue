<template>
  <el-drawer v-model="state.drawerHistoryShow" direction="rtl" size="60%">
    <template #header>
      <h4>收藏key列表</h4>
    </template>
    <template #default>
      <div>
        <el-input
            type="text"
            v-model="state.filterValue"
            style="width: 91%"
            placeholder="输入搜索过滤,空格多个条件"
            @input="filterList"
        ></el-input>
        <el-table :data="state.filterStarList" stripe style="width: 100%;">
          <el-table-column prop="name" label="name" width="300" style="font-size:12px;"/>
          <el-table-column label="key" >
            <template #default="scope">
              <el-button link type="primary" size="small" @click="CallStarListSearch(scope.row)">
                {{scope.row.key}}
              </el-button>
            </template>
          </el-table-column>
          <el-table-column label="op" width="150">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="copyKey(scope.row.key)">
                复制
              </el-button>
              <el-popconfirm
                  cancel-button-text="取消"
                  confirm-button-text="删除"
                  icon-color="#626AEF"
                  title="确定删除吗?"
                  @confirm="starDelete(scope.row)"
              >
                <template #reference>
                <el-button link type="danger" size="small" >
                  删除
                </el-button>
                </template>
              </el-popconfirm>
              <el-button link type="primary" size="small" @click="starEdit(scope.row)">
                编辑
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>
    <template #footer>
      <div style="flex: auto">
      </div>
    </template>
  </el-drawer>

  <el-dialog v-model="state.dialogStarCache" title="收藏缓存key" width="500">
    <el-form :model="state.starForm">
      <el-form-item label="name" :label-width="80">
        <el-input v-model="state.starForm.name" autocomplete="off" />
      </el-form-item>
      <el-form-item label="key" :label-width="80">
        <el-input v-model="state.starForm.key" autocomplete="off" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogStarCache = false">取消</el-button>
        <el-button type="primary" @click="starSave">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive} from 'vue';
import list from '../../utils/base/list'
import star_api from '../../utils/base/star'
import copy from '@/utils/base/copy'
export default defineComponent({
  props: {
  },
  data() {
    return {
    }
  },
  setup() {
    const instance = getCurrentInstance().appContext.config.globalProperties
    const _callStarListSearch = inject('callStarListSearch');
    //点击搜索
    const CallStarListSearch = function (value){
      _callStarListSearch(value)
      state.drawerHistoryShow = false
    };
    const copyKey = function (key){
      let index = copy.SetCopyContent(key)
      copy.handleCopy(index)
    }
    //收藏方法
    const star = function(value) {
      GetStarList()
      state.dialogStarCache = true
      let searchValue = list.SearchSetValue(state.starList , 'key' , {key : value.cacheKey})
      if(searchValue.key){
        state.starForm.name = searchValue.name
        state.starForm.key = value.cacheKey
      }else{
        state.starForm.name = ''
        state.starForm.key = value.cacheKey
      }
    };
    const starEdit = function(value) {
      state.dialogStarCache = true
      state.starForm = value
    };
    //收藏保存
    const starSave = function (){
      if(state.starForm.name === '' || state.starForm.key === ''){
        instance.$helperNotify.error('name和key不能为空')
        return
      }
      star_api.StarAdd(state.starForm.id , state.starForm.name , state.starForm.key , state.starForm.key , 'redis' , function (response){

      })
      instance.$helperNotify.success('success')
      state.dialogStarCache = false
      GetStarList()
    };
    //删除收藏
    const starDelete = function (value){
      star_api.StarDel(value.id, function (response){})
      GetStarList()
    };
    //展示列表方法
    const showStarList  = function (){
      state.drawerHistoryShow = !state.drawerHistoryShow
      GetStarList()
    };
    //筛选
    const filterList = function (){
      let searchRet = list.QuickSearch(state.filterValue , [...state.starList] , ['key' , 'name'])
      state.filterStarList = searchRet.list
    };
    //固有属性
    const state = reactive({
      drawerHistoryShow: false, //展示抽屉
      dialogStarCache : false,//展示弹窗
      starList: [], //收藏列表
      filterStarList : [], //过滤后的列表
      starListLocalKey : 'redisKeyStarListV3',
      filterValue : '', //搜索的值
      starForm : { //编辑表单
        id : '',
        name : '',
        key : '',
        type : 'redis',
        value : '',
      },
    })
    //初始化
    const GetStarList = function () {
      star_api.StarList('redis' , function (response){
        if(response.ErrCode === 1){
          return
        }
        state.starList = response.Data
        filterList()
      })
    };

    return {
      star,
      starEdit,
      state,
      starSave,
      starDelete,
      showStarList,
      filterList,
      CallStarListSearch,
      GetStarList,
      copyKey,
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

<style >
</style>