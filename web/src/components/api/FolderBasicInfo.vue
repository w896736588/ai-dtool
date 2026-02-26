<template>
  <div class="folder-basic-info">
    <el-form
        ref="formRef"
        :model="form"
        label-width="100px"
        label-position="left"
        class="info-form"
    >
      <el-form-item label="文件夹名称">
        <el-input
            v-model="form.name"
            placeholder="请输入文件夹名称"
            maxlength="50"
            show-word-limit
        />
      </el-form-item>

      <el-form-item label="描述信息">
        <el-input
            v-model="form.desc"
            type="textarea"
            :rows="3"
            placeholder="请输入文件夹描述"
            maxlength="200"
            show-word-limit
        />
      </el-form-item>

      <el-form-item label="接口数量">
        <el-tag type="info">{{ form.apiCount || 0 }} 个接口</el-tag>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="handleSave">保存更改</el-button>
        <el-button type="danger" @click="handleDelete" v-if="folder.id">删除文件夹</el-button>
      </el-form-item>
    </el-form>

  </div>
</template>

<script>
export default {
  name: 'FolderBasicInfo',
  props: {
    folder: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      form: {}
    }
  },
  watch: {
    folder: {
      handler(newVal) {
        this.loadFolderData(newVal)
      },
      immediate: true,
      deep: true
    }
  },
  methods: {
    loadFolderData(folder) {
      this.form = {
        name: folder.name || '',
        desc: folder.desc || '',
        apiCount: folder.apiCount || 0,
      }
    },

    formatTime(timeString) {
      if (!timeString) return '-'
      return new Date(timeString).toLocaleString('zh-CN')
    },

    handleSave() {
      if (!this.form.name.trim()) {
        this.$message.error('请输入文件夹名称')
        return
      }

      this.$emit('update', {
        ...this.folder,
        ...this.form,
        updateTime: new Date().toISOString()
      })
    },

    handleReset() {
      this.loadFolderData(this.folder)
      this.$message.info('已重置')
    },
    handleDelete() {
      let _that = this
      _that.$emit('delete', _that.folder)
    }
  }
}
</script>

<style scoped>
.folder-basic-info {
  padding: 20px 0;
}

.info-form {
  max-width: 600px;
}

.readonly-text {
  color: #606266;
  font-size: 14px;
}

.stats-row {
  margin-top: 20px;
}

.stat-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}
</style>