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

      <el-form-item label="默认环境">
        <el-select v-model="form.env_id" placeholder="选择环境" style="width: 100%" @change="handleEnvChange">
          <el-option
              v-for="env in envs"
              :key="env.id"
              :label="env.name"
              :value="env.id"
          />
        </el-select>
        <div class="env-hint" style="color: #909399; font-size: 12px; margin-top: 4px;">
          接口未单独配置环境时，将使用此环境
        </div>
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
import Api from '@/utils/base/api'

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
      envs: [],
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
        env_id: parseInt(folder.env_id) || 0,
        apiCount: folder.apiCount || 0,
      }
      this.loadEnvs()
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

    handleEnvChange(envId) {
      this.form.env_id = envId
    },

    loadEnvs() {
      if (!this.folder.collection_id) return
      let _that = this
      Api.CollectionEnvs({
        collection_id: this.folder.collection_id,
      }, function (res) {
        if (res.ErrCode !== 0) return
        _that.envs = (res.Data.list || []).map(function(env) { return {id: parseInt(env.id), name: env.name} })
        _that.envs.unshift({id: 0, name: '不设置'})
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

<style scoped src="@/css/components/api/FolderBasicInfo.css"></style>

