const assert = require('assert')
const path = require('path')

const {
  mergeHomeTaskFragmentOptions,
} = require(path.join(__dirname, '../src/utils/home_task_fragment_options.cjs'))

const merged = mergeHomeTaskFragmentOptions([
  { file_id: '1744260592119040000', title: '片段A', tags: ['需求'] },
  { file_id: '1744260592119040256', title: '片段B', tags: [] },
], {
  file_id: '1744260592119040512',
  title: '当前关联片段',
  tags: ['任务'],
})

assert.deepStrictEqual(
  merged.map((item) => item.id),
  ['1744260592119040512', '1744260592119040000', '1744260592119040256'],
  '当前编辑任务关联的知识片段即使不在列表接口中，也应保留在下拉选项首位'
)

assert.deepStrictEqual(
  mergeHomeTaskFragmentOptions([
    { file_id: '1744260592119040512', title: '当前关联片段', tags: ['任务'] },
    { file_id: '1744260592119040000', title: '片段A', tags: [] },
  ], {
    file_id: '1744260592119040512',
    title: '当前关联片段',
    tags: ['任务'],
  }).map((item) => item.id),
  ['1744260592119040512', '1744260592119040000'],
  '当前关联片段已在列表中时，不应重复插入'
)

console.log('home_task_fragment_options tests passed')
