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
      <div v-else class="memory-share-layout">
        <article class="memory-share-viewer">
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
              :onGetCatalog="onGetCatalog"
            />
          </section>
        </article>
        <!-- 右侧目录导航 -->
        <aside v-if="tocItems.length > 0" class="memory-share-toc">
          <div class="toc-title">目录</div>
          <nav class="toc-nav">
            <a
              v-for="item in tocItems"
              :key="item.id"
              :class="['toc-link', 'toc-h' + item.level, { 'toc-active': activeHeading === item.id }]"
              @click.prevent="scrollToHeading(item.id)"
            >{{ item.text }}</a>
          </nav>
        </aside>
      </div>
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
      tocItems: [],
      activeHeading: '',
      scrollObserver: null,
    }
  },
  mounted() {
    this.loadShareInfo()
  },
  beforeUnmount() {
    this.cleanupScrollObserver()
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
      this.tocItems = []
      this.activeHeading = ''
      this.cleanupScrollObserver()
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

    /**
     * MdPreview 渲染完成后的目录回调，结合 DOM 获取标题 ID
     */
    onGetCatalog(catalogList) {
      if (!catalogList || catalogList.length === 0) {
        this.tocItems = []
        return
      }
      this.$nextTick(() => {
        const previewEl = this.$el.querySelector('.md-editor-preview')
        if (!previewEl) {
          this.tocItems = catalogList.map((item) => ({
            level: item.level,
            text: item.text,
            id: '',
          }))
          return
        }
        // 从 DOM 获取实际标题 ID，用于滚动定位
        const domHeadings = previewEl.querySelectorAll('h1, h2, h3, h4')
        this.tocItems = catalogList.map((item, i) => ({
          level: item.level,
          text: item.text,
          id: domHeadings[i] ? domHeadings[i].id : '',
        }))
        if (this.tocItems.length > 0) {
          this.$nextTick(() => this.setupScrollSpy())
        }
      })
    },

    /**
     * 使用 IntersectionObserver 追踪当前可见标题
     */
    setupScrollSpy() {
      this.cleanupScrollObserver()
      const previewEl = this.$el.querySelector('.md-editor-preview')
      if (!previewEl) return

      const headingEls = previewEl.querySelectorAll('h1, h2, h3, h4')
      if (headingEls.length === 0) return

      this.scrollObserver = new IntersectionObserver(
        (entries) => {
          for (const entry of entries) {
            if (entry.isIntersecting && entry.target.id) {
              this.activeHeading = entry.target.id
              break
            }
          }
        },
        { rootMargin: '-80px 0px -60% 0px', threshold: 0.1 }
      )
      headingEls.forEach((el) => this.scrollObserver.observe(el))
    },

    /**
     * 清理 IntersectionObserver
     */
    cleanupScrollObserver() {
      if (this.scrollObserver) {
        this.scrollObserver.disconnect()
        this.scrollObserver = null
      }
    },

    /**
     * 点击目录项，平滑滚动到对应标题
     */
    scrollToHeading(id) {
      const el = document.getElementById(id)
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'start' })
        this.activeHeading = id
      }
    },
  },
}
</script>

<style scoped src="@/css/components/memory/MemoryFragmentShare.css"></style>
