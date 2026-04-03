const assert = require('assert')
const fs = require('fs')
const path = require('path')

const filePath = path.join(__dirname, '../src/components/Api.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert.match(
  source,
  /async openCopyApiDialog\(api\)/,
  '复制接口弹窗应升级为异步流程，以便先拉取完整详情'
)

assert.match(
  source,
  /await this\.requestApi\('ApisDetailByIds',\s*\{\s*ids:\s*\[api\.id\]\s*\}\)/,
  '复制接口前应先请求完整接口详情'
)

assert.match(
  source,
  /const copySource = .*detail.*\|\|\s*api/,
  '复制接口时应优先使用详情接口返回的数据作为复制源'
)

assert.match(
  source,
  /this\.dialogData\.copyApi = JSON\.parse\(JSON\.stringify\(copySource\)\)/,
  '复制接口弹窗应基于完整详情构造表单数据'
)

console.log('api_copy_full_detail tests passed')
