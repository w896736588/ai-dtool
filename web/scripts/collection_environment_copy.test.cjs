const assert = require('assert')
const fs = require('fs')
const path = require('path')

const collectionEnvironmentVuePath = path.join(__dirname, '../src/components/api/CollectionEnvironment.vue')
const source = fs.readFileSync(collectionEnvironmentVuePath, 'utf8')

const run = () => {
  assert.ok(
    /<pl-button[^>]*@click="handleCopyEnv\(row\)"[^>]*>复制<\/pl-button>/.test(source),
    'Environment list should provide a copy button for each row'
  )

  assert.ok(
    source.includes('handleCopyEnv(env) {') &&
    source.includes("name: `${env.name || '新环境'}-复制`") &&
    source.includes('copiedVariables: this.cloneEnvironmentVariables(env.variables)'),
    'Copy action should create a draft environment with duplicated variables'
  )

  assert.ok(
    source.includes('cloneEnvironmentVariables(variables) {') &&
    source.includes('saveCopiedVariables(env, savedEnvId) {'),
    'Environment copy flow should include variable cloning and persistence helpers'
  )

  console.log('collection_environment_copy tests passed')
}

run()
