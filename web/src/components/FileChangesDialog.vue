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
              <span class="file-changes-stat file-changes-stat--committed">{{ summary.committed || 0 }} C</span>
              <span class="file-changes-stat file-changes-stat--staged">{{ summary.staged || 0 }} S</span>
              <span class="file-changes-stat file-changes-stat--modified">{{ summary.modified || 0 }} M</span>
              <span class="file-changes-stat file-changes-stat--untracked">{{ summary.untracked || 0 }} U</span>
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
          <div class="file-changes-detail__tree-title">改动文件列表 ({{ files.length }})</div>
          <div class="file-changes-detail__tree">
            <template v-if="flatTreeItems.length > 0">
              <div
                v-for="item in flatTreeItems"
                :key="item.key"
                :class="item.isDir ? 'file-changes-tree__dir' : 'file-changes-tree__file'"
                :style="{ paddingLeft: (12 + item.depth * 16) + 'px' }"
                @click="item.isDir ? toggleDir(item.key) : selectFile(item)"
              >
                <template v-if="item.isDir">
                  <span class="file-changes-tree__dir-icon">{{ expandedDirs[item.key] !== false ? '▼' : '▶' }}</span>
                  <span class="file-changes-tree__dir-name">{{ item.name }}/</span>
                  <span class="file-changes-tree__dir-count">({{ item.fileCount }})</span>
                </template>
                <template v-else>
                  <span class="file-changes-tree__file-type" :class="'file-changes-tree__file-type--' + item.type">{{ item.status_code }}</span>
                  <span class="file-changes-tree__file-name" :title="item.path">{{ item.name }}</span>
                </template>
              </div>
            </template>
            <div v-else class="file-changes-detail__tree-empty">暂无文件变更</div>
          </div>
        </div>
        <!-- 右侧 diff 视图 -->
        <div class="file-changes-detail__diff-panel">
          <template v-if="diffLoading">
            <div class="file-changes-detail__diff-placeholder">
              <span>加载 diff 中...</span>
            </div>
          </template>
          <template v-else-if="diffError">
            <div class="file-changes-detail__diff-error">{{ diffError }}</div>
          </template>
          <template v-else-if="selectedFile">
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

// CodeMirror MergeView
import CodeMirror from 'codemirror'
import 'codemirror/lib/codemirror.css'
import 'codemirror/addon/merge/merge.js'
import 'codemirror/addon/merge/merge.css'
import 'codemirror/addon/fold/foldgutter.css'
import 'codemirror/addon/fold/foldcode.js'
import 'codemirror/addon/fold/foldgutter.js'

// 语法高亮模式（按需导入）
import 'codemirror/mode/javascript/javascript.js'
import 'codemirror/mode/go/go.js'
import 'codemirror/mode/python/python.js'
import 'codemirror/mode/css/css.js'
import 'codemirror/mode/htmlmixed/htmlmixed.js'
import 'codemirror/mode/xml/xml.js'
import 'codemirror/mode/sql/sql.js'
import 'codemirror/mode/markdown/markdown.js'
import 'codemirror/mode/shell/shell.js'
import 'codemirror/mode/yaml/yaml.js'
import 'codemirror/mode/vue/vue.js'
import 'codemirror/mode/php/php.js'
import 'codemirror/mode/ruby/ruby.js'
import 'codemirror/mode/clike/clike.js'
import 'codemirror/mode/nginx/nginx.js'
import 'codemirror/mode/dockerfile/dockerfile.js'
import 'codemirror/mode/toml/toml.js'
import 'codemirror/mode/jsx/jsx.js'
import 'codemirror/mode/sass/sass.js'
import 'codemirror/mode/powershell/powershell.js'

// diff-match-patch（CodeMirror MergeView 依赖）
import DiffMatchPatch from 'diff-match-patch'
// CodeMirror merge addon 直接引用全局常量，需手动挂载
window.diff_match_patch = DiffMatchPatch
window.DIFF_EQUAL = DiffMatchPatch.DIFF_EQUAL
window.DIFF_INSERT = DiffMatchPatch.DIFF_INSERT
window.DIFF_DELETE = DiffMatchPatch.DIFF_DELETE

// 文件扩展名 → CodeMirror mode 映射
const EXT_MODE_MAP = {
  '.js': 'javascript',
  '.mjs': 'javascript',
  '.cjs': 'javascript',
  '.jsx': 'jsx',
  '.ts': 'javascript',
  '.tsx': 'jsx',
  '.vue': 'vue',
  '.go': 'go',
  '.py': 'python',
  '.css': 'css',
  '.scss': 'sass',
  '.less': 'css',
  '.html': 'htmlmixed',
  '.htm': 'htmlmixed',
  '.xml': 'xml',
  '.svg': 'xml',
  '.sql': 'sql',
  '.md': 'markdown',
  '.markdown': 'markdown',
  '.sh': 'shell',
  '.bash': 'shell',
  '.zsh': 'shell',
  '.yml': 'yaml',
  '.yaml': 'yaml',
  '.php': 'php',
  '.rb': 'ruby',
  '.java': 'text/x-java',
  '.c': 'text/x-csrc',
  '.cpp': 'text/x-c++src',
  '.h': 'text/x-csrc',
  '.hpp': 'text/x-c++src',
  '.cs': 'text/x-csharp',
  '.json': { name: 'javascript', json: true },
  '.toml': 'toml',
  '.ini': 'toml',
  '.cfg': 'toml',
  '.conf': 'nginx',
  '.dockerfile': 'dockerfile',
  '.ps1': 'shell',
  '.bat': 'shell',
  '.cmd': 'shell',
}

function getModeForFile(filePath) {
  if (!filePath) return 'text'
  const fileName = filePath.split('/').pop() || filePath
  // 特殊文件名
  if (fileName === 'Dockerfile' || fileName.startsWith('Dockerfile.')) return 'dockerfile'
  if (fileName === 'Makefile' || fileName === 'Taskfile.yml') return 'yaml'
  if (fileName === '.gitignore' || fileName === '.dockerignore' || fileName === '.eslintignore') return 'shell'
  const dotIdx = fileName.lastIndexOf('.')
  if (dotIdx < 0) return 'text'
  const ext = fileName.substring(dotIdx).toLowerCase()
  return EXT_MODE_MAP[ext] || 'text'
}

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
      selectedFile: '',
      currentDiff: '',
      oldContent: '',
      newContent: '',
      diffError: '',
      diffLoading: false,
      diffMode: 'side-by-side',
      expandedDirs: {},
      mergeView: null,
    }
  },
  computed: {
    treeNodes() {
      return this.buildTree(this.files)
    },
    flatTreeItems() {
      const result = []
      const flatten = (nodes, depth) => {
        for (const node of nodes) {
          if (node.isDir) {
            result.push({
              isDir: true,
              name: node.name,
              key: node.key,
              path: node.path,
              fileCount: node.fileCount,
              depth,
            })
            if (this.expandedDirs[node.key] !== false) {
              flatten(node.children, depth + 1)
              for (const file of node.files) {
                result.push({
                  isDir: false,
                  name: file.name,
                  key: 'file_' + file.path,
                  path: file.path,
                  status_code: file.status_code,
                  type: file.type,
                  depth: depth + 1,
                })
              }
            }
          } else {
            result.push({
              isDir: false,
              name: node.name,
              key: node.key,
              path: node.path,
              status_code: node.status_code,
              type: node.type,
              depth,
            })
          }
        }
      }
      flatten(this.treeNodes, 0)
      return result
    },
  },
  watch: {
    visible(val) {
      if (val) {
        this.initData()
      } else {
        this.destroyMergeView()
      }
    },
  },
  beforeUnmount() {
    this.destroyMergeView()
  },
  methods: {
    initData() {
      this.summary = this.initialSummary ? { ...this.initialSummary } : null
      this.files = Array.isArray(this.initialFiles) ? [...this.initialFiles] : []
      this.selectedFile = ''
      this.currentDiff = ''
      this.oldContent = ''
      this.newContent = ''
      this.diffError = ''
      this.diffLoading = false
      this.expandedDirs = {}
      this.loadFileList()
    },
    loadFileList() {
      if (!this.localDir) return
      this.loading = true
      taskWorkflowApi.TaskWorkflowFileChangesDetail(this.localDir, this.parentBranch, (response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data) {
          if (response.Data.summary) {
            this.summary = response.Data.summary
          }
          if (Array.isArray(response.Data.files)) {
            this.files = response.Data.files
          }
          if (response.Data.error) {
            this.diffError = response.Data.error
          }
        }
      })
    },
    buildTree(files) {
      if (!files || files.length === 0) return []

      const dirMap = {}

      const ensureDir = (dirPath) => {
        if (dirMap[dirPath]) return dirMap[dirPath]
        const parts = dirPath.split('/')
        const name = parts[parts.length - 1]
        const parentPath = parts.slice(0, -1).join('/')
        const dirNode = {
          isDir: true,
          name,
          key: 'dir_' + dirPath,
          path: dirPath,
          files: [],
          children: [],
          fileCount: 0,
        }
        dirMap[dirPath] = dirNode
        if (parentPath) {
          ensureDir(parentPath).children.push(dirNode)
        }
        return dirNode
      }

      const rootFiles = []

      for (const file of files) {
        const path = file.path || ''
        const slashIdx = path.lastIndexOf('/')
        const dirPath = slashIdx >= 0 ? path.substring(0, slashIdx) : ''
        const fileName = slashIdx >= 0 ? path.substring(slashIdx + 1) : path
        const fileObj = { ...file, name: fileName }

        if (dirPath === '') {
          rootFiles.push({ ...fileObj, isDir: false, key: 'file_' + file.path })
        } else {
          ensureDir(dirPath)
          dirMap[dirPath].files.push(fileObj)
        }
      }

      const topLevelDirPaths = new Set()
      for (const dirPath of Object.keys(dirMap)) {
        const parts = dirPath.split('/')
        if (parts.length === 1) {
          topLevelDirPaths.add(dirPath)
        }
      }

      const computeFileCount = (dirNode) => {
        let count = dirNode.files.length
        for (const child of dirNode.children) {
          count += computeFileCount(child)
        }
        dirNode.fileCount = count
        return count
      }

      const sortDir = (dirNode) => {
        dirNode.children.sort((a, b) => a.name.localeCompare(b.name))
        dirNode.files.sort((a, b) => a.name.localeCompare(b.name))
        for (const child of dirNode.children) {
          sortDir(child)
        }
      }

      const topDirs = []
      for (const dirPath of topLevelDirPaths) {
        topDirs.push(dirMap[dirPath])
      }

      for (const dir of topDirs) {
        computeFileCount(dir)
        sortDir(dir)
      }

      const result = []
      for (const dir of topDirs.sort((a, b) => a.name.localeCompare(b.name))) {
        result.push(dir)
      }
      for (const file of rootFiles.sort((a, b) => a.name.localeCompare(b.name))) {
        result.push(file)
      }

      return result
    },
    toggleDir(key) {
      this.expandedDirs = {
        ...this.expandedDirs,
        [key]: this.expandedDirs[key] !== false ? false : true,
      }
    },
    selectFile(item) {
      this.selectedFile = item.path
      this.diffError = ''
      this.loadFileDiff(item)
    },
    loadFileDiff(file) {
      if (!this.parentBranch) {
        this.diffError = '未指定基分支，无法获取 diff。'
        return
      }
      this.diffLoading = true
      this.currentDiff = ''
      this.oldContent = ''
      this.newContent = ''
      this.destroyMergeView()
      taskWorkflowApi.TaskWorkflowFileChangesFileDiff(this.localDir, this.parentBranch, file.path, (response) => {
        this.diffLoading = false
        if (response && response.ErrCode === 0 && response.Data) {
          const data = response.Data
          // 优先使用 old_content/new_content（CodeMirror MergeView）
          if (data.old_content !== undefined && data.new_content !== undefined) {
            this.oldContent = data.old_content || ''
            this.newContent = data.new_content || ''
          }
          if (data.diff) {
            this.currentDiff = data.diff
          }
          // 如果没有 old/new 内容但有 diff 文本，降级到 diff2html
          if (!this.oldContent && !this.newContent && this.currentDiff) {
            this.renderDiff2Html(this.currentDiff)
          } else {
            this.renderDiff()
          }
        } else {
          if (file.type === 'untracked') {
            this.diffError = '文件 "' + file.path + '" 是未跟踪的新文件，暂无法对比差异。'
          } else {
            this.diffError = (response && response.ErrMsg) || '获取 diff 失败'
          }
        }
      })
    },
    handleClose() {
      this.$emit('update:visible', false)
      this.selectedFile = ''
      this.currentDiff = ''
      this.oldContent = ''
      this.newContent = ''
      this.destroyMergeView()
    },
    destroyMergeView() {
      if (this.mergeView) {
        try {
          this.mergeView.editor().toTextArea?.()
        } catch (e) { /* ignore */ }
        this.mergeView = null
      }
      const container = this.$refs.diffContainer
      if (container) {
        container.innerHTML = ''
      }
    },
    renderDiff() {
      this.$nextTick(() => {
        const container = this.$refs.diffContainer
        if (!container) return

        this.destroyMergeView()
        container.innerHTML = ''

        const mode = getModeForFile(this.selectedFile)

        if (this.diffMode === 'side-by-side') {
          // 横向对比：使用 CodeMirror MergeView（左右双面板 + 中间连接线）
          this.mergeView = CodeMirror.MergeView(container, {
            value: this.newContent || '',
            orig: this.oldContent || '',
            lineNumbers: true,
            mode: mode,
            theme: 'default',
            collapseIdentical: true,
            revertButtons: false,
            readOnly: true,
            lineWrapping: false,
            scrollbarStyle: 'native',
          })
          // 注入箭头到连接线上
          this.$nextTick(() => this.addConnectArrows())
          // 监听滚动和更新事件，重绘箭头
          const editor = this.mergeView.editor()
          const orig = this.mergeView.leftOriginal()
          if (editor) editor.on('scroll', () => this.scheduleRedrawArrows())
          if (orig) orig.on('scroll', () => this.scheduleRedrawArrows())
        } else {
          // 纵向对比：使用 CodeMirror 单面板 + diff 高亮
          this.renderInlineDiff(container, mode)
        }
      })
    },
    renderInlineDiff(container, mode) {
      // 纵向对比：交替显示删除行（红底）和新增行（绿底）
      const Diff = require('diff')
      const changes = Diff.diffLines(this.oldContent || '', this.newContent || '')

      // 构建带类型标注的行列表
      let annotatedLines = []
      for (const change of changes) {
        const text = change.value.replace(/\n$/, '')
        const lines = text.split('\n')
        for (let i = 0; i < lines.length; i++) {
          // 跳过末尾空行（diffLines 会多出一个换行）
          if (i === lines.length - 1 && lines[i] === '' && change.value.endsWith('\n')) continue
          let type = 'unchanged'
          if (change.added) type = 'added'
          else if (change.removed) type = 'removed'
          annotatedLines.push({ text: lines[i], type })
        }
      }

      // 构建 CodeMirror 显示内容（删除行前加 - 前缀，新增行前加 + 前缀）
      const displayText = annotatedLines.map(l => {
        if (l.type === 'removed') return '- ' + l.text
        if (l.type === 'added') return '+ ' + l.text
        return '  ' + l.text
      }).join('\n')

      const editor = CodeMirror(container, {
        value: displayText,
        lineNumbers: true,
        mode: mode,
        readOnly: true,
        lineWrapping: false,
        gutters: ['CodeMirror-linenumbers', 'diff-gutter'],
      })

      // 标记每行的背景色和 gutter 符号
      for (let i = 0; i < annotatedLines.length; i++) {
        const line = annotatedLines[i]
        if (line.type === 'added') {
          editor.addLineClass(i, 'background', 'diff-line-added')
          editor.setGutterMarker(i, 'diff-gutter', (() => {
            const el = document.createElement('div')
            el.className = 'diff-gutter-marker diff-gutter-marker--add'
            el.textContent = '+'
            return el
          })())
        } else if (line.type === 'removed') {
          editor.addLineClass(i, 'background', 'diff-line-removed')
          editor.setGutterMarker(i, 'diff-gutter', (() => {
            const el = document.createElement('div')
            el.className = 'diff-gutter-marker diff-gutter-marker--del'
            el.textContent = '-'
            return el
          })())
        }
      }

      editor.setSize('100%', '100%')
    },
    renderDiff2Html(diffText) {
      this.destroyMergeView()
      try {
        const diffHtml = Diff2Html.html(diffText, {
          drawFileList: false,
          matching: 'lines',
          outputFormat: 'line-by-line',
          renderNothingWhenEmpty: false,
        })
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
      } catch (e) {
        this.diffError = '渲染 diff 失败: ' + (e.message || String(e))
      }
    },
    renderDiff2HtmlFromContent(oldText, newText) {
      // 使用 diff 库生成 unified diff，再用 diff2html 渲染
      try {
        const Diff = require('diff')
        const patch = Diff.createPatch(this.selectedFile, oldText || '', newText || '', '旧版本', '新版本')
        if (patch && patch.trim()) {
          this.renderDiff2Html(patch)
        } else {
          const container = this.$refs.diffContainer
          if (container) {
            container.innerHTML = '<div style="padding:20px;color:#909399;text-align:center;">文件内容无差异</div>'
          }
        }
      } catch (e) {
        this.diffError = '渲染 diff 失败: ' + (e.message || String(e))
      }
    },
    renderCurrentDiff() {
      if (this.selectedFile && (this.oldContent || this.newContent || this.currentDiff)) {
        this.renderDiff()
      }
    },
    // 箭头重绘节流
    scheduleRedrawArrows() {
      if (this._arrowTimer) return
      this._arrowTimer = setTimeout(() => {
        this._arrowTimer = null
        this.addConnectArrows()
      }, 80)
    },
    // 在 MergeView 的 SVG 连接线上注入方向箭头
    addConnectArrows() {
      const container = this.$refs.diffContainer
      if (!container) return
      const svg = container.querySelector('.CodeMirror-merge-gap svg')
      if (!svg) return

      const svgNS = 'http://www.w3.org/2000/svg'
      // 清除旧箭头
      svg.querySelectorAll('.diff-arrow').forEach(el => el.remove())

      // 注入 marker 定义（只需一次）
      if (!svg.querySelector('#diff-arrow-marker')) {
        const defs = svg.querySelector('defs') || svg.insertBefore(document.createElementNS(svgNS, 'defs'), svg.firstChild)
        const marker = document.createElementNS(svgNS, 'marker')
        marker.setAttribute('id', 'diff-arrow-marker')
        marker.setAttribute('viewBox', '0 0 10 10')
        marker.setAttribute('refX', '5')
        marker.setAttribute('refY', '5')
        marker.setAttribute('markerWidth', '6')
        marker.setAttribute('markerHeight', '6')
        marker.setAttribute('orient', 'auto-start-reverse')
        const polygon = document.createElementNS(svgNS, 'polygon')
        polygon.setAttribute('points', '0,0 10,5 0,10')
        polygon.setAttribute('fill', '#0366d6')
        marker.appendChild(polygon)
        defs.appendChild(marker)
      }

      // 为每个连接路径添加箭头线
      const paths = svg.querySelectorAll('path')
      paths.forEach(path => {
        const d = path.getAttribute('d')
        if (!d) return

        // 解析路径提取关键坐标点
        // 路径格式：M -1 topRpx C... L (w+2) botLpx C... z
        // 顶部连线中点和底部连线中点
        const coords = []
        const re = /[-+]?[\d.]+/g
        let match
        while ((match = re.exec(d)) !== null) {
          coords.push(parseFloat(match[0]))
        }
        // coords: [x1, topR, cx1, cy1, cx2, cy2, x2, topL, x3, botL, cx3, cy3, cx4, cy4, x4, botR]
        if (coords.length < 16) return

        const w = coords[6] // 右侧 x (约 w+2)
        const topRightY = coords[1]
        const topLeftY = coords[7]
        const botLeftY = coords[9]
        const botRightY = coords[15]
        const midX = w / 2

        // 顶部连线中点（从右到左）
        const topMidY = (topRightY + topLeftY) / 2
        // 底部连线中点（从左到右）
        const botMidY = (botLeftY + botRightY) / 2

        // 绘制箭头线：从右（old）到左（new）方向，在顶部连线中点
        const arrowLine1 = document.createElementNS(svgNS, 'line')
        arrowLine1.setAttribute('x1', String(w * 0.8))
        arrowLine1.setAttribute('y1', String(topMidY))
        arrowLine1.setAttribute('x2', String(w * 0.2))
        arrowLine1.setAttribute('y2', String(topMidY))
        arrowLine1.setAttribute('stroke', '#0366d6')
        arrowLine1.setAttribute('stroke-width', '2')
        arrowLine1.setAttribute('marker-end', 'url(#diff-arrow-marker)')
        arrowLine1.setAttribute('class', 'diff-arrow')
        svg.appendChild(arrowLine1)

        // 绘制箭头线：从左（new）到右（old）方向，在底部连线中点
        const arrowLine2 = document.createElementNS(svgNS, 'line')
        arrowLine2.setAttribute('x1', String(w * 0.2))
        arrowLine2.setAttribute('y1', String(botMidY))
        arrowLine2.setAttribute('x2', String(w * 0.8))
        arrowLine2.setAttribute('y2', String(botMidY))
        arrowLine2.setAttribute('stroke', '#0366d6')
        arrowLine2.setAttribute('stroke-width', '2')
        arrowLine2.setAttribute('marker-end', 'url(#diff-arrow-marker)')
        arrowLine2.setAttribute('class', 'diff-arrow')
        svg.appendChild(arrowLine2)
      })
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
  gap: 6px;
}

.file-changes-stat {
  font-size: 12px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 3px;
}

.file-changes-stat--committed {
  background: #dafbe1;
  color: #1a7f37;
}

.file-changes-stat--staged {
  background: #f0f7ff;
  color: #0366d6;
}

.file-changes-stat--modified {
  background: #fff8c5;
  color: #9a6700;
}

.file-changes-stat--untracked {
  background: #f3f4f6;
  color: #656d76;
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
  flex-shrink: 0;
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
  min-width: 20px;
  text-align: center;
  line-height: 18px;
}

.file-changes-tree__file-type--committed {
  background: #dafbe1;
  color: #1a7f37;
}

.file-changes-tree__file-type--staged {
  background: #f0f7ff;
  color: #0366d6;
}

.file-changes-tree__file-type--modified {
  background: #fff8c5;
  color: #9a6700;
}

.file-changes-tree__file-type--untracked {
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
  overflow: hidden;
  min-height: 0;
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

/* ===== CodeMirror MergeView 适配样式 ===== */
.file-changes-detail__diff-content :deep(.CodeMirror) {
  height: 100%;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 13px;
  line-height: 1.5;
}

.file-changes-detail__diff-content :deep(.CodeMirror-merge) {
  height: 100%;
}

.file-changes-detail__diff-content :deep(.CodeMirror-merge-pane) {
  height: 100%;
}

/* 改动行背景色 */
.file-changes-detail__diff-content :deep(.CodeMirror-merge-r-deleted),
.file-changes-detail__diff-content :deep(.CodeMirror-merge-l-deleted) {
  background-color: #fecdd3;
}

.file-changes-detail__diff-content :deep(.CodeMirror-merge-r-inserted),
.file-changes-detail__diff-content :deep(.CodeMirror-merge-l-inserted) {
  background-color: #bbf7d0;
}

/* 中间连接线区域 */
.file-changes-detail__diff-content :deep(.CodeMirror-merge-gap) {
  background: #f6f8fa;
  border-left: 1px solid #d0d7de;
  border-right: 1px solid #d0d7de;
  z-index: 2;
  min-width: 50px;
}

/* 连接线路径 — 蓝色填充 + 描边 */
.file-changes-detail__diff-content :deep(.CodeMirror-merge-l-connect),
.file-changes-detail__diff-content :deep(.CodeMirror-merge-r-connect) {
  fill: rgba(3, 102, 214, 0.12);
  stroke: rgba(3, 102, 214, 0.4);
  stroke-width: 1;
  stroke-linejoin: round;
}


.file-changes-detail__diff-content :deep(.CodeMirror-gutters) {
  background: #f6f8fa;
  border-right: 1px solid #e8ecf1;
}

/* ===== 纵向对比 CodeMirror 行标记样式 ===== */
.file-changes-detail__diff-content :deep(.diff-line-added) {
  background-color: #bbf7d0 !important;
}

.file-changes-detail__diff-content :deep(.diff-line-removed) {
  background-color: #fecdd3 !important;
}

.file-changes-detail__diff-content :deep(.diff-gutter-marker) {
  font-weight: 700;
  font-size: 14px;
  width: 16px;
  text-align: center;
  line-height: 1;
  margin-top: 2px;
}

.file-changes-detail__diff-content :deep(.diff-gutter-marker--add) {
  color: #1a7f37;
}

.file-changes-detail__diff-content :deep(.diff-gutter-marker--del) {
  color: #cf222e;
}

/* diff2html 纵向对比适配 */
.file-changes-detail__diff-content :deep(.d2h-wrapper) {
  height: 100%;
  overflow: auto;
}

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
