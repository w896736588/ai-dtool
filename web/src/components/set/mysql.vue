<template>
  配置mysql <el-button type="primary" link @click="ShowAddMysql">添加</el-button>
  <el-table :data="state.mysqlList" style="width: 100%">
    <el-table-column prop="id" label="#id" />
    <el-table-column prop="name" label="name" width="180" />
    <el-table-column prop="ssh_name" label="ssh" width="140"/>
    <el-table-column prop="host" label="Host" width="180" />
    <el-table-column prop="port" label="port" />
    <el-table-column prop="username" label="username" />
    <el-table-column prop="dbname" label="dbname" />
    <el-table-column label="操作" width="200">
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditMysql(scope.row , true)">复制新增</el-button>
        <el-button type="primary" link @click="ShowEditMysql(scope.row , false)">编辑</el-button>
        <el-button link type="danger" @click="DeleteMysql(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>

  <el-dialog v-model="state.dialogEditMysql" title="编辑" width="500">
    <el-form :model="state.starForm">
      <el-form-item label="name" :label-width="80">
        <el-input v-model="state.editMysqlConfig.name" autocomplete="off" />
      </el-form-item>
      <el-form-item label="host" :label-width="80">
        <el-input v-model="state.editMysqlConfig.host" autocomplete="off" />
      </el-form-item>
      <el-form-item label="port" :label-width="80">
        <el-input v-model="state.editMysqlConfig.port" autocomplete="off" />
      </el-form-item>
      <el-form-item label="username" :label-width="80">
        <el-input v-model="state.editMysqlConfig.username" autocomplete="off" />
      </el-form-item>
      <el-form-item label="password" :label-width="80">
        <el-input v-model="state.editMysqlConfig.password" type="password" autocomplete="off" />
      </el-form-item>
      <el-form-item label="dbname" :label-width="80">
        <el-input v-model="state.editMysqlConfig.dbname" autocomplete="off" />
      </el-form-item>
      <el-form-item label="ssh" :label-width="80">
        <el-select v-model="state.editMysqlConfig.ssh_id" placeholder="选择分组" style="width: 140px">
          <el-option v-for="item in state.sshList" :key="item.id" :label="item.name" :value="item.id"/>
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditMysql = false">取消</el-button>
        <el-button type="primary" @click="EditMysql">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

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

</style>