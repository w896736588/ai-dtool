<template>
  <div class="memory-raw-page">
    <main class="memory-raw-shell">
      <div v-if="loading" class="memory-raw-state">
        <el-icon class="memory-raw-loading"><Loading /></el-icon>
        <span>正在打开分享链接...</span>
      </div>
      <el-empty
        v-else-if="errorMessage"
        :description="errorMessage"
      />
      <article v-else class="memory-raw-viewer">
        <header class="memory-raw-header">
          <div class="memory-raw-heading">
            <h1>{{ fragment.title || '未命名片段' }}</h1>
            <el-button
              class="memory-raw-copy"
              type="primary"
              size="small"
              :icon="CopyDocument"
              @click="copyRaw"
            >复制原文</el-button>
          </div>
          <div class="memory-raw-meta">
            <span v-if="fragment.update_time_desc">更新：{{ fragment.update_time_desc }}</span>
            <span v-if="share.expire_at_desc">链接有效期至：{{ share.expire_at_desc }}</span>
          </div>
        </header>
        <section class="memory-raw-content">
          <pre class="memory-raw-pre">{{ fragment.content || '' }}</pre>
        </section>
      </article>
    </main>
  </div>
</template>

<script>
import { Loading, CopyDocument } from '@element-plus/icons-vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'

export default {
  name: 'MemoryFragmentRaw',
  components: {
    Loading,
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
    '$route.query.id'() {
      this.loadShareInfo()
    },
  },
  methods: {
    loadShareInfo() {
      const id = String(this.$route.query.id || '').trim()
      if (!id) {
        this.errorMessage = '分享链接缺少片段 ID'
        return
      }
      this.loading = true
      this.errorMessage = ''
      MemoryFragmentApi.MemoryFragmentInfo(id, (response) => {
        this.loading = false
        if (response.ErrCode !== 0 || !response.Data) {
          this.errorMessage = response.ErrMsg || '分享链接不可用'
          return
        }
        this.fragment = response.Data || {}
        this.share = {}
      })
    },
    copyRaw() {
      const value = String(this.fragment.content || '').trim()
      if (!value) {
        this.$helperNotify.error('没有可复制的内容')
        return
      }
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(value).then(() => {
          this.$helperNotify.success('原文已复制到剪贴板')
        }).catch(() => {
          this.fallbackCopy(value)
        })
        return
      }
      this.fallbackCopy(value)
    },
    fallbackCopy(text) {
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      try {
        document.execCommand('copy')
        this.$helperNotify.success('原文已复制到剪贴板')
      } catch (error) {
        this.$helperNotify.error('复制失败')
      }
      document.body.removeChild(textArea)
    },
  },
}
</script>

<style scoped src="@/css/components/memory/MemoryFragmentRaw.css"></style>
