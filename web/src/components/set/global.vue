<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">全局配置</h3>
      <p class="set-config-desc">管理全局参数（密钥字段自动脱敏显示）</p>
      <div class="set-config-actions">
        <pl-button type="primary" @click="ShowAddGit">添加</pl-button>
      </div>
    </div>
    <div class="set-config-table-card">
      <el-table :data="state.gitList" class="set-config-table">
        <el-table-column prop="id" label="#id" width="80" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="key" label="Key" min-width="180" />
        <el-table-column prop="value" label="Value" min-width="180">
          <template #default="scope">
            <span v-if="scope.row.key.includes('token') || scope.row.key.includes('password')">******</span>
            <span v-else>{{scope.row.value}}</span>
          </template>
        </el-table-column>
        <el-table-column prop="desc" label="备注" min-width="160" />
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

    <el-dialog v-model="state.dialogEditGit" title="编辑全局配置" width="500">
      <el-form label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="state.editGitConfig.name" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Key">
          <el-input v-model="state.editGitConfig.key" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Value">
          <el-input v-model="state.editGitConfig.value" autocomplete="off" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="state.editGitConfig.desc" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogEditGit = false">取消</pl-button>
          <pl-button type="primary" @click="EditGit">保存</pl-button>
        </div>
      </template>
    </el-dialog>
  </div>
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
      set.GlobalList(function (response){
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
      set.GlobalAdd(state.editGitConfig , function (response){
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
        set.GlobalDelete(rowData , function (response){
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

<style scoped src="@/css/components/set/global.css"></style>

