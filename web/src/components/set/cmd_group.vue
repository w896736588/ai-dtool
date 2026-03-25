<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">命令组管理</h3>
      <div class="set-config-actions">
        <pl-button type="primary" @click="ShowAddCmdGroup">添加命令组</pl-button>
      </div>
    </div>
    <div class="set-config-table-card">
      <el-table :data="state.CmdGroupList" class="set-config-table">
        <el-table-column prop="id" label="#id" width="80" />
        <el-table-column prop="name" label="组名" min-width="180" />
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <div class="set-op-group">
              <pl-button type="primary" link @click="ShowEditCmdGroup(scope.row)">编辑</pl-button>
              <pl-button link type="danger" @click="DeleteCmdGroup(scope.row)">删除</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog v-model="state.dialogEditCmdGroup" title="编辑命令组" width="500">
      <el-form label-width="80px">
        <el-form-item label="组名">
          <el-input v-model="state.editCmdGroupConfig.name" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogEditCmdGroup = false">取消</pl-button>
          <pl-button type="primary" @click="EditCmdGroup">保存</pl-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive} from 'vue';
import set from '../../utils/base/cmd_set'
import common from '../../utils/common'
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
    const CmdGroupList = function (){
      set.SetCmdGroupList(function (response){
        if(response.ErrCode === 0){
          state.CmdGroupList = response.Data
        }
      })
    }
    const ShowEditCmdGroup = function (CmdGroupConfig){
      state.dialogEditCmdGroup = true
      state.editCmdGroupConfig = CmdGroupConfig
    }
    const ShowAddCmdGroup = function (){
      state.dialogEditCmdGroup = true
      state.editCmdGroupConfig = {}
    }
    const EditCmdGroup = function (){
      set.SetCmdGroupAdd(state.editCmdGroupConfig , function (response){
        if(response.ErrCode === 0){
          CmdGroupList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditCmdGroup = false
      })
    }
    const DeleteCmdGroup = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.SetCmdGroupDelete(rowData , function (response){
          if(response.ErrCode === 0){
            CmdGroupList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
        })
      })
    }
    //固有属性
    const state = reactive({
      sshList :[],
      CmdGroupList : [],
      dialogEditCmdGroup : false,
      editCmdGroupConfig : {},
      quickFilterKeysResult : [],
      dialogEditCmdQuick : false,
      loading : {
        quick : false
      }
    })
    //初始化
    CmdGroupList()
    return {
      state,
      ShowEditCmdGroup,
      ShowAddCmdGroup,
      EditCmdGroup,
      DeleteCmdGroup,
      CmdGroupList,
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

