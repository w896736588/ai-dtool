<template>
  配置account组 <el-button type="primary" link @click="ShowAddAccountGroup">添加</el-button>
  <el-table :data="state.accountGroupList" style="width: 100%">
    <el-table-column prop="id" label="#id" />
    <el-table-column prop="name" label="name" />
    <el-table-column label="操作" >
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditAccountGroup(scope.row)">编辑</el-button>
        <el-button link type="danger" @click="DeleteAccountGroup(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>
  <p></p>
  <el-dialog v-model="state.dialogEditAccountGroup" title="编辑" width="500">
    <el-form>
      <el-form-item label="组名" :label-width="80">
        <el-input v-model="state.editAccountGroupConfig.name" autocomplete="off" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditAccountGroup = false">取消</el-button>
        <el-button type="primary" @click="EditAccountGroup">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive} from 'vue';
import ssh_set from '../../utils/base/ssh_set'
import set from '../../utils/base/account_set'
import common from '../../utils/common'
import list from "@/utils/base/list";
export default defineComponent({
  props: {
  },
  data() {
    return {
    }
  },
  setup(props , {emit}) {
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties

    const AccountGroupList = function (){
      set.AccountGroupList(function (response){
        if(response.ErrCode === 0){
          state.accountGroupList = response.Data
        }
      })
    }
    const ShowEditAccountGroup = function (accountGroupConfig){
      state.dialogEditAccountGroup = true
      state.editAccountGroupConfig = accountGroupConfig
    }
    const ShowAddAccountGroup = function (){
      state.dialogEditAccountGroup = true
      state.editAccountGroupConfig = {}
    }
    const EditAccountGroup = function (){
      set.AccountGroupAdd(state.editAccountGroupConfig , function (response){
        if(response.ErrCode === 0){
          AccountGroupList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditAccountGroup = false
        emit('update-group')
      })
    }
    const DeleteAccountGroup = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.AccountGroupDelete(rowData , function (response){
          if(response.ErrCode === 0){
            AccountGroupList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
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
    //固有属性
    const state = reactive({
      sshList :[],
      accountGroupList : [],
      dialogEditAccountGroup : false,
      editAccountGroupConfig : {},
    })
    //初始化
    AccountGroupList()
    SshList()
    return {
      state,
      ShowEditAccountGroup,
      ShowAddAccountGroup,
      EditAccountGroup,
      DeleteAccountGroup,
      AccountGroupList,
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