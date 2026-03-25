<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">MySQL 配置管理</h3>
      <p class="set-config-desc">管理数据库连接配置与 SSH 隧道映射</p>
      <div class="set-config-actions">
        <pl-button type="primary" @click="ShowAddMysql">添加 MySQL</pl-button>
      </div>
    </div>
    <div class="set-config-table-card">
      <el-table :data="state.mysqlList" class="set-config-table">
        <el-table-column prop="id" label="#id" width="80" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column prop="ssh_name" label="SSH" width="140" />
        <el-table-column prop="host" label="Host" min-width="180" />
        <el-table-column prop="port" label="Port" width="90" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="dbname" label="数据库" min-width="140" />
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <div class="set-op-group">
              <pl-button type="primary" link @click="ShowEditMysql(scope.row , true)">复制新增</pl-button>
              <pl-button type="primary" link @click="ShowEditMysql(scope.row , false)">编辑</pl-button>
              <pl-button link type="danger" @click="DeleteMysql(scope.row)">删除</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="state.dialogEditMysql" title="编辑MySQL配置" width="520">
      <el-form :model="state.starForm" label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="state.editMysqlConfig.name" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Host">
          <el-input v-model="state.editMysqlConfig.host" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Port">
          <el-input v-model="state.editMysqlConfig.port" autocomplete="off" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="state.editMysqlConfig.username" autocomplete="off" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="state.editMysqlConfig.password" type="password" autocomplete="off" />
        </el-form-item>
        <el-form-item label="数据库">
          <el-input v-model="state.editMysqlConfig.dbname" autocomplete="off" />
        </el-form-item>
        <el-form-item label="SSH">
          <el-select v-model="state.editMysqlConfig.ssh_id" placeholder="选择SSH" style="width: 140px">
            <el-option v-for="item in state.sshList" :key="item.id" :label="item.name" :value="item.id"/>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogEditMysql = false">取消</pl-button>
          <pl-button type="primary" @click="EditMysql">保存</pl-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
<script>
import {defineExpose, defineComponent, inject, defineEmits, getCurrentInstance, reactive, onActivated} from 'vue';
import set from '../../utils/base/mysql_set'
import common from '../../utils/common'
import ssh_set from "@/utils/base/ssh_set";
import Init from "@/utils/base/set_init";
export default defineComponent({
  props: {
  },
  data() {
    return {
    }
  },
  setup() {
    onActivated(() => {
      if(Init.GetIsInit('mysql') === true){
        MysqlList()
        SshList()
        Init.DelInit('mysql')
      }
    });
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const MysqlList = function (){
      set.MysqlList(function (response){
        if(response.ErrCode === 0){
          state.mysqlList = response.Data
        }
      })
    }
    const ShowEditMysql = function (mysqlConfig , isCopy){
      state.dialogEditMysql = true
      state.editMysqlConfig = mysqlConfig
      if(isCopy){
        state.editMysqlConfig.id = 0
      }
    }
    const ShowAddMysql = function (){
      state.dialogEditMysql = true
      state.editMysqlConfig = {}
    }
    const EditMysql = function (){
      set.MysqlAdd(state.editMysqlConfig , function (response){
        if(response.ErrCode === 0){
          MysqlList()
        }else{
          instance.$helperNotify.success(response.ErrMsg)
        }
        state.dialogEditMysql = false
      })
    }
    const DeleteMysql = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.MysqlDelete(rowData , function (response){
          if(response.ErrCode === 0){
            MysqlList()
          }else{
            instance.$helperNotify.success(response.ErrMsg)
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
    SshList()
    //固有属性
    const state = reactive({
      sshList : [],
      mysqlList : [],
      dialogEditMysql : false,
      editMysqlConfig : {},
    })
    //初始化
    MysqlList()
    return {
      state,
      ShowEditMysql,
      ShowAddMysql,
      EditMysql,
      DeleteMysql,
      MysqlList,
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

