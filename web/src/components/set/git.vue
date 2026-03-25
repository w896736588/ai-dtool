<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">Git 配置管理</h3>
      <p class="set-config-desc">管理仓库目录、SSH 环境与分组映射</p>
      <div class="set-config-actions">
        <pl-button type="primary" @click="ShowAddGit">添加</pl-button>
        <pl-button @click="ShowQuickAddGit">快速添加</pl-button>
        <pl-button @click="ShowGitGroup">Git分组</pl-button>
      </div>
    </div>
    <div class="set-config-table-card">
      <el-table :data="state.gitList" class="set-config-table">
        <el-table-column prop="id" label="#id" width="80" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="ssh_name" label="SSH" width="140" />
        <el-table-column prop="git_group_name" label="分组" width="120" />
        <el-table-column prop="code_path" label="目录" min-width="260">
          <template #default="scope">
            <code class="set-mono">{{ scope.row.code_path }}</code>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <div class="set-op-group">
              <pl-button type="primary" link @click="ShowEditGit(scope.row , true)">复制新增</pl-button>
              <pl-button type="primary" link @click="ShowEditGit(scope.row , false)">编辑</pl-button>
              <pl-button link type="danger" @click="DeleteGit(scope.row)">删除</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

  <el-dialog v-model="state.dialogEditGit" title="编辑Git配置" width="500">
    <el-form>
      <el-form-item label="名称" :label-width="80">
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
      <el-form-item label="SSH" :label-width="80">
        <el-select v-model="state.editGitConfig.ssh_id" placeholder="选择ssh" style="width: 140px">
          <el-option v-for="item in state.sshList" :key="item.id" :label="item.name" :value="item.id"/>
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="state.dialogEditGit = false">取消</pl-button>
        <pl-button type="primary" @click="EditGit">
          保存
        </pl-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="state.dialogEditGitQuick" title="快速查找" width="1000">
    <el-form :inline="true">
      <el-form-item label="搜索目录" :label-width="80">
        <el-input v-model="state.quickDir" autocomplete="off" />
      </el-form-item>
      <el-form-item>
        <pl-button type="primary" @click="GitQuickList" v-loading="state.loading.quick">
          查找
        </pl-button>
      </el-form-item>
      <el-form-item>
        <el-input type="text" v-model="state.filterValue" style="width: 91%" placeholder="输入搜索过滤,空格多个条件" @input="FilterQuickList">
        </el-input>
      </el-form-item>
    </el-form>
    <el-table :data="state.quickFilterKeysResult" class="set-config-table">
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
          <pl-button type="primary" link @click="QuickEditGit(scope.row)">保存</pl-button>
        </template>
      </el-table-column>
    </el-table>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="state.dialogEditGitQuick = false">取消</pl-button>
        <pl-button type="primary" @click="EditGit">
          保存
        </pl-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="state.dialogGitGroup" title="Git分组" width="1000">
      <git_group ref="git_group"></git_group>
  </el-dialog>
  </div>
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
@import "@/css/set_module_unified.css";
</style>

