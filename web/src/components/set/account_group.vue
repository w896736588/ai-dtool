<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">Account 分组管理</h3>
      <div class="set-config-actions">
        <pl-button type="primary" @click="ShowAddAccountGroup">添加分组</pl-button>
      </div>
    </div>
    <div class="set-config-table-card">
      <el-table :data="state.accountGroupList" class="set-config-table">
        <el-table-column prop="id" label="#id" width="80" />
        <el-table-column prop="name" label="分组名" min-width="180" />
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <div class="set-op-group">
              <pl-button type="primary" link @click="ShowEditAccountGroup(scope.row)">编辑</pl-button>
              <pl-button link type="danger" @click="DeleteAccountGroup(scope.row)">删除</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog v-model="state.dialogEditAccountGroup" title="编辑分组" width="500">
      <el-form label-width="80px">
        <el-form-item label="组名">
          <el-input v-model="state.editAccountGroupConfig.name" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogEditAccountGroup = false">取消</pl-button>
          <pl-button type="primary" @click="EditAccountGroup">保存</pl-button>
        </div>
      </template>
    </el-dialog>
  </div>
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

<style scoped src="@/css/components/set/account_group.css"></style>

