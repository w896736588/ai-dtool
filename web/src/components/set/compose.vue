<template>
  配置docker compose <el-button type="primary" link @click="ShowAddCompose">添加</el-button>
  <el-table :data="state.composeList" style="width: 100%">
    <el-table-column prop="id" label="#id" width="60"/>
    <el-table-column prop="name" label="名称"  width="150"/>
    <el-table-column prop="compose_yml_path" label="compose.yml目录"  width="200"/>
    <el-table-column prop="env_file" label="env file" width="140" />
    <el-table-column prop="ssh_name" label="ssh" width="140"/>
    <el-table-column prop="docker_cmd" label="命令" width="140"/>
    <el-table-column prop="default_service" label="默认服务" />
<!--    <el-table-column prop="upload_exes" label="上传重启" />-->
    <el-table-column label="操作" width="200">
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditCompose(scope.row , true)">复制新增</el-button>
        <el-button type="primary" link @click="ShowEditCompose(scope.row , false)">编辑</el-button>
        <el-button link type="danger" @click="DeleteCompose(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>

  <el-dialog v-model="state.dialogEditCompose" title="编辑" width="70%">
    <el-form :model="state.starForm">
      <el-form-item label="名称" :label-width="180">
        <el-input v-model="state.editComposeConfig.name" autocomplete="off" />
      </el-form-item>
      <el-form-item label="compose.yml" :label-width="180">
        <el-input v-model="state.editComposeConfig.compose_yml_path" autocomplete="off" />
      </el-form-item>
      <el-form-item label="env file(为空默认为.env)" :label-width="180">
        <el-input v-model="state.editComposeConfig.env_file" autocomplete="off" />
      </el-form-item>
      <el-form-item label="ssh" :label-width="180">
        <el-select v-model="state.editComposeConfig.ssh_id" placeholder="选择分组" style="width: 140px">
          <el-option v-for="item in state.sshList" :key="item.id" :label="item.name" :value="item.id"/>
        </el-select>
      </el-form-item>
      <el-form-item label="docker命令" :label-width="180">
        <el-input v-model="state.editComposeConfig.docker_cmd" autocomplete="off" />
      </el-form-item>
      <el-form-item label="默认服务(多个英文逗号分割)" :label-width="180">
        <el-input v-model="state.editComposeConfig.default_service" autocomplete="off" />
      </el-form-item>
<!--      <el-form-item label="上传重启定义" :label-width="180">-->
<!--        <el-input type="textarea" rows="5" v-model="state.editComposeConfig.upload_exes" autocomplete="off" />-->
<!--      </el-form-item>-->
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditCompose = false">取消</el-button>
        <el-button type="primary" @click="EditCompose">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive} from 'vue';
import set from '../../utils/base/compose_set'
import common from '../../utils/common'
import ssh_set from "@/utils/base/ssh_set";
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
    const ComposeList = function (){
      set.ComposeList(function (response){
        if(response.ErrCode === 0){
          state.composeList = response.Data
        }
      })
    }
    const ShowEditCompose = function (composeConfig , isCopy){
      state.dialogEditCompose = true
      state.editComposeConfig = composeConfig
      if(isCopy){
        state.editComposeConfig.id = 0
      }
    }
    const ShowAddCompose = function (){
      state.dialogEditCompose = true
      state.editComposeConfig = {}
    }
    const EditCompose = function (){
      set.ComposeAdd(state.editComposeConfig , function (response){
        if(response.ErrCode === 0){
          ComposeList()
        }else{
          instance.$helperNotify.success(response.ErrMsg)
        }
        state.dialogEditCompose = false
      })
    }
    const DeleteCompose = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.ComposeDelete(rowData , function (response){
          if(response.ErrCode === 0){
            ComposeList()
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
    //固有属性
    const state = reactive({
      sshList : [],
      composeList : [],
      dialogEditCompose : false,
      editComposeConfig : {},
    })
    //初始化
    ComposeList()
    SshList()

    return {
      state,
      ShowEditCompose,
      ShowAddCompose,
      EditCompose,
      DeleteCompose,
      ComposeList,
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