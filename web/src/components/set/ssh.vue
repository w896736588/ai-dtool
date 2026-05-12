<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">SSH 配置管理</h3>
      <p class="set-config-desc">管理远程连接与当前连接状态</p>
      <div class="set-config-actions">
        <pl-button type="primary" @click="ShowAddSsh">添加 SSH</pl-button>
      </div>
    </div>
    <div class="set-config-table-card">
      <el-table :data="state.sshList" class="set-config-table">
        <el-table-column prop="id" label="#id" width="60" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column prop="host" label="Host" min-width="180" />
        <el-table-column prop="port" label="Port" width="90" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="home" label="家目录" min-width="180">
          <template #default="scope">
            <code class="set-mono">{{ scope.row.home || "-" }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="连接状态" width="100" />
        <el-table-column label="当前连接数" width="120" align="center">
          <template #default="scope">
            <pl-button type="primary" link @click="ShowConnections(scope.row)">{{ GetConnectionCount(scope.row.id) }}</pl-button>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="180">
          <template #default="scope">
            <div class="set-op-group">
              <pl-button type="primary" link @click="ShowEditSsh(scope.row , true)">复制新增</pl-button>
              <pl-button type="primary" link @click="ShowEditSsh(scope.row , false)">编辑</pl-button>
              <pl-button link type="danger" @click="DeleteSsh(scope.row)">删除</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="state.dialogEditSsh" title="编辑SSH配置" width="520">
      <el-form :model="state.starForm" label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="state.editSshConfig.name" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Host">
          <el-input v-model="state.editSshConfig.host" autocomplete="off" />
        </el-form-item>
        <el-form-item label="Port">
          <el-input v-model="state.editSshConfig.port" autocomplete="off" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="state.editSshConfig.username" autocomplete="off" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="state.editSshConfig.password" type="password" autocomplete="off" />
        </el-form-item>
        <el-form-item label="家目录">
          <el-input v-model="state.editSshConfig.home" type="text" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogEditSsh = false">取消</pl-button>
          <pl-button type="primary" @click="EditSsh">保存</pl-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="state.dialogConnections" title="连接详情" width="80%">
      <el-table :data="state.connections" class="set-config-table">
        <el-table-column prop="shell_client_id" label="客户端ID" min-width="180" />
        <el-table-column prop="current_command" label="当前命令" min-width="220" />
        <el-table-column prop="status" label="状态" width="90" />
        <el-table-column prop="connect_time" label="连接开始时间" width="180" />
        <el-table-column prop="connect_seconds" label="连接时长(秒)" width="120" />
        <el-table-column prop="type" label="类型" width="90" />
        <el-table-column label="操作" width="90">
          <template #default="scope">
            <pl-button type="primary" link @click="Reconnect(scope.row)">重连</pl-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <div class="dialog-footer">
          <pl-button type="primary" @click="state.dialogConnections = false">关闭</pl-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive , onMounted , onBeforeUnmount} from 'vue';
import set from '../../utils/base/ssh_set'
import common from '../../utils/common'
import Init  from '@/utils/base/set_init'
import sseDistribute from '@/utils/base/sse_distribute'
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
      }, {is_check_connection: 1})
    }
    // 处理SSE推送的连接状态更新
    const handleConnectionsUpdate = function (data){
      if(!data || typeof data !== 'object'){
        state.allConnections = []
        return
      }
      state.allConnections = SortConnectionsByDuration(data.connections || [])
      // 如果对话框打开，更新过滤后的连接列表
      if(state.dialogConnections && state.selectedSshId){
        state.connections = SortConnectionsByDuration(state.allConnections.filter(conn => {
          const sshId = String(conn.shell_client_id || '').split('#')[0]
          return sshId === String(state.selectedSshId)
        }))
      }
    }
    const LoadConnections = function (){
      // 连接数据现在通过SSE自动推送，无需主动获取
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
          state.dialogEditSsh = false
          SshList()
          SetInit()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
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
      // 使用已有的 allConnections 数据进行过滤
      state.connections = SortConnectionsByDuration(state.allConnections.filter(conn => {
        const sshId = String(conn.shell_client_id || '').split('#')[0]
        return sshId === String(sshConfig.id)
      }))
    }
    const GetConnectionCount = function (sshId){
      if(!state.allConnections || state.allConnections.length === 0){
        return 0
      }
      return state.allConnections.filter(conn => {
        const connSshId = String(conn.shell_client_id || '').split('#')[0]
        return connSshId === String(sshId)
      }).length
    }
    const Reconnect = function (connection){
      set.ReconnectConnection(connection.shell_client_id, function (response){
        if(response.ErrCode === 0){
          instance.$helperNotify.success('重连成功')
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }
    const RefreshAll = function (){
      SshList()
      // 连接数据现在通过SSE自动推送
    }
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
    // 注册Shell连接状态SSE监听
    sseDistribute.RegisterReceive('shell_connections', handleConnectionsUpdate)
    onBeforeUnmount(() => {
      sseDistribute.UnRegisterReceive('shell_connections')
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

<style scoped src="@/css/components/set/ssh.css"></style>

