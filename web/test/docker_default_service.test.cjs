const assert = require('assert').strict
const {
  isDockerDefaultServiceEnabled,
  normalizeDockerDefaultServices,
  stringifyDockerDefaultServices,
  toggleDockerDefaultService,
} = require('../src/utils/docker_default_service.cjs')

assert.deepEqual(
  normalizeDockerDefaultServices('api, worker,api, , web'),
  ['api', 'worker', 'web'],
  '默认服务列表应去空格、去空项并去重'
)

assert.equal(
  isDockerDefaultServiceEnabled('api,worker', 'worker'),
  true,
  '已存在的服务应识别为默认服务'
)

assert.equal(
  toggleDockerDefaultService('api,worker', 'web', true),
  'api,worker,web',
  '加入默认服务时应追加到配置字符串'
)

assert.equal(
  toggleDockerDefaultService('api,worker', 'worker', false),
  'api',
  '移除默认服务时应从配置字符串里删除对应项'
)

assert.equal(
  stringifyDockerDefaultServices(['api', 'worker', 'api']),
  'api,worker',
  '序列化默认服务时应保持去重后的逗号分隔格式'
)

console.log('docker_default_service tests passed')
