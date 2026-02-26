<template>
  <div class="api-document-page">
    <div v-if="loading" class="loading-container">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>Loading...</span>
    </div>
    <div v-else-if="error" class="error-container">
      <el-alert :title="error" type="error" :closable="false" show-icon />
    </div>
    <api-document
        v-else
        :apis="folderApis"
        :folder-name="folderName"
    />
  </div>
</template>

<script>
import Api from '@/utils/base/api'
import ApiDocument from '@/components/api/ApiDocument.vue'
import { Loading } from '@element-plus/icons-vue'

export default {
  name: 'ApiDocumentPage',
  components: {
    ApiDocument,
    Loading
  },
  data() {
    return {
      loading: false,
      error: '',
      folderApis: [],
      folderName: ''
    }
  },
  mounted() {
    this.loadFolderData()
  },
  methods: {
    loadFolderData() {
      const folderId = this.$route.params.folderId
      
      if (!folderId) {
        this.error = 'Missing required parameter: folderId'
        return
      }

      this.loading = true
      this.error = ''
      
      const _this = this
      
      // Use new FolderDetail API to get folder info and APIs
      Api.FolderDetail({
        dir_id: folderId
      }, function (res) {
        _this.loading = false
        if (res.ErrCode === 0) {
          _this.folderName = res.Data.dir.name || ''
          _this.folderApis = res.Data.dir.children || []
        } else {
          _this.error = res.ErrMsg || 'Failed to load folder data'
        }
      })
    }
  }
}
</script>

<style scoped>
.api-document-page {
  height: 100vh;
  overflow: hidden;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  gap: 16px;
  color: #909399;
}

.error-container {
  padding: 20px;
}
</style>
