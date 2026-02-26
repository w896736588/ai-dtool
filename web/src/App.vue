<template>
  <div id="app">
    <router-view/>
  </div>
</template>

<script>
import base from '@/utils/base'
export default {
  name: 'App',
  components: {
  },
  mounted() {
    base.DisableSaveShortcut()
    let favicon = document.querySelector('link[rel="icon"]')
    if (process.env.NODE_ENV === 'production') {
      favicon.href = './favicon.ico'
    }

  }
}
</script>

<style>
#app {
  font-family: Consolas , Avenir, Helvetica, Arial, sans-serif !important;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

/* 隐藏所有滚动条 */
::-webkit-scrollbar {
  display: none; /* Chrome/Safari */
}

* {
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE/Edge */
}

body{
  margin : 0;
}
</style>
