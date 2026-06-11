<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    title="文件变更详情"
    width="90%"
    top="3vh"
    :close-on-click-modal="true"
    @close="handleClose"
  >
    <div class="file-changes-detail" v-loading="loading">
      <!-- 工具栏 -->
      <div class="file-changes-detail__toolbar">
        <div class="file-changes-detail__info">
          <span class="file-changes-detail__dir">{{ localDir || '-' }}</span>
          <span class="file-changes-detail__branch">基分支: {{ parentBranch || '-' }}</span>
          <span class="file-changes-detail__summary">
            <template v-if="summary">
              <span class="file-changes-stat file-changes-stat--added">待add {{ summary.untracked || 0 }}</span>
              <span class="file-changes-stat file-changes-stat--modified">编辑 {{ summary.modified || 0 }}</span>
              <span class="file-changes-stat file-changes-stat--deleted">删除 {{ summary.deleted || 0 }}</span>
              <span v-if="summary.other > 0" class="file-changes-stat file-changes-stat--other">其他 {{ summary.other }}</span>
            </template>
          </span>
        </div>
        <div class="file-changes-detail__actions">
          <el-radio-group v-model="diffMode" size="small" @change="renderCurrentDiff">
            <el-radio-button value="side-by-side">横向对比</el-radio-button>
            <el-radio-button value="line-by-line">纵向对比</el-radio-button>
          </el-radio-group>
        </div>
      </div>
      <!-- 主体：左侧文件树 + 右侧 diff -->
      <div class="file-changes-detail__body">
        <!-- 左侧文件树 -->
        <div class="file-changes-detail__tree-panel">
          <div class="file-changes-detail__tree-title">改动文件列表 ({{ fileTreeTotal }})</div>
          <div class="file-changes-detail__tree">
            <template v-if="fileTree.length > 0 || rootFiles.length > 0">
              <div v-for="node in fileTree" :key="node.key">
                <div
                  v-if="node.isDir"
                  class="file-changes-tree__dir"
                  @click="toggleDir(node.key)"
                >
                  <span class="file-changes-tree__dir-icon">{{ node.expanded ? '&#x25BC;' : '&#x25B6;' }}</span>
                  <span class="file-changes-tree__dir-name">{{ node.name }}/</span>
                  <span class="file-changes-tree__dir-count">({{ node.fileCount }})</span>
                </div>
                <div v-if="node.isDir && node.expanded">
                  <div
                    v-for="file in node.files"
                    :key="file.path"
                    class="file-changes-tree__file"
                    :class="{ 'file-changes-tree__file--active': selectedFile === file.path }"
                    @click="selectFile(file)"
                  >
                    <span class="file-changes-tree__file-type file-changes-tree__file-type--{{ file.type }}">{{ file.status_code }}</span>
                    <span class="file-changes-tree__file-name" :title="file.path">{{ file.name }}</span>
                  </div>
                </div>
              </div>
              <!-- 根目录文件 -->
              <div v-if="rootFiles.length > 0">
                <div
                  v-for="file in rootFiles"
                  :key="file.path"
                  class="file-changes-tree__file"
                  :class="{ 'file-changes-tree__file--active': selectedFile === file.path }"
                  @click="selectFile(file)"
                >
                  <span class="file-changes-tree__file-type file-changes-tree__file-type--{{ file.type }}">{{ file.status_code }}</span>
                  <span class="file-changes-tree__file-name" :title="file.path">{{ file.path }}</span>
                </div>
              </div>
            </template>
            <div v-else class="file-changes-detail__tree-empty">暂无文件变更</div>
          </div>
        </div>
        <!-- 右侧 diff 视图 -->
        <div class="file-changes-detail__diff-panel">
          <template v-if="diffError">
            <div class="file-changes-detail__diff-error">{{ diffError }}</div>
          </template>
          <template v-else-if="selectedFile && currentDiff">
            <div class="file-changes-detail__diff-header">
              <span class="file-changes-detail__diff-file">{{ selectedFile }}</span>
            </div>
            <div ref="diffContainer" class="file-changes-detail__diff-content"></div>
          </template>
          <template v-else>
            <div class="file-changes-detail__diff-placeholder">
              <span>请从左侧选择文件查看变更详情</span>
            </div>
          </template>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="file-changes-detail__footer">
        <GitActionButton compact variant="info" @click="handleClose">关闭</GitActionButton>
      </div>
    </template>
  </el-dialog>
</template>

<script>
import GitActionButton from '@/components/base/GitActionButton.vue'
import taskWorkflowApi from '@/utils/base/task_workflow'
import * as Diff2Html from 'diff2html'
import 'diff2html/bundles/css/diff2html.min.css'

export default {
  name: 'FileChangesDialog',
  components: { GitActionButton },
  props: {
    visible: { type: Boolean, default: false },
    localDir: { type: String, default: '' },
    parentBranch: { type: String, default: '' },
    initialSummary: { type: Object, default: null },
    initialFiles: { type: Array, default: () => [] },
  },
  emits: ['update:visible'],
  data() {
    return {
      loading: false,
      summary: null,
      files: [],
      diffs: {},
      selectedFile: '',
      currentDiff: '',
      diffError: '',
      diffMode: 'side-by-side',
      treeExpanded: {},
      _rootFiles: [],
    }
  },
  computed: {
    fileTree() {
      const dirMap = {}
      const rootFiles = []
      for (const file of this.files) {
        const path = file.path || ''
        const slashIdx = path.indexOf('/')
        if (slashIdx < 0) {
          rootFiles.push(file)
        } else {
          const dirName = path.substring(0, slashIdx)
          const fileName = path.substring(slashIdx + 1)
          if (!dirMap[dirName]) {
            dirMap[dirName] = { name: dirName, files: [], fileCount: 0 }
          }
          dirMap[dirName].files.push({ ...file, name: fileName })
          dirMap[dirName].fileCount++
        }
      }
      const tree = []
      for (const dirName of Object.keys(dirMap).sort()) {
        const node = dirMap[dirName]
        node.isDir = true
        node.key = 'dir_' + dirName
        node.expanded = this.treeExpanded[node.key] !== false
        tree.push(node)
      }
      this._rootFiles = rootFiles
      return tree
    },
    rootFiles() {
      return this._rootFiles || []
    },
    fileTreeTotal() {
      return this.files.length
    },
  },
  watch: {
    visible(val) {
      if (val) {
        this.initData()
        this.loadDetail()
      }
    },
  },
  methods: {
    initData() {
      this.summary = this.initialSummary ? { ...this.initialSummary } : null
      this.files = Array.isArray(this.initialFiles) ? [...this.initialFiles] : []
      this.diffs = {}
      this.selectedFile = ''
      this.currentDiff = ''
      this.diffError = ''
      this.treeExpanded = {}
    },
    loadDetail() {
      if (!this.localDir) return
      // 如果有初始数据，不显示 loading 遮罩，优先展示已有数据
      if (!this.summary && this.files.length === 0) {
        this.loading = true
      }
      taskWorkflowApi.TaskWorkflowFileChangesDetail(this.localDir, this.parentBranch, (response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data) {
          const data = response.Data
          // 修复 JS 空数组 truthy 陷阱：[] || [...] 返回 []，需要检查 length
          if (data.summary && typeof data.summary === 'object' && Object.keys(data.summary).length > 0) {
            this.summary = data.summary
          }
          if (Array.isArray(data.files) && data.files.length > 0) {
            this.files = data.files
          }
          if (data.diffs && typeof data.diffs === 'object') {
            this.diffs = data.diffs
          }
        } else {
          this.diffError = (response && response.ErrMsg) || '加载详情失败'
        }
      })
    },
    handleClose() {
      this.$emit('update:visible', false)
      this.selectedFile = ''
      this.currentDiff = ''
      this.diffError = ''
    },
    toggleDir(key) {
      this.treeExpanded = {
        ...this.treeExpanded,
        [key]: this.treeExpanded[key] === true ? false : true,
      }
    },
    selectFile(file) {
      this.selectedFile = file.path
      this.diffError = ''
      if (this.diffs && this.diffs[file.path]) {
        this.renderDiff(this.diffs[file.path])
      } else if (file.type === 'untracked') {
        this.diffError = `文件 "${file.path}" 是未跟踪的新文件，暂无法对比差异。`
      } else {
        this.diffError = `该文件暂无差异数据。`
      }
    },
    renderDiff(diffText) {
      try {
        const diffHtml = Diff2Html.html(diffText, {
          drawFileList: false,
          matching: 'lines',
          outputFormat: this.diffMode,
          renderNothingWhenEmpty: false,
        })
        this.currentDiff = diffHtml
        this.$nextTick(() => {
          const container = this.$refs.diffContainer
          if (container) {
            container.innerHTML = diffHtml
            const codeLines = container.querySelectorAll('.d2h-code-line')
            codeLines.forEach(line => {
              const contentCells = line.querySelectorAll('.d2h-code-line-ctn')
              contentCells.forEach(cell => {
                cell.innerHTML = cell.textContent.replace(/\n/g, '<br>')
              })
            })
          }
        })
      } catch (e) {
        this.diffError = '渲染 diff 失败: ' + (e.message || String(e))
      }
    },
    renderCurrentDiff() {
      if (this.selectedFile && this.diffs && this.diffs[this.selectedFile]) {
        this.renderDiff(this.diffs[this.selectedFile])
      }
    },
  },
}
</script>

<style scoped>
/* ===== 文件变更弹窗样式 ===== */
.file-changes-detail {
  height: 75vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.file-changes-detail__toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0 12px;
  border-bottom: 1px solid #e8ecf1;
  margin-bottom: 8px;
  flex-shrink: 0;
  gap: 12px;
}

.file-changes-detail__info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.file-changes-detail__dir {
  font-weight: 600;
  font-size: 13px;
  color: #303133;
  max-width: 350px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-changes-detail__branch {
  font-size: 12px;
  color: #909399;
}

.file-changes-detail__summary {
  display: flex;
  gap: 4px;
}

.file-changes-detail__body {
  flex: 1;
  display: flex;
  gap: 0;
  overflow: hidden;
  min-height: 0;
}

.file-changes-detail__tree-panel {
  width: 280px;
  min-width: 200px;
  border: 1px solid #e8ecf1;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  margin-right: 8px;
}

.file-changes-detail__tree-title {
  font-size: 12px;
  font-weight: 600;
  color: #606266;
  padding: 10px 12px;
  border-bottom: 1px solid #e8ecf1;
  background: #f6f8fa;
  flex-shrink: 0;
}

.file-changes-detail__tree {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.file-changes-detail__tree-empty {
  padding: 20px;
  text-align: center;
  color: #c0c4cc;
  font-size: 13px;
}

.file-changes-tree__dir {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  cursor: pointer;
  font-size: 13px;
  color: #0366d6;
  user-select: none;
}

.file-changes-tree__dir:hover {
  background: #f0f7ff;
}

.file-changes-tree__dir-icon {
  font-size: 10px;
  width: 14px;
  text-align: center;
}

.file-changes-tree__dir-name {
  font-weight: 500;
}

.file-changes-tree__dir-count {
  font-size: 11px;
  color: #8b949e;
}

.file-changes-tree__file {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 12px 3px 32px;
  cursor: pointer;
  font-size: 12px;
  color: #24292e;
  transition: background 0.15s;
}

.file-changes-tree__file:hover {
  background: #f0f7ff;
}

.file-changes-tree__file--active {
  background: #ddf4ff !important;
  font-weight: 600;
}

.file-changes-tree__file-type {
  display: inline-block;
  font-size: 10px;
  font-family: monospace;
  padding: 0 4px;
  border-radius: 3px;
  min-width: 24px;
  text-align: center;
  line-height: 18px;
}

.file-changes-tree__file-type--added {
  background: #dafbe1;
  color: #1a7f37;
}

.file-changes-tree__file-type--modified {
  background: #fff8c5;
  color: #9a6700;
}

.file-changes-tree__file-type--deleted {
  background: #ffebe9;
  color: #cf222e;
}

.file-changes-tree__file-type--untracked {
  background: #f3f4f6;
  color: #656d76;
}

.file-changes-tree__file-type--renamed {
  background: #f0f7ff;
  color: #0366d6;
}

.file-changes-tree__file-type--other {
  background: #f3f4f6;
  color: #656d76;
}

.file-changes-tree__file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-changes-detail__diff-panel {
  flex: 1;
  border: 1px solid #e8ecf1;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.file-changes-detail__diff-header {
  padding: 8px 12px;
  border-bottom: 1px solid #e8ecf1;
  background: #f6f8fa;
  flex-shrink: 0;
}

.file-changes-detail__diff-file {
  font-size: 13px;
  font-family: monospace;
  color: #24292e;
}

.file-changes-detail__diff-content {
  flex: 1;
  overflow: auto;
  padding: 8px;
}

.file-changes-detail__diff-placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #c0c4cc;
  font-size: 14px;
}

.file-changes-detail__diff-error {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #e53935;
  font-size: 13px;
  padding: 20px;
  text-align: center;
}

.file-changes-detail__footer {
  text-align: right;
}

/* diff2html 适配 */
.file-changes-detail__diff-content :deep(.d2h-code-line-ctn) {
  white-space: pre-wrap;
  word-break: break-all;
}

.file-changes-detail__diff-content :deep(.d2h-file-side-diff) {
  display: flex;
}

.file-changes-detail__diff-content :deep(.d2h-code-line) {
  display: flex;
}

.file-changes-detail__diff-content :deep(.d2h-code-side-linenumber) {
  width: 40px;
  min-width: 40px;
  padding: 0 10px;
  background-color: #f6f8fa;
  color: rgba(27, 31, 35, 0.3);
  text-align: right;
  user-select: none;
}
</style>
