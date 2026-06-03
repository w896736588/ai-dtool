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

function testParsesCompletedErrorItemAsErrorMessage() {
  const messages = codexParser.parseChatLines([
    JSON.stringify({
      type: 'item.completed',
      item: {
        id: 'item_6',
        type: 'error',
        message: 'in-process app-server event stream lagged; dropped 7 events',
      },
    }),
  ])

  assert.equal(messages.length, 1)
  assert.equal(messages[0].type, 'error')
  assert.equal(messages[0].text, 'in-process app-server event stream lagged; dropped 7 events')
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

function testParsesSystemTaskUpdatedEvent() {
  const messages = codexParser.parseChatLines([
    JSON.stringify({
      type: 'system',
      subtype: 'task_updated',
      task_id: 'b1kekq04w',
      patch: {
        status: 'completed',
        end_time: 1780474188733,
      },
      uuid: 'f5d3683c-ccf5-4f6b-afbc-250fc7e2330d',
      session_id: '88b38486-b8ed-4aa5-b065-4c5b44c72bf6',
    }),
  ])

  assert.equal(messages.length, 1)
  assert.equal(messages[0].type, 'system_task')
  assert.equal(messages[0].taskId, 'b1kekq04w')
  assert.equal(messages[0].status, 'completed')
  assert.equal(messages[0].description, '任务 b1kekq04w')
  assert.equal(messages[0].sessionId, '88b38486-b8ed-4aa5-b065-4c5b44c72bf6')
}

function main() {
  testParsesErrorTextField()
  testParsesCompletedErrorItemAsErrorMessage()
  testParsesThreadStartedWithoutModelAsEmptyString()
  testTurnCompletedDoesNotDuplicateCommandResultText()
  testParsesUnifiedResultEventInjectedByBackend()
  testParsesSystemTaskUpdatedEvent()
  console.log('codex_chat_parser tests passed')
}

main()
