<template>
  配置git
  <el-button type="primary" link @click="ShowAddGit">添加</el-button>
  <el-button type="primary" link @click="ShowQuickAddGit">快速添加</el-button>
  <el-button type="primary" link @click="ShowGitGroup">Git分组</el-button>
  <el-table :data="state.gitList" style="width: 100%">
    <el-table-column prop="id" label="#id" width="80" />
    <el-table-column prop="name" label="name"  width="120"/>
    <el-table-column prop="ssh_name" label="ssh" width="140"/>
    <el-table-column prop="git_group_name" label="git分组"  width="120"/>
    <el-table-column prop="code_path" label="目录" />
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
      <el-form-item label="分组" :label-width="80">
        <el-select v-model="state.editGitConfig.git_group_id" placeholder="选择分组" style="width: 140px">
          <el-option v-for="item in state.gitGroupList" :key="item.id" :label="item.name" :value="item.id"/>
        </el-select>
      </el-form-item>
<!--      <el-form-item label="指定分支切换（仓库过大）" :label-width="80">-->
<!--        <el-select v-model="state.editGitConfig.assign_check" placeholder="是否指定切换" style="width: 140px">-->
<!--          <el-option key="1" label="指定分支" value="1"/>-->
<!--          <el-option key="0" label="不指定分支" value="0"/>-->
<!--        </el-select>-->
<!--      </el-form-item>-->
      <el-form-item label="目录" :label-width="80">
        <el-input v-model="state.editGitConfig.code_path" autocomplete="off" />
      </el-form-item>
      <el-form-item label="ssh" :label-width="80">
        <el-select v-model="state.editGitConfig.ssh_id" placeholder="选择ssh" style="width: 140px">
          <el-option v-for="item in state.sshList" :key="item.id" :label="item.name" :value="item.id"/>
        </el-select>
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

  <el-dialog v-model="state.dialogEditGitQuick" title="快速查找" width="1000">
    <el-form :inline="true">
      <el-form-item label="搜索目录" :label-width="80">
        <el-input v-model="state.quickDir" autocomplete="off" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="GitQuickList" v-loading="state.loading.quick">
          查找
        </el-button>
      </el-form-item>
      <el-form-item>
        <el-input type="text" v-model="state.filterValue" style="width: 91%" placeholder="输入搜索过滤,空格多个条件" @input="FilterQuickList">
        </el-input>
      </el-form-item>
    </el-form>
    <el-table :data="state.quickFilterKeysResult" style="width: 100%">
      <el-table-column prop="code_path" label="目录" width="300"/>
      <el-table-column prop="ssh_name" label="ssh" />
      <el-table-column label="分组" >
        <template #default="scope">
          <el-select v-model="scope.row.git_group_id" placeholder="选择分组" style="width: 140px">
            <el-option v-for="item in state.gitGroupList" :key="item.id" :label="item.name" :value="item.id"/>
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="操作" >
        <template #default="scope">
          <el-button type="primary" link @click="QuickEditGit(scope.row)">保存</el-button>
        </template>
      </el-table-column>
    </el-table>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditGitQuick = false">取消</el-button>
        <el-button type="primary" @click="EditGit">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="state.dialogGitGroup" title="Git分组" width="1000">
      <git_group ref="git_group"></git_group>
  </el-dialog>

</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive , onActivated } from 'vue';
import ssh_set from '../../utils/base/ssh_set'
import set from '../../utils/base/git_set'
import common from '../../utils/common'
import list from "@/utils/base/list";
import git_group from "@/components/set/git_group.vue";
import Init from "@/utils/base/set_init";
import base from "@/utils/base";
export default defineComponent({
  components: {git_group},
  props: {
  },
  data() {
    return {
    }
  },

  setup() {
    onActivated(() => {
      if(Init.GetIsInit('git') === true){
        GitList()
        GitGroupList()
        SshList()
        Init.DelInit('git')
      }
    });

    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const GitList = function (){
      set.GitList(function (response){
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
    const ShowGitGroup = function (){
      state.dialogGitGroup = true
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
    const EditGit = function (){
      set.GitAdd(state.editGitConfig , function (response){
        if(response.ErrCode === 0){
          GitList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditGit = false
        SetInit()
      })
    }
    const QuickEditGit = function (rowData){
      set.GitAdd(rowData , function (response){
        if(response.ErrCode === 0){
          GitList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        SetInit()
      })
    }
    const DeleteGit = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.GitDelete(rowData , function (response){
          if(response.ErrCode === 0){
            GitList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
          SetInit()
        })
      })
    }

    const GitGroupList = function (){
      set.GitGroupList(function (response){
        if(response.ErrCode === 0){
          state.gitGroupList = response.Data
        }
      })
    }
    const SshList = function (){
      ssh_set.SshList(function (response){
        if(response.ErrCode === 0){
          state.sshList = response.Data
        }
      })
    }
    const SetInit = function(){
      Init.SetIsInit('git')
    }
    const FilterQuickList = function (){
      let searchRet = list.QuickSearch(state.filterValue , [...state.gitQuickList] , ['code_path' , 'name'])
      state.quickFilterKeysResult = searchRet.list
    }
    //固有属性
    const state = reactive({
      sshList :[],
      gitGroupList : [],
      gitList : [],
      dialogEditGit : false,
      editGitConfig : {},
      gitQuickList : [],
      filterValue : '',
      quickFilterKeysResult : [],
      dialogEditGitQuick : false,
      dialogGitGroup: false,
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
      EditGit,
      DeleteGit,
      ShowQuickAddGit,
      ShowGitGroup,
      GitQuickList,
      QuickEditGit,
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