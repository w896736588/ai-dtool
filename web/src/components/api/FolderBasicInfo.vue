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

      <el-form-item label="默认请求头">
        <headers-value-editor
            v-model="form.headers"
            :keys="headerSuggestions"
            @update="handleHeadersUpdate"
        />
      </el-form-item>

      <el-form-item label="接口数量">
        <el-tag type="info">{{ form.apiCount || 0 }} 个接口</el-tag>
      </el-form-item>

      <el-form-item>
        <pl-button type="primary" @click="handleSave">保存更改</pl-button>
        <pl-button type="danger" @click="handleDelete" v-if="folder.id">删除文件夹</pl-button>
      </el-form-item>
    </el-form>

  </div>
</template>

<script>
import HeadersValueEditor from './HeadersValueEditor.vue'

// 中文注释：复用接口详情页中常见的请求头候选项。
const HEADER_SUGGESTIONS = [
  'Content-Type',
  'Authorization',
  'User-Agent',
  'Accept',
  'Cookie',
  'Token',
]

// 中文注释：统一解析文件夹上保存的请求头，兼容字符串和对象两种入参。
function parseFolderHeaders(headersValue) {
  if (!headersValue) {
    return {}
  }
  if (typeof headersValue === 'object') {
    return { ...headersValue }
  }
  try {
    const parsedHeaders = JSON.parse(headersValue)
    return parsedHeaders && typeof parsedHeaders === 'object' ? parsedHeaders : {}
  } catch (error) {
    return {}
  }
}

export default {
  name: 'FolderBasicInfo',
  components: {
    HeadersValueEditor,
  },
  props: {
    folder: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      form: {},
      headerSuggestions: HEADER_SUGGESTIONS,
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
        headers: parseFolderHeaders(folder.headers),
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
    handleHeadersUpdate(headers) {
      this.form.headers = headers || {}
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
  padding: 12px;
  border: 1px solid #e8eee5;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 6px 18px rgba(80, 110, 80, 0.08);
}

.info-form {
  max-width: 600px;
}

.folder-basic-info :deep(.el-input__wrapper),
.folder-basic-info :deep(.el-textarea__inner) {
  border-radius: 8px;
}

.folder-basic-info :deep(.el-form-item:last-child .el-form-item__content) {
  gap: 10px;
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
  background: #f7f9f5;
  border-radius: 8px;
  border: 1px solid #e6ece0;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #4f7d4f;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}
</style>

