const assert = require('assert')

const MODULE_PATH = '../src/utils/dashboard_history_rank.cjs'

const loadHistoryRankModule = () => require(MODULE_PATH)

const run = () => {
  const {
    buildTopHistoryCommands,
    normalizeHistoryCommandText,
  } = loadHistoryRankModule()

  assert.strictEqual(normalizeHistoryCommandText('  git   status  '), 'git status')
  assert.strictEqual(normalizeHistoryCommandText(''), '')

  const rankedCommands = buildTopHistoryCommands({
    historyList: [
      'git status',
      'docker ps',
      'git pull',
      'git status',
      'npm run dev',
      'docker ps',
      'git status',
      'go test ./...',
    ],
    usageMap: {
      'git status': 7,
      'docker ps': 4,
      'npm run dev': 4,
      'git pull': 2,
      'go test ./...': 1,
      'unused command': 99,
    },
    limit: 3,
  })

  assert.deepStrictEqual(
    rankedCommands,
    ['git status', 'docker ps', 'npm run dev'],
    '应按使用次数降序排列；次数相同时按最近使用优先 / should rank by usage desc and recency on ties'
  )

  const deduplicatedCommands = buildTopHistoryCommands({
    historyList: [
      '  git   status  ',
      'docker ps',
      'git status',
      'npm run dev',
      'git status',
    ],
    usageMap: {
      'git status': 3,
      'docker ps': 2,
      'npm run dev': 1,
    },
    limit: 10,
  })

  assert.deepStrictEqual(
    deduplicatedCommands,
    ['git status', 'docker ps', 'npm run dev'],
    '应去重并保持同名命令只展示一次 / should deduplicate normalized commands'
  )

  console.log('dashboard_history_rank tests passed')
}

run()
