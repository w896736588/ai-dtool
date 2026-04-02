<template>
  <div class="collection-environment">
    <div class="environment-header">
      <pl-button type="primary" @click="handleAddEnvironment">新增环境</pl-button>
      <pl-button @click="handleImport">导入环境</pl-button>
      <pl-button @click="handleExport">导出环境</pl-button>
    </div>

    <el-table
        :data="environmentList"
        style="width: 100%"
        row-key="id"
        v-loading="loading"
    >
      <el-table-column prop="name" label="环境名称" width="150">
        <template #default="{ row }">
          <el-input
              v-if="row.editing"
              v-model="row.name"
              size="small"
              placeholder="环境名称"
              class="edit-input"
          />
          <span v-else>{{ row.name }}</span>
        </template>
      </el-table-column>

      <el-table-column prop="desc" label="环境描述">
        <template #default="{ row }">
          <el-input
              v-if="row.editing"
              v-model="row.desc"
              size="small"
              placeholder="环境描述"
              class="edit-input"
          />
          <span v-else>{{ row.desc }}</span>
        </template>
      </el-table-column>

      <el-table-column label="变量数量" width="100" align="center">
        <template #default="{ row }">
          <el-tag size="small">{{ row.variables ? Object.keys(row.variables).length : 0 }}</el-tag>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="240" align="center">
        <template #default="{ row, $index }">
          <div v-if="row.editing">
            <pl-button type="primary" link @click="handleSaveEnv(row)">保存</pl-button>
            <pl-button link @click="handleCancelEdit(row, $index)">取消</pl-button>
          </div>
          <div v-else>
            <pl-button type="primary" link @click="handleCopyEnv(row)">复制</pl-button>
            <pl-button type="primary" link @click="handleEditEnv(row)">编辑</pl-button>
            <pl-button type="primary" link @click="handleManageVariables(row)">变量管理</pl-button>
            <pl-button type="danger" link @click="handleDeleteEnv(row)">删除</pl-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- 变量管理对话框 -->
    <el-dialog
        v-model="variableDialogVisible"
        :title="`环境变量管理 - ${currentEnv.name}`"
        width="800px"
    >
      <variable-manager
          ref="refVariableManager"
          @update="handleVariablesUpdate"
      />
      <template #footer>
        <pl-button @click="variableDialogVisible = false">取消</pl-button>
<!--        <pl-button type="primary" @click="handleSaveVariables">保存变量</pl-button>-->
      </template>
    </el-dialog>
  </div>
</template>

<script>
import VariableManager from './VariableManager.vue'
import Api from "@/utils/base/api";

export default {
  name: 'CollectionEnvironment',
  components: {
    VariableManager
  },
  props: {
    collection: {
      type: Object,
      required: true
    },
    environments: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      loading: false,
      environmentList: [],
      variableDialogVisible: false,
      currentEnv: {},
      backupData: null
    }
  },
  watch: {
    collection: {
      handler(newVal) {
        this.loadEnvironmentData(newVal)
      },
      immediate: true
    }
  },
  methods: {
    loadEnvironmentData(collection) {
      let _that = this
      _that.loading = true
      Api.CollectionEnvs({
        collection_id: collection.id
      } , function (res){
        _that.loading = false
        if(res.ErrCode !== 0){
          _that.$message.error(res.ErrMsg)
          return
        }
        _that.environmentList = res.Data.list
      })
    },

    handleAddEnvironment() {
      const newEnv = {
        id: 0,
        name: '',
        desc: '',
        variables: {},
        editing: true
      }
      this.environmentList.unshift(newEnv)
    },

    handleEditEnv(env) {
      this.backupData = { ...env }
      env.editing = true
    },

    handleCopyEnv(env) {
      const copiedEnv = {
        id: 0,
        collection_id: env.collection_id || this.collection.id,
        name: `${env.name || '新环境'}-复制`,
        desc: env.desc || '',
        variables: this.cloneEnvironmentVariables(env.variables),
        copiedVariables: this.cloneEnvironmentVariables(env.variables),
        editing: true
      }
      this.backupData = null
      this.environmentList.unshift(copiedEnv)
    },

    cloneEnvironmentVariables(variables) {
      if (!Array.isArray(variables)) {
        return []
      }
      return variables.map(item => ({
        ...item,
        id: 0,
        env_id: 0
      }))
    },

    saveCopiedVariables(env, savedEnvId) {
      const variables = Array.isArray(env.copiedVariables) ? env.copiedVariables : []
      const validVariables = variables.filter(item => String(item.key || '').trim() !== '')
      if (validVariables.length === 0) {
        return Promise.resolve()
      }
      const tasks = validVariables.map(item => new Promise((resolve, reject) => {
        Api.CreateCollectionEnvItem({
          collection_id: this.collection.id,
          env_id: savedEnvId,
          name: item.name || '',
          desc: item.desc || '',
          key: item.key || '',
          value: item.value || ''
        }, function (res) {
          if (res.ErrCode !== 0) {
            reject(new Error(res.ErrMsg || '复制环境变量失败'))
            return
          }
          resolve(res.Data)
        })
      }))
      return Promise.all(tasks)
    },

    handleSaveEnv(env) {
      if (!env.name.trim()) {
        this.$message.error('请输入环境名称')
        return
      }
      let _that = this
      _that.loading = true
      Api.CreateCollectionEnv({
        collection_id: _that.collection.id,
        name : env.name,
        desc : env.desc,
        id : env.id,
      } , async function (res){
        _that.loading = false
        if(res.ErrCode !== 0){
          _that.$message.error(res.ErrMsg)
          return
        }
        let saveWarning = ''
        try {
          const isNewEnv = parseInt(env.id) === 0
          if (isNewEnv) {
            await _that.saveCopiedVariables(env, res.Data.id)
            if (Array.isArray(env.copiedVariables) && env.copiedVariables.length > 0) {
              const variableRes = await new Promise((resolve, reject) => {
                Api.CollectionEnvItems({
                  collection_id: _that.collection.id,
                  env_id: res.Data.id,
                }, function (itemsRes) {
                  if (itemsRes.ErrCode !== 0) {
                    reject(new Error(itemsRes.ErrMsg || '刷新环境变量失败'))
                    return
                  }
                  resolve(itemsRes.Data.list)
                })
              })
              res.Data.variables = variableRes
            }
          }
        } catch (error) {
          saveWarning = error.message || '环境已保存，但复制环境变量失败'
        }
        env.editing = false
        for (let i in _that.environmentList) {
          if (parseInt(res.Data.id) === parseInt(_that.environmentList[i].id) || (parseInt(env.id) === 0 && parseInt(_that.environmentList[i].id) === 0)) {
            _that.environmentList[i] = res.Data
            _that.$emit('environmentUpdate', _that.environmentList)
            if (saveWarning) {
              _that.$message.warning(saveWarning)
            } else {
              _that.$message.success('保存成功')
            }
            return
          }
        }
        _that.environmentList.push(res.Data)
        _that.$emit('environmentUpdate', _that.environmentList)
        if (saveWarning) {
          _that.$message.warning(saveWarning)
        } else {
          _that.$message.success('保存成功')
        }
      })
    },

    handleCancelEdit(env, index) {
      if (this.backupData) {
        Object.assign(env, this.backupData)
        env.editing = false
      } else {
        this.environmentList.splice(index, 1)
      }
      this.backupData = null
    },

    handleDeleteEnv(env) {
      this.$confirm(`确定要删除环境 "${env.name}" 吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        const index = this.environmentList.findIndex(item => item.id === env.id)
        if (index !== -1) {
          this.environmentList.splice(index, 1)
          this.$message.success('删除成功')
        }
      })
    },

    handleManageVariables(env) {
      let _that = this
      Api.CollectionEnvItems({
        collection_id: env.collection_id,
        env_id : env.id,
      } , function (res){
        _that.loading = false
        if(res.ErrCode !== 0){
          _that.$message.error(res.ErrMsg)
          return
        }
        env.variables = res.Data.list
        _that.currentEnv = env
        _that.variableDialogVisible = true
        _that.$nextTick(() => {
          _that.$refs.refVariableManager.loadVariables(env)
        })
      })

    },

    handleVariablesUpdate(variables) {
      this.currentEnv.variables = variables
    },

    // handleSaveVariables() {
    //   this.variableDialogVisible = false
    //   this.$message.success('变量保存成功')
    // },

    handleImport() {
      // 实现导入环境逻辑
      this.$message.info('导入环境功能开发中')
    },

    handleExport() {
      // 实现导出环境逻辑
      this.$message.info('导出环境功能开发中')
    }
  }
}
</script>

<style scoped>
.collection-environment {
  padding: 12px;
  border: 1px solid #e8eee5;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 6px 18px rgba(80, 110, 80, 0.08);
}

.environment-header {
  margin-bottom: 20px;
  padding: 10px 12px;
  border: 1px solid #e6ece0;
  border-radius: 10px;
  background: #f7f9f5;
}

/* 为编辑状态下的输入框设置字体大小 */
.collection-environment :deep(.edit-input .el-input__wrapper) {
  font-size: 14px;
}

.collection-environment :deep(.edit-input .el-input__inner) {
  font-size: 14px;
  height: 32px;
  line-height: 32px;
}

.collection-environment :deep(.el-table) {
  border: 1px solid #e6ece0;
  border-radius: 10px;
  overflow: hidden;
}
</style>




