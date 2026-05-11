import base from '../base'

// MemoryFragmentList 查询知识片段列表。
function MemoryFragmentStatus(callBack) {
  base.BasePost('/api/MemoryFragmentStatus', {}, callBack)
}

// MemoryFragmentList 查询知识片段列表。
function MemoryFragmentList(limit, offset, callBack) {
  base.BasePost('/api/MemoryFragmentList', {
    limit: limit,
    offset: offset,
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

// MemoryFragmentSearch 搜索知识片段。
function MemoryFragmentSearch(query, mode, selectedTags, limit, callBack) {
  base.BasePost('/api/MemoryFragmentSearch', {
    query: query,
    mode: mode,
    limit: limit,
  }, callBack)
}

// MemoryFragmentOrganize 使用 AI 整理当前知识片段内容。
function MemoryFragmentOrganize(id, title, content, tags, sseDistributeId, callBack) {
  base.BasePost('/api/MemoryFragmentOrganize', {
    id: id,
    title: title,
    content: content,
  }, callBack)
}

// MemoryFragmentShareCreate 创建 24 小时有效的片段分享链接 token。
function MemoryFragmentShareCreate(id, callBack) {
  base.BasePost('/api/MemoryFragmentShareCreate', {
    id: id,
  }, callBack)
}

// MemoryFragmentShareInfo 通过分享 token 查询只读片段详情。
function MemoryFragmentShareInfo(token, callBack) {
  base.BasePost('/api/MemoryFragmentShareInfo', {
    token: token,
  }, callBack)
}

// MemoryFragmentTagList 查询可用标签列表。
function MemoryFragmentTagList(callBack) {
  callBack({ ErrCode: 0, Data: [] })
}

// GitPendingStatus 检测主库和记忆库是否有待 commit 的 git 变更。
function GitPendingStatus(callBack) {
  base.BasePost('/api/GitPendingStatus', {}, callBack)
}

// MemoryFragmentImageUpload 上传图片到记忆库，返回可访问的 URL。
function MemoryFragmentImageUpload(file, callBack) {
  const form = new FormData()
  form.append('file', file)
  base.BasePostForm('/api/MemoryFragmentImageUpload', form, callBack)
}

// MemoryFragmentUploadZip 上传 ZIP 文件，解析 content.md + images/ 创建知识片段。
function MemoryFragmentUploadZip(file, apiBaseURL, callBack) {
  const form = new FormData()
  form.append('file', file)
  form.append('api_base_url', apiBaseURL)
  base.BasePostForm('/api/MemoryFragmentUploadZip', form, callBack)
}

// MemoryFragmentBatchInfoByPaths 批量按文件路径查询片段摘要（id + title）。
function MemoryFragmentBatchInfoByPaths(paths, callBack) {
  base.BasePost('/api/MemoryFragmentBatchInfoByPaths', {
    paths: paths,
  }, callBack)
}

// MemoryGitPull 手动拉取记忆库远程仓库最新内容。
function MemoryGitPull(callBack) {
  base.BasePost('/api/MemoryGitPull', {}, callBack)
}

export default {
  GitPendingStatus,
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
  MemoryFragmentShareCreate,
  MemoryFragmentShareInfo,
  MemoryFragmentImageUpload,
  MemoryFragmentUploadZip,
  MemoryFragmentBatchInfoByPaths,
  MemoryGitPull,
}
