<template>
  <div class="mcp-list-page">
    <div class="mcp-list-header">
      <h3>MCP 管理</h3>
      <el-button type="primary" size="small" @click="openAgentTargetDialog">管理目标智能体</el-button>
    </div>

    <el-table :data="typeList" stripe border style="width: 100%">
      <el-table-column prop="name" label="MCP 类型" width="160" />
      <el-table-column prop="package_name" label="包名" min-width="240" />
      <el-table-column label="已绑定" width="120">
        <template #default="scope">
          <el-tag size="small" type="info">{{ getTotalBindCount(scope.row) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="说明" min-width="200" />
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="scope">
          <el-button type="primary" link size="small" @click="goToBinding(scope.row)">查看详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 目标智能体管理弹窗 -->
    <el-dialog v-model="agentTargetDialogVisible" title="目标智能体管理" width="70%">
      <div style="margin-bottom: 12px;">
        <el-button type="primary" size="small" @click="addAgentTarget">新增</el-button>
      </div>
      <el-table :data="agentTargetList" stripe border style="width: 100%">
        <el-table-column prop="agent_name" label="名称" width="150" />
        <el-table-column prop="config_filename" label="配置文件名" width="160" />
        <el-table-column prop="config_dir" label="配置目录" min-width="300" />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="scope">
            <el-button type="primary" link size="small" @click="editAgentTarget(scope.row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="deleteAgentTarget(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 目标智能体编辑弹窗 -->
    <el-dialog v-model="agentTargetFormVisible" :title="agentTargetFormId ? '编辑智能体' : '新增智能体'" width="500">
      <el-form :model="agentTargetForm" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="agentTargetForm.agent_name" placeholder="如 claude_code" />
        </el-form-item>
        <el-form-item label="配置文件名">
          <el-input v-model="agentTargetForm.config_filename" placeholder="如 .claude.json" />
        </el-form-item>
        <el-form-item label="配置目录">
          <el-input v-model="agentTargetForm.config_dir" placeholder="配置文件所在目录的完整路径" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="agentTargetFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="agentTargetSaving" @click="saveAgentTarget">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import mcpApi from '@/utils/base/mcp'

export default {
  data() {
    return {
      typeList: [],
      agentTargetList: [],
      agentTargetDialogVisible: false,
      agentTargetFormVisible: false,
      agentTargetFormId: 0,
      agentTargetForm: {
        agent_name: '',
        config_filename: '',
        config_dir: '',
      },
      agentTargetSaving: false,
    }
  },
  mounted() {
    this.loadTypeList()
  },
  methods: {
    loadTypeList() {
      mcpApi.McpTypeList((response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.typeList = response.Data || []
        }
      })
    },
    getTotalBindCount(row) {
      let total = 0
      const map = row.bind_count_map || {}
      for (const key in map) {
        total += map[key]
      }
      return total
    },
    goToBinding(row) {
      this.$router.push('/Mcp/' + row.mcp_type)
    },
    openAgentTargetDialog() {
      this.loadAgentTargetList()
      this.agentTargetDialogVisible = true
    },
    loadAgentTargetList() {
      mcpApi.McpAgentTargetList((response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.agentTargetList = response.Data || []
        }
      })
    },
    addAgentTarget() {
      this.agentTargetFormId = 0
      this.agentTargetForm = { agent_name: '', config_filename: '', config_dir: '' }
      this.agentTargetFormVisible = true
    },
    editAgentTarget(row) {
      this.agentTargetFormId = row.id
      this.agentTargetForm = {
        agent_name: row.agent_name,
        config_filename: row.config_filename,
        config_dir: row.config_dir,
      }
      this.agentTargetFormVisible = true
    },
    saveAgentTarget() {
      this.agentTargetSaving = true
      const data = {
        ...this.agentTargetForm,
        id: this.agentTargetFormId,
      }
      mcpApi.McpAgentTargetSave(data, (response) => {
        this.agentTargetSaving = false
        if (response && response.ErrCode === 0) {
          this.agentTargetFormVisible = false
          this.loadAgentTargetList()
          this.loadTypeList()
          this.$helperNotify.success('保存成功')
        } else {
          this.$helperNotify.error(response && response.ErrMsg ? response.ErrMsg : '保存失败')
        }
      })
    },
    deleteAgentTarget(row) {
      this.$confirm('确定删除该目标智能体？', '提示', {
        type: 'warning',
      }).then(() => {
        mcpApi.McpAgentTargetDelete(row.id, (response) => {
          if (response && response.ErrCode === 0) {
            this.loadAgentTargetList()
            this.$helperNotify.success('删除成功')
          } else {
            this.$helperNotify.error(response && response.ErrMsg ? response.ErrMsg : '删除失败')
          }
        })
      }).catch(function() {})
    },
  },
}
</script>

<style scoped>
.mcp-list-page {
  padding: 20px;
}
.mcp-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.mcp-list-header h3 {
  margin: 0;
}
</style>
