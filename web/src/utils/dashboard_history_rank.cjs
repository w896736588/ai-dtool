// DEFAULT_TOP_HISTORY_LIMIT 默认展示的高频历史命令数量。
// DEFAULT_TOP_HISTORY_LIMIT defines the default number of top history commands to show.
const DEFAULT_TOP_HISTORY_LIMIT = 10

// normalizeHistoryCommandText 统一规范历史命令文本，避免空白差异影响统计。
// normalizeHistoryCommandText normalizes history command text to avoid whitespace affecting ranking.
function normalizeHistoryCommandText(rawText) {
  return String(rawText || '').trim().replace(/\s+/g, ' ')
}

// buildTopHistoryCommands 按使用次数和最近使用顺序构建高频历史命令列表。
// buildTopHistoryCommands builds top history commands by usage count and recency.
function buildTopHistoryCommands({ historyList = [], usageMap = {}, limit = DEFAULT_TOP_HISTORY_LIMIT } = {}) {
  const normalizedLimit = Number(limit) > 0 ? Number(limit) : DEFAULT_TOP_HISTORY_LIMIT
  const rankedCommandMap = new Map()

  ;(Array.isArray(historyList) ? historyList : []).forEach((item, index) => {
    const commandText = normalizeHistoryCommandText(item)
    if (!commandText) {
      return
    }
    const commandUsage = Number(usageMap?.[commandText]) || 0
    if (commandUsage <= 0) {
      return
    }

    const lastUsedIndex = index
    const previous = rankedCommandMap.get(commandText)
    if (!previous || lastUsedIndex > previous.lastUsedIndex) {
      rankedCommandMap.set(commandText, {
        commandText,
        usage: commandUsage,
        lastUsedIndex,
      })
    }
  })

  return Array.from(rankedCommandMap.values())
    .sort((left, right) => {
      // 使用次数高的排前面；次数相同时最近使用的排前面。
      // Higher usage ranks first; when tied, the more recently used command ranks first.
      if (right.usage !== left.usage) {
        return right.usage - left.usage
      }
      return right.lastUsedIndex - left.lastUsedIndex
    })
    .slice(0, normalizedLimit)
    .map(item => item.commandText)
}

module.exports = {
  DEFAULT_TOP_HISTORY_LIMIT,
  normalizeHistoryCommandText,
  buildTopHistoryCommands,
}
