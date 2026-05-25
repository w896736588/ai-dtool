import assert from 'node:assert/strict'
import codexParser from '../src/utils/codex_chat_parser.js'

function testParsesErrorTextField() {
  const messages = codexParser.parseChatLines([
    JSON.stringify({
      type: 'error',
      text: 'Codex CLI 配置解析失败: missing api key',
    }),
  ])

  assert.equal(messages.length, 1)
  assert.equal(messages[0].type, 'error')
  assert.equal(messages[0].text, 'Codex CLI 配置解析失败: missing api key')
}

function testParsesThreadStartedWithoutModelAsEmptyString() {
  const messages = codexParser.parseChatLines([
    JSON.stringify({
      type: 'thread.started',
      thread_id: 'thread_123',
    }),
  ])

  assert.equal(messages.length, 1)
  assert.equal(messages[0].type, 'system_init')
  assert.equal(messages[0].text, '会话已创建')
  assert.equal(messages[0].model, '')
  assert.equal(messages[0].sessionId, 'thread_123')
}

function testTurnCompletedDoesNotDuplicateCommandResultText() {
  const messages = codexParser.parseChatLines([
    JSON.stringify({
      type: 'item.completed',
      item: {
        id: 'cmd_1',
        type: 'command_execution',
        command: 'echo hello',
        stdout: 'hello',
        exit_code: 0,
      },
    }),
    JSON.stringify({
      type: 'turn.completed',
      usage: {
        input_tokens: 10,
        output_tokens: 5,
      },
    }),
  ])

  assert.equal(messages.length, 2)
  assert.equal(messages[0].type, 'assistant')
  assert.equal(messages[0].content[0].type, 'tool_use')
  assert.equal(messages[0].content[0]._result.text, 'hello')
  assert.equal(messages[1].type, 'result')
  assert.equal(messages[1].resultText, undefined)
}

function testParsesUnifiedResultEventInjectedByBackend() {
  const messages = codexParser.parseChatLines([
    JSON.stringify({
      type: 'result',
      subtype: 'completed',
      duration_ms: 263200,
      num_turns: 12,
      usage: {
        input_tokens: 32531,
        output_tokens: 10578,
        cache_read_input_tokens: 184064,
      },
      modelUsage: {
        'deepseek-v4-pro[1m]': {
          inputTokens: 32531,
          outputTokens: 10578,
          cacheReadInputTokens: 184064,
        },
      },
    }),
  ])

  assert.equal(messages.length, 1)
  assert.equal(messages[0].type, 'result')
  assert.equal(messages[0].durationMs, 263200)
  assert.equal(messages[0].numTurns, 12)
  assert.equal(messages[0].modelUsage.length, 1)
  assert.equal(messages[0].modelUsage[0].name, 'deepseek-v4-pro[1m]')
}

function main() {
  testParsesErrorTextField()
  testParsesThreadStartedWithoutModelAsEmptyString()
  testTurnCompletedDoesNotDuplicateCommandResultText()
  testParsesUnifiedResultEventInjectedByBackend()
  console.log('codex_chat_parser tests passed')
}

main()
