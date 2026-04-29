<template>
  <div class="code-diff-container">
    <div class="diff-header">
      <div class="file-title">{{ title }}</div>
    </div>
    <div ref="diffContainer" class="diff-content"></div>
  </div>
</template>

<script>
import { onMounted, ref, watch } from 'vue'
import * as Diff2Html from 'diff2html'
import * as Diff from 'diff'
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

    // 处理文本中的换行符
    const normalizeLineEndings = (text) => {
      return text.replace(/\r\n/g, '\n').replace(/\r/g, '\n')
    }

    // 生成差异对比的HTML（固定使用split视图）
    const generateDiffHtml = () => {
      const normalizedOld = normalizeLineEndings(props.oldText)
      const normalizedNew = normalizeLineEndings(props.newText)

      const diff = Diff.createPatch('file.txt', normalizedOld, normalizedNew, '旧版本', '新版本')
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

    // 监听变化
    watch([() => props.oldText, () => props.newText], () => {
      updateDiffView()
    }, { immediate: true })

    // 初始化
    onMounted(() => {
      updateDiffView()
    })

    return {
      diffContainer
    }
  }
}
</script>

<style scoped src="@/css/components/base/diff_code.css"></style>