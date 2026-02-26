<template>
  配置account
  <el-button type="primary" link @click="ShowAddAccount">添加</el-button>
  <el-button type="primary" link @click="ShowAccountGroup">Account分组</el-button>
  <el-table :data="state.accountList" style="width: 100%">
    <el-table-column prop="id" label="#id" width="80" />
    <el-table-column prop="username" label="username"  />
    <el-table-column prop="password" label="password" />
    <el-table-column prop="account_group_name" label="account分组"  width="180"/>
    <el-table-column label="操作" >
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditAccount(scope.row , true)">复制新增</el-button>
        <el-button type="primary" link @click="ShowEditAccount(scope.row , false)">编辑</el-button>
        <el-button link type="danger" @click="DeleteAccount(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>

  <el-dialog v-model="state.dialogEditAccount" title="编辑" width="500">
    <el-form>
      <el-form-item label="username" :label-width="80">
        <el-input v-model="state.editAccountConfig.username" autocomplete="off" />
      </el-form-item>
      <el-form-item label="password" :label-width="80">
        <el-input v-model="state.editAccountConfig.password" autocomplete="off" />
      </el-form-item>
      <el-form-item label="分组" :label-width="80">
        <el-select v-model="state.editAccountConfig.account_group_id" placeholder="选择分组" style="width: 140px">
          <el-option v-for="item in state.accountGroupList" :key="item.id" :label="item.name" :value="item.id"/>
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditAccount = false">取消</el-button>
        <el-button type="primary" @click="EditAccount">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="state.dialogAccountGroup" title="Account分组" width="1000">
    <account_group v-if="state.dialogAccountGroup" @update-group="UpdateGroup" ref="account_group"></account_group>
  </el-dialog>

</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive} from 'vue';
import set from '../../utils/base/account_set'
import common from '../../utils/common'
import list from "@/utils/base/list";
import account_group from "@/components/set/account_group.vue";
import Init from "@/utils/base/set_init";
export default defineComponent({
  components: {account_group},
  props: {
  },
  data() {
    return {
    }
  },
  setup() {
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const AccountList = function (){
      set.AccountList(function (response){
        if(response.ErrCode === 0){
          state.accountList = response.Data
        }
      })
    }
    const ShowEditAccount = function (accountConfig , isCopy){
      state.dialogEditAccount = true
      state.editAccountConfig = accountConfig
      if(isCopy){
        state.editAccountConfig.id = 0
      }
    }
    const ShowAddAccount = function (){
      state.dialogEditAccount = true
      state.editAccountConfig = {}
    }
    const ShowAccountGroup = function (){
      state.dialogAccountGroup = true
    }

    const EditAccount = function (){
      set.AccountAdd(state.editAccountConfig , function (response){
        if(response.ErrCode === 0){
          AccountList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditAccount = false
        Init.SetIsInit('smart_link')
      })
    }
    const DeleteAccount = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.AccountDelete(rowData , function (response){
          if(response.ErrCode === 0){
            AccountList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
          Init.SetIsInit('smart_link')
        })
      })
    }

    const AccountGroupList = function (){
      set.AccountGroupList(function (response){
        if(response.ErrCode === 0){
          state.accountGroupList = response.Data
        }
      })
    }
    const UpdateGroup = function (){
      AccountGroupList()
      Init.SetIsInit('smart_link')
    }
    //固有属性
    const state = reactive({
      accountGroupList : [],
      accountList : [],
      dialogEditAccount : false,
      editAccountConfig : {},
      filterValue : '',
      dialogAccountGroup: false,
    })
    //初始化
    AccountList()
    AccountGroupList()
    return {
      state,
      ShowEditAccount,
      ShowAddAccount,
      EditAccount,
      DeleteAccount,
      ShowAccountGroup,
      AccountList,
      AccountGroupList,
      UpdateGroup,
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