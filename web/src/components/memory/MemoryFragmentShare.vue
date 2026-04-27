<template>
  <div class="memory-share-page">
    <main class="memory-share-shell">
      <div v-if="loading" class="memory-share-state">
        <el-icon class="memory-share-loading"><Loading /></el-icon>
        <span>正在打开分享链接...</span>
      </div>
      <el-empty
        v-else-if="errorMessage"
        :description="errorMessage"
      />
      <article v-else class="memory-share-viewer">
        <header class="memory-share-header">
          <h1>{{ fragment.title || '未命名片段' }}</h1>
          <div class="memory-share-meta">
            <span v-if="fragment.update_time_desc">更新：{{ fragment.update_time_desc }}</span>
            <span v-if="share.expire_at_desc">链接有效期至：{{ share.expire_at_desc }}</span>
          </div>
        </header>
        <section class="memory-share-content">
          <MdPreview
            :model-value="fragment.content || ''"
            preview-theme="github"
          />
        </section>
      </article>
    </main>
  </div>
</template>

<script>
import { Loading } from '@element-plus/icons-vue'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import MemoryFragmentApi from '@/utils/base/memory_fragment'

export default {
  name: 'MemoryFragmentShare',
  components: {
    Loading,
    MdPreview,
  },
  data() {
    return {
      loading: false,
      errorMessage: '',
      fragment: {
        title: '',
        content: '',
        update_time_desc: '',
      },
      share: {
        expire_at_desc: '',
      },
    }
  },
  mounted() {
    this.loadShareInfo()
  },
  watch: {
    '$route.query.token'() {
      this.loadShareInfo()
    },
  },
  methods: {
    loadShareInfo() {
      const token = String(this.$route.query.token || '').trim()
      if (!token) {
        this.errorMessage = '分享链接缺少 token'
        return
      }
      this.loading = true
      this.errorMessage = ''
      MemoryFragmentApi.MemoryFragmentShareInfo(token, (response) => {
        this.loading = false
        if (response.ErrCode !== 0 || !response.Data) {
          this.errorMessage = response.ErrMsg || '分享链接不可用'
          return
        }
        this.fragment = response.Data.fragment || {}
        this.share = response.Data.share || {}
      })
    },
  },
}
</script>

<style scoped>
.memory-share-page {
  min-height: 100vh;
  background: #f5f7f2;
  color: #2f3c2b;
}

.memory-share-shell {
  width: min(960px, calc(100% - 32px));
  margin: 0 auto;
  padding: 32px 0;
}

.memory-share-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 320px;
  color: #5f7059;
  font-size: 14px;
}

.memory-share-loading {
  animation: memory-share-spin 1s linear infinite;
}

.memory-share-viewer {
  min-height: calc(100vh - 64px);
  border: 1px solid #e2e8d8;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(54, 74, 54, 0.08);
  overflow: hidden;
}

.memory-share-header {
  padding: 24px 28px 18px;
  border-bottom: 1px solid #e8eee0;
  background: #f8faf5;
}

.memory-share-header h1 {
  margin: 0;
  color: #263523;
  font-size: 24px;
  line-height: 1.35;
  font-weight: 700;
  word-break: break-word;
}

.memory-share-meta {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 10px;
  color: #687762;
  font-size: 13px;
}

.memory-share-content {
  padding: 22px 28px 32px;
}

.memory-share-content :deep(.md-editor-preview) {
  font-size: 14px;
  color: #33422f;
}

@keyframes memory-share-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 720px) {
  .memory-share-shell {
    width: calc(100% - 20px);
    padding: 10px 0;
  }

  .memory-share-viewer {
    min-height: calc(100vh - 20px);
  }

  .memory-share-header,
  .memory-share-content {
    padding-left: 16px;
    padding-right: 16px;
  }
}
</style>
