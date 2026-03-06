<template>
  <div ref="markdownContainer" class="markdown-container" v-html="compiledMarkdown"></div>
</template>

<script>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue';
import MarkdownIt from 'markdown-it';

export default {
  name: "MarkdownRenderer",
  props: {
    source: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const md = new MarkdownIt({
      html: true,
      breaks: true,
      linkify: true,
      typographer: true,
    });

    const compiledMarkdown = computed(() => md.render(props.source));
    const markdownContainer = ref(null);

    // 是否允许自动滚动的状态
    const allowAutoScroll = ref(true);

    // 后备复制方法（用于不支持 Clipboard API 的情况）
    const fallbackCopyText = (text, copyBtn) => {
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      textArea.style.top = '-999999px';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();

      try {
        const successful = document.execCommand('copy');
        if (successful) {
          copyBtn.textContent = '已复制!';
          setTimeout(() => {
            copyBtn.textContent = '复制';
          }, 2000);
        } else {
          copyBtn.textContent = '复制失败';
        }
      } catch (err) {
        console.error('后备复制失败:', err);
        copyBtn.textContent = '复制失败';
      }

      document.body.removeChild(textArea);
    };

    // 添加复制按钮到所有代码块
    const addCopyButtons = () => {
      nextTick(() => {
        if (!markdownContainer.value) return;

        const codeBlocks = markdownContainer.value.querySelectorAll('pre code');
        codeBlocks.forEach((codeBlock) => {
          // 如果已经添加过复制按钮，则跳过
          if (codeBlock.parentNode.querySelector('.copy-btn')) return;

          const copyBtn = document.createElement('button');
          copyBtn.className = 'copy-btn';
          copyBtn.textContent = '复制';
          copyBtn.title = '复制代码';

          copyBtn.addEventListener('click', () => {
            const textToCopy = codeBlock.textContent;
            if (navigator.clipboard && navigator.clipboard.writeText) {
              navigator.clipboard.writeText(textToCopy).then(() => {
                copyBtn.textContent = '已复制!';
                setTimeout(() => {
                  copyBtn.textContent = '复制';
                }, 2000);
              }).catch(err => {
                console.error('复制失败:', err);
                fallbackCopyText(textToCopy, copyBtn);
              });
            } else {
              fallbackCopyText(textToCopy, copyBtn);
            }
          });

          const btnContainer = document.createElement('div');
          btnContainer.className = 'copy-btn-container';
          btnContainer.appendChild(copyBtn);

          codeBlock.parentNode.insertBefore(btnContainer, codeBlock);
        });
      });
    };

    // 检查是否触底
    const isScrolledToBottom = () => {
      if (!markdownContainer.value) return false;
      const { scrollTop, scrollHeight, clientHeight } = markdownContainer.value;
      return scrollHeight - scrollTop <= clientHeight + 5; // 允许 5px 的误差
    };

    // 滚动事件处理函数
    const handleScroll = () => {
      if (!markdownContainer.value) return;

      // 如果用户向上滚动，停止自动滚动
      if (markdownContainer.value.scrollTop < markdownContainer.value.scrollHeight - markdownContainer.value.clientHeight) {
        allowAutoScroll.value = false;
      }

      // 如果用户向下滚动并触底，恢复自动滚动
      if (isScrolledToBottom()) {
        allowAutoScroll.value = true;
      }
    };

    // 监听 source 变化，更新内容并根据状态决定是否自动滚动
    watch(compiledMarkdown, () => {
      nextTick(() => {
        addCopyButtons();

        // 如果允许自动滚动，则滚动到底部
        if (allowAutoScroll.value) {
          setTimeout(() => {
            if (markdownContainer.value) {
              markdownContainer.value.scrollTop = markdownContainer.value.scrollHeight;
            }
          }, 0);
        }
      });
    });

    // 生命周期钩子
    onMounted(() => {
      addCopyButtons();
      if (markdownContainer.value) {
        markdownContainer.value.scrollTop = markdownContainer.value.scrollHeight;
      }

      // 添加滚动事件监听器
      if (markdownContainer.value) {
        markdownContainer.value.addEventListener('scroll', handleScroll);
      }
    });

    onUnmounted(() => {
      // 移除滚动事件监听器
      if (markdownContainer.value) {
        markdownContainer.value.removeEventListener('scroll', handleScroll);
      }
    });

    return {
      compiledMarkdown,
      markdownContainer,
    };
  },
};
</script>

<style scoped>

.markdown-container {
  width: 100%;
  min-width: 0;
  height: 560px;
  overflow-y: auto;
  border: 1px solid #ddd;
  padding: 10px;
  box-sizing: border-box;
  font-family: 'Consolas', monospace;
  font-size: 14px;
  line-height: 1.7;
}

code{
  text-shadow: none !important;
}

/* 代码块样式 */
pre {
  position: relative;
  margin: 1em 0;
  border-radius: 4px;
  background: #282c34;
}

/* 使用 :deep() 处理动态内容 */
:deep(pre) {
  position: relative;
  margin: 1em 0;
  border-radius: 4px;
  background: #282c34 !important;
}

:deep(pre code) {
  display: block;
  padding: 1em;
  overflow-x: auto;
  color: #abb2bf;
  background: transparent !important;
  text-shadow: none !important;
}

pre code {
  display: block;
  padding: 1em;
  overflow-x: auto;
  color: #abb2bf;
}

/* 复制按钮样式 */
.copy-btn-container {
  position: absolute;
  top: 0.5em;
  right: 0.5em;
  z-index: 1;
}




:deep(blockquote) {
  border-left: 4px solid #4285f4;
  background: #f8f9fa !important;   /* 保险起见加 important */
  padding: 6px 12px;
  margin: 16px 0;
  border-radius: 0 4px 4px 0;
  font-style: italic;
  color: #555;
}

strong code {
  font-weight: bold;
  background: #f0f0f0;
  padding: 0.2em 0.4em;
  border-radius: 3px;
}
</style>
