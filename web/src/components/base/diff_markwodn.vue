<template>
  <div class="code-diff-container">
    <div class="diff-header">
      <div class="file-title">{{ title }}</div>
      <div class="view-toggle">
        <pl-button
            :class="{ active: viewMode === 'diff' }"
            link
            @click="viewMode = 'diff'"
        >
          差异对比
        </pl-button>
        <pl-button
            :class="{ active: viewMode === 'rendered' }"
            link
            @click="viewMode = 'rendered'"
        >
          完整渲染
        </pl-button>
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

<style scoped src="@/css/components/base/diff_markwodn.css"></style>
