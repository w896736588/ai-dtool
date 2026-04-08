const assert = require('assert')

const {
  buildMemoryDetailSearchMatches,
  normalizeActiveMatchIndex,
  getNextMemoryDetailMatchIndex,
} = require('../src/utils/memory_detail_search.cjs')

const matches = buildMemoryDetailSearchMatches(
  '缓存设计说明',
  [
    '# 缓存设计说明',
    '',
    '这里记录缓存命中率和缓存淘汰策略。',
    '再次提到缓存命中率。',
  ].join('\n'),
  '缓存'
)

assert.deepStrictEqual(
  matches.map(item => ({ field: item.field, index: item.index, text: item.text })),
  [
    { field: 'title', index: 0, text: '缓存' },
    { field: 'content', index: 2, text: '缓存' },
    { field: 'content', index: 14, text: '缓存' },
    { field: 'content', index: 20, text: '缓存' },
    { field: 'content', index: 32, text: '缓存' },
  ],
  '应返回标题和正文中的全部命中项，并保留字段与偏移信息'
)

assert.strictEqual(normalizeActiveMatchIndex(matches, -1), 0)
assert.strictEqual(normalizeActiveMatchIndex(matches, 2), 2)
assert.strictEqual(normalizeActiveMatchIndex(matches, 99), 0)

assert.strictEqual(getNextMemoryDetailMatchIndex(matches, 0, 1), 1)
assert.strictEqual(getNextMemoryDetailMatchIndex(matches, 4, 1), 0)
assert.strictEqual(getNextMemoryDetailMatchIndex(matches, 0, -1), 4)
assert.strictEqual(getNextMemoryDetailMatchIndex([], 0, 1), -1)

assert.deepStrictEqual(
  buildMemoryDetailSearchMatches('缓存设计说明', '缓存命中率', '   '),
  [],
  '空白关键字不应产生命中结果'
)

console.log('memory_detail_search tests passed')
