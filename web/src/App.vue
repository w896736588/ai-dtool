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
html,
body,
#app {
  /* 根节点统一占满视口，给内部 100% 高度布局提供稳定基线。
     Let root nodes fill the viewport so inner 100% layouts have a reliable baseline. */
  height: 100%;
}

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
