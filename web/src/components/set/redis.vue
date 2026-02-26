<template>
  配置redis <el-button type="primary" link @click="ShowAddRedis">添加</el-button>
  <el-table :data="state.redisList" style="width: 100%">
    <el-table-column prop="id" label="#id" />
    <el-table-column prop="name" label="name" width="180" />
    <el-table-column prop="ssh_name" label="ssh" width="140"/>
    <el-table-column prop="host" label="Host" width="180" />
    <el-table-column prop="port" label="port" />
    <el-table-column prop="status" label="连接状态" />
    <el-table-column prop="username" label="username" width="200"/>
    <el-table-column label="操作" width="200">
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditRedis(scope.row , true)">复制新增</el-button>
        <el-button type="primary" link @click="ShowEditRedis(scope.row , false)">编辑</el-button>
        <el-button link type="danger" @click="DeleteRedis(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>

  <el-dialog v-model="state.dialogEditRedis" title="编辑" width="500">
    <el-form :model="state.starForm">
      <el-form-item label="name" :label-width="80">
        <el-input v-model="state.editRedisConfig.name" autocomplete="off" />
      </el-form-item>
      <el-form-item label="host" :label-width="80">
        <el-input v-model="state.editRedisConfig.host" autocomplete="off" />
      </el-form-item>
      <el-form-item label="port" :label-width="80">
        <el-input v-model="state.editRedisConfig.port" autocomplete="off" />
      </el-form-item>
      <el-form-item label="username" :label-width="80">
        <el-input v-model="state.editRedisConfig.username" autocomplete="off" />
      </el-form-item>
      <el-form-item label="password" :label-width="80">
        <el-input v-model="state.editRedisConfig.password" type="password" autocomplete="off" />
      </el-form-item>
      <el-form-item label="ssh" :label-width="80">
        <el-select v-model="state.editRedisConfig.ssh_id" placeholder="选择分组" style="width: 140px">
          <el-option v-for="item in state.sshList" :key="item.id" :label="item.name" :value="item.id"/>
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditRedis = false">取消</el-button>
        <el-button type="primary" @click="EditRedis">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

</template>
<script>
import {defineExpose, defineComponent, inject, defineEmits, getCurrentInstance, reactive, onActivated} from 'vue';
import set from '../../utils/base/redis_set'
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
      if(Init.GetIsInit('redis') === true){
        RedisList()
        SshList()
        Init.DelInit('redis')
      }
    });
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const RedisList = function (){
      set.RedisList(function (response){
        if(response.ErrCode === 0){
          state.redisList = response.Data
        }
      })
    }
    const ShowEditRedis = function (redisConfig , isCopy){
      state.dialogEditRedis = true
      state.editRedisConfig = redisConfig
      if(isCopy){
        state.editRedisConfig.id = 0
      }
    }
    const ShowAddRedis = function (){
      state.dialogEditRedis = true
      state.editRedisConfig = {}
    }
    const EditRedis = function (){
      set.RedisAdd(state.editRedisConfig , function (response){
        if(response.ErrCode === 0){
          RedisList()
        }else{
          instance.$helperNotify.success(response.ErrMsg)
        }
        state.dialogEditRedis = false
        SetInit()
      })
    }
    const DeleteRedis = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.RedisDelete(rowData , function (response){
          if(response.ErrCode === 0){
            RedisList()
          }else{
            instance.$helperNotify.success(response.ErrMsg)
          }
          SetInit()
        })
      })
    }
    const SshList = function (){
      ssh_set.SshList(function (response){
        if(response.ErrCode === 0){
          state.sshList = response.Data
          state.sshList.unshift({id : "0" , name : '请选择'})
        }
      })
    }
    const SetInit = function(){
      Init.SetIsInit('redis')
    }
    SshList()
    //固有属性
    const state = reactive({
      sshList : [],
      redisList : [],
      dialogEditRedis : false,
      editRedisConfig : {},
    })
    //初始化
    RedisList()
    return {
      state,
      ShowEditRedis,
      ShowAddRedis,
      EditRedis,
      DeleteRedis,
      RedisList,
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