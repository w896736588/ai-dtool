import taskStore from '@/utils/task_progress_store'
import codexParser from '@/utils/codex_chat_parser'

// tryParse 尝试将值解析为 JSON 对象/数组，若已是对象/数组则直接返回
function tryParse(v) {
  if (v === null || v === undefined) return v
  if (typeof v === 'string') {
    try { return JSON.parse(v) } catch (e) { return v }
  }
  return v
}

// buildToolDisplayInput 根据工具名和解析后的参数生成可读摘要。
function buildToolDisplayInput(name, parsed) {
  if (!name || !parsed) return null
  const n = name.toLowerCase()
  if (n === 'bash' && parsed.command) return parsed.command
  if ((n === 'write' || n === 'edit') && parsed.command) return parsed.command
  if (n === 'read' && parsed.file_path) return parsed.file_path
  if (parsed.pattern) return parsed.pattern
  if (parsed.description) return parsed.description
  if (parsed.command) return parsed.command
  if (parsed.file_path) return parsed.file_path
  // TodoWrite: 支持数组、{ newTodos }、{ todos }、或 JSON 字符串
  if (n === 'todowrite') {
    const obj = tryParse(parsed)
    const todos = Array.isArray(obj) ? obj : (obj && Array.isArray(obj.newTodos) ? obj.newTodos : (obj && Array.isArray(obj.todos) ? obj.todos : null))
    if (todos && todos.length) {
      const total = todos.length
      const completed = todos.filter(t => t.status === 'completed').length
      if (completed > 0 && completed < total) return `${total} 个任务 (${completed}/${total} 完成)`
      if (completed === total) return `${total} 个任务 (全部完成)`
      return `${total} 个任务`
    }
  }
  // AskUserQuestion: 提取第一个问题文本
  if (n === 'askuserquestion' && parsed.questions && parsed.questions.length) {
    return parsed.questions[0].question || ''
  }
  // TaskCreate / TaskUpdate: 数组或 JSON 字符串
  if (n === 'taskcreate' || n === 'taskupdate') {
    const obj = tryParse(parsed)
    if (Array.isArray(obj) && obj.length) {
      const total = obj.length
      const completed = obj.filter(t => t.status === 'completed').length
      if (completed > 0 && completed < total) return `${total} 个任务 (${completed}/${total} 完成)`
      if (completed === total) return `${total} 个任务 (全部完成)`
      return `${total} 个任务`
    }
  }
  const keys = Object.keys(parsed)
  if (keys.length === 1) return parsed[keys[0]]
  return null
}

// extractTasks 从工具参数中提取任务列表。
// TodoWrite 支持数组、{ newTodos }、或 JSON 字符串；TaskCreate/TaskUpdate 为数组或字符串。
function extractTasks(name, parsed) {
  if (!name || !parsed) return null
  const n = name.toLowerCase()
  const obj = tryParse(parsed)
  if (n === 'todowrite') {
    const todos = Array.isArray(obj) ? obj : (obj && Array.isArray(obj.newTodos) ? obj.newTodos : (obj && Array.isArray(obj.todos) ? obj.todos : null))
    if (todos && todos.length) return todos
  }
  if ((n === 'taskcreate' || n === 'taskupdate') && Array.isArray(obj) && obj.length) {
    return obj
  }
  return null
}

// extractAskUserQuestions 从 AskUserQuestion 工具参数中提取问题列表。
function extractAskUserQuestions(name, parsed) {
  if (!name || !parsed) return null
  const n = name.toLowerCase()
  if (n !== 'askuserquestion') return null
  const questions = parsed.questions
  if (Array.isArray(questions) && questions.length) return questions
  return null
}

// extractTasksFromToolUseResult 从顶层 tool_use_result 字段提取任务列表。
// tool_use_result 格式: { oldTodos: [...], newTodos: [...], verificationNudgeNeeded: false }
function extractTasksFromToolUseResult(toolUseResult) {
  if (!toolUseResult) return null
  const todos = toolUseResult.newTodos || toolUseResult.todos
  if (Array.isArray(todos) && todos.length) return todos
  return null
}

// parseChatLines 将 claude stream-json 输出解析为可渲染的消息数组。
// parseOneLine 解析单行 SSE 数据，更新 messages、currentMessage 和 toolUseMap。
// 提取为独立函数以便全量解析和增量解析共用。
function parseOneLine(line, messages, currentMessageRef, toolUseMap, msgIndexOffset, skipTypes) {
  const idxOffset = msgIndexOffset || 0
  if (!line || !line.trim()) return currentMessageRef.value
  let obj = null
  try {
    obj = JSON.parse(line)
  } catch (e) {
    messages.push({ type: 'raw_text', text: line })
    return currentMessageRef.value
  }

  const lineType = obj.type || ''

  // 当 stream_event 行存在时，跳过 assistant 行以避免重复渲染
  // assistant 行是会话历史的汇总格式，其内容已通过 stream_event 行呈现
  if (skipTypes && skipTypes.has(lineType)) {
    return currentMessageRef.value
  }
  let cm = currentMessageRef.value

  if (lineType === 'system') {
    const subtype = obj.subtype || ''
    if (subtype === 'init') {
      messages.push({ type: 'system_init', text: obj.is_resume ? '继续对话' : '会话已创建', model: obj.model || '', sessionId: obj.session_id || '' })
    } else if (subtype === 'command') {
      messages.push({ type: 'system_command', text: obj.text || '', cliType: obj.cli_type || '', cmdLine: obj.cmd_line || '', collapsed: true })
    } else if (subtype === 'hook_started' || subtype === 'hook_response') {
      messages.push({ type: 'system_hook', text: subtype === 'hook_started' ? 'Hook started: ' + (obj.hook_name || '') : 'Hook response: ' + (obj.hook_name || ''), collapsed: true })
    } else if (subtype === 'hook_progress') {
      messages.push({ type: 'system_hook', text: 'Hook progress: ' + (obj.hook_name || ''), collapsed: true, stderr: obj.stderr || '', output: obj.output || '' })
    } else if (subtype === 'status') {
      const statusMap = { requesting: '请求中', compressing: '压缩中' }
      messages.push({ type: 'system_status', status: obj.status || '', text: statusMap[obj.status] || obj.status })
    } else if (subtype === 'task_started') {
      const taskMsg = { type: 'system_task', description: obj.description || '', taskId: obj.task_id || '', status: 'started', _msgIndex: idxOffset + messages.length }
      messages.push(taskMsg)
      taskStore.updateFromMessage(taskMsg)
    } else if (subtype === 'task_progress') {
      const taskMsg = {
        type: 'system_task',
        description: obj.description || '',
        taskId: obj.task_id || '',
        status: 'running',
        usage: obj.usage || null,
        lastToolName: obj.last_tool_name || '',
        uuid: obj.uuid || '',
        sessionId: obj.session_id || '',
        _msgIndex: idxOffset + messages.length,
      }
      messages.push(taskMsg)
      taskStore.updateFromMessage(taskMsg)
    } else if (subtype === 'task_notification') {
      const taskMsg = { type: 'system_task', description: obj.summary || '', taskId: obj.task_id || '', status: obj.status || '', _msgIndex: idxOffset + messages.length }
      messages.push(taskMsg)
      taskStore.updateFromMessage(taskMsg)
    } else {
      messages.push({ type: 'system', text: JSON.stringify(obj) })
    }
  } else if (lineType === 'stream_event') {
    const event = obj.event || {}
    const eventType = event.type || ''

    if (eventType === 'message_start') {
      if (cm && (cm.content.length > 0 || cm.thinking)) {
        messages.push(cm)
      }
      cm = {
        type: 'assistant',
        role: event.message?.role || 'assistant',
        model: event.message?.model || '',
        content: [],
        thinking: '',
        usage: null,
        _thinkingTiming: { startMs: 0, durationMs: 0 },
        _thinkingCollapsed: false,
      }
    } else if (eventType === 'content_block_start') {
      if (!cm) return cm
      cm._blockType = event.content_block?.type || ''
      if (cm._blockType === 'tool_use') {
        const block = {
          type: 'tool_use',
          name: event.content_block?.name || '',
          id: event.content_block?.id || '',
          input: '',
        }
        cm.content.push(block)
        if (block.id) {
          toolUseMap.set(block.id, { msg: null, block, isNew: true })
        }
      }
    } else if (eventType === 'content_block_delta') {
      if (!cm) return cm
      const delta = event.delta || {}
      if (delta.type === 'text_delta') {
        if (cm._blockType === 'tool_use' && cm.content.length > 0) {
          const last = cm.content[cm.content.length - 1]
          if (last.type === 'tool_use') {
            last.input += (delta.text || '')
          }
        } else {
          const lastContent = cm.content[cm.content.length - 1]
          if (lastContent && lastContent.type === 'text') {
            lastContent.text += (delta.text || '')
          } else {
            cm.content.push({ type: 'text', text: delta.text || '' })
          }
        }
      } else if (delta.type === 'thinking_delta') {
        cm.thinking += (delta.thinking || '')
      } else if (delta.type === 'input_json_delta') {
        if (cm.content.length > 0) {
          const last = cm.content[cm.content.length - 1]
          if (last.type === 'tool_use') {
            last.input += (delta.partial_json || '')
          }
        }
      }
    } else if (eventType === 'content_block_stop') {
      if (cm) {
        if (cm._blockType === 'tool_use' && cm.content.length > 0) {
          const last = cm.content[cm.content.length - 1]
          if (last.type === 'tool_use' && last.input) {
            try {
              const parsed = JSON.parse(last.input)
              last.inputObj = parsed
              last.input = JSON.stringify(parsed, null, 2)
              const di = buildToolDisplayInput(last.name, parsed)
              if (di) last.displayInput = di
              // TodoWrite / TaskCreate / TaskUpdate: 挂载任务列表
              const tasks = extractTasks(last.name, parsed)
              if (tasks) last._tasks = tasks
              // AskUserQuestion: 挂载问题列表
              const questions = extractAskUserQuestions(last.name, parsed)
              if (questions) last._askQuestions = questions
            } catch (e) {
              // 解析失败，保留原始字符串
            }
          }
        }
        cm._blockType = ''
      }
    } else if (eventType === 'message_delta') {
      if (cm) {
        cm.usage = event.delta?.usage || event.usage || null
      }
    } else if (eventType === 'message_stop') {
      if (cm && (cm.content.length > 0 || cm.thinking)) {
        messages.push(cm)
      }
      cm = null
    }
  } else if (lineType === 'user') {
    const content = obj.message?.content || []
    // 提取顶层 tool_use_result 中的任务数据（如有）
    const toolUseResultTasks = extractTasksFromToolUseResult(obj.tool_use_result)
    for (const part of content) {
      if (part.type === 'tool_result') {
        const text = typeof part.content === 'string' ? part.content : JSON.stringify(part.content || '')
        const toolUseId = part.tool_use_id || ''
        // 尝试即时配对
        const target = toolUseMap.get(toolUseId)
        if (target) {
          const resultData = { text, collapsed: true }
          if (toolUseResultTasks) resultData._tasks = toolUseResultTasks
          if (target.isNew) {
            // 当前 batch 内新创建的对象，直接赋值即可（尚未变为响应式）
            if (target.block) {
              target.block._result = resultData
            } else if (target.msg) {
              target.msg._result = resultData
            }
          } else {
            // 来自前序 batch（已推入响应式数组），标记为待补丁
            target._pendingResult = resultData
          }
        } else {
          // 未找到 tool_use，保留为独立消息
          const standalone = { type: 'tool_result', text, collapsed: true, toolUseId }
          if (toolUseResultTasks) standalone._tasks = toolUseResultTasks
          messages.push(standalone)
        }
      }
    }
  } else if (lineType === 'result') {
    // 格式化模型用量数据，便于前端展示
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
  } else if (lineType === 'assistant') {
    const content = obj.message?.content || []
    for (const part of content) {
      if (part.type === 'text') {
        messages.push({ type: 'assistant_text', text: part.text || '' })
      } else if (part.type === 'thinking') {
        messages.push({ type: 'assistant_thinking', text: part.thinking || '', collapsed: true })
      } else if (part.type === 'tool_use') {
        const inputObj = tryParse(part.input || {})
        const di = buildToolDisplayInput(part.name, inputObj)
        const tuMsg = { type: 'tool_use', name: part.name || '', id: part.id || '', input: typeof inputObj === 'object' ? JSON.stringify(inputObj, null, 2) : (inputObj || ''), displayInput: di }
        // TodoWrite / TaskCreate / TaskUpdate: 挂载任务列表
        const tasks = extractTasks(part.name, inputObj)
        if (tasks) tuMsg._tasks = tasks
        // AskUserQuestion: 挂载问题列表
        const questions = extractAskUserQuestions(part.name, inputObj)
        if (questions) tuMsg._askQuestions = questions
        messages.push(tuMsg)
        if (part.id) {
          toolUseMap.set(part.id, { msg: tuMsg, isNew: true })
        }
      }
    }
  } else if (lineType === 'chat') {
    if (obj.subtype === 'completed') {
      // 先将待完成的 assistant 消息推入，确保 chat_completed 在内容之后
      if (currentMessageRef.value && (currentMessageRef.value.content.length > 0 || currentMessageRef.value.thinking)) {
        messages.push(currentMessageRef.value)
        currentMessageRef.value = null
      }
      messages.push({ type: 'chat_completed', text: '对话已完成' })
    }
  } else if (lineType === 'parse_error') {
    const data = obj.data || {}
    messages.push({ type: 'parse_error', text: data.line || obj.line || '', error: data.error || '' })
  } else if (lineType === 'raw_text') {
    const data = obj.data || {}
    messages.push({ type: 'raw_text', text: data.text || obj.text || '' })
  } else if (lineType === 'error') {
    // 先将待完成的 assistant 消息推入，确保错误信息在内容之后
    if (currentMessageRef.value && (currentMessageRef.value.content.length > 0 || currentMessageRef.value.thinking)) {
      messages.push(currentMessageRef.value)
      currentMessageRef.value = null
    }
    messages.push({ type: 'error', text: obj.text || '' })
  }

  return cm
}

function parseChatLines(lines) {
  if (!lines || lines.length === 0) return []

  const messages = []
  let currentMessage = null
  const toolUseMap = new Map()

  // 检查是否包含 stream_event 行，若有则跳过 assistant 行以避免重复
  // assistant 行是会话历史的汇总格式，其内容已通过 stream_event 行呈现
  let skipTypes = null
  for (const line of lines) {
    try {
      const obj = JSON.parse(line)
      if (obj.type === 'stream_event') {
        skipTypes = new Set(['assistant'])
        break
      }
    } catch (e) { /* ignore parse errors in quick scan */ }
  }

  for (const line of lines) {
    currentMessage = parseOneLine(line, messages, { value: currentMessage }, toolUseMap, 0, skipTypes)
  }

  // flush remaining message
  if (currentMessage && (currentMessage.content.length > 0 || currentMessage.thinking)) {
    messages.push(currentMessage)
  }

  // 后处理：将 tool_result 配对到对应的 tool_use 下
  const toolUseMapFull = new Map()
  for (let i = 0; i < messages.length; i++) {
    const msg = messages[i]
    if (msg.type === 'assistant') {
      for (const block of (msg.content || [])) {
        if (block.type === 'tool_use' && block.id) {
          toolUseMapFull.set(block.id, { msg, block, index: i })
        }
      }
    } else if (msg.type === 'tool_use' && msg.id) {
      toolUseMapFull.set(msg.id, { msg })
    }
  }

  const toRemove = new Set()
  for (let i = 0; i < messages.length; i++) {
    const msg = messages[i]
    if (msg.type === 'tool_result' && msg.toolUseId) {
      const target = toolUseMapFull.get(msg.toolUseId)
      if (target) {
        const resultData = { text: msg.text, collapsed: msg.collapsed }
        if (target.block) {
          target.block._result = resultData
        } else {
          target.msg._result = resultData
        }
        toRemove.add(i)
      }
    }
  }

  if (toRemove.size > 0) {
    return messages.filter((_, i) => !toRemove.has(i))
  }
  return messages
}

// parseChatLinesIncremental 增量解析新增的 SSE 行。
// parseState 包含 { currentMessage, toolUseMap, pendingPatches, skipTypes }。
// 返回 { newMessages, parseState }。
// - newMessages: 新增的消息数组，调用方通过 push 追加即可
// - parseState: 更新后的解析状态，供下一次增量调用
// - parseState.pendingPatches: 需要在已渲染消息上通过 $set 补丁的 { msgIndex, blockId, resultData } 列表
function parseChatLinesIncremental(newLines, parseState, msgIndexOffset) {
  if (!newLines || newLines.length === 0) {
    return {
      newMessages: [],
      parseState: parseState || { currentMessage: null, toolUseMap: new Map(), pendingPatches: [] },
    }
  }

  let currentMessage = (parseState && parseState.currentMessage) || null
  const toolUseMap = (parseState && parseState.toolUseMap) || new Map()
  const messages = []

  // 继承或检测 skipTypes：若已确认存在 stream_event，则跳过 assistant 行
  let skipTypes = (parseState && parseState.skipTypes) || null
  if (!skipTypes) {
    for (const line of newLines) {
      try {
        const obj = JSON.parse(line)
        if (obj.type === 'stream_event') {
          skipTypes = new Set(['assistant'])
          break
        }
      } catch (e) { /* ignore parse errors in quick scan */ }
    }
  }

  for (const line of newLines) {
    currentMessage = parseOneLine(line, messages, { value: currentMessage }, toolUseMap, msgIndexOffset || 0, skipTypes)
  }

  // 不 flush currentMessage —— 保留到下次增量或最终 flush
  const pendingPatches = []
  for (const [toolUseId, entry] of toolUseMap) {
    if (entry._pendingResult) {
      const resultData = entry._pendingResult
      delete entry._pendingResult
      if (entry.isNew) {
        if (entry.block) {
          entry.block._result = resultData
        } else if (entry.msg) {
          entry.msg._result = resultData
        }
      } else {
        // 需要调用方在已渲染消息上通过 $set 打补丁
        pendingPatches.push({
          toolUseId,
          blockId: entry.block ? entry.block.id : entry.msg ? entry.msg.id : '',
          resultData,
        })
      }
    }
  }

  // 本批处理完毕，后续批次中的 tool_result 需通过 pendingPatches (Vue $set) 补丁到已渲染消息
  for (const [, entry] of toolUseMap) {
    entry.isNew = false
  }

  return {
    newMessages: messages,
    parseState: { currentMessage, toolUseMap, pendingPatches, skipTypes },
  }
}

// flushParseState 刷新解析状态的最终 currentMessage。
function flushParseState(parseState) {
  const msgs = []
  if (parseState && parseState.currentMessage && (parseState.currentMessage.content.length > 0 || parseState.currentMessage.thinking)) {
    msgs.push(parseState.currentMessage)
    parseState.currentMessage = null
  }
  return msgs
}

// 带 cliType 分发的包装函数（cliType 默认 'claude'，不影响已有调用）
function parseChatLinesDispatch(lines, cliType) {
  if (cliType === 'codex') return codexParser.parseChatLines(lines)
  return parseChatLines(lines)
}

function parseChatLinesIncrementalDispatch(newLines, parseState, msgIndexOffset, cliType) {
  if (cliType === 'codex') return codexParser.parseChatLinesIncremental(newLines, parseState, msgIndexOffset)
  return parseChatLinesIncremental(newLines, parseState, msgIndexOffset)
}

function flushParseStateDispatch(parseState, cliType) {
  if (cliType === 'codex') return codexParser.flushParseState(parseState)
  return flushParseState(parseState)
}

export default {
  parseChatLines: parseChatLinesDispatch,
  parseChatLinesIncremental: parseChatLinesIncrementalDispatch,
  flushParseState: flushParseStateDispatch,
  // 保留直接访问原始 Claude 解析器的途径（向后兼容）
  _claudeParseChatLines: parseChatLines,
  _claudeParseChatLinesIncremental: parseChatLinesIncremental,
  _claudeFlushParseState: flushParseState,
}
