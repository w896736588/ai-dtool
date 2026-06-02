<template>
  <el-drawer v-model="state.drawerHistoryShow" direction="rtl" size="50%" class="star-drawer">
    <template #header>
      <div class="drawer-header">
        <el-icon class="header-icon"><Star /></el-icon>
        <span>收藏 Key 列表</span>
      </div>
    </template>
    <template #default>
      <div class="drawer-content">
        <div class="search-bar">
          <el-input
            type="text"
            v-model="state.filterValue"
            placeholder="搜索收藏,空格多个条件"
            @input="filterList"
            clearable
            class="search-input"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
        <el-table :data="state.filterStarList" stripe style="width: 100%;" class="star-table">
          <el-table-column prop="name" label="名称" width="200">
            <template #default="scope">
              <div class="name-cell">
                <el-icon class="star-icon"><StarFilled /></el-icon>
                <span class="name-text">{{ scope.row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="Key" min-width="300">
            <template #default="scope">
              <pl-button link type="primary" size="small" @click="CallStarListSearch(scope.row)" class="key-link">
                <el-icon><Key /></el-icon>
                {{ scope.row.key }}
              </pl-button>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="scope">
              <div class="action-cell">
                <pl-button link type="primary" size="small" @click="copyKey(scope.row.key)">
                  <el-icon><DocumentCopy /></el-icon>复制
                </pl-button>
                <el-popconfirm
                  cancel-button-text="取消"
                  confirm-button-text="删除"
                  icon-color="#626AEF"
                  title="确定删除吗?"
                  @confirm="starDelete(scope.row)"
                >
                  <template #reference>
                    <pl-button link type="danger" size="small">
                      <el-icon><Delete /></el-icon>删除
                    </pl-button>
                  </template>
                </el-popconfirm>
                <pl-button link type="primary" size="small" @click="starEdit(scope.row)">
                  <el-icon><Edit /></el-icon>编辑
                </pl-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>
    <template #footer>
      <div class="drawer-footer">
        <span class="footer-count">共 {{ state.filterStarList.length }} 条收藏</span>
      </div>
    </template>
  </el-drawer>

  <el-dialog v-model="state.dialogStarCache" title="收藏缓存Key" width="450px" class="star-dialog">
    <el-form :model="state.starForm" label-width="60px" class="star-form">
      <el-form-item label="名称">
        <el-input v-model="state.starForm.name" autocomplete="off" placeholder="为收藏命名">
          <template #prefix>
            <el-icon><CollectionTag /></el-icon>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item label="Key">
        <el-input v-model="state.starForm.key" autocomplete="off" placeholder="Redis Key">
          <template #prefix>
            <el-icon><Key /></el-icon>
          </template>
        </el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="state.dialogStarCache = false">取消</pl-button>
        <pl-button type="primary" @click="starSave">
          <el-icon><Check /></el-icon>保存
        </pl-button>
      </div>
    </template>
  </el-dialog>
</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive} from 'vue';
import { Star, StarFilled, Search, Key, DocumentCopy, Delete, Edit, CollectionTag, Check } from '@element-plus/icons-vue';
import list from '../../utils/base/list'
import star_api from '../../utils/base/star'
import copy from '@/utils/base/copy'
export default defineComponent({
  components: { Star, StarFilled, Search, Key, DocumentCopy, Delete, Edit, CollectionTag, Check },
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

<style scoped src="@/css/components/redis/starRecord.css"></style>

