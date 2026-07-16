<template>
  <div id="app">
    <router-view/>
  </div>
</template>

<script>
import base from '@/utils/base'

export default {
  name: 'App',
  mounted() {
    base.DisableSaveShortcut()
    const favicon = document.querySelector('link[rel="icon"]')
    if (process.env.NODE_ENV === 'production' && favicon) {
      favicon.href = './favicon.ico'
    }
  },
}
</script>

<style>
html,
body,
#app {
  height: 100%;
}

#app {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

body {
  margin: 0;
}
</style>
