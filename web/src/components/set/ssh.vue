<template>
  配置ssh <el-button type="primary" link @click="ShowAddSsh">添加</el-button>
  <el-table :data="state.sshList" style="width: 100%">
    <el-table-column prop="id" label="#id" width="50" />
    <el-table-column prop="name" label="name" width="180" />
    <el-table-column prop="host" label="Host" width="180" />
    <el-table-column prop="port" label="port" />
    <el-table-column prop="username" label="username" />
    <el-table-column prop="home" label="家目录" />
    <el-table-column prop="status" label="连接状态" />
    <el-table-column label="当前连接数">
      <template #default="scope">
        <el-button type="primary" link @click="ShowConnections(scope.row)">{{ GetConnectionCount(scope.row.id) }}</el-button>
      </template>
    </el-table-column>
    <el-table-column label="操作" >
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditSsh(scope.row , true)">复制新增</el-button>
        <el-button type="primary" link @click="ShowEditSsh(scope.row , false)">编辑</el-button>
        <el-button link type="danger" @click="DeleteSsh(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>

  <el-dialog v-model="state.dialogEditSsh" title="编辑" width="500">
    <el-form :model="state.starForm">
      <el-form-item label="name" :label-width="80">
        <el-input v-model="state.editSshConfig.name" autocomplete="off" />
      </el-form-item>
      <el-form-item label="host" :label-width="80">
        <el-input v-model="state.editSshConfig.host" autocomplete="off" />
      </el-form-item>
      <el-form-item label="port" :label-width="80">
        <el-input v-model="state.editSshConfig.port" autocomplete="off" />
      </el-form-item>
      <el-form-item label="username" :label-width="80">
        <el-input v-model="state.editSshConfig.username" autocomplete="off" />
      </el-form-item>
      <el-form-item label="password" :label-width="80">
        <el-input v-model="state.editSshConfig.password" type="password"  autocomplete="off" />
      </el-form-item>
      <el-form-item label="家目录" :label-width="80">
        <el-input v-model="state.editSshConfig.home" type="text"  autocomplete="off" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditSsh = false">取消</el-button>
        <el-button type="primary" @click="EditSsh">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="state.dialogConnections" title="连接详情" width="80%">
    <el-table :data="state.connections" style="width: 100%">
      <el-table-column prop="shell_client_id" label="客户端ID" />
      <el-table-column prop="current_command" label="当前命令" />
      <el-table-column prop="status" label="状态" />
      <el-table-column prop="connect_time" label="连接开始时间" width="180" />
      <el-table-column prop="connect_seconds" label="连接时长(秒)" width="120" />
      <el-table-column prop="type" label="类型" />
      <el-table-column label="操作">
        <template #default="scope">
          <el-button type="primary" link @click="Reconnect(scope.row)">重连</el-button>
        </template>
      </el-table-column>
    </el-table>
    <template #footer>
      <div class="dialog-footer">
        <el-button type="primary" @click="state.dialogConnections = false">关闭</el-button>
      </div>
    </template>
  </el-dialog>

</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive , onMounted , onBeforeUnmount} from 'vue';
import set from '../../utils/base/ssh_set'
import common from '../../utils/common'
import Init  from '@/utils/base/set_init'
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
    const SortConnectionsByDuration = function (list){
      return [...(list || [])].sort((a, b) => {
        const aSeconds = Number(a.connect_seconds || 0)
        const bSeconds = Number(b.connect_seconds || 0)
        if(aSeconds === bSeconds){
          return String(a.shell_client_id || '').localeCompare(String(b.shell_client_id || ''))
        }
        return aSeconds - bSeconds
      })
    }
    const SshList = function (){
      set.SshList(function (response){
        if(response.ErrCode === 0){
          // Sort by ID ascending
          state.sshList = response.Data.sort((a, b) => a.id - b.id)
        }
      })
    }
    const LoadConnections = function (){
      set.GetConnections(function (response){
        if(response.ErrCode === 0){
          state.allConnections = SortConnectionsByDuration(response.Data.connections || [])
        }
      })
    }
    const ShowEditSsh = function (sshConfig , isCopy){
      state.dialogEditSsh = true
      state.editSshConfig = sshConfig
      if(isCopy){
        state.editSshConfig.id = 0
      }
    }
    const ShowAddSsh = function (){
      state.dialogEditSsh = true
      state.editSshConfig = {}
    }
    const EditSsh = function (){
      set.SshAdd(state.editSshConfig , function (response){
        if(response.ErrCode === 0){
          SshList()
        }else{
          instance.$helperNotify.success(response.ErrMsg)
        }
        state.dialogEditSsh = false
        SetInit()
      })
    }

    const SetInit = function(){
      Init.SetIsInit('git') //git配置页面
      Init.SetIsInit('supervisor') //supervisor设置页面
      Init.SetIsInit('redis')
      Init.SetIsInit('mysql')
    }

    const DeleteSsh = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.SshDelete(rowData , function (response){
          if(response.ErrCode === 0){
            SshList()
          }else{
            instance.$helperNotify.success(response.ErrMsg)
          }
          SetInit()
        })
      })
    }
    const ShowConnections = function (sshConfig){
      state.dialogConnections = true
      state.selectedSshId = sshConfig.id
      set.GetConnections(function (response){
        if(response.ErrCode === 0){
          state.allConnections = SortConnectionsByDuration(response.Data.connections || [])
          // Filter connections for the selected SSH
          state.connections = SortConnectionsByDuration(state.allConnections.filter(conn => {
            const sshId = conn.shell_client_id.split('#')[0]
            return sshId === String(sshConfig.id)
          }))
        }else{
          instance.$helperNotify.success(response.ErrMsg)
        }
      })
    }
    const GetConnectionCount = function (sshId){
      if(!state.allConnections || state.allConnections.length === 0){
        return 0
      }
      return state.allConnections.filter(conn => {
        const connSshId = conn.shell_client_id.split('#')[0]
        return connSshId === String(sshId)
      }).length
    }
    const Reconnect = function (connection){
      set.ReconnectConnection(connection.shell_client_id, function (response){
        if(response.ErrCode === 0){
          instance.$helperNotify.success('重连成功')
          // Refresh connections
          if(state.selectedSshId){
            ShowConnections({id: state.selectedSshId})
          }else{
            LoadConnections()
          }
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }
    const RefreshAll = function (){
      SshList()
      LoadConnections()
      // Also refresh dialog connections if dialog is open
      if(state.dialogConnections && state.selectedSshId){
        set.GetConnections(function (response){
          if(response.ErrCode === 0){
            state.allConnections = SortConnectionsByDuration(response.Data.connections || [])
            state.connections = SortConnectionsByDuration(state.allConnections.filter(conn => {
              const sshId = conn.shell_client_id.split('#')[0]
              return sshId === String(state.selectedSshId)
            }))
          }
        })
      }
    }
    let timer = null
    //固有属性
    const state = reactive({
      sshList : [],
      dialogEditSsh : false,
      editSshConfig : {},
      dialogConnections : false,
      connections : [],
      allConnections : [],
      selectedSshId : null,
    })
    //初始化
    SshList()
    LoadConnections()
    onMounted(() => {
      timer = setInterval(() => {
        RefreshAll()
      }, 3000)
    })
    onBeforeUnmount(() => {
      if(timer){
        clearInterval(timer)
        timer = null
      }
    })
    return {
      state,
      ShowEditSsh,
      ShowAddSsh,
      EditSsh,
      DeleteSsh,
      SshList,
      ShowConnections,
      GetConnectionCount,
      Reconnect,
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
