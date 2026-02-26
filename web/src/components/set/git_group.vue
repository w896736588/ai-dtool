<template>
  配置git组 <el-button type="primary" link @click="ShowAddGitGroup">添加</el-button>
  <el-table :data="state.gitGroupList" style="width: 100%">
    <el-table-column prop="id" label="#id" />
    <el-table-column prop="name" label="name" />
    <el-table-column label="操作" >
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditGitGroup(scope.row)">编辑</el-button>
        <el-button link type="danger" @click="DeleteGitGroup(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>
  <p></p>
  <el-dialog v-model="state.dialogEditGitGroup" title="编辑" width="500">
    <el-form>
      <el-form-item label="组名" :label-width="80">
        <el-input v-model="state.editGitGroupConfig.name" autocomplete="off" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditGitGroup = false">取消</el-button>
        <el-button type="primary" @click="EditGitGroup">
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
import Init from "@/utils/base/set_init";
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
      set.GitList(function (response){
        if(response.ErrCode === 0){
          state.gitList = response.Data
        }
      })
    }
    const ShowEditGit = function (gitConfig){
      state.dialogEditGit = true
      state.editGitConfig = gitConfig
    }
    const ShowAddGit = function (){
      state.dialogEditGit = true
      state.editGitConfig = {}
    }
    const GitQuickList = function (){
      state.loading.quick = true
      set.GitQuickList({dir : state.quickDir} , function (response) {
        if(response.ErrCode === 0){
          state.gitQuickList = response.Data
          state.quickFilterKeysResult = state.gitQuickList
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.loading.quick = false
      })
    }
    const SetInit = function(){
      Init.SetIsInit('git')
    }

    const GitGroupList = function (){
      set.GitGroupList(function (response){
        if(response.ErrCode === 0){
          state.gitGroupList = response.Data
        }
      })
    }
    const ShowEditGitGroup = function (gitGroupConfig){
      state.dialogEditGitGroup = true
      state.editGitGroupConfig = gitGroupConfig
    }
    const ShowAddGitGroup = function (){
      state.dialogEditGitGroup = true
      state.editGitGroupConfig = {}
    }
    const EditGitGroup = function (){
      set.GitGroupAdd(state.editGitGroupConfig , function (response){
        if(response.ErrCode === 0){
          GitGroupList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditGitGroup = false
        SetInit()
      })
    }
    const DeleteGitGroup = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.GitGroupDelete(rowData , function (response){
          if(response.ErrCode === 0){
            GitGroupList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
          SetInit()
        })
      })
    }
    const SshList = function (){
      ssh_set.SshList(function (response){
        if(response.ErrCode === 0){
          state.sshList = response.Data
        }
      })
    }
    const FilterQuickList = function (){
      let searchRet = list.QuickSearch(state.filterValue , [...state.gitQuickList] , ['code_path' , 'name'])
      state.quickFilterKeysResult = searchRet.list
    }
    //固有属性
    const state = reactive({
      sshList :[],
      gitGroupList : [],
      dialogEditGitGroup : false,
      editGitGroupConfig : {},
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
    GitGroupList()
    SshList()
    return {
      state,
      ShowEditGit,
      ShowAddGit,
      ShowEditGitGroup,
      ShowAddGitGroup,
      EditGitGroup,
      DeleteGitGroup,
      GitQuickList,
      FilterQuickList,
      GitList,
      GitGroupList,
      SshList,
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