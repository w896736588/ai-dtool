import assert from 'node:assert/strict'
import { formatInlineUsage } from '../src/utils/chat_usage.mjs'

{
  const text = formatInlineUsage({
    input_tokens: 1825,
    output_tokens: 89,
    cache_read_input_tokens: 120,
    cache_creation_input_tokens: 45,
  })

  assert.equal(text, 'input: 1825 | output: 89 | cache: 120 | cache create: 45')
}

{
  const text = formatInlineUsage({
    input_tokens: 20,
    output_tokens: 8,
  })

  assert.equal(text, 'input: 20 | output: 8 | cache: 0 | cache create: 0')
}
