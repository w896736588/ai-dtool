<template>
  gitlab token
  <el-button type="primary" link @click="ShowAddGit">添加</el-button>
  <el-table :data="state.gitList" style="width: 100%">
    <el-table-column prop="id" label="#id" width="80" />
    <el-table-column prop="name" label="name"  width="120"/>
    <el-table-column prop="url" label="url" />
    <el-table-column prop="access_token" label="access_token"  />
    <el-table-column label="操作" width="200">
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditGit(scope.row , true)">复制新增</el-button>
        <el-button type="primary" link @click="ShowEditGit(scope.row , false)">编辑</el-button>
        <el-button link type="danger" @click="DeleteGit(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>

  <el-dialog v-model="state.dialogEditGit" title="编辑" width="500">
    <el-form>
      <el-form-item label="name" :label-width="80">
        <el-input v-model="state.editGitConfig.name" autocomplete="off" />
      </el-form-item>
      <el-form-item label="url" :label-width="80">
        <el-input v-model="state.editGitConfig.url" autocomplete="off" />
      </el-form-item>
      <el-form-item label="access_token" :label-width="80">
        <el-input v-model="state.editGitConfig.access_token" autocomplete="off" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditGit = false">取消</el-button>
        <el-button type="primary" @click="EditGit">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive} from 'vue';
import ssh_set from '../../utils/base/ssh_set'
import set from '../../utils/base/git_set'
import common from '../../utils/common'
import list from "@/utils/base/list";
export default defineComponent({
  props: {
  },
  data() {
    return {
    }
  },
  setup() {
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const GitList = function (){
      set.GitlabTokenList(function (response){
        if(response.ErrCode === 0){
          state.gitList = response.Data
        }
      })
    }
    const ShowEditGit = function (gitConfig , isCopy){
      state.dialogEditGit = true
      state.editGitConfig = gitConfig
      if(isCopy){
        state.editGitConfig.id = 0
      }
    }
    const ShowAddGit = function (){
      state.dialogEditGit = true
      state.editGitConfig = {}
    }
    const ShowQuickAddGit = function (){
      state.dialogEditGitQuick = true

    }
    const EditGit = function (){
      set.GitlabTokenAdd(state.editGitConfig , function (response){
        if(response.ErrCode === 0){
          GitList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditGit = false
      })
    }
    const DeleteGit = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.GitlabTokenDelete(rowData , function (response){
          if(response.ErrCode === 0){
            GitList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
        })
      })
    }

    const FilterQuickList = function (){
      let searchRet = list.QuickSearch(state.filterValue , [...state.gitQuickList] , ['code_path' , 'name'])
      state.quickFilterKeysResult = searchRet.list
    }
    //固有属性
    const state = reactive({
      gitList : [],
      dialogEditGit : false,
      editGitConfig : {},
      gitQuickList : [],
      filterValue : '',
      quickFilterKeysResult : [],
      dialogEditGitQuick : false,
      quickDir : '',
      loading : {
        quick : false
      }
    })
    //初始化
    GitList()
    return {
      state,
      ShowEditGit,
      ShowAddGit,
      EditGit,
      DeleteGit,
      ShowQuickAddGit,
      FilterQuickList,
      GitList,
    }
  },
  mounted() {

  },
  methods: {
  },
})
</script>

<style scoped>

</style>