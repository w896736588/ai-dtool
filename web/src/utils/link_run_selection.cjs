function normalizeCommandPart(value) {
  if (value === null || value === undefined) return ''
  return String(value).trim()
}

function hasConfiguredLinkAccounts(envCmdOrEnvData) {
  const envData = envCmdOrEnvData && envCmdOrEnvData.data && envCmdOrEnvData.data.env
    ? envCmdOrEnvData.data.env
    : envCmdOrEnvData
  const userList = Array.isArray(envData && envData.userList) ? envData.userList : []
  return userList.length > 0
}

function getLinkRunSelection(stack) {
  const sourceStack = Array.isArray(stack) ? stack : []
  const actionIndex = sourceStack.findIndex(item => item && item.action === 'linkRun')
  if (actionIndex < 0) {
    return {
      configCmd: null,
      envCmd: null,
      accountCmd: null,
    }
  }
  const tailStack = sourceStack.slice(actionIndex + 1)
  const envCmd = tailStack.find(item => item && item.data && item.data.__linkType === 'env') || null
  return {
    configCmd: tailStack.find(item => item && item.data && item.data.__linkType === 'config') || (envCmd && envCmd.data && envCmd.data.config ? { data: envCmd.data.config } : null),
    envCmd,
    accountCmd: tailStack.find(item => item && item.data && item.data.__linkType === 'account') || null,
  }
}

function buildLinkEnvOptionsFromConfig(configCmd, normalize = normalizeCommandPart) {
  const linkList = Array.isArray(configCmd && configCmd.data && configCmd.data.linkList) ? configCmd.data.linkList : []
  return linkList.map((item, index) => {
    const envName = normalize(item && item.label) || `环境${index + 1}`
    return {
      command: envName,
      name: envName,
      dynamicChildren: hasConfiguredLinkAccounts(item) ? 'linkAccountList' : undefined,
      data: {
        __linkType: 'env',
        env: item || {},
        config: (configCmd && configCmd.data) || {},
      },
    }
  })
}

function buildLinkAccountOptionsFromEnv(envCmd, normalize = normalizeCommandPart) {
  const userListRaw = Array.isArray(envCmd && envCmd.data && envCmd.data.env && envCmd.data.env.userList) ? envCmd.data.env.userList : []
  const userList = userListRaw.length > 0 ? userListRaw : [{ user_name: '默认账号(空)', password: '' }]
  return userListRaw.map((item, index) => {
    const userName = normalize(item && item.user_name) || `账号${index + 1}`
    return {
      command: userName,
      name: userName,
      data: {
        __linkType: 'account',
        account: {
          user_name: normalize(item && item.user_name),
          password: normalize(item && item.password),
        },
      },
    }
  })
}

function isLinkRunSelectionComplete(selection) {
  if (!(selection && selection.envCmd)) {
    return false
  }
  if (!hasConfiguredLinkAccounts(selection.envCmd)) {
    return true
  }
  return !!selection.accountCmd
}

function buildLinkRunPayload(selection, sseDistributeId, normalize = normalizeCommandPart) {
  const configData = ((selection && selection.configCmd) || {}).data || (((selection && selection.envCmd) || {}).data || {}).config || {}
  const envData = (((selection && selection.envCmd) || {}).data || {}).env || {}
  const accountData = (((selection && selection.accountCmd) || {}).data || {}).account || {}

  return {
    id: configData.id,
    label: normalize(envData.label),
    user_name: normalize(accountData.user_name),
    password: normalize(accountData.password),
    open_num: normalize(configData.open_num),
    open_type: normalize(configData.open_type),
    sse_distribute_id: sseDistributeId,
  }
}

module.exports = {
  buildLinkAccountOptionsFromEnv,
  buildLinkEnvOptionsFromConfig,
  buildLinkRunPayload,
  getLinkRunSelection,
  hasConfiguredLinkAccounts,
  isLinkRunSelectionComplete,
  normalizeCommandPart,
}
