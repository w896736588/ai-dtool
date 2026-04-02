const assert = require('assert')
const feedback = require('../src/utils/memory_fragment_feedback.cjs')

assert.strictEqual(feedback.isMemoryFragmentTabName('fragment-12'), true)
assert.strictEqual(feedback.isMemoryFragmentTabName('home'), false)

const emptyState = {}
const nextState = feedback.activateMemorySaveFeedback(emptyState, 12, 1000, 1000)
assert.deepStrictEqual(Object.keys(nextState), ['12'])
assert.strictEqual(nextState['12'].visible, true)
assert.strictEqual(nextState['12'].expiresAt, 2000)

const cleanedState = feedback.clearExpiredMemorySaveFeedback(nextState, 1999)
assert.deepStrictEqual(cleanedState, nextState)
assert.deepStrictEqual(feedback.clearExpiredMemorySaveFeedback(nextState, 2000), {})

console.log('memory_fragment_feedback tests passed')
