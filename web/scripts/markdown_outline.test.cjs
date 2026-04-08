const assert = require('assert')

const {
  buildMarkdownOutline,
} = require('../src/utils/markdown_outline.cjs')

const markdown = [
  '# 总览',
  '',
  '正文段落',
  '',
  '## 安装',
  '',
  '### Windows',
  '',
  '## 安装',
  '',
  '```md',
  '# 代码块里的标题不应出现在目录',
  '```',
  '',
  '#### 太深的标题不纳入目录',
].join('\n')

const outline = buildMarkdownOutline(markdown)

assert.deepStrictEqual(
  outline,
  [
    { level: 1, text: '总览', slug: 'u603bu89c8', lineNumber: 1 },
    { level: 2, text: '安装', slug: 'u5b89u88c5', lineNumber: 5 },
    { level: 3, text: 'Windows', slug: 'windows', lineNumber: 7 },
    { level: 2, text: '安装', slug: 'u5b89u88c5-2', lineNumber: 9 },
  ],
  '目录应只提取 h1-h3 标题，生成唯一锚点、记录行号，并忽略代码块与过深标题'
)

console.log('markdown_outline tests passed')
