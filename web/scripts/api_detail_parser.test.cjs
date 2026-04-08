const assert = require('assert')
const path = require('path')

const parser = require(path.join(__dirname, '../src/utils/api_detail_parser.cjs'))

const run = () => {
  assert.deepStrictEqual(
    parser.parseApiObjectField('', {}),
    {},
    'Empty string object payload should fall back to an empty object'
  )

  assert.deepStrictEqual(
    parser.parseApiObjectField({ Authorization: 'Bearer demo' }, {}),
    { Authorization: 'Bearer demo' },
    'Object payload should be returned as a shallow clone'
  )

  assert.deepStrictEqual(
    parser.parseApiArrayField('', []),
    [],
    'Empty string array payload should fall back to an empty array'
  )

  assert.deepStrictEqual(
    parser.parseApiArrayField([{ key: 'token', value: 'demo' }], []),
    [{ key: 'token', value: 'demo' }],
    'Array payload should be returned as a shallow clone'
  )

  assert.deepStrictEqual(
    parser.parseApiArrayField('[', []),
    [],
    'Truncated JSON array payload should fall back to an empty array'
  )

  assert.deepStrictEqual(
    parser.parseApiObjectField('{"ok":true}', {}),
    { ok: true },
    'Valid JSON object string should still parse successfully'
  )

  console.log('api_detail_parser tests passed')
}

run()
