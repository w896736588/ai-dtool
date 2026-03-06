import base from '../base'

// MemoryFragmentList 查询记忆片段列表。
function MemoryFragmentList(limit, callBack) {
  base.BasePost('/api/MemoryFragmentList', {
    limit: limit,
  }, callBack)
}

// MemoryFragmentInfo 查询记忆片段详情。
function MemoryFragmentInfo(id, callBack) {
  base.BasePost('/api/MemoryFragmentInfo', {
    id: id,
  }, callBack)
}

// MemoryFragmentSave 保存记忆片段。
function MemoryFragmentSave(id, title, content, tags, callBack) {
  base.BasePost('/api/MemoryFragmentSave', {
    id: id,
    title: title,
    content: content,
    tags: tags,
  }, callBack)
}

// MemoryFragmentDelete 删除记忆片段。
function MemoryFragmentDelete(id, callBack) {
  base.BasePost('/api/MemoryFragmentDelete', {
    id: id,
  }, callBack)
}

// MemoryFragmentHistoryList 查询片段历史记录。
function MemoryFragmentHistoryList(id, callBack) {
  base.BasePost('/api/MemoryFragmentHistoryList', {
    id: id,
  }, callBack)
}

// MemoryFragmentTagList 查询可用标签列表。
function MemoryFragmentTagList(callBack) {
  base.BasePost('/api/MemoryFragmentTagList', {}, callBack)
}

// MemoryFragmentSearch 搜索记忆片段。
function MemoryFragmentSearch(query, mode, selectedTags, limit, callBack) {
  base.BasePost('/api/MemoryFragmentSearch', {
    query: query,
    mode: mode,
    selected_tags: selectedTags,
    limit: limit,
  }, callBack)
}

export default {
  MemoryFragmentList,
  MemoryFragmentInfo,
  MemoryFragmentSave,
  MemoryFragmentDelete,
  MemoryFragmentHistoryList,
  MemoryFragmentTagList,
  MemoryFragmentSearch,
}
