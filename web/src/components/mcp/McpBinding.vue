<template>
  <div class="mcp-binding-page">
    <div class="mcp-binding-header">
      <el-button v-if="!isEmbedded" type="info" link @click="goBack">&larr; 返回 MCP 列表</el-button>
      <h3>{{ pageTitle }}</h3>
    </div>

    <!-- Chrome DevTools 配置管理模式 -->
    <template v-if="isChromeDevtools">
      <div style="margin-bottom: 12px;">
        <el-button type="primary" size="small" @click="openConfigDialog">新增配置</el-button>
      </div>

      <el-table v-loading="loading" :data="configList" stripe border style="width: 100%">
        <el-table-column prop="name" label="名称" width="200" />
        <el-table-column prop="port" label="端口" width="100" />
        <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip />
        <el-table-column label="使用状态" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.is_used" type="success" size="small">使用中</el-tag>
            <el-tag v-else type="info" size="small">空闲</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="scope">
            <el-button type="primary" link size="small" @click="editConfig(scope.row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="deleteConfig(scope.row)">删除</el-button>
            <el-button type="info" link size="small" @click="viewConfig(scope.row)">查看配置</el-button>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- 其他 MCP 类型绑定管理模式（保持原有逻辑） -->
    <template v-else>
      <el-tabs v-model="activeTargetId" @tab-change="onTargetChange">
        <el-tab-pane
          v-for="target in agentTargetList"
          :key="target.id"
          :label="target.agent_name"
          :name="String(target.id)"
        />
      </el-tabs>

      <el-table v-loading="loading" :data="bindingList" stripe border style="width: 100%">
        <el-table-column prop="mapping_key" label="Mapping Key" width="200" />
        <el-table-column prop="label" label="标签" width="150" />
        <el-table-column prop="user_data_dir" label="用户数据目录" min-width="300" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.is_bound" size="small" type="success">已绑定</el-tag>
            <el-tag v-else size="small" type="info">未绑定</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <el-button
              v-if="!scope.row.is_bound"
              type="primary"
              link
              size="small"
              @click="addBinding(scope.row)"
            >添加绑定</el-button>
            <el-button
              v-if="scope.row.is_bound"
              type="danger"
              link
              size="small"
              @click="removeBinding(scope.row)"
            >移除绑定</el-button>
            <el-button
              v-if="scope.row.is_bound"
              type="info"
              link
              size="small"
              @click="copyInstruction(scope.row)"
            >复制说明</el-button>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- 配置编辑弹窗 -->
    <el-dialog v-model="configDialogVisible" :title="configFormId ? '编辑配置' : '新增配置'" width="500px">
      <el-form :model="configForm" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="configForm.name" placeholder="如 chrome-devtools-1" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="configForm.port" :min="1" :max="65535" placeholder="如 9222" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="configForm.remark" placeholder="备注说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="configSaving" @click="saveConfig">保存</el-button>
      </template>
    </el-dialog>

    <!-- 查看配置 JSON 弹窗 -->
    <el-dialog v-model="configJsonVisible" title="MCP 配置" width="600px">
      <el-input
        type="textarea"
        :rows="10"
        :model-value="configJsonContent"
        readonly
      />
      <template #footer>
        <el-button type="primary" @click="copyConfigJson">复制</el-button>
        <el-button @click="configJsonVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import mcpApi from '@/utils/base/mcp'
import copy from '@/utils/base/copy'
import sseDistribute from '@/utils/base/sse_distribute'

const CHROME_DEVTOOLS_TYPE = 'chrome-devtools'

// MCP 类型定义（与后端保持一致）
const MCP_TYPE_DEFS = {
  'chrome-devtools': {
    Name: 'ChromeDevTools',
    PackageName: 'chrome-devtools-mcp@latest',
  },
}

export default {
  data() {
    return {
      mcpType: '',
      mcpTypeDef: null,
      // 绑定管理模式数据
      agentTargetList: [],
      activeTargetId: '',
      bindingList: [],
      loading: false,
      // Chrome DevTools 配置管理数据
      configList: [],
      configDialogVisible: false,
      configFormId: 0,
      configForm: { name: '', port: 9222, remark: '' },
      configSaving: false,
      configJsonVisible: false,
      configJsonContent: '',
    }
  },
  computed: {
    isEmbedded() {
      return String(this.$route.query.embed || '') === '1'
    },
    isChromeDevtools() {
      return this.mcpType === CHROME_DEVTOOLS_TYPE
    },
    pageTitle() {
      if (this.isChromeDevtools) {
        return 'ChromeDevTools 调试端口配置'
      }
      return this.mcpTypeDef ? this.mcpTypeDef.Name + ' 绑定管理' : this.mcpType + ' 绑定管理'
    },
  },
  mounted() {
    this.mcpType = this.$route.params.mcpType || ''
    this.mcpTypeDef = MCP_TYPE_DEFS[this.mcpType] || null
    if (this.isChromeDevtools) {
      this.loadConfigList()
      // 注册 SSE 端口状态变更回调，端口分配/释放时自动刷新列表
      this._ssePortStatusId = 'chrome_devtools_port_status'
      sseDistribute.RegisterReceive(this._ssePortStatusId, () => {
        this.loadConfigList()
      })
    } else {
      this.loadAgentTargets()
    }
  },
  beforeDestroy() {
    if (this._ssePortStatusId) {
      sseDistribute.UnRegisterReceive(this._ssePortStatusId)
    }
  },
  methods: {
    goBack() {
      this.$router.push('/Mcp')
    },
    // ========== Chrome DevTools 配置管理方法 ==========
    loadConfigList() {
      this.loading = true
      mcpApi.McpChromeDevtoolsConfigList((response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data) {
          this.configList = response.Data || []
        }
      })
    },
    openConfigDialog() {
      this.configFormId = 0
      this.configForm = { name: '', port: 9222, remark: '' }
      this.configDialogVisible = true
    },
    editConfig(row) {
      this.configFormId = row.id
      this.configForm = { name: row.name, port: row.port, remark: row.remark }
      this.configDialogVisible = true
    },
    saveConfig() {
      if (!this.configForm.name || !this.configForm.port) {
        this.$helperNotify.error('名称和端口不能为空')
        return
      }
      this.configSaving = true
      const data = {
        ...this.configForm,
        id: this.configFormId,
      }
      mcpApi.McpChromeDevtoolsConfigSave(data, (response) => {
        this.configSaving = false
        if (response && response.ErrCode === 0) {
          this.configDialogVisible = false
          this.loadConfigList()
          this.$helperNotify.success('保存成功')
        } else {
          this.$helperNotify.error(response && response.ErrMsg ? response.ErrMsg : '保存失败')
        }
      })
    },
    deleteConfig(row) {
      this.$confirm('确定删除该配置？', '提示', { type: 'warning' }).then(() => {
        mcpApi.McpChromeDevtoolsConfigDelete(row.id, (response) => {
          if (response && response.ErrCode === 0) {
            this.loadConfigList()
            this.$helperNotify.success('删除成功')
          } else {
            this.$helperNotify.error(response && response.ErrMsg ? response.ErrMsg : '删除失败')
          }
        })
      }).catch(function() {})
    },
    viewConfig(row) {
      const cfg = {
        mcpServers: {
          [row.name]: {
            command: 'npx',
            args: ['chrome-devtools-mcp@latest', '--browser-url=http://127.0.0.1:' + row.port],
          },
        },
      }
      this.configJsonContent = JSON.stringify(cfg, null, 2)
      this.configJsonVisible = true
    },
    copyConfigJson() {
      const index = copy.SetCopyContent(this.configJsonContent)
      copy.handleCopy(index)
      this.$helperNotify.success('已复制到剪贴板')
    },
    // ========== 绑定管理方法（保持原有逻辑） ==========
    loadAgentTargets() {
      mcpApi.McpAgentTargetList((response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.agentTargetList = response.Data || []
          if (this.agentTargetList.length > 0 && !this.activeTargetId) {
            this.activeTargetId = String(this.agentTargetList[0].id)
            this.loadBindingList()
          }
        }
      })
    },
    onTargetChange() {
      this.loadBindingList()
    },
    loadBindingList() {
      if (!this.mcpType || !this.activeTargetId) return
      this.loading = true
      mcpApi.McpBindingList(this.mcpType, Number(this.activeTargetId), (response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data) {
          this.bindingList = response.Data || []
        }
      })
    },
    addBinding(row) {
      mcpApi.McpBindingAdd(this.mcpType, row.mapping_id, Number(this.activeTargetId), (response) => {
        if (response && response.ErrCode === 0) {
          this.$helperNotify.success('绑定成功，配置文件已同步')
          this.loadBindingList()
        } else {
          this.$helperNotify.error(response && response.ErrMsg ? response.ErrMsg : '绑定失败')
        }
      })
    },
    removeBinding(row) {
      this.$confirm('确定移除该绑定？配置文件将同步更新。', '提示', { type: 'warning' }).then(() => {
        mcpApi.McpBindingRemove(row.binding_id, (response) => {
          if (response && response.ErrCode === 0) {
            this.$helperNotify.success('绑定已移除，配置文件已同步')
            this.loadBindingList()
          } else {
            this.$helperNotify.error(response && response.ErrMsg ? response.ErrMsg : '移除失败')
          }
        })
      }).catch(function() {})
    },
    copyInstruction(row) {
      if (row.instruction) {
        const index = copy.SetCopyContent(row.instruction)
        copy.handleCopy(index)
        this.$helperNotify.success('已复制到剪贴板')
      }
    },
  },
}
</script>

<style scoped>
.mcp-binding-page {
  padding: 20px;
}
.mcp-binding-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.mcp-binding-header h3 {
  margin: 0;
}
</style>
