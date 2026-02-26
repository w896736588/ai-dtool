<template>
  配置supervisor
  <el-button type="primary" link @click="ShowAddSupervisor">添加</el-button>
  <el-table :data="state.supervisorList" style="width: 100%">
    <el-table-column prop="id" label="#id" width="80" />
    <el-table-column prop="name" label="name"  width="120"/>
    <el-table-column prop="ssh_name" label="ssh" width="140"/>
    <el-table-column prop="docker_name" label="docker name" width="140"/>
    <el-table-column prop="config_dir" label="配置目录" />
    <el-table-column label="操作" width="200">
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditSupervisor(scope.row , true)">复制新增</el-button>
        <el-button type="primary" link @click="ShowEditSupervisor(scope.row , false)">编辑</el-button>
        <el-button link type="danger" @click="DeleteSupervisor(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>

  <el-dialog v-model="state.dialogEditSupervisor" title="编辑" width="500" :close-on-click-modal="false">
    <el-form>
      <el-form-item label="name" :label-width="80">
        <el-input v-model="state.editSupervisorConfig.name" autocomplete="off" />
      </el-form-item>
      <el-form-item label="docker_name" :label-width="80">
        <el-input v-model="state.editSupervisorConfig.docker_name" autocomplete="off" />
      </el-form-item>
      <el-form-item label="目录" :label-width="80">
        <el-input v-model="state.editSupervisorConfig.config_dir" autocomplete="off" />
      </el-form-item>
      <el-form-item label="ssh" :label-width="80">
        <el-select v-model="state.editSupervisorConfig.ssh_id" placeholder="选择分组" style="width: 140px">
          <el-option v-for="item in state.sshList" :key="item.id" :label="item.name" :value="item.id"/>
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditSupervisor = false">取消</el-button>
        <el-button type="primary" @click="EditSupervisor">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

</template>
<script>
import {defineExpose, defineComponent, inject, defineEmits, getCurrentInstance, reactive, onActivated} from 'vue';
import ssh_set from '../../utils/base/ssh_set'
import set from '../../utils/base/supervisor_set'
import common from '../../utils/common'
import list from "@/utils/base/list";
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
      if(Init.GetIsInit('supervisor') === true){
        SupervisorList()
        SshList()
        Init.DelInit('supervisor')
      }
    });
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const SupervisorList = function (){
      set.SupervisorList(function (response){
        if(response.ErrCode === 0){
          state.supervisorList = response.Data
        }
      })
    }
    const ShowEditSupervisor = function (supervisorConfig , isCopy){
      state.dialogEditSupervisor = true
      state.editSupervisorConfig = supervisorConfig
      if(isCopy){
        state.editSupervisorConfig.id = 0
      }
    }
    const SetInit = function(){
      Init.SetIsInit('supervisor')
    }
    const ShowAddSupervisor = function (){
      state.dialogEditSupervisor = true
      state.editSupervisorConfig = {}
    }
    const EditSupervisor = function (){
      set.SupervisorAdd(state.editSupervisorConfig , function (response){
        if(response.ErrCode === 0){
          SupervisorList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditSupervisor = false
        SetInit()
      })
    }
    const DeleteSupervisor = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.SupervisorDelete(rowData , function (response){
          if(response.ErrCode === 0){
            SupervisorList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
          SetInit()
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
      supervisorList : [],
      dialogEditSupervisor : false,
      editSupervisorConfig : {},
      filterValue : '',
    })
    //初始化
    SupervisorList()
    SshList()
    return {
      state,
      ShowEditSupervisor,
      ShowAddSupervisor,
      EditSupervisor,
      DeleteSupervisor,
      SupervisorList,
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