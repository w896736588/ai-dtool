const assert = require('assert')
const fs = require('fs')
const path = require('path')

const homeVuePath = path.join(__dirname, '../src/components/Home.vue')
const source = fs.readFileSync(homeVuePath, 'utf8')

const run = () => {
  assert.ok(
    !source.includes("menuKeyStore: 'lastMenuName.v2'"),
    'Home page should remove the last menu cache key'
  )

  assert.ok(
    !source.includes("this.menuName = this.$helperStore.getStore(this.menuKeyStore)") &&
    !source.includes("this.$helperStore.setStore(_that.menuKeyStore, this.menuName)"),
    'Home page should no longer read or write last menu cache'
  )

  assert.ok(
    source.includes("this.menuName = this.$route.path || '/Dashboard'"),
    'Home page should initialize menu highlight from the current route instead of cache'
  )

  console.log('home_menu_cache_removal tests passed')
}

run()
