<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">账号配置管理</h3>
      <p class="set-config-desc">管理账号信息及账号分组</p>
      <div class="set-config-actions">
        <pl-button type="primary" @click="ShowAddAccount">添加账号</pl-button>
        <pl-button @click="ShowAccountGroup">Account分组</pl-button>
        <el-select v-model="state.filterGroupId" clearable placeholder="按分组筛选" style="width: 160px">
          <el-option v-for="item in state.accountGroupList" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
      </div>
    </div>
    <div class="set-config-table-card">
      <el-table :data="filteredAccountList" class="set-config-table" max-height="460">
        <el-table-column prop="id" label="#id" width="80" />
        <el-table-column prop="username" label="用户名" min-width="160" />
        <el-table-column prop="password" label="密码" min-width="140">
          <template #default>
            <span>******</span>
          </template>
        </el-table-column>
        <el-table-column prop="account_group_name" label="Account分组" min-width="180"/>
        <el-table-column label="操作" min-width="180">
          <template #default="scope">
            <div class="set-op-group">
              <pl-button type="primary" link @click="ShowEditAccount(scope.row , true)">复制新增</pl-button>
              <pl-button type="primary" link @click="ShowEditAccount(scope.row , false)">编辑</pl-button>
              <pl-button link type="danger" @click="DeleteAccount(scope.row)">删除</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="state.dialogEditAccount" title="编辑账号" width="500">
      <el-form label-width="90px">
        <el-form-item label="用户名">
          <el-input v-model="state.editAccountConfig.username" autocomplete="off" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="state.editAccountConfig.password" autocomplete="off" />
        </el-form-item>
        <el-form-item label="分组">
          <el-select v-model="state.editAccountConfig.account_group_id" placeholder="选择分组" style="width: 140px">
            <el-option v-for="item in state.accountGroupList" :key="item.id" :label="item.name" :value="item.id"/>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogEditAccount = false">取消</pl-button>
          <pl-button type="primary" @click="EditAccount">保存</pl-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="state.dialogAccountGroup" title="Account分组" width="1000">
      <account_group v-if="state.dialogAccountGroup" @update-group="UpdateGroup" ref="account_group"></account_group>
    </el-dialog>
  </div>
</template>
<script>
import {defineComponent , getCurrentInstance , reactive, computed} from 'vue';
import set from '../../utils/base/account_set'
import common from '../../utils/common'
import list from "@/utils/base/list";
import account_group from "@/components/set/account_group.vue";
import Init from "@/utils/base/set_init";
export default defineComponent({
  components: {account_group},
  props: {
  },
  emits: ['changed'],
  data() {
    return {
    }
  },
  setup(props, { emit }) {
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

    // emitChanged 告知宿主页面账号配置已变化，便于自定义网页页内立即刷新。
    // Notify host pages that account settings changed so the custom web page can refresh immediately.
    const emitChanged = function (){
      emit('changed')
    }
    const EditAccount = function (){
      set.AccountAdd(state.editAccountConfig , function (response){
        if(response.ErrCode === 0){
          AccountList()
          emitChanged()
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
            emitChanged()
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
      emitChanged()
    }
    //固有属性
    const state = reactive({
      accountGroupList : [],
      accountList : [],
      dialogEditAccount : false,
      editAccountConfig : {},
      filterValue : '',
      filterGroupId : '',
      dialogAccountGroup: false,
    })

    const filteredAccountList = computed(() => {
      if (!state.filterGroupId) return state.accountList
      return state.accountList.filter(item => item.account_group_id === state.filterGroupId)
    })
    //初始化
    AccountList()
    AccountGroupList()
    return {
      state,
      filteredAccountList,
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

<style scoped src="@/css/components/set/account.css"></style>

