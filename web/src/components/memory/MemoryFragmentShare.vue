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

<style scoped src="@/css/components/memory/MemoryFragmentShare.css"></style>
