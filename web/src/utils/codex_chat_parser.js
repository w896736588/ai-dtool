// codex_chat_parser.js — Codex CLI JSONL 解析器
// 将 codex exec --json 的 JSONL 事件流解析为与 chat_parser.js 兼容的可渲染消息格式。
// 独立于 Claude 的 stream-json 解析逻辑，通过 chat_parser.js 顶层分发调用。

// buildSystemTaskMessage 将 system.task_updated 等任务事件归一化为可渲染任务消息。
function buildSystemTaskMessage(obj) {
  if (!obj || obj.type !== 'system') return null
  const subtype = obj.subtype || ''
  if (subtype !== 'task_updated') return null

  const patch = obj.patch && typeof obj.patch === 'object' ? obj.patch : {}
  const status = String(patch.status || obj.status || 'running').trim() || 'running'
  const description = String(obj.description || patch.description || '').trim() || `任务 ${obj.task_id || ''}`.trim() || '任务更新'

  return {
    type: 'system_task',
    description,
    taskId: obj.task_id || '',
    status,
    uuid: obj.uuid || '',
    sessionId: obj.session_id || '',
  }
}

// parseOneLine 解析单行 JSONL 事件，更新 messages 和 currentItems。
function parseOneLine(line, messages, currentItems) {
  if (!line || !line.trim()) return
  let obj = null
  try {
    obj = JSON.parse(line)
  } catch (e) {
    messages.push({ type: 'raw_text', text: line })
    return
  }

  const eventType = obj.type || ''

  if (eventType === 'thread.started') {
    messages.push({
      type: 'system_init',
      text: '会话已创建',
      model: obj.model || '',
      sessionId: obj.thread_id || '',
    })
  } else if (eventType === 'result') {
    let modelUsageList = null
    if (obj.modelUsage) {
      modelUsageList = Object.entries(obj.modelUsage).map(([name, info]) => ({
        name,
        inputTokens: info.inputTokens || 0,
        outputTokens: info.outputTokens || 0,
        cacheReadInputTokens: info.cacheReadInputTokens || 0,
        cacheCreationInputTokens: info.cacheCreationInputTokens || 0,
        costUSD: info.costUSD || 0,
      }))
    }
    messages.push({
      type: 'result',
      subtype: obj.subtype || '',
      durationMs: obj.duration_ms || 0,
      durationApiMs: obj.duration_api_ms || 0,
      numTurns: obj.num_turns || 0,
      usage: obj.usage || null,
      isError: obj.is_error || false,
      totalCostUsd: obj.total_cost_usd ?? null,
      modelUsage: modelUsageList,
      stopReason: obj.stop_reason || '',
      permissionDenials: obj.permission_denials || null,
      resultText: obj.result || '',
      uuid: obj.uuid || '',
      sessionId: obj.session_id || '',
    })
  } else if (eventType === 'turn.started') {
    // 内部状态，不显示
  } else if (eventType === 'turn.completed') {
    messages.push({
      type: 'result',
      subtype: 'completed',
      text: '回合完成',
      usage: obj.usage || null,
    })
  } else if (eventType === 'turn.failed') {
    messages.push({
      type: 'error',
      text: obj.error || obj.message || '回合执行失败',
    })
  } else if (eventType === 'error') {
    // Codex 可能发出临时 reconnect 通知，也可能是真正的错误
    const errorMsg = obj.text || obj.message || obj.error || '未知错误'
    messages.push({ type: 'error', text: errorMsg })
  } else if (eventType === 'item.started' || eventType === 'item.updated' || eventType === 'item.completed') {
    handleItemEvent(eventType, obj, messages, currentItems)
  } else if (eventType === 'chat' && obj.subtype === 'completed') {
    // 由后端注入的终止标记，与 Claude 保持一致
    messages.push({ type: 'chat_completed', text: '对话已完成' })
  } else if (eventType === 'system') {
    // 后端注入的 system/command 提示词展示（runCodexCommand 推送）
    if (obj.subtype === 'command') {
      messages.push({ type: 'system_command', text: obj.text || '', cliType: obj.cli_type || '', cmdLine: obj.cmd_line || '', collapsed: true })
    } else if (obj.subtype === 'task_updated') {
      messages.push(buildSystemTaskMessage(obj))
    } else {
      messages.push({ type: 'system', text: JSON.stringify(obj) })
    }
  } else {
    // 未知事件类型，原样输出
    messages.push({ type: 'raw_text', text: line })
  }
}

// handleItemEvent 处理 item.started / item.updated / item.completed 事件。
function handleItemEvent(eventType, obj, messages, currentItems) {
  const item = obj.item || {}
  const itemId = item.id || ''
  const itemType = item.type || ''

  if (eventType === 'item.started') {
    // 创建进行中的 item 追踪
    const tracked = { itemId, itemType, data: item }

    if (itemType === 'agent_message') {
      // 助手消息
      const msg = {
        type: 'assistant',
        role: 'assistant',
        model: '',
        content: [{ type: 'text', text: item.text || '' }],
        thinking: '',
        _codexItemId: itemId,
      }
      tracked.msgIndex = messages.length
      messages.push(msg)
    } else if (itemType === 'reasoning') {
      // 推理过程（类似 Claude thinking）
      const msg = {
        type: 'assistant',
        role: 'assistant',
        model: '',
        content: [],
        thinking: item.text || item.summary || '',
        _thinkingTiming: { startMs: Date.now(), durationMs: 0 },
        _thinkingCollapsed: false,
        _codexItemId: itemId,
      }
      tracked.msgIndex = messages.length
      messages.push(msg)
    } else if (itemType === 'command_execution') {
      // 命令执行（类似 Claude Bash tool_use）
      const msg = {
        type: 'assistant',
        role: 'assistant',
        model: '',
        content: [{
          type: 'tool_use',
          name: 'Bash',
          id: itemId,
          input: JSON.stringify({ command: item.command || '' }, null, 2),
          displayInput: item.command || '',
          _codexStatus: item.status || 'in_progress',
        }],
        thinking: '',
        _codexItemId: itemId,
      }
      tracked.msgIndex = messages.length
      messages.push(msg)
    } else if (itemType === 'file_change') {
      // 文件变更（类似 Claude Edit/Write tool_use）
      const changes = item.changes || []
      const changesSummary = changes.map(ch => `${ch.type || 'modify'}: ${ch.path || ''}`).join('\n')
      const msg = {
        type: 'assistant',
        role: 'assistant',
        model: '',
        content: [{
          type: 'tool_use',
          name: 'FileChange',
          id: itemId,
          input: JSON.stringify(item, null, 2),
          displayInput: changesSummary || '文件变更',
          _codexStatus: item.status || 'in_progress',
        }],
        thinking: '',
        _codexItemId: itemId,
      }
      tracked.msgIndex = messages.length
      messages.push(msg)
    } else if (itemType === 'mcp_tool_call') {
      // MCP 工具调用
      const msg = {
        type: 'assistant',
        role: 'assistant',
        model: '',
        content: [{
          type: 'tool_use',
          name: `MCP:${item.server || ''}/${item.tool || ''}`,
          id: itemId,
          input: JSON.stringify(item.arguments || {}, null, 2),
          displayInput: `${item.server || ''}/${item.tool || ''}`,
          _codexStatus: item.status || 'in_progress',
        }],
        thinking: '',
        _codexItemId: itemId,
      }
      tracked.msgIndex = messages.length
      messages.push(msg)
    } else if (itemType === 'web_search') {
      // 网络搜索
      const msg = {
        type: 'assistant',
        role: 'assistant',
        model: '',
        content: [{
          type: 'tool_use',
          name: 'WebSearch',
          id: itemId,
          input: JSON.stringify(item, null, 2),
          displayInput: item.query || '网络搜索',
          _codexStatus: item.status || 'in_progress',
        }],
        thinking: '',
        _codexItemId: itemId,
      }
      tracked.msgIndex = messages.length
      messages.push(msg)
    } else if (itemType === 'todo_list') {
      // 任务列表（类似 Claude TodoWrite）
      const msg = {
        type: 'system_task',
        description: '任务列表',
        status: 'running',
        _codexItemId: itemId,
        _todoItems: item.items || item.todos || [],
      }
      tracked.msgIndex = messages.length
      messages.push(msg)
    } else if (itemType === 'error') {
      const msg = {
        type: 'error',
        text: item.message || item.text || item.error || '未知错误',
        _codexItemId: itemId,
      }
      tracked.msgIndex = messages.length
      messages.push(msg)
    } else {
      // 未知 item 类型
      messages.push({ type: 'raw_text', text: JSON.stringify(obj) })
    }

    if (itemId) {
      currentItems.set(itemId, tracked)
    }
  } else if (eventType === 'item.updated') {
    // 更新已有 item
    const tracked = itemId ? currentItems.get(itemId) : null
    if (tracked && tracked.msgIndex !== undefined && tracked.msgIndex < messages.length) {
      const existingMsg = messages[tracked.msgIndex]
      if (itemType === 'agent_message' && existingMsg.content && existingMsg.content.length > 0) {
        existingMsg.content[0].text = item.text || ''
      } else if (itemType === 'reasoning') {
        existingMsg.thinking = item.text || item.summary || ''
      } else if (itemType === 'error') {
        existingMsg.text = item.message || item.text || item.error || existingMsg.text
      } else if (itemType === 'todo_list') {
        existingMsg._todoItems = item.items || item.todos || []
      } else if (existingMsg.content && existingMsg.content.length > 0) {
        existingMsg.content[0]._codexStatus = item.status || existingMsg.content[0]._codexStatus
      }
    } else if (!tracked) {
      // 兜底路径：未收到 item.started，从 item.updated 创建进行中的消息
      if (itemType === 'agent_message') {
        const msg = {
          type: 'assistant',
          role: 'assistant',
          model: '',
          content: [{ type: 'text', text: item.text || '' }],
          thinking: '',
          _codexItemId: itemId,
        }
        const newTracked = { itemId, itemType, data: item, msgIndex: messages.length }
        messages.push(msg)
        if (itemId) currentItems.set(itemId, newTracked)
      } else if (itemType === 'reasoning') {
        const msg = {
          type: 'assistant',
          role: 'assistant',
          model: '',
          content: [],
          thinking: item.text || item.summary || '',
          _thinkingTiming: { startMs: Date.now(), durationMs: 0 },
          _thinkingCollapsed: false,
          _codexItemId: itemId,
        }
        const newTracked = { itemId, itemType, data: item, msgIndex: messages.length }
        messages.push(msg)
        if (itemId) currentItems.set(itemId, newTracked)
      } else if (itemType === 'command_execution') {
        const msg = {
          type: 'assistant',
          role: 'assistant',
          model: '',
          content: [{
            type: 'tool_use',
            name: 'Bash',
            id: itemId,
            input: JSON.stringify({ command: item.command || '' }, null, 2),
            displayInput: item.command || '',
            _codexStatus: item.status || 'in_progress',
          }],
          thinking: '',
          _codexItemId: itemId,
        }
        const newTracked = { itemId, itemType, data: item, msgIndex: messages.length }
        messages.push(msg)
        if (itemId) currentItems.set(itemId, newTracked)
      }
    }
  } else if (eventType === 'item.completed') {
    const tracked = itemId ? currentItems.get(itemId) : null
    if (tracked && tracked.msgIndex !== undefined && tracked.msgIndex < messages.length) {
      // 正常路径：已收到 item.started，更新已有消息
      const existingMsg = messages[tracked.msgIndex]
      if (itemType === 'agent_message' && existingMsg.content && existingMsg.content.length > 0) {
        existingMsg.content[0].text = item.text || existingMsg.content[0].text
      } else if (itemType === 'reasoning') {
        existingMsg.thinking = item.text || item.summary || existingMsg.thinking
        if (existingMsg._thinkingTiming) {
          existingMsg._thinkingTiming.durationMs = Date.now() - existingMsg._thinkingTiming.startMs
        }
      } else if (itemType === 'error') {
        existingMsg.text = item.message || item.text || item.error || existingMsg.text
      } else if (itemType === 'command_execution' && existingMsg.content && existingMsg.content.length > 0) {
        const block = existingMsg.content[0]
        block._codexStatus = 'completed'
        // 添加执行结果（优先使用 aggregated_output，兼容 stdout/stderr）
        const resultParts = []
        if (item.aggregated_output !== undefined && item.aggregated_output !== null) {
          // Codex CLI 的 command_execution 在 completed 时输出内容在 aggregated_output 字段
          const output = item.aggregated_output.trim()
          if (output) resultParts.push(output)
        } else {
          if (item.stdout) resultParts.push(item.stdout)
          if (item.stderr) resultParts.push('[stderr] ' + item.stderr)
        }
        if (item.exit_code !== undefined && item.exit_code !== 0) resultParts.push('[exit_code] ' + item.exit_code)
        if (resultParts.length > 0) {
          block._result = { text: resultParts.join('\n'), collapsed: true }
        } else {
          // 即使无输出内容，也设置空结果以停止 spinner
          block._result = { text: '', collapsed: true }
        }
      } else if (itemType === 'file_change' && existingMsg.content && existingMsg.content.length > 0) {
        existingMsg.content[0]._codexStatus = 'completed'
        existingMsg.content[0].input = JSON.stringify(item, null, 2)
      } else if (itemType === 'todo_list') {
        existingMsg._todoItems = item.items || item.todos || existingMsg._todoItems
        existingMsg.status = 'completed'
      } else if (existingMsg.content && existingMsg.content.length > 0) {
        existingMsg.content[0]._codexStatus = 'completed'
      }
    } else {
      // 兜底路径：未收到 item.started，直接从 item.completed 创建完整消息
      // 典型场景：历史数据加载、消息流丢失、或 Codex CLI 仅输出 completed 事件
      if (itemType === 'agent_message') {
        const msg = {
          type: 'assistant',
          role: 'assistant',
          model: '',
          content: [{ type: 'text', text: item.text || '' }],
          thinking: '',
          _codexItemId: itemId,
        }
        messages.push(msg)
      } else if (itemType === 'reasoning') {
        const msg = {
          type: 'assistant',
          role: 'assistant',
          model: '',
          content: [],
          thinking: item.text || item.summary || '',
          _thinkingTiming: { startMs: 0, durationMs: 0 },
          _thinkingCollapsed: true,
          _codexItemId: itemId,
        }
        messages.push(msg)
      } else if (itemType === 'command_execution') {
        // 优先使用 aggregated_output，兼容 stdout/stderr
        const resultParts = []
        if (item.aggregated_output !== undefined && item.aggregated_output !== null) {
          const output = item.aggregated_output.trim()
          if (output) resultParts.push(output)
        } else {
          if (item.stdout) resultParts.push(item.stdout)
          if (item.stderr) resultParts.push('[stderr] ' + item.stderr)
        }
        if (item.exit_code !== undefined && item.exit_code !== 0) resultParts.push('[exit_code] ' + item.exit_code)
        const block = {
          type: 'tool_use',
          name: 'Bash',
          id: itemId,
          input: JSON.stringify({ command: item.command || '' }, null, 2),
          displayInput: item.command || '',
          _codexStatus: 'completed',
        }
        if (resultParts.length > 0) {
          block._result = { text: resultParts.join('\n'), collapsed: true }
        } else {
          // 即使无输出内容，也设置空结果以停止 spinner
          block._result = { text: '', collapsed: true }
        }
        const msg = {
          type: 'assistant',
          role: 'assistant',
          model: '',
          content: [block],
          thinking: '',
          _codexItemId: itemId,
        }
        messages.push(msg)
      } else if (itemType === 'file_change') {
        const changes = item.changes || []
        const changesSummary = changes.map(ch => `${ch.type || 'modify'}: ${ch.path || ''}`).join('\n')
        const msg = {
          type: 'assistant',
          role: 'assistant',
          model: '',
          content: [{
            type: 'tool_use',
            name: 'FileChange',
            id: itemId,
            input: JSON.stringify(item, null, 2),
            displayInput: changesSummary || '文件变更',
            _codexStatus: 'completed',
          }],
          thinking: '',
          _codexItemId: itemId,
        }
        messages.push(msg)
      } else if (itemType === 'mcp_tool_call') {
        const msg = {
          type: 'assistant',
          role: 'assistant',
          model: '',
          content: [{
            type: 'tool_use',
            name: `MCP:${item.server || ''}/${item.tool || ''}`,
            id: itemId,
            input: JSON.stringify(item.arguments || {}, null, 2),
            displayInput: `${item.server || ''}/${item.tool || ''}`,
            _codexStatus: 'completed',
          }],
          thinking: '',
          _codexItemId: itemId,
        }
        messages.push(msg)
      } else if (itemType === 'web_search') {
        const msg = {
          type: 'assistant',
          role: 'assistant',
          model: '',
          content: [{
            type: 'tool_use',
            name: 'WebSearch',
            id: itemId,
            input: JSON.stringify(item, null, 2),
            displayInput: item.query || '网络搜索',
            _codexStatus: 'completed',
          }],
          thinking: '',
          _codexItemId: itemId,
        }
        messages.push(msg)
      } else if (itemType === 'todo_list') {
        const msg = {
          type: 'system_task',
          description: '任务列表',
          status: 'completed',
          _codexItemId: itemId,
          _todoItems: item.items || item.todos || [],
        }
        messages.push(msg)
      } else if (itemType === 'error') {
        messages.push({
          type: 'error',
          text: item.message || item.text || item.error || '未知错误',
          _codexItemId: itemId,
        })
      } else {
        // 未知 item 类型
        messages.push({ type: 'raw_text', text: JSON.stringify(obj) })
      }
    }
    // 清理追踪
    if (itemId) {
      currentItems.delete(itemId)
    }
  }
}

// parseChatLines 全量解析 Codex JSONL 行为消息数组。
function parseChatLines(lines) {
  if (!lines || lines.length === 0) return []

  const messages = []
  const currentItems = new Map()

  for (const line of lines) {
    parseOneLine(line, messages, currentItems)
  }

  return messages
}

// parseChatLinesIncremental 增量解析新增的 SSE 行。
// 与 chat_parser.js 保持相同的调用接口。
function parseChatLinesIncremental(newLines, parseState, msgIndexOffset) {
  if (!newLines || newLines.length === 0) {
    return {
      newMessages: [],
      parseState: parseState || { currentItems: new Map(), pendingPatches: [] },
    }
  }

  const currentItems = (parseState && parseState.currentItems) || new Map()
  const messages = []

  // 调整 tracked.msgIndex 的偏移量
  const baseOffset = msgIndexOffset || 0

  for (const line of newLines) {
    const beforeLen = messages.length
    parseOneLine(line, messages, currentItems)
    // 更新新消息的 msgIndex（相对于全局消息数组）
    for (let i = beforeLen; i < messages.length; i++) {
      const msg = messages[i]
      if (msg._codexItemId) {
        const tracked = currentItems.get(msg._codexItemId)
        if (tracked) {
          tracked.msgIndex = baseOffset + i
        }
      }
    }
  }

  return {
    newMessages: messages,
    parseState: { currentItems, pendingPatches: [] },
  }
}

// flushParseState 刷新解析状态（Codex 不需要 flush currentMessage，保持接口一致）。
function flushParseState(parseState) {
  return []
}

export default {
  parseChatLines,
  parseChatLinesIncremental,
  flushParseState,
}
