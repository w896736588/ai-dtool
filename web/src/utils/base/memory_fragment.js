import base from '../base'

// MemoryFragmentList 查询知识片段列表。
function MemoryFragmentStatus(callBack) {
  base.BasePost('/api/MemoryFragmentStatus', {}, callBack)
}

// MemoryFragmentList 查询知识片段列表。
function MemoryFragmentList(limit, callBack) {
  base.BasePost('/api/MemoryFragmentList', {
    limit: limit,
  }, callBack)
}

// MemoryFragmentInfo 查询知识片段详情。
function MemoryFragmentInfo(id, callBack) {
  base.BasePost('/api/MemoryFragmentInfo', {
    id: id,
  }, callBack)
}

// MemoryFragmentSave 保存知识片段。
function MemoryFragmentSave(id, title, content, tags, callBack) {
  base.BasePost('/api/MemoryFragmentSave', {
    id: id,
    title: title,
    content: content,
    tags: tags,
  }, callBack)
}

// MemoryFragmentDelete 删除知识片段。
function MemoryFragmentDelete(id, callBack) {
  base.BasePost('/api/MemoryFragmentDelete', {
    id: id,
  }, callBack)
}

// MemoryFragmentTrashList 查询回收站中的片段列表。
function MemoryFragmentTrashList(limit, callBack) {
  base.BasePost('/api/MemoryFragmentTrashList', {
    limit: limit,
  }, callBack)
}

// MemoryFragmentRestore 恢复回收站中的片段。
function MemoryFragmentRestore(id, callBack) {
  base.BasePost('/api/MemoryFragmentRestore', {
    id: id,
  }, callBack)
}

// MemoryFragmentHardDelete 彻底删除回收站中的片段。
function MemoryFragmentHardDelete(id, callBack) {
  base.BasePost('/api/MemoryFragmentHardDelete', {
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

// MemoryFragmentSearch 搜索知识片段。
function MemoryFragmentSearch(query, mode, selectedTags, limit, callBack) {
  base.BasePost('/api/MemoryFragmentSearch', {
    query: query,
    mode: mode,
    selected_tags: selectedTags,
    limit: limit,
  }, callBack)
}

// MemoryFragmentOrganize 使用 AI 整理当前知识片段内容。
function MemoryFragmentOrganize(id, title, content, tags, callBack) {
  base.BasePost('/api/MemoryFragmentOrganize', {
    id: id,
    title: title,
    content: content,
    tags: tags,
  }, callBack)
}

export default {
  MemoryFragmentStatus,
  MemoryFragmentList,
  MemoryFragmentInfo,
  MemoryFragmentSave,
  MemoryFragmentDelete,
  MemoryFragmentTrashList,
  MemoryFragmentRestore,
  MemoryFragmentHardDelete,
  MemoryFragmentHistoryList,
  MemoryFragmentTagList,
  MemoryFragmentSearch,
  MemoryFragmentOrganize,
}
