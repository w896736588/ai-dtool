function normalizeCommandPart(value) {
  if (value === null || value === undefined) return ''
  return String(value).trim()
}

function hasConfiguredLinkAccounts(linkCmd) {
  const linkData = linkCmd && linkCmd.data ? linkCmd.data : linkCmd
  const userList = Array.isArray(linkData && linkData.userList) ? linkData.userList : []
  return userList.length > 0
}

function getLinkRunSelection(stack) {
  const sourceStack = Array.isArray(stack) ? stack : []
  const actionIndex = sourceStack.findIndex(item => item && item.action === 'linkRun')
  if (actionIndex < 0) {
    return { linkCmd: null, accountCmd: null }
  }
  const tailStack = sourceStack.slice(actionIndex + 1)
  const linkCmd = tailStack.find(item => item && item.data && item.data.__linkType === 'link') || null
  return {
    linkCmd,
    accountCmd: tailStack.find(item => item && item.data && item.data.__linkType === 'account') || null,
  }
}

function buildLinkAccountOptionsFromLink(linkCmd, normalize = normalizeCommandPart) {
  const userListRaw = Array.isArray(linkCmd && linkCmd.data && linkCmd.data.userList) ? linkCmd.data.userList : []
  if (userListRaw.length === 0) return []
  return userListRaw.map((item, index) => {
    const userName = normalize(item && item.user_name) || `账号${index + 1}`
    return {
      command: userName,
      name: userName,
      data: {
        __linkType: 'account',
        account: { user_name: normalize(item.user_name), password: normalize(item.password) },
      },
    }
  })
}

function isLinkRunSelectionComplete(selection) {
  if (!(selection && selection.linkCmd)) return false
  if (!hasConfiguredLinkAccounts(selection.linkCmd)) return true
  return !!selection.accountCmd
}

function buildLinkRunPayload(selection, sseDistributeId, normalize = normalizeCommandPart) {
  const linkData = (((selection && selection.linkCmd) || {}).data) || {}
  const accountData = (((selection && selection.accountCmd) || {}).data || {}).account || {}

  return {
    id: linkData.id,
    label: normalize(linkData.label),
    user_name: normalize(accountData.user_name),
    password: normalize(accountData.password),
    open_num: normalize(linkData.open_num),
    open_type: normalize(linkData.open_type),
    sse_distribute_id: sseDistributeId,
  }
}

module.exports = {
  buildLinkAccountOptionsFromLink,
  buildLinkRunPayload,
  getLinkRunSelection,
  hasConfiguredLinkAccounts,
  isLinkRunSelectionComplete,
  normalizeCommandPart,
}
