import taskStore from '@/utils/task_progress_store'

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
  const keys = Object.keys(parsed)
  if (keys.length === 1) return parsed[keys[0]]
  return null
}

// parseChatLines 将 claude stream-json 输出解析为可渲染的消息数组。
// parseOneLine 解析单行 SSE 数据，更新 messages、currentMessage 和 toolUseMap。
// 提取为独立函数以便全量解析和增量解析共用。
function parseOneLine(line, messages, currentMessageRef, toolUseMap, msgIndexOffset) {
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
  let cm = currentMessageRef.value

  if (lineType === 'system') {
    const subtype = obj.subtype || ''
    if (subtype === 'init') {
      messages.push({ type: 'system_init', text: obj.is_resume ? '继续对话' : '会话已创建', model: obj.model || '', sessionId: obj.session_id || '' })
    } else if (subtype === 'command') {
      messages.push({ type: 'system_command', text: obj.text || '', collapsed: true })
    } else if (subtype === 'hook_started' || subtype === 'hook_response') {
      messages.push({ type: 'system_hook', text: subtype === 'hook_started' ? 'Hook started: ' + (obj.hook_name || '') : 'Hook response: ' + (obj.hook_name || ''), collapsed: true })
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
    for (const part of content) {
      if (part.type === 'tool_result') {
        const text = typeof part.content === 'string' ? part.content : JSON.stringify(part.content || '')
        const toolUseId = part.tool_use_id || ''
        // 尝试即时配对
        const target = toolUseMap.get(toolUseId)
        if (target) {
          const resultData = { text, collapsed: true }
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
          messages.push({ type: 'tool_result', text, collapsed: true, toolUseId })
        }
      }
    }
  } else if (lineType === 'result') {
    messages.push({
      type: 'result',
      subtype: obj.subtype || '',
      durationMs: obj.duration_ms || 0,
      numTurns: obj.num_turns || 0,
      usage: obj.usage || null,
      isError: obj.is_error || false,
    })
  } else if (lineType === 'assistant') {
    const content = obj.message?.content || []
    for (const part of content) {
      if (part.type === 'text') {
        messages.push({ type: 'assistant_text', text: part.text || '' })
      } else if (part.type === 'thinking') {
        messages.push({ type: 'assistant_thinking', text: part.thinking || '', collapsed: true })
      } else if (part.type === 'tool_use') {
        const inputObj = part.input || {}
        const di = buildToolDisplayInput(part.name, inputObj)
        const tuMsg = { type: 'tool_use', name: part.name || '', id: part.id || '', input: JSON.stringify(inputObj, null, 2), displayInput: di }
        messages.push(tuMsg)
        if (part.id) {
          toolUseMap.set(part.id, { msg: tuMsg, isNew: true })
        }
      }
    }
  } else if (lineType === 'chat') {
    if (obj.subtype === 'completed') {
      messages.push({ type: 'chat_completed', text: '对话已完成' })
    }
  } else if (lineType === 'parse_error') {
    const data = obj.data || {}
    messages.push({ type: 'parse_error', text: data.line || obj.line || '', error: data.error || '' })
  } else if (lineType === 'raw_text') {
    const data = obj.data || {}
    messages.push({ type: 'raw_text', text: data.text || obj.text || '' })
  } else if (lineType === 'error') {
    messages.push({ type: 'error', text: obj.text || '' })
  }

  return cm
}

function parseChatLines(lines) {
  if (!lines || lines.length === 0) return []

  const messages = []
  let currentMessage = null
  const toolUseMap = new Map()

  for (const line of lines) {
    currentMessage = parseOneLine(line, messages, { value: currentMessage }, toolUseMap)
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
// parseState 包含 { currentMessage, toolUseMap, pendingPatches }。
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

  for (const line of newLines) {
    currentMessage = parseOneLine(line, messages, { value: currentMessage }, toolUseMap, msgIndexOffset || 0)
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
    parseState: { currentMessage, toolUseMap, pendingPatches },
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

export default {
  parseChatLines,
  parseChatLinesIncremental,
  flushParseState,
}
