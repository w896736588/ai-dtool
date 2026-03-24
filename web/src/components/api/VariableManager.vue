<template>
  <div class="variable-manager">
    <div class="manager-header">
      <pl-button type="primary" @click="addVariable">
        <el-icon><Plus /></el-icon>
        添加变量
      </pl-button>
      <pl-button @click="importVariables">
        <el-icon><Upload /></el-icon>
        导入
      </pl-button>
      <pl-button @click="exportVariables">
        <el-icon><Download /></el-icon>
        导出
      </pl-button>
    </div>

    <el-table
        :data="variableList"
        style="width: 100%"
        v-loading="loading"
        empty-text="暂无变量数据"
    >
      <el-table-column prop="key" label="变量名" width="200">
        <template #default="{ row }">
          <el-input
              v-if="row.editing"
              v-model="row.key"
              placeholder="变量名"
              size="small"
              class="edit-input"
          />
          <span v-else class="variable-key">{{ '$' + row.key + '$' }}</span>
        </template>
      </el-table-column>

      <el-table-column prop="value" label="变量值">
        <template #default="{ row }">
          <el-input
              v-if="row.editing"
              v-model="row.value"
              placeholder="变量值"
              size="small"
              class="edit-input"
          />
          <el-tooltip
              v-else
              :content="row.value"
              placement="top"
              effect="light"
              :hide-after="0"
          >
            <span class="variable-value truncated">{{ row.value }}</span>
            <pl-button @click="copyVal(row.value)" link >Copy</pl-button>
          </el-tooltip>
        </template>
      </el-table-column>

      <el-table-column prop="desc" label="描述">
        <template #default="{ row }">
          <el-input
              v-if="row.editing"
              v-model="row.desc"
              placeholder="描述信息"
              size="small"
              class="edit-input"
          />
          <el-tooltip
              v-else
              :content="row.desc || ''"
              placement="top"
              effect="light"
              :hide-after="0"
          >
            <span class="variable-desc truncated">{{ row.desc || '' }}</span>
          </el-tooltip>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="150" align="center">
        <template #default="{ row, $index }">
          <div v-if="row.editing">
            <pl-button type="primary" link @click="saveVariable(row)">保存</pl-button>
            <pl-button link @click="cancelEdit(row, $index)">取消</pl-button>
          </div>
          <div v-else>
            <pl-button type="primary" link @click="editVariable(row)">编辑</pl-button>
            <pl-button type="danger" link @click="deleteVariable(row, $index)">删除</pl-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- 批量导入对话框 -->
    <el-dialog
        v-model="importDialogVisible"
        title="导入变量"
        width="500px"
    >
      <el-alert
          title="导入格式：每行一个变量，格式为 key=value 或 key=value#描述"
          type="info"
          :closable="false"
          style="margin-bottom: 16px;"
      />
      <el-input
          v-model="importText"
          type="textarea"
          :rows="10"
          placeholder="例如：&#10;baseUrl=https://api.example.com&#10;apiKey=123456#API密钥&#10;timeout=5000#超时时间"
          class="import-textarea"
      />
      <template #footer>
        <pl-button @click="importDialogVisible = false">取消</pl-button>
        <pl-button type="primary" @click="handleImport">导入</pl-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { Plus, Upload, Download } from '@element-plus/icons-vue'
import typ from '@/utils/base/type'
import Api from "@/utils/base/api";
import Copy from "@/utils/base/copy"

export default {
  name: 'VariableManager',
  components: {
    Plus,
    Upload,
    Download
  },
  props: {
    // env: {
    //   type: Object,
    //   default: () => ({})
    // }
  },
  data() {
    return {
      loading: false,
      variableList: [],
      importDialogVisible: false,
      importText: '',
      backupData: null,
      nextId: 1,
      env : {},
    }
  },
  watch: {
  },
  methods: {
    copyVal : function (val){
      let index = Copy.SetCopyContent(val)
      Copy.handleCopy(index)
    },
    loadVariables(env) {
      let _that = this
      if (!typ.IsArray(env.variables)) {
        _that.variableList = []
      }else{
        _that.variableList = env.variables
      }
      _that.env = env
    },
    addVariable() {
      let _that = this
      _that.variableList.push({
        id: 0,
        key : '',
        collection_id : _that.env.collection_id,
        env_id : _that.env.id,
        value: '',
        desc: '',
        editing: true
      })
    },

    editVariable(variable) {
      this.backupData = { ...variable }
      variable.editing = true
    },

    saveVariable(variable) {
      let _that = this
      Api.CreateCollectionEnvItem({
        collection_id: _that.env.collection_id,
        env_id : _that.env.id,
        key : variable.key,
        value : variable.value,
        desc : variable.desc,
        id : variable.id,
      } , function (res){
        variable.editing = false
        if(res.ErrCode !== 0){
          _that.$message.error(res.ErrMsg)
          return
        }
        for (let i in _that.variableList) {
          if (parseInt(res.Data.id) === parseInt(_that.variableList[i].id) || (parseInt(variable.id) === 0 && parseInt(_that.variableList[i].id) === 0 ) ) {
            _that.variableList[i] = res.Data
            return
          }
        }
        _that.variableList.push(res.Data)
      })
      this.$message.success('保存成功')
    },

    cancelEdit(variable, index) {
      if (this.backupData) {
        Object.assign(variable, this.backupData)
        variable.editing = false
      } else {
        this.variableList.splice(index, 1)
      }
      this.backupData = null
    },

    deleteVariable(variable, index) {
      this.$confirm(`确定要删除变量 "${variable.key}" 吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.variableList.splice(index, 1)
        this.emitUpdate()
        this.$message.success('删除成功')
      })
    },

    emitUpdate() {
      const result = {}
      this.variableList.forEach(item => {
        if (item.key.trim()) {
          // 根据类型转换值
          let value = item.value
          if (item.type === 'number') {
            value = Number(item.value)
          } else if (item.type === 'boolean') {
            value = item.value === 'true'
          }
          result[item.key.trim()] = value
        }
      })
      this.$emit('update', result)
    },

    importVariables() {
      this.importText = ''
      this.importDialogVisible = true
    },

    handleImport() {
      const lines = this.importText.split('\n').filter(line => line.trim())
      const importedCount = 0

      lines.forEach(line => {
        const [keyValue, ...descriptionParts] = line.split('#')
        const [key, ...valueParts] = keyValue.split('=')

        if (key && key.trim() && valueParts.length > 0) {
          const newVariable = {
            id: this.nextId++,
            key: key.trim(),
            value: valueParts.join('=').trim(),
            description: descriptionParts.join('#') || '',
            editing: false
          }

          // 检查是否已存在，如果存在则更新，否则添加
          const existingIndex = this.variableList.findIndex(item => item.key === newVariable.key)
          if (existingIndex >= 0) {
            this.variableList[existingIndex] = newVariable
          } else {
            this.variableList.push(newVariable)
          }
        }
      })

      this.emitUpdate()
      this.importDialogVisible = false
      this.$message.success(`成功导入 ${lines.length} 个变量`)
    },

    exportVariables() {
      const exportText = this.variableList
          .map(item => {
            let line = `${item.key}=${item.value}`
            if (item.description) {
              line += `#${item.description}`
            }
            return line
          })
          .join('\n')

      this.copyToClipboard(exportText)
      this.$message.success('变量已复制到剪贴板')
    },

    copyToClipboard(text) {
      const textArea = document.createElement('textarea')
      textArea.value = text
      document.body.appendChild(textArea)
      textArea.select()
      document.execCommand('copy')
      document.body.removeChild(textArea)
    }
  }
}
</script>

<style scoped>
.variable-manager {
  padding: 12px;
  border: 1px solid #e8eee5;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 6px 18px rgba(80, 110, 80, 0.08);
}

.manager-header {
  margin-bottom: 16px;
  display: flex;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid #e6ece0;
  border-radius: 10px;
  background: #f7f9f5;
}

.variable-key {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-weight: 500;
  color: #4f7d4f;
  font-size: 14px;
}

.variable-value {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background: #f4faf2;
  border: 1px solid #d9e7d4;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 14px;
  display: inline-block;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

.variable-desc {
  display: inline-block;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

/* 为编辑状态下的输入框设置字体大小 */
.variable-manager :deep(.edit-input .el-input__wrapper) {
  font-size: 14px;
}

.variable-manager :deep(.edit-input .el-input__inner) {
  font-size: 14px;
  height: 32px;
  line-height: 32px;
}

/* 为导入对话框中的文本域设置字体大小 */
.variable-manager :deep(.import-textarea .el-textarea__inner) {
  font-size: 14px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.variable-manager :deep(.el-table) {
  border: 1px solid #e6ece0;
  border-radius: 10px;
  overflow: hidden;
}
</style>




