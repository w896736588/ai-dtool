<template>
  <div class="memory-fullscreen-page">
    <!-- 加载中 -->
    <div v-if="loading" class="memory-fullscreen-loading">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <span>正在加载知识片段...</span>
    </div>

    <!-- 加载失败 -->
    <el-empty
      v-else-if="errorMessage"
      :description="errorMessage"
    />

    <!-- MemoryEditor 全屏编辑区 -->
    <div v-else-if="fragmentData.id" class="memory-fullscreen-editor-wrapper">
      <MemoryEditor
        ref="editorRef"
        :fragment="fragmentData"
        :saved-fragment="savedFragment"
        :available-tags="[]"
        :show-outline-sidebar="true"
        @change="handleEditorChange"
        @saved="handleEditorSaved"
        @deleted="handleEditorDeleted"
        @show-history="handleShowHistory"
      />
    </div>

    <!-- 历史记录弹窗 -->
    <MemoryHistoryDialog
      v-model="historyDialogVisible"
      :fragment-id="fragmentData.id"
      :git-repo-enabled="memoryIsGitRepo"
      :is-git-repo="memoryIsGitRepo"
    />
  </div>
</template>

<script>
import { Loading } from '@element-plus/icons-vue'
import MemoryEditor from '@/components/memory/MemoryEditor.vue'
import MemoryHistoryDialog from '@/components/memory/MemoryHistoryDialog.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'

/**
 * MemoryFragmentFullscreen
 * 知识片段全屏编辑页面。
 * 从 URL query 读取 fragment_id，加载片段数据后以编辑模式独立展示 MemoryEditor。
 * 用于从 MemoryFragment 页面右侧工具栏"全屏"按钮跳转打开的新页面。
 */
export default {
  name: 'MemoryFragmentFullscreen',
  components: {
    Loading,
    MemoryEditor,
    MemoryHistoryDialog,
  },
  data() {
    return {
      loading: false,
      errorMessage: '',
      fragmentData: {
        id: '',
        title: '',
        content: '',
        file_path: '',
        folder_name: 'fragments',
        update_time_desc: '',
        create_time_desc: '',
      },
      savedFragment: {
        id: '',
        title: '',
        content: '',
        file_path: '',
        folder_name: 'fragments',
        update_time_desc: '',
        create_time_desc: '',
      },
      historyDialogVisible: false,
      memoryIsGitRepo: false,
    }
  },
  computed: {
    /**
     * fragmentId 从路由 query 中提取 fragment_id。
     */
    fragmentId() {
      return String(this.$route.query.fragment_id || '').trim()
    },
  },
  mounted() {
    if (!this.fragmentId) {
      this.errorMessage = '缺少 fragment_id 参数'
      return
    }
    this.loadFragment()
  },
  methods: {
    /**
     * loadFragment 通过 API 加载片段详情数据。
     */
    loadFragment() {
      this.loading = true
      this.errorMessage = ''
      MemoryFragmentApi.MemoryFragmentInfo(this.fragmentId, (response) => {
        this.loading = false
        if (response.ErrCode !== 0 || !response.Data) {
          this.errorMessage = response.ErrMsg || '加载片段失败'
          return
        }
        const data = response.Data
        const normalized = this.normalizeFragment(data)
        this.fragmentData = normalized
        this.savedFragment = this.cloneFragment(normalized)
        // 进入编辑模式
        this.$nextTick(() => {
          if (this.$refs.editorRef && this.$refs.editorRef.setContentEditMode) {
            this.$refs.editorRef.setContentEditMode(true)
          }
        })
      })
    },

    /**
     * normalizeFragment 统一片段对象结构。
     */
    normalizeFragment(fragment) {
      return {
        id: String(fragment.id || fragment.file_id || '').trim(),
        title: fragment.title || '',
        content: fragment.content || '',
        file_path: fragment.file_path || '',
        folder_name: fragment.folder_name || 'fragments',
        update_time_desc: fragment.update_time_desc || '',
        create_time_desc: fragment.create_time_desc || '',
      }
    },

    /**
     * cloneFragment 深拷贝片段对象。
     */
    cloneFragment(fragment) {
      return JSON.parse(JSON.stringify(fragment))
    },

    /**
     * handleEditorChange 编辑器内容变更回调。
     * @param {object} fragment 变更后的草稿片段
     */
    handleEditorChange(fragment) {
      // 编辑器内部自行管理草稿，父组件仅需感知变更
    },

    /**
     * handleEditorSaved 编辑器保存成功回调，更新 savedFragment。
     * @param {object} fragment 保存后的片段数据
     */
    handleEditorSaved(fragment) {
      if (fragment && fragment.id) {
        const normalized = this.normalizeFragment(fragment)
        this.savedFragment = normalized
        // 同步 fragmentData 中的标题等字段
        this.fragmentData = { ...this.fragmentData, ...normalized }
      }
    },

    /**
     * handleEditorDeleted 编辑器删除片段回调，返回知识片段主页。
     */
    handleEditorDeleted() {
      this.$router.replace('/MemoryFragment')
    },

    /**
     * handleShowHistory 打开历史记录弹窗。
     */
    handleShowHistory() {
      this.historyDialogVisible = true
    },

  },
}
</script>

<style scoped src="@/css/components/memory/MemoryFragmentFullscreen.css"></style>
