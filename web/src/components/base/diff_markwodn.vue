<template>
  <div class="code-diff-container">
    <div class="diff-header">
      <div class="file-title">{{ title }}</div>
      <div class="view-toggle">
        <button
            :class="{ active: viewMode === 'diff' }"
            @click="viewMode = 'diff'"
        >
          差异对比
        </button>
        <button
            :class="{ active: viewMode === 'rendered' }"
            @click="viewMode = 'rendered'"
        >
          完整渲染
        </button>
      </div>
    </div>
    <div v-show="viewMode === 'diff'" ref="diffContainer" class="diff-content"></div>
    <div v-show="viewMode === 'rendered'" class="markdown-rendered-view">
      <div class="rendered-column old-content" v-html="renderedOldText"></div>
      <div class="rendered-column new-content" v-html="renderedNewText"></div>
    </div>
  </div>
</template>

<script>
import { onMounted, ref, watch, computed } from 'vue'
import * as Diff2Html from 'diff2html'
import * as Diff from 'diff'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import 'diff2html/bundles/css/diff2html.min.css'

export default {
  name: 'SplitViewCodeDiff',
  props: {
    oldText: {
      type: String,
      default: ''
    },
    newText: {
      type: String,
      default: ''
    },
    title: {
      type: String,
      default: '代码对比'
    }
  },
  setup(props) {
    const diffContainer = ref(null)
    const viewMode = ref('diff') // 'diff' 或 'rendered'

    // 配置marked解析器
    marked.setOptions({
      breaks: true,
      gfm: true,
      highlight: null
    })

    // 处理文本中的换行符
    const normalizeLineEndings = (text) => {
      return text.replace(/\r\n/g, '\n').replace(/\r/g, '\n')
    }

    // 生成差异对比的HTML
    const generateDiffHtml = () => {
      const normalizedOld = normalizeLineEndings(props.oldText)
      const normalizedNew = normalizeLineEndings(props.newText)

      const diff = Diff.createPatch('file.md', normalizedOld, normalizedNew, '旧版本', '新版本')
      const diffHtml = Diff2Html.html(diff, {
        drawFileList: false,
        matching: 'lines',
        outputFormat: 'side-by-side',
        renderNothingWhenEmpty: false
      })
      return diffHtml
    }

    // 更新差异显示
    const updateDiffView = () => {
      if (diffContainer.value) {
        const html = generateDiffHtml()
        diffContainer.value.innerHTML = html

        // 修复换行显示
        setTimeout(() => {
          const codeLines = diffContainer.value.querySelectorAll('.d2h-code-line')
          codeLines.forEach(line => {
            const contentCells = line.querySelectorAll('.d2h-code-line-ctn')
            contentCells.forEach(cell => {
              cell.innerHTML = cell.textContent.replace(/\n/g, '<br>')
            })
          })
        }, 0)
      }
    }

    // 计算渲染后的Markdown
    const renderedOldText = computed(() => {
      return DOMPurify.sanitize(marked(props.oldText || '无内容'))
    })

    const renderedNewText = computed(() => {
      return DOMPurify.sanitize(marked(props.newText || '无内容'))
    })

    // 监听变化
    watch([() => props.oldText, () => props.newText], () => {
      if (viewMode.value === 'diff') {
        updateDiffView()
      }
    }, { immediate: true })

    // 监听视图模式变化
    watch(viewMode, (newVal) => {
      if (newVal === 'diff') {
        updateDiffView()
      }
    })

    // 初始化
    onMounted(() => {
      updateDiffView()
    })

    return {
      diffContainer,
      viewMode,
      renderedOldText,
      renderedNewText
    }
  }
}
</script>

<style scoped>
/* 样式保持不变，与之前相同 */
.code-diff-container {
  border: 1px solid #e1e4e8;
  border-radius: 6px;
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
  background-color: #fff;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.diff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background-color: #f6f8fa;
  border-bottom: 1px solid #e1e4e8;
}

.file-title {
  font-weight: 600;
  font-size: 14px;
  color: #24292e;
}

.view-toggle {
  display: flex;
  gap: 8px;
}

.view-toggle button {
  padding: 4px 8px;
  font-size: 12px;
  background: #f3f4f6;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  cursor: pointer;
}

.view-toggle button.active {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.diff-content {
  flex: 1;
  overflow: auto;
}

.markdown-rendered-view {
  flex: 1;
  display: flex;
  overflow: auto;
}

.rendered-column {
  flex: 1;
  padding: 16px;
  overflow: auto;
  box-sizing: border-box;
}

.rendered-column.old-content {
  border-right: 1px solid #e1e4e8;
  background-color: #fff8f8;
}

.rendered-column.new-content {
  background-color: #f8fff8;
}

/* 增强换行显示 */
:deep(.d2h-code-line-ctn) {
  white-space: pre-wrap;
  word-break: break-all;
}

/* 强制左右分屏样式 */
:deep(.d2h-file-side-diff) {
  display: flex;
}

:deep(.d2h-code-line) {
  display: flex;
}

:deep(.d2h-code-side-linenumber) {
  width: 40px;
  min-width: 40px;
  padding: 0 10px;
  background-color: #f6f8fa;
  color: rgba(27, 31, 35, 0.3);
  text-align: right;
  user-select: none;
}

/* 差异高亮样式 */
:deep(.d2h-del) {
  background-color: #ffebe9;
}

:deep(.d2h-ins) {
  background-color: #e6ffec;
}

:deep(.d2h-info) {
  background-color: #f1f8ff;
  color: #0366d6;
}

/* Markdown渲染样式 */
:deep(.rendered-column) h1,
:deep(.rendered-column) h2,
:deep(.rendered-column) h3,
:deep(.rendered-column) h4,
:deep(.rendered-column) h5,
:deep(.rendered-column) h6 {
  margin-top: 0;
  margin-bottom: 16px;
  font-weight: 600;
  line-height: 1.25;
}

:deep(.rendered-column) p {
  margin-top: 0;
  margin-bottom: 16px;
}

:deep(.rendered-column) pre {
  background-color: #f6f8fa;
  border-radius: 6px;
  padding: 16px;
  overflow: auto;
}

:deep(.rendered-column) code {
  font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
  font-size: 85%;
  background-color: rgba(27, 31, 35, 0.05);
  border-radius: 3px;
  padding: 0.2em 0.4em;
}

:deep(.rendered-column) blockquote {
  border-left: 4px solid #dfe2e5;
  color: #6a737d;
  padding: 0 1em;
  margin: 0 0 16px 0;
}

:deep(.rendered-column) table {
  border-collapse: collapse;
  width: 100%;
  margin-bottom: 16px;
}

:deep(.rendered-column) th,
:deep(.rendered-column) td {
  border: 1px solid #dfe2e5;
  padding: 6px 13px;
}

:deep(.rendered-column) tr {
  background-color: #fff;
  border-top: 1px solid #c6cbd1;
}

:deep(.rendered-column) tr:nth-child(2n) {
  background-color: #f6f8fa;
}
</style>