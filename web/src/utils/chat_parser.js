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
function parseChatLines(lines) {
  if (!lines || lines.length === 0) return []

  const messages = []
  let currentMessage = null

  for (const line of lines) {
    if (!line || !line.trim()) continue
    let obj = null
    try {
      obj = JSON.parse(line)
    } catch (e) {
      // 非 JSON 行直接显示原始文本，不丢弃任何内容
      messages.push({ type: 'raw_text', text: line })
      continue
    }

    const lineType = obj.type || ''

    if (lineType === 'system') {
      const subtype = obj.subtype || ''
      if (subtype === 'init') {
        messages.push({ type: 'system_init', text: '会话已创建', model: obj.model || '', sessionId: obj.session_id || '' })
      } else if (subtype === 'command') {
        messages.push({ type: 'system_command', text: obj.text || '', collapsed: true })
      } else if (subtype === 'hook_started' || subtype === 'hook_response') {
        messages.push({ type: 'system_hook', text: subtype === 'hook_started' ? 'Hook started: ' + (obj.hook_name || '') : 'Hook response: ' + (obj.hook_name || ''), collapsed: true })
      } else if (subtype === 'status') {
        const statusMap = { requesting: '请求中', compressing: '压缩中' }
        messages.push({ type: 'system_status', status: obj.status || '', text: statusMap[obj.status] || obj.status })
      } else if (subtype === 'task_started') {
        const taskMsg = { type: 'system_task', description: obj.description || '', taskId: obj.task_id || '', status: 'started', _msgIndex: messages.length }
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
          _msgIndex: messages.length,
        }
        messages.push(taskMsg)
        taskStore.updateFromMessage(taskMsg)
      } else if (subtype === 'task_notification') {
        const taskMsg = { type: 'system_task', description: obj.summary || '', taskId: obj.task_id || '', status: obj.status || '', _msgIndex: messages.length }
        messages.push(taskMsg)
        taskStore.updateFromMessage(taskMsg)
      } else {
        messages.push({ type: 'system', text: JSON.stringify(obj) })
      }
    } else if (lineType === 'stream_event') {
      const event = obj.event || {}
      const eventType = event.type || ''

      if (eventType === 'message_start') {
        currentMessage = {
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
        if (!currentMessage) continue
        currentMessage._blockType = event.content_block?.type || ''
        if (currentMessage._blockType === 'tool_use') {
          currentMessage.content.push({
            type: 'tool_use',
            name: event.content_block?.name || '',
            id: event.content_block?.id || '',
            input: '',
          })
        }
      } else if (eventType === 'content_block_delta') {
        if (!currentMessage) continue
        const delta = event.delta || {}
        if (delta.type === 'text_delta') {
          if (currentMessage._blockType === 'tool_use' && currentMessage.content.length > 0) {
            const last = currentMessage.content[currentMessage.content.length - 1]
            if (last.type === 'tool_use') {
              last.input += (delta.text || '')
            }
          } else {
            const lastContent = currentMessage.content[currentMessage.content.length - 1]
            if (lastContent && lastContent.type === 'text') {
              lastContent.text += (delta.text || '')
            } else {
              currentMessage.content.push({ type: 'text', text: delta.text || '' })
            }
          }
        } else if (delta.type === 'thinking_delta') {
          currentMessage.thinking += (delta.thinking || '')
        } else if (delta.type === 'input_json_delta') {
          if (currentMessage.content.length > 0) {
            const last = currentMessage.content[currentMessage.content.length - 1]
            if (last.type === 'tool_use') {
              last.input += (delta.partial_json || '')
            }
          }
        }
      } else if (eventType === 'content_block_stop') {
        if (currentMessage) {
          // 格式化 tool_use 的 input
          if (currentMessage._blockType === 'tool_use' && currentMessage.content.length > 0) {
            const last = currentMessage.content[currentMessage.content.length - 1]
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
          currentMessage._blockType = ''
        }
      } else if (eventType === 'message_delta') {
        if (currentMessage) {
          currentMessage.usage = event.delta?.usage || event.usage || null
        }
      } else if (eventType === 'message_stop') {
        if (currentMessage && (currentMessage.content.length > 0 || currentMessage.thinking)) {
          messages.push(currentMessage)
        }
        currentMessage = null
      }
    } else if (lineType === 'user') {
      const content = obj.message?.content || []
      for (const part of content) {
        if (part.type === 'tool_result') {
          const text = typeof part.content === 'string' ? part.content : JSON.stringify(part.content || '')
          messages.push({ type: 'tool_result', text: text, collapsed: true, toolUseId: part.tool_use_id || '' })
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
          messages.push({ type: 'tool_use', name: part.name || '', id: part.id || '', input: JSON.stringify(inputObj, null, 2), displayInput: di })
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
  }

  // flush remaining message
  if (currentMessage && (currentMessage.content.length > 0 || currentMessage.thinking)) {
    messages.push(currentMessage)
  }

  // 后处理：将 tool_result 配对到对应的 tool_use 下
  const toolUseMap = new Map()
  for (let i = 0; i < messages.length; i++) {
    const msg = messages[i]
    if (msg.type === 'assistant') {
      for (const block of (msg.content || [])) {
        if (block.type === 'tool_use' && block.id) {
          toolUseMap.set(block.id, { msg, block, index: i })
        }
      }
    } else if (msg.type === 'tool_use' && msg.id) {
      toolUseMap.set(msg.id, { msg })
    }
  }

  const toRemove = new Set()
  for (let i = 0; i < messages.length; i++) {
    const msg = messages[i]
    if (msg.type === 'tool_result' && msg.toolUseId) {
      const target = toolUseMap.get(msg.toolUseId)
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

export default {
  parseChatLines,
}
