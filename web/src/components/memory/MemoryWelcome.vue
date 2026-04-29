<template>
  <div class="memory-welcome">
    <section v-if="!configured" class="welcome-card welcome-empty-card">
      <div class="welcome-title">知识片段</div>
      <div class="welcome-subtitle">当前未配置记忆库，请先在配置文件中设置记忆库目录和数据库名，再回到此页面使用。</div>
      <div class="welcome-actions">
        <pl-button type="primary" @click="$emit('go-memory-setting')">查看配置说明</pl-button>
      </div>
    </section>

    <template v-else>
    <div class="welcome-hero">
      <div>
        <div class="welcome-title">知识片段</div>
        <div class="welcome-subtitle">把零散的 Markdown 知识片段收进一个可检索、可追溯、可多标签组织的工作区。</div>
      </div>
      <div class="welcome-actions">
        <pl-button type="primary" @click="$emit('create-fragment')">
          <el-icon><Plus /></el-icon>
          新建片段
        </pl-button>
        <pl-button plain @click="$emit('clear-filter')">
          清空筛选
        </pl-button>
      </div>
    </div>

    <div class="welcome-grid">
      <section class="welcome-card">
        <div class="section-head">
          <span>{{ hasFilter ? '搜索结果' : '最近更新' }}</span>
          <el-tag size="small" effect="light">{{ displayList.length }}</el-tag>
        </div>
        <el-empty v-if="!loading && displayList.length === 0" :description="hasFilter ? '没有匹配的片段' : '还没有片段，先创建一个'" />
        <div v-else class="fragment-list">
          <button
            v-for="item in displayList"
            :key="item.id"
            class="fragment-item"
            @click="$emit('open-fragment', item.id)"
          >
            <div class="fragment-main">
              <div class="fragment-title">{{ item.title }}</div>
              <div class="fragment-meta">
                <span>{{ item.update_time_desc || '-' }}</span>
              </div>
            </div>
            <div v-if="item.tags && item.tags.length > 0" class="fragment-tags">
              <el-tag
                v-for="tag in item.tags.slice(0, 4)"
                :key="tag"
                size="small"
                effect="plain"
              >
                {{ tag }}
              </el-tag>
            </div>
          </button>
        </div>
      </section>

      
    </div>
    </template>
  </div>
</template>

<script>
import { Plus } from '@element-plus/icons-vue'

export default {
  name: 'MemoryWelcome',
  components: {
    Plus,
  },
  props: {
    recentList: {
      type: Array,
      default: () => []
    },
    searchResults: {
      type: Array,
      default: () => []
    },
    tagList: {
      type: Array,
      default: () => []
    },
    selectedTags: {
      type: Array,
      default: () => []
    },
    query: {
      type: String,
      default: ''
    },
    loading: {
      type: Boolean,
      default: false
    },
    configured: {
      type: Boolean,
      default: true
    }
  },
  computed: {
    // hasFilter 判断当前是否存在搜索条件。
    hasFilter() {
      return this.query.trim() !== '' || this.selectedTags.length > 0
    },
    // displayList 返回欢迎页要展示的片段列表。
    displayList() {
      return this.hasFilter ? this.searchResults : this.recentList
    }
  }
}
</script>

<style scoped src="@/css/components/memory/MemoryWelcome.css"></style>

