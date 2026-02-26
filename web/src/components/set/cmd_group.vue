<template>
  配置命令组 <el-button type="primary" link @click="ShowAddCmdGroup">添加</el-button>
  <el-table :data="state.CmdGroupList" style="width: 100%">
    <el-table-column prop="id" label="#id" />
    <el-table-column prop="name" label="name" />
    <el-table-column label="操作" >
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditCmdGroup(scope.row)">编辑</el-button>
        <el-button link type="danger" @click="DeleteCmdGroup(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>
  <p></p>
  <el-dialog v-model="state.dialogEditCmdGroup" title="编辑" width="500">
    <el-form>
      <el-form-item label="组名" :label-width="80">
        <el-input v-model="state.editCmdGroupConfig.name" autocomplete="off" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditCmdGroup = false">取消</el-button>
        <el-button type="primary" @click="EditCmdGroup">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

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

</style>