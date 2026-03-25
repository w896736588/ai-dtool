<template>
  <div class="memory-welcome">
    <section v-if="!configured" class="welcome-card welcome-empty-card">
      <div class="welcome-title">知识片段</div>
      <div class="welcome-subtitle">请先到设置页面配置记忆目录和数据库名</div>
      <div class="welcome-actions">
        <pl-button type="primary" @click="$emit('go-memory-setting')">去设置</pl-button>
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
                <span>{{ item.index_status_desc || '待索引' }}</span>
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

      <section class="welcome-card">
        <div class="section-head">
          <span>标签面板</span>
          <el-tag size="small" effect="light">{{ tagList.length }}</el-tag>
        </div>
        <div v-if="selectedTags.length > 0" class="selected-tags">
          <span class="selected-label">当前筛选</span>
          <el-tag
            v-for="tag in selectedTags"
            :key="tag"
            size="small"
            closable
            @close="$emit('toggle-tag', tag)"
          >
            {{ tag }}
          </el-tag>
        </div>
        <el-empty v-if="tagList.length === 0" description="暂无标签" />
        <div v-else class="tag-cloud">
          <button
            v-for="tag in tagList"
            :key="tag.tag_name"
            class="tag-chip"
            :class="{ active: selectedTags.includes(tag.tag_name) }"
            @click="$emit('toggle-tag', tag.tag_name)"
          >
            <span>{{ tag.tag_name }}</span>
            <span class="tag-count">{{ tag.use_count }}</span>
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

<style scoped>
.memory-welcome {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100%;
}

.welcome-empty-card {
  align-items: flex-start;
  justify-content: center;
}

.welcome-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 22px 24px;
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  background: linear-gradient(135deg, #f7fbf2 0%, #ffffff 55%, #eef5e8 100%);
  box-shadow: 0 4px 14px rgba(66, 88, 57, 0.05);
}

.welcome-title {
  font-size: 26px;
  font-weight: 700;
  color: #47643f;
  margin-bottom: 8px;
}

.welcome-subtitle {
  max-width: 720px;
  line-height: 1.7;
  color: #5d6f58;
  font-size: 14px;
}

.welcome-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.welcome-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(320px, 0.8fr);
  gap: 16px;
}

.welcome-card {
  min-height: 280px;
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 4px 12px rgba(54, 74, 54, 0.05);
  padding: 18px;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  font-size: 15px;
  font-weight: 600;
  color: #4e5f49;
}

.fragment-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.fragment-item {
  border: 1px solid #edf1e8;
  border-radius: 12px;
  background: #fbfcf8;
  padding: 14px;
  cursor: pointer;
  text-align: left;
  transition: all 0.2s ease;
}

.fragment-item:hover {
  border-color: #cfe0c8;
  background: #f5f9ef;
  transform: translateY(-1px);
}

.fragment-main {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}

.fragment-title {
  color: #354531;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
}

.fragment-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  color: #7a8773;
  font-size: 12px;
  white-space: nowrap;
}

.fragment-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: 10px;
}

.selected-tags {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.selected-label {
  font-size: 13px;
  color: #66785f;
}

.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border: 1px solid #dbe7d4;
  border-radius: 999px;
  padding: 8px 12px;
  background: #f8fbf5;
  color: #4f6448;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tag-chip.active {
  border-color: #81a478;
  background: #edf6e7;
  color: #35512f;
}

.tag-count {
  min-width: 20px;
  padding: 1px 6px;
  border-radius: 999px;
  background: rgba(86, 123, 76, 0.12);
  font-size: 12px;
}

@media (max-width: 1080px) {
  .welcome-grid {
    grid-template-columns: 1fr;
  }

  .welcome-hero {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

